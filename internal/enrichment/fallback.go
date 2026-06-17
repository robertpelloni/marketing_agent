package enrichment

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type FallbackSource struct {
	mu      sync.RWMutex
	sources []EnrichmentSource
	names   []string
}

func NewFallbackSource(sources []EnrichmentSource, names []string) *FallbackSource {
	resolvedNames := make([]string, len(sources))
	for i := range sources {
		if i < len(names) && names[i] != "" {
			resolvedNames[i] = names[i]
		} else {
			resolvedNames[i] = fmt.Sprintf("Source%d", i+1)
		}
	}
	return &FallbackSource{
		sources: sources,
		names:   resolvedNames,
	}
}

func (f *FallbackSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for i, source := range f.sources {
		name := f.names[i]
		slog.Info("FallbackSource: Trying source", "attempt", i+1, "total", len(f.sources), "source", name, "company", company.Name)

		contacts, err := source.Enrich(ctx, company)
		if err != nil {
			slog.Warn("FallbackSource: source failed", "source", name, "error", err)
			continue
		}

		if len(contacts) > 0 {
			slog.Info("FallbackSource: source succeeded", "source", name, "count", len(contacts))
			return contacts, nil
		}
	}

	return nil, nil
}

func (f *FallbackSource) HealthCheck(ctx context.Context) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for i, source := range f.sources {
		if err := source.HealthCheck(ctx); err != nil {
			return fmt.Errorf("fallback source health check failed at %s: %w", f.names[i], err)
		}
	}
	return nil
}

func (f *FallbackSource) Status() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("FallbackSource: %d source(s) configured\n", len(f.sources)))
	for i, name := range f.names {
		b.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, name))
	}
	return b.String()
}
