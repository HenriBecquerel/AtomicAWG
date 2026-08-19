package main

import (
	"image"
	"image/color"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func vspace(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(nil)
	r.SetMinSize(fyne.NewSize(1, h))
	return r
}

// sectionLabel renders a small-caps, muted section title the way macOS
// System Settings labels a grouped list above the group itself.
func sectionLabel(text string) fyne.CanvasObject {
	t := canvas.NewText(strings.ToUpper(text), currentPalette().textMuted)
	t.TextSize = 11
	t.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewPadded(t)
}

func hairline() fyne.CanvasObject {
	r := canvas.NewRectangle(currentPalette().separator)
	r.SetMinSize(fyne.NewSize(1, 1))
	return r
}

func monoEntry(minLines int) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapOff
	e.TextStyle = fyne.TextStyle{Monospace: true}
	e.Disable()
	e.SetMinRowsVisible(minLines)
	return e
}

// fixedWidthLayout constrains its single child to an exact width while
// letting its natural (minimum) height through.
type fixedWidthLayout struct{ width float32 }

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(f.width, 0)
	}
	return fyne.NewSize(f.width, objects[0].MinSize().Height)
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(f.width, size.Height))
	}
}

func (g *guiApp) buildUI() fyne.CanvasObject {
	sidebar := g.buildSidebar()

	g.pages[pageStatus] = g.buildStatusPage()
	g.pages[pageProfiles] = g.buildProfilesPage()
	g.pages[pageConnection] = g.buildConnectionPage()
	g.pages[pageLog] = g.buildLogPage()
	g.pages[pageAbout] = g.buildAboutPage()
	for _, page := range g.pages {
		g.pageStack.Add(page)
	}

	g.loadSettingsIntoInterface()
	g.interfaceLoaded = true
	g.showPage(pageStatus)

	return container.NewBorder(nil, nil, sidebar, nil, g.pageStack)
}

// --- Status page -----------------------------------------------------------

func (g *guiApp) buildStatusPage() fyne.CanvasObject {
	p := currentPalette()

	heroBg := canvas.NewRaster(func(w, h int) image.Image {
		pp := currentPalette()
		return roundedGradient(w, h, 16, pp.heroFrom, pp.heroTo)
	})
	heroTitle := canvas.NewText(appDisplayName, color.White)
	heroTitle.TextStyle = fyne.TextStyle{Bold: true}
	heroTitle.TextSize = 24
	heroSubtitle := canvas.NewText("WireGuard и AmneziaWG через SOCKS5-прокси", color.NRGBA{R: 255, G: 255, B: 255, A: 220})
	heroSubtitle.TextSize = 12
	g.heroProfile = canvas.NewText("Профиль не выбран", color.NRGBA{R: 255, G: 255, B: 255, A: 235})
	g.heroProfile.TextStyle = fyne.TextStyle{Bold: true, Monospace: false}
	g.heroProfile.TextSize = 13
	heroText := container.NewVBox(heroTitle, heroSubtitle, vspace(10), g.heroProfile)
	hero := container.NewStack(heroBg, container.NewPadded(container.NewPadded(heroText)))
	heroSized := container.New(&fixedHeightLayout{height: 128}, hero)

	g.statusCapsule = newStatusCapsule()
	g.activityLabel = widget.NewLabel("Нет данных")
	g.activityLabel.Alignment = fyne.TextAlignTrailing
	g.activityLabel.Importance = widget.LowImportance
	statusRow := container.NewBorder(nil, nil, g.statusCapsule, g.activityLabel, nil)

	g.statusDetail = widget.NewLabel("Импортируйте конфигурацию и запустите прокси.")
	g.statusDetail.Importance = widget.LowImportance
	g.statusDetail.Wrapping = fyne.TextWrapWord

	g.receivedLabel = widget.NewLabel("↓ Получено: 0 Б")
	g.receivedLabel.TextStyle = fyne.TextStyle{Bold: true}
	g.sentLabel = widget.NewLabel("↑ Отправлено: 0 Б")
	g.sentLabel.TextStyle = fyne.TextStyle{Bold: true}
	g.receivedSparkline = newSparkline(p.accent)
	g.sentSparkline = newSparkline(p.sentAccent)

	receivedBlock := container.NewVBox(g.receivedLabel, g.receivedSparkline)
	sentBlock := container.NewVBox(g.sentLabel, g.sentSparkline)
	trafficRow := container.NewGridWithColumns(2, receivedBlock, sentBlock)

	g.speedLabel = widget.NewLabel("Скорость: ↓ 0 Б/с  ↑ 0 Б/с")
	g.handshakeLabel = widget.NewLabel("Handshake: —")
	g.handshakeLabel.Importance = widget.LowImportance
	g.peakLabel = widget.NewLabel("Пик: —")
	g.peakLabel.Importance = widget.LowImportance
	g.avgLabel = widget.NewLabel("Средняя: —")
	g.avgLabel.Importance = widget.LowImportance
	metaRow := container.NewHBox(g.speedLabel, layout.NewSpacer(), g.peakLabel, vspace(1), g.avgLabel, vspace(1), g.handshakeLabel)

	g.toggleButton = newBigButton("Запустить прокси", g.onToggleProxy)
	buttonRow := container.NewBorder(nil, nil, nil, container.New(&fixedWidthLayout{width: 190}, g.toggleButton))

	statusCard := newCard(container.NewVBox(
		statusRow,
		g.statusDetail,
		vspace(6),
		trafficRow,
		vspace(4),
		metaRow,
		vspace(8),
		buttonRow,
	))

	content := container.NewVBox(
		heroSized,
		vspace(4),
		sectionLabel("Состояние"),
		statusCard,
	)
	return container.NewVScroll(container.NewPadded(content))
}

