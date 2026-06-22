package communication

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"log/slog"
>>>>>>> origin/main
	"os"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// RAGResponseGenerator provides technically grounded replies using Pseudo-RAG.
type RAGResponseGenerator struct {
<<<<<<< HEAD
	db       *db.DB
	llm      llm.LLMProvider
	tormentNexusDocs string
=======
	db               *db.DB
	llm              llm.LLMProvider
	tormentNexusDocs string
	objectionLib     *ObjectionLibrary
>>>>>>> origin/main
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
<<<<<<< HEAD
			log.Printf("RAG: Successfully loaded TormentNexus documentation from %s", path)
=======
			slog.Info(fmt.Sprintf("RAG: Successfully loaded TormentNexus documentation from %s", path))
>>>>>>> origin/main
			break
		}
	}

	if err != nil {
<<<<<<< HEAD
		log.Printf("RAG: Warning: could not load TormentNexus documentation: %v", err)
	}

	return &RAGResponseGenerator{
		db:       database,
		llm:      provider,
		tormentNexusDocs: string(content),
=======
		slog.Info(fmt.Sprintf("RAG: Warning: could not load TormentNexus documentation: %v", err))
	}

	return &RAGResponseGenerator{
<<<<<<< HEAD
		db:			database,
		llm:			provider,
		tormentNexusDocs:	string(content),
		objectionLib:	NewObjectionLibrary(),
=======
		db:               database,
		llm:              provider,
		tormentNexusDocs: string(content),
		objectionLib:     NewObjectionLibrary(),
>>>>>>> origin/main
	}
}

