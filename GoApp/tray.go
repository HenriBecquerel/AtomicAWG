package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// setupTray configures the menu-bar presence: the app lives in the tray,
// clicking the icon shows/raises the window (the same interaction WireGuard's
// own macOS client uses), and closing the window hides it rather than
// quitting — the proxy keeps running until "Выход" is chosen explicitly.
func (g *guiApp) setupTray(appIcon fyne.Resource) {
	trayIcon := fyne.NewStaticResource("tray.svg", trayIconBytes)

	showItem := fyne.NewMenuItem("Открыть AtomicAWG", func() {
		g.window.Show()
		g.window.RequestFocus()
	})
	g.trayToggleItem = fyne.NewMenuItem("Запустить прокси", g.onToggleProxy)
	quitItem := fyne.NewMenuItem("Выход", g.quit)
	quitItem.IsQuit = true

	g.trayMenu = fyne.NewMenu("AtomicAWG", showItem, g.trayToggleItem, fyne.NewMenuItemSeparator(), quitItem)

	if desk, ok := g.fyneApp.(desktop.App); ok {
		desk.SetSystemTrayIcon(trayIcon)
		desk.SetSystemTrayMenu(g.trayMenu)
		desk.SetSystemTrayWindow(g.window)
	}

	g.window.SetCloseIntercept(func() {
		g.window.Hide()
	})
}

func (g *guiApp) updateTray(running bool) {
	if g.trayToggleItem == nil {
		return
	}
	if running {
		g.trayToggleItem.Label = "Остановить прокси"
	} else {
		g.trayToggleItem.Label = "Запустить прокси"
	}
	if g.trayMenu != nil {
		g.trayMenu.Refresh()
	}
}

// quit stops the tunnel cleanly before terminating the process.
func (g *guiApp) quit() {
	g.stopStatistics()
	g.runtime.Stop()
	g.fyneApp.Quit()
}
