package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// DeepSeekLLMProvider implements LLMProvider by calling the DeepSeek API
// (or any OpenAI-compatible gateway such as Xiaomi MiMo / OpenRouter)
// via its /v1/chat/completions endpoint.
//
// This replaces Hermes or Mock with a direct DeepSeek API connection.
// Setup: set DEEPSEEK_API_KEY in your environment.
type DeepSeekLLMProvider struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// DeepSeekConfig holds the configuration for connecting to DeepSeek API.
type DeepSeekConfig struct {
	APIKey  string // DeepSeek API key (sk-...)
	Model   string // e.g. "deepseek-chat" or "deepseek-reasoner"
	BaseURL string // Override base URL (default: https://api.deepseek.com)
}

func normalizeOpenAIBaseRoot(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	trimmed = strings.TrimSuffix(trimmed, "/v1")
	return strings.TrimRight(trimmed, "/")
}

// NewDeepSeekLLMProvider creates a provider that calls DeepSeek directly.
func NewDeepSeekLLMProvider(cfg DeepSeekConfig) *DeepSeekLLMProvider {
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	baseRoot := normalizeOpenAIBaseRoot(cfg.BaseURL)
	if baseRoot == "" {
		baseRoot = "https://api.deepseek.com"
	}
	return &DeepSeekLLMProvider{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: baseRoot,
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

	baseRoot := normalizeOpenAIBaseRoot(d.BaseURL)
	if baseRoot == "" {
		baseRoot = "https://api.deepseek.com"
	}
	url := baseRoot + "/v1/chat/completions"
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

// HealthCheck verifies the DeepSeek / OpenAI-compatible API key is valid by
// probing /v1/models (with fallback to /models for gateways that mount at root).
func (d *DeepSeekLLMProvider) HealthCheck(ctx context.Context) error {
	if d.APIKey == "" {
		return fmt.Errorf("deepseek: API key not configured")
	}

	baseRoot := normalizeOpenAIBaseRoot(d.BaseURL)
	if baseRoot == "" {
		baseRoot = "https://api.deepseek.com"
	}

	paths := []string{"/v1/models", "/models"}
	var lastStatus int
	var lastErrType string

	for _, p := range paths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseRoot+p, nil)
		if err != nil {
			return fmt.Errorf("deepseek health: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+d.APIKey)

		resp, err := d.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("deepseek health: request failed: %w", err)
		}
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		lastStatus = resp.StatusCode
		var errPayload struct {
			Error struct {
				Type string `json:"type"`
				Code string `json:"code"`
			} `json:"error"`
		}
		if json.Unmarshal(bodyBytes, &errPayload) == nil {
			if errPayload.Error.Type != "" {
				lastErrType = errPayload.Error.Type
			} else if errPayload.Error.Code != "" {
				lastErrType = errPayload.Error.Code
			}
		}

		// Only try fallback /models path if /v1/models returned 404.
		if resp.StatusCode != http.StatusNotFound {
			break
		}
	}

	if lastErrType != "" {
		return fmt.Errorf("deepseek health: API returned status %d (%s)", lastStatus, lastErrType)
	}
	return fmt.Errorf("deepseek health: API returned status %d", lastStatus)
}

// compile-time interface check
var _ LLMProvider = (*DeepSeekLLMProvider)(nil)
