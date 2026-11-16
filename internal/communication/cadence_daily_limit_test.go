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
