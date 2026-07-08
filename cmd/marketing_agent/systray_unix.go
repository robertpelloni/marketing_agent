//go:build !windows

package main

import (
	"context"
	"log/slog"
)

// setupSystray blocks on Linux by waiting forever.
// The graceful shutdown goroutine in main() triggers os.Exit(0).
func setupSystray(cancel context.CancelFunc, port string) {
	slog.Info("Linux headless mode: running (signal handler manages shutdown)...")
	select {}
}
