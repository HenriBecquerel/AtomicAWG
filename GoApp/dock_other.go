//go:build !darwin

package main

// hideDockIcon is a no-op outside macOS; taskbar presence there is governed
// by the platform's own conventions instead.
func hideDockIcon() {}
