package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/blackfyre/bopen/internal/discovery"
)

// frame runs one frame of w against router and returns after layout.
func frame(w *window, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Exact(image.Pt(800, 600)), Now: time.Now(), Source: router.Source()}
	w.handle(gtx)
	w.layout(gtx)
	router.Frame(&ops)
}

func press(w *window, router *input.Router, name key.Name) {
	router.Queue(key.Event{Name: name, State: key.Press})
	frame(w, router)
}

func keyedWindow(t *testing.T) (*window, *input.Router, *[]string) {
	m := model(t, "https://example.com/?fbclid=x")
	m.Browsers = []discovery.Browser{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}, {ID: "c", Name: "C"}}
	var opened []string
	m.Open = func(b discovery.Browser, url string) error { opened = append(opened, b.ID+" "+url); return nil }
	w := newWindow(m)
	router := new(input.Router)
	frame(w, router)
	return w, router, &opened
}

func TestKeyEnterOpensPreselected(t *testing.T) {
	w, router, opened := keyedWindow(t)
	press(w, router, key.NameReturn)
	if !w.done || len(*opened) != 1 || (*opened)[0] != "a https://example.com/" {
		t.Fatalf("done %v opened %v", w.done, *opened)
	}
}

func TestKeyEscapeCancels(t *testing.T) {
	w, router, opened := keyedWindow(t)
	press(w, router, key.NameEscape)
	if !w.done || len(*opened) != 0 {
		t.Fatalf("done %v opened %v", w.done, *opened)
	}
}

func TestKeySelection(t *testing.T) {
	w, router, opened := keyedWindow(t)
	press(w, router, key.NameDownArrow)
	press(w, router, key.NameDownArrow)
	press(w, router, key.NameDownArrow)
	if w.m.Selected != 2 || w.browsers.Value != "2" {
		t.Fatalf("selected %d (%q)", w.m.Selected, w.browsers.Value)
	}
	press(w, router, key.NameUpArrow)
	if w.m.Selected != 1 {
		t.Fatalf("selected %d", w.m.Selected)
	}
	press(w, router, "1")
	if w.m.Selected != 0 {
		t.Fatalf("selected %d", w.m.Selected)
	}
	press(w, router, "9")
	if w.m.Selected != 0 {
		t.Fatalf("out-of-range digit changed selection to %d", w.m.Selected)
	}
	press(w, router, "3")
	press(w, router, key.NameReturn)
	if len(*opened) != 1 || (*opened)[0] != "c https://example.com/" {
		t.Fatalf("opened %v", *opened)
	}
}

func TestKeyEnterWithBlockerDoesNothing(t *testing.T) {
	w, router, opened := keyedWindow(t)
	w.m.Blocker = "no"
	press(w, router, key.NameReturn)
	if w.done || len(*opened) != 0 {
		t.Fatalf("done %v opened %v", w.done, *opened)
	}
}
