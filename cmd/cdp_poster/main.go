package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// cdp_poster — Local Windows companion that posts to Twitter/X and Reddit
// via Chrome DevTools Protocol. Bypasses both the Twitter $100/mo paywall
// and Reddit's dead API. Just needs Chrome running locally.
//
// Usage:
//   chrome --remote-debugging-port=9222
//   cdp_poster -platform twitter
//   cdp_poster -platform reddit
//   cdp_poster -platform both

var (
	contentURL  = flag.String("url", "http://5.161.250.43:8084/api/v1/social/reddit", "Content API URL")
	platform    = flag.String("platform", "both", "Platform to post: twitter, reddit, both")
	chromePort  = flag.Int("chrome", 9222, "Chrome remote debugging port")
	twitterUser = flag.String("tw-user", "", "Twitter username/email")
	twitterPass = flag.String("tw-pass", "", "Twitter password")
	redditUser  = flag.String("rd-user", "", "Reddit username")
	redditPass  = flag.String("rd-pass", "", "Reddit password")
	dryRun      = flag.Bool("dry-run", false, "Print content but don't post")
)

type ContentResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Brand   string `json:"brand"`
}

func main() {
	flag.Parse()

	// Load credentials from env if not passed as flags
	if *twitterUser == "" {
		*twitterUser = os.Getenv("CDP_TWITTER_USER")
	}
	if *twitterPass == "" {
		*twitterPass = os.Getenv("CDP_TWITTER_PASS")
	}
	if *redditUser == "" {
		*redditUser = os.Getenv("REDDIT_USERNAME")
	}
	if *redditPass == "" {
		*redditPass = os.Getenv("REDDIT_PASSWORD")
	}

	// Fetch content from VPS
	content, err := fetchContent(*contentURL)
	if err != nil {
		log.Printf("⚠ Failed to fetch content from VPS: %v — using fallback", err)
		content = &ContentResponse{
			Title:   "TormentNexus — The OS for AI Models",
			Content: "Progressive MCP tool routing, cross-harness parity, LLM waterfall, 14K+ memories. Open source: github.com/HyperNexusSoft/HyperNexus",
			Brand:   "tormentnexus",
		}
	}

	log.Printf("📋 Content: %s", truncate(content.Content, 100))

	if *dryRun {
		log.Printf("🔍 DRY RUN — would post to: %s\nContent: %s", *platform, content.Content)
		return
	}

	// Connect to local Chrome
	log.Printf("🔌 Connecting to Chrome on port %d...", *chromePort)
	browser, err := connectToChrome(*chromePort)
	if err != nil {
		log.Fatalf("❌ Cannot connect to Chrome: %v\n   Start Chrome with: chrome --remote-debugging-port=%d", err, *chromePort)
	}
	defer browser.Close()
	log.Println("✅ Connected to Chrome")

	ctx := context.Background()

	switch *platform {
	case "twitter", "both":
		if *twitterUser == "" {
			log.Println("⚠ Twitter credentials not set — skipping (set CDP_TWITTER_USER and CDP_TWITTER_PASS)")
		} else {
			postToTwitter(ctx, browser, *twitterUser, *twitterPass, content)
		}
	}

	switch *platform {
	case "reddit", "both":
		if *redditUser == "" {
			log.Println("⚠ Reddit credentials not set — skipping (set REDDIT_USERNAME and REDDIT_PASSWORD)")
		} else {
			postToReddit(ctx, browser, *redditUser, *redditPass, content)
		}
	}
}

// currentHour returns the current hour modulo 24 for subreddit rotation.
func currentHour() int {
	return time.Now().Hour()
}

// varyTitle slightly rephrases the title to avoid identical cross-posts.
var titleVariants = []string{
	"[Showcase] %s",
	"%s — feedback welcome",
	"Built this: %s",
	"%s — thoughts?",
	"Sharing my project: %s",
	"%s [Open Source]",
}

func varyTitle(baseTitle, subreddit string) string {
	variant := titleVariants[rand.Intn(len(titleVariants))]
	title := fmt.Sprintf(variant, baseTitle)
	// Keep it under 300 chars (Reddit limit)
	if len(title) > 290 {
		title = title[:287] + "..."
	}
	return title
}

func fetchContent(url string) (*ContentResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c ContentResponse
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, fmt.Errorf("decode: %w (body: %s)", err, string(body[:100]))
	}
	return &c, nil
}

func connectToChrome(port int) (*rod.Browser, error) {
	// Try connecting to existing Chrome with remote debugging
	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	resp, err := http.Get(url + "/json/version")
	if err != nil {
		// Chrome not running with remote debugging — try launching it
		log.Println("Chrome not running with remote debugging, launching new instance...")
		u := launcher.New().
			Headless(false).
			RemoteDebuggingPort(port).
			MustLaunch()
		return rod.New().ControlURL(u).MustConnect(), nil
	}
	defer resp.Body.Close()

	return rod.New().ControlURL(launcher.MustResolveURL(url)).MustConnect(), nil
}

