package communication

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
<<<<<<< HEAD

	"github.com/go-rod/rod"
=======
>>>>>>> origin/main
)

// LinkedInSender sends messages via LinkedIn messaging.
// Note: LinkedIn does not provide a public API for direct messaging.
// This implementation uses:
// - Simulation fallback when credentials are not configured
<<<<<<< HEAD
// - Headless browser automation (rod) for actual sending
type LinkedInSender struct {
	Username string
	Password string
=======
// - Placeholder for future headless browser automation (rod/chromedp)
// Production use requires LinkedIn API partnership or browser automation.
// Headless browser approach (future) will use chromedp/rod for:
// 1. Authenticating with LINKEDIN_USERNAME/LINKEDIN_PASSWORD
// 2. Navigating to recipient's LinkedIn profile
// 3. Clicking "Message" button
// 4. Typing and sending the message
// 5. Rate limiting and anti-detection measures
type LinkedInSender struct {
	Username	string
	Password	string
>>>>>>> origin/main
}

// LinkedInMessage represents a message to send via LinkedIn.
type LinkedInMessage struct {
<<<<<<< HEAD
	ProfileURL string // LinkedIn profile URL of the recipient
	Subject    string // Subject line (note subject)
	Body       string // Message body content
=======
	ProfileURL	string	// LinkedIn profile URL of the recipient
	Subject		string	// Subject line (note subject)
	Body		string	// Message body content
>>>>>>> origin/main
}

// NewLinkedInSender creates a new LinkedInSender.
func NewLinkedInSender() *LinkedInSender {
	return &LinkedInSender{
<<<<<<< HEAD
		Username: os.Getenv("LINKEDIN_USERNAME"),
		Password: os.Getenv("LINKEDIN_PASSWORD"),
=======
		Username:	os.Getenv("LINKEDIN_USERNAME"),
		Password:	os.Getenv("LINKEDIN_PASSWORD"),
>>>>>>> origin/main
	}
}

// Send sends a LinkedIn message to a recipient.
// Falls back to simulation when credentials are not configured.
func (l *LinkedInSender) Send(ctx context.Context, msg LinkedInMessage) error {
	if l.Username == "" || l.Password == "" {
		slog.Info("LinkedInSender: No LINKEDIN_USERNAME/PASSWORD configured, logging message (simulation)")
		return l.simulateSend(ctx, msg)
	}

<<<<<<< HEAD
	slog.Info("LinkedInSender: Credentials configured, attempting headless message send to", "profile", msg.ProfileURL)

	err := l.sendViaRod(ctx, msg)
	if err != nil {
		slog.Warn("LinkedInSender: Headless send failed, falling back to simulation", "error", err)
		return l.simulateSend(ctx, msg)
	}

	return nil
}

// sendViaRod uses the go-rod headless browser to send a message on LinkedIn.
func (l *LinkedInSender) sendViaRod(ctx context.Context, msg LinkedInMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("headless message send panicked: %v", r)
		}
	}()

	browser := rod.New().MustConnect()
	defer browser.Close()

	page := browser.MustPage()

	// 1. Navigate to login
	page.MustNavigate("https://www.linkedin.com/login")
	page.MustWaitLoad()

	// 2. Fill login form
	page.MustElement("#username").MustInput(l.Username)
	page.MustElement("#password").MustInput(l.Password)
	page.MustElement("button[type='submit']").MustClick()

	// Wait for successful login
	page.MustWaitElementsMoreThan(".global-nav__me-photo", 1)

	// 3. Navigate to profile URL
	page.MustNavigate(msg.ProfileURL)
	page.MustWaitLoad()

	// 4. Click message button
	// LinkedIn's message button classes change frequently. We look for a button containing "Message" text.
	// We might need to handle cases where we aren't connected yet.
	messageBtn := page.MustElementR("button", "Message")
	messageBtn.MustClick()

	// Wait for message box to appear
	page.MustWaitElementsMoreThan(".msg-form__contenteditable", 1)

	// 5. Type subject (if the form supports it, sometimes it doesn't for 1st degree connections)
	// We might need to skip subject if the element doesn't exist
	if hasSubject, _, _ := page.Has(".msg-form__subject"); hasSubject {
		page.MustElement(".msg-form__subject").MustInput(msg.Subject)
	}

	// 6. Type message body
	msgBox := page.MustElement(".msg-form__contenteditable")
	msgBox.MustInput(msg.Body)

	// 7. Send
	sendBtn := page.MustElement(".msg-form__send-button")
	sendBtn.MustClick()

	// 8. Sleep for rate limiting (rudimentary)
	time.Sleep(2 * time.Second)

	return nil
