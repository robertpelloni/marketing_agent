package communication

import (
	"context"
	"fmt"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator implements ResponseGenerator with technical context awareness.
type RAGResponseGenerator struct {
	db  *db.DB
	llm llm.LLMProvider
}

func NewRAGResponseGenerator(database *db.DB, llmProvider llm.LLMProvider) *RAGResponseGenerator {
	return &RAGResponseGenerator{
		db:  database,
		llm: llmProvider,
	}
}

func (r *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	// 1. Check for known objections first
	if salesCtx.LatestIntent == IntentObjection {
		// Use first inbound interaction for context
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

	// 2. Standard LLM generation with RAG context
	if r.llm != nil {
		prompt := llm.Prompt{
			System: "You are an elite enterprise sales engineer for TormentNexus. Ground your response in the provided technical dossier.",
			User:   fmt.Sprintf("Intent: %s. Dossier: %s. Generate a professional and persuasive reply.", salesCtx.LatestIntent, salesCtx.Deal.TechnicalDossier),
		}
		return r.llm.Generate(ctx, prompt)
	}

	return "Hello, I'd like to follow up on our previous technical discussion regarding TormentNexus.", nil
}
