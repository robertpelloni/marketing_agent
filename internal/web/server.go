package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

// Server handles web dashboard requests.
type Server struct {
	db      *db.DB
	deploy  *deploy.Deployer
	tracker deploy.CITracker
	tasks   *autodev.TaskManager
}

// NewServer creates a new Server instance.
func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager) *Server {
	return &Server{
		db:      database,
		deploy:  deployer,
		tracker: tracker,
		tasks:   taskManager,
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/health/detailed", s.handleDetailedHealth)
	mux.HandleFunc("/api/v1/webhook/github", s.handleGitHubWebhook)
	mux.ServeHTTP(w, r)
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Web dashboard starting on %s", addr)
	return http.ListenAndServe(addr, s)
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
		case "build":
			if err := s.deploy.ExecuteBuild(); err != nil {
				log.Printf("UI: Build error: %v", err)
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
		log.Printf("UI: Error retrieving metrics: %v", err)
		metrics = &db.PerformanceMetrics{
			LeadsByState: make(map[db.LeadState]int),
		}
	}

	prs, err := s.db.ListActivePullRequests(r.Context())
	if err != nil {
		log.Printf("UI: Error listing PRs: %v", err)
	}

	taskList, _ := s.tasks.ListAllTasks(r.Context())

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
				.status-Discovered { background-color: #e2e3e5; color: #383d41; }
				.status-Researched { background-color: #cce5ff; color: #004085; }
				.status-PR { background-color: #fff3cd; color: #856404; }
				.action-btn { background-color: #28a745; color: white; border: none; padding: 6px 12px; border-radius: 4px; cursor: pointer; }
				.action-btn:hover { background-color: #218838; }
				.deploy-section { margin-top: 30px; padding: 20px; border: 1px solid #ccc; border-radius: 8px; background: #fff; }
				.deploy-btn { background-color: #007bff; margin-right: 10px; }
			</style>
		</head>
		<body>
			<h1>Sales Bot Lead Dashboard</h1>
			<p>Total Recent Leads: %d</p>
			<table>
				<tr>
					<th>Deal ID</th>
					<th>Company ID</th>
					<th>State</th>
					<th>Last Updated</th>
					<th>Actions</th>
				</tr>`, len(deals))

	for _, d := range deals {
		statusTitle := ""
		switch d.CurrentState {
		case db.StateDiscovered:
			statusTitle = "Company identified, awaiting technical research."
		case db.StateResearched:
			statusTitle = "Key engineering contacts found and technical dossier compiled."
		}

		// Retrieve latest interaction ID to allow manual flagging from UI
		contacts, _ := s.db.ListContactsByCompany(r.Context(), d.CompanyID)
		latestInteractionID := int64(0)
		if len(contacts) > 0 {
			interactions, _ := s.db.ListInteractionsByContact(r.Context(), contacts[0].ID)
			if len(interactions) > 0 {
				latestInteractionID = interactions[0].ID
			}
		}

		fmt.Fprintf(w, `
				<tr>
					<td>%d</td>
					<td>%d</td>
					<td><span class="status status-%s" title="%s">%s</span></td>
					<td>%s</td>
					<td>
						<form method="POST" style="display:inline;">
							<input type="hidden" name="action" value="enrich">
							<input type="hidden" name="deal_id" value="%d">
							<button type="submit" class="action-btn">Trigger Enrichment</button>
						</form>
						<form method="POST" style="display:inline;">
							<input type="hidden" name="action" value="flag_success">
							<input type="hidden" name="interaction_id" value="%d">
							<input type="hidden" name="success" value="true">
							<button type="submit" class="action-btn" style="background-color: #6f42c1;">Flag Success</button>
						</form>
					</td>
				</tr>`, d.ID, d.CompanyID, d.CurrentState, statusTitle, d.CurrentState, d.UpdatedAt.Format("2006-01-02 15:04:05"), d.ID, latestInteractionID)
	}

	fmt.Fprintf(w, `
			</table>

			<div class="deploy-section" style="border-top: 5px solid #17a2b8;">
				<h2>Performance Metrics</h2>
				<p>Real-time pipeline statistics and conversion rates.</p>
				<div style="display: flex; gap: 20px; flex-wrap: wrap;">
					<div style="background: #e9ecef; padding: 15px; border-radius: 8px; min-width: 150px;">
						<strong>Total Leads:</strong> %d
					</div>
					<div style="background: #d4edda; padding: 15px; border-radius: 8px; min-width: 150px;">
						<strong>Won Deals:</strong> %d
					</div>
					<div style="background: #f8d7da; padding: 15px; border-radius: 8px; min-width: 150px;">
						<strong>Win Rate:</strong> %.1f%%
					</div>
					<div style="background: #fff3cd; padding: 15px; border-radius: 8px; min-width: 150px;">
						<strong>Successful Outreach:</strong> %d
					</div>
				</div>
				<h3>Leads by State</h3>
				<ul>
					<li><strong>Discovered:</strong> %d</li>
					<li><strong>Researched:</strong> %d</li>
					<li><strong>Outreach Sent:</strong> %d</li>
					<li><strong>Engaged:</strong> %d</li>
					<li><strong>Negotiating:</strong> %d</li>
				</ul>
			</div>`,
		metrics.TotalLeads, metrics.LeadsByState[db.StateClosedWon], metrics.WinRate, metrics.SuccessfulOutreach,
		metrics.LeadsByState[db.StateDiscovered], metrics.LeadsByState[db.StateResearched], metrics.LeadsByState[db.StateOutreachSent],
		metrics.LeadsByState[db.StateEngaged], metrics.LeadsByState[db.StateNegotiating])

	fmt.Fprintf(w, `
			<div class="deploy-section">
				<h2>Autonomous Task Board</h2>
				<p>Prioritized development roadmap and execution status.</p>
				<table>
					<tr>
						<th>Description</th>
						<th>Status</th>
					</tr>`)
	for _, t := range taskList {
		status := "Pending"
		if t.Completed {
			status = "Completed"
		}
		fmt.Fprintf(w, `
					<tr>
						<td>%s</td>
						<td><span class="status status-%s">%s</span></td>
					</tr>`, html.EscapeString(t.Description), status, status)
	}
	fmt.Fprintf(w, `
				</table>
			</div>

			<div class="deploy-section">
				<h2>Autonomous Pull Requests</h2>
				<p>Active feature branches and automated merge status.</p>
				<table>
					<tr>
						<th>PR ID</th>
						<th>Branch</th>
						<th>Title</th>
						<th>Status</th>
						</tr>`)
	for _, pr := range prs {
		fmt.Fprintf(w, `
					<tr>
						<td>%s</td>
						<td>%s</td>
						<td>%s</td>
						<td><span class="status status-PR">%s</span></td>
					</tr>`, html.EscapeString(pr.ID), html.EscapeString(pr.Branch), html.EscapeString(pr.Title), html.EscapeString(string(pr.Status)))
	}

	fmt.Fprintf(w, `
				</table>
			</div>

			<div class="deploy-section">
				<h2>Self-Service Deployment</h2>
				<p>Manage repository state and trigger system builds autonomously.</p>
				<form method="POST" style="display:inline;">
					<input type="hidden" name="action" value="sync">
					<button type="submit" class="action-btn deploy-btn">Sync Repository</button>
				</form>
				<form method="POST" style="display:inline;">
					<input type="hidden" name="action" value="build">
					<button type="submit" class="action-btn deploy-btn" style="background-color: #6c757d;">Trigger Build</button>
				</form>
			</div>

			<div class="deploy-section" style="border-left: 5px solid #28a745;">
				<h2>System Health & CI Status</h2>
				<p>Real-time monitoring of the autonomous deployment pipeline.</p>
				<ul>
					<li><strong>Global Health:</strong> <span style="color: #28a745;">%s</span></li>
				</ul>
			</div>
		</body>
				</html>`, health)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
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

	// 3. Worker liveness (Simulated)
	health["workers"] = "active"

	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Web: Error encoding health JSON: %v", err)
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

	// SECURITY: Verify GitHub Webhook Signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		log.Println("Webhook Security: Missing X-Hub-Signature-256 header")
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
			log.Println("Webhook Security: Invalid signature")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	} else {
		log.Println("Webhook Security: GITHUB_WEBHOOK_SECRET not set, skipping verification (Insecure!)")
	}

	log.Println("Webhook: Received GitHub push event, triggering deployment...")

	// Trigger sync and build in a goroutine to avoid blocking the webhook response
	go func() {
		if err := s.deploy.ExecuteSync(); err != nil {
			log.Printf("Webhook: Sync failed: %v", err)
			return
		}
		if err := s.deploy.ExecuteBuild(); err != nil {
			log.Printf("Webhook: Build failed: %v", err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Deployment triggered")
}
