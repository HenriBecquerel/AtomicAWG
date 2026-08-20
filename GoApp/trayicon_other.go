//go:build !darwin

package main

import "fyne.io/fyne/v2"

// trayIconResource returns the full-color gradient app icon (same as the
// window title bar and taskbar): unlike macOS, Windows and Linux tray areas
// conventionally show colorful icons rather than monochrome glyphs, and a
// plain white line-art glyph reads as a broken/black icon there.
func trayIconResource(appIcon fyne.Resource) fyne.Resource {
	return appIcon
}
