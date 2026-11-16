package web

import (
	"strings"
	"testing"
)

func TestProviderPanelListReflectsLaneFlags(t *testing.T) {
	t.Setenv("TWITTER_DISABLED", "true")
	t.Setenv("TWITTER_API_KEY", "present")
	t.Setenv("GITHUB_POSTING_DISABLED", "true")
	t.Setenv("GITHUB_TOKEN", "present")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("HERMES_API_KEY", "")
	t.Setenv("HERMES_API_URL", "")

	html := providerPanelList("stripe-ok")
	if !strings.Contains(html, "X/Twitter (social posting):") || !strings.Contains(html, "disabled via TWITTER_DISABLED") {
		t.Errorf("expected X lane to report disabled, got: %s", html)
	}
	if !strings.Contains(html, "GitHub comment posting:") || !strings.Contains(html, "disabled via GITHUB_POSTING_DISABLED") {
		t.Errorf("expected GitHub posting to report disabled, got: %s", html)
	}
	if !strings.Contains(html, "Apollo/Hunter:") {
		t.Errorf("expected retired Apollo/Hunter line, got: %s", html)
	}
}

func TestProviderPanelListTwitterEnabledWhenConfigured(t *testing.T) {
	t.Setenv("TWITTER_DISABLED", "")
	t.Setenv("TWITTER_API_KEY", "present")
	t.Setenv("GITHUB_POSTING_DISABLED", "")
	t.Setenv("GITHUB_TOKEN", "present")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("HERMES_API_KEY", "")
	t.Setenv("HERMES_API_URL", "")

	html := providerPanelList("stripe-ok")
	if strings.Contains(html, "disabled via TWITTER_DISABLED") {
		t.Errorf("expected X lane not disabled, got: %s", html)
	}
	if !strings.Contains(html, "X/Twitter (social posting):") {
		t.Errorf("expected X lane line, got: %s", html)
	}
}
