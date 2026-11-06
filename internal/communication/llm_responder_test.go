package communication

import (
	"context"
	"strings"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
	"gitlab.com/robertpelloni/marketing_agent/internal/llm"
)

func TestLLMResponseGenerator_Generate(t *testing.T) {
	provider := &llm.MockLLMProvider{}
	generator := NewRAGResponseGenerator(nil, provider)

	salesCtx := SalesContext{
		Contact: db.Contact{Name: "John Doe", Role: "CTO"},
		Company: db.Company{Name: "TechCorp"},
		Deal:    db.Deal{TechnicalDossier: "INFRASTRUCTURE_BOTTLENECK detected in legacy k8s clusters."},
		LatestIntent: IntentTechnical,
		Interactions: []db.Interaction{
			{RawText: "How can TormentNexus help with our scaling?"},
		},
	}

	resp, err := generator.Generate(context.Background(), salesCtx, ActionRespond)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if !strings.Contains(resp, "John Doe") {
		t.Errorf("Response should contain contact name")
	}
	if !strings.Contains(resp, "INFRASTRUCTURE_BOTTLENECK") {
		t.Errorf("Response should contain technical dossier context")
	}
}

func TestLLMResponseGenerator_MCPSkillTargeting(t *testing.T) {
	provider := &llm.MockLLMProvider{}
	generator := NewRAGResponseGenerator(nil, provider)

	salesCtx := SalesContext{
		Contact: db.Contact{Name: "Alice Smith", Role: "Head of AI", Email: "alice@enterprise.com"},
		Company: db.Company{Name: "AgenticCorp", Domain: "enterprise.com", TechStack: []string{"mcp-server", "Go"}},
		Deal:    db.Deal{TechnicalDossier: "Evaluating model-context-protocol tool routing."},
		LatestIntent: IntentTechnical,
		Interactions: []db.Interaction{
			{RawText: "We need efficient tool calling for our custom MCP servers."},
		},
	}

	resp, err := generator.Generate(context.Background(), salesCtx, ActionRespond)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if !strings.Contains(resp, "MCP & Skill-Based Targeting") {
		t.Errorf("Expected response prompt to include MCP & Skill-Based Targeting guidance")
	}
	if !strings.Contains(resp, "progressive tool routing") {
		t.Errorf("Expected response prompt to highlight progressive tool routing")
	}
}
