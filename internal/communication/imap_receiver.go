package communication

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

// IMAPConfig holds the configuration for IMAP email polling.
type IMAPConfig struct {
	Host     string // e.g. "imap.gmail.com"
	Port     int    // e.g. 993 for SSL
	Username string // email address
	Password string // app password
	Folder   string // mailbox to poll (default: "INBOX")
}

// InboundEmail represents a parsed inbound email from IMAP.
type InboundEmail struct {
	From      string // sender email address
	Subject   string
	Body      string // plain text body (decoded from MIME)
	MessageID string // RFC 5322 Message-ID header when present
	Date      time.Time
	UID       uint32 // IMAP UID for tracking
}

// EmailReceiver polls an IMAP inbox for new inbound emails.
type EmailReceiver struct {
	config    IMAPConfig
	manager   *Manager
	lastUID   uint32 // tracks the last processed message UID
	seededUID bool
	connected bool
}

// NewEmailReceiver creates a new IMAP email receiver.
func NewEmailReceiver(cfg IMAPConfig, manager *Manager) *EmailReceiver {
	if cfg.Port == 0 {
		cfg.Port = 993
	}
	if cfg.Folder == "" {
		cfg.Folder = "INBOX"
	}
	return &EmailReceiver{
		config:  cfg,
		manager: manager,
	}
}

// Run starts the IMAP polling loop. It checks for new emails at the specified interval.
func (r *EmailReceiver) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info(fmt.Sprintf("IMAP: Email receiver started (polling %s every %v)", r.config.Host, interval))

	// Run an initial poll shortly after startup so bounces/replies are caught promptly.
	if err := r.pollNewEmails(ctx); err != nil {
		slog.Info(fmt.Sprintf("IMAP: Initial poll error: %v", err))
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("IMAP: Email receiver stopping...")
			return
		case <-ticker.C:
			if err := r.pollNewEmails(ctx); err != nil {
				slog.Info(fmt.Sprintf("IMAP: Poll error: %v", err))
			}
		}
	}
}

func (r *EmailReceiver) ensureSeededUID(ctx context.Context) {
	if r.seededUID {
		return
	}
	r.seededUID = true
	if r.manager != nil && r.manager.db != nil {
		if maxUID, err := r.manager.db.GetMaxProcessedInboundUID(ctx, r.config.Folder); err == nil && maxUID > r.lastUID {
			r.lastUID = maxUID
			slog.Info(fmt.Sprintf("IMAP: Restored last processed UID=%d for folder %s from database", r.lastUID, r.config.Folder))
		}
	}
}

// pollNewEmails connects to IMAP, fetches unread messages, and processes them idempotently.
func (r *EmailReceiver) pollNewEmails(ctx context.Context) error {
	r.ensureSeededUID(ctx)

	c, err := r.connect()
	if err != nil {
		return fmt.Errorf("imap: connect failed: %w", err)
	}
	defer r.disconnect(c)

	// Select the mailbox read-only (BODY.PEEK[] preserves user unread state while
	// processed_inbound_emails guarantees strict idempotency across restarts).
	mbox, err := c.Select(r.config.Folder, true)
	if err != nil {
		return fmt.Errorf("imap: select %s failed: %w", r.config.Folder, err)
	}

	if mbox.Messages == 0 {
		return nil
	}

	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	if r.lastUID > 0 {
		criteria.Uid = new(imap.SeqSet)
		criteria.Uid.AddRange(r.lastUID+1, 0)
	}

	rawUIDs, err := c.UidSearch(criteria)
	if err != nil {
		return fmt.Errorf("imap: search failed: %w", err)
	}

	// RFC 3501 §6.4.4 quirk: `UID SEARCH UID <lastUID+1>:*` normalizes `<lastUID+1>:*`
	// to `<maxUID>:<lastUID+1>` when no newer messages exist, returning `<maxUID>` again!
	// Strictly filter `uid > r.lastUID` so already-seen high-water UIDs are never re-fetched.
	var uids []uint32
	for _, uid := range rawUIDs {
		if uid > r.lastUID {
			uids = append(uids, uid)
		}
	}

	if len(uids) == 0 {
		return nil
	}

	// Process up to 50 unread messages per cycle so DSN bounces cannot starve human replies.
	if len(uids) > 50 {
		uids = uids[len(uids)-50:]
	}

	slog.Info(fmt.Sprintf("IMAP: Found %d new unread messages above UID %d", len(uids), r.lastUID))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	bodySection := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, bodySection.FetchItem()}
	messages := make(chan *imap.Message, len(uids))

	go func() {
		if err := c.UidFetch(seqSet, items, messages); err != nil {
			slog.Info(fmt.Sprintf("IMAP: Fetch error: %v", err))
		}
	}()

	var maxUID uint32
	for msg := range messages {
		if msg == nil {
			continue
		}
		if msg.Uid > maxUID {
			maxUID = msg.Uid
		}

		email := r.parseMessage(msg, bodySection)
		if email == nil {
			continue
		}

		r.processInboundEmail(ctx, *email)
	}

	if maxUID > r.lastUID {
		r.lastUID = maxUID
	}

	return nil
}

