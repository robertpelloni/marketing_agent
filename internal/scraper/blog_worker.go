package scraper

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// RSSFeed represents a basic RSS feed structure.
type RSSFeed struct {
	Items []RSSItem `xml:"channel>item"`
}

// RSSItem represents a single item in an RSS feed.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// BlogWorker periodically polls engineering blogs for technical signals.
type BlogWorker struct {
	db     *db.DB
	client *http.Client
	feeds  []string
}

// NewBlogWorker creates a new BlogWorker.
func NewBlogWorker(database *db.DB) *BlogWorker {
	return &BlogWorker{
		db:     database,
		client: &http.Client{Timeout: 30 * time.Second},
		feeds: []string{
			"https://netflixtechblog.com/feed",
			"https://eng.uber.com/feed/",
			"https://doordash.engineering/feed/",
			"https://engineering.fb.com/feed/",
			"https://medium.com/feed/airbnb-engineering",
		},
	}
}

// Run starts the blog ingestion loop.
func (bw *BlogWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("BlogWorker: Started engineering blog monitor.")

	// Initial run
	bw.pollFeeds(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bw.pollFeeds(ctx)
		}
	}
}

func (bw *BlogWorker) pollFeeds(ctx context.Context) {
	for _, feedURL := range bw.feeds {
		if err := bw.processFeed(ctx, feedURL); err != nil {
			log.Printf("BlogWorker Error: Failed to process feed %s: %v", feedURL, err)
		}
	}
}

func (bw *BlogWorker) processFeed(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := bw.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return err
	}

	for _, item := range feed.Items {
		bw.analyzePost(ctx, item)
	}

	return nil
}

func (bw *BlogWorker) analyzePost(ctx context.Context, item RSSItem) {
	text := strings.ToLower(item.Title + " " + item.Description)

	// Keywords signaling technical bottlenecks or relevance to TormentNexus
	signals := []string{
		"orchestration", "bottleneck", "latency", "scalability",
		"microservices", "infrastructure", "ai platform", "llm",
		"agent", "workflow", "state management", "distributed systems",
	}

	foundSignals := []string{}
	for _, s := range signals {
		if strings.Contains(text, s) {
			foundSignals = append(foundSignals, s)
		}
	}

	if len(foundSignals) > 0 {
		// In a real system, we'd map the feed URL back to a company in our DB.
		// For now, we log the discovery.
		log.Printf("BlogWorker: Discovered high-value signal at %s: %v", item.Link, foundSignals)

		// If we can identify the company from the link, we'd update its hiring signals or dossier.
		// Example logic placeholder:
		// if company := bw.db.FindCompanyByBlogURL(item.Link); company != nil {
		//     bw.db.AddHiringSignal(ctx, company.ID, fmt.Sprintf("Engineering Blog: %s", item.Title))
		// }
	}
}
