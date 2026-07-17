package enrichment

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// FallbackSource wraps multiple EnrichmentSource instances and tries each one
// in order until one returns contacts or all sources are exhausted. It provides
// structured logging and status reporting for observability.
type FallbackSource struct {
	mu      sync.RWMutex
	sources []EnrichmentSource
	names   []string // human-readable names for each source (for logging/status)
}

// sourceResult captures the outcome of a single enrichment source attempt.
type sourceResult struct {
	Name    string
	Success bool
	Count   int
	Error   string
}

// FallbackReport contains the result of a full fallback chain attempt.
type FallbackReport struct {
	CompanyName   string
	CompanyDomain string
	SourceResults []sourceResult
	FinalSuccess  bool
	TotalContacts int
}

// NewFallbackSource creates a FallbackSource from an ordered list of
// EnrichmentSource instances. Each source can be optionally named via
// the names parameter; if omitted or shorter than sources, generic names
// are generated (Source1, Source2, ...).
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

// Enrich implements EnrichmentSource. It tries each wrapped source in order
// until one returns contacts without error. Logs each attempt with clear
// pass/fail indicators.
func (f *FallbackSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Skip companies with obviously garbage domains/names to avoid
	// wasting API calls on HN comment parsing artifacts.
	if isGarbageLead(company) {
		slog.Info(fmt.Sprintf("FallbackSource: Skipping garbage lead %q (domain: %q)", company.Name, company.Domain))
		return nil, nil
	}

	report := &FallbackReport{
		CompanyName:   company.Name,
		CompanyDomain: company.Domain,
	}

	for i, source := range f.sources {
		name := f.names[i]

		select {
		case <-ctx.Done():
			slog.Info(fmt.Sprintf("FallbackSource: Context cancelled at source %d/%d (%s) for %s",
				i+1, len(f.sources), name, company.Domain))
			return nil, ctx.Err()
		default:
		}

		slog.Info(fmt.Sprintf("FallbackSource: Trying source %d/%d — %s for %s (%s)",
			i+1, len(f.sources), name, company.Name, company.Domain))

		contacts, err := source.Enrich(ctx, company)

		result := sourceResult{Name: name}
		if err != nil {
			result.Error = err.Error()
			slog.Info(fmt.Sprintf("FallbackSource: ✗ %s failed: %v", name, err))
			report.SourceResults = append(report.SourceResults, result)
			continue
		}

		if len(contacts) == 0 {
			slog.Info(fmt.Sprintf("FallbackSource: ✗ %s returned no contacts for %s", name, company.Domain))
			report.SourceResults = append(report.SourceResults, result)
			continue
		}

		// Success!
		result.Success = true
		result.Count = len(contacts)
		report.SourceResults = append(report.SourceResults, result)
		report.FinalSuccess = true
		report.TotalContacts = len(contacts)

		slog.Info(fmt.Sprintf("FallbackSource: ✓ %s returned %d contacts for %s (%s)",
			name, len(contacts), company.Name, company.Domain))
		return contacts, nil
	}

	// All sources exhausted
	slog.Info(fmt.Sprintf("FallbackSource: ✗ All sources exhausted for %s (%s) — no contacts found",
		company.Name, company.Domain))
	return nil, nil
}

// Status returns a human-readable summary of the configured fallback chain.
func (f *FallbackSource) Status() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var b strings.Builder
	fmt.Fprintf(&b, "FallbackSource: %d source(s) configured\n", len(f.sources))
	for i, name := range f.names {
		_ = i
		fmt.Fprintf(&b, "  [%d] %s\n", i+1, name)
	}
	return b.String()
}

// Sources returns the underlying source list (for testing/inspection).
func (f *FallbackSource) Sources() []EnrichmentSource {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cp := make([]EnrichmentSource, len(f.sources))
	copy(cp, f.sources)
	return cp
}

// isGarbageLead returns true if the company has an obviously invalid domain
// or name that resulted from HN comment parsing errors. Skips the entire
// enrichment chain to avoid wasting API calls.
func isGarbageLead(company db.Company) bool {
	domain := strings.ToLower(company.Domain)
	name := strings.ToLower(company.Name)

	// Domain must have at least one dot and valid TLD-like structure
	if domain == "" || domain == ".com" {
		return true
	}
	// Domain must contain a dot
	if !strings.Contains(domain, ".") {
		return true
	}
	// Reject domains that contain URL encoding artifacts or HTML fragments
	garbageDomainPatterns := []string{
		"x2f", "x26", "x3b", "x3d", "x27", // URL-encoded HTML entities
		"&#",                           // Raw HTML entities
		"https", "http", "www.", "://", // URLs embedded in domain
		"%20", "%2f", "%26", "%23", // URL-encoded chars
	}
	for _, p := range garbageDomainPatterns {
		if strings.Contains(domain, p) {
			return true
		}
	}
	// Reject names that are clearly garbage from HN parsing
	if strings.Contains(name, "&#") || strings.Contains(name, "x2f") ||
		strings.Contains(name, "___") || strings.Contains(name, "http") ||
		len(name) > 120 {
		return true
	}
	// Reject domains that look like concatenated garbage
	if strings.Count(domain, ".") > 3 {
		return true // Too many dots
	}
	return false
}

// Names returns the human-readable names for each source.
func (f *FallbackSource) Names() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	cp := make([]string, len(f.names))
	copy(cp, f.names)
	return cp
}
