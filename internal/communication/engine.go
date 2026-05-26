package communication

import (
	"context"
	"log"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// LearningSalesEngine implements the SalesStrategy interface.
type LearningSalesEngine struct {
	db *db.DB
}

// NewLearningSalesEngine creates a new instance of the engine.
func NewLearningSalesEngine(database *db.DB) *LearningSalesEngine {
	return &LearningSalesEngine{db: database}
}

// Decide determines the next action for a lead.
func (e *LearningSalesEngine) Decide(ctx context.Context, salesCtx SalesContext) (Action, error) {
	log.Printf("LearningSalesEngine: Deciding next action for deal %d (Latest Intent: %s)", salesCtx.Deal.ID, salesCtx.LatestIntent)

	// 1. Analyze historical performance
	if e.shouldAdvanceState(salesCtx) {
		log.Printf("LearningSalesEngine: Advancing deal %d to Negotiating state.", salesCtx.Deal.ID)
		if e.db != nil {
			if err := e.db.UpdateDealState(ctx, salesCtx.Deal.ID, db.StateNegotiating); err != nil {
				log.Printf("LearningSalesEngine: Error updating deal state: %v", err)
			}
		}
		return ActionAdvanceState, nil
	}

	// 2. Base intent-driven logic
	switch salesCtx.LatestIntent {
	case IntentTechnical:
		return ActionRespond, nil
	case IntentPricing:
		if e.isHighValueLead(salesCtx) {
			return ActionRespond, nil
		}
		return ActionEscalate, nil // Escalate high-tier pricing negotiation
	case IntentObjection:
		// Attempt one autonomous rebuttal, then escalate
		if e.countInteractionTypes(salesCtx.Interactions, "Outbound") < 2 {
			return ActionRespond, nil
		}
		return ActionEscalate, nil
	case IntentSpam:
		return ActionWait, nil
	}

	return ActionRespond, nil
}

func (e *LearningSalesEngine) shouldAdvanceState(ctx SalesContext) bool {
	// Logic to advance state from Engaged to Negotiating if interest is high
	return len(ctx.Interactions) > 3 && ctx.LatestIntent == IntentPricing
}

func (e *LearningSalesEngine) isHighValueLead(ctx SalesContext) bool {
	return ctx.Company.MarketCapTier == "Enterprise"
}

func (e *LearningSalesEngine) countInteractionTypes(interactions []db.Interaction, direction string) int {
	count := 0
	for _, i := range interactions {
		if i.Direction == direction {
			count++
		}
	}
	return count
}
