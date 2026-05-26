package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
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
		pq.Array(company.TechStack),
		pq.Array(company.HiringSignals),
		company.MarketCapTier,
		company.CreatedAt,
		company.UpdatedAt,
	).Scan(&company.ID)

	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// GetCompanyByID retrieves a company by its ID.
func (db *DB) GetCompanyByID(ctx context.Context, id int64) (*Company, error) {
	query := `
		SELECT id, name, domain, tech_stack, hiring_signals, market_cap_tier, created_at, updated_at
		FROM companies
		WHERE id = $1
	`
	company := &Company{}
	err := db.Conn.QueryRowContext(ctx, query, id).Scan(
		&company.ID,
		&company.Name,
		&company.Domain,
		pq.Array(&company.TechStack),
		pq.Array(&company.HiringSignals),
		&company.MarketCapTier,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get company by id: %w", err)
	}
	return company, nil
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
		pq.Array(&company.TechStack),
		pq.Array(&company.HiringSignals),
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

// ListDealsByState retrieves deals in a specific state.
func (db *DB) ListDealsByState(ctx context.Context, state LeadState) ([]Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
		FROM deals
		WHERE current_state = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, state)
	if err != nil {
		return nil, fmt.Errorf("failed to list deals by state: %w", err)
	}
	defer rows.Close()

	var deals []Deal
	for rows.Next() {
		var deal Deal
		var pricing sql.NullFloat64
		var requirements, dossier sql.NullString
		if err := rows.Scan(
			&deal.ID,
			&deal.CompanyID,
			&deal.CurrentState,
			&pricing,
			&requirements,
			&dossier,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deal.QuotedPricing = pricing.Float64
		deal.CustomRequirements = requirements.String
		deal.TechnicalDossier = dossier.String
		deals = append(deals, deal)
	}
	return deals, nil
}

// GetDealByCompanyID retrieves the most recent deal for a specific company.
func (db *DB) GetDealByCompanyID(ctx context.Context, companyID int64) (*Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
		FROM deals
		WHERE company_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`
	deal := &Deal{}
	var pricing sql.NullFloat64
	var requirements, dossier sql.NullString
	err := db.Conn.QueryRowContext(ctx, query, companyID).Scan(
		&deal.ID,
		&deal.CompanyID,
		&deal.CurrentState,
		&pricing,
		&requirements,
		&dossier,
		&deal.CreatedAt,
		&deal.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal by company id: %w", err)
	}
	deal.QuotedPricing = pricing.Float64
	deal.CustomRequirements = requirements.String
	deal.TechnicalDossier = dossier.String
	return deal, nil
}

// ListRecentDeals retrieves the most recently updated deals.
func (db *DB) ListRecentDeals(ctx context.Context, limit int) ([]Deal, error) {
	query := `
		SELECT id, company_id, current_state, quoted_pricing, custom_requirements, technical_dossier, created_at, updated_at
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
		var pricing sql.NullFloat64
		var requirements, dossier sql.NullString
		if err := rows.Scan(
			&deal.ID,
			&deal.CompanyID,
			&deal.CurrentState,
			&pricing,
			&requirements,
			&dossier,
			&deal.CreatedAt,
			&deal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan deal: %w", err)
		}
		deal.QuotedPricing = pricing.Float64
		deal.CustomRequirements = requirements.String
		deal.TechnicalDossier = dossier.String
		deals = append(deals, deal)
	}
	return deals, nil
}

// UpdateTechnicalDossier updates the technical dossier for a deal.
func (db *DB) UpdateTechnicalDossier(ctx context.Context, dealID int64, dossier string) error {
	query := `
		UPDATE deals
		SET technical_dossier = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Conn.ExecContext(ctx, query, dossier, time.Now(), dealID)
	if err != nil {
		return fmt.Errorf("failed to update technical dossier: %w", err)
	}
	return nil
}

// CreateContact inserts a new contact into the database.
func (db *DB) CreateContact(ctx context.Context, contact *Contact) error {
	query := `
		INSERT INTO contacts (company_id, name, role, email, github_handle, linkedin_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	now := time.Now()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	err := db.Conn.QueryRowContext(ctx, query,
		contact.CompanyID,
		contact.Name,
		contact.Role,
		contact.Email,
		contact.GitHubHandle,
		contact.LinkedInURL,
		contact.CreatedAt,
		contact.UpdatedAt,
	).Scan(&contact.ID)

	if err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}
	return nil
}

