package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// FallbackProvider implements LLMProvider by routing requests through a prioritized list of providers.
type FallbackProvider struct {
	providers []LLMProvider
	names     []string
}

// NewFallbackProvider creates a provider that tries multiple LLMs in order until one succeeds.
func NewFallbackProvider(providers []LLMProvider, names []string) *FallbackProvider {
	return &FallbackProvider{
		providers: providers,
		names:     names,
	}
}

// Generate tries to generate a response, falling back to the next provider on failure.
func (f *FallbackProvider) Generate(ctx context.Context, prompt Prompt) (string, error) {
	var errs []string

	for i, provider := range f.providers {
		name := f.names[i]
		slog.Debug(fmt.Sprintf("LLMFallback: Attempting generation with provider %s", name))

		res, err := provider.Generate(ctx, prompt)
		if err == nil {
			return res, nil
		}

		slog.Warn(fmt.Sprintf("LLMFallback: Provider %s failed: %v", name, err))
		errs = append(errs, fmt.Sprintf("%s: %v", name, err))
	}

	return "", fmt.Errorf("llm fallback: all providers failed: %s", strings.Join(errs, ", "))
}
