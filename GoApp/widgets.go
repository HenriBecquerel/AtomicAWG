package main

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	cardRadius  float32 = 13
	cardPadding float32 = 16
)

// card — белая (или тёмная) панель со скруглёнными углами и волосяной обводкой,
// как группированные секции в настройках macOS.
type card struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func newCard(content fyne.CanvasObject) *card {
	c := &card{content: content}
	c.ExtendBaseWidget(c)
	return c
}

func (c *card) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = cardRadius
	bg.StrokeWidth = 1
	r := &cardRenderer{card: c, bg: bg}
	r.Refresh()
	return r
}

type cardRenderer struct {
	card *card
	bg   *canvas.Rectangle
}

func (r *cardRenderer) Destroy() {}

func (r *cardRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.card.content.Move(fyne.NewPos(cardPadding, cardPadding))
	r.card.content.Resize(size.SubtractWidthHeight(cardPadding*2, cardPadding*2))
}

func (r *cardRenderer) MinSize() fyne.Size {
	return r.card.content.MinSize().AddWidthHeight(cardPadding*2, cardPadding*2)
}

func (r *cardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.card.content}
}

func (r *cardRenderer) Refresh() {
	p := currentPalette()
	r.bg.FillColor = p.card
	r.bg.StrokeColor = p.cardStroke
	r.bg.Refresh()
	r.card.content.Refresh()
}

// bigButton — главная кнопка приложения: крупная, заливка системным синим
// (красным в режиме остановки), с состояниями hover и нажатия.
type bigButton struct {
	widget.BaseWidget
	label    string
	danger   bool
	disabled bool
	onTap    func()

	hovered bool
	pressed bool
}

var _ fyne.Tappable = (*bigButton)(nil)
var _ desktop.Hoverable = (*bigButton)(nil)
var _ desktop.Cursorable = (*bigButton)(nil)

func newBigButton(label string, onTap func()) *bigButton {
	b := &bigButton{label: label, onTap: onTap}
	b.ExtendBaseWidget(b)
	return b
}

func (b *bigButton) SetState(label string, danger bool) {
	b.label = label
	b.danger = danger
	b.Refresh()
}

func (b *bigButton) SetDisabled(disabled bool) {
	b.disabled = disabled
	b.Refresh()
}

func (b *bigButton) Tapped(*fyne.PointEvent) {
	if b.disabled || b.onTap == nil {
		return
	}
	b.onTap()
}

func (b *bigButton) MouseIn(*desktop.MouseEvent)    { b.hovered = true; b.Refresh() }
func (b *bigButton) MouseMoved(*desktop.MouseEvent) {}
func (b *bigButton) MouseOut()                      { b.hovered = false; b.pressed = false; b.Refresh() }

func (b *bigButton) MouseDown(*desktop.MouseEvent) { b.pressed = true; b.Refresh() }
func (b *bigButton) MouseUp(*desktop.MouseEvent)   { b.pressed = false; b.Refresh() }

func (b *bigButton) Cursor() desktop.Cursor {
	if b.disabled {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

func (b *bigButton) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	rect.CornerRadius = 12
	text := canvas.NewText(b.label, color.White)
	text.TextStyle = fyne.TextStyle{Bold: true}
	text.TextSize = 15
	text.Alignment = fyne.TextAlignCenter
	r := &bigButtonRenderer{button: b, rect: rect, text: text}
	r.Refresh()
	return r
}

type bigButtonRenderer struct {
	button *bigButton
	rect   *canvas.Rectangle
	text   *canvas.Text
}

func (r *bigButtonRenderer) Destroy() {}

func (r *bigButtonRenderer) Layout(size fyne.Size) {
	r.rect.Resize(size)
	r.text.Resize(fyne.NewSize(size.Width, r.text.MinSize().Height))
	r.text.Move(fyne.NewPos(0, (size.Height-r.text.MinSize().Height)/2))
}

func (r *bigButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.text.MinSize().Width+48, 46)
}

func (r *bigButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.rect, r.text}
}

