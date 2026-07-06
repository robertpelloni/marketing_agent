package main

import (
	"context"
	"fmt"
	"os"

	"github.com/getlantern/systray"
)

// setupSystray initializes and runs the system tray icon
func setupSystray(cancel context.CancelFunc, port string) {
	systray.Run(func() {
		onReady(cancel, port)
	}, onExit)
}

func onReady(cancel context.CancelFunc, port string) {
	systray.SetTitle("Marketing Agent")
	systray.SetTooltip("TormentNexus / HyperNexus Marketing Agent")

	mOpen := systray.AddMenuItem("Open Dashboard", "Open the web dashboard")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				// Open browser
				url := fmt.Sprintf("http://localhost:%s", port)
				// Using some generic cross-platform way might be needed, or just printing for now.
				fmt.Println("Please open: ", url)
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// Let the context cancellation happen in the main thread
	// Or we can just call os.Exit(0)
	os.Exit(0)
}
