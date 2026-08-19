//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

char *awgproxy_choose_config_file(void);
void awgproxy_free_string(char *s);
*/
import "C"

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// pickConfigFile shows the real macOS "Open" panel (NSOpenPanel) rather than
// Fyne's own cross-platform file browser widget, so the picker looks and
// behaves exactly like every other native app on the system.
func pickConfigFile(_ fyne.Window, callback func(name string, data []byte, err error)) {
	cPath := C.awgproxy_choose_config_file()
	if cPath == nil {
		callback("", nil, nil) // Cancelled.
		return
	}
	path := C.GoString(cPath)
	C.awgproxy_free_string(cPath)

	data, err := os.ReadFile(path)
	callback(filepath.Base(path), data, err)
}
