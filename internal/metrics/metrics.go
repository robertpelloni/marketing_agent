package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PromptImpressions counts the number of times a prompt variant is used.
	PromptImpressions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sales_bot_prompt_impressions_total",
		Help: "The total number of prompt impressions, labeled by variant",
	}, []string{"variant"})

	// PromptConversions counts the number of successful meetings booked per variant.
	PromptConversions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sales_bot_prompt_conversions_total",
		Help: "The total number of prompt conversions, labeled by variant",
	}, []string{"variant"})
)

// RecordImpression records an impression for a prompt variant.
func RecordImpression(variant string) {
	PromptImpressions.WithLabelValues(variant).Inc()
}

// RecordConversion records a successful conversion for a prompt variant.
func RecordConversion(variant string) {
	PromptConversions.WithLabelValues(variant).Inc()
}
