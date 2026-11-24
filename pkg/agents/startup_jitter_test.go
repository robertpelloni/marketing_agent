package agents

import (
	"testing"
	"time"
)

func TestFirstRunDelay(t *testing.T) {
	t.Setenv("STARTUP_JITTER_SECONDS", "0")
	if d := firstRunDelay(); d != 0 {
		t.Errorf("jitter 0 must disable the delay, got %v", d)
	}

	t.Setenv("STARTUP_JITTER_SECONDS", "45")
	if d := firstRunDelay(); d != 45*time.Second {
		t.Errorf("explicit jitter must be honored, got %v", d)
	}

	t.Setenv("STARTUP_JITTER_SECONDS", "")
	for i := 0; i < 20; i++ {
		d := firstRunDelay()
		if d < 30*time.Second || d > 90*time.Second {
			t.Fatalf("default jitter must be within [30s, 90s], got %v", d)
		}
	}

	t.Setenv("STARTUP_JITTER_SECONDS", "garbage")
	if d := firstRunDelay(); d < 30*time.Second || d > 90*time.Second {
		t.Errorf("garbage jitter must fall back to default range, got %v", d)
	}
}
