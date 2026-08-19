package main

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

//go:embed assets/Inter-Regular.ttf
var interRegular []byte

//go:embed assets/Inter-Medium.ttf
var interMedium []byte

//go:embed assets/Inter-SemiBold.ttf
var interSemiBold []byte

//go:embed assets/Inter-Bold.ttf
var interBold []byte

var (
	fontRegular  = fyne.NewStaticResource("Inter-Regular.ttf", interRegular)
	fontMedium   = fyne.NewStaticResource("Inter-Medium.ttf", interMedium)
	fontSemiBold = fyne.NewStaticResource("Inter-SemiBold.ttf", interSemiBold)
	fontBold     = fyne.NewStaticResource("Inter-Bold.ttf", interBold)
)

// Палитра в духе Apple Human Interface Guidelines.
type palette struct {
	background   color.Color // фон окна
	card         color.Color // фон карточек
	cardStroke   color.Color // волосяная обводка карточек
	text         color.Color // основной текст
	textMuted    color.Color // вторичный текст
	accent       color.Color // системный синий
	accentPress  color.Color
	success      color.Color // системный зелёный
	danger       color.Color // системный красный
	dangerPress  color.Color
	inputFill    color.Color // заливка полей ввода
	inputStroke  color.Color
	hover        color.Color
	separator    color.Color
	logBackdrop  color.Color // фон журнала
	capsuleIdle  color.Color // капсула статуса «остановлен»
	capsuleLive  color.Color // капсула статуса «работает»
	textOnAccent color.Color

	sidebarBg     color.Color // фон боковой панели навигации
	navSelectedBg color.Color // подложка активного пункта меню
	navHoverBg    color.Color
	heroFrom      color.Color // градиент героя на странице «Статус»
	heroTo        color.Color
	sentAccent    color.Color // цвет графика «Отправлено» (в паре с accent для «Получено»)
}

var lightPalette = palette{
	background:   color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF7, A: 0xFF},
	card:         color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
	cardStroke:   color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x14},
	text:         color.NRGBA{R: 0x1D, G: 0x1D, B: 0x1F, A: 0xFF},
	textMuted:    color.NRGBA{R: 0x6E, G: 0x6E, B: 0x73, A: 0xFF},
	accent:       color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF},
	accentPress:  color.NRGBA{R: 0x00, G: 0x62, B: 0xCC, A: 0xFF},
	success:      color.NRGBA{R: 0x34, G: 0xC7, B: 0x59, A: 0xFF},
	danger:       color.NRGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF},
	dangerPress:  color.NRGBA{R: 0xD7, G: 0x2F, B: 0x26, A: 0xFF},
	inputFill:    color.NRGBA{R: 0xF2, G: 0xF2, B: 0xF7, A: 0xFF},
	inputStroke:  color.NRGBA{R: 0xD1, G: 0xD1, B: 0xD6, A: 0xFF},
	hover:        color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x0A},
	separator:    color.NRGBA{R: 0xE5, G: 0xE5, B: 0xEA, A: 0xFF},
	logBackdrop:  color.NRGBA{R: 0xF7, G: 0xF7, B: 0xF9, A: 0xFF},
	capsuleIdle:  color.NRGBA{R: 0x78, G: 0x78, B: 0x80, A: 0x24},
	capsuleLive:  color.NRGBA{R: 0x34, G: 0xC7, B: 0x59, A: 0x2E},
	textOnAccent: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},

	sidebarBg:     color.NRGBA{R: 0xEC, G: 0xEC, B: 0xF1, A: 0xFF},
	navSelectedBg: color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x1E},
	navHoverBg:    color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x08},
	heroFrom:      color.NRGBA{R: 0x2F, G: 0x6F, B: 0xFF, A: 0xFF},
	heroTo:        color.NRGBA{R: 0x7B, G: 0x5C, B: 0xFF, A: 0xFF},
	sentAccent:    color.NRGBA{R: 0x8B, G: 0x5C, B: 0xF6, A: 0xFF},
}

