package db

import "time"

// LeadState represents the current state of a lead in the pipeline.
type LeadState string

const (
	StateDiscovered   LeadState = "Discovered"
	StateResearched   LeadState = "Researched"
	StateOutreachSent LeadState = "Outreach_Sent"
	StateEngaged      LeadState = "Engaged"
	StateNegotiating  LeadState = "Negotiating"
	StateClosedWon    LeadState = "Closed_Won"
	StateClosedLost   LeadState = "Closed_Lost"
)

// Company represents a target organization.
type Company struct {
	ID            int64     `db:"id"`
	Name          string    `db:"name"`
	Domain        string    `db:"domain"`
	TechStack     []string  `db:"tech_stack"`
	HiringSignals []string  `db:"hiring_signals"`
	MarketCapTier string    `db:"market_cap_tier"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Contact represents an individual decision-maker at a company.
type Contact struct {
	ID             int64     `db:"id"`
	CompanyID      int64     `db:"company_id"`
	Name           string    `db:"name"`
	Role           string    `db:"role"`
	Email          string    `db:"email"`
	GitHubHandle   string    `db:"github_handle"`
	LinkedInURL    string    `db:"linkedin_url"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// Interaction tracks communications with a contact.
type Interaction struct {
	ID        int64     `db:"id"`
	ContactID int64     `db:"contact_id"`
	Channel   string    `db:"channel"`   // e.g., Email, LinkedIn, GitHub
	Direction string    `db:"direction"` // e.g., Inbound, Outbound
	RawText   string    `db:"raw_text"`
	Summary   string    `db:"summary"`
	Sentiment string    `db:"sentiment"`
	CreatedAt time.Time `db:"created_at"`
}

// Deal tracks the financial and state progress of a lead.
type Deal struct {
	ID                 int64     `db:"id"`
	CompanyID          int64     `db:"company_id"`
	CurrentState       LeadState `db:"current_state"`
	QuotedPricing      float64   `db:"quoted_pricing"`
	CustomRequirements string    `db:"custom_requirements"`
	TechnicalDossier   string    `db:"technical_dossier"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}
