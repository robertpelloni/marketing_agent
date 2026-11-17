package communication

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGmailAuthBreakerTripsAndSuppresses(t *testing.T) {
	s := NewGmailOAuth2Sender()
	calls := 0
	s.sendFn = func(to string, msg []byte) error {
		calls++
		return errors.New(`failed to get token: oauth2: "disabled_client" "The OAuth client was disabled."`)
	}
	base := time.Now()
	s.now = func() time.Time { return base }

	ctx := context.Background()
	msg := EmailMessage{To: "a@b.com", Subject: "s", Body: "b"}

	if err := s.Send(ctx, msg); err == nil {
		t.Fatal("first send must surface the auth error")
	}
	if calls != 1 {
		t.Fatalf("expected 1 underlying attempt, got %d", calls)
	}

	// Within the window: suppressed without another underlying attempt.
	err := s.Send(ctx, msg)
	if err == nil || !strings.Contains(err.Error(), "auth breaker") {
		t.Fatalf("expected breaker error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("breaker must suppress repeat attempts, got %d calls", calls)
	}

	// After the window: a real attempt happens again.
	s.now = func() time.Time { return base.Add(gmailAuthBreakerWindow + time.Minute) }
	if err := s.Send(ctx, msg); err == nil {
		t.Fatal("send after window must attempt (and fail with the stub error)")
	}
	if calls != 2 {
		t.Fatalf("expected retry after window, got %d calls", calls)
	}
}

func TestGmailBreakerIgnoresNonAuthErrors(t *testing.T) {
	s := NewGmailOAuth2Sender()
	calls := 0
	s.sendFn = func(to string, msg []byte) error {
		calls++
		return errors.New("failed to connect: dial tcp: i/o timeout")
	}
	ctx := context.Background()
	msg := EmailMessage{To: "a@b.com", Subject: "s", Body: "b"}

	_ = s.Send(ctx, msg)
	_ = s.Send(ctx, msg)
	if calls != 2 {
		t.Fatalf("non-auth failures must not trip the breaker, got %d calls", calls)
	}
}

func TestGmailBreakerClearsOnSuccess(t *testing.T) {
	s := NewGmailOAuth2Sender()
	fail := true
	s.sendFn = func(to string, msg []byte) error {
		if fail {
			return errors.New(`failed to get token: oauth2: "disabled_client"`)
		}
		return nil
	}
	base := time.Now()
	s.now = func() time.Time { return base }
	ctx := context.Background()
	msg := EmailMessage{To: "a@b.com", Subject: "s", Body: "b"}

	_ = s.Send(ctx, msg) // trips
	fail = false
	s.now = func() time.Time { return base.Add(gmailAuthBreakerWindow + time.Minute) }
	if err := s.Send(ctx, msg); err != nil {
		t.Fatalf("send after window must succeed once provider recovers: %v", err)
	}
	// Breaker cleared: immediate next send also goes through.
	if err := s.Send(ctx, msg); err != nil {
		t.Fatalf("breaker must stay cleared after success: %v", err)
	}
}
