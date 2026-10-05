package ui

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

func model(t *testing.T, raw string) *Model {
	rules, err := clean.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	a := clean.Analyse(raw, rules)
	return &Model{Input: raw, Analysis: a, Accepted: a.Defaults(),
		Browsers: []discovery.Browser{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}}}
}

func TestRunsCoverURLAndMarkKinds(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/?fbclid=x&id=1")
	m := model(t, raw)
	var b strings.Builder
	tracking := ""
	for _, r := range m.runs() {
		b.WriteString(r.text)
		if r.kind == clean.KindTracking {
			tracking += r.text
		}
	}
	if b.String() != raw {
		t.Fatalf("runs do not reproduce the URL: %q", b.String())
	}
	if tracking != "fbclid%3Dx" {
		t.Fatalf("tracking runs = %q", tracking)
	}
}

func TestRunsReflectRejection(t *testing.T) {
	m := model(t, "https://example.com/?utm_source=x")
	m.Accepted[0] = false
	for _, r := range m.runs() {
		if r.kind == clean.KindTracking && r.accepted {
			t.Fatal("rejected suggestion drawn as accepted")
		}
	}
}

func TestResultUpdatesWithToggles(t *testing.T) {
	m := model(t, "https://example.com/?a=1&utm_source=x")
	if got := m.Result(); got != "https://example.com/?a=1" {
		t.Fatalf("result %q", got)
	}
	m.Accepted[0] = false
	if got := m.Result(); got != "https://example.com/?a=1&utm_source=x" {
		t.Fatalf("result %q", got)
	}
}

func TestMoveClamps(t *testing.T) {
	m := model(t, "https://example.com/")
	m.Move(-1)
	if m.Selected != 0 {
		t.Fatal(m.Selected)
	}
	m.Move(5)
	if m.Selected != 1 {
		t.Fatal(m.Selected)
	}
}

func TestOpenSelectedFailureKeepsWindow(t *testing.T) {
	m := model(t, "https://example.com/?fbclid=x")
	var opened string
	m.Open = func(b discovery.Browser, u string) error { opened = b.ID + " " + u; return nil }
	m.Selected = 1
	if !m.OpenSelected() || opened != "b https://example.com/" {
		t.Fatalf("opened %q", opened)
	}
	m.Open = func(discovery.Browser, string) error { return errors.New("boom") }
	if m.OpenSelected() || !strings.Contains(m.LaunchError, "boom") {
		t.Fatalf("launch error %q", m.LaunchError)
	}
}

func TestBlockerPreventsOpen(t *testing.T) {
	m := model(t, "https://example.com/")
	m.Blocker = "no browsers"
	m.Open = func(discovery.Browser, string) error { t.Fatal("opened"); return nil }
	if m.OpenSelected() {
		t.Fatal("opened despite blocker")
	}
}

func envModel(t *testing.T, raw string) *Model {
	rules, err := clean.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	m := &Model{Input: raw, Analysis: clean.Analyse(raw, rules), Env: &Env{
		Config:      prefs.DefaultConfig(),
		Builtin:     rules,
		AllBrowsers: []discovery.Browser{{ID: "a", Name: "A"}, {ID: "b", Name: "B"}, {ID: "c", Name: "C"}},
	}}
	m.Accepted = m.Analysis.Defaults()
	m.Refresh()
	return m
}

func TestRefreshCarriesTogglesBySpan(t *testing.T) {
	m := envModel(t, "https://example.com/?utm_source=x&fbclid=y")
	if len(m.Analysis.Suggestions) != 2 {
		t.Fatalf("suggestions %+v", m.Analysis.Suggestions)
	}
	m.Accepted[1] = false // keep fbclid
	_ = m.Env.Update(func(c *prefs.Config) { c.Rules.Disabled = []string{"utm"} })
	m.Refresh()
	if len(m.Analysis.Suggestions) != 1 || m.Analysis.Suggestions[0].Text != "fbclid=y" || m.Accepted[0] {
		t.Fatalf("after disabling utm: %+v accepted %v", m.Analysis.Suggestions, m.Accepted)
	}
	_ = m.Env.Update(func(c *prefs.Config) { c.Rules.Disabled = nil })
	m.Refresh()
	if len(m.Analysis.Suggestions) != 2 || !m.Accepted[0] || m.Accepted[1] {
		t.Fatalf("after re-enabling: accepted %v", m.Accepted)
	}
}

func TestRefreshKeepsSelectionAndHonoursHidden(t *testing.T) {
	m := envModel(t, "https://example.com/")
	m.Selected = 2 // C
	_ = m.Env.Update(func(c *prefs.Config) { c.Browsers.Hidden = []string{"a"} })
	m.Refresh()
	if len(m.Browsers) != 2 || m.Browsers[m.Selected].ID != "c" {
		t.Fatalf("browsers %v selected %d", m.Browsers, m.Selected)
	}
	_ = m.Env.Update(func(c *prefs.Config) { c.Browsers.Hidden = []string{"a", "b", "c"} })
	m.Refresh()
	if m.CanOpen() || !strings.Contains(m.Blocker, "hidden") {
		t.Fatalf("blocker %q", m.Blocker)
	}
}

func TestRunsMarkParameters(t *testing.T) {
	m := model(t, "https://example.com/a?ref=home&fbclid=x")
	got := map[string]int{}
	for _, r := range m.runs() {
		got[r.text] = r.param
	}
	if got["ref=home"] < 0 || got["fbclid=x"] < 0 || got["https://example.com/a?"] != -1 {
		t.Fatalf("runs %v", got)
	}
	if p := m.Analysis.Params[got["ref=home"]]; p.Name != "ref" {
		t.Fatalf("param %+v", p)
	}
}
