//go:build !darwin && !windows && !linux

package main

import "fyne.io/fyne/v2"

// pickConfigFile falls back to Fyne's portable file dialog on platforms
// without a native picker wired up.
func pickConfigFile(window fyne.Window, callback func(name string, data []byte, err error)) {
	fyneFallbackPicker(window, callback)
}
