package web
import (
	"fmt"

	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robertpelloni/enterprise_sales_bot/internal/auth"
	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"golang.org/x/time/rate"
)
type Server struct {
	db *db.DB; deploy *deploy.Deployer; tracker deploy.CITracker; tasks *autodev.TaskManager; comm *communication.Manager; auth *auth.Authenticator; crm crm.CRMClient; provider string; limiter *rate.Limiter; mux *http.ServeMux
}
func NewServer(database *db.DB, deployer *deploy.Deployer, tracker deploy.CITracker, taskManager *autodev.TaskManager, crmClient crm.CRMClient, commManager *communication.Manager, provider string) *Server {
	s := &Server{db: database, deploy: deployer, tracker: tracker, tasks: taskManager, comm: commManager, auth: auth.NewAuthenticator(), crm: crmClient, provider: provider, limiter: rate.NewLimiter(5, 10), mux: http.NewServeMux()}
	s.routes(); return s
}
func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleDashboard); s.mux.HandleFunc("/login", s.auth.HandleLogin); s.mux.HandleFunc("/health", s.handleHealth); s.mux.Handle("/metrics", promhttp.Handler()); s.mux.HandleFunc("/api/v1/test/simulate_inbound", s.handleSimulateInbound)
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.limiter.Allow() { http.Error(w, "Rate limit", 429); return }
	s.auth.Middleware(s.mux).ServeHTTP(w, r)
}
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "<html><body><h1>Dashboard v0.4.8</h1></body></html>") }
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "OK") }
func (s *Server) handleSimulateInbound(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email"); text := r.FormValue("text")
	contact, _ := s.db.GetContactByEmail(r.Context(), email)
	if contact == nil { http.Error(w, "Not found", 404); return }
	reply, _ := s.comm.ProcessInbound(r.Context(), *contact, text)
	fmt.Fprintf(w, "Autonomous reply: %s", reply)
}
