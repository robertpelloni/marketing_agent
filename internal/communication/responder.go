package communication

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

// RAGResponseGenerator provides technically grounded replies using Pseudo-RAG.
type RAGResponseGenerator struct {
	db			*db.DB
	llm			llm.LLMProvider
	tormentNexusDocs	string
	objectionLib	*ObjectionLibrary
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
			slog.Info(fmt.Sprintf("RAG: Successfully loaded TormentNexus documentation from %s", path))
			break
		}
	}

	if err != nil {
		slog.Info(fmt.Sprintf("RAG: Warning: could not load TormentNexus documentation: %v", err))
	}

	return &RAGResponseGenerator{
		db:			database,
		llm:			provider,
		tormentNexusDocs:	string(content),
		objectionLib:	NewObjectionLibrary(),
	}
}

func (g *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
	slog.Info(fmt.Sprintf("RAGResponseGenerator: Generating response for intent: %s", salesCtx.LatestIntent))

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
	// Implement A/B testing for prompt generation
	// Use interaction count or deal ID to determine group
	useExamples := false
	if salesCtx.Deal.ID%2 == 0 {
		useExamples = true
		slog.Info(fmt.Sprintf("A/B Testing: Injecting few-shot examples for Deal ID %d", salesCtx.Deal.ID))
	} else {
		slog.Info(fmt.Sprintf("A/B Testing: NOT injecting few-shot examples for Deal ID %d (control group)", salesCtx.Deal.ID))
	}

	if g.db != nil && useExamples {
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

	// Objection handling: use library to find a matched rebuttal
	if salesCtx.LatestIntent == IntentObjection && len(salesCtx.Interactions) > 0 {
		matched := g.objectionLib.MatchObjection(ctx, latestMsg, SentimentResult{}, salesCtx.Deal.CurrentState)
		if matched != nil {
			slog.Info(fmt.Sprintf("ObjectionLibrary: Matched \"%s\" (score=%.2f)", matched.Objection.Title, matched.Score))
			return matched.Response.Text, nil
		}
		slog.Info("ObjectionLibrary: No match found for intent, falling through to LLM")
	}

	isCorp := IsCorporate(salesCtx.Contact.Email, salesCtx.Company.Domain)

	systemPrompt := "You are a senior sales engineer at HyperNexus (hypernexus.site). Use the provided technical and pricing context to draft a hyper-personalized response. Ground technical descriptions on the HyperNexus platform (the corporate version of TormentNexus, with the open-source fork at github.com/HyperNexusSoft/HyperNexus). Focus on core enterprise capabilities: custom tool registry, persistent multi-tier memory (L1/L2), SSO/OIDC auth, RBAC, and audit logs. Do NOT pitch LLM provider routing, model fallbacks, or gateway waterfalls, as enterprise targets already possess their own reliable LLM provider infrastructure."
	if !isCorp {
		systemPrompt = "You are an expert developer advocate for TormentNexus (tormentnexus.site). Use the provided technical context to draft a hyper-personalized response. Ground technical descriptions on TormentNexus (the local-first cognitive control plane and open-source model hypervisor at github.com/NexusSoftMDMA/TormentNexus). Focus on self-hosting, open-source freedom, developer velocity, custom tools, and local-first memory. Do NOT mention LLM routing, fallbacks, or waterfalls."
	}

	// Focus on keeping initial outreach very short, direct, and to the point
	userPrompt := fmt.Sprintf("Draft a reply to %s at %s. Intent: %s. Action: %s. %s\nLatest Message: %s\nTechnical Dossier: %s",
		salesCtx.Contact.Name, salesCtx.Company.Name, salesCtx.LatestIntent, action, contextInjection, latestMsg, salesCtx.Deal.TechnicalDossier)
	if latestMsg == "START_OUTREACH" {
		userPrompt += "\nCRITICAL: This is the initial outreach email. Keep it extremely short (under 4 sentences) and to the point. Focus strictly on key features (custom tools, skills, multi-tier memory) and how they directly help their engineering work. Avoid generic corporate fluff."
	}

	prompt := llm.Prompt{
		System: systemPrompt,
		User:   userPrompt,
	}

	return g.llm.Generate(ctx, prompt)
}

func (g *RAGResponseGenerator) truncateDocs(docs string) string {
	if len(docs) > 2000 {
		return docs[:2000] + "... [truncated]"
	}
	return docs
}

// GenerateFromTemplate renders a template with context-specific placeholders.
// It returns the rendered body and subject.
func (g *RAGResponseGenerator) GenerateFromTemplate(ctx context.Context, tmpl *db.Template, salesCtx SalesContext) (subject, body string, err error) {
	subject = tmpl.Subject
	body = tmpl.Body

	// Helper to safely get a string value
	getValue := func(parts ...string) string {
		for _, part := range parts {
			if part != "" {
				return part
			}
		}
		return ""
	}

	isCorp := IsCorporate(salesCtx.Contact.Email, salesCtx.Company.Domain)

	// Placeholder replacements for body
	replacements := map[string]string{
		"{{contact}}":        getValue(salesCtx.Contact.Name),
		"{{company}}":        getValue(salesCtx.Company.Name),
		"{{tech_stack}}":     strings.Join(salesCtx.Company.TechStack, ", "),
		"{{role}}":           getValue(salesCtx.Contact.Role),
		"{{github_handle}}":  getValue(salesCtx.Contact.GitHubHandle),
		"{{linkedin_url}}":   getValue(salesCtx.Contact.LinkedInURL),
		"{{email}}":          getValue(salesCtx.Contact.Email),
		"{{specific_project}}": func() string {
			if isCorp {
				return "HyperNexus"
			}
			return "TormentNexus"
		}(),
		"{{repo}}":           getValue(salesCtx.Company.Name, "AI-Platform"),
		"{{market_cap_tier}}": getValue(salesCtx.Company.MarketCapTier, "Corporate"),
	}

	// Replace placeholders in body
	for placeholder, value := range replacements {
		body = strings.ReplaceAll(body, placeholder, value)
	}

	// Replace placeholders in subject
	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
	}

	// If the recipient is not corporate, we dynamically replace HyperNexus references with TormentNexus.
	if !isCorp {
		body = strings.ReplaceAll(body, "HyperNexus (hypernexus.site)", "TormentNexus (tormentnexus.site)")
		body = strings.ReplaceAll(body, "HyperNexus", "TormentNexus")
		body = strings.ReplaceAll(body, "hypernexus.site", "tormentnexus.site")
		body = strings.ReplaceAll(body, "the enterprise-ready cloud-hosted version of TormentNexus", "the open-source, local-first model hypervisor")
		body = strings.ReplaceAll(body, "the enterprise-grade cloud version of TormentNexus", "the open-source, local-first model hypervisor")
		body = strings.ReplaceAll(body, "stable fork of TormentNexus at github.com/HyperNexusSoft/HyperNexus", "open-source repo at github.com/NexusSoftMDMA/TormentNexus")
		body = strings.ReplaceAll(body, "github.com/HyperNexusSoft/HyperNexus", "github.com/NexusSoftMDMA/TormentNexus")
		body = strings.ReplaceAll(body, "Corporate", "Developer")

		subject = strings.ReplaceAll(subject, "HyperNexus", "TormentNexus")
	}

	return subject, body, nil
}
