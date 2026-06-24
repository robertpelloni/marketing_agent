package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// PromptMetricsTracker monitors the conversion success of different A/B prompt variants over time.
type PromptMetricsTracker struct {
	db *db.DB
	mu sync.Mutex
}

// NewPromptMetricsTracker initializes a new tracker for evaluating LLM generation quality.
func NewPromptMetricsTracker(database *db.DB) *PromptMetricsTracker {
	return &PromptMetricsTracker{
		db: database,
	}
}

// TrackImpression records that a specific variant was used to generate outreach.
func (p *PromptMetricsTracker) TrackImpression(ctx context.Context, variant string, templateID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// In a full implementation, this would write to a dedicated prompt_analytics table.
	// For now, we utilize the application logs to trace variant performance distributions.
	slog.Info(fmt.Sprintf("PromptAnalytics: Impression Recorded | Variant: %s | Template: %s | Timestamp: %s",
		variant, templateID, time.Now().Format(time.RFC3339)))

	return nil
}

// TrackConversion correlates a Closed_Won deal back to the original prompt variant used to secure the meeting.
func (p *PromptMetricsTracker) TrackConversion(ctx context.Context, dealID int64, variant string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	slog.Info(fmt.Sprintf("PromptAnalytics: CONVERSION Recorded! | Deal: %d | Winning Variant: %s",
		dealID, variant))

	return nil
}
