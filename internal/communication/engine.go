package communication
import (
	"context"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
	"strings"
)
type LearningSalesEngine struct {
	db *db.DB; crmClient crm.CRMClient; llmProvider llm.LLMProvider
}
func NewLearningSalesEngine(database *db.DB, crmClient crm.CRMClient, llmProvider llm.LLMProvider) *LearningSalesEngine {
	return &LearningSalesEngine{db: database, crmClient: crmClient, llmProvider: llmProvider}
}
func (e *LearningSalesEngine) ScoreLead(ctx SalesContext) int {
	score := 0
	if ctx.Company.MarketCapTier == "Enterprise" { score += 50 }
	if ctx.Deal.TechnicalDossier != "" {
		if strings.Contains(strings.ToLower(ctx.Deal.TechnicalDossier), "bottleneck") { score += 30 }
	}
	score += len(ctx.Interactions) * 2
	return score
}
func (e *LearningSalesEngine) QualifyLead(ctx SalesContext) int {
	score := e.ScoreLead(ctx); qual := score / 2
	inboundCount := 0
	for _, i := range ctx.Interactions { if i.Direction == "Inbound" { inboundCount++ } }
	if inboundCount > 2 { qual += 20 }
	switch ctx.LatestIntent {
	case IntentMeetingRequest: qual += 25
	case IntentPricing: qual += 15
	}
	return qual
}
func (e *LearningSalesEngine) Decide(ctx context.Context, salesCtx SalesContext) (Action, error) {
	qual := e.QualifyLead(salesCtx)
	if qual >= 70 && salesCtx.Deal.CurrentState == db.StateEngaged { return ActionAdvanceState, nil }
	if (salesCtx.LatestIntent == IntentPricing || qual >= 70) && salesCtx.Deal.CurrentState == db.StateNegotiating { return ActionAdvanceState, nil }
	return ActionRespond, nil
}
