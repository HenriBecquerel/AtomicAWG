package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type navSpec struct {
	id    pageID
	kind  navIconKind
	label string
}

var navSpecs = []navSpec{
	{pageStatus, iconStatus, "Статус"},
	{pageProfiles, iconProfiles, "Профили"},
	{pageConnection, iconConnection, "Подключение"},
	{pageLog, iconLog, "Журнал"},
	{pageAbout, iconAbout, "О приложении"},
}

// buildSidebar assembles the fixed-width navigation rail: app icon + name,
// the page list, and a pinned quick-toggle at the bottom — the same
// structure as Mail/Settings-style macOS apps.
func (g *guiApp) buildSidebar() fyne.CanvasObject {
	p := currentPalette()

	tile := newIconTile(fyne.NewStaticResource("brandmark.svg", brandmarkBytes), 40)
	appName := canvas.NewText(appDisplayName, p.text)
	appName.TextStyle = fyne.TextStyle{Bold: true}
	appName.TextSize = 14
	appVersionText := canvas.NewText("v"+appVersion, p.textMuted)
	appVersionText.TextSize = 11
	titleBlock := container.NewVBox(appName, appVersionText)
	header := container.NewPadded(container.NewHBox(tile, titleBlock))

	navList := container.NewVBox()
	for _, spec := range navSpecs {
		spec := spec
		item := newNavItem(spec.kind, spec.label, func() { g.showPage(spec.id) })
		g.navItems[spec.id] = item
		navList.Add(item)
	}

	g.sidebarStatusDot = newStatusCapsule()
	g.sidebarToggle = newToggleSwitch(func(checked bool) {
		if checked != g.runtime.IsRunning() {
			g.onToggleProxy()
		}
	})
	bottomRow := container.NewBorder(nil, nil, g.sidebarStatusDot, g.sidebarToggle)

	content := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator(), container.NewPadded(navList)),
		container.NewPadded(bottomRow),
		nil, nil,
	)

	bg := canvas.NewRectangle(p.sidebarBg)
	return container.New(&fixedWidthLayout{width: 220}, container.NewStack(bg, content))
}

func (g *guiApp) showPage(id pageID) {
	for pid, item := range g.navItems {
		item.SetSelected(pid == id)
	}
	for pid, page := range g.pages {
		if pid == id {
			page.Show()
		} else {
			page.Hide()
		}
	}
}
