package ui

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/blackfyre/bopen/internal/appearance"
)

func TestPaletteLegibility(t *testing.T) {
	accents := map[string]*color.NRGBA{
		"default": nil,
		"cosmic":  {R: 122, G: 162, B: 247, A: 255},
		"windows": {R: 0x00, G: 0x78, B: 0xD4, A: 255},
		"yellow":  {R: 255, G: 255, B: 0, A: 255},
		"white":   {R: 255, G: 255, B: 255, A: 255},
		"black":   {A: 255},
		"red":     {R: 230, G: 30, B: 30, A: 255},
	}
	for _, dark := range []bool{false, true} {
		for _, high := range []bool{false, true} {
			for name, accent := range accents {
				s := appearance.Settings{Dark: dark, HighContrast: high}
				if accent != nil {
					s.Accent, s.HasAccent = *accent, true
				}
				p := NewPalette(s)
				label := fmt.Sprintf("dark=%v high=%v accent=%s", dark, high, name)
				check := func(what string, fg, bg color.NRGBA, min float64) {
					if r := contrast(fg, bg); r < min {
						t.Errorf("%s: %s %.2f < %.1f", label, what, r, min)
					}
				}
				for _, bg := range []struct {
					name string
					c    color.NRGBA
				}{{"bg", p.Bg}, {"surface", p.Surface}, {"surfaceAlt", p.SurfaceAlt}} {
					check("fg on "+bg.name, p.Fg, bg.c, 4.5)
					check("muted on "+bg.name, p.Muted, bg.c, 4.5)
					check("tracking on "+bg.name, p.Tracking, bg.c, 4.5)
					check("affiliate on "+bg.name, p.Affiliate, bg.c, 4.5)
					check("redirect on "+bg.name, p.Redirect, bg.c, 4.5)
				}
				check("warning text", p.WarnFg, p.WarnBg, 4.5)
				check("error text", p.ErrFg, p.ErrBg, 4.5)
				check("accent text", p.OnAccent, p.Accent, 4.5)
				check("accent on surface", p.Accent, p.Surface, 3)
				if high && p.Muted != p.Fg {
					t.Errorf("%s: muted text in high contrast", label)
				}
			}
		}
	}
}

func TestSystemAccentUsedWhenLegible(t *testing.T) {
	// A mid-tone accent that already works is used unchanged.
	accent := color.NRGBA{R: 0x00, G: 0x5A, B: 0xC8, A: 0xFF}
	p := NewPalette(appearance.Settings{Accent: accent, HasAccent: true})
	if p.Accent != accent || p.OnAccent != (color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}) {
		t.Fatalf("accent %v on %v", p.Accent, p.OnAccent)
	}
	cosmic := color.NRGBA{R: 122, G: 162, B: 247, A: 255}
	if p := NewPalette(appearance.Settings{Dark: true, Accent: cosmic, HasAccent: true}); p.Accent != cosmic {
		t.Fatalf("COSMIC accent changed in dark mode: %v", p.Accent)
	}
}

func TestAppearanceChangeRethemesWindow(t *testing.T) {
	w, router := ruleWindow(t, "https://example.com/?fbclid=x")
	if w.pal.Bg != lightPalette.Bg {
		t.Fatal("window does not start light")
	}
	w.looks <- appearance.Settings{Dark: true}
	frame(w, router)
	if w.pal.Bg != darkPalette.Bg || w.th.Palette.Bg != darkPalette.Surface || w.th.Palette.Fg != darkPalette.Fg {
		t.Fatalf("not re-themed: %+v", w.pal)
	}
	w.looks <- appearance.Settings{Dark: true, HighContrast: true}
	frame(w, router)
	if w.pal.Muted != w.pal.Fg {
		t.Fatal("high contrast not applied")
	}
}
