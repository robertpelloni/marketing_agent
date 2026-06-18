package communication

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

type RAGResponseGenerator struct {
	db               *db.DB
	llm              llm.LLMProvider
	registry         *llm.PromptRegistry
	tormentNexusDocs string
}

func NewRAGResponseGenerator(database *db.DB, provider llm.LLMProvider, registry *llm.PromptRegistry) *RAGResponseGenerator {
	docsPaths := []string{
		"borg/docs/ARCHITECTURE.md",
		"../../borg/docs/ARCHITECTURE.md",
		"../../../borg/docs/ARCHITECTURE.md",
	}

	var content []byte
	var err error
	for _, path := range docsPaths {
		content, err = os.ReadFile(path)
		if err == nil {
			slog.Info("RAG: Successfully loaded TormentNexus documentation", "path", path)
			break
		}
	}

	return &RAGResponseGenerator{
		db:               database,
		llm:              provider,
		registry:         registry,
		tormentNexusDocs: string(content),
	}
}

func (r *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	slog.Info("RAGResponseGenerator: Generating response", "intent", salesCtx.LatestIntent)

	contextInjection := ""
	if salesCtx.LatestIntent == IntentTechnical && r.tormentNexusDocs != "" {
		contextInjection = fmt.Sprintf("\nTechnical Context:\n%s\n", r.truncateDocs(r.tormentNexusDocs))
	}
	if salesCtx.LatestIntent == IntentPricing {
		pricing := sales.CalculateQuote(salesCtx.Company.MarketCapTier)
		contextInjection = fmt.Sprintf("\nPricing Context: Annual subscription approx $%d.\n", pricing)
	}

	negativeContext := "AVOID past mistakes: being too generic, missing technical specifics."

	if r.registry != nil && r.llm != nil {
		data := map[string]string{
			"intent":   string(salesCtx.LatestIntent),
			"dossier":  salesCtx.Deal.TechnicalDossier,
			"company":  salesCtx.Company.Name,
			"negative": negativeContext,
			"context":  contextInjection,
		}
		promptText, err := r.registry.ResolvePrompt("outreach-reply", data)
		if err == nil {
			return r.llm.Generate(ctx, llm.Prompt{
				System: "You are an elite enterprise sales engineer for TormentNexus.",
				User:   promptText,
			})
		}
	}

	if r.llm != nil {
		prompt := llm.Prompt{
			System: "You are an elite enterprise sales engineer for TormentNexus.",
			User:   fmt.Sprintf("Intent: %s. Action: %s. Context: %s. Dossier: %s. Generate a professional reply.", salesCtx.LatestIntent, action, contextInjection, salesCtx.Deal.TechnicalDossier),
		}
		return r.llm.Generate(ctx, prompt)
	}

	return "Hello, I'd like to follow up on our discussion regarding TormentNexus.", nil
}

func (r *RAGResponseGenerator) truncateDocs(docs string) string {
	if len(docs) > 2000 {
		return docs[:2000] + "..."
	}
	return docs
}

func (g *RAGResponseGenerator) GenerateFromTemplate(ctx context.Context, tmpl *db.Template, salesCtx SalesContext) (subject, body string, err error) {
	subject = tmpl.Subject
	body = tmpl.Body

	replacements := map[string]string{
		"{{contact}}":        salesCtx.Contact.Name,
		"{{company}}":        salesCtx.Company.Name,
		"{{tech_stack}}":     strings.Join(salesCtx.Company.TechStack, ", "),
		"{{role}}":           salesCtx.Contact.Role,
	}

	for placeholder, value := range replacements {
		body = strings.ReplaceAll(body, placeholder, value)
		subject = strings.ReplaceAll(subject, placeholder, value)
	}

	return subject, body, nil
}
