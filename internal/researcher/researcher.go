package researcher

import (
	"context"

	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

type Crawler interface {
	Crawl(ctx context.Context, target string) (string, error)
}

type DossierProcessor interface {
	Process(findings []string) (string, error)
}

type Researcher struct {
	db        *db.DB
	crawlers  []Crawler
	processor DossierProcessor
	crmClient crm.CRMClient
}

func NewResearcher(database *db.DB, crawlers []Crawler, processor DossierProcessor, crm crm.CRMClient) *Researcher {
	return &Researcher{db: database, crawlers: crawlers, processor: processor, crmClient: crm}
}
