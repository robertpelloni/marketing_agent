package communication

import (
	"strings"
	"testing"
)

func TestParseRawMIMEBody_MultipartQuotedPrintableAndDSN(t *testing.T) {
	raw := "From: Dheeraj Kiran <dheerajkiran.e@gmail.com>\r\n" +
		"Subject: Re: TormentNexus for autonomous-urban-driving-stack -- Quick Question\r\n" +
		"Message-ID: <CAEhE9oEG+96V003@mail.gmail.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"000000000000abc123\"\r\n\r\n" +
		"--000000000000abc123\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: quoted-printable\r\n\r\n" +
		"Hi HyperNexus Team,\r\n\r\n" +
		"I=E2=80=99m not currently exploring new platforms or services.\r\n\r\n" +
		"On Sat, Nov 14, 2026 at 7:00 PM HyperNexus Official <\r\n" +
		"hypernexusofficialllc@gmail.com> wrote:\r\n" +
		"> Pricing and architecture overview\r\n" +
		"--000000000000abc123\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		"<div>HTML copy</div>\r\n" +
		"--000000000000abc123--\r\n"

	body, msgID := ParseRawMIMEBody([]byte(raw))
	if msgID != "<CAEhE9oEG+96V003@mail.gmail.com>" {
		t.Fatalf("unexpected Message-ID: %q", msgID)
	}
	if !strings.Contains(body, "not currently exploring new platforms or services") {
		t.Fatalf("expected decoded quoted-printable plain text body, got: %q", body)
	}

	fresh := ExtractFreshInboundText("Subject: Re: Quick Question\n\n" + body)
	if strings.Contains(fresh, "Pricing and architecture") {
		t.Fatalf("expected quoted outreach block to be stripped, got: %q", fresh)
	}
	if !isPoliteDecline(fresh) {
		t.Fatalf("expected polite decline to be detected from fresh reply: %q", fresh)
	}
}

func TestExtractPermanentBounceRecipients_FailureVsDelay(t *testing.T) {
	permBody := "** Address not found **\n" +
		"Your message wasn't delivered to nkirchner@zeenea.com because the address couldn't be found.\n" +
		"Final-Recipient: rfc822; nkirchner@zeenea.com\n" +
		"Action: failed\n" +
		"Status: 5.1.1\n" +
		"Reporting-MTA: dns; googlemail.com (sent by hypernexusofficialllc@gmail.com)"

	addrs, isDSN, isDelay := ExtractPermanentBounceRecipients(
		"mailer-daemon@googlemail.com",
		"Delivery Status Notification (Failure)",
		permBody,
		"hypernexusofficialllc@gmail.com",
	)
	if !isDSN || isDelay {
		t.Fatalf("expected permanent DSN (isDSN=true, isDelay=false), got isDSN=%v isDelay=%v", isDSN, isDelay)
	}
	if len(addrs) != 1 || addrs[0] != "nkirchner@zeenea.com" {
		t.Fatalf("expected [nkirchner@zeenea.com], got %v", addrs)
	}

	delayBody := "** Delivery incomplete **\n" +
		"There was a temporary problem delivering your message to arc@cpan.org. Gmail will retry for 46 more hours."
	delayAddrs, isDSN, isDelay := ExtractPermanentBounceRecipients(
		"mailer-daemon@googlemail.com",
		"Delivery Status Notification (Delay)",
		delayBody,
		"hypernexusofficialllc@gmail.com",
	)
	if !isDSN || !isDelay || len(delayAddrs) != 0 {
		t.Fatalf("expected temporary delay DSN to be ignored safely, got isDSN=%v isDelay=%v addrs=%v", isDSN, isDelay, delayAddrs)
	}

	_, humanDSN, _ := ExtractPermanentBounceRecipients(
		"lancey3@gmail.com",
		"Re: TormentNexus for rustbgpd -- Quick Question",
		"How can I help?\nLance",
		"hypernexusofficialllc@gmail.com",
	)
	if humanDSN {
		t.Fatal("expected human reply not to be flagged as DSN")
	}
}

func TestSanitizeOutboundText_RemovesNamePlaceholders(t *testing.T) {
	dirty := "Hi Lance,\n\nCheck out hypernexus.site {{unused_var}}.\n\nBest,\n[Your Name]"
	got := SanitizeOutboundText(dirty, "HyperNexus Team")
	if strings.Contains(strings.ToLower(got), "[your name]") || strings.Contains(got, "{{") {
		t.Fatalf("expected placeholders removed, got: %q", got)
	}
	if !strings.Contains(got, "Best,\nHyperNexus Team") {
		t.Fatalf("expected fallback signoff injected, got: %q", got)
	}
}

func TestPoliteDecline_DoesNotFalsePositiveOnEngagedReply(t *testing.T) {
	if isPoliteDecline("How can I help?\nLance") {
		t.Fatal("expected engaged reply not to trigger polite decline")
	}
	if !isPoliteDecline("Hi, it's just a personal project not anywhere in the production, thanks!") {
		t.Fatal("expected personal-project reply to trigger polite decline")
	}
	if !isPoliteDecline("Thats a project for fun nothing more. No thank you") {
		t.Fatal("expected 'no thank you' hobby reply to trigger polite decline")
	}
}
