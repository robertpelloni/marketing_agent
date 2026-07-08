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

// DeepSeekLLMProvider implements LLMProvider by calling the DeepSeek API
// via its OpenAI-compatible /v1/chat/completions endpoint.
//
// This replaces Hermes or Mock with a direct DeepSeek API connection.
// Setup: set DEEPSEEK_API_KEY in your environment.
type DeepSeekLLMProvider struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// DeepSeekConfig holds the configuration for connecting to DeepSeek API.
type DeepSeekConfig struct {
	APIKey string // DeepSeek API key (sk-...)
	Model  string // e.g. "deepseek-chat" or "deepseek-reasoner"
}

// NewDeepSeekLLMProvider creates a provider that calls DeepSeek directly.
func NewDeepSeekLLMProvider(cfg DeepSeekConfig) *DeepSeekLLMProvider {
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	return &DeepSeekLLMProvider{
		APIKey: cfg.APIKey,
		Model:  cfg.Model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Generate sends a prompt to the DeepSeek API and returns the response.
func (d *DeepSeekLLMProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	messages := []chatMessage{}

	if prompt.System != "" {
		messages = append(messages, chatMessage{
			Role:    "system",
			Content: prompt.System,
		})
	}

	messages = append(messages, chatMessage{
		Role:    "user",
		Content: prompt.User,
	})

	reqBody := chatRequest{
		Model:    d.Model,
		Messages: messages,
	}

	if prompt.MaxTokens > 0 {
		reqBody.MaxTokens = prompt.MaxTokens
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to marshal request: %w", err)
	}

	url := "https://api.deepseek.com/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepseek: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("deepseek: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("deepseek: failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("deepseek: no choices in response")
	}

	content := chatResp.Choices[0].Message.Content

	slog.Info(fmt.Sprintf("DeepSeekLLM: model=%s tokens=%d+%d=%d finish=%s",
		d.Model,
		chatResp.Usage.PromptTokens,
		chatResp.Usage.CompletionTokens,
		chatResp.Usage.TotalTokens,
		chatResp.Choices[0].FinishReason),
	)

	return content, nil
}

// HealthCheck verifies the DeepSeek API key is valid by making a lightweight request.
func (d *DeepSeekLLMProvider) HealthCheck(ctx context.Context) error {
	if d.APIKey == "" {
		return fmt.Errorf("deepseek: API key not configured")
	}

	// Try fetching models list as a lightweight auth check
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.deepseek.com/models", nil)
	if err != nil {
		return fmt.Errorf("deepseek health: failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.APIKey)

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("deepseek health: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deepseek health: API returned status %d", resp.StatusCode)
	}

	return nil
}

// compile-time interface check
var _ LLMProvider = (*DeepSeekLLMProvider)(nil)
