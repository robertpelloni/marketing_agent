package communication

import (
	"errors"
	"testing"
)

func TestIsDailyLimitError(t *testing.T) {
	gmailQuota := errors.New(`DATA failed: 550 "5.4.5 Daily user sending limit exceeded. For more information on Gmail sending limits go to https://support.google.com/a/answer/166852 - gsmtp"`)
	if !isDailyLimitError(gmailQuota) {
		t.Error("expected Gmail daily-limit error to be detected")
	}
	if !isDailyLimitError(errors.New("550 5.4.5 something")) {
		t.Error("expected bare 5.4.5 to be detected")
	}
	if isDailyLimitError(errors.New("connection refused")) {
		t.Error("unrelated error must not be treated as daily limit")
	}
	if isDailyLimitError(nil) {
		t.Error("nil must not be treated as daily limit")
	}
}

func TestLinkedInLaneDisabled(t *testing.T) {
	t.Setenv("LINKEDIN_DISABLED", "true")
	if !linkedInLaneDisabled() {
		t.Error("expected LinkedIn lane disabled when LINKEDIN_DISABLED=true")
	}
	t.Setenv("LINKEDIN_DISABLED", "")
	if linkedInLaneDisabled() {
		t.Error("expected LinkedIn lane enabled when LINKEDIN_DISABLED unset")
	}
	t.Setenv("LINKEDIN_DISABLED", "false")
	if linkedInLaneDisabled() {
		t.Error("expected LinkedIn lane enabled when LINKEDIN_DISABLED=false")
	}
}

func TestIsProviderAuthError(t *testing.T) {
	disabled := errors.New(`failed to get token: oauth2: "disabled_client" "The OAuth client was disabled."`)
	if !isProviderAuthError(disabled) {
		t.Error("expected disabled_client token error to be detected")
	}
	if !isProviderAuthError(errors.New("oauth2: invalid_grant: token revoked")) {
		t.Error("expected invalid_grant to be detected")
	}
	if isProviderAuthError(errors.New("550 5.4.5 Daily user sending limit exceeded")) {
		t.Error("daily-limit quota error is not an auth error (handled separately)")
	}
	if isProviderAuthError(errors.New("connection refused")) || isProviderAuthError(nil) {
		t.Error("unrelated/nil errors must not be treated as auth errors")
	}
}
