package main

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type pageID int

const (
	pageStatus pageID = iota
	pageProfiles
	pageConnection
	pageLog
	pageAbout
)

// guiApp holds every widget reference and piece of mutable UI state for the
// main window. It mirrors the shape of the previous MainWindow.axaml.cs but
// with no external process to manage — the tunnel is an in-process object.
type guiApp struct {
	fyneApp fyne.App
	window  fyne.Window

	settings appSettings
	runtime  *tunnelRuntime

	// Navigation.
	navItems         map[pageID]*navItem
	pages            map[pageID]fyne.CanvasObject
	pageStack        *fyne.Container
	sidebarStatusDot *statusCapsule
	sidebarToggle    *toggleSwitch

	// Status page.
	statusCapsule     *statusCapsule
	activityLabel     *widget.Label
	toggleButton      *bigButton
	receivedLabel     *widget.Label
	sentLabel         *widget.Label
	speedLabel        *widget.Label
	handshakeLabel    *widget.Label
	statusDetail      *widget.Label
	peakLabel         *widget.Label
	avgLabel          *widget.Label
	heroProfile       *canvas.Text
	receivedSparkline *sparkline
	sentSparkline     *sparkline

	// Profiles page.
	profileSelect   *widget.Select
	importButton    *widget.Button
	deleteButton    *widget.Button
	showSecretsChk  *widget.Check
	protocolLabel   *widget.Label
	configViewer    *widget.Entry
	selectedProfile string

	// Connection page.
	portEntry    *widget.Entry
	lanCheck     *widget.Check
	lanWarning   *widget.Label
	addressLabel *widget.Label

	// Log page.
	debugCheck *widget.Check
	logBox     *widget.Entry

	// Tray.
	trayMenu       *fyne.Menu
	trayToggleItem *fyne.MenuItem

	// Runtime state.
	interfaceLoaded bool
	hasHandshake    bool
	statsCancel     context.CancelFunc
	prevReceived    int64
	prevSent        int64
	prevSampleTime  time.Time
	hasPrevSample   bool

	receivedHistory []float64
	sentHistory     []float64
	peakBps         float64
	sessionStart    time.Time
}

func newGUIApp(fyneApp fyne.App, window fyne.Window) *guiApp {
	return &guiApp{
		fyneApp:   fyneApp,
		window:    window,
		settings:  loadSettings(),
		runtime:   newTunnelRuntime(),
		navItems:  make(map[pageID]*navItem),
		pages:     make(map[pageID]fyne.CanvasObject),
		pageStack: container.NewStack(),
	}
}
