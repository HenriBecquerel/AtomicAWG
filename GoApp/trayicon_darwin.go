//go:build darwin

package main

import "fyne.io/fyne/v2"

// trayIconResource returns the plain white glyph (tray.svg): macOS menu-bar
// icons are conventionally monochrome so they read correctly against both
// light and dark menu bars.
func trayIconResource(fyne.Resource) fyne.Resource {
	return fyne.NewStaticResource("tray.svg", trayIconBytes)
}
