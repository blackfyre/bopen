package ui

import (
	"image"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/blackfyre/bopen/internal/appearance"
	"github.com/blackfyre/bopen/internal/discovery"
)

// TestScreenshots renders the inspector off screen, light and dark, into
// the README's PNG files: BOPEN_SCREENSHOTS=docs go test -run Screenshots
// ./internal/ui. It is skipped otherwise.
func TestScreenshots(t *testing.T) {
	dir := os.Getenv("BOPEN_SCREENSHOTS")
	if dir == "" {
		t.Skip("set BOPEN_SCREENSHOTS to a directory")
	}
	const width, height, scale = 760, 960, 2
	target := "https://www.amazon.de/dp/B0CX23V2ZK?tag=deals-21&utm_source=newsletter&fbclid=IwAR3x"
	raw := "https://www.google.com/url?q=" + url.QueryEscape(target) + "&sa=D"
	for _, dark := range []bool{false, true} {
		w, router := ruleWindow(t, raw)
		w.m.Env.AllBrowsers = []discovery.Browser{
			{ID: "zen", Name: "Zen Browser · Work", Kind: discovery.KindFlatpak},
			{ID: "firefox", Name: "Firefox", Kind: discovery.KindSystem},
			{ID: "chromium", Name: "Chromium", Kind: discovery.KindSystem},
		}
		w.m.Refresh()
		w.syncInspector()
		w.applyLook(appearance.Settings{Dark: dark})
		frame(w, router)

		hw, err := headless.NewWindow(width*scale, height*scale)
		if err != nil {
			t.Skip("no headless GPU:", err)
		}
		var ops op.Ops
		gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: scale, PxPerSp: scale},
			Constraints: layout.Exact(image.Pt(width*scale, height*scale)), Now: nextFrameTime(), Source: router.Source()}
		w.layout(gtx)
		if err := hw.Frame(&ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rectangle{Max: image.Pt(width*scale, height*scale)})
		if err := hw.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		hw.Release()
		name := "inspector-light.png"
		if dark {
			name = "inspector-dark.png"
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
