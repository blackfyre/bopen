package ui

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/blackfyre/bopen/internal/appearance"
)

func TestPaletteLegibility(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, high := range []bool{false, true} {
			p := NewPalette(appearance.Settings{Dark: dark, HighContrast: high})
			label := fmt.Sprintf("dark=%v high=%v", dark, high)
			check := func(what string, fg, bg color.NRGBA, min float64) {
				if r := contrast(fg, bg); r < min {
					t.Errorf("%s: %s %.2f < %.1f", label, what, r, min)
				}
			}
			for _, bg := range []struct {
				name string
				c    color.NRGBA
			}{{"bg", p.Bg}, {"surface", p.Surface}} {
				check("fg on "+bg.name, p.Fg, bg.c, 4.5)
				check("muted on "+bg.name, p.Muted, bg.c, 4.5)
				check("tracking on "+bg.name, p.Tracking, bg.c, 4.5)
				check("affiliate on "+bg.name, p.Affiliate, bg.c, 4.5)
				check("redirect on "+bg.name, p.Redirect, bg.c, 4.5)
			}
			check("warning text", p.WarnFg, p.WarnBg, 4.5)
			check("error text", p.ErrFg, p.ErrBg, 4.5)
			for _, k := range []buttonKind{kindNeutral, kindPrimary, kindQuiet} {
				face, fg, _ := p.buttonColours(k)
				check(fmt.Sprintf("button %d text", k), fg, face, 4.5)
			}
			check("disabled button", p.Muted, p.SurfaceAlt, 3)
			check("accent on surface", p.Accent, p.Surface, 3)
			if high && p.Muted != p.Fg {
				t.Errorf("%s: muted text in high contrast", label)
			}
		}
	}
}

// Spec "Blue system accent": the Blackfyre orange with ink text.
func TestBlackfyreAccentIgnoresSystemAccent(t *testing.T) {
	cosmic := color.NRGBA{R: 122, G: 162, B: 247, A: 255}
	for _, dark := range []bool{false, true} {
		p := NewPalette(appearance.Settings{Dark: dark, Accent: cosmic, HasAccent: true})
		if p.Accent != rgb(0xE2551D) || p.OnAccent != rgb(0x1B1B19) {
			t.Fatalf("dark=%v: accent %v on %v", dark, p.Accent, p.OnAccent)
		}
	}
	if p := NewPalette(appearance.Settings{}); p.Bg != rgb(0xE6E4DF) || p.Fg != rgb(0x1B1B19) {
		t.Fatalf("light %+v", p)
	}
	if p := NewPalette(appearance.Settings{Dark: true}); p.Bg != rgb(0x1B1B19) || p.Fg != rgb(0xF3F2EE) {
		t.Fatalf("dark %+v", p)
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
