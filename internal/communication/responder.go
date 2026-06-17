package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator implements ResponseGenerator with technical context awareness and prompt versioning.
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

	// 2. Use PromptRegistry for A/B tested responses
	if r.registry != nil && r.llm != nil {
		data := map[string]string{
			"intent":  string(salesCtx.LatestIntent),
			"dossier": salesCtx.Deal.TechnicalDossier,
			"company": salesCtx.Company.Name,
		}

		promptText, err := r.registry.ResolvePrompt("outreach-reply", data)
		if err == nil {
			prompt := llm.Prompt{
				System: "You are an elite enterprise sales engineer for TormentNexus.",
				User:   promptText,
			}
			return r.llm.Generate(ctx, prompt)
		}
		log.Printf("RAGResponder: PromptRegistry resolution failed: %v, falling back to hardcoded prompt", err)
	}

	// 3. Fallback to hardcoded LLM generation if registry fails
	if r.llm != nil {
		prompt := llm.Prompt{
			System: "You are an elite enterprise sales engineer for TormentNexus. Ground your response in the provided technical dossier.",
			User:   fmt.Sprintf("Intent: %s. Dossier: %s. Generate a professional and persuasive reply.", salesCtx.LatestIntent, salesCtx.Deal.TechnicalDossier),
		}
		return r.llm.Generate(ctx, prompt)
	}

	return "Hello, I'd like to follow up on our previous technical discussion regarding TormentNexus.", nil
}