func (r *bigButtonRenderer) Refresh() {
	p := currentPalette()
	fill := p.accent
	pressFill := p.accentPress
	if r.button.danger {
		fill = p.danger
		pressFill = r.button.dangerFill(p)
	}
	switch {
	case r.button.disabled:
		r.rect.FillColor = withAlpha(fill, 0x59)
	case r.button.pressed:
		r.rect.FillColor = pressFill
	case r.button.hovered:
		r.rect.FillColor = blendTowards(fill, pressFill)
	default:
		r.rect.FillColor = fill
	}
	r.text.Text = r.button.label
	r.text.Color = p.textOnAccent
	r.rect.Refresh()
	r.text.Refresh()
}

func (b *bigButton) dangerFill(p palette) color.Color { return p.dangerPress }

func withAlpha(c color.Color, alpha uint8) color.Color {
	red, green, blue, _ := c.RGBA()
	return color.NRGBA{R: uint8(red >> 8), G: uint8(green >> 8), B: uint8(blue >> 8), A: alpha}
}

func blendTowards(from, to color.Color) color.Color {
	fr, fg, fb, _ := from.RGBA()
	tr, tg, tb, _ := to.RGBA()
	mix := func(a, b uint32) uint8 { return uint8(((a*3 + b) / 4) >> 8) }
	return color.NRGBA{R: mix(fr, tr), G: mix(fg, tg), B: mix(fb, tb), A: 0xFF}
}

// statusCapsule — капсула состояния с точкой-индикатором, как бейджи в macOS.
type statusCapsule struct {
	widget.BaseWidget
	text string
	live bool
}

func newStatusCapsule() *statusCapsule {
	c := &statusCapsule{text: "Остановлен"}
	c.ExtendBaseWidget(c)
	return c
}

func (c *statusCapsule) SetState(text string, live bool) {
	c.text = text
	c.live = live
	c.Refresh()
}

func (c *statusCapsule) CreateRenderer() fyne.WidgetRenderer {
	pill := canvas.NewRectangle(color.Transparent)
	dot := canvas.NewCircle(color.Transparent)
	label := canvas.NewText(c.text, color.Black)
	label.TextSize = 12
	label.TextStyle = fyne.TextStyle{Bold: true}
	r := &capsuleRenderer{capsule: c, pill: pill, dot: dot, label: label}
	r.Refresh()
	return r
}

type capsuleRenderer struct {
	capsule *statusCapsule
	pill    *canvas.Rectangle
	dot     *canvas.Circle
	label   *canvas.Text
}

func (r *capsuleRenderer) Destroy() {}

func (r *capsuleRenderer) Layout(size fyne.Size) {
	r.pill.Resize(size)
	r.pill.CornerRadius = size.Height / 2
	const dotSize float32 = 8
	r.dot.Resize(fyne.NewSize(dotSize, dotSize))
	r.dot.Move(fyne.NewPos(12, (size.Height-dotSize)/2))
	textMin := r.label.MinSize()
	r.label.Resize(textMin)
	r.label.Move(fyne.NewPos(12+dotSize+7, (size.Height-textMin.Height)/2))
}

func (r *capsuleRenderer) MinSize() fyne.Size {
	textMin := r.label.MinSize()
	return fyne.NewSize(12+8+7+textMin.Width+12, textMin.Height+12)
}

func (r *capsuleRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.pill, r.dot, r.label}
}

func (r *capsuleRenderer) Refresh() {
	p := currentPalette()
	if r.capsule.live {
		r.pill.FillColor = p.capsuleLive
		r.dot.FillColor = p.success
		r.label.Color = p.success
	} else {
		r.pill.FillColor = p.capsuleIdle
		r.dot.FillColor = p.textMuted
		r.label.Color = p.textMuted
	}
	r.label.Text = r.capsule.text
	r.pill.Refresh()
	r.dot.Refresh()
	r.label.Refresh()
}

// toggleSwitch — компактный переключатель в стиле iOS/macOS (капсула с
// ползунком), используется вместо стандартного чекбокса там, где нужен
// компактный вкл/выкл-контрол (боковая панель).
type toggleSwitch struct {
	widget.BaseWidget
	checked  bool
	onChange func(bool)
}

var _ fyne.Tappable = (*toggleSwitch)(nil)
var _ desktop.Cursorable = (*toggleSwitch)(nil)

