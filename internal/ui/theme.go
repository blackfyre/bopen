package ui

import (
	"image/color"
	"math"

	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/appearance"
)

// Palette holds the semantic colours the UI draws with: the Blackfyre
// tokens of blackfyre.ninja, as in the Spooder desktop calibrator.
type Palette struct {
	Bg         color.NRGBA // window background (paper in light mode)
	Surface    color.NRGBA // cards
	SurfaceAlt color.NRGBA // disabled buttons
	Fg         color.NRGBA // primary text (ink)
	Muted      color.NRGBA // secondary text
	Border     color.NRGBA // card and field outlines

	Accent        color.NRGBA // the orange: primary actions and selection
	AccentPressed color.NRGBA // the accent's shadow
	OnAccent      color.NRGBA // text on Accent

	// Neutral buttons (ink in light mode) and quiet buttons (small ones).
	ButtonBg, ButtonFg, ButtonShadow color.NRGBA
	QuietBg, QuietFg, QuietShadow    color.NRGBA

	Tracking  color.NRGBA
	Affiliate color.NRGBA
	Redirect  color.NRGBA
	WarnBg    color.NRGBA
	WarnFg    color.NRGBA
	ErrBg     color.NRGBA
	ErrFg     color.NRGBA
	Scrim     color.NRGBA
}

func rgb(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

// Blackfyre tokens (blackfyre.ninja styles.css; the dark shades as in the
// calibrator). The kind and banner colours are bopen's own roles.
var (
	ink   = rgb(0x1B1B19)
	paper = rgb(0xE6E4DF)

	lightPalette = Palette{
		Bg: paper, Surface: rgb(0xF3F2EE), SurfaceAlt: rgb(0xDEDBD4),
		Fg: ink, Muted: rgb(0x5E5C56), Border: rgb(0xD3D0C9),
		Accent: rgb(0xE2551D), AccentPressed: rgb(0xA93A0F), OnAccent: ink,
		ButtonBg: ink, ButtonFg: rgb(0xF3F2EE), ButtonShadow: rgb(0x000000),
		QuietBg: rgb(0xDEDBD4), QuietFg: ink, QuietShadow: rgb(0xC4C0B8),
		Tracking: rgb(0xA3390C), Affiliate: rgb(0x6E5300), Redirect: rgb(0x2A52A8),
		WarnBg: rgb(0xFFF6C9), WarnFg: rgb(0x6E5300), ErrBg: rgb(0xF6DED4), ErrFg: rgb(0x7F2C09),
		Scrim: color.NRGBA{A: 0x8C},
	}
	darkPalette = Palette{
		Bg: ink, Surface: rgb(0x2C2C29), SurfaceAlt: rgb(0x3A3935),
		Fg: rgb(0xF3F2EE), Muted: rgb(0xA19E96), Border: rgb(0x3A3935),
		Accent: rgb(0xE2551D), AccentPressed: rgb(0xA93A0F), OnAccent: ink,
		ButtonBg: rgb(0x3A3935), ButtonFg: rgb(0xF3F2EE), ButtonShadow: rgb(0x121211),
		QuietBg: rgb(0x242421), QuietFg: rgb(0xF3F2EE), QuietShadow: rgb(0x121211),
		Tracking: rgb(0xFF8A4C), Affiliate: rgb(0xE0AF68), Redirect: rgb(0x7AA2F7),
		WarnBg: rgb(0x33301F), WarnFg: rgb(0xE8DC9A), ErrBg: rgb(0x3D2219), ErrFg: rgb(0xFFB59A),
		Scrim: color.NRGBA{A: 0xA6},
	}
)

// NewPalette returns the Blackfyre palette for the host appearance: light
// (paper) or dark, each with a high-contrast variant. The system accent
// colour is not used; the accent is Blackfyre's orange.
func NewPalette(s appearance.Settings) Palette {
	p := lightPalette
	if s.Dark {
		p = darkPalette
	}
	if s.HighContrast {
		if s.Dark {
			p.Bg, p.Surface, p.SurfaceAlt, p.Fg = rgb(0x000000), rgb(0x000000), rgb(0x1A1A1A), rgb(0xFFFFFF)
		} else {
			p.Bg, p.Surface, p.SurfaceAlt, p.Fg = rgb(0xFFFFFF), rgb(0xFFFFFF), rgb(0xEBEBEB), rgb(0x000000)
		}
		p.Muted, p.Border = p.Fg, p.Fg
	}
	return p
}

// buttonColours are a button's face, text and shadow.
func (p Palette) buttonColours(k buttonKind) (face, fg, shadow color.NRGBA) {
	switch k {
	case kindPrimary:
		return p.Accent, p.OnAccent, p.AccentPressed
	case kindQuiet:
		return p.QuietBg, p.QuietFg, p.QuietShadow
	}
	return p.ButtonBg, p.ButtonFg, p.ButtonShadow
}

// luminance is the WCAG relative luminance of an opaque colour.
func luminance(c color.NRGBA) float64 {
	ch := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(c.R) + 0.7152*ch(c.G) + 0.0722*ch(c.B)
}

// contrast is the WCAG contrast ratio between two opaque colours.
func contrast(a, b color.NRGBA) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// apply sets the stock widget theme from the palette.
func (p Palette) apply(th *material.Theme) {
	th.Palette = material.Palette{Bg: p.Surface, Fg: p.Fg, ContrastBg: p.Accent, ContrastFg: p.OnAccent}
}
