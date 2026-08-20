//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
)

var (
	user32                = syscall.NewLazyDLL("user32.dll")
	procFindWindowW       = user32.NewProc("FindWindowW")
	procSetWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProcW   = user32.NewProc("CallWindowProcW")
	procShowWindow        = user32.NewProc("ShowWindow")
)

const (
	gwlpWndProc    = ^uintptr(3) // GWLP_WNDPROC (-4) as uintptr
	wmSysCommand   = 0x0112
	scMinimizeMask = 0xFFF0
	scMinimize     = 0xF020
	swHide         = 0
)

var origWndProc uintptr

// wndProc intercepts WM_SYSCOMMAND/SC_MINIMIZE (fired by both the window's
// minimize button and Windows+Down) and hides the window instead of the
// default minimize-to-taskbar-thumbnail behavior; everything else is
// forwarded to the original window procedure unchanged.
func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if uint32(msg) == wmSysCommand && (wParam&scMinimizeMask) == scMinimize {
		procShowWindow.Call(hwnd, swHide)
		return 0
	}
	ret, _, _ := procCallWindowProcW.Call(origWndProc, hwnd, msg, wParam, lParam)
	return ret
}

// installMinimizeToTray makes the minimize button hide the window straight
// to the tray, matching the close button's existing hide-to-tray behavior
// (SetCloseIntercept in tray.go) instead of leaving a taskbar thumbnail.
func installMinimizeToTray(window fyne.Window) {
	go func() {
		titlePtr, err := syscall.UTF16PtrFromString(window.Title())
		if err != nil {
			return
		}

		var hwnd uintptr
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			h, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
			if h != 0 {
				hwnd = h
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if hwnd == 0 {
			return
		}

		callback := syscall.NewCallback(wndProc)
		orig, _, _ := procSetWindowLongPtrW.Call(hwnd, uintptr(gwlpWndProc), callback)
		origWndProc = orig
	}()
}