=======
	// Future: Implement real LinkedIn message sending via headless browser
	// This requires chromedp or rod for browser automation.
	// Implementation pattern:
	//   1. Navigate to linkedin.com/login
	//   2. Fill in username/password
	//   3. Handle 2FA if prompted
	//   4. Navigate to msg.ProfileURL
	//   5. Click message button
	//   6. Type message body
	//   7. Send
	//   8. Sleep for rate limiting
	slog.Info("LinkedInSender: Credentials configured but browser automation not yet implemented. Logging message.")
	return l.simulateSend(ctx, msg)
>>>>>>> origin/main
}

// simulateSend logs the message that would be sent.
func (l *LinkedInSender) simulateSend(ctx context.Context, msg LinkedInMessage) error {
	slog.Info(fmt.Sprintf(`LinkedInSender [SIMULATION] — Would send LinkedIn message:
  To: %s
  Subject: %s
  Body: %s
  Sent at: %s`,
		msg.ProfileURL,
		msg.Subject,
		msg.Body,
		time.Now().Format(time.RFC3339)),
	)
	return nil
}

// HealthCheck verifies LinkedIn credentials are configured.
func (l *LinkedInSender) HealthCheck(ctx context.Context) error {
	if l.Username == "" || l.Password == "" {
		return fmt.Errorf("LINKEDIN_USERNAME and LINKEDIN_PASSWORD must be configured for LinkedIn messaging")
	}
	slog.Info("LinkedInSender: Health check passed (credentials configured)")
	return nil
}

// SendConnectionRequest sends a LinkedIn connection request with a note.
<<<<<<< HEAD
=======
// Future implementation will use headless browser automation.
>>>>>>> origin/main
func (l *LinkedInSender) SendConnectionRequest(ctx context.Context, profileURL, note string) error {
	if l.Username == "" || l.Password == "" {
		slog.Info(fmt.Sprintf("LinkedInSender [SIMULATION] — Would send connection request to %s with note: %s", profileURL, note))
		return nil
	}

<<<<<<< HEAD
	slog.Info(fmt.Sprintf("LinkedInSender: Connection request to %s via headless browser", profileURL))
	err := l.connectViaRod(ctx, profileURL, note)
	if err != nil {
		slog.Warn("LinkedInSender: Headless connection request failed, falling back to simulation", "error", err)
		slog.Info(fmt.Sprintf("LinkedInSender [SIMULATION] — Would send connection request to %s with note: %s", profileURL, note))
	}

	return nil
}

// connectViaRod uses the go-rod headless browser to send a connection request.
func (l *LinkedInSender) connectViaRod(ctx context.Context, profileURL, note string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("headless connection request panicked: %v", r)
		}
	}()

	browser := rod.New().MustConnect()
	defer browser.Close()

	page := browser.MustPage()

	// 1. Navigate to login
	page.MustNavigate("https://www.linkedin.com/login")
	page.MustWaitLoad()

	// 2. Fill login form
	page.MustElement("#username").MustInput(l.Username)
	page.MustElement("#password").MustInput(l.Password)
	page.MustElement("button[type='submit']").MustClick()

	// Wait for successful login
	page.MustWaitElementsMoreThan(".global-nav__me-photo", 1)

	// 3. Navigate to profile URL
	page.MustNavigate(profileURL)
	page.MustWaitLoad()

	// 4. Click Connect button
	connectBtn := page.MustElementR("button", "Connect")
	connectBtn.MustClick()

	// Wait for modal
	page.MustWaitElementsMoreThan("#custom-message", 1)

	// 5. Add note
	page.MustElementR("button", "Add a note").MustClick()
	page.MustElement("#custom-message").MustInput(note)

	// 6. Send
	page.MustElementR("button", "Send").MustClick()

	// 7. Sleep for rate limiting (rudimentary)
	time.Sleep(2 * time.Second)

=======
	slog.Info(fmt.Sprintf("LinkedInSender: Connection request to %s — browser automation pending", profileURL))
>>>>>>> origin/main
	return nil
}
