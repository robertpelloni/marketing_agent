package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator implements ResponseGenerator with technical context awareness, prompt versioning, and negative constraint injection.
type RAGResponseGenerator struct {
	db       *db.DB
	llm      llm.LLMProvider
	registry *llm.PromptRegistry
}

func NewRAGResponseGenerator(database *db.DB, llmProvider llm.LLMProvider, registry *llm.PromptRegistry) *RAGResponseGenerator {
	return &RAGResponseGenerator{
		db:       database,
		llm:      llmProvider,
		registry: registry,
	}
}

func (r *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	// 1. Check for known objections first
	if salesCtx.LatestIntent == IntentObjection {
		var lastInbound string
		for i := len(salesCtx.Interactions) - 1; i >= 0; i-- {
			if salesCtx.Interactions[i].Direction == "Inbound" {
				lastInbound = salesCtx.Interactions[i].RawText
				break
			}
		}
		if lastInbound != "" {
			rebuttal := GetBestRebuttal(lastInbound)
			log.Printf("RAGResponder: Using library rebuttal for objection.")
			return rebuttal, nil
		}
	}

	// 2. Fetch negative examples (failed interactions) to improve response quality
	negativeContext := ""
	if r.db != nil {
		// We could fetch interactions marked with success=false
		// For now, simulate fetching some common "failures" or actual past data
		negativeContext = "AVOID the following patterns found in past unsuccessful outreach: being too generic, missing technical specifics, and failing to provide a clear CTA."
	}

	// 3. Use PromptRegistry for A/B tested responses
	if r.registry != nil && r.llm != nil {
		data := map[string]string{
			"intent":   string(salesCtx.LatestIntent),
			"dossier":  salesCtx.Deal.TechnicalDossier,
			"company":  salesCtx.Company.Name,
			"negative": negativeContext,
		}

		promptText, err := r.registry.ResolvePrompt("outreach-reply", data)
		if err == nil {
			prompt := llm.Prompt{
				System: "You are an elite enterprise sales engineer for TormentNexus.",
				User:   promptText,
			}
			return r.llm.Generate(ctx, prompt)
		}
	}

	// 4. Fallback to hardcoded LLM generation with negative context
	if r.llm != nil {
		prompt := llm.Prompt{
			System: "You are an elite enterprise sales engineer for TormentNexus. Ground your response in the technical dossier and AVOID past mistakes.",
			User:   fmt.Sprintf("Intent: %s. Dossier: %s. %s Generate a professional reply.", salesCtx.LatestIntent, salesCtx.Deal.TechnicalDossier, negativeContext),
		}
		return r.llm.Generate(ctx, prompt)
	}

	return "Hello, I'd like to follow up on our previous technical discussion regarding TormentNexus.", nil
}