func newToggleSwitch(onChange func(bool)) *toggleSwitch {
	t := &toggleSwitch{onChange: onChange}
	t.ExtendBaseWidget(t)
	return t
}

func (t *toggleSwitch) SetChecked(checked bool) {
	if t.checked == checked {
		return
	}
	t.checked = checked
	t.Refresh()
}

func (t *toggleSwitch) Tapped(*fyne.PointEvent) {
	t.checked = !t.checked
	t.Refresh()
	if t.onChange != nil {
		t.onChange(t.checked)
	}
}

func (t *toggleSwitch) Cursor() desktop.Cursor { return desktop.PointerCursor }

const (
	toggleWidth  float32 = 38
	toggleHeight float32 = 22
)

func (t *toggleSwitch) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(color.Transparent)
	track.CornerRadius = toggleHeight / 2
	knob := canvas.NewCircle(color.White)
	r := &toggleSwitchRenderer{toggle: t, track: track, knob: knob}
	r.Refresh()
	return r
}

type toggleSwitchRenderer struct {
	toggle *toggleSwitch
	track  *canvas.Rectangle
	knob   *canvas.Circle
}

func (r *toggleSwitchRenderer) Destroy() {}

func (r *toggleSwitchRenderer) Layout(fyne.Size) {
	r.track.Resize(fyne.NewSize(toggleWidth, toggleHeight))
	const knobSize float32 = 18
	const pad float32 = 2
	x := pad
	if r.toggle.checked {
		x = toggleWidth - knobSize - pad
	}
	r.knob.Resize(fyne.NewSize(knobSize, knobSize))
	r.knob.Move(fyne.NewPos(x, pad))
}

func (r *toggleSwitchRenderer) MinSize() fyne.Size {
	return fyne.NewSize(toggleWidth, toggleHeight)
}

func (r *toggleSwitchRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.track, r.knob}
}

func (r *toggleSwitchRenderer) Refresh() {
	p := currentPalette()
	if r.toggle.checked {
		r.track.FillColor = p.success
	} else {
		r.track.FillColor = p.textMuted
	}
	r.track.Refresh()
	r.Layout(fyne.Size{})
	r.knob.Refresh()
}

// navIconKind selects which small geometric mark a navItem draws. Icons are
// composed from plain canvas primitives (not SVG resources) so their color
// can be swapped in place between muted and accent as selection changes.
type navIconKind int

const (
	iconStatus navIconKind = iota
	iconProfiles
	iconConnection
	iconLog
	iconAbout
)

// navItem — строка бокового меню: иконка + подпись, с подсветкой при наведении
// и заливкой акцентным цветом при выборе, как пункты в Finder/Mail/System Settings.
type navItem struct {
	widget.BaseWidget
	kind     navIconKind
	label    string
	selected bool
	hovered  bool
	onTap    func()
}

var _ fyne.Tappable = (*navItem)(nil)
var _ desktop.Hoverable = (*navItem)(nil)
var _ desktop.Cursorable = (*navItem)(nil)

func newNavItem(kind navIconKind, label string, onTap func()) *navItem {
	n := &navItem{kind: kind, label: label, onTap: onTap}
	n.ExtendBaseWidget(n)
	return n
}

func (n *navItem) SetSelected(selected bool) {
	if n.selected == selected {
		return
	}
	n.selected = selected
	n.Refresh()
}

func (n *navItem) Tapped(*fyne.PointEvent) {
	if n.onTap != nil {
		n.onTap()
	}
}

func (n *navItem) MouseIn(*desktop.MouseEvent)    { n.hovered = true; n.Refresh() }
func (n *navItem) MouseMoved(*desktop.MouseEvent) {}
func (n *navItem) MouseOut()                      { n.hovered = false; n.Refresh() }
func (n *navItem) Cursor() desktop.Cursor         { return desktop.PointerCursor }

const navIconBox float32 = 17