func (g *RAGResponseGenerator) Generate(ctx context.Context, salesCtx SalesContext, action Action) (string, error) {
<<<<<<< HEAD
	log.Printf("RAGResponseGenerator: Generating response for intent: %s", salesCtx.LatestIntent)
=======
	slog.Info(fmt.Sprintf("RAGResponseGenerator: Generating response for intent: %s", salesCtx.LatestIntent))
>>>>>>> origin/main

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
<<<<<<< HEAD
	if g.db != nil {
=======
	if g.db != nil && g.db.Conn != nil {
>>>>>>> origin/main
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

<<<<<<< HEAD
	prompt := llm.Prompt{
		System: "You are a senior sales engineer at TormentNexus. Use the provided technical and pricing context to draft a hyper-personalized response.",
		User: fmt.Sprintf("Draft a reply to %s at %s. Intent: %s. Action: %s. %s\nLatest Message: %s\nTechnical Dossier: %s",
			salesCtx.Contact.Name, salesCtx.Company.Name, salesCtx.LatestIntent, action, contextInjection, latestMsg, salesCtx.Deal.TechnicalDossier),
=======
	// Objection handling: use library to find a matched rebuttal
	if salesCtx.LatestIntent == IntentObjection && len(salesCtx.Interactions) > 0 {
		matched := g.objectionLib.MatchObjection(ctx, latestMsg, SentimentResult{}, salesCtx.Deal.CurrentState)
		if matched != nil {
			slog.Info(fmt.Sprintf("ObjectionLibrary: Matched \"%s\" (score=%.2f)", matched.Objection.Title, matched.Score))
			return matched.Response.Text, nil
		}
		slog.Info("ObjectionLibrary: No match found for intent, falling through to LLM")
	}

	prompt := llm.Prompt{
<<<<<<< HEAD
		System:	"You are a senior sales engineer at TormentNexus. Use the provided technical and pricing context to draft a hyper-personalized response.",
		User: fmt.Sprintf("Draft a reply to %s at %s. Intent: %s. Action: %s. %s\nLatest Message: %s\nTechnical Dossier: %s",
			salesCtx.Contact.Name, salesCtx.Company.Name, salesCtx.LatestIntent, action, contextInjection, latestMsg, salesCtx.Deal.TechnicalDossier),
=======
		System: `salesPersona:
You are the world's best technical sales engineer.
You are a master of these methodologies. USE THEM:

1. SPIN Selling: Ask about Situation, dig for Problem, amplify Implication, present Need-payoff.
2. The Challenger Sale: Teach the customer something new about their own business. Tailor the pitch to their specific stack. Take control of the conversation.
3. Sandler Rule: Qualify early. If they're not a fit, don't chase. Build reciprocity by giving value first.
4. Cialdini's Principles:
   - Reciprocity: Give something valuable (insight, benchmark, config snippet)
   - Scarcity: Time-limited implementation slots, limited beta access
   - Authority: Reference TormentNexus's 11K+ server catalog, enterprise deployments
   - Social Proof: "Teams using similar stacks to yours have seen 3-5x improvements"
   - Consistency: Get a small yes first ("does that sound like a challenge you're facing?")
   - Liking: Reference their specific work, be genuinely helpful
5. Loss Aversion: Frame inaction as a cost: "Every month without this, you're losing X in engineering hours to coordination overhead."
6. Feel-Felt-Found for objections: "I understand how you feel. Other teams felt the same way. What they found was..."

TONE:
- Confident but not arrogant. Expert peer, not pushy salesperson.
- Technically fluent. Reference their specific tech stack and pain points.
- Use "we" and "you" language. Build partnership.
- Be concise. Engineers value brevity.
- Never use buzzwords without substance. Always ground claims in specifics.

OUTREACH STRUCTURE:
1. Hook: Reference something specific about their work (repo, blog post, tech choice)
2. Problem: Identify a pain point they likely have based on their stack
3. Implication: What that pain costs them (time, money, engineer-hours)
4. Solution: How TormentNexus solves it (specific, technical)
5. Social Proof: Other teams with similar stacks
6. Low-friction CTA: "Worth a 15-min screen share to see if this applies?"

Remember: You're not selling software. You're selling engineering hours back to them.`,
		User: fmt.Sprintf("Draft a reply to %s at %s. Intent: %s. Action: %s. %s\nLatest Message: %s\nTechnical Dossier: %s\nCompany Tech Stack: %s\nContact Role: %s",
			salesCtx.Contact.Name, salesCtx.Company.Name, salesCtx.LatestIntent, action, contextInjection, latestMsg, salesCtx.Deal.TechnicalDossier, strings.Join(salesCtx.Company.TechStack, ", "), salesCtx.Contact.Role),
>>>>>>> origin/main
	}

	return g.llm.Generate(ctx, prompt)
}

func (g *RAGResponseGenerator) truncateDocs(docs string) string {
	if len(docs) > 2000 {
		return docs[:2000] + "... [truncated]"
	}
	return docs
}
<<<<<<< HEAD
=======

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

	// Placeholder replacements for body
	replacements := map[string]string{
<<<<<<< HEAD
		"{{contact}}":        getValue(salesCtx.Contact.Name),
		"{{company}}":        getValue(salesCtx.Company.Name),
		"{{tech_stack}}":     strings.Join(salesCtx.Company.TechStack, ", "),
		"{{role}}":           getValue(salesCtx.Contact.Role),
		"{{github_handle}}":  getValue(salesCtx.Contact.GitHubHandle),
		"{{linkedin_url}}":   getValue(salesCtx.Contact.LinkedInURL),
		"{{email}}":          getValue(salesCtx.Contact.Email),
		"{{specific_project}}": "TormentNexus",
		"{{repo}}":           getValue(salesCtx.Company.Name, "AI-Platform"),
		"{{market_cap_tier}}": getValue(salesCtx.Company.MarketCapTier, "Enterprise"),
=======
		"{{contact}}":          getValue(salesCtx.Contact.Name),
		"{{company}}":          getValue(salesCtx.Company.Name),
		"{{tech_stack}}":       strings.Join(salesCtx.Company.TechStack, ", "),
		"{{role}}":             getValue(salesCtx.Contact.Role),
		"{{github_handle}}":    getValue(salesCtx.Contact.GitHubHandle),
		"{{linkedin_url}}":     getValue(salesCtx.Contact.LinkedInURL),
		"{{email}}":            getValue(salesCtx.Contact.Email),
		"{{specific_project}}": "TormentNexus",
		"{{repo}}":             getValue(salesCtx.Company.Name, "AI-Platform"),
		"{{market_cap_tier}}":  getValue(salesCtx.Company.MarketCapTier, "Enterprise"),
>>>>>>> origin/main
	}

	// Replace placeholders in body
	for placeholder, value := range replacements {
		body = strings.ReplaceAll(body, placeholder, value)
	}

	// Replace placeholders in subject
	for placeholder, value := range replacements {
		subject = strings.ReplaceAll(subject, placeholder, value)
	}

	return subject, body, nil
}
>>>>>>> origin/main
