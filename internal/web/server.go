package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"

	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
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
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.limiter.Allow() {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Web dashboard starting on %s", addr)
	return http.ListenAndServe(addr, s)
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
		case "enrich": log.Printf("UI: Enrich deal %s", r.FormValue("deal_id"))
		case "sync": _ = s.deploy.ExecuteSync()
		case "approve":
			var id int64
			_, _ = fmt.Sscanf(r.FormValue("deal_id"), "%d", &id)
			_ = s.db.SetApprovalRequired(r.Context(), id, false)
		case "build": _ = s.deploy.ExecuteBuild()
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
<h1>TormentNexus Autonomous Sales v0.7.0</h1>
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

<h2>Prompt Performance & A/B Analytics</h2>
<table><tr><th>Experiment</th><th>Variant</th><th>Win Rate</th></tr>`, page-1, page, page+1)

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
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "OK"})
}

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
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
