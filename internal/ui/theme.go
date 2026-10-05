package ui

import (
	"image/color"
	"math"

	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/appearance"
)

// Palette holds the semantic colours the UI draws with.
type Palette struct {
	Bg         color.NRGBA // window background
	Surface    color.NRGBA // cards
	SurfaceAlt color.NRGBA // hovered and selected rows, chips
	Fg         color.NRGBA // primary text
	Muted      color.NRGBA // secondary text
	Border     color.NRGBA // card and field outlines
	Accent     color.NRGBA // primary actions and selection
	OnAccent   color.NRGBA // text on Accent
	Tracking   color.NRGBA
	Affiliate  color.NRGBA
	Redirect   color.NRGBA
	WarnBg     color.NRGBA
	WarnFg     color.NRGBA
	ErrBg      color.NRGBA
	ErrFg      color.NRGBA
	Scrim      color.NRGBA
}

func rgb(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xFF}
}

var (
	lightPalette = Palette{
		Bg: rgb(0xF2F3F5), Surface: rgb(0xFFFFFF), SurfaceAlt: rgb(0xECEEF2),
		Fg: rgb(0x1B1D22), Muted: rgb(0x5A606B), Border: rgb(0xD6DAE1),
		Tracking: rgb(0xB3261E), Affiliate: rgb(0x8A5000), Redirect: rgb(0x0B57D0),
		WarnBg: rgb(0xFFF3DC), WarnFg: rgb(0x5C3B00), ErrBg: rgb(0xFDECEA), ErrFg: rgb(0x8C1D18),
		Scrim: color.NRGBA{A: 0x8C},
	}
	darkPalette = Palette{
		Bg: rgb(0x15171B), Surface: rgb(0x1F2228), SurfaceAlt: rgb(0x2A2E36),
		Fg: rgb(0xE6E8EC), Muted: rgb(0xA4AAB5), Border: rgb(0x383D47),
		Tracking: rgb(0xFF8A80), Affiliate: rgb(0xF5C26B), Redirect: rgb(0x8AB4F8),
		WarnBg: rgb(0x3A2C10), WarnFg: rgb(0xFFE2AE), ErrBg: rgb(0x3D1D1B), ErrFg: rgb(0xFFB4AB),
		Scrim: color.NRGBA{A: 0xA6},
	}
	// defaultAccent is used when the system reports none.
	defaultAccent = rgb(0x3D6FD9)
)

// NewPalette derives the palette for the host appearance.
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
	accent := defaultAccent
	if s.HasAccent {
		accent = s.Accent
	}
	p.Accent, p.OnAccent = fitAccent(accent, p.Surface, s.Dark)
	return p
}

// fitAccent adjusts accent until it stands out from surface (3:1) and its
// text colour, black or white, reaches 4.5:1. In dark mode the accent is
// lightened, otherwise darkened; both raise the two ratios monotonically.
func fitAccent(accent, surface color.NRGBA, dark bool) (color.NRGBA, color.NRGBA) {
	target := color.NRGBA{A: 0xFF}
	if dark {
		target = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	}
	for step := 0; step <= 50; step++ {
		c := mix(accent, target, float64(step)/50)
		on := onColour(c)
		if contrast(c, surface) >= 3 && contrast(on, c) >= 4.5 {
			return c, on
		}
	}
	return target, onColour(target)
}

// onColour returns black or white, whichever contrasts more with c.
func onColour(c color.NRGBA) color.NRGBA {
	black, white := color.NRGBA{A: 0xFF}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	if contrast(black, c) >= contrast(white, c) {
		return black
	}
	return white
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	l := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return color.NRGBA{R: l(a.R, b.R), G: l(a.G, b.G), B: l(a.B, b.B), A: 0xFF}
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

// materialTheme derives the stock widget theme from the palette.
func (p Palette) apply(th *material.Theme) {
	th.Palette = material.Palette{Bg: p.Surface, Fg: p.Fg, ContrastBg: p.Accent, ContrastFg: p.OnAccent}
}
