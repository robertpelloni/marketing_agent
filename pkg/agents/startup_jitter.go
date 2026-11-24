package agents

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// firstRunDelay spreads first-run work at process startup. Every worker
// used to fire immediately on boot, so each restart produced a
// thundering herd (LLM generations + social posts + cadence all at
// once) that starved cycle-time network calls on the shared 2-vCPU
// VPS. STARTUP_JITTER_SECONDS overrides (0 disables, e.g. in tests).
func firstRunDelay() time.Duration {
	if v := os.Getenv("STARTUP_JITTER_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Second
		}
	}
	return time.Duration(30+rand.Intn(61)) * time.Second // 30-90s
}

// waitFirstRun sleeps the startup jitter unless ctx ends first.
// Returns false when the caller should stop instead of running.
func waitFirstRun(ctx context.Context, worker string) bool {
	d := firstRunDelay()
	if d <= 0 {
		return true
	}
	slog.Info(fmt.Sprintf("%s: first run in %v (startup jitter)", worker, d))
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
