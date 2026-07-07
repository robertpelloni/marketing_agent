//go:build !windows

package main

import "context"

// setupSystray is a no-op on headless Linux servers
func setupSystray(cancel context.CancelFunc, port string) {
	// Systray is only supported on Windows.
	// On headless Linux, we rely on the web dashboard and graceful shutdown via SIGTERM.
	_ = cancel
	_ = port
}
