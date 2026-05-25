package db

import (
	"context"
	"fmt"
	"time"
)

// CreateCompany inserts a new company into the database.
func (db *DB) CreateCompany(ctx context.Context, company *Company) error {
	query := `
		INSERT INTO companies (name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	company.CreatedAt = now
	company.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		company.Name,
		company.Domain,
		company.TechStack,
		company.HiringSignals,
		company.MarketCapTier,
		company.CreatedAt,
		company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// GetCompanyByDomain retrieves a company by its domain.
func (db *DB) GetCompanyByDomain(ctx context.Context, domain string) (*Company, error) {
	query := `
		SELECT id, name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at
		FROM companies
		WHERE domain = $1
	`
	company := &Company{}
	err := db.Conn.QueryRowContext(ctx, query, domain).Scan(
		&company.ID,
		&company.Name,
		&company.Domain,
		&company.TechStack,
		&company.HiringSignals,
		&company.MarketCapTier,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by domain: %w", err)
	}
	return company, nil
}

// CreateDeal inserts a new deal for a company.
func (db *DB) CreateDeal(ctx context.Context, deal *Deal) error {
	query := `
		INSERT INTO deals (company_id, current_state, quoted_pricing, custom_requirements, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	now := time.Now()
	deal.CreatedAt = now
	deal.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		deal.CompanyID,
		deal.CurrentState,
		deal.QuotedPricing,
		deal.CustomRequirements,
		deal.CreatedAt,
		deal.UpdatedAt,
	).Scan(&deal.ID)

	if err != nil {
		return fmt.Errorf("failed to create deal: %w", err)
	}
	return nil
}

// UpdateDealState updates the state of an existing deal.
func (db *DB) UpdateDealState(ctx context.Context, dealID int64, newState LeadState) error {
	query := `
		UPDATE deals
		SET current_state = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Conn.ExecContext(ctx, query, newState, time.Now(), dealID)
	if err != nil {
		return fmt.Errorf("failed to update deal state: %w", err)
	}
	return nil
}

// ListRecentDeals retrieves the most recently updated deals.
func (db *DB) ListRecentDeals(ctx context.Context, limit int) ([]Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, created_at, updated_at
		FROM deals
		ORDER BY updated_at DESC
		LIMIT $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent deals: %w", err)
	}
	defer rows.Close()

	var deals []Deal
	for rows.Next() {
		var deal Deal
		if err := rows.Scan(
			&deal.ID,
			&deal.CompanyID,
			&deal.CurrentState,
			&deal.QuotedPricing,
			&deal.CustomRequirements,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deals = append(deals, deal)
	}
	return deals, nil
}
