package communication

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator provides technically grounded replies using Pseudo-RAG.
type RAGResponseGenerator struct {
	db       *db.DB
	llm      llm.LLMProvider
	tormentNexusDocs string
}

// NewRAGResponseGenerator creates a new generator with TormentNexus context.
func NewRAGResponseGenerator(database *db.DB, provider llm.LLMProvider) *RAGResponseGenerator {
	// Documentation path resolution to support both root execution and package-level tests
	docsPaths := []string{
		"borg/docs/ARCHITECTURE.md",
		"../../borg/docs/ARCHITECTURE.md",
		"../../../borg/docs/ARCHITECTURE.md",
	}

	var content []byte
	var err error
	for _, path := range docsPaths {
		// #nosec G304 -- Documentation paths are internal to the repository structure
		content, err = os.ReadFile(path)
		if err == nil {
			log.Printf("RAG: Successfully loaded TormentNexus documentation from %s", path)
			break
		}
	}

	if err != nil {
		log.Printf("RAG: Warning: could not load TormentNexus documentation: %v", err)
	}

	return &RAGResponseGenerator{
		db:       database,
		llm:      provider,
		tormentNexusDocs: string(content),
	}
}

func (g *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	log.Printf("RAGResponseGenerator: Generating response for intent: %s", salesCtx.LatestIntent)

	// Inject technical context if the intent is technical
	contextInjection := ""
	if salesCtx.LatestIntent == IntentTechnical && g.tormentNexusDocs != "" {
		contextInjection = fmt.Sprintf("\nRelevant Technical Context from TormentNexus Docs:\n%s\n", g.truncateDocs(g.tormentNexusDocs))
	}

	// Inject pricing context if the intent is pricing
	if salesCtx.LatestIntent == IntentPricing {
		pricing := CalculateQuote(salesCtx.Company.MarketCapTier)
		contextInjection = fmt.Sprintf("\nPricing Context: Annual subscription is approximately $%d based on company size.\n", pricing)
	}

	// SELF-IMPROVING PROMPTS: Inject successful past interactions as few-shot examples
	if g.db != nil && g.db.Conn != nil {
		successes, err := g.db.ListSuccessfulInteractions(ctx, 3)
		if err == nil && len(successes) > 0 {
			examples := []string{}
			for _, s := range successes {
				examples = append(examples, fmt.Sprintf("- Successful Response: %s", s.Summary))
			}
			contextInjection += fmt.Sprintf("\nSuccessful Past Outreach Examples:\n%s\n", strings.Join(examples, "\n"))
		}
	}

	latestMsg := "START_OUTREACH"
	if len(salesCtx.Interactions) > 0 {
		latestMsg = salesCtx.Interactions[0].RawText
	}

	prompt := llm.Prompt{
		System: "You are a senior sales engineer at TormentNexus. Use the provided technical and pricing context to draft a hyper-personalized response.",
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
