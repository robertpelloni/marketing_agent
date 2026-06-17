package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type EventType string

const (
	EventDealStateChange EventType = "deal_state_change"
)

type WebhookPayload struct {
	Event     EventType    `json:"event"`
	DealID    int64        `json:"deal_id"`
	NewState  db.LeadState `json:"new_state"`
	Timestamp time.Time    `json:"timestamp"`
}

type Dispatcher struct {
	URL    string
	Client *http.Client
}

func NewDispatcher(url string) *Dispatcher {
	return &Dispatcher{
		URL: url,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, dealID int64, newState db.LeadState) error {
	if d.URL == "" { return nil }

	payload := WebhookPayload{
		Event:     EventDealStateChange,
		DealID:    dealID,
		NewState:  newState,
		Timestamp: time.Now(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Retry logic
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", d.URL, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := d.Client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				slog.Info("Webhook dispatched successfully", "deal", dealID, "state", newState)
				return nil
			}
			lastErr = fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		slog.Warn("Webhook dispatch failed, retrying", "attempt", i+1, "error", lastErr)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("webhook dispatch failed after %d attempts: %w", maxRetries, lastErr)
}
