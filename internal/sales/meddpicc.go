package sales

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/robertpelloni/marketing_agent/internal/db"
	"github.com/robertpelloni/marketing_agent/internal/llm"
)

// MEDDPICCEvaluator analyzes interaction history to populate MEDDPICC fields.
type MEDDPICCEvaluator struct {
	llm llm.LLMProvider
}

// NewMEDDPICCEvaluator creates a new evaluator.
func NewMEDDPICCEvaluator(provider llm.LLMProvider) *MEDDPICCEvaluator {
	return &MEDDPICCEvaluator{llm: provider}
}

// Evaluate Deal updates MEDDPICC fields by asking the LLM to parse the interaction history.
func (e *MEDDPICCEvaluator) EvaluateDeal(ctx context.Context, deal *db.Deal, interactions []db.Interaction) error {
	if e.llm == nil {
		return fmt.Errorf("no LLM provider configured for MEDDPICC evaluation")
	}

	if len(interactions) == 0 {
		return nil // Nothing to evaluate
	}

	// Create a summary of recent interactions
	var history strings.Builder
	for i, interaction := range interactions {
		// Only send the last 5 interactions to save context
		if i >= 5 {
			break
		}
		history.WriteString(fmt.Sprintf("[%s]: %s\n", interaction.Direction, interaction.RawText))
	}

	systemPrompt := `You are an elite enterprise sales MEDDPICC evaluator.
Analyze the interaction history and extract any signals related to the following categories.
Return a comma-separated list of updates in this format:
KEY: Value

Keys must be one of: METRICS, ECO_BUYER, DECISION_CRITERIA, DECISION_PROCESS, PAPER_PROCESS, IDENTIFY_PAIN, CHAMPION, COMPETITION.
If no signal is found for a key, do not include it. Keep the value concise and direct.`

	prompt := llm.Prompt{
		System: systemPrompt,
		User:   fmt.Sprintf("Deal ID: %d\nHistory:\n%s", deal.ID, history.String()),
	}

	response, err := e.llm.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("MEDDPICC LLM generation failed: %w", err)
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Only update if we found something meaningful
			if val != "" && val != "None" && val != "N/A" && val != "Not mentioned" {
				switch key {
				case "METRICS":
					deal.MEDDPICCMetrics = val
				case "ECO_BUYER":
					deal.MEDDPICCEcoBuyer = val
				case "DECISION_CRITERIA":
					deal.MEDDPICCDecisionCriteria = val
				case "DECISION_PROCESS":
					deal.MEDDPICCDecisionProcess = val
				case "PAPER_PROCESS":
					deal.MEDDPICCPaperProcess = val
				case "IDENTIFY_PAIN":
					deal.MEDDPICCIdentifyPain = val
				case "CHAMPION":
					deal.MEDDPICCChampion = val
				case "COMPETITION":
					deal.MEDDPICCCompetition = val
				}
				slog.Info(fmt.Sprintf("MEDDPICC Tracking: Updated %s for Deal %d", key, deal.ID))
			}
		}
	}

	return nil
}
