package researcher

import (
	"strings"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// IntentSignal describes a single point of interest for scoring lead intent.
type IntentSignal struct {
	Type   string // "hiring", "tech_stack", "domain", "product_fit", "github", "rss", "blog", "competitor"
	Weight int    // Base score multiplier
	Value  string // Raw signal data
}

// IntentScore aggregates signals into a normalized intent score (0-100).
type IntentScore struct {
	Score   int
	Signals []IntentSignal
}

// IntentAggregator unifies disparate signals from a company's profile
// into a singular intent score that prioritizes outreach execution.
type IntentAggregator struct{}

// Aggregate synthesizes a comprehensive intent score based on available data.
func (a *IntentAggregator) Aggregate(company db.Company, dossier string) IntentScore {
	var signals []IntentSignal
	totalScore := 0

	// 1. Hiring & Issue Signals
	for _, hiring := range company.HiringSignals {
		lowerHiring := strings.ToLower(hiring)
		weight := 5 // Base for generic signal

		if strings.Contains(lowerHiring, "open source") {
			weight = 8
		}
		if strings.Contains(lowerHiring, "ai") || strings.Contains(lowerHiring, "ml") || strings.Contains(lowerHiring, "machine learning") {
			weight = 15
		}
		if strings.Contains(lowerHiring, "mcp") || strings.Contains(lowerHiring, "agent") || strings.Contains(lowerHiring, "llm") || strings.Contains(lowerHiring, "tool") {
			weight = 18
		}
		if strings.Contains(lowerHiring, "infrastructure") || strings.Contains(lowerHiring, "platform") || strings.Contains(lowerHiring, "orchestration") {
			weight = 20
		}

		signals = append(signals, IntentSignal{
			Type:   "hiring",
			Weight: weight,
			Value:  hiring,
		})
		totalScore += weight
	}

	// 2. Tech Stack & Company Maturity Signals
	if len(company.TechStack) > 0 {
		stackWeight := 0
		var matched []string
		for _, tech := range company.TechStack {
			lt := strings.ToLower(strings.TrimSpace(tech))
			if lt == "" {
				continue
			}
			matched = append(matched, tech)
			switch lt {
			case "go", "golang", "rust", "typescript":
				if stackWeight < 12 {
					stackWeight = 12
				}
			case "python", "javascript", "java", "c#", "kotlin":
				if stackWeight < 8 {
					stackWeight = 8
				}
			default:
				if stackWeight < 4 {
					stackWeight = 4
				}
			}
		}
		if stackWeight > 0 {
			signals = append(signals, IntentSignal{
				Type:   "tech_stack",
				Weight: stackWeight,
				Value:  strings.Join(matched, ", "),
			})
			totalScore += stackWeight
		}
	}

	lowerDomain := strings.ToLower(strings.TrimSpace(company.Domain))
	if lowerDomain != "" && !strings.HasPrefix(lowerDomain, "github.com/") && !db.IsPersonalEmailDomain(lowerDomain) {
		signals = append(signals, IntentSignal{
			Type:   "domain",
			Weight: 10,
			Value:  company.Domain,
		})
		totalScore += 10
	}

	switch strings.ToLower(strings.TrimSpace(company.MarketCapTier)) {
	case "enterprise":
		signals = append(signals, IntentSignal{Type: "tier", Weight: 15, Value: company.MarketCapTier})
		totalScore += 15
	case "mid-market":
		signals = append(signals, IntentSignal{Type: "tier", Weight: 10, Value: company.MarketCapTier})
		totalScore += 10
	case "startup":
		signals = append(signals, IntentSignal{Type: "tier", Weight: 5, Value: company.MarketCapTier})
		totalScore += 5
	}

	// 3. Product-Fit Keyword Alignment (MCP, multi-agent routing, memory, gateways)
	combinedProfile := strings.ToLower(company.Name + " " + company.Domain + " " + strings.Join(company.HiringSignals, " "))
	fitKeywords := []string{"mcp", "context", "agent", "router", "gateway", "memory", "swarm", "workflow", "rag", "llm"}
	fitHits := 0
	for _, kw := range fitKeywords {
		if strings.Contains(combinedProfile, kw) {
			fitHits++
		}
	}
	if fitHits > 0 {
		fitWeight := fitHits * 6
		if fitWeight > 24 {
			fitWeight = 24
		}
		signals = append(signals, IntentSignal{
			Type:   "product_fit",
			Weight: fitWeight,
			Value:  company.Name,
		})
		totalScore += fitWeight
	}

	// 4. Dossier Analysis (GitHub, RSS, Blogs, Competitors)
	lowerDossier := strings.ToLower(dossier)

	if strings.Contains(lowerDossier, "github insight") {
		weight := 15
		if strings.Contains(lowerDossier, "orchestration") || strings.Contains(lowerDossier, "state management") {
			weight = 25
		}
		signals = append(signals, IntentSignal{
			Type:   "github",
			Weight: weight,
			Value:  "Verified repository activity indicates multi-agent workflow complexity",
		})
		totalScore += weight
	} else if strings.Contains(lowerDossier, "bottleneck detected") {
		signals = append(signals, IntentSignal{
			Type:   "github",
			Weight: 8,
			Value:  "Heuristic repository orchestration bottleneck candidate",
		})
		totalScore += 8
	}

	if strings.Contains(lowerDossier, "rss:") || strings.Contains(lowerDossier, "atom:") {
		signals = append(signals, IntentSignal{
			Type:   "rss",
			Weight: 20,
			Value:  "Matched live engineering RSS/Atom mention",
		})
		totalScore += 20
	}

	if strings.Contains(lowerDossier, "blog insight") {
		weight := 10
		if strings.Contains(lowerDossier, "evaluating move") || strings.Contains(lowerDossier, "scaling") {
			weight = 12
		}
		signals = append(signals, IntentSignal{
			Type:   "blog",
			Weight: weight,
			Value:  "Technical blog post analyzing orchestration constraints",
		})
		totalScore += weight
	}

	if strings.Contains(lowerDossier, "competitive intelligence") {
		weight := 0
		if strings.Contains(lowerDossier, "struggling") || strings.Contains(lowerDossier, "rate-limiting") {
			weight = 30
			signals = append(signals, IntentSignal{
				Type:   "competitor",
				Weight: weight,
				Value:  "Evident friction with existing competitive solution",
			})
		} else if strings.Contains(lowerDossier, "evaluation phase") || strings.Contains(lowerDossier, "starred") {
			weight = 20
			signals = append(signals, IntentSignal{
				Type:   "competitor",
				Weight: weight,
				Value:  "Actively evaluating competitive frameworks",
			})
		}
		totalScore += weight
	}

	// Normalize Score (Cap at 100)
	if totalScore > 100 {
		totalScore = 100
	}

	return IntentScore{
		Score:   totalScore,
		Signals: signals,
	}
}
