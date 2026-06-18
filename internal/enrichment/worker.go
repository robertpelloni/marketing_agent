package enrichment

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// Run starts the background enrichment process.
func (e *Enricher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Enricher worker started...")

	// Run immediately on startup
	e.executeEnrichment(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Enricher worker stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			e.executeEnrichment(ctx)
		}
	}
}

// ExecuteEnrichment manually triggers an enrichment cycle (exported for testing).
func (e *Enricher) ExecuteEnrichment(ctx context.Context) {
	e.executeEnrichment(ctx)
}

func (e *Enricher) executeEnrichment(ctx context.Context) {
	if e.db == nil {
		// DB not available – skip enrichment to avoid panic in dev/test environments
		slog.Info("Enricher: DB unavailable, skipping enrichment cycle")
		return
	}
	// 1. Find deals in Discovered state
	deals, err := e.db.ListDealsByState(ctx, db.StateDiscovered)
	if err != nil {
		slog.Info(fmt.Sprintf("Enricher: Error listing discovered deals: %v", err))
		return
	}

	for _, deal := range deals {
		company, err := e.db.GetCompanyByID(ctx, deal.CompanyID)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error getting company %d: %v", deal.CompanyID, err))
			continue
		}

		err = e.enrichCompany(ctx, deal, *company)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error enriching company %s: %v", company.Name, err))
		}
	}
}

func (e *Enricher) enrichCompany(ctx context.Context, deal db.Deal, company db.Company) error {
	for _, source := range e.sources {
		contacts, err := source.Enrich(ctx, company)
		if err != nil {
			slog.Info(fmt.Sprintf("Enricher: Error from source: %v", err))
			continue
		}

		for _, contact := range contacts {
			contact.CompanyID = company.ID
			err := e.db.CreateContact(ctx, &contact)
			if err != nil {
				slog.Info(fmt.Sprintf("Enricher: Error persisting contact %s: %v", contact.Name, err))
			}
		}

		if len(contacts) > 0 {
			// Advance deal state to Researched
			err = e.db.UpdateDealState(ctx, deal.ID, db.StateResearched)
			if err != nil {
				return fmt.Errorf("failed to update deal state: %w", err)
			}

			// Synchronize newly found contacts with the CRM (with retry logic)
			if e.crmClient != nil {
				go func() {
					maxRetries := 3
					for i := 0; i < maxRetries; i++ {
						if err := e.crmClient.SyncContacts(ctx, company.ID, contacts); err != nil {
							slog.Info(fmt.Sprintf("Enricher Warning: Failed to sync contacts to CRM (attempt %d/%d): %v", i+1, maxRetries, err))
							time.Sleep(time.Duration(i+1) * 2 * time.Second)
							continue
						}
						return
					}
					slog.Info(fmt.Sprintf("Enricher Error: CRM contact synchronization failed after %d attempts for company %d", maxRetries, company.ID))
				}()
			}

			slog.Info(fmt.Sprintf("Enricher: Successfully enriched %s with %d contacts", company.Name, len(contacts)))
			return nil
		}
	}
	return nil
}

// MockApolloSource is a simulated enrichment source.
// It generates plausible contacts for ANY domain to enable full pipeline flow.
type MockApolloSource struct {
	mu        sync.Mutex
	callCount int
}

var mockFirstNames = []string{"Alex", "Jordan", "Morgan", "Casey", "Riley", "Taylor", "Avery", "Quinn", "Drew", "Reese",
	"Sam", "Blake", "Cameron", "Dakota", "Ellis", "Finley", "Harper", "Jade", "Kai", "Logan"}
var mockLastNames = []string{"Chen", "Patel", "Kim", "Singh", "Garcia", "Martinez", "Thompson", "Zhang", "Kumar", "Okafor",
	"Anders", "Bennett", "Crawford", "Donovan", "Espinoza", "Fischer", "Guevara", "Huang", "Ito", "Johansson"}
var mockRoles = []string{"VP of Engineering", "Director of AI", "CTO", "Principal Architect", "Head of ML",
	"Chief Scientist", "Engineering Manager", "Staff Engineer", "Tech Lead", "Head of Infrastructure"}

func (m *MockApolloSource) Enrich(ctx context.Context, company db.Company) ([]db.Contact, error) {
	m.mu.Lock()
	m.callCount++
	idx := m.callCount
	m.mu.Unlock()

	slog.Info(fmt.Sprintf("MockApolloSource: Generating contacts for %s (call #%d)", company.Domain, idx))

	// Generate 1-3 mock contacts per company with deterministic names based on domain
	domainHash := sha256.Sum256([]byte(company.Domain))
	r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(domainHash[:8]))))

	numContacts := r.Intn(3) + 1
	contacts := make([]db.Contact, 0, numContacts)

	for i := 0; i < numContacts; i++ {
		firstName := mockFirstNames[r.Intn(len(mockFirstNames))]
		lastName := mockLastNames[r.Intn(len(mockLastNames))]
		role := mockRoles[r.Intn(len(mockRoles))]

		email := fmt.Sprintf("%s.%s@%s", strings.ToLower(firstName), strings.ToLower(lastName), company.Domain)
		if strings.Contains(company.Domain, "github.com") {
			email = fmt.Sprintf("%s.%s@gmail.com", strings.ToLower(firstName), strings.ToLower(lastName))
		}

		contacts = append(contacts, db.Contact{
			Name:         fmt.Sprintf("%s %s", firstName, lastName),
			Role:         role,
			Email:        email,
			GitHubHandle: fmt.Sprintf("%s-%s", strings.ToLower(firstName), strings.ToLower(lastName)),
		})
	}

	slog.Info(fmt.Sprintf("MockApolloSource: Generated %d contacts for %s", len(contacts), company.Domain))
	return contacts, nil
}
