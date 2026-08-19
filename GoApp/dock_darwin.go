//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa

void awgproxy_hide_dock_icon(void);
*/
import "C"

import "time"

// hideDockIcon repeatedly asks Cocoa to switch this process to an "accessory"
// application (no Dock icon, no Cmd+Tab entry — menu-bar only, like
// WireGuard's own macOS client). This is necessary because Fyne's underlying
// GLFW backend unconditionally forces the app to
// NSApplicationActivationPolicyRegular from its own applicationDidFinishLaunching
// delegate method, which runs on Cocoa's launch sequence and overrides the
// LSUIElement=true setting already present in Info.plist. Re-asserting the
// accessory policy for a brief window after launch reliably wins that race;
// any Dock icon flash is a single frame at most.
func hideDockIcon() {
	go func() {
		deadline := time.Now().Add(1500 * time.Millisecond)
		for time.Now().Before(deadline) {
			C.awgproxy_hide_dock_icon()
			time.Sleep(20 * time.Millisecond)
		}
	}()
}
