package metrics
import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)
var (
	LeadsDiscovered = promauto.NewCounter(prometheus.CounterOpts{Name: "sales_bot_leads_discovered_total", Help: "Total leads discovered"})
	InteractionsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{Name: "sales_bot_interactions_processed_total", Help: "Total interactions"}, []string{"direction", "channel"})
	DealsWon = promauto.NewCounter(prometheus.CounterOpts{Name: "sales_bot_deals_won_total", Help: "Total deals won"})
)
