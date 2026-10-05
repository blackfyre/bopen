package ui

import (
	"testing"

	"gioui.org/io/key"

	"github.com/blackfyre/bopen/internal/desktopentry"
	"github.com/blackfyre/bopen/internal/discovery"
)

func privateWindow(t *testing.T) (*window, *[]discovery.Browser) {
	m := model(t, "https://example.com/?fbclid=x")
	m.Browsers = []discovery.Browser{
		{ID: "a", Name: "A", PrivateEntry: &desktopentry.Entry{Exec: "a --private-window %u"}},
		{ID: "b", Name: "B"},
	}
	var opened []discovery.Browser
	m.Open = func(b discovery.Browser, u string) error { opened = append(opened, b); return nil }
	w := newWindow(m, m.Env)
	return w, &opened
}

func TestPrivateShortcutOpensPrivately(t *testing.T) {
	w, opened := privateWindow(t)
	router := newRouter(w)
	press(w, router, "P")
	if !w.m.Private || !w.private.Value {
		t.Fatal("P did not turn the option on")
	}
	press(w, router, key.NameReturn)
	if len(*opened) != 1 || !(*opened)[0].OpenPrivate || (*opened)[0].ID != "a" {
		t.Fatalf("opened %+v", *opened)
	}
}

func TestPrivateUnavailableForUnsupportedBrowser(t *testing.T) {
	w, opened := privateWindow(t)
	router := newRouter(w)
	press(w, router, "P")
	press(w, router, key.NameDownArrow) // B has no private window
	if w.m.Private || w.private.Value {
		t.Fatal("private option kept for a browser without one")
	}
	press(w, router, "P")
	if w.m.Private {
		t.Fatal("P turned the option on for an unsupported browser")
	}
	press(w, router, key.NameReturn)
	if len(*opened) != 1 || (*opened)[0].OpenPrivate {
		t.Fatalf("opened %+v", *opened)
	}
}
