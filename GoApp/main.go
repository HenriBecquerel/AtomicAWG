// Command atomicawg is a single-binary, native desktop client for
// WireGuard and AmneziaWG. The proxy engine (package github.com/awgproxy/awgproxy)
// is linked in directly rather than shelled out to as a bundled second
// executable, so there is nothing to extract to disk and launch at runtime.
package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed assets/appicon.png
var appIconBytes []byte

//go:embed assets/tray.svg
var trayIconBytes []byte

//go:embed assets/brandmark.svg
var brandmarkBytes []byte

const appDisplayName = "AtomicAWG"

// appVersion is set via -ldflags "-X main.appVersion=..." at build time.
var appVersion = "dev"

func main() {
	fyneApp := app.NewWithID("com.atomicawg.app")
	fyneApp.Settings().SetTheme(&luxTheme{})

	appIcon := fyne.NewStaticResource("appicon.png", appIconBytes)
	fyneApp.SetIcon(appIcon)

	window := fyneApp.NewWindow(appDisplayName + " " + appVersion)
	window.SetIcon(appIcon)
	window.Resize(fyne.NewSize(860, 760))

	g := newGUIApp(fyneApp, window)
	window.SetContent(g.buildUI())
	g.setupTray(appIcon)

	window.CenterOnScreen()
	window.Show()
	hideDockIcon()
	fyneApp.Run()
}
