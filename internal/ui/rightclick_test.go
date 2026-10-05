package ui

import (
	"image"
	"testing"
	"time"

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
		Constraints: layout.Exact(image.Pt(1140, 760)), Now: time.Now(), Source: router.Source()}
	w.handle(gtx)
	w.layout(gtx)
	router.Frame(ops)
}

func TestRightClickTargetsHoveredParameter(t *testing.T) {
	w, _ := ruleWindow(t, "https://news.example.com/article?ref=home&utm_source=x")
	router := new(input.Router)
	scaledFrame(w, router)
	// "https://news.example.com/article?" is 33 monospace characters of
	// about 13.5 px at this scale, after the 24 px inset.
	pos := f32.Pt(24+36*13.5, 80)
	for _, e := range []pointer.Event{
		{Kind: pointer.Move, Position: pos, Source: pointer.Mouse},
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
