package communication

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator provides technically grounded replies using Pseudo-RAG.
type RAGResponseGenerator struct {
	llm      llm.LLMProvider
	borgDocs string
}

// NewRAGResponseGenerator creates a new generator with Borg context.
func NewRAGResponseGenerator(provider llm.LLMProvider) *RAGResponseGenerator {
	docsPath := "borg/docs/ARCHITECTURE.md"
	content, err := os.ReadFile(docsPath)
	if err != nil {
		log.Printf("RAG: Warning: could not load Borg documentation: %v", err)
	}

	return &RAGResponseGenerator{
		llm:      provider,
		borgDocs: string(content),
	}
}

func (g *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	log.Printf("RAGResponseGenerator: Generating response for intent: %s", salesCtx.LatestIntent)

	// Inject technical context if the intent is technical
	contextInjection := ""
	if salesCtx.LatestIntent == IntentTechnical && g.borgDocs != "" {
		contextInjection = fmt.Sprintf("\nRelevant Technical Context from Borg Docs:\n%s\n", g.truncateDocs(g.borgDocs))
	}

	// Inject pricing context if the intent is pricing
	if salesCtx.LatestIntent == IntentPricing {
		pricing := CalculateQuote(salesCtx.Company.MarketCapTier)
		contextInjection = fmt.Sprintf("\nPricing Context: Annual subscription is approximately $%d based on company size.\n", pricing)
	}

	latestMsg := "START_OUTREACH"
	if len(salesCtx.Interactions) > 0 {
		latestMsg = salesCtx.Interactions[0].RawText
	}

	prompt := llm.Prompt{
		System: "You are a senior sales engineer at Borg. Use the provided technical and pricing context to draft a hyper-personalized response.",
		User: fmt.Sprintf("Draft a reply to %s at %s. Intent: %s. Action: %s. %s\nLatest Message: %s\nTechnical Dossier: %s",
			salesCtx.Contact.Name, salesCtx.Company.Name, salesCtx.LatestIntent, action, contextInjection, latestMsg, salesCtx.Deal.TechnicalDossier),
	}

	return g.llm.Generate(ctx, prompt)
}

func (g *RAGResponseGenerator) truncateDocs(docs string) string {
	if len(docs) > 2000 {
		return docs[:2000] + "... [truncated]"
	}
	return docs
}
