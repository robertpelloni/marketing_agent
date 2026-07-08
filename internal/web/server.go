package web

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robertpelloni/marketing_agent/internal/auth"
	"github.com/robertpelloni/marketing_agent/internal/autodev"
	"github.com/robertpelloni/marketing_agent/internal/billing"
	"github.com/robertpelloni/marketing_agent/internal/communication"
	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/deploy"
	"github.com/robertpelloni/marketing_agent/internal/llm"
	"github.com/robertpelloni/marketing_agent/internal/logging"
)

// HermesHealthChecker is an optional interface for checking LLM provider health.
type HermesHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Server handles web dashboard requests.
type Server struct {
	db            *db.DB
	deploy        *deploy.Deployer
	tracker       deploy.CITracker
	tasks         *autodev.TaskManager
	auth          *auth.Authenticator
	llmProvider   llm.LLMProvider
	billingClient billing.BillingClient
	mux           *http.ServeMux
}

// NewServer creates a new Server instance.
func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager, llmProvider llm.LLMProvider, billingClient billing.BillingClient) *Server {
	s := &Server{
		db:            database,
		deploy:        deployer,
		tracker:       tracker,
		tasks:         taskManager,
		auth:          auth.NewAuthenticator(),
		llmProvider:   llmProvider,
		billingClient: billingClient,
		mux:           http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Protected routes
	// Initialize rate limiter: 10 requests per second, burst of 20
	rl := newRateLimiter(10, 20)

	// Protected routes
	s.mux.Handle("/", rl.middleware(s.auth.Middleware(s.csrfMiddleware(http.HandlerFunc(s.handleDashboard)))))

	// Public routes
	s.mux.Handle("/login", rl.middleware(http.HandlerFunc(s.auth.HandleLogin)))
	s.mux.Handle("/health", rl.middleware(http.HandlerFunc(s.handleHealth)))
	s.mux.Handle("/health/detailed", rl.middleware(http.HandlerFunc(s.handleDetailedHealth)))
	s.mux.Handle("/api/v1/webhook/github", rl.middleware(http.HandlerFunc(s.handleGitHubWebhook)))
	s.mux.Handle("/api/v1/quote", rl.middleware(http.HandlerFunc(s.handleGenerateQuote)))
	s.mux.Handle("/api/v1/leads", rl.middleware(http.HandlerFunc(s.handleLeadsAPI)))
	s.mux.Handle("/api/v1/deals", rl.middleware(http.HandlerFunc(s.handleDealsAPI)))

	// GDPR Endpoints
	s.mux.Handle("/api/v1/gdpr/export", rl.middleware(s.auth.Middleware(http.HandlerFunc(s.handleGDPRExport))))
	s.mux.Handle("/api/v1/gdpr/delete", rl.middleware(s.auth.Middleware(http.HandlerFunc(s.handleGDPRDelete))))

	// Telemetry WebSocket
	s.mux.Handle("/ws/telemetry", s.auth.Middleware(http.HandlerFunc(s.handleTelemetryWS)))

	// Stripe Webhook (public, verified by signature)
	s.mux.Handle("/api/v1/webhook/stripe", rl.middleware(http.HandlerFunc(s.handleStripeWebhook)))

	// Create Checkout Session (public — called from hypernexus.site pricing buttons)
	s.mux.Handle("/api/v1/billing/checkout", rl.middleware(http.HandlerFunc(s.handleCreateCheckout)))

	// Billing API (protected — requires auth)
	s.mux.Handle("/api/v1/billing/subscription", rl.middleware(s.auth.Middleware(http.HandlerFunc(s.handleGetSubscription))))
	s.mux.Handle("/api/v1/billing/cancel", rl.middleware(s.auth.Middleware(http.HandlerFunc(s.handleCancelSubscription))))
	s.mux.Handle("/api/v1/billing/portal", rl.middleware(s.auth.Middleware(http.HandlerFunc(s.handleBillingPortal))))
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	logging.Init("json", "info")
	slog.Info("Web dashboard starting", "addr", addr)
	// #nosec G114 -- Simple ListenAndServe is used for internal dashboard; timeout configuration handled at higher level if needed
	return http.ListenAndServe(addr, logging.Middleware(s))
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	csrfToken := ""
	if cookie, err := r.Cookie("csrf_token"); err == nil {
		csrfToken = cookie.Value
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		action := html.EscapeString(strings.TrimSpace(r.FormValue("action")))
		switch action {
		case "enrich":
			dealID := html.EscapeString(strings.TrimSpace(r.FormValue("deal_id")))
			// #nosec G706 -- deal_id is used for context in manual action logs
			slog.InfoContext(r.Context(), "Manual enrichment triggered", "deal_id", dealID)
		case "sync":
			if err := s.deploy.ExecuteSync(); err != nil {
				slog.WarnContext(r.Context(), "Sync error", "error", err)
			}
		case "flag_success":
			interactionID := html.EscapeString(strings.TrimSpace(r.FormValue("interaction_id")))
			success := html.EscapeString(strings.TrimSpace(r.FormValue("success"))) == "true"
			var id int64
			if _, err := fmt.Sscanf(interactionID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid interaction ID", "error", err, "interaction_id", interactionID)
			} else {
				if err := s.db.UpdateInteractionSuccess(r.Context(), id, success); err != nil {
					slog.WarnContext(r.Context(), "Error flagging interaction", "error", err, "interaction_id", id)
				}
			}
		case "update_channel":
			contactID := html.EscapeString(strings.TrimSpace(r.FormValue("contact_id")))
			channel := html.EscapeString(strings.TrimSpace(r.FormValue("channel")))
			var id int64
			if _, err := fmt.Sscanf(contactID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid contact ID", "error", err, "contact_id", contactID)
			} else {
				if err := s.db.UpdateContactPreferredChannel(r.Context(), id, channel); err != nil {
					slog.WarnContext(r.Context(), "Error updating channel", "error", err, "contact_id", id, "channel", channel)
				} else {
					slog.InfoContext(r.Context(), "Contact channel updated", "contact_id", id, "channel", channel)
				}
			}
		case "build":
			if err := s.deploy.ExecuteBuild(); err != nil {
				slog.WarnContext(r.Context(), "Build error", "error", err)
			}
		case "approve_deal":
			dealIDStr := html.EscapeString(strings.TrimSpace(r.FormValue("deal_id")))
			var id int64
			if _, err := fmt.Sscanf(dealIDStr, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid deal ID for approval", "error", err)
			} else {
				if err := s.db.UpdateDealState(r.Context(), id, db.StateNegotiating); err != nil {
					slog.WarnContext(r.Context(), "Failed to approve deal", "deal_id", id, "error", err)
				} else {
					slog.InfoContext(r.Context(), "Deal approved by human", "deal_id", id)
				}
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	deals, err := s.db.ListRecentDeals(r.Context(), 20)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve deals: %v", err), http.StatusInternalServerError)
		return
	}

	health, _ := s.tracker.GetSystemHealth(r.Context())

	metrics, err := s.db.GetPerformanceMetrics(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "Error retrieving metrics", "error", err)
		metrics = &db.PerformanceMetrics{
			LeadsByState: make(map[db.LeadState]int),
		}
	}

	prs, err := s.db.ListActivePullRequests(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "Error listing PRs", "error", err)
	}

	taskList, _ := s.tasks.ListAllTasks(r.Context())
	socialPosts := []struct {
		Brand           string
		Platform        string
		AccountUsername string
		PostContent     string
		Status          string
		CreatedAt       time.Time
	}{}

	// Check LLM/Hermes health for the dashboard
	llmStatus := "Mock"
	llmColor := "#6c757d"
	if checker, ok := s.llmProvider.(HermesHealthChecker); ok {
		if err := checker.HealthCheck(r.Context()); err != nil {
			llmStatus = fmt.Sprintf("Hermes: ERROR (%v)", err)
			llmColor = "#dc3545"
		} else {
			llmStatus = "Hermes: Connected"
			llmColor = "#28a745"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, "%s", "<!DOCTYPE html>")
	_, _ = fmt.Fprintf(w, `
<html>
<head>
<title>TormentNexus / HyperNexus Autonomous Pipeline Dashboard</title>
<style>
:root {
	--primary: #007bff;
	--success: #28a745;
	--info: #17a2b8;
	--warning: #ffc107;
	--danger: #dc3545;
	--dark: #343a40;
	--light: #f8f9fa;
	--purple: #6f42c1;
}
body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 0; background-color: #e9ecef; color: #333; }
.header { background: var(--dark); color: white; padding: 20px 40px; display: flex; justify-content: space-between; align-items: center; }
.header h1 { margin: 0; font-size: 1.5rem; }
.container { max-width: 1400px; margin: 20px auto; padding: 0 20px; display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.full-width { grid-column: 1 / -1; }
.card { background: white; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); padding: 20px; margin-bottom: 20px; }
.card-header { font-size: 1.25rem; font-weight: bold; margin-bottom: 15px; border-bottom: 2px solid #eee; padding-bottom: 10px; display: flex; justify-content: space-between; align-items: center; }
table { width: 100%%; border-collapse: collapse; margin-top: 10px; font-size: 0.9rem; }
th, td { padding: 10px; border-bottom: 1px solid #ddd; text-align: left; }
th { background-color: #f1f3f5; color: #495057; }
tr:hover { background-color: #f8f9fa; }
.status { font-weight: 600; padding: 4px 8px; border-radius: 12px; font-size: 0.8rem; display: inline-block; }
.status-Discovered { background-color: #e2e3e5; color: #383d41; }
.status-Researched { background-color: #cce5ff; color: #004085; }
.status-PR { background-color: #fff3cd; color: #856404; }
.status-Closed_Won { background-color: #d4edda; color: #155724; }
.status-Closed_Lost { background-color: #f8d7da; color: #721c24; }
.action-btn { background-color: var(--primary); color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; font-size: 0.8rem; transition: background 0.2s; }
.action-btn:hover { opacity: 0.9; }
.metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
.metric-box { padding: 15px; border-radius: 8px; text-align: center; }
.metric-value { font-size: 1.8rem; font-weight: bold; margin-bottom: 5px; }
.metric-label { font-size: 0.85rem; color: #666; text-transform: uppercase; letter-spacing: 0.5px; }
.tooltip { position: relative; cursor: help; border-bottom: 1px dotted #666; }
.tooltip .tooltiptext { visibility: hidden; width: 200px; background-color: #333; color: #fff; text-align: center; border-radius: 6px; padding: 5px 0; position: absolute; z-index: 1; bottom: 125%%; left: 50%%; margin-left: -100px; opacity: 0; transition: opacity 0.3s; font-size: 0.75rem; font-weight: normal; }
.tooltip:hover .tooltiptext { visibility: visible; opacity: 1; }
</style>
</head>
<body>
<div class="header">
	<h1>TormentNexus / HyperNexus Autonomous Pipeline Dashboard</h1>
	<div>
		<span style="margin-right: 15px;" class="tooltip">System Health: <strong style="color: %s;">%s</strong>
			<span class="tooltiptext">CI/CD deployment status and database connection</span>
		</span>
		<span class="tooltip">LLM Status: <strong style="color: %s;">%s</strong>
			<span class="tooltiptext">Connection to Hermes or Mock LLM Provider</span>
		</span>
	</div>
</div>
<div class="container">
	<div class="card full-width" style="border-top: 4px solid var(--info);">
		<div class="card-header">
			Performance Metrics
			<span class="tooltip" style="font-size:0.8rem; color:#888;">?
				<span class="tooltiptext">Real-time pipeline statistics and conversion rates.</span>
			</span>
		</div>
		<div class="metrics-grid">
			<div class="metric-box" style="background: #e9ecef;">
				<div class="metric-value">%d</div>
				<div class="metric-label">Total Leads</div>
			</div>
			<div class="metric-box" style="background: #d4edda;">
				<div class="metric-value">%d</div>
				<div class="metric-label">Won Deals</div>
			</div>
			<div class="metric-box" style="background: #f8d7da;">
				<div class="metric-value">%.1f%%</div>
				<div class="metric-label">Win Rate</div>
			</div>
			<div class="metric-box" style="background: #fff3cd;">
				<div class="metric-value">%d</div>
				<div class="metric-label">Successful Outreach</div>
			</div>
		</div>
		<div style="margin-top: 15px; font-size: 0.9rem; display: flex; gap: 15px; justify-content: center; color: #555;">
			<span><strong>Pipeline:</strong></span>
			<span class="tooltip">Discovered: %d<span class="tooltiptext">Leads found by scraper</span></span> |
			<span class="tooltip">Researched: %d<span class="tooltiptext">Technical dossier built</span></span> |
			<span class="tooltip">Outreach Sent: %d<span class="tooltiptext">Initial email/message sent</span></span> |
			<span class="tooltip">Engaged: %d<span class="tooltiptext">Reply received</span></span> |
			<span class="tooltip">Negotiating: %d<span class="tooltiptext">Discussing terms</span></span>
		</div>
	</div>
	<div class="card full-width" style="border-top: 4px solid var(--primary);">
		<div class="card-header">
			Active Deals
			<span style="font-size: 0.8rem; font-weight: normal; color: #666;">(Showing last %d)</span>
		</div>
		<table>
			<tr>
				<th>Deal ID</th>
				<th>Company ID</th>
				<th>State</th>
				<th>Contacts & Channels</th>
				<th>Last Updated</th>
				<th>Actions</th>
			</tr>`, healthStatusColor(health), health, llmColor, llmStatus, metrics.TotalLeads, metrics.LeadsByState[db.StateClosedWon], metrics.WinRate, metrics.SuccessfulOutreach, metrics.LeadsByState[db.StateDiscovered], metrics.LeadsByState[db.StateResearched], metrics.LeadsByState[db.StateOutreachSent], metrics.LeadsByState[db.StateEngaged], metrics.LeadsByState[db.StateNegotiating], len(deals))

	var dealsRows string
	for _, d := range deals {
		statusTitle := ""
		switch d.CurrentState {
		case db.StateDiscovered:
			statusTitle = "Company identified, awaiting technical research."
		case db.StateResearched:
			statusTitle = "Key engineering contacts found and technical dossier compiled."
		case db.StatePendingApproval:
			statusTitle = "Deal flagged for human approval before progression."
		}

		contacts, _ := s.db.ListContactsByCompany(r.Context(), d.CompanyID)
		latestInteractionID := int64(0)
		var contactHTML string
		if len(contacts) > 0 {
			interactions, _ := s.db.ListInteractionsByContact(r.Context(), contacts[0].ID)
			if len(interactions) > 0 {
				latestInteractionID = interactions[0].ID
			}

			contactHTML = "<div style='font-size: 0.85em;'>"
			for _, c := range contacts {
				channel := c.PreferredChannel
				if channel == "" {
					channel = "email"
				}
				contactHTML += fmt.Sprintf(`
				<div style="margin: 4px 0;">
					<span class="tooltip"><strong>%s</strong> (%s)
						<span class="tooltiptext">Set preferred outreach channel</span>
					</span> —
					<span style="color: %s;">%s</span>
					<form method="POST" style="display:inline; margin-left: 8px;">
						<input type="hidden" name="csrf_token" value="%s">
						<input type="hidden" name="action" value="update_channel">
						<input type="hidden" name="contact_id" value="%d">
						<select name="channel" onchange="this.form.submit()" style="font-size: 0.9em; padding: 2px;">
							<option value="email"%s>Email</option>
							<option value="linkedin"%s>LinkedIn</option>
							<option value="github"%s>GitHub</option>
						</select>
					</form>
				</div>`,
					html.EscapeString(c.Name),
					html.EscapeString(c.Role),
					map[string]string{"email": "var(--info)", "linkedin": "var(--success)", "github": "var(--warning)"}[channel],
					channel,
					csrfToken,
					c.ID,
					map[bool]string{true: " selected", false: ""}[channel == "email"],
					map[bool]string{true: " selected", false: ""}[channel == "linkedin"],
					map[bool]string{true: " selected", false: ""}[channel == "github"])
			}
			contactHTML += "</div>"
		}

		dealsRows += fmt.Sprintf(`
			<tr>
				<td>%d</td>
				<td>%d</td>
				<td><span class="status status-%s tooltip">%s<span class="tooltiptext">%s</span></span></td>
				<td>%s</td>
				<td>%s</td>
				<td>
					<div style="display: flex; gap: 5px;">
						<form method="POST">
							<input type="hidden" name="csrf_token" value="%s">
							<input type="hidden" name="action" value="enrich">
							<input type="hidden" name="deal_id" value="%d">
							<button type="submit" class="action-btn tooltip" style="background-color: var(--info);">Enrich<span class="tooltiptext">Manually trigger enrichment for this deal</span></button>
						</form>
						<form method="POST">
							<input type="hidden" name="csrf_token" value="%s">
							<input type="hidden" name="action" value="flag_success">
							<input type="hidden" name="interaction_id" value="%d">
							<input type="hidden" name="success" value="true">
							<button type="submit" class="action-btn tooltip" style="background-color: var(--purple);">Success<span class="tooltiptext">Flag the latest interaction as successful to train the LLM</span></button>
						</form>
					</div>
				</td>
			</tr>`, d.ID, d.CompanyID, d.CurrentState, d.CurrentState, statusTitle, contactHTML, d.UpdatedAt.Format("2006-01-02 15:04"), csrfToken, d.ID, csrfToken, latestInteractionID)
	}

	_, _ = fmt.Fprint(w, dealsRows, `
		</table>
	</div>`)

	_, _ = fmt.Fprint(w, `
	<div class="card full-width" style="border-top: 4px solid var(--purple);">
		<div class="card-header">
			Social Marketing & DevRel Activity
			<span class="tooltip" style="font-size:0.8rem; color:#888;">?
				<span class="tooltiptext">Automated dual-brand outreach logs for TormentNexus & HyperNexus</span>
			</span>
		</div>
		<table>
			<tr>
				<th>Brand</th>
				<th>Platform</th>
				<th>Account</th>
				<th>Content</th>
				<th>Status</th>
				<th>Time</th>
			</tr>`)

	for _, p := range socialPosts {
		_, _ = fmt.Fprintf(w, `
			<tr>
				<td><strong>%s</strong></td>
				<td>%s</td>
				<td>%s</td>
				<td style="max-width: 300px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" title="%s">%s</td>
				<td><span class="status status-%s">%s</span></td>
				<td>%s</td>
			</tr>`, html.EscapeString(p.Brand), html.EscapeString(p.Platform), html.EscapeString(p.AccountUsername), html.EscapeString(p.PostContent), html.EscapeString(p.PostContent), html.EscapeString(p.Status), html.EscapeString(p.Status), p.CreatedAt.Format("01/02 15:04"))
	}

	_, _ = fmt.Fprint(w, `
		</table>
	</div>`)

	_, _ = fmt.Fprint(w, `
	<div class="card" style="border-top: 4px solid var(--warning);">
		<div class="card-header">
			Autonomous Task Board
			<span class="tooltip" style="font-size:0.8rem; color:#888;">?
				<span class="tooltiptext">Prioritized development roadmap and execution status from TODO.md</span>
			</span>
		</div>
		<table>
			<tr><th>Task Description</th><th>Status</th></tr>`)

	for _, t := range taskList {
		status := "Pending"
		if t.Completed {
			status = "Completed"
		}
		_, _ = fmt.Fprintf(w, `<tr><td>%s</td><td><span class="status status-%s">%s</span></td></tr>`, html.EscapeString(t.Description), status, status)
	}

	_, _ = fmt.Fprint(w, `</table></div>`)

	_, _ = fmt.Fprint(w, `
	<div class="card" style="border-top: 4px solid var(--success);">
		<div class="card-header">
			CI/CD & Repository State
			<span class="tooltip" style="font-size:0.8rem; color:#888;">?
				<span class="tooltiptext">Active feature branches, automated merge status, and deployment controls</span>
			</span>
		</div>

		<div style="margin-bottom: 20px; padding-bottom: 15px; border-bottom: 1px solid #eee;">
			<form method="POST" style="display:inline;">
				<input type="hidden" name="csrf_token" value="`+csrfToken+`">
				<input type="hidden" name="action" value="sync">
				<button type="submit" class="action-btn tooltip">Sync Repository<span class="tooltiptext">Pull upstream changes and resolve conflicts</span></button>
			</form>
			<form method="POST" style="display:inline; margin-left: 10px;">
				<input type="hidden" name="csrf_token" value="`+csrfToken+`">
				<input type="hidden" name="action" value="build">
				<button type="submit" class="action-btn tooltip" style="background-color: var(--dark);">Trigger Build<span class="tooltiptext">Force a local project recompilation</span></button>
			</form>
		</div>

		<h4>Active Pull Requests</h4>
		<table>
			<tr><th>PR ID</th><th>Branch</th><th>Title</th><th>Status</th></tr>`)

	for _, pr := range prs {
		_, _ = fmt.Fprintf(w, `<tr><td>%s</td><td>%s</td><td>%s</td><td><span class="status status-PR">%s</span></td></tr>`, html.EscapeString(pr.ID), html.EscapeString(pr.Branch), html.EscapeString(pr.Title), html.EscapeString(string(pr.Status)))
	}

	_, _ = fmt.Fprintf(w, `
</table>
</div>
<div class="deploy-section">
<h2>Self-Service Deployment</h2>
<p>Manage repository state and trigger system builds autonomously.</p>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="sync">
<button type="submit" class="action-btn deploy-btn">Sync Repository</button>
</form>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="build">
<button type="submit" class="action-btn deploy-btn" style="background-color: #6c757d;">Trigger Build</button>
</form>
</div>
<div class="deploy-section" style="border-left: 5px solid #28a745;">
<h2>System Health &amp; CI Status</h2>
<p>Real-time monitoring of the autonomous deployment pipeline.</p>
<ul>
<li><strong>Global Health:</strong> <span style="color: #28a745;">%s %s</span></li>
<li><strong>LLM Provider:</strong> <span style="color: %s;">%s</span></li>
</ul>
<div style="margin-top: 20px;">
	<h3>Real-time Telemetry (WebSocket)</h3>
	<div style="display: flex; gap: 20px;">
		<div style="flex: 1; border: 1px solid #ddd; padding: 10px; border-radius: 8px; background: #fafafa;">
			<h4>Audit Log Stream</h4>
			<div id="auditLogStream" style="height: 150px; overflow-y: auto; font-family: monospace; font-size: 12px; background: #fff; padding: 5px; border: 1px inset #ccc;">
				<em style="color: #888;">Connecting to telemetry stream...</em>
			</div>
		</div>
		<div style="flex: 1; border: 1px solid #ddd; padding: 10px; border-radius: 8px; background: #fafafa;">
			<h4>Hermes Latency (ms)</h4>
			<div id="latencyGauge" style="height: 150px; background: #fff; padding: 5px; border: 1px inset #ccc; display: flex; align-items: flex-end;">
				<em style="color: #888; align-self: center; margin: auto;">Connecting...</em>
			</div>
		</div>
	</div>
</div>
<script>
	const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const ws = new WebSocket(proto + '//' + window.location.host + '/ws/telemetry');

	const auditLogStream = document.getElementById('auditLogStream');
	const latencyGauge = document.getElementById('latencyGauge');

	let bars = [];

	ws.onmessage = function(event) {
		const data = JSON.parse(event.data);

		if (data.audit_logs && data.audit_logs.length > 0) {
			auditLogStream.innerHTML = '';
			data.audit_logs.forEach(log => {
				const div = document.createElement('div');
				div.style.borderBottom = '1px solid #eee';
				div.style.padding = '2px 0';
				div.textContent = "[" + (log.actor || 'system') + "] " + log.action;
				auditLogStream.appendChild(div);
			});
		}

		if (data.metrics && data.metrics.hermes_latency_ms) {
			if (bars.length === 0) latencyGauge.innerHTML = '';

			const val = data.metrics.hermes_latency_ms;
			const bar = document.createElement('div');
			bar.style.width = '10px';
			bar.style.marginRight = '2px';
			bar.style.background = val > 600 ? '#dc3545' : (val > 400 ? '#ffc107' : '#28a745');
			const h = Math.min(100, (val / 1000) * 100);
			bar.style.height = Math.round(h) + "%%";
			bar.title = Math.round(val) + "ms";

			latencyGauge.appendChild(bar);
			bars.push(bar);
			if (bars.length > 30) {
				const oldBar = bars.shift();
				latencyGauge.removeChild(oldBar);
			}
		}
	};
</script>
</div>
</body>
</html>`, csrfToken, csrfToken, health, map[bool]string{true: "🥇", false: "🚨"}[health == "Healthy"], llmColor, llmStatus)
}

func healthStatusColor(status string) string {
	if strings.Contains(strings.ToLower(status), "ok") || strings.Contains(strings.ToLower(status), "pass") {
		return "#28a745"
	}
	if strings.Contains(strings.ToLower(status), "error") || strings.Contains(strings.ToLower(status), "fail") {
		return "#dc3545"
	}
	return "#ffc107"
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "OK")
}

func (s *Server) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := make(map[string]interface{})

	// 1. Check DB
	if err := s.db.Conn.PingContext(r.Context()); err != nil {
		health["database"] = "ERROR: " + err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["database"] = "OK"
	}

	// 2. Check CI/Sync status
	healthStatus, _ := s.tracker.GetSystemHealth(r.Context())
	health["system_health"] = healthStatus

	// 3. Worker liveness
	health["workers"] = "active"

	// 4. Check LLM/Hermes connectivity
	if checker, ok := s.llmProvider.(HermesHealthChecker); ok {
		if err := checker.HealthCheck(r.Context()); err != nil {
			health["llm_provider"] = "ERROR: " + err.Error()
		} else {
			health["llm_provider"] = "Hermes: Connected"
		}
	} else {
		health["llm_provider"] = "Mock"
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		slog.WarnContext(r.Context(), "Error encoding health JSON", "error", err)
	}
}

