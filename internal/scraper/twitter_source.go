package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// TwitterSource implements LeadSource by searching Twitter/X API v2 for
// companies discussing AI infrastructure pain points.
type TwitterSource struct {
	Client            *http.Client
	BearerToken       string
	APIKey            string
	APIKeySecret      string
	AccessToken       string
	AccessTokenSecret string
}

const twitterAPIBase = "https://api.twitter.com/2"

// tweet represents a tweet from the Twitter API v2 response.
type tweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
}

// twitterUser represents a user from the Twitter API v2 includes.
type twitterUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Bio      string `json:"description,omitempty"`
}

// twitterResponse is the top-level response from the Twitter API v2.
type twitterResponse struct {
	Data     []tweet `json:"data"`
	Includes struct {
		Users []twitterUser `json:"users"`
	} `json:"includes"`
	Meta struct {
		ResultCount int `json:"result_count"`
	} `json:"meta"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// searchQueries are the Twitter search queries to find potential leads.
var searchQueries = []string{
	`"model context protocol" OR "mcp server" OR "mcp tool" lang:en -is:reply`,
	`"ai orchestration" OR "llm orchestration" OR "agent orchestration" lang:en -is:reply`,
	`"multi-agent" OR "agentic" AND "infrastructure" lang:en -is:reply`,
	`"llm routing" OR "model routing" OR "provider fallback" lang:en -is:reply`,
	`"tool calling" OR "function calling" AND "llm" lang:en -is:reply`,
	`"vector search" OR "semantic search" AND "production" lang:en -is:reply`,
	`"ai platform" OR "ml platform" AND "scaling" lang:en -is:reply`,
	`"hiring" AND ("ai engineer" OR "llm engineer" OR "ml engineer") lang:en -is:reply`,
}

// Discover searches Twitter/X API for companies discussing AI infrastructure.
func (t *TwitterSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("TwitterSource: Searching Twitter API for AI/LLM signals...")

	if t.BearerToken == "" {
		slog.Info("TwitterSource: No bearer token configured, using simulated results")
		return t.simulate(ctx, keywords)
	}

	companies, err := t.searchTwitterAPI(ctx)
	if err != nil {
		slog.Info(fmt.Sprintf("TwitterSource: API search failed (%v), falling back to simulation", err))
		return t.simulate(ctx, keywords)
	}

	if len(companies) == 0 {
		slog.Info("TwitterSource: No companies found from API, using simulation")
		return t.simulate(ctx, keywords)
	}

	slog.Info(fmt.Sprintf("TwitterSource: Discovered %d companies from Twitter API", len(companies)))
	return companies, nil
}

// searchTwitterAPI searches the Twitter API v2 for relevant tweets.
func (t *TwitterSource) searchTwitterAPI(ctx context.Context) ([]db.Company, error) {
	seen := make(map[string]bool)
	var companies []db.Company

	for _, query := range searchQueries {
		if len(companies) >= 15 {
			break
		}

		tweets, users, err := t.searchTweets(ctx, query)
		if err != nil {
			slog.Info(fmt.Sprintf("TwitterSource: Query %q failed: %v", query[:30], err))
			continue
		}

		userMap := make(map[string]twitterUser)
		for _, u := range users {
			userMap[u.ID] = u
		}

		for _, tw := range tweets {
			if len(companies) >= 15 {
				break
			}

			user, hasUser := userMap[tw.AuthorID]
			if !hasUser {
				continue
			}

			// Extract potential company name from user name or bio
			companyName := extractCompanyName(user.Name, user.Bio)
			if companyName == "" || seen[companyName] {
				continue
			}

			domain := extractDomain(user.Bio)
			if domain == "" {
				domain = strings.ToLower(user.Username) + ".io"
			}

			techStack := extractTechStackFromText(tw.Text + " " + user.Bio)
			seen[companyName] = true

			companies = append(companies, db.Company{
				Name:          companyName,
				Domain:        domain,
				TechStack:     techStack,
				HiringSignals: []string{truncateText(tw.Text, 200)},
				MarketCapTier: classifyMarketCapFromText(tw.Text + " " + user.Bio),
			})
		}
	}

	return companies, nil
}

// searchTweets makes a single Twitter API v2 search request.
func (t *TwitterSource) searchTweets(ctx context.Context, query string) ([]tweet, []twitterUser, error) {
	url := fmt.Sprintf("%s/tweets/search/recent?query=%s&max_results=10&expansions=author_id&user.fields=description,name,username&tweet.fields=created_at",
		twitterAPIBase, urlQueryEncode(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+t.BearerToken)

	resp, err := t.client().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("Twitter API returned status %d", resp.StatusCode)
	}

	var result twitterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, err
	}

	if len(result.Errors) > 0 {
		return nil, nil, fmt.Errorf("Twitter API errors: %v", result.Errors)
	}

	return result.Data, result.Includes.Users, nil
}

// extractCompanyName tries to extract a company name from a Twitter user's profile.
func extractCompanyName(name, bio string) string {
	// Check bio for common company indicators
	bioLower := strings.ToLower(bio)
	lines := strings.Split(bio, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "@") || strings.Contains(line, "founder") || strings.Contains(line, "building") || strings.Contains(line, "ceo") || strings.Contains(line, "cto") {
			// Try to extract company from common patterns
			parts := strings.Split(line, "|")
			if len(parts) > 1 {
				candidate := strings.TrimSpace(parts[len(parts)-1])
				if len(candidate) > 2 && len(candidate) < 50 {
					return candidate
				}
			}
		}
	}

	// Try to find "at X" pattern
	if idx := strings.LastIndex(bioLower, " at "); idx > 0 {
		after := strings.TrimSpace(bio[idx+4:])
		if len(after) > 2 && len(after) < 40 && !strings.Contains(after, " ") {
			return after
		}
	}

	// Fall back to the display name if it looks like a company
	if len(name) > 5 && !strings.Contains(strings.ToLower(name), " ") {
		return name
	}

	return ""
}

// extractDomain tries to find a domain in the user's profile.
func extractDomain(bio string) string {
	// Check for URLs in bio
	lower := strings.ToLower(bio)
	if idx := strings.Index(lower, "http"); idx >= 0 {
		remainder := bio[idx:]
		// Extract domain from URL
		start := strings.Index(remainder, "://")
		if start >= 0 {
			remainder = remainder[start+3:]
		}
		if end := strings.Index(remainder, "/"); end >= 0 {
			remainder = remainder[:end]
		}
		if end := strings.Index(remainder, " "); end >= 0 {
			remainder = remainder[:end]
		}
		remainder = strings.TrimPrefix(remainder, "www.")
		if strings.Contains(remainder, ".") {
			return remainder
		}
	}

	return ""
}

// extractTechStackFromText parses tech keywords from text.
func extractTechStackFromText(text string) []string {
	lower := strings.ToLower(text)
	var stack []string
	techMap := map[string]string{
		"go": "Go", "golang": "Go", "rust": "Rust", "python": "Python",
		"typescript": "TypeScript", "kubernetes": "Kubernetes", "k8s": "Kubernetes",
		"docker": "Docker", "aws": "AWS", "gcp": "GCP", "azure": "Azure",
		"postgres": "PostgreSQL", "redis": "Redis", "kafka": "Kafka",
		"pytorch": "PyTorch", "tensorflow": "TensorFlow", "langchain": "LangChain",
		"llm": "LLMs", "openai": "OpenAI", "anthropic": "Anthropic",
		"grpc": "gRPC", "graphql": "GraphQL", "react": "React",
	}
	seen := make(map[string]bool)
	for kw, display := range techMap {
		if strings.Contains(lower, kw) && !seen[display] {
			seen[display] = true
			stack = append(stack, display)
		}
	}
	return stack
}

// classifyMarketCapFromText estimates company tier from text.
func classifyMarketCapFromText(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "series") || strings.Contains(lower, "seed") || strings.Contains(lower, "yc ") || strings.Contains(lower, "startup") {
		return "Startup"
	}
	if strings.Contains(lower, "enterprise") || strings.Contains(lower, "fortune") {
		return "Enterprise"
	}
	return "Startup"
}

// truncateText truncates text to maxLen.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// urlQueryEncode encodes a query string for a URL.
func urlQueryEncode(q string) string {
	q = strings.ReplaceAll(q, " ", "%20")
	q = strings.ReplaceAll(q, "\"", "%22")
	q = strings.ReplaceAll(q, ":", "%3A")
	q = strings.ReplaceAll(q, "-", "%2D")
	return q
}

// simulate returns empty — no mock data. Only real API results are used.
func (t *TwitterSource) simulate(_ context.Context, _ []string) ([]db.Company, error) {
	return nil, nil
}

func (t *TwitterSource) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}
