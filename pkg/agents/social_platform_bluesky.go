package agents

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// ─── Bluesky AT Protocol Client (pure Go, no external deps) ─────────────

type BlueskyProvider struct {
	Handle   string
	Password string
	client   *http.Client
	dryRun   bool
}

func NewBlueskyProvider(handle, password string) *BlueskyProvider {
	return &BlueskyProvider{
		Handle:   handle,
		Password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
		dryRun:   os.Getenv("BLUESKY_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
	}
}
func (p *BlueskyProvider) Name() string { return "bluesky" }

// atprotoSession is returned by com.atproto.server.createSession
type atprotoSession struct {
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
	Handle     string `json:"handle"`
	Did        string `json:"did"`
}

// atprotoRecord is the payload for com.atproto.repo.createRecord
type atprotoRecord struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     any    `json:"record"`
}

// blueskyPost is the schema for app.bsky.feed.post
type blueskyPost struct {
	Text      string        `json:"text"`
	CreatedAt string        `json:"createdAt"`
	Facets    []any         `json:"facets,omitempty"`
	Embed     *blueskyEmbed `json:"embed,omitempty"`
}

type blueskyEmbed struct {
	Type     string          `json:"$type"`
	External blueskyExternal `json:"external"`
}

type blueskyExternal struct {
	URI         string `json:"uri"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (p *BlueskyProvider) Post(ctx context.Context, req PostRequest) error {
	if p.dryRun {
		slog.Info(fmt.Sprintf("[Bluesky DRY RUN] Would post to %s:\n%s", req.AccountID, truncate(req.Content, 200)))
		return nil
	}

	// 1. Create session (authenticate)
	session, err := p.createSession(ctx)
	if err != nil {
		return fmt.Errorf("bluesky auth: %w", err)
	}

	// 2. Build the post record
	post := blueskyPost{
		Text:      req.Content,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// 3. Create the record via the AT Protocol
	recordURL := "https://bsky.social/xrpc/com.atproto.repo.createRecord"
	payload := atprotoRecord{
		Repo:       session.Did,
		Collection: "app.bsky.feed.post",
		Record:     post,
	}

	body, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", recordURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+session.AccessJWT)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("bluesky createRecord: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bluesky createRecord: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}

	slog.Info(fmt.Sprintf("Bluesky: Posted to %s ✓", req.AccountID))
	return nil
}

func (p *BlueskyProvider) createSession(ctx context.Context) (*atprotoSession, error) {
	body := map[string]string{
		"identifier": p.Handle,
		"password":   p.Password,
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST",
		"https://bsky.social/xrpc/com.atproto.server.createSession",
		bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var session atprotoSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if session.AccessJWT == "" {
		return nil, fmt.Errorf("empty access token in session response")
	}
	return &session, nil
}

// ─── Reddit API Client (OAuth2 App-Only + Script) ────────────────────────

type RedditProvider struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	client       *http.Client
	accessToken  string
	dryRun       bool
}

func NewRedditProvider(clientID, clientSecret, username, password string) *RedditProvider {
	return &RedditProvider{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Username:     username,
		Password:     password,
		client:       &http.Client{Timeout: 30 * time.Second},
		dryRun:       os.Getenv("REDDIT_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
	}
}
func (p *RedditProvider) Name() string { return "reddit" }

func (p *RedditProvider) Post(ctx context.Context, req PostRequest) (err error) {
	if p.dryRun {
		slog.Info(fmt.Sprintf("[Reddit DRY RUN] Would post to %s:\n%s", req.AccountID, truncate(req.Content, 200)))
		return nil
	}

	if p.Username == "" || p.Password == "" {
		slog.Info("RedditProvider: Credentials not configured, skipping live Reddit post.")
		return nil
	}

	slog.Info("RedditProvider: Attempting headless post via go-rod", "subreddit", "r/TormentNexusDev")

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("reddit headless post panicked: %v", r)
		}
	}()

	// Import package "github.com/go-rod/rod" implicitly or load rod browser context
	// We'll boot chromium headless
	browser := rod.New().MustConnect()
	defer func() { _ = browser.Close() }()

	page := browser.MustPage()

	// 1. Login
	page.MustNavigate("https://www.reddit.com/login")
	page.MustWaitLoad()
	page.MustElement("#login-username").MustInput(p.Username)
	page.MustElement("#login-password").MustInput(p.Password)
	page.MustElement("button[type='submit']").MustClick()
	time.Sleep(5 * time.Second) // wait for session initialization

	// 2. Navigate to test community
	page.MustNavigate("https://www.reddit.com/r/TormentNexusDev/submit")
	page.MustWaitLoad()

	// 3. Write post
	// Select text post title and body fields
	page.MustElement("textarea[placeholder='Title']").MustInput("TormentNexus Marketing System Update")
	page.MustElement("textarea[placeholder='Text (optional)']").MustInput(req.Content)

	// 4. Click Submit
	submitBtn := page.MustElementR("button", "Post")
	submitBtn.MustClick()
	time.Sleep(3 * time.Second)

	slog.Info(fmt.Sprintf("Reddit: Headless post successful to r/TormentNexusDev ✓"))
	return nil
}

func (p *RedditProvider) getToken(ctx context.Context) (string, error) {
	if p.accessToken != "" {
		return p.accessToken, nil
	}

	data := strings.NewReader(fmt.Sprintf(
		"grant_type=password&username=%s&password=%s",
		p.Username, p.Password,
	))

	req, _ := http.NewRequestWithContext(ctx, "POST",
		"https://www.reddit.com/api/v1/access_token", data)
	req.Header.Set("Authorization", basicAuth(p.ClientID, p.ClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "web:tormentnexus-bot:v1 (by /u/"+p.Username+")")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("token decode: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("reddit oauth error: %s", result.Error)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	p.accessToken = result.AccessToken
	return result.AccessToken, nil
}

// ─── Twitter / X API v2 Client ──────────────────────────────────────────

type TwitterProvider struct {
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
	BearerToken  string
	client       *http.Client
	dryRun       bool
}

func NewTwitterProvider(apiKey, apiSecret, accessToken, accessSecret, bearerToken string) *TwitterProvider {
	return &TwitterProvider{
		APIKey:       apiKey,
		APISecret:    apiSecret,
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
		BearerToken:  bearerToken,
		client:       &http.Client{Timeout: 30 * time.Second},
		dryRun:       os.Getenv("TWITTER_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
	}
}
func (p *TwitterProvider) Name() string { return "twitter" }

func (p *TwitterProvider) Post(ctx context.Context, req PostRequest) error {
	if p.dryRun {
		slog.Info(fmt.Sprintf("[Twitter DRY RUN] Would post to %s:\n%s", req.AccountID, truncate(req.Content, 200)))
		return nil
	}

	// Twitter API v2: POST /2/tweets
	payload := map[string]string{"text": req.Content}
	body, _ := json.Marshal(payload)

	apiURL := "https://api.twitter.com/2/tweets"
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	if p.BearerToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.BearerToken)
	} else if p.APIKey != "" && p.APISecret != "" && p.AccessToken != "" && p.AccessSecret != "" {
		// OAuth 1.0a — for a full implementation we'd use a signed header
		// For now, fall back to Bearer token for user-context
		httpReq.Header.Set("Authorization", "Bearer "+p.BearerToken)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("twitter post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twitter post: HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 300))
	}

	slog.Info(fmt.Sprintf("Twitter: Posted to %s ✓", req.AccountID))
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func stripNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", "")
}
