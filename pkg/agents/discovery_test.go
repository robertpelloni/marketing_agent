package agents

import (
	"context"
	"strings"
	"testing"
	"time"

	"gitlab.com/robertpelloni/marketing_agent/internal/llm"
)

func TestBuildDiscoveryQuery_WithoutLLM(t *testing.T) {
	ctx := context.Background()
	query := BuildDiscoveryQuery(nil, ctx, 24*time.Hour)

	if !strings.Contains(query, "model-context-protocol") {
		t.Errorf("Expected base query terms in query, got %q", query)
	}

	today := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	expectedQualifier := "pushed:>" + today
	if !strings.Contains(query, expectedQualifier) {
		t.Errorf("Expected recency qualifier %q in query, got %q", expectedQualifier, query)
	}
}

func TestBuildDiscoveryQuery_ZeroLookback(t *testing.T) {
	ctx := context.Background()
	query := BuildDiscoveryQuery(nil, ctx, 0)

	if strings.Contains(query, "pushed:>") {
		t.Errorf("Expected no pushed qualifier with 0 lookback, got %q", query)
	}
}

type fixedMockLLM struct {
	response string
}

func (m *fixedMockLLM) Generate(ctx context.Context, prompt llm.Prompt) (string, error) {
	return m.response, nil
}

func (m *fixedMockLLM) GenerateJSON(ctx context.Context, prompt llm.Prompt, v any) error {
	return nil
}

func TestBuildDiscoveryQuery_ValidLLMQuery(t *testing.T) {
	ctx := context.Background()
	mock := &fixedMockLLM{response: "mcp tool router language:Rust"}
	query := BuildDiscoveryQuery(mock, ctx, 24*time.Hour)

	if !strings.Contains(query, "mcp tool router language:Rust") {
		t.Errorf("Expected custom LLM query, got %q", query)
	}
	if !strings.Contains(query, "pushed:>") {
		t.Errorf("Expected pushed qualifier in query, got %q", query)
	}
}

func TestBuildDiscoveryQuery_InvalidLLMQueryFallback(t *testing.T) {
	ctx := context.Background()
	mock := &fixedMockLLM{response: "topic:mcp language:Go"}
	query := BuildDiscoveryQuery(mock, ctx, 24*time.Hour)

	// Should reject only qualifiers and fallback to default
	if !strings.Contains(query, "model-context-protocol") {
		t.Errorf("Expected fallback to default query, got %q", query)
	}
}
