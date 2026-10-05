package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"

	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

type fakeRegistrar struct {
	isDefault, known bool
	registered       int
}

func (f *fakeRegistrar) Register() (string, error) {
	f.registered++
	f.isDefault, f.known = true, true
	return "registered", nil
}
func (f *fakeRegistrar) IsDefault() (bool, bool) { return f.isDefault, f.known }

func settingsWindow(t *testing.T) (*window, *input.Router) {
	m := envModel(t, "https://example.com/?utm_source=x&fbclid=y")
	m.Env.ConfigDir = t.TempDir()
	m.Env.Registrar = &fakeRegistrar{}
	m.Open = func(discovery.Browser, string) error { return nil }
	w := newWindow(m, m.Env)
	router := new(input.Router)
	frame(w, router)
	return w, router
}

func TestSettingsRoundTripKeepsToggles(t *testing.T) {
	w, router := settingsWindow(t)
	w.toggles[1].Value = false
	w.m.Accepted[1] = false
	w.openSettings()
	frame(w, router)
	press(w, router, key.NameEscape)
	if w.view != viewInspector || w.done {
		t.Fatalf("view %v done %v", w.view, w.done)
	}
	if w.m.Accepted[1] || w.toggles[1].Value || !w.toggles[0].Value {
		t.Fatalf("toggles lost: accepted %v", w.m.Accepted)
	}
}

func TestSettingsChangesApplyOnReturn(t *testing.T) {
	w, router := settingsWindow(t)
	w.openSettings()
	w.save(func(c *prefs.Config) { c.Browsers.Hidden = []string{"a"} })
	w.save(func(c *prefs.Config) { c.Rules.Disabled = []string{"utm"} })
	w.save(func(c *prefs.Config) { c.Window = prefs.WindowWhenSuggestions })
	frame(w, router)
	w.closeSettings()
	if len(w.m.Browsers) != 2 || w.m.Browsers[0].ID != "b" {
		t.Fatalf("browsers %v", w.m.Browsers)
	}
	if len(w.m.Analysis.Suggestions) != 1 || len(w.toggles) != 1 {
		t.Fatalf("suggestions %+v", w.m.Analysis.Suggestions)
	}
	cfg, problems := prefs.LoadConfig(w.env.ConfigDir)
	if len(problems) != 0 || !cfg.IsHidden("a") || !cfg.IsDisabled("utm") || cfg.Window != prefs.WindowWhenSuggestions {
		t.Fatalf("saved %+v %v", cfg, problems)
	}
}

func TestSettingsSaveFailureReverts(t *testing.T) {
	w, _ := settingsWindow(t)
	if err := os.WriteFile(filepath.Join(w.env.ConfigDir, "config.toml"), []byte("window = "), 0o644); err != nil {
		t.Fatal(err)
	}
	w.openSettings()
	w.settings.windowMode.Value = string(prefs.WindowWhenSuggestions)
	w.save(func(c *prefs.Config) { c.Window = prefs.WindowWhenSuggestions })
	if !w.settings.isError || !strings.Contains(w.settings.message, "Could not save") {
		t.Fatalf("message %q", w.settings.message)
	}
	if w.settings.windowMode.Value != string(prefs.WindowAlways) {
		t.Fatalf("widget shows unsaved value %q", w.settings.windowMode.Value)
	}
}

func TestStandaloneSettingsEscapeCloses(t *testing.T) {
	m := envModel(t, "https://example.com/")
	w := newWindow(nil, m.Env)
	w.openSettings()
	router := new(input.Router)
	frame(w, router)
	press(w, router, key.NameEscape)
	if !w.done {
		t.Fatal("standalone settings did not close")
	}
}

func TestRegistrationStatusAndReregister(t *testing.T) {
	w, _ := settingsWindow(t)
	w.openSettings()
	if !strings.Contains(w.settings.status, "Unknown") {
		t.Fatalf("status %q", w.settings.status)
	}
	reg := w.env.Registrar.(*fakeRegistrar)
	reg.known = true
	w.refreshStatus()
	if !strings.Contains(w.settings.status, "not the default") {
		t.Fatalf("status %q", w.settings.status)
	}
	if _, err := reg.Register(); err != nil {
		t.Fatal(err)
	}
	w.refreshStatus()
	if !strings.Contains(w.settings.status, "is the default") {
		t.Fatalf("status %q", w.settings.status)
	}
}

func TestRulePattern(t *testing.T) {
	m := envModel(t, "https://example.com/")
	got := map[string]string{}
	for _, r := range m.Env.Builtin {
		got[r.ID] = rulePattern(r)
	}
	if got["redirect-google"] != "www.google.*/url, google.*/url" || got["fbclid"] != "fbclid" ||
		got["affiliate-amazon-tag"] != "tag  on amazon.*, *.amazon.*" {
		t.Fatalf("patterns %v", got)
	}
}
