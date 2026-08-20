//go:build windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"fyne.io/fyne/v2"
)

var (
	comdlg32            = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileName = comdlg32.NewProc("GetOpenFileNameW")
)

// openFileName mirrors the Win32 OPENFILENAMEW struct (unicode).
type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	fileOffset        uint16
	fileExtension     uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

const (
	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800
	ofnExplorer      = 0x00080000
)

// utf16FilterString builds a Win32 filter string: pairs of
// "description\0pattern\0" terminated by an extra NUL, without going
// through syscall.UTF16PtrFromString (which rejects embedded NULs).
func utf16FilterString(parts ...string) *uint16 {
	var buf []uint16
	for _, p := range parts {
		u, _ := syscall.UTF16FromString(p) // already NUL-terminated
		buf = append(buf, u...)
	}
	buf = append(buf, 0) // final list terminator
	return &buf[0]
}

// pickConfigFile shows the real Windows "Open" dialog (GetOpenFileNameW)
// rather than Fyne's own cross-platform file browser widget, so it looks
// and behaves exactly like every other native app on the system.
func pickConfigFile(_ fyne.Window, callback func(name string, data []byte, err error)) {
	const maxPath = 32768
	fileBuf := make([]uint16, maxPath)

	title, _ := syscall.UTF16PtrFromString("Выберите конфигурацию WireGuard")
	filter := utf16FilterString(
		"WireGuard config (*.conf)", "*.conf",
		"All files (*.*)", "*.*",
	)

	ofn := openFileName{
		lpstrFilter: filter,
		lpstrFile:   &fileBuf[0],
		nMaxFile:    uint32(len(fileBuf)),
		lpstrTitle:  title,
		flags:       ofnFileMustExist | ofnPathMustExist | ofnExplorer,
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))

	ret, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		callback("", nil, nil) // Cancelled.
		return
	}

	path := syscall.UTF16ToString(fileBuf)
	data, err := os.ReadFile(path)
	callback(filepath.Base(path), data, err)
}