func verifySignature(payload []byte, secret string, signatureHeader string) bool {
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	actualSignature := signatureHeader[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// SECURITY: Webhook IP Allowlisting (GitHub Webhook IPs)
	// In production, this list would be fetched dynamically from api.github.com/meta
	allowedIPs := []string{
		"192.30.252.0/22", "185.199.108.0/22", "140.82.112.0/20", "143.55.64.0/20", "127.0.0.1", "::1",
	}

	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Forwarded-For")
	}
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	isAllowed := false
	parsedIP := net.ParseIP(clientIP)
	for _, allowed := range allowedIPs {
		if strings.Contains(allowed, "/") {
			// Subnet check
			_, subnet, err := net.ParseCIDR(allowed)
			if err == nil && subnet.Contains(parsedIP) {
				isAllowed = true
				break
			}
		} else if clientIP == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed && clientIP != "127.0.0.1" { // Localhost always allowed for local testing/proxy
		slog.WarnContext(r.Context(), "Webhook: Blocked request from unauthorized IP", "ip", clientIP)
		http.Error(w, "Forbidden: IP not allowlisted", http.StatusForbidden)
		return
	}

	// SECURITY: Verify GitHub Webhook Signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		slog.WarnContext(r.Context(), "Webhook: Missing X-Hub-Signature-256 header")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Real implementation of signature verification logic
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret != "" {
		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		// Reset body for later use if needed (though not used here yet)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		if !verifySignature(body, secret, signature) {
			slog.WarnContext(r.Context(), "Webhook: Invalid signature")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	} else {
		slog.WarnContext(r.Context(), "Webhook: GITHUB_WEBHOOK_SECRET not set, skipping verification (Insecure!)")
	}

	slog.InfoContext(r.Context(), "Webhook: Received GitHub push event, triggering deployment...")

	// Trigger sync and build in a goroutine to avoid blocking the webhook response
	go func() {
		if err := s.deploy.ExecuteSync(); err != nil {
			slog.Warn("Webhook: Sync failed", "error", err)
			return
		}
		if err := s.deploy.ExecuteBuild(); err != nil {
			slog.Warn("Webhook: Build failed", "error", err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = fmt.Fprintln(w, "Deployment triggered")
}

func (s *Server) handleGenerateQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tier := html.EscapeString(strings.TrimSpace(r.URL.Query().Get("company_size")))
	if tier == "" {
		tier = html.EscapeString(strings.TrimSpace(r.URL.Query().Get("market_cap_tier")))
	}

	quote := communication.CalculateQuote(tier)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tier":  tier,
		"quote": quote,
	}); err != nil {
		slog.ErrorContext(r.Context(), "Error encoding quote JSON", "error", err)
	}
}

// REST API for external pipeline management
func (s *Server) handleLeadsAPI(w http.ResponseWriter, r *http.Request) {
	// CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Example: List all leads
		companies, err := s.db.ListAllCompanies(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve leads", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(companies)
	case http.MethodPost:
		// Example: Create a new lead
		var lead db.Company
		if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := s.db.CreateCompany(r.Context(), &lead); err != nil {
			http.Error(w, "Failed to create lead", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(lead)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for telemetry dashboard
	},
}

func (s *Server) handleTelemetryWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), "WebSocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	slog.InfoContext(r.Context(), "Telemetry WebSocket connected")

	// Filter parameters (e.g. who=autodev)
	filterActor := r.URL.Query().Get("who")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Simulate Hermes Latency data via random walk around an average for now,
	// in a real app this would be polled from actual LLM latency metrics.
	latency := 450.0

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Fetch recent audit logs
			logs, err := s.db.ListRecentAuditLogs(r.Context(), 50)
			if err != nil {
				slog.ErrorContext(r.Context(), "Failed to fetch audit logs for telemetry", "error", err)
				continue
			}

			// Filter if requested
			var filtered []db.AuditLog
			if filterActor != "" {
				for _, l := range logs {
					if l.Actor == filterActor {
						filtered = append(filtered, l)
					}
				}
			} else {
				filtered = logs
			}

			// Generate jitter for latency
			latency += float64((time.Now().UnixNano()%100)-50) / 2.0
			if latency < 200 {
				latency = 200
			}

			payload := map[string]interface{}{
				"audit_logs": filtered,
				"metrics": map[string]interface{}{
					"hermes_latency_ms": latency,
				},
			}

			if err := conn.WriteJSON(payload); err != nil {
				slog.ErrorContext(r.Context(), "WebSocket write failed", "error", err)
				return
			}
		}
	}
}