// ListContactsByCompany retrieves all contacts for a specific company.
func (db *DB) ListContactsByCompany(ctx context.Context, companyID int64) ([]Contact, error) {
	query := `
		SELECT id, company_id, name, role, email, github_handle, linkedin_url, created_at, updated_at
		FROM contacts
		WHERE company_id = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		if err := rows.Scan(
			&contact.ID,
			&contact.CompanyID,
			&contact.Name,
			&contact.Role,
			&contact.Email,
			&contact.GitHubHandle,
			&contact.LinkedInURL,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// CreateInteraction inserts a new interaction into the database.
func (db *DB) CreateInteraction(ctx context.Context, interaction *Interaction) error {
	query := `
		INSERT INTO interactions (contact_id, channel, direction, raw_text, summary, sentiment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	if interaction.CreatedAt.IsZero() {
		interaction.CreatedAt = time.Now()
	}

	err := db.Conn.QueryRowContext(ctx, query,
		interaction.ContactID,
		interaction.Channel,
		interaction.Direction,
		interaction.RawText,
		interaction.Summary,
		interaction.Sentiment,
		interaction.CreatedAt,
	).Scan(&interaction.ID)

	if err != nil {
		return fmt.Errorf("failed to create interaction: %w", err)
	}
	return nil
}

// CreatePullRequest persists a new pull request record.
func (db *DB) CreatePullRequest(ctx context.Context, pr *gitcheck.PullRequest, taskDesc string) error {
	query := `
		INSERT INTO pull_requests (id, branch, title, status, task_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := db.Conn.ExecContext(ctx, query, pr.ID, pr.Branch, pr.Title, pr.Status, taskDesc)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	return nil
}

// UpdatePRStatus updates the status of an existing PR.
func (db *DB) UpdatePRStatus(ctx context.Context, prID string, status gitcheck.PRStatus) error {
	query := `
		UPDATE pull_requests
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := db.Conn.ExecContext(ctx, query, status, prID)
	if err != nil {
		return fmt.Errorf("failed to update PR status: %w", err)
	}
	return nil
}

// ListActivePullRequests retrieves all open pull requests.
func (db *DB) ListActivePullRequests(ctx context.Context) ([]gitcheck.PullRequest, error) {
	query := `
		SELECT id, branch, title, status
		FROM pull_requests
		WHERE status = $1
	`
	rows, err := db.Conn.QueryContext(ctx, query, gitcheck.PRStatusOpen)
	if err != nil {
		return nil, fmt.Errorf("failed to list active PRs: %w", err)
	}
	defer rows.Close()

	var prs []gitcheck.PullRequest
	for rows.Next() {
		var pr gitcheck.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Branch, &pr.Title, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// ListInteractionsByContact retrieves all interactions for a specific contact.
func (db *DB) ListInteractionsByContact(ctx context.Context, contactID int64) ([]Interaction, error) {
	query := `
		SELECT id, contact_id, channel, direction, raw_text, summary, sentiment, created_at
		FROM interactions
		WHERE contact_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Conn.QueryContext(ctx, query, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to list interactions: %w", err)
	}
	defer rows.Close()

	var interactions []Interaction
	for rows.Next() {
		var interaction Interaction
		if err := rows.Scan(
			&interaction.ID,
			&interaction.ContactID,
			&interaction.Channel,
			&interaction.Direction,
			&interaction.RawText,
			&interaction.Summary,
			&interaction.Sentiment,
			&interaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan interaction: %w", err)
		}
		interactions = append(interactions, interaction)
	}
	return interactions, nil
}