// buildNavIconShapes returns the small set of colorable primitives for one
// icon kind, laid out inside a navIconBox x navIconBox square.
func buildNavIconShapes(kind navIconKind) []fyne.CanvasObject {
	switch kind {
	case iconStatus: // ring with a filled center dot — "current state"
		ring := canvas.NewCircle(color.Transparent)
		ring.StrokeWidth = 1.6
		ring.Resize(fyne.NewSize(16, 16))
		ring.Move(fyne.NewPos(0.5, 0.5))
		dot := canvas.NewCircle(color.Transparent)
		dot.Resize(fyne.NewSize(6, 6))
		dot.Move(fyne.NewPos(5.5, 5.5))
		return []fyne.CanvasObject{ring, dot}

	case iconProfiles: // two stacked cards
		back := canvas.NewRectangle(color.Transparent)
		back.CornerRadius = 2.5
		back.Resize(fyne.NewSize(12, 9))
		back.Move(fyne.NewPos(4, 1))
		front := canvas.NewRectangle(color.Transparent)
		front.CornerRadius = 2.5
		front.StrokeWidth = 1.6
		front.Resize(fyne.NewSize(14, 10))
		front.Move(fyne.NewPos(1.5, 6.5))
		return []fyne.CanvasObject{back, front}

	case iconConnection: // two nodes joined by a link
		left := canvas.NewCircle(color.Transparent)
		left.Resize(fyne.NewSize(6, 6))
		left.Move(fyne.NewPos(1, 5.5))
		right := canvas.NewCircle(color.Transparent)
		right.Resize(fyne.NewSize(6, 6))
		right.Move(fyne.NewPos(10, 5.5))
		link := canvas.NewLine(color.Transparent)
		link.StrokeWidth = 1.8
		link.Position1 = fyne.NewPos(6.5, 8.5)
		link.Position2 = fyne.NewPos(10.5, 8.5)
		return []fyne.CanvasObject{link, left, right}

	case iconLog: // three ruled lines of decreasing width
		l1 := canvas.NewLine(color.Transparent)
		l1.StrokeWidth = 1.8
		l1.Position1 = fyne.NewPos(1, 3)
		l1.Position2 = fyne.NewPos(16, 3)
		l2 := canvas.NewLine(color.Transparent)
		l2.StrokeWidth = 1.8
		l2.Position1 = fyne.NewPos(1, 8.5)
		l2.Position2 = fyne.NewPos(16, 8.5)
		l3 := canvas.NewLine(color.Transparent)
		l3.StrokeWidth = 1.8
		l3.Position1 = fyne.NewPos(1, 14)
		l3.Position2 = fyne.NewPos(11, 14)
		return []fyne.CanvasObject{l1, l2, l3}

	default: // iconAbout: ring with an "i" (dot + stem)
		ring := canvas.NewCircle(color.Transparent)
		ring.StrokeWidth = 1.6
		ring.Resize(fyne.NewSize(16, 16))
		ring.Move(fyne.NewPos(0.5, 0.5))
		dot := canvas.NewCircle(color.Transparent)
		dot.Resize(fyne.NewSize(2.4, 2.4))
		dot.Move(fyne.NewPos(7.8, 4.6))
		stem := canvas.NewLine(color.Transparent)
		stem.StrokeWidth = 1.8
		stem.Position1 = fyne.NewPos(9, 8.5)
		stem.Position2 = fyne.NewPos(9, 13)
		return []fyne.CanvasObject{ring, dot, stem}
	}
}

// tintNavIconShapes applies the current icon color to every primitive built
// by buildNavIconShapes, filling closed shapes and stroking lines/rings.
func tintNavIconShapes(shapes []fyne.CanvasObject, c color.Color) {
	for _, obj := range shapes {
		switch shape := obj.(type) {
		case *canvas.Circle:
			if shape.StrokeWidth > 0 {
				shape.StrokeColor = c
			} else {
				shape.FillColor = c
			}
		case *canvas.Rectangle:
			if shape.StrokeWidth > 0 {
				shape.StrokeColor = c
			} else {
				shape.FillColor = c
			}
		case *canvas.Line:
			shape.StrokeColor = c
		}
		obj.Refresh()
	}
}

func (n *navItem) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 8
	iconShapes := buildNavIconShapes(n.kind)
	basePositions := make([]fyne.Position, len(iconShapes))
	for i, shape := range iconShapes {
		basePositions[i] = shape.Position()
	}
	label := canvas.NewText(n.label, color.Black)
	label.TextSize = 13
	r := &navItemRenderer{item: n, bg: bg, iconShapes: iconShapes, iconBasePos: basePositions, label: label}
	r.Refresh()
	return r
}