// --- Profiles page -----------------------------------------------------

func (g *guiApp) buildProfilesPage() fyne.CanvasObject {
	title := canvas.NewText("Профили", currentPalette().text)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 20

	g.profileSelect = widget.NewSelect(nil, g.onProfileSelected)
	g.profileSelect.PlaceHolder = "Нет сохранённых конфигураций"

	g.importButton = widget.NewButtonWithIcon("Добавить .conf", theme.ContentAddIcon(), g.onImportConfiguration)
	g.importButton.Importance = widget.HighImportance
	g.deleteButton = widget.NewButtonWithIcon("", theme.DeleteIcon(), g.onDeleteSelectedProfile)
	g.deleteButton.Importance = widget.DangerImportance

	selectRow := container.NewBorder(nil, nil, nil,
		container.NewHBox(g.deleteButton, g.importButton), g.profileSelect)
	selectCard := newCard(selectRow)

	g.showSecretsChk = widget.NewCheck("Показать PrivateKey и PresharedKey", g.onShowSecretsChanged)
	g.protocolLabel = widget.NewLabel("Протокол: не определён")
	g.protocolLabel.TextStyle = fyne.TextStyle{Bold: true}
	protocolRow := container.NewBorder(nil, nil, nil, g.protocolLabel, g.showSecretsChk)
	g.configViewer = monoEntry(8)
	previewCard := newCard(container.NewVBox(protocolRow, g.configViewer))

	content := container.NewVBox(
		title,
		vspace(10),
		sectionLabel("Активный профиль"),
		selectCard,
		vspace(4),
		sectionLabel("Просмотр конфигурации"),
		previewCard,
	)
	return container.NewPadded(content)
}

// --- Connection page -----------------------------------------------------

func (g *guiApp) buildConnectionPage() fyne.CanvasObject {
	title := canvas.NewText("Подключение", currentPalette().text)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 20

	g.portEntry = widget.NewEntry()
	g.portEntry.OnChanged = g.onPortChanged
	portField := container.New(&fixedWidthLayout{width: 110}, g.portEntry)
	portRow := container.NewBorder(nil, nil, widget.NewLabel("Порт прокси"), nil, portField)

	g.lanCheck = widget.NewCheck("Разрешить подключения из локальной сети", g.onLanChanged)

	g.lanWarning = widget.NewLabel("В режиме LAN открывайте порт только в доверенных сетях.")
	g.lanWarning.Importance = widget.WarningImportance
	g.lanWarning.Wrapping = fyne.TextWrapWord
	g.lanWarning.Hide()

	g.addressLabel = widget.NewLabel("socks5://127.0.0.1:1080")
	g.addressLabel.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	copyButton := widget.NewButtonWithIcon("Копировать адрес", theme.ContentCopyIcon(), g.onCopyAddress)
	addressRow := container.NewBorder(nil, nil, nil, copyButton, g.addressLabel)

	connectionCard := newCard(container.NewVBox(
		portRow,
		g.lanCheck,
		g.lanWarning,
		hairline(),
		addressRow,
	))

	content := container.NewVBox(title, vspace(10), connectionCard)
	return container.NewPadded(content)
}