var darkPalette = palette{
	background:   color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1E, A: 0xFF},
	card:         color.NRGBA{R: 0x2C, G: 0x2C, B: 0x2E, A: 0xFF},
	cardStroke:   color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x17},
	text:         color.NRGBA{R: 0xF5, G: 0xF5, B: 0xF7, A: 0xFF},
	textMuted:    color.NRGBA{R: 0x98, G: 0x98, B: 0x9D, A: 0xFF},
	accent:       color.NRGBA{R: 0x0A, G: 0x84, B: 0xFF, A: 0xFF},
	accentPress:  color.NRGBA{R: 0x30, G: 0x99, B: 0xFF, A: 0xFF},
	success:      color.NRGBA{R: 0x30, G: 0xD1, B: 0x58, A: 0xFF},
	danger:       color.NRGBA{R: 0xFF, G: 0x45, B: 0x3A, A: 0xFF},
	dangerPress:  color.NRGBA{R: 0xFF, G: 0x66, B: 0x5C, A: 0xFF},
	inputFill:    color.NRGBA{R: 0x3A, G: 0x3A, B: 0x3C, A: 0xFF},
	inputStroke:  color.NRGBA{R: 0x48, G: 0x48, B: 0x4A, A: 0xFF},
	hover:        color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x10},
	separator:    color.NRGBA{R: 0x38, G: 0x38, B: 0x3A, A: 0xFF},
	logBackdrop:  color.NRGBA{R: 0x1C, G: 0x1C, B: 0x1E, A: 0xFF},
	capsuleIdle:  color.NRGBA{R: 0x78, G: 0x78, B: 0x80, A: 0x3D},
	capsuleLive:  color.NRGBA{R: 0x30, G: 0xD1, B: 0x58, A: 0x3D},
	textOnAccent: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},

	sidebarBg:     color.NRGBA{R: 0x18, G: 0x18, B: 0x1A, A: 0xFF},
	navSelectedBg: color.NRGBA{R: 0x0A, G: 0x84, B: 0xFF, A: 0x33},
	navHoverBg:    color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x0C},
	heroFrom:      color.NRGBA{R: 0x3D, G: 0x74, B: 0xFF, A: 0xFF},
	heroTo:        color.NRGBA{R: 0x8A, G: 0x66, B: 0xFF, A: 0xFF},
	sentAccent:    color.NRGBA{R: 0xA0, G: 0x7C, B: 0xFF, A: 0xFF},
}

func currentPalette() palette {
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		return darkPalette
	}
	return lightPalette
}

// luxTheme — тема приложения: типографика Inter, скругления и цвета в стиле macOS.
type luxTheme struct{}

var _ fyne.Theme = (*luxTheme)(nil)

func (luxTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	p := lightPalette
	if variant == theme.VariantDark {
		p = darkPalette
	}

	switch name {
	case theme.ColorNameBackground:
		return p.background
	case theme.ColorNameForeground:
		return p.text
	case theme.ColorNamePrimary, theme.ColorNameHyperlink, theme.ColorNameFocus:
		return p.accent
	case theme.ColorNameButton:
		return p.inputFill
	case theme.ColorNameInputBackground:
		return p.inputFill
	case theme.ColorNameInputBorder:
		return p.inputStroke
	case theme.ColorNameHover:
		return p.hover
	case theme.ColorNamePressed:
		return p.hover
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return p.textMuted
	case theme.ColorNameSeparator:
		return p.separator
	case theme.ColorNameSuccess:
		return p.success
	case theme.ColorNameError:
		return p.danger
	case theme.ColorNameForegroundOnPrimary, theme.ColorNameForegroundOnError, theme.ColorNameForegroundOnSuccess:
		return p.textOnAccent
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return p.card
	case theme.ColorNameSelection:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x0A, G: 0x84, B: 0xFF, A: 0x59}
		}
		return color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0x40}
	case theme.ColorNameScrollBar:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x45}
		}
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x35}
	case theme.ColorNameShadow:
		return color.NRGBA{A: 0x21}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (luxTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		return fontSemiBold
	}
	return fontRegular
}

func (luxTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (luxTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameHeadingText:
		return 26
	case theme.SizeNameSubHeadingText:
		return 15
	case theme.SizeNamePadding:
		return 5
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameLineSpacing:
		return 5
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 9
	case theme.SizeNameSelectionRadius:
		return 7
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 4
	}
	return theme.DefaultTheme().Size(name)
}
