package web

import (
	"os"

	"encoding/json"
	"io"
	"log/slog"
	"net/http"
    "strconv"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
    "github.com/robertpelloni/marketing_agent/internal/db"
)

func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(r.Context(), "Error reading request body", "error", err)
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// If no secret is configured, skip verification (useful for local dev/testing without forwarding)
	var event stripe.Event
	if webhookSecret == "" {
		slog.WarnContext(r.Context(), "Stripe webhook secret is not configured; skipping signature verification")
		err = json.Unmarshal(payload, &event)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to parse webhook JSON", "error", err)
			http.Error(w, "Failed to parse webhook JSON", http.StatusBadRequest)
			return
		}
	} else {
		event, err = webhook.ConstructEvent(payload, sigHeader, webhookSecret)
		if err != nil {
			slog.ErrorContext(r.Context(), "Error verifying webhook signature", "error", err)
			http.Error(w, "Error verifying webhook signature", http.StatusBadRequest)
			return
		}
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			slog.ErrorContext(r.Context(), "Error parsing checkout session", "error", err)
			http.Error(w, "Error parsing checkout session", http.StatusBadRequest)
			return
		}

		slog.InfoContext(r.Context(), "Received checkout.session.completed", "session_id", session.ID)

		// Extract deal ID from metadata
		dealIDStr, ok := session.Metadata["deal_id"]
		if !ok {
			slog.WarnContext(r.Context(), "No deal_id found in checkout session metadata", "session_id", session.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		dealID, err := strconv.ParseInt(dealIDStr, 10, 64)
		if err != nil {
			slog.ErrorContext(r.Context(), "Invalid deal_id in metadata", "deal_id_str", dealIDStr, "error", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Update deal state and pricing
		amountTotal := session.AmountTotal // Amount is in cents
		actualRevenue := float64(amountTotal) / 100.0

		// Use the DB to update the deal details
        // Update Deal State to StateClosedWon
        err = s.db.UpdateDealState(r.Context(), dealID, db.StateClosedWon)
        if err != nil {
            slog.ErrorContext(r.Context(), "Failed to update deal state", "deal_id", dealID, "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Ensure deal pricing matches actual paid amount
        err = s.db.UpdateDealDetails(r.Context(), dealID, actualRevenue, "Processed via Stripe Checkout")
        if err != nil {
            slog.ErrorContext(r.Context(), "Failed to update deal details", "deal_id", dealID, "error", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

		// Log audit event for this state transition
		err = s.db.CreateAuditLog(r.Context(), &db.AuditLog{
			EntityID: dealID,
			Type:     "deal_transition",
			Action:   string(db.StateClosedWon),
			Actor:    "stripe_webhook",
			Metadata: string(event.Data.Raw),
		})
        if err != nil {
            slog.WarnContext(r.Context(), "Failed to create audit log", "deal_id", dealID, "error", err)
        }

        slog.InfoContext(r.Context(), "Successfully processed checkout.session.completed", "deal_id", dealID, "revenue", actualRevenue)
	}

	w.WriteHeader(http.StatusOK)
}
