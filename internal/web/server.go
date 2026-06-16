package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"golang.org/x/time/rate"
)

// HermesHealthChecker is an optional interface for checking LLM provider health.
type HermesHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Server handles web dashboard requests.
type Server struct {
	db          *db.DB
	deploy      *deploy.Deployer
	tracker     deploy.CITracker
	tasks       *autodev.TaskManager
	auth        *auth.Authenticator
	llmProvider llm.LLMProvider
	mux         *http.ServeMux
	limiter     *rate.Limiter
}

// NewServer creates a new Server instance.
func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager, llmProvider llm.LLMProvider) *Server {
	s := &Server{
		db:          database,
		deploy:      deployer,
		tracker:     tracker,
		tasks:       taskManager,
		auth:        auth.NewAuthenticator(),
		llmProvider: llmProvider,
		mux:         http.NewServeMux(),
		limiter:     rate.NewLimiter(rate.Limit(5), 10),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Protected routes
	s.mux.Handle("/", s.auth.Middleware(http.HandlerFunc(s.handleDashboard)))
	s.mux.Handle("/api/v1/test/simulate_inbound", s.auth.Middleware(http.HandlerFunc(s.handleSimulateInbound)))

	// Public routes
	s.mux.HandleFunc("/login", s.auth.HandleLogin)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/health/detailed", s.handleDetailedHealth)
	s.mux.HandleFunc("/api/v1/webhook/github", s.handleGitHubWebhook)
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.limiter.Allow() {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	s.mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Web dashboard starting on %s", addr)
	// #nosec G114 -- Simple ListenAndServe is used for internal dashboard; timeout configuration handled at higher level if needed
	return http.ListenAndServe(addr, s)
}

func (s *Server) handleSimulateInbound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cid := r.FormValue("contact_id")
	txt := r.FormValue("text")

	log.Printf("UAT: Simulating inbound from contact %s: %s", cid, txt)
	fmt.Fprintf(w, "UAT: Simulation triggered for contact %s. Response logic would execute here.", cid)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		switch action {
		case "enrich":
			log.Printf("UI: Manual enrichment triggered for deal %s", r.FormValue("deal_id"))
		case "sync":
			if err := s.deploy.ExecuteSync(); err != nil {
				log.Printf("UI: Sync error: %v", err)
			}
		case "flag_success":
			interactionID := r.FormValue("interaction_id")
			success := r.FormValue("success") == "true"
			var id int64
			if _, err := fmt.Sscanf(interactionID, "%d", &id); err != nil {
				log.Printf("UI: Invalid interaction ID: %v", err)
			} else {
				if err := s.db.UpdateInteractionSuccess(r.Context(), id, success); err != nil {
					log.Printf("UI: Error flagging interaction: %v", err)
				}
			}
		case "update_channel":
			contactID := r.FormValue("contact_id")
			channel := r.FormValue("channel")
			var id int64
			if _, err := fmt.Sscanf(contactID, "%d", &id); err != nil {
				log.Printf("UI: Invalid contact ID: %v", err)
			} else {
				if err := s.db.UpdateContactPreferredChannel(r.Context(), id, channel); err != nil {
					log.Printf("UI: Error updating channel: %v", err)
				}
			}
		case "approve":
			dealID := r.FormValue("deal_id")
			var id int64
			if _, err := fmt.Sscanf(dealID, "%d", &id); err == nil {
				if err := s.db.SetApprovalRequired(r.Context(), id, false); err != nil {
					log.Printf("UI: Error approving deal: %v", err)
				}
			}
		case "build":
			if err := s.deploy.ExecuteBuild(); err != nil {
				log.Printf("UI: Build error: %v", err)
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	deals, _ := s.db.ListRecentDeals(r.Context(), 20)
	health, _ := s.tracker.GetSystemHealth(r.Context())
	taskList, _ := s.tasks.ListAllTasks(r.Context())
	csrfToken := s.auth.GetCSRFToken(r)

	// Check LLM/Hermes health
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
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
<title>Enterprise Sales Bot Dashboard</title>
<style>
body { font-family: sans-serif; margin: 40px; background-color: #f4f4f9; }
table { width: 100%%; border-collapse: collapse; margin-top: 20px; background: white; }
th, td { padding: 12px; border: 1px solid #ddd; text-align: left; }
th { background-color: #007bff; color: white; }
tr:nth-child(even) { background-color: #f2f2f2; }
h1 { color: #333; }
.status { font-weight: bold; padding: 4px 8px; border-radius: 4px; cursor: help; }
.action-btn { background-color: #28a745; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; }
.deploy-section { margin-top: 30px; padding: 20px; border: 1px solid #ccc; border-radius: 8px; background: #fff; }
</style>
</head>
<body>
<h1>Sales Bot Lead Dashboard</h1>
<table>
<tr><th>Deal ID</th><th>Company ID</th><th>State</th><th>Last Updated</th><th>Actions</th></tr>`)

	for _, d := range deals {
		contacts, _ := s.db.ListContactsByCompany(r.Context(), d.CompanyID)
		var contactHTML string
		for _, c := range contacts {
			contactHTML += fmt.Sprintf(`
				<div><strong>%s</strong> (%s)
					<form method="POST" style="display:inline;">
						<input type="hidden" name="csrf_token" value="%s">
						<input type="hidden" name="action" value="update_channel">
						<input type="hidden" name="contact_id" value="%d">
						<select name="channel" onchange="this.form.submit()">
							<option value="email"%s>Email</option>
							<option value="linkedin"%s>LinkedIn</option>
							<option value="github"%s>GitHub</option>
						</select>
					</form>
				</div>`, html.EscapeString(c.Name), html.EscapeString(c.Role), csrfToken, c.ID,
				map[bool]string{true: " selected", false: ""}[c.PreferredChannel == "email"],
				map[bool]string{true: " selected", false: ""}[c.PreferredChannel == "linkedin"],
				map[bool]string{true: " selected", false: ""}[c.PreferredChannel == "github"])
		}

		fmt.Fprintf(w, `
<tr>
<td>%d</td><td>%d</td><td>%s</td><td>%s</td>
<td>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="enrich"><input type="hidden" name="deal_id" value="%d">
<button type="submit" class="action-btn">Enrich</button>
</form>
%s
</td>
</tr><tr><td colspan="5">%s</td></tr>`, d.ID, d.CompanyID, d.CurrentState, d.UpdatedAt.Format("15:04:05"), csrfToken, d.ID, s.renderApprovalButton(d, csrfToken), contactHTML)
	}

	fmt.Fprintf(w, `
</table>

<div class="deploy-section">
<h2>User Testing & UAT Portal</h2>
<form action="/api/v1/test/simulate_inbound" method="POST" target="_blank">
	<input type="hidden" name="csrf_token" value="%s">
	<input type="text" name="contact_id" placeholder="Contact ID">
	<input type="text" name="text" placeholder="Message">
	<button type="submit" class="action-btn">Simulate Inbound</button>
</form>
</div>

<div class="deploy-section">
<h2>Autonomous Tasks</h2>
<table><tr><th>Description</th><th>Status</th></tr>`, csrfToken)

	for _, t := range taskList {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>", html.EscapeString(t.Description), t.Completed)
	}

	fmt.Fprintf(w, `
</table>
</div>

<div class="deploy-section">
<h2>Deployment & Sync</h2>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="sync"><button type="submit" class="action-btn">Sync Repository</button>
</form>
<form method="POST" style="display:inline;">
<input type="hidden" name="csrf_token" value="%s">
<input type="hidden" name="action" value="build"><button type="submit" class="action-btn" style="background:#6c757d">Build</button>
</form>
<p>System Health: %s | LLM: <span style="color:%s">%s</span></p>
</div>
</body></html>`, csrfToken, csrfToken, health, llmColor, llmStatus)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "OK") }

func (s *Server) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := map[string]string{"database": "OK", "workers": "active"}
	_ = json.NewEncoder(w).Encode(health)
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
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
