//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/getlantern/systray"
)

// setupSystray initializes and runs the system tray icon on Windows.
func setupSystray(cancel context.CancelFunc, port string) {
	systray.Run(func() {
		onReady(cancel, port)
	}, onExit)
}

func onReady(cancel context.CancelFunc, port string) {
	systray.SetTitle("Marketing Agent")
	systray.SetTooltip("TormentNexus / HyperNexus Marketing Agent")

	mOpen := systray.AddMenuItem("Open Dashboard", "Open the web dashboard")
	systray.AddSeparator()
	mBlog := systray.AddMenuItem("Open Blog", "View generated blog posts")
	mSite := systray.AddMenuItem("Open TormentNexus Site", "Visit tormentnexus.site")
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart Service", "Restart the marketing agent")
	mQuit := systray.AddMenuItem("Quit", "Quit the servers and exit")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser(fmt.Sprintf("http://localhost:%s", port))
			case <-mBlog.ClickedCh:
				openBrowser(fmt.Sprintf("http://localhost:%s/blog", port))
			case <-mSite.ClickedCh:
				openBrowser("https://tormentnexus.site")
			case <-mRestart.ClickedCh:
				// Cancel current context, quit tray, and re-launch self
				cancel()
				systray.Quit()
				exe, _ := os.Executable()
				if exe != "" {
					cmd := exec.Command(exe, os.Args[1:]...)
					cmd.Start()
				}
				return
			case <-mQuit.ClickedCh:
				cancel()
				systray.Quit()
				return
			}
		}
	}()
}

func openBrowser(url string) {
	// Use cmd /c start on Windows to open the default browser
	_ = exec.Command("cmd", "/c", "start", url).Start()
}

func onExit() {
	os.Exit(0)
}
