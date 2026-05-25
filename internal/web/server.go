package web

import (
	"fmt"
	"net/http"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
)

// Server handles web dashboard requests.
type Server struct {
	db     *db.DB
	deploy *deploy.Deployer
}

// NewServer creates a new Server instance.
func NewServer(database *db.DB, deployer *deploy.Deployer) *Server {
	return &Server{
		db:     database,
		deploy: deployer,
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/health", s.handleHealth)

	log.Printf("Web dashboard starting on %s", addr)
	return http.ListenAndServe(addr, mux)
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
					</td>
				</tr>`, d.ID, d.CompanyID, d.CurrentState, statusTitle, d.CurrentState, d.UpdatedAt.Format("2006-01-02 15:04:05"), d.ID)
	}

	fmt.Fprintf(w, `
			</table>

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
		</body>
		</html>`)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}
