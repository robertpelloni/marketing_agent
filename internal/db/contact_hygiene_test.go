package db

import "testing"

func TestIsAutomatedOrUnroutableEmail(t *testing.T) {
	blocked := []string{
		"",
		"not-an-email",
		"user@localhost",
		"dev@machine.local",
		"test@example.com",
		"12345+user@users.noreply.github.com",
		"noreply@github.com",
		"actions@github.com",
		"action@github.com",
		"lopesjrodrigo@gmail.com.unlinked202609291341",
		"user@127.0.0.1",
		"codex@openai.com",
		"claude@anthropic.com",
		"cursoragent@cursor.com",
		"bot@renovateapp.com",
		"bot@mcp-scanner.github.io",
		"maintainer@project.github.io",
		"dependabot[bot]@users.noreply.github.com",
		"snyk-bot@snyk.io",
		"bot-release@acme.io",
		"support@acme.io",
		"admin@acme.io",
		"info@acme.io",
		"security@acme.io",
		"do-not-reply@acme.io",
	}
	for _, email := range blocked {
		if !IsAutomatedOrUnroutableEmail(email) {
			t.Errorf("expected %q to be classified as automated/unroutable", email)
		}
	}

	allowed := []string{
		"jane@acme.dev",
		"founder@botpress.com",
		"talbot@gmail.com",
		"abbott@startup.io",
		"cabot@company.ai",
		"ryu.yamamoto@icloud.com",
		"cto@hypernexus.site",
	}
	for _, email := range allowed {
		if IsAutomatedOrUnroutableEmail(email) {
			t.Errorf("expected human email %q to be allowed", email)
		}
	}
}

func TestIsAutomatedCommitAuthor(t *testing.T) {
	if !IsAutomatedCommitAuthor("renovate[bot]", "renovate@acme.dev") {
		t.Fatal("expected [bot] author name to be blocked")
	}
	if !IsAutomatedCommitAuthor("Claude Code", "dev@acme.dev") {
		t.Fatal("expected Claude Code author name to be blocked")
	}
	if !IsAutomatedCommitAuthor("ellmos-multi-bot", "notify@ellmos.dev") {
		t.Fatal("expected -bot author name to be blocked")
	}
	if !IsAutomatedCommitAuthor("MCP Scanner Bot", "scanner@acme.dev") {
		t.Fatal("expected ' Bot' author name to be blocked")
	}
	if IsAutomatedCommitAuthor("Jane Talbot", "talbot@acme.dev") {
		t.Fatal("expected human author Jane Talbot to be allowed")
	}
}

func TestSelectOutreachContact_SkipsLeadingBotContact(t *testing.T) {
	contacts := []Contact{
		{ID: 1, Name: "Codex", Email: "codex@openai.com"},
		{ID: 2, Name: "Bot", Email: "bot@mcp-scanner.github.io"},
		{ID: 3, Name: "Real Maintainer", Email: "maintainer@acme.dev"},
	}
	got, ok := SelectOutreachContact(contacts)
	if !ok || got.ID != 3 || got.Email != "maintainer@acme.dev" {
		t.Fatalf("expected human contact ID 3, got ok=%v contact=%+v", ok, got)
	}

	onlyBots := []Contact{
		{ID: 10, Name: "Renovate", Email: "bot@renovateapp.com"},
	}
	fallback, ok := SelectOutreachContact(onlyBots)
	if ok || fallback.ID != 10 {
		t.Fatalf("expected ok=false with fallback ID 10 when only bot contacts exist, got ok=%v contact=%+v", ok, fallback)
	}
}
