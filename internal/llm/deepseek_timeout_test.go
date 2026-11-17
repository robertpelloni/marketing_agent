package llm

import (
	"testing"
	"time"
)

func TestDeepSeekTimeoutConfigurable(t *testing.T) {
	p := NewDeepSeekLLMProvider(DeepSeekConfig{APIKey: "k", TimeoutSeconds: 300})
	if p.HTTPClient.Timeout != 300*time.Second {
		t.Errorf("expected 300s timeout, got %v", p.HTTPClient.Timeout)
	}

	def := NewDeepSeekLLMProvider(DeepSeekConfig{APIKey: "k"})
	if def.HTTPClient.Timeout != 120*time.Second {
		t.Errorf("expected default 120s timeout, got %v", def.HTTPClient.Timeout)
	}

	neg := NewDeepSeekLLMProvider(DeepSeekConfig{APIKey: "k", TimeoutSeconds: -5})
	if neg.HTTPClient.Timeout != 120*time.Second {
		t.Errorf("expected non-positive timeout to fall back to 120s, got %v", neg.HTTPClient.Timeout)
	}
}