// parseMessage extracts an InboundEmail (including full decoded MIME body) from an IMAP message.
func (r *EmailReceiver) parseMessage(msg *imap.Message, bodySection *imap.BodySectionName) *InboundEmail {
	if msg.Envelope == nil {
		return nil
	}

	from := ""
	if len(msg.Envelope.From) > 0 {
		addr := msg.Envelope.From[0]
		from = fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName)
	}
	from = strings.ToLower(strings.TrimSpace(from))
	if from == "" || from == "@" {
		return nil
	}

	var rawBytes []byte
	if bodySection != nil {
		if lit := msg.GetBody(bodySection); lit != nil {
			rawBytes, _ = io.ReadAll(io.LimitReader(lit, 256*1024))
		}
	}
	if len(rawBytes) == 0 && msg.Body != nil {
		for _, lit := range msg.Body {
			if lit == nil {
				continue
			}
			data, err := io.ReadAll(io.LimitReader(lit, 256*1024))
			if err == nil && len(data) > 0 {
				rawBytes = data
				break
			}
		}
	}

	plainBody, msgID := ParseRawMIMEBody(rawBytes)
	subject := decodeMIMEHeader(msg.Envelope.Subject)

	return &InboundEmail{
		From:      from,
		Subject:   subject,
		Body:      plainBody,
		MessageID: msgID,
		Date:      msg.Envelope.Date,
		UID:       msg.Uid,
	}
}

func (r *EmailReceiver) messageDedupKey(email InboundEmail) string {
	if id := strings.TrimSpace(email.MessageID); id != "" {
		return "msgid:" + strings.ToLower(id)
	}
	return fmt.Sprintf("imap:%s:%d", strings.ToUpper(strings.TrimSpace(r.config.Folder)), email.UID)
}

// processInboundEmail feeds a received email or DSN bounce into the sales pipeline idempotently.
func (r *EmailReceiver) processInboundEmail(ctx context.Context, email InboundEmail) {
	msgKey := r.messageDedupKey(email)

	if r.manager != nil && r.manager.db != nil {
		alreadyDone, err := r.manager.db.IsInboundEmailProcessed(ctx, msgKey)
		if err == nil && alreadyDone {
			return
		}
	}

	slog.Info(fmt.Sprintf("IMAP: Processing email UID=%d from %s (subject: %s)", email.UID, email.From, email.Subject))

	// 1. Handle DSN / Mailer-Daemon bounce notifications first (before sender contact lookup,
	// because bounces arrive from mailer-daemon@... rather than the prospect's email address).
	bouncedAddrs, isDSN, isTempDelay := ExtractPermanentBounceRecipients(email.From, email.Subject, email.Body, r.config.Username)
	if isDSN {
		if isTempDelay {
			slog.Info(fmt.Sprintf("IMAP: Temporary delivery delay DSN (UID=%d) — leaving contact active", email.UID))
			r.markProcessed(ctx, msgKey, email, "dsn_delay_ignored")
			return
		}

		closedCount := 0
		if r.manager != nil && r.manager.db != nil {
			for _, bounced := range bouncedAddrs {
				contact, err := r.manager.db.GetContactByEmail(ctx, bounced)
				if err != nil || contact == nil {
					continue
				}
				if deal, dealErr := r.manager.db.GetDealByCompanyID(ctx, contact.CompanyID); dealErr == nil && deal != nil {
					_ = r.manager.db.UpdateDealState(ctx, deal.ID, db.StateClosedLost)
				}
				_ = r.manager.db.UpdateContactEmail(ctx, contact.ID, "")
				closedCount++
				slog.Info(fmt.Sprintf("IMAP: Permanent bounce processed for %s (contact %d) — cleared email and closed deal", bounced, contact.ID))
			}
		}
		r.markProcessed(ctx, msgKey, email, fmt.Sprintf("dsn_bounce_processed:%d", closedCount))
		return
	}

	if r.manager == nil || r.manager.db == nil {
		return
	}

	// 2. Look up the contact by sender email address.
	contact, err := r.manager.db.GetContactByEmail(ctx, email.From)
	if err != nil || contact == nil {
		slog.Info(fmt.Sprintf("IMAP: No contact found for %s — marking non-pipeline message processed", email.From))
		r.markProcessed(ctx, msgKey, email, "ignored_not_in_pipeline")
		return
	}

	// 3. Feed full Subject + decoded MIME Body into the communication manager.
	text := email.Subject
	if strings.TrimSpace(email.Body) != "" {
		text = fmt.Sprintf("Subject: %s\n\n%s", email.Subject, email.Body)
	}

	reply, err := r.manager.ProcessInbound(ctx, *contact, text)
	if err != nil {
		slog.Info(fmt.Sprintf("IMAP: Error generating reply for %s (inbound interaction persisted): %v", email.From, err))
		// Inbound interaction and state transition are already persisted at the top of ProcessInbound;
		// mark the IMAP message processed so transient LLM errors do not duplicate Inbound rows every poll.
		r.markProcessed(ctx, msgKey, email, "inbound_recorded_reply_err")
		return
	}

	if reply != "" {
		slog.Info(fmt.Sprintf("IMAP: Generated reply to %s: %s", email.From, truncate(reply, 100)))
		r.markProcessed(ctx, msgKey, email, "replied")
		return
	}
	r.markProcessed(ctx, msgKey, email, "processed_no_reply")
}

