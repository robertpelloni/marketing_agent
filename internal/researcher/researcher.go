package researcher

import (
	"context"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

<<<<<<< HEAD
=======
// Crawler defines an interface for extracting technical insights from web sources.
>>>>>>> origin/main
type Crawler interface {
	Crawl(ctx context.Context, target string) (string, error)
}

<<<<<<< HEAD
=======
// DossierProcessor defines an interface for synthesizing research findings.
>>>>>>> origin/main
type DossierProcessor interface {
	Process(findings []string) (string, error)
}

<<<<<<< HEAD
=======
// Researcher coordinates technical deep research for leads.
>>>>>>> origin/main
type Researcher struct {
	db        *db.DB
	crawlers  []Crawler
	processor DossierProcessor
	crmClient crm.CRMClient
}

<<<<<<< HEAD
func NewResearcher(database *db.DB, crawlers []Crawler, processor DossierProcessor, crm crm.CRMClient) *Researcher {
	return &Researcher{db: database, crawlers: crawlers, processor: processor, crmClient: crm}
=======
// NewResearcher creates a new Researcher instance.
func NewResearcher(database *db.DB, crawlers []Crawler, processor DossierProcessor, crmClient crm.CRMClient) *Researcher {
	return &Researcher{
		db:        database,
		crawlers:  crawlers,
		processor: processor,
		crmClient: crmClient,
	}
>>>>>>> origin/main
}
