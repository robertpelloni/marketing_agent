package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
<<<<<<< HEAD
=======
	"regexp"
>>>>>>> origin/main
	"strings"
	"time"
)

// HermesLLMProvider implements LLMProvider by calling a Hermes Agent API server
// via its OpenAI-compatible /v1/chat/completions endpoint.
//
// This replaces MockLLMProvider with a real LLM backed by Hermes's provider
// routing (OpenRouter, NVIDIA NIM, LM Studio, Ollama, etc.).
//
// Setup: set HERMES_API_URL and HERMES_API_KEY in your environment.
// The Hermes gateway must be running with api_server enabled.
type HermesLLMProvider struct {
<<<<<<< HEAD
	BaseURL		string
	APIKey		string
	Model		string
	HTTPClient	*http.Client
=======
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
>>>>>>> origin/main
}

// HermesConfig holds the configuration for connecting to a Hermes API server.
type HermesConfig struct {
<<<<<<< HEAD
	BaseURL	string	// e.g. "http://172.21.116.32:8642"
	APIKey	string	// the API_SERVER_KEY set in Hermes .env
	Model	string	// e.g. "free-llm" or any model available in Hermes
=======
	BaseURL string // e.g. "http://172.21.116.32:8642"
	APIKey  string // the API_SERVER_KEY set in Hermes .env
	Model   string // e.g. "free-llm" or any model available in Hermes
>>>>>>> origin/main
}

// NewHermesLLMProvider creates a provider that routes LLM calls through Hermes.
func NewHermesLLMProvider(cfg HermesConfig) *HermesLLMProvider {
	return &HermesLLMProvider{
<<<<<<< HEAD
		BaseURL:	strings.TrimRight(cfg.BaseURL, "/"),
		APIKey:		cfg.APIKey,
		Model:		cfg.Model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,	// LLM calls can be slow
=======
		BaseURL: strings.TrimRight(cfg.BaseURL, "/"),
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // LLM calls can be slow
>>>>>>> origin/main
		},
	}
}

// chatMessage represents a single message in the OpenAI chat format.
type chatMessage struct {
<<<<<<< HEAD
	Role	string	`json:"role"`
	Content	string	`json:"content"`
=======
	Role    string `json:"role"`
	Content string `json:"content"`
>>>>>>> origin/main
}

// chatRequest is the request body for /v1/chat/completions.
type chatRequest struct {
<<<<<<< HEAD
	Model		string		`json:"model"`
	Messages	[]chatMessage	`json:"messages"`
	MaxTokens	int		`json:"max_tokens,omitempty"`
=======
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
>>>>>>> origin/main
}

// chatResponse is the response body from /v1/chat/completions.
type chatResponse struct {
<<<<<<< HEAD
	Choices	[]struct {
		Message	struct {
			Content string `json:"content"`
		}	`json:"message"`
		FinishReason	string	`json:"finish_reason"`
	}	`json:"choices"`
	Usage	struct {
		PromptTokens		int	`json:"prompt_tokens"`
		CompletionTokens	int	`json:"completion_tokens"`
		TotalTokens		int	`json:"total_tokens"`
	}	`json:"usage"`
=======
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
>>>>>>> origin/main
}

// Generate sends a prompt to the Hermes API server and returns the LLM response.
func (h *HermesLLMProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	messages := []chatMessage{}

	if prompt.System != "" {
		messages = append(messages, chatMessage{
<<<<<<< HEAD
			Role:		"system",
			Content:	prompt.System,
=======
			Role:    "system",
			Content: prompt.System,
>>>>>>> origin/main
		})
	}

	messages = append(messages, chatMessage{
<<<<<<< HEAD
		Role:		"user",
		Content:	prompt.User,
	})

	reqBody := chatRequest{
		Model:		h.Model,
		Messages:	messages,
=======
		Role:    "user",
		Content: prompt.User,
	})

	reqBody := chatRequest{
		Model:    h.Model,
		Messages: messages,
>>>>>>> origin/main
	}

	if prompt.MaxTokens > 0 {
		reqBody.MaxTokens = prompt.MaxTokens
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("hermes: failed to marshal request: %w", err)
	}

	url := h.BaseURL + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("hermes: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.APIKey)

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("hermes: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("hermes: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("hermes: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("hermes: failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("hermes: no choices in response")
	}

	content := chatResp.Choices[0].Message.Content

<<<<<<< HEAD
=======
	// Strip reasoning prefixes from OpenCode Zen / FreeLLM responses
	// e.g. "[Model: north-mini-code-free | Provider: opencode_zen]\n\n"
	content = stripReasoningPrefix(content)

>>>>>>> origin/main
	slog.Info(fmt.Sprintf("HermesLLM: model=%s tokens=%d+%d=%d finish=%s",
		h.Model,
		chatResp.Usage.PromptTokens,
		chatResp.Usage.CompletionTokens,
		chatResp.Usage.TotalTokens,
		chatResp.Choices[0].FinishReason),
	)

	return content, nil
}

<<<<<<< HEAD
=======
var reasoningPrefixRegex = regexp.MustCompile(`(?s)\[Model:[^\]]+\]\s*\n\s*`)
var reasoningContinuedRegex = regexp.MustCompile(`(?s)\[Continued with Model:[^\]]+\]\s*\n\s*`)
var providerPrefixRegex = regexp.MustCompile(`(?s)^[A-Z][a-z]+\s+thinks?[^\n]*\n\s*`)

// stripReasoningPrefix removes model/provider reasoning prefixes that some
// OpenCode Zen / FreeLLM models prepend before the actual response.
func stripReasoningPrefix(s string) string {
	s = reasoningPrefixRegex.ReplaceAllString(s, "")
	s = reasoningContinuedRegex.ReplaceAllString(s, "")
	s = providerPrefixRegex.ReplaceAllString(s, "")
	// Remove leading newlines/spaces after stripping
	for len(s) > 0 && (s[0] == '\n' || s[0] == '\r' || s[0] == ' ') {
		s = s[1:]
	}
	return strings.TrimSpace(s)
}

>>>>>>> origin/main
// HealthCheck verifies the Hermes API server is reachable and responding.
func (h *HermesLLMProvider) HealthCheck(ctx context.Context) error {
	url := h.BaseURL + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("hermes health: failed to create request: %w", err)
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("hermes health: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hermes health: API returned status %d", resp.StatusCode)
	}

	return nil
}
