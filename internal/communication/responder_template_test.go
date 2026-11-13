package communication

import (
    "context"
	"strings"
    "testing"
    "gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func TestGenerateFromTemplate(t *testing.T) {
    tmpl := &db.Template{
        ID:      "test-email",
        Subject: "Hello {{contact}} from {{company}}",
        Body:    "Hi {{contact}},\nWe noticed your work on {{tech_stack}} and think TormentNexus can help.",
        Channel: "email",
    }

    salesCtx := SalesContext{
        Company: db.Company{Name: "Acme Corp", TechStack: []string{"Go", "Kubernetes"}},
        Contact: db.Contact{Name: "Alice"},
    }

    rg := &RAGResponseGenerator{}
    subject, body, err := rg.GenerateFromTemplate(context.Background(), tmpl, salesCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if subject != "Hello Alice from Acme Corp" {
        t.Errorf("subject mismatch: got %q", subject)
    }
    expectedBody := "Hi Alice,\nWe noticed your work on Go, Kubernetes and think TormentNexus can help."
    if body != expectedBody {
        t.Errorf("body mismatch: got %q expected %q", body, expectedBody)
    }
}

func TestGenerateFromTemplate_ResearchInsight(t *testing.T) {
    rg := &RAGResponseGenerator{}
    tmpl := &db.Template{
        Subject: "Following up on {{research_insight}}",
        Body:    "Hi {{contact}}, I reviewed your stack and noted: {{technical_dossier}}.",
    }
    salesCtx := SalesContext{
        Contact: db.Contact{Name: "Bob", Email: "bob@company.com"},
        Company: db.Company{Name: "MegaCorp", Domain: "company.com"},
        Deal: db.Deal{
            TechnicalDossier: "High GPU cluster latency in vector memory retrieval",
        },
    }

    subject, body, err := rg.GenerateFromTemplate(context.Background(), tmpl, salesCtx)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !strings.Contains(subject, "High GPU cluster latency") {
        t.Errorf("expected research insight in subject, got %q", subject)
    }
    if !strings.Contains(body, "High GPU cluster latency in vector memory retrieval") {
        t.Errorf("expected technical dossier in body, got %q", body)
    }
}
