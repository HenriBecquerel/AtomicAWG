package main

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// fyneFallbackPicker uses Fyne's own portable file dialog. It's a fallback
// for platforms without a native picker wired up, or where the native tool
// (e.g. zenity/kdialog on Linux) isn't installed.
func fyneFallbackPicker(window fyne.Window, callback func(name string, data []byte, err error)) {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			callback("", nil, err)
			return
		}
		if reader == nil {
			callback("", nil, nil) // Cancelled.
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		callback(reader.URI().Name(), data, err)
	}, window)
	fd.Show()
}