// --- Log page ------------------------------------------------------------

func (g *guiApp) buildLogPage() fyne.CanvasObject {
	title := canvas.NewText("Журнал", currentPalette().text)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 20

	g.debugCheck = widget.NewCheck("Подробный журнал (DEBUG)", g.onDebugChanged)
	g.logBox = monoEntry(20)
	g.logBox.Wrapping = fyne.TextWrapOff
	logScroll := container.NewVScroll(g.logBox)

	logCard := newCard(container.NewBorder(g.debugCheck, nil, nil, nil, logScroll))

	return container.NewBorder(
		container.NewPadded(container.NewVBox(title, vspace(4))),
		nil, nil, nil,
		container.NewPadded(logCard),
	)
}

// --- About page ------------------------------------------------------------

func (g *guiApp) buildAboutPage() fyne.CanvasObject {
	p := currentPalette()
	tile := newIconTile(fyne.NewStaticResource("brandmark-about.svg", brandmarkBytes), 72)

	name := canvas.NewText(appDisplayName, p.text)
	name.TextStyle = fyne.TextStyle{Bold: true}
	name.TextSize = 22
	version := canvas.NewText("Версия "+appVersion, p.textMuted)
	version.TextSize = 12

	description := widget.NewLabel(
		"Клиент WireGuard и AmneziaWG с локальным SOCKS5-прокси. " +
			"Движок работает в том же процессе, что и интерфейс — без " +
			"отдельного исполняемого файла на диске.")
	description.Wrapping = fyne.TextWrapWord
	description.Alignment = fyne.TextAlignCenter

	engine := widget.NewLabel("Ядро: AmneziaWG v3 (github.com/amnezia-vpn/amneziawg-go)")
	engine.Importance = widget.LowImportance
	engine.Alignment = fyne.TextAlignCenter
	license := widget.NewLabel("Распространяется по лицензии MIT")
	license.Importance = widget.LowImportance
	license.Alignment = fyne.TextAlignCenter

	developersTitle := widget.NewLabel("Разработчик")
	developersTitle.Importance = widget.LowImportance
	developersTitle.Alignment = fyne.TextAlignCenter
	developer1 := widget.NewLabel("Antoine Henri Becquerel")
	developer1.Alignment = fyne.TextAlignCenter
	developer1.TextStyle = fyne.TextStyle{Bold: true}

	githubURL, _ := url.Parse("https://github.com/HenriBecquerel")
	githubLink := widget.NewHyperlink("github.com/HenriBecquerel", githubURL)
	githubLink.Alignment = fyne.TextAlignCenter

	divider := container.New(&fixedWidthLayout{width: 160}, hairline())

	block := container.NewVBox(
		container.NewCenter(tile),
		vspace(8),
		container.NewCenter(name),
		container.NewCenter(version),
		vspace(14),
		description,
		vspace(14),
		container.NewCenter(divider),
		vspace(14),
		developersTitle,
		developer1,
		githubLink,
		vspace(14),
		engine,
		license,
	)
	return container.NewCenter(container.NewPadded(block))
}

// fixedHeightLayout constrains its single child to an exact height while
// letting the available width through — used for the status page's hero
// banner so it reads as a deliberate masthead rather than growing with
// its text content.
type fixedHeightLayout struct{ height float32 }

func (f *fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, f.height)
	}
	return fyne.NewSize(objects[0].MinSize().Width, f.height)
}

func (f *fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(size.Width, f.height))
	}
}
