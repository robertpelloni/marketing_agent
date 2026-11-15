package researcher

import (
	"strings"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func TestResearcher_Run(t *testing.T) {
	r := NewResearcher(nil, []Crawler{&GitHubCrawler{}}, &DefaultDossierProcessor{}, nil)
	if r == nil {
		t.Fatal("Failed to create researcher")
	}
}

func TestBuildResearchTargets_StripsEmailsAndPersonalDomains(t *testing.T) {
	company := &db.Company{
		Name:   "mcp-router",
		Domain: "github.com/velella-sh/mcp-router",
	}
	contact := db.Contact{
		Name:         "Ryu",
		Email:        "ryu.yamamoto@icloud.com",
		GitHubHandle: "velella-sh",
	}

	targets := buildResearchTargets(company, contact)
	if len(targets) == 0 {
		t.Fatal("expected non-empty research targets")
	}
	for _, tgt := range targets {
		if strings.Contains(tgt, "@") || strings.Contains(tgt, "icloud.com") {
			t.Fatalf("expected raw email and consumer webmail domain to be stripped, got %q in %v", tgt, targets)
		}
	}
	if targets[0] != "velella-sh" {
		t.Fatalf("expected GitHub handle/owner first, got %v", targets)
	}
}

func TestIntentAggregator_ProducesDifferentiatedScores(t *testing.T) {
	agg := &IntentAggregator{}

	lowLead := agg.Aggregate(db.Company{
		Name:          "misc-notes",
		Domain:        "github.com/jdoe/misc-notes",
		TechStack:     []string{"HTML"},
		HiringSignals: []string{"Active Open Source contributor"},
		MarketCapTier: "SMB",
	}, "BOTTLENECK DETECTED: jdoe/llm-orchestration uses high-latency serial state updates.")

	highLead := agg.Aggregate(db.Company{
		Name:          "acme-mcp-agent-router",
		Domain:        "acme-ai.dev",
		TechStack:     []string{"Go", "TypeScript"},
		HiringSignals: []string{"GitHub Issue: MCP multi-agent orchestration bottleneck - https://github.com/acme-ai/router/issues/9"},
		MarketCapTier: "Mid-Market",
	}, "REAL-TIME GitHub INSIGHT for acme-ai: Active repositories [router (Go)]. Analyzing codebase for state management and orchestration patterns.\nRSS: Acme is hiring ML engineers – https://acme-ai.dev/jobs")

	if lowLead.Score >= highLead.Score {
		t.Fatalf("expected high-fit MCP Go company (%d) to outscore generic HTML repo (%d)", highLead.Score, lowLead.Score)
	}
	if highLead.Score < 70 {
		t.Fatalf("expected high-fit MCP lead score >= 70, got %d", highLead.Score)
	}
	if lowLead.Score > 40 {
		t.Fatalf("expected generic HTML repo score <= 40, got %d", lowLead.Score)
	}
}
