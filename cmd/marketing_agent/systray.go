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

	mOpen := systray.AddMenuItem("Open Dashboard", "Open the unified web dashboard")
	mOutreach := systray.AddMenuItem("View Cold Outreach Tracking", "Open cold outreach delivery monitor")
	systray.AddSeparator()
	mHyperSite := systray.AddMenuItem("Open HyperNexus Site", "Visit corporate hypernexus.site")
	mTormentSite := systray.AddMenuItem("Open TormentNexus Site", "Visit developer tormentnexus.site")
	mBlog := systray.AddMenuItem("Open Technical Blog", "View generated blog posts")
	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart Service", "Restart the marketing agent server")
	mQuit := systray.AddMenuItem("Quit Servers & Exit", "Gracefully terminate servers and exit")

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser(fmt.Sprintf("http://localhost:%s", port))
			case <-mOutreach.ClickedCh:
				openBrowser(fmt.Sprintf("http://localhost:%s/#outreach", port))
			case <-mHyperSite.ClickedCh:
				openBrowser("https://hypernexus.site")
			case <-mTormentSite.ClickedCh:
				openBrowser("https://tormentnexus.site")
			case <-mBlog.ClickedCh:
				// There is no local /blog route; the blog is published on the site.
				openBrowser("https://hypernexus.site/blog/")
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
