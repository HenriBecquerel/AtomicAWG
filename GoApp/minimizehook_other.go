//go:build !windows

package main

import "fyne.io/fyne/v2"

// installMinimizeToTray is a no-op outside Windows; macOS already hides via
// LSUIElement/hideDockIcon, and Linux desktop environments vary too much to
// target generically here.
func installMinimizeToTray(fyne.Window) {}