func postToTwitter(ctx context.Context, browser *rod.Browser, username, password string, content *ContentResponse) {
	log.Println("🐦 Posting to Twitter/X...")

	page := browser.MustPage()
	defer page.MustClose()

	// Navigate to compose
	page.MustNavigate("https://x.com/compose/post")
	page.MustWaitLoad()

	// Check if we need to login
	time.Sleep(2 * time.Second)
	hasLogin, _, _ := page.Has("input[autocomplete='username']")
	if hasLogin {
		log.Println("   Logging into Twitter...")
		page.MustElement("input[autocomplete='username']").MustInput(username)
		page.MustElementR("span", "Next").MustClick()
		time.Sleep(2 * time.Second)

		// Check for unusual activity
		if has, _, _ := page.Has("input[name='password']"); has {
			page.MustElement("input[name='password']").MustInput(password)
			page.MustElementR("span", "Log in").MustClick()
		} else {
			// Might have phone/email verification step
			hasVerify, _, _ := page.Has("input[data-testid='ocfEnterTextTextInput']")
			if hasVerify {
				log.Println("   ⚠ Twitter verification requested — check your phone/email")
			}
			page.MustElement("input[name='password']").MustInput(password)
			page.MustElementR("span", "Log in").MustClick()
		}
		time.Sleep(3 * time.Second)
		page.MustNavigate("https://x.com/compose/post")
		page.MustWaitLoad()
		time.Sleep(2 * time.Second)
	}

	// Type the tweet
	tweetText := truncate(content.Content, 280)
	editor := page.MustElement("div[data-testid='tweetTextarea_0']")
	editor.MustClick()
	editor.MustInput(tweetText)
	time.Sleep(1 * time.Second)

	// Click Tweet button
	postBtn := page.MustElementR("span", "Post")
	postBtn.MustClick()
	time.Sleep(3 * time.Second)

	log.Printf("🐦 ✅ Posted to Twitter: %s", truncate(tweetText, 80))
}

func postToReddit(ctx context.Context, browser *rod.Browser, username, password string, content *ContentResponse) {
	log.Println("🔴 Posting to Reddit...")

	// Rotate through AI/LLM/coding subreddits — one per cycle
	subreddits := []string{
		"LocalLLaMA", "selfhosted", "MachineLearning", "artificial",
		"LLMDevs", "AI_Agents", "opensource", "coolgithubprojects",
		"golang", "programming",
	}
	subreddit := subreddits[currentHour()%len(subreddits)]

	// Add timeout so we don't hang forever
	postCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	page := browser.MustPage()
	defer page.MustClose()

	// Try old.reddit.com first (simpler DOM)
	page.MustNavigate("https://old.reddit.com/login")
	page.MustWaitLoad()

	// Check if already logged in (old Reddit shows username in header)
	alreadyLoggedIn := false
	if el, err := page.Element(".user a"); err == nil && el != nil {
		txt, _ := el.Text()
		if txt != "" && txt != "log in" && txt != "sign up" {
			alreadyLoggedIn = true
			log.Printf("   Already logged in as %s", txt)
		}
	}

	if !alreadyLoggedIn {
		log.Println("   Logging into Reddit via old.reddit.com...")
		// old.reddit.com uses name="user" and name="passwd"
		if el, err := page.Element("input[name='user']"); err == nil && el != nil {
			el.MustInput(username)
		} else {
			log.Println("   ⚠ Could not find login form — cookies may have expired")
			return
		}
		page.MustElement("input[name='passwd']").MustInput(password)
		page.MustElement("button[type='submit']").MustClick()
		time.Sleep(3 * time.Second)

		// Verify login succeeded
		if el, err := page.Element(".user a"); err != nil || el == nil {
			log.Println("   ⚠ Login failed — check credentials or CAPTCHA")
			return
		}
	}

	// Navigate to submit
	page.MustNavigate(fmt.Sprintf("https://old.reddit.com/r/%s/submit", subreddit))
	page.MustWaitLoad()

	// Vary the title so each post looks unique
	title := varyTitle(content.Title, subreddit)

	// Fill and submit
	page.MustElement("input[name='title']").MustInput(title)
	page.MustElement("textarea[name='text']").MustInput(content.Content)
	page.MustElement("button[type='submit']").MustClick()
	time.Sleep(3 * time.Second)

	log.Printf("🔴 ✅ Posted to r/%s: %s", subreddit, truncate(title, 80))
	_ = postCtx // timeout context handles cleanup
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