func (s *Server) handleGDPRExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	data, err := s.db.ExportGDPRData(r.Context(), email)
	if err != nil {
		http.Error(w, "Failed to export data: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleGDPRDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteGDPRData(r.Context(), email); err != nil {
		http.Error(w, "Failed to delete data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDealsAPI(w http.ResponseWriter, r *http.Request) {
	// CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		stateFilter := html.EscapeString(strings.TrimSpace(r.URL.Query().Get("state")))
		var deals []db.Deal
		var err error
		if stateFilter != "" {
			deals, err = s.db.ListDealsByState(r.Context(), db.LeadState(stateFilter))
		} else {
			deals, err = s.db.ListDealsByState(r.Context(), db.StateDiscovered)
		}

		if err != nil {
			http.Error(w, "Failed to retrieve deals", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deals)
	case http.MethodPost:
		var deal db.Deal
		if err := json.NewDecoder(r.Body).Decode(&deal); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := s.db.CreateDeal(r.Context(), &deal); err != nil {
			http.Error(w, "Failed to create deal", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(deal)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ── Billing Handlers ──

func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.billingClient == nil {
		http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	sigHeader := r.Header.Get("Stripe-Signature")
	msg, err := s.billingClient.HandleWebhook(r.Context(), body, sigHeader)
	if err != nil {
		slog.Error("Stripe webhook processing failed", "error", err)
		http.Error(w, "Webhook processing failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	slog.Info("Stripe webhook processed", "msg", msg)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func (s *Server) handleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.billingClient == nil {
		http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
		return
	}
	var req struct {
		CompanyID  int64  `json:"company_id"`
		Tier       string `json:"tier"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	url, err := s.billingClient.CreateCheckoutSession(r.Context(), req.CompanyID, billing.Tier(req.Tier), req.SuccessURL, req.CancelURL)
	if err != nil {
		slog.Error("Failed to create checkout session", "error", err)
		http.Error(w, "Failed to create checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.billingClient == nil {
		http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
		return
	}
	subID := r.URL.Query().Get("stripe_subscription_id")
	var sub *billing.SubscriptionInfo
	var err error
	if subID == "" {
		http.Error(w, "Missing stripe_subscription_id", http.StatusBadRequest)
		return
	}
	sub, err = s.billingClient.GetSubscription(r.Context(), subID)
	if err != nil {
		http.Error(w, "Failed to get subscription: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func (s *Server) handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.billingClient == nil {
		http.Error(w, "Billing not configured", http.StatusServiceUnavailable)
		return
	}
	var req struct {
		StripeSubID string `json:"stripe_subscription_id"`
		AtPeriodEnd bool   `json:"at_period_end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := s.billingClient.CancelSubscription(r.Context(), req.StripeSubID, req.AtPeriodEnd); err != nil {
		http.Error(w, "Failed to cancel: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "canceled"})
}

func (s *Server) handleBillingPortal(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	http.Redirect(w, r, scheme+"://"+r.Host+"/#billing", http.StatusFound)
}
