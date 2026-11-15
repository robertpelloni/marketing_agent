package db

import "strings"

var exactBlockedEmails = map[string]struct{}{
	"codex@openai.com":       {},
	"noreply@openai.com":     {},
	"claude@anthropic.com":   {},
	"noreply@anthropic.com":  {},
	"cursoragent@cursor.com": {},
	"agent@cursor.com":       {},
	"noreply@cursor.com":     {},
	"devin@cognition.ai":     {},
	"noreply@cognition.ai":   {},
	"actions@github.com":     {},
	"noreply@github.com":     {},
	"support@github.com":     {},
	"bot@renovateapp.com":    {},
	"bot@mend.io":            {},
}

var exactBlockedDomains = map[string]struct{}{
	"localhost":                {},
	"localdomain":              {},
	"example.com":              {},
	"example.org":              {},
	"example.net":              {},
	"test.com":                 {},
	"none.com":                 {},
	"invalid":                  {},
	"github.io":                {},
	"users.noreply.github.com": {},
	"noreply.github.com":       {},
	"renovateapp.com":          {},
	"dependabot.com":           {},
}

var blockedDomainSuffixes = []string{
	".github.io",
	".noreply.github.com",
	".local",
	".localhost",
	".internal",
	".invalid",
	".test",
	".example",
	".lan",
}

var exactBlockedLocalParts = map[string]struct{}{
	"bot":                  {},
	"dependabot":           {},
	"renovate":             {},
	"renovate-bot":         {},
	"codecov":              {},
	"codecov-io":           {},
	"greenkeeper":          {},
	"snyk-bot":             {},
	"imgbot":               {},
	"allcontributors":      {},
	"semantic-release-bot": {},
	"pyup-bot":             {},
	"web-flow":             {},
	"webflow":              {},
	"github-actions":       {},
	"support":              {},
	"admin":                {},
	"info":                 {},
	"security":             {},
	"abuse":                {},
	"postmaster":           {},
	"mailer-daemon":        {},
	"webmaster":            {},
	"hostmaster":           {},
	"automated":            {},
	"ci":                   {},
}

var personalEmailDomains = map[string]struct{}{
	"gmail.com":      {},
	"googlemail.com": {},
	"icloud.com":     {},
	"me.com":         {},
	"mac.com":        {},
	"yahoo.com":      {},
	"ymail.com":      {},
	"hotmail.com":    {},
	"outlook.com":    {},
	"live.com":       {},
	"msn.com":        {},
	"aol.com":        {},
	"proton.me":      {},
	"protonmail.com": {},
	"pm.me":          {},
	"gmx.com":        {},
	"gmx.de":         {},
	"mail.com":       {},
	"fastmail.com":   {},
	"hey.com":        {},
	"zoho.com":       {},
	"yandex.com":     {},
	"yandex.ru":      {},
	"qq.com":         {},
	"163.com":        {},
	"126.com":        {},
	"att.net":        {},
	"comcast.net":    {},
	"verizon.net":    {},
	"sbcglobal.net":  {},
}

// IsAutomatedOrUnroutableEmail reports whether email belongs to an AI coding
// agent, CI/dependency bot, system/no-reply mailbox, or non-routable domain
// (such as *.github.io Pages hosts, *.local, or example.com).
func IsAutomatedOrUnroutableEmail(email string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(email))
	if cleaned == "" {
		return true
	}
	if strings.ContainsAny(cleaned, " \t\r\n<>\"'") {
		return true
	}
	if strings.Count(cleaned, "@") != 1 {
		return true
	}

	if _, blocked := exactBlockedEmails[cleaned]; blocked {
		return true
	}

	parts := strings.SplitN(cleaned, "@", 2)
	local, domain := parts[0], parts[1]
	if local == "" || domain == "" {
		return true
	}
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return true
	}

	if _, blocked := exactBlockedDomains[domain]; blocked {
		return true
	}
	for _, suffix := range blockedDomainSuffixes {
		if strings.HasSuffix(domain, suffix) {
			return true
		}
	}

	if _, blocked := exactBlockedLocalParts[local]; blocked {
		return true
	}

	for _, token := range []string{"noreply", "no-reply", "donotreply", "do-not-reply", "[bot]", "github-actions", "mailer-daemon"} {
		if strings.Contains(local, token) {
			return true
		}
	}

	// Match explicit delimited bot handles (e.g. snyk-bot, bot-ci) without
	// rejecting human surnames ending in "bot" (e.g. talbot, abbott, cabot).
	if strings.HasSuffix(local, "-bot") || strings.HasSuffix(local, "_bot") || strings.HasSuffix(local, ".bot") ||
		strings.HasPrefix(local, "bot-") || strings.HasPrefix(local, "bot_") || strings.HasPrefix(local, "bot.") {
		return true
	}

	return false
}

// IsAutomatedCommitAuthor reports whether a Git commit author name/email pair
// belongs to a bot, CI runner, or autonomous coding agent rather than a human.
func IsAutomatedCommitAuthor(name, email string) bool {
	if IsAutomatedOrUnroutableEmail(email) {
		return true
	}

	lowerName := strings.ToLower(strings.TrimSpace(name))
	if lowerName == "" {
		return false
	}
	if strings.Contains(lowerName, "[bot]") {
		return true
	}

	switch lowerName {
	case "github-actions", "github actions",
		"dependabot", "dependabot-preview",
		"renovate", "renovate bot", "renovate-bot",
		"semantic-release-bot", "snyk-bot", "snyk bot",
		"codecov", "codecov-io", "greenkeeper",
		"openai codex", "claude code", "cursor agent", "cursoragent", "devin ai":
		return true
	}

	return false
}

// IsPersonalEmailDomain reports whether domain is a consumer webmail provider
// (gmail.com, icloud.com, etc.) rather than a company-specific domain.
func IsPersonalEmailDomain(domain string) bool {
	d := strings.ToLower(strings.TrimSpace(domain))
	_, ok := personalEmailDomains[d]
	return ok
}

// SelectOutreachContact returns the first contact with a valid, non-automated
// email address. If none of the contacts have a routable human email, it
// returns the first contact (if any) and validEmail=false so callers can skip
// email touches while preserving non-email channel metadata.
func SelectOutreachContact(contacts []Contact) (Contact, bool) {
	if len(contacts) == 0 {
		return Contact{}, false
	}
	for _, c := range contacts {
		if !IsAutomatedOrUnroutableEmail(c.Email) {
			return c, true
		}
	}
	return contacts[0], false
}
