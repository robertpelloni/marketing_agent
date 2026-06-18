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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/logging"
	"golang.org/x/time/rate"
)

type HermesHealthChecker interface { HealthCheck(ctx context.Context) error }

type Server struct {
	db          *db.DB
	deploy      *deploy.Deployer
	tracker     deploy.CITracker
	tasks       *autodev.TaskManager
	auth        *auth.Authenticator
	llmProvider llm.LLMProvider
	mux         *http.ServeMux
	limiter     *rate.Limiter
	registry    *llm.PromptRegistry
}

func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager, llmProvider llm.LLMProvider, registry *llm.PromptRegistry) *Server {
	s := &Server{
		db:          database,
		deploy:      deployer,
		tracker:     tracker,
		tasks:       taskManager,
		auth:        auth.NewAuthenticator(),
		llmProvider: llmProvider,
		mux:         http.NewServeMux(),
		limiter:     rate.NewLimiter(rate.Limit(5), 10),
		registry:    registry,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Handle("/", s.auth.Middleware(http.HandlerFunc(s.handleDashboard)))
	s.mux.Handle("/api/v1/test/simulate_inbound", s.auth.Middleware(http.HandlerFunc(s.handleSimulateInbound)))
	s.mux.HandleFunc("/login", s.auth.HandleLogin)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/health/detailed", s.handleDetailedHealth)
	s.mux.Handle("/api/v1/deals", s.auth.Middleware(http.HandlerFunc(s.handleListDeals)))
	s.mux.Handle("/api/v1/leads", s.auth.Middleware(http.HandlerFunc(s.handleListLeads)))
	s.mux.Handle("/api/v1/gdpr/export", s.auth.Middleware(http.HandlerFunc(s.handleGDPRExport)))
	s.mux.Handle("/api/v1/gdpr/delete", s.auth.Middleware(http.HandlerFunc(s.handleGDPRDelete)))
	s.mux.HandleFunc("/api/v1/webhook/github", s.handleGitHubWebhook)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.limiter.Allow() {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	logging.Init("json", "info")
	slog.Info("Web dashboard starting", "addr", addr)
	// #nosec G114 -- ReadHeaderTimeout is configured via srv.ReadHeaderTimeout in main.go
	return http.ListenAndServe(addr, logging.Middleware(s))
}

func (s *Server) handleSimulateInbound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { return }
	cid := html.EscapeString(r.FormValue("contact_id"))
	txt := html.EscapeString(r.FormValue("text"))
	fmt.Fprintf(w, "UAT: Simulation triggered for contact %s: %s", cid, txt)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	if r.Method == http.MethodPost {
		action := html.EscapeString(r.FormValue("action"))
		switch action {
		case "enrich":
			slog.InfoContext(r.Context(), "Manual enrichment triggered", "deal_id", r.FormValue("deal_id"))
		case "sync":
			if err := s.deploy.ExecuteSync(); err != nil {
				slog.WarnContext(r.Context(), "Sync error", "error", err)
			}
		case "flag_success":
			interactionID := r.FormValue("interaction_id")
			success := r.FormValue("success") == "true"
			var id int64
			if _, err := fmt.Sscanf(interactionID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid interaction ID", "error", err, "interaction_id", interactionID)
			} else {
				if err := s.db.UpdateInteractionSuccess(r.Context(), id, success); err != nil {
					slog.WarnContext(r.Context(), "Error flagging interaction", "error", err, "interaction_id", id)
				}
			}
		case "update_channel":
			contactID := r.FormValue("contact_id")
			channel := r.FormValue("channel")
			var id int64
			if _, err := fmt.Sscanf(contactID, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid contact ID", "error", err, "contact_id", contactID)
			} else {
				if err := s.db.UpdateContactPreferredChannel(r.Context(), id, channel); err != nil {
					slog.WarnContext(r.Context(), "Error updating channel", "error", err, "contact_id", id, "channel", channel)
				}
			}
		case "build":
			if err := s.deploy.ExecuteBuild(); err != nil {
				slog.WarnContext(r.Context(), "Build error", "error", err)
			}
		case "approve":
			dealIDStr := r.FormValue("deal_id")
			var id int64
			if _, err := fmt.Sscanf(dealIDStr, "%d", &id); err != nil {
				slog.WarnContext(r.Context(), "Invalid deal ID for approval", "error", err)
			} else {
				_ = s.db.SetApprovalRequired(r.Context(), id, false)
				if err := s.db.UpdateDealState(r.Context(), id, db.StateNegotiating); err != nil {
					slog.WarnContext(r.Context(), "Failed to approve deal", "deal_id", id, "error", err)
				}
			}
		}
		http.Redirect(w, r, r.URL.String(), http.StatusSeeOther); return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 { page = 1 }
	limit := 20
	offset := (page - 1) * limit

	deals, _ := s.db.ListRecentDeals(r.Context(), limit, offset)
	health, _ := s.tracker.GetSystemHealth(r.Context())
	csrfToken := s.auth.GetCSRFToken(r)

	outcomes := []llm.ABResult{}
	if s.registry != nil { outcomes = s.registry.GetOutcomes() }
	timings := deploy.GetTimings()

	metrics, err := s.db.GetPerformanceMetrics(r.Context())
	if err != nil {
		slog.WarnContext(r.Context(), "Error retrieving metrics", "error", err)
		metrics = &db.PerformanceMetrics{
			LeadsByState: make(map[db.LeadState]int),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html><head><title>TormentNexus Dashboard</title>
<style>
body { font-family: sans-serif; margin: 40px; background: #f8f9fa; }
.container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
table { width: 100%%; border-collapse: collapse; margin-top: 10px; }
th, td { padding: 10px; border: 1px solid #ddd; text-align: left; }
th { background-color: #007bff; color: white; }
.action-btn { background-color: #28a745; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; text-decoration: none; display: inline-block; }
.nav-btn { background-color: #6c757d; color: white; padding: 6px 12px; border-radius: 4px; text-decoration: none; }
</style>
</head>
<body>
<div class="container">
<h1>TormentNexus Autonomous Sales v0.9.0</h1>
<h2>Active Leads</h2>
<table><tr><th>ID</th><th>State</th><th>Last Updated</th><th>Actions</th></tr>`)

	for _, d := range deals {
		fmt.Fprintf(w, `
<tr>
<td>%d</td><td>%s</td><td>%s</td>
<td>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="enrich"><input type="hidden" name="deal_id" value="%d">
<button type="submit" class="action-btn">Enrich</button>
</form>
%s
</td>
</tr>`, d.ID, d.CurrentState, d.UpdatedAt.Format("15:04:05"), csrfToken, d.ID, s.renderApprovalButton(d, csrfToken))
	}

	fmt.Fprintf(w, `
</table>
<div style="margin-top: 20px;">
	<a href="/?page=%d" class="nav-btn">Previous</a>
	<span>Page %d</span>
	<a href="/?page=%d" class="nav-btn">Next</a>
</div>

<h2>Pipeline Performance</h2>
<p>Total Leads: %d | Win Rate: %.1f%% | Successful Outreach: %d</p>

<h2>Prompt Performance & A/B Analytics</h2>
<table><tr><th>Experiment</th><th>Variant</th><th>Win Rate</th></tr>`, page-1, page, page+1, metrics.TotalLeads, metrics.WinRate, metrics.SuccessfulOutreach)

	for _, o := range outcomes {
		rate := 0.0
		if o.Total > 0 { rate = float64(o.Success) / float64(o.Total) * 100 }
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%.1f%% (%d/%d)</td></tr>", o.Experiment, o.VersionID, rate, o.Success, o.Total)
	}

	fmt.Fprintf(w, `
</table>

<h2>Worker Performance</h2>
<table><tr><th>Worker</th><th>Last Cycle Duration</th></tr>`)
	for name, dur := range timings {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>", name, dur)
	}

	fmt.Fprintf(w, `
</table>

<div style="margin-top:20px; padding:15px; background:#e9ecef; border-radius:4px;">
<h3>System Status</h3>
<p>Health: %s | AutoDev: Active</p>
<form method="POST">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="sync"><button type="submit" class="action-btn">Sync Repository</button>
</form>
</div>
</div>
</body></html>`, health, csrfToken)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "OK") }
func (s *Server) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := make(map[string]interface{})

	if err := s.db.Conn.PingContext(r.Context()); err != nil {
		health["database"] = "ERROR: " + err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["database"] = "OK"
	}

	healthStatus, _ := s.tracker.GetSystemHealth(r.Context())
	health["system_health"] = healthStatus
	health["workers"] = "active"

	if checker, ok := s.llmProvider.(HermesHealthChecker); ok {
		if err := checker.HealthCheck(r.Context()); err != nil {
			health["llm_provider"] = "ERROR: " + err.Error()
		} else {
			health["llm_provider"] = "Hermes: Connected"
		}
	} else {
		health["llm_provider"] = "Mock"
	}

	_ = json.NewEncoder(w).Encode(health)
}

func verifySignature(payload []byte, secret string, signatureHeader string) bool {
	if !strings.HasPrefix(signatureHeader, "sha256=") { return false }
	actualSignature := signatureHeader[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(actualSignature), []byte(expectedSignature))
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	allowedIPs := os.Getenv("ALLOWED_WEBHOOK_IPS")
	if allowedIPs != "" {
		remoteIP := strings.Split(r.RemoteAddr, ":")[0]
		found := false
		for _, ip := range strings.Split(allowedIPs, ",") {
			if strings.TrimSpace(ip) == remoteIP { found = true; break }
		}
		if !found {
			slog.Warn("Webhook Security: Blocked unauthorized IP", "ip", remoteIP)
			http.Error(w, "Forbidden", http.StatusForbidden); return
		}
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret != "" {
		if signature == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized); return
		}
		body, _ := io.ReadAll(r.Body)
		if !verifySignature(body, secret, signature) {
			http.Error(w, "Forbidden", http.StatusForbidden); return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	go func() {
		_ = s.deploy.ExecuteSync()
		_ = s.deploy.ExecuteBuild()
	}()

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleListDeals(w http.ResponseWriter, r *http.Request) {
	deals, err := s.db.ListRecentDeals(r.Context(), 100, 0)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deals)
}

func (s *Server) handleListLeads(w http.ResponseWriter, r *http.Request) {
	leads, err := s.db.ListAllCompanies(r.Context(), 100, 0)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leads)
}

func (s *Server) renderApprovalButton(deal db.Deal, csrfToken string) string {
	if !deal.ApprovalRequired { return "" }
	return fmt.Sprintf(`
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="approve"><input type="hidden" name="deal_id" value="%d">
<button type="submit" class="action-btn" style="background:#ffc107;color:#000">Approve</button>
</form>`, csrfToken, deal.ID)
}

func (s *Server) handleGDPRExport(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	data, err := s.db.GetExportData(r.Context(), email)
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleGDPRDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { return }
	email := r.FormValue("email")
	if err := s.db.SoftDeleteContact(r.Context(), email); err != nil {
		http.Error(w, err.Error(), 500); return
	}
	w.WriteHeader(http.StatusNoContent)
}