func (r *EmailReceiver) markProcessed(ctx context.Context, msgKey string, email InboundEmail, outcome string) {
	if r.manager == nil || r.manager.db == nil {
		return
	}
	_ = r.manager.db.MarkInboundEmailProcessed(ctx, msgKey, r.config.Folder, email.UID, email.From, email.Subject, outcome)
}

var (
	htmlScriptStyleRe = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	htmlBlockBreakRe  = regexp.MustCompile(`(?i)</?(p|div|br|li|tr|h[1-6])[^>]*>`)
	htmlTagRe         = regexp.MustCompile(`(?s)<[^>]+>`)
	emailExtractRe    = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z0-9\-]{2,}`)
)

func decodeMIMEHeader(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	dec := new(mime.WordDecoder)
	if decoded, err := dec.DecodeHeader(trimmed); err == nil {
		return decoded
	}
	return trimmed
}

// ParseRawMIMEBody parses a raw RFC 5322 email payload and returns its decoded
// plain-text body (including delivery-status diagnostic blocks for DSNs) and Message-ID.
func ParseRawMIMEBody(raw []byte) (string, string) {
	if len(raw) == 0 {
		return "", ""
	}

	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return strings.TrimSpace(string(raw)), ""
	}

	msgID := strings.TrimSpace(msg.Header.Get("Message-ID"))
	if msgID == "" {
		msgID = strings.TrimSpace(msg.Header.Get("Message-Id"))
	}

	var plainParts []string
	var htmlParts []string
	if failedHdr := strings.TrimSpace(msg.Header.Get("X-Failed-Recipients")); failedHdr != "" {
		plainParts = append(plainParts, "X-Failed-Recipients: "+failedHdr)
	}

	extractEntityParts(msg.Header.Get("Content-Type"), msg.Header.Get("Content-Transfer-Encoding"), msg.Body, &plainParts, &htmlParts, 0)

	if len(plainParts) > 0 {
		return strings.TrimSpace(strings.Join(plainParts, "\n\n")), msgID
	}
	if len(htmlParts) > 0 {
		return strings.TrimSpace(stripHTMLToText(strings.Join(htmlParts, "\n\n"))), msgID
	}
	return "", msgID
}

func extractEntityParts(contentType, transferEncoding string, body io.Reader, plainParts, htmlParts *[]string, depth int) {
	if depth > 8 || body == nil {
		return
	}

	mediaType := "text/plain"
	var params map[string]string
	if strings.TrimSpace(contentType) != "" {
		if mt, p, err := mime.ParseMediaType(contentType); err == nil {
			mediaType = strings.ToLower(mt)
			params = p
		}
	}

	if strings.HasPrefix(mediaType, "multipart/") && params["boundary"] != "" {
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			ct := part.Header.Get("Content-Type")
			cte := part.Header.Get("Content-Transfer-Encoding")
			disp := strings.ToLower(part.Header.Get("Content-Disposition"))
			if strings.HasPrefix(disp, "attachment") && !strings.Contains(strings.ToLower(ct), "delivery-status") {
				_ = part.Close()
				continue
			}
			extractEntityParts(ct, cte, part, plainParts, htmlParts, depth+1)
			_ = part.Close()
		}
		return
	}

	decodedReader := decodeTransferEncoding(body, transferEncoding)
	data, err := io.ReadAll(io.LimitReader(decodedReader, 128*1024))
	if err != nil || len(data) == 0 {
		return
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return
	}

	switch {
	case mediaType == "text/plain" || mediaType == "message/delivery-status" || mediaType == "text/rfc822-headers":
		*plainParts = append(*plainParts, text)
	case mediaType == "text/html":
		*htmlParts = append(*htmlParts, text)
	}
}

func decodeTransferEncoding(r io.Reader, encoding string) io.Reader {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "quoted-printable":
		return quotedprintable.NewReader(r)
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, r)
	default:
		return r
	}
}

func stripHTMLToText(s string) string {
	clean := htmlScriptStyleRe.ReplaceAllString(s, " ")
	clean = htmlBlockBreakRe.ReplaceAllString(clean, "\n")
	clean = htmlTagRe.ReplaceAllString(clean, " ")
	clean = html.UnescapeString(clean)
	lines := strings.Split(clean, "\n")
	var out []string
	for _, ln := range lines {
		t := strings.Join(strings.Fields(ln), " ")
		if t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n")
}

// ExtractPermanentBounceRecipients inspects an inbound email to determine if it is a
// Mailer-Daemon / DSN notification, whether it is only a temporary delay, and which
// target recipient addresses permanently bounced.
func ExtractPermanentBounceRecipients(from, subject, fullBody string, selfEmails ...string) ([]string, bool, bool) {
	lowerFrom := strings.ToLower(strings.TrimSpace(from))
	lowerSubj := strings.ToLower(strings.TrimSpace(subject))
	combined := lowerSubj + "\n" + strings.ToLower(fullBody)

	isDSNSender := strings.HasPrefix(lowerFrom, "mailer-daemon@") ||
		strings.HasPrefix(lowerFrom, "postmaster@")
	isDSNContent := strings.Contains(lowerSubj, "delivery status notification") ||
		strings.Contains(lowerSubj, "undelivered mail") ||
		strings.Contains(lowerSubj, "mail delivery failed") ||
		strings.Contains(lowerSubj, "returned mail") ||
		strings.Contains(lowerSubj, "failure notice") ||
		(isDSNSender && isBounce(combined))

	if !isDSNSender && !isDSNContent {
		return nil, false, false
	}

	// Temporary deferral DSNs (e.g., "Delivery Status Notification (Delay)", Status: 4.x.x)
	// should NOT wipe a valid contact's email while the MTA is still retrying.
	isDelay := strings.Contains(lowerSubj, "(delay)") ||
		strings.Contains(combined, "action: delayed") ||
		strings.Contains(combined, "will retry for") ||
		(strings.Contains(combined, "delivery incomplete") && !strings.Contains(combined, "permanent"))
	if isDelay {
		return nil, true, true
	}

	excluded := map[string]struct{}{
		lowerFrom: {},
	}
	for _, self := range selfEmails {
		s := strings.ToLower(strings.TrimSpace(self))
		if s != "" {
			excluded[s] = struct{}{}
			if local, _, ok := strings.Cut(s, "@"); ok && local != "" {
				excluded[local+"@gmail.com"] = struct{}{}
				excluded[local+"@googlemail.com"] = struct{}{}
			}
		}
	}

	seen := map[string]struct{}{}
	var recipients []string
	for _, match := range emailExtractRe.FindAllString(fullBody, -1) {
		cand := strings.ToLower(strings.Trim(match, ".,;:<>()[]\"'"))
		if cand == "" {
			continue
		}
		if _, skip := excluded[cand]; skip {
			continue
		}
		if strings.HasPrefix(cand, "mailer-daemon@") ||
			strings.HasPrefix(cand, "postmaster@") ||
			strings.HasPrefix(cand, "noreply@") ||
			strings.HasSuffix(cand, "@mx.google.com") ||
			strings.HasSuffix(cand, "@google.com") {
			continue
		}
		if _, exists := seen[cand]; !exists {
			seen[cand] = struct{}{}
			recipients = append(recipients, cand)
		}
	}

	return recipients, true, false
}

// connect establishes an IMAP connection.
func (r *EmailReceiver) connect() (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", r.config.Host, r.config.Port)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("imap: TLS dial to %s failed: %w", addr, err)
	}

	if err := c.Login(r.config.Username, r.config.Password); err != nil {
		_ = c.Logout()
		return nil, fmt.Errorf("imap: login failed: %w", err)
	}

	r.connected = true
	return c, nil
}

// disconnect closes the IMAP connection gracefully.
func (r *EmailReceiver) disconnect(c *client.Client) {
	if c != nil {
		_ = c.Logout()
		r.connected = false
	}
}

// HealthCheck verifies IMAP connectivity.
func (r *EmailReceiver) HealthCheck(ctx context.Context) error {
	c, err := r.connect()
	if err != nil {
		return err
	}
	defer r.disconnect(c)
	return nil
}

// MockEmailReceiver is a no-op email receiver for testing.
type MockEmailReceiver struct{}

func (m *MockEmailReceiver) Run(ctx context.Context, interval time.Duration) {
	slog.Info("MockEmailReceiver: Running (no-op)")
	<-ctx.Done()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
