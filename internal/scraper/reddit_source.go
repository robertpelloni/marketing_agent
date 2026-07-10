package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/marketing_agent/internal/db"
)

type RedditSource struct {
	Client *http.Client
}

type redditListing struct {
	Data struct {
		Children []struct {
			Data struct {
				Title       string `json:"title"`
				SelfText    string `json:"selftext"`
				Author      string `json:"author"`
				Subreddit   string `json:"subreddit"`
				URL         string `json:"url"`
				CreatedUTC  float64 `json:"created_utc"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func NewRedditSource() *RedditSource {
	return &RedditSource{
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (r *RedditSource) Discover(ctx context.Context, keywords []string) ([]db.Company, error) {
	slog.Info("RedditSource: Starting discovery across subreddits")

	// Pre-defined target subreddits
	subreddits := []string{"MachineLearning", "LocalLLaMA", "artificial", "OpenAI", "devops"}
	var companies []db.Company

	for _, sub := range subreddits {
		url := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=10", sub)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			slog.Warn("RedditSource: Failed to create request", "error", err)
			continue
		}

		// Reddit requires a unique User-Agent
		req.Header.Set("User-Agent", "TormentNexus/1.0 (MarketingAgent; bot)")

		resp, err := r.Client.Do(req)
		if err != nil {
			slog.Warn("RedditSource: Failed to fetch subreddit", "subreddit", sub, "error", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			slog.Warn("RedditSource: Non-200 status", "subreddit", sub, "status", resp.StatusCode)
			continue
		}

		var listing redditListing
		if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
			resp.Body.Close()
			slog.Warn("RedditSource: Failed to parse JSON", "subreddit", sub, "error", err)
			continue
		}
		resp.Body.Close()

		for _, child := range listing.Data.Children {
			post := child.Data

			// Filter by keyword relevance
			isRelevant := false
			content := strings.ToLower(post.Title + " " + post.SelfText)
			for _, kw := range keywords {
				if strings.Contains(content, strings.ToLower(kw)) {
					isRelevant = true
					break
				}
			}

			if !isRelevant {
				continue
			}

			// Generate a pseudo-company mapping for the individual Redditor
			// since Reddit represents indie developers/creators usually.
			company := db.Company{
				Name:          fmt.Sprintf("Indie Dev: %s", post.Author),
				Domain:        fmt.Sprintf("reddit.com/user/%s", post.Author),
				MarketCapTier: "Startup", // Default to startup/indie for Reddit
				HiringSignals: []string{
					fmt.Sprintf("Reddit Post [%s]: %s - %s", post.Subreddit, post.Title, post.URL),
				},
				TechStack: []string{"Reddit Lead"},
			}

			companies = append(companies, company)
		}

		// Polite delay between subreddit requests
		time.Sleep(2 * time.Second)
	}

	slog.Info("RedditSource: Discovery complete", "found", len(companies))
	return companies, nil
}
