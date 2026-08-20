//go:build linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
)

// pickConfigFile shows the desktop's native "Open" dialog via zenity (GTK,
// GNOME and most distros) or kdialog (KDE) — whichever is installed — so it
// looks and behaves like every other native app on the system, instead of
// Fyne's own cross-platform file browser widget. Falls back to that Fyne
// dialog only if neither tool is present.
func pickConfigFile(window fyne.Window, callback func(name string, data []byte, err error)) {
	path, handled := runNativeLinuxPicker()
	if !handled {
		fyneFallbackPicker(window, callback)
		return
	}
	if path == "" {
		callback("", nil, nil) // Cancelled (or the tool failed) — no error to show.
		return
	}

	data, err := os.ReadFile(path)
	callback(filepath.Base(path), data, err)
}

// runNativeLinuxPicker runs whichever native file-chooser tool is available.
// handled reports whether a tool was found and run at all (as opposed to
// falling through to the Fyne dialog); path is empty on cancel or failure.
func runNativeLinuxPicker() (path string, handled bool) {
	if bin, err := exec.LookPath("zenity"); err == nil {
		out, err := exec.Command(bin,
			"--file-selection",
			"--title=Выберите конфигурацию WireGuard",
			"--file-filter=WireGuard config (*.conf) | *.conf",
			"--file-filter=All files | *",
		).Output()
		if err != nil {
			return "", true // Non-zero exit: user cancelled (or a real error) — treat as cancel.
		}
		return strings.TrimSpace(string(out)), true
	}

	if bin, err := exec.LookPath("kdialog"); err == nil {
		out, err := exec.Command(bin,
			"--title", "Выберите конфигурацию WireGuard",
			"--getopenfilename", os.Getenv("HOME"),
			"*.conf|WireGuard config\n*|All files",
		).Output()
		if err != nil {
			return "", true
		}
		return strings.TrimSpace(string(out)), true
	}

	return "", false
}