type navItemRenderer struct {
	item        *navItem
	bg          *canvas.Rectangle
	iconShapes  []fyne.CanvasObject
	iconBasePos []fyne.Position
	label       *canvas.Text
}

func (r *navItemRenderer) Destroy() {}

func (r *navItemRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	const leftPad float32 = 10
	iconY := (size.Height - navIconBox) / 2
	for i, shape := range r.iconShapes {
		base := r.iconBasePos[i]
		shape.Move(fyne.NewPos(leftPad+base.X, iconY+base.Y))
	}
	textMin := r.label.MinSize()
	r.label.Resize(textMin)
	r.label.Move(fyne.NewPos(leftPad+navIconBox+9, (size.Height-textMin.Height)/2))
}

func (r *navItemRenderer) MinSize() fyne.Size {
	return fyne.NewSize(210, 34)
}

func (r *navItemRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, 0, len(r.iconShapes)+2)
	objs = append(objs, r.bg)
	objs = append(objs, r.iconShapes...)
	objs = append(objs, r.label)
	return objs
}

func (r *navItemRenderer) Refresh() {
	p := currentPalette()
	switch {
	case r.item.selected:
		r.bg.FillColor = p.navSelectedBg
	case r.item.hovered:
		r.bg.FillColor = p.navHoverBg
	default:
		r.bg.FillColor = color.Transparent
	}
	textColor := p.textMuted
	if r.item.selected {
		textColor = p.accent
	}
	r.label.Text = r.item.label
	r.label.Color = textColor
	r.label.TextStyle = fyne.TextStyle{Bold: r.item.selected}
	tintNavIconShapes(r.iconShapes, textColor)
	r.bg.Refresh()
	r.label.Refresh()
}

// iconTile — скруглённая плитка с градиентной заливкой и белым брендовым
// значком поверх, как иконка приложения в шапке боковой панели.
type iconTile struct {
	widget.BaseWidget
	glyph fyne.Resource
	size  float32
}

func newIconTile(glyph fyne.Resource, size float32) *iconTile {
	t := &iconTile{glyph: glyph, size: size}
	t.ExtendBaseWidget(t)
	return t
}

func (t *iconTile) CreateRenderer() fyne.WidgetRenderer {
	corner := t.size * 0.24
	bg := canvas.NewRaster(func(w, h int) image.Image {
		p := currentPalette()
		return roundedGradient(w, h, corner, p.heroFrom, p.heroTo)
	})
	glyph := canvas.NewImageFromResource(t.glyph)
	glyph.FillMode = canvas.ImageFillContain
	r := &iconTileRenderer{tile: t, bg: bg, glyph: glyph}
	return r
}

type iconTileRenderer struct {
	tile  *iconTile
	bg    *canvas.Raster
	glyph *canvas.Image
}

func (r *iconTileRenderer) Destroy() {}

func (r *iconTileRenderer) Layout(fyne.Size) {
	s := r.tile.size
	r.bg.Resize(fyne.NewSize(s, s))
	inset := s * 0.26
	r.glyph.Resize(fyne.NewSize(s-inset*2, s-inset*2))
	r.glyph.Move(fyne.NewPos(inset, inset))
}

func (r *iconTileRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.tile.size, r.tile.size)
}

func (r *iconTileRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.glyph}
}

func (r *iconTileRenderer) Refresh() {
	r.bg.Refresh()
	r.glyph.Refresh()
}

// roundedGradient rasterizes a diagonal two-color gradient clipped to a
// rounded-rectangle mask, with a ~1px antialiased edge. Used for the sidebar
// app-icon tile and the status page's hero card.
func roundedGradient(w, h int, corner float32, from, to color.Color) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return img
	}
	fr, fg, fb, fa := from.RGBA()
	tr, tg, tb, ta := to.RGBA()
	fw, fh := float64(w), float64(h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			alpha := roundedRectCoverage(float64(x)+0.5, float64(y)+0.5, fw, fh, float64(corner))
			if alpha <= 0 {
				continue
			}
			t := ((float64(x)/fw + float64(y)/fh) / 2)
			mix := func(a, b uint32) uint8 {
				v := uint32(float64(a)*(1-t) + float64(b)*t)
				return uint8(v >> 8)
			}
			a := (float64(fa)*(1-t) + float64(ta)*t) / 0xffff
			img.SetNRGBA(x, y, color.NRGBA{
				R: mix(fr, tr), G: mix(fg, tg), B: mix(fb, tb),
				A: uint8(alpha * a * 255),
			})
		}
	}
	return img
}

