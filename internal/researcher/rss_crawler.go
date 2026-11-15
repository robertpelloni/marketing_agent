package researcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RSSItem represents a single entry in an RSS 2.0 feed.
type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
}

type RSSChannel struct {
	Items []RSSItem `xml:"item"`
}

type RSS struct {
	Channel RSSChannel `xml:"channel"`
}

// AtomEntry represents a single entry in an Atom feed.
type AtomEntry struct {
	Title   string `xml:"title"`
	Summary string `xml:"summary"`
	Link    struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Updated string `xml:"updated"`
}

type Atom struct {
	Entries []AtomEntry `xml:"entry"`
}

type cachedFeedEntry struct {
	prefix  string
	title   string
	content string
	link    string
	pubTime time.Time
	hasTime bool
}

type cachedFeed struct {
	entries   []cachedFeedEntry
	expiresAt time.Time
}

// RSSFeedCrawler scans configured RSS/Atom feeds for hiring signals and company mentions.
// It implements the researcher.Crawler interface.
type RSSFeedCrawler struct {
	// HTTP client used for fetching feeds. If nil, http.DefaultClient is used.
	Client *http.Client
	// Feeds is a list of RSS/Atom feed URLs to scan.
	Feeds []string
	// Keywords used to detect hiring or interest signals.
	Keywords []string
	// MaxAge limits how old an article can be to be considered (e.g., 30 days).
	MaxAge time.Duration
	// CacheTTL controls how long parsed feed entries are cached in memory across
	// Crawl invocations (defaults to 30m; 429/5xx responses back off for 15m).
	CacheTTL time.Duration

	mu        sync.Mutex
	feedCache map[string]cachedFeed
}

// defaultKeywords returns a set of terms that commonly indicate hiring or technical interest.
func defaultKeywords() []string {
	return []string{"hiring", "engineer", "ml", "ai", "machine learning", "data science", "open position", "job", "career", "vacancy", "join us", "team", "role"}
}

func (r *RSSFeedCrawler) cacheTTL() time.Duration {
	if r.CacheTTL > 0 {
		return r.CacheTTL
	}
	return 30 * time.Minute
}

func (r *RSSFeedCrawler) loadFeedEntries(ctx context.Context, feedURL string) []cachedFeedEntry {
	now := time.Now()

	r.mu.Lock()
	if r.feedCache == nil {
		r.feedCache = make(map[string]cachedFeed)
	}
	if cached, ok := r.feedCache[feedURL]; ok && now.Before(cached.expiresAt) {
		entries := cached.entries
		r.mu.Unlock()
		return entries
	}
	r.mu.Unlock()

	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		slog.Info(fmt.Sprintf("RSSFeedCrawler: failed to create request for %s: %v", feedURL, err))
		return nil
	}
	req.Header.Set("User-Agent", "HyperNexus-ResearchBot/1.0 (+https://hypernexus.site)")

	resp, err := client.Do(req)
	if err != nil {
		slog.Info(fmt.Sprintf("RSSFeedCrawler: request error for %s: %v", feedURL, err))
		r.storeCachedFeed(feedURL, nil, 5*time.Minute)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		slog.Info(fmt.Sprintf("RSSFeedCrawler: non-200 status %d for %s", resp.StatusCode, feedURL))
		_ = resp.Body.Close()
		backoff := 5 * time.Minute
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			backoff = 15 * time.Minute
		}
		r.storeCachedFeed(feedURL, nil, backoff)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		slog.Info(fmt.Sprintf("RSSFeedCrawler: read error for %s: %v", feedURL, err))
		r.storeCachedFeed(feedURL, nil, 5*time.Minute)
		return nil
	}

	var entries []cachedFeedEntry
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err == nil && len(rss.Channel.Items) > 0 {
		for _, item := range rss.Channel.Items {
			entry := cachedFeedEntry{
				prefix:  "RSS",
				title:   item.Title,
				content: strings.ToLower(item.Title + " " + item.Description),
				link:    item.Link,
			}
			if item.PubDate != "" {
				if t, perr := parseDate(item.PubDate); perr == nil {
					entry.pubTime = t
					entry.hasTime = true
				}
			}
			entries = append(entries, entry)
		}
		r.storeCachedFeed(feedURL, entries, r.cacheTTL())
		return entries
	}

	var atom Atom
	if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entries) > 0 {
		for _, item := range atom.Entries {
			entry := cachedFeedEntry{
				prefix:  "Atom",
				title:   item.Title,
				content: strings.ToLower(item.Title + " " + item.Summary),
				link:    item.Link.Href,
			}
			if item.Updated != "" {
				if t, perr := parseDate(item.Updated); perr == nil {
					entry.pubTime = t
					entry.hasTime = true
				}
			}
			entries = append(entries, entry)
		}
		r.storeCachedFeed(feedURL, entries, r.cacheTTL())
		return entries
	}

	r.storeCachedFeed(feedURL, nil, r.cacheTTL())
	return nil
}

func (r *RSSFeedCrawler) storeCachedFeed(feedURL string, entries []cachedFeedEntry, ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.feedCache == nil {
		r.feedCache = make(map[string]cachedFeed)
	}
	r.feedCache[feedURL] = cachedFeed{
		entries:   entries,
		expiresAt: time.Now().Add(ttl),
	}
}

// Crawl implements the Crawler interface. It scans all configured feeds, looking for
// any article that mentions the target (company domain or name) and any of the hiring keywords.
// It returns a simple concatenated string of findings, or an empty string if none are found.
func (r *RSSFeedCrawler) Crawl(ctx context.Context, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" || strings.Contains(target, "@") {
		return "", nil
	}

	// Use supplied keywords if set, otherwise defaults.
	keywords := r.Keywords
	if len(keywords) == 0 {
		keywords = defaultKeywords()
	}
	// MaxAge defaults to 30 days if not set.
	maxAge := r.MaxAge
	if maxAge == 0 {
		maxAge = 30 * 24 * time.Hour
	}

	// Prepare for deduplication of matching items.
	var matches []string
	seen := make(map[string]struct{})
	targetLower := strings.ToLower(target)

	for _, feedURL := range r.Feeds {
		entries := r.loadFeedEntries(ctx, feedURL)
		for _, entry := range entries {
			if entry.hasTime && time.Since(entry.pubTime) > maxAge {
				continue
			}
			if strings.Contains(entry.content, targetLower) && containsKeyword(entry.content, keywords) {
				if _, ok := seen[entry.link]; ok {
					continue
				}
				seen[entry.link] = struct{}{}
				matches = append(matches, fmt.Sprintf("%s: %s – %s", entry.prefix, entry.title, entry.link))
			}
		}
	}

	if len(matches) == 0 {
		return "", nil
	}
	// Join findings with newlines.
	return strings.Join(matches, "\n"), nil
}

// containsKeyword reports true if any of the supplied keywords appear in the text.
func containsKeyword(text string, kw []string) bool {
	textLower := strings.ToLower(text)
	for _, k := range kw {
		if strings.Contains(textLower, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

// parseDate attempts to parse common RSS/Atom date formats.
func parseDate(dateStr string) (time.Time, error) {
	// Try multiple known formats.
	layouts := []string{
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00", // fallback same as RFC3339
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	// If parsing fails, return zero time with error.
	return time.Time{}, fmt.Errorf("unable to parse date %s", dateStr)
}
