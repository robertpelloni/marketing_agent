package agents

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/robertpelloni/marketing_agent/internal/communication"
)

// ContentType indicates the style/mood of a social post.
type ContentType string

const (
	ContentTech ContentType = "tech"
	ContentBiz  ContentType = "business"
)

// PostRequest holds everything needed to publish on a platform.
type PostRequest struct {
	Brand     string // "tormentnexus" or "hypernexus"
	Platform  string // "bluesky", "reddit", "twitter", "linkedin"
	Content   string // The body text
	AccountID string // The account name / identifier
}

// SocialProvider posts content to a single platform.
type SocialProvider interface {
	Post(ctx context.Context, req PostRequest) error
	Name() string
}

// SocialPosterBlog orchestrates multiple SocialProviders.
type SocialPosterHub struct {
	providers map[string]SocialProvider
	accounts  map[string]map[string]string // brand -> platform -> accountID
}

// NewSocialPosterHub builds the hub, auto-detecting which platforms are configured.
func NewSocialPosterHub() *SocialPosterHub {
	hub := &SocialPosterHub{
		providers: make(map[string]SocialProvider),
		accounts:  defaultAccounts(),
	}

	register := func(p SocialProvider) {
		slog.Info(fmt.Sprintf("SocialPosterHub: %s provider ready", p.Name()))
		hub.providers[p.Name()] = p
	}

	// Bluesky (Go AT Protocol client — uses env BLUESKY_HANDLE, BLUESKY_APP_PASSWORD)
	if h := os.Getenv("BLUESKY_HANDLE"); h != "" {
		register(NewBlueskyProvider(h, os.Getenv("BLUESKY_APP_PASSWORD")))
	} else {
		slog.Info("SocialPosterHub: Bluesky not configured (set BLUESKY_HANDLE + BLUESKY_APP_PASSWORD)")
	}

	// Reddit (uses env REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET, REDDIT_USERNAME, REDDIT_PASSWORD)
	if id := os.Getenv("REDDIT_CLIENT_ID"); id != "" {
		register(NewRedditProvider(
			id, os.Getenv("REDDIT_CLIENT_SECRET"),
			os.Getenv("REDDIT_USERNAME"), os.Getenv("REDDIT_PASSWORD"),
		))
	} else {
		slog.Info("SocialPosterHub: Reddit not configured (set REDDIT_CLIENT_ID + REDDIT_CLIENT_SECRET + REDDIT_USERNAME + REDDIT_PASSWORD)")
	}

	// Twitter / X (uses env TWITTER_BEARER_TOKEN, or TWITTER_API_KEY + TWITTER_API_SECRET + TWITTER_ACCESS_TOKEN + TWITTER_ACCESS_SECRET)
	if bt := os.Getenv("TWITTER_BEARER_TOKEN"); bt != "" || os.Getenv("TWITTER_API_KEY") != "" {
		register(NewTwitterProvider(
			os.Getenv("TWITTER_API_KEY"), os.Getenv("TWITTER_API_SECRET"),
			os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_SECRET"),
			os.Getenv("TWITTER_BEARER_TOKEN"),
		))
	} else {
		slog.Info("SocialPosterHub: Twitter not configured (set TWITTER_BEARER_TOKEN or TWITTER_API_KEY + ...)")
	}

	// LinkedIn — uses LINKEDIN_USERNAME, LINKEDIN_PASSWORD (headless browser via go-rod)
	if u := os.Getenv("LINKEDIN_USERNAME"); u != "" {
		register(NewLinkedInProvider(communication.NewLinkedInSender()))
	} else {
		slog.Info("SocialPosterHub: LinkedIn not configured (set LINKEDIN_USERNAME + LINKEDIN_PASSWORD)")
	}

	return hub
}

// Provider returns the SocialProvider for the named platform, or nil if not configured.
func (h *SocialPosterHub) Provider(name string) SocialProvider {
	return h.providers[name]
}

// Post sends content to every configured provider for the given brand.
// It returns a summary of what was actually posted vs simulated.
func (h *SocialPosterHub) Post(ctx context.Context, brand string) []string {
	var results []string
	platforms := []string{"bluesky", "reddit", "twitter", "linkedin"}

	for _, plat := range platforms {
		provider, ok := h.providers[plat]
		if !ok {
			results = append(results, fmt.Sprintf("%s: not configured (skipped)", plat))
			continue
		}

		acct, ok := h.accounts[brand][plat]
		if !ok {
			acct = brand + "-" + plat
		}

		req := PostRequest{
			Brand:     brand,
			Platform:  plat,
			Content:   "", // filled in by caller
			AccountID: acct,
		}

		if err := provider.Post(ctx, req); err != nil {
			results = append(results, fmt.Sprintf("%s: POST FAILED — %v", plat, err))
		} else {
			results = append(results, fmt.Sprintf("%s: posted to %s", plat, acct))
		}
	}
	return results
}

// Registered returns the list of configured platform names.
func (h *SocialPosterHub) Registered() []string {
	var names []string
	for _, p := range h.providers {
		names = append(names, p.Name())
	}
	return names
}

func defaultAccounts() map[string]map[string]string {
	return map[string]map[string]string{
		"tormentnexus": {
			"bluesky":  "tormentnexus.bsky.social",
			"reddit":   "MDMAtk",
			"reddit":   "tormentnexus-reddit",
			"twitter":  "@tormentnexus",
			"linkedin": "TormentNexus Community",
		},
		"hypernexus": {
			"bluesky":  "hypernexus.bsky.social",
			"reddit":   "MDMAtk",
			"reddit":   "hypernexus-reddit",
			"twitter":  "@hypernexus_site",
			"linkedin": "HyperNexus Enterprise",
		},
	}
}

// ─── LinkedIn Provider (wraps existing LinkedInSender from communication) ───

type LinkedInProvider struct {
	sender *communication.LinkedInSender
}

func NewLinkedInProvider(s *communication.LinkedInSender) *LinkedInProvider {
	return &LinkedInProvider{sender: s}
}
func (p *LinkedInProvider) Name() string { return "linkedin" }
func (p *LinkedInProvider) Post(ctx context.Context, req PostRequest) error {
	msg := communication.LinkedInMessage{
		ProfileURL: "https://www.linkedin.com/feed/", // posts to own feed
		Subject:    "",
		Body:       req.Content,
	}
	return p.sender.Send(ctx, msg)
}

// ─── Provider Error Helpers ──────────────────────────────────────

func truncate(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n-1]) + "…"
}
