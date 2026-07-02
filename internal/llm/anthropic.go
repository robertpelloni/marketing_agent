package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// AnthropicLLMProvider implements LLMProvider by calling the Anthropic API.
type AnthropicLLMProvider struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// AnthropicConfig holds the configuration for connecting to the Anthropic API.
type AnthropicConfig struct {
	APIKey string
	Model  string
}

// NewAnthropicLLMProvider creates a provider that routes LLM calls through Anthropic.
func NewAnthropicLLMProvider(cfg AnthropicConfig) *AnthropicLLMProvider {
	model := cfg.Model
	if model == "" {
		model = "claude-3-opus-20240229"
	}
	return &AnthropicLLMProvider{
		APIKey: cfg.APIKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // LLM calls can be slow
		},
	}
}

// anthropicMessage represents a single message in the Anthropic messages format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest is the request body for the Messages API.
type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
}

// anthropicResponse is the response body from the Messages API.
type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Generate sends a prompt to the Anthropic API and returns the LLM response.
func (a *AnthropicLLMProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	messages := []anthropicMessage{
		{
			Role:    "user",
			Content: prompt.User,
		},
	}

	maxTokens := prompt.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	reqBody := anthropicRequest{
		Model:     a.Model,
		Messages:  messages,
		System:    prompt.System,
		MaxTokens: maxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp anthropicResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("anthropic: failed to decode response: %w", err)
	}

	if len(chatResp.Content) == 0 {
		return "", fmt.Errorf("anthropic: no choices in response")
	}

	content := chatResp.Content[0].Text

	slog.Info(fmt.Sprintf("AnthropicLLM: model=%s tokens=%d+%d=%d",
		a.Model,
		chatResp.Usage.InputTokens,
		chatResp.Usage.OutputTokens,
		chatResp.Usage.InputTokens+chatResp.Usage.OutputTokens),
	)

	return content, nil
}
