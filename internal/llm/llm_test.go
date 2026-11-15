package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeepSeekLLMProvider_BaseURLNormalizationAndHealthCheck(t *testing.T) {
	var seenPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPaths = append(seenPaths, r.URL.Path)
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("Authorization") != "Bearer valid-key" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":{"type":"invalid_key","code":"401"}}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"mimo-v2.5"}]}`))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"pong"},"finish_reason":"stop"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Both root URL and /v1-suffixed URL must resolve cleanly without /v1/v1 duplication or /models 404.
	for _, rawBase := range []string{srv.URL, srv.URL + "/", srv.URL + "/v1", srv.URL + "/v1/"} {
		p := NewDeepSeekLLMProvider(DeepSeekConfig{
			APIKey:  "valid-key",
			Model:   "mimo-v2.5",
			BaseURL: rawBase,
		})
		if err := p.HealthCheck(context.Background()); err != nil {
			t.Fatalf("HealthCheck(%q) failed: %v", rawBase, err)
		}
		out, err := p.Generate(context.Background(), Prompt{User: "ping"})
		if err != nil || out != "pong" {
			t.Fatalf("Generate(%q) = %q, err=%v", rawBase, out, err)
		}
	}

	// Invalid key must surface 401 (invalid_key) rather than misleading 404.
	bad := NewDeepSeekLLMProvider(DeepSeekConfig{
		APIKey:  "expired-key",
		Model:   "mimo-v2.5",
		BaseURL: srv.URL,
	})
	err := bad.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "invalid_key") {
		t.Fatalf("expected 401 (invalid_key) health error, got: %v", err)
	}
}

func TestHermesLLMProvider_OpenAIGatewayFallbackHealthCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			if r.Header.Get("Authorization") != "Bearer hm-key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/v1/chat/completions":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	h := NewHermesLLMProvider(HermesConfig{
		BaseURL: srv.URL + "/v1",
		APIKey:  "hm-key",
		Model:   "mimo-v2.5",
	})
	if err := h.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected Hermes OpenAI-compatible gateway health check to pass via /v1/models fallback, got: %v", err)
	}
	out, err := h.Generate(context.Background(), Prompt{User: "hi"})
	if err != nil || out != "ok" {
		t.Fatalf("Hermes Generate failed: out=%q err=%v", out, err)
	}
}
