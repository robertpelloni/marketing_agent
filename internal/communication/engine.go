package communication

import (
	"context"
	"log/slog"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

type LearningSalesEngine struct {
	db        *db.DB
	crmClient crm.CRMClient
	llm       llm.LLMProvider
}

func NewLearningSalesEngine(database *db.DB, crmClient crm.CRMClient, llmProvider llm.LLMProvider) *LearningSalesEngine {
	return &LearningSalesEngine{
		db:        database,
		crmClient: crmClient,
		llm:       llmProvider,
	}
}

func (e *LearningSalesEngine) Decide(ctx context.Context, salesCtx SalesContext) (Action, error) {
	slog.Info("LearningSalesEngine: Deciding next action", "deal", salesCtx.Deal.ID, "intent", salesCtx.LatestIntent)

	// HITL GATE: If high value and not approved, wait for human
	if e.isHighValueLead(salesCtx) && salesCtx.Deal.CurrentState == db.StateResearched && !salesCtx.Deal.ApprovalRequired {
		if e.db != nil {
			_ = e.db.SetApprovalRequired(ctx, salesCtx.Deal.ID, true)
		}
		return ActionWait, nil
	}

	if e.shouldAdvanceState(salesCtx) {
		newState := db.StateNegotiating
		if e.QualifyLead(salesCtx) >= 80 && salesCtx.LatestIntent == IntentFollowUp {
			newState = db.StateClosedWon
		}
		if e.isHighValueLead(salesCtx) && newState != db.StatePendingApproval && newState != db.StateClosedWon {
			newState = db.StatePendingApproval
		}

		slog.Info("LearningSalesEngine: Advancing deal state", "deal", salesCtx.Deal.ID, "to", newState)
		if e.db != nil {
			if err := e.db.UpdateDealState(ctx, salesCtx.Deal.ID, newState); err == nil {
				_ = e.db.MarkTemplateSuccessForDeal(ctx, salesCtx.Deal.ID)
				if e.crmClient != nil {
					go func() {
						updatedDeal := salesCtx.Deal
						updatedDeal.CurrentState = newState
						_ = e.crmClient.PushDeal(ctx, updatedDeal, salesCtx.Company, e.RouteLead(salesCtx))
					}()
				}
			}
		}
		return ActionAdvanceState, nil
	}

	switch salesCtx.LatestIntent {
	case IntentPricing:
		if e.isHighValueLead(salesCtx) { return ActionEscalate, nil }
		return ActionRespond, nil
	case IntentObjection:
		if e.countInteractionTypes(salesCtx.Interactions, "Outbound") >= 2 { return ActionEscalate, nil }
		return ActionRespond, nil
	case IntentSpam:
		return ActionWait, nil
	case IntentMeetingRequest:
		return ActionRespond, nil
	}

	return ActionRespond, nil
}

func (e *LearningSalesEngine) shouldAdvanceState(ctx SalesContext) bool {
	if ctx.Deal.CurrentState == db.StateEngaged && (len(ctx.Interactions) > 3 || e.QualifyLead(ctx) > 70) {
		return true
	}
	if ctx.Deal.CurrentState == db.StateNegotiating && (ctx.LatestIntent == IntentFollowUp || ctx.LatestIntent == IntentPricing) {
		return e.QualifyLead(ctx) > 85
	}
	return false
}

func (e *LearningSalesEngine) isHighValueLead(ctx SalesContext) bool {
	return ctx.Company.MarketCapTier == "Enterprise" || ctx.Deal.QuotedPricing >= 100000 || e.ScoreLead(ctx) > 80
}

func (e *LearningSalesEngine) countInteractionTypes(interactions []db.Interaction, direction string) int {
	count := 0
	for _, i := range interactions {
		if i.Direction == direction { count++ }
	}
	return count
}

func (e *LearningSalesEngine) ScoreLead(salesCtx SalesContext) int {
	score := 0
	switch strings.ToLower(salesCtx.Company.MarketCapTier) {
	case "enterprise": score += 50
	case "mid-market": score += 25
	}
	if strings.Contains(salesCtx.Deal.TechnicalDossier, "BOTTLENECK") { score += 30 }
	score += len(salesCtx.Interactions) * 2
	if score > 100 { return 100 }
	return score
}

func (e *LearningSalesEngine) QualifyLead(ctx SalesContext) int {
	qual := e.ScoreLead(ctx) / 2
	if e.countInteractionTypes(ctx.Interactions, "Inbound") > 2 { qual += 20 }
	switch ctx.LatestIntent {
	case IntentPricing: qual += 15
	case IntentMeetingRequest: qual += 25
	case IntentFollowUp: qual += 20
	}
	if ctx.LatestIntent == IntentObjection { qual -= 10 }
	if qual > 100 { return 100 }
	if qual < 0 { return 0 }
	return qual
}

func (e *LearningSalesEngine) RouteLead(salesCtx SalesContext) string {
	if salesCtx.Company.MarketCapTier == "Enterprise" && salesCtx.LatestIntent == IntentTechnical {
		return "Lead Solutions Architect"
	}
	if e.ScoreLead(salesCtx) > 80 { return "Senior Account Executive" }
	return "Sales Representative"
}

func CalculateQuote(tier string) int {
	switch strings.ToLower(tier) {
	case "enterprise": return 50000
	case "mid-market": return 25000
	default: return 10000
	}
}