// roundedRectCoverage returns 1 inside a corner-radius rectangle, 0 outside,
// and a linear falloff across the ~1px boundary for antialiasing.
func roundedRectCoverage(x, y, w, h, corner float64) float64 {
	if corner <= 0 {
		return 1
	}
	var ccx, ccy float64
	switch {
	case x < corner && y < corner:
		ccx, ccy = corner, corner
	case x > w-corner && y < corner:
		ccx, ccy = w-corner, corner
	case x < corner && y > h-corner:
		ccx, ccy = corner, h-corner
	case x > w-corner && y > h-corner:
		ccx, ccy = w-corner, h-corner
	default:
		return 1
	}
	d := math.Hypot(x-ccx, y-ccy)
	switch {
	case d <= corner-0.5:
		return 1
	case d >= corner+0.5:
		return 0
	default:
		return corner + 0.5 - d
	}
}

// sparkline — компактный график последних значений (используется для трафика
// приёма/отдачи), заливка мягким градиентом под линией, как в системных
// виджетах Activity Monitor / System Settings.
type sparkline struct {
	widget.BaseWidget
	samples []float64
	stroke  color.Color
}

func newSparkline(stroke color.Color) *sparkline {
	s := &sparkline{stroke: stroke}
	s.ExtendBaseWidget(s)
	return s
}

func (s *sparkline) SetSamples(samples []float64) {
	s.samples = samples
	s.Refresh()
}

func (s *sparkline) CreateRenderer() fyne.WidgetRenderer {
	raster := canvas.NewRaster(s.draw)
	return &sparklineRenderer{sparkline: s, raster: raster}
}

func (s *sparkline) draw(w, h int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 || len(s.samples) < 2 {
		return img
	}

	maxV := 0.0
	for _, v := range s.samples {
		if v > maxV {
			maxV = v
		}
	}
	if maxV <= 0 {
		maxV = 1
	}

	n := len(s.samples)
	strokeR, strokeG, strokeB, _ := s.stroke.RGBA()
	topPad := float64(h) * 0.12

	// yAt returns the curve height (in pixels from the top) at pixel column x,
	// via linear interpolation between the two nearest samples.
	yAt := func(x int) float64 {
		pos := float64(x) / float64(w-1) * float64(n-1)
		i := int(pos)
		if i >= n-1 {
			i = n - 2
		}
		frac := pos - float64(i)
		v := s.samples[i]*(1-frac) + s.samples[i+1]*frac
		usable := float64(h) - topPad
		return topPad + usable*(1-v/maxV)
	}

	for x := 0; x < w; x++ {
		curveY := yAt(x)
		for y := 0; y < h; y++ {
			fy := float64(y)
			if fy < curveY-1 {
				continue // above the line: fully transparent
			}
			// Soft vertical gradient fill under the curve, fading toward the bottom.
			depth := (fy - curveY) / (float64(h) - curveY + 1)
			alpha := 0.0
			switch {
			case fy < curveY+1.4:
				alpha = 0.9 // the stroke itself, drawn slightly thicker via near-line band
			default:
				alpha = math.Max(0, 0.32*(1-depth))
			}
			if alpha <= 0.01 {
				continue
			}
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8(strokeR >> 8), G: uint8(strokeG >> 8), B: uint8(strokeB >> 8),
				A: uint8(alpha * 255),
			})
		}
	}
	return img
}

type sparklineRenderer struct {
	sparkline *sparkline
	raster    *canvas.Raster
}

func (r *sparklineRenderer) Destroy() {}
func (r *sparklineRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}
func (r *sparklineRenderer) MinSize() fyne.Size           { return fyne.NewSize(60, 34) }
func (r *sparklineRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.raster} }
func (r *sparklineRenderer) Refresh()                     { r.raster.Refresh() }
