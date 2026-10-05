package ui

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func scaledFrame(w *window, router *input.Router) {
	ops := new(op.Ops)
	gtx := layout.Context{Ops: ops, Metric: unit.Metric{PxPerDp: 1.5, PxPerSp: 1.5},
		Constraints: layout.Exact(image.Pt(1140, 760)), Now: nextFrameTime(), Source: router.Source()}
	w.handle(gtx)
	w.layout(gtx)
	router.Frame(ops)
}

func TestRightClickTargetsHoveredParameter(t *testing.T) {
	w, _ := ruleWindow(t, "https://news.example.com/article?ref=home&utm_source=x")
	router := new(input.Router)
	scaledFrame(w, router)
	ref := paramIndex(t, w.m.Analysis, "ref")
	// Find a point over "ref=home" by moving the pointer down the window
	// until the hovered parameter is ref.
	var pos f32.Point
	found := false
	for y := float32(0); y < 400 && !found; y += 4 {
		for x := float32(16); x < 1100 && !found; x += 12 {
			pos = f32.Pt(x, y)
			router.Queue(pointer.Event{Kind: pointer.Move, Position: pos, Source: pointer.Mouse})
			scaledFrame(w, router)
			found = w.menus.hoverParam == ref
		}
	}
	if !found {
		t.Fatal("never hovered the ref parameter")
	}
	for _, e := range []pointer.Event{
		{Kind: pointer.Press, Position: pos, Buttons: pointer.ButtonSecondary, Source: pointer.Mouse},
		{Kind: pointer.Release, Position: pos, Source: pointer.Mouse},
	} {
		router.Queue(e)
		scaledFrame(w, router)
	}
	if !w.menus.link.Active() {
		t.Fatal("context menu did not open")
	}
	p, ok := w.menuParam()
	if !ok || p.Name != "ref" || p.Host != "news.example.com" {
		t.Fatalf("menu targets %+v (ok %v)", p, ok)
	}
}
