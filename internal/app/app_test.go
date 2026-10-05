package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

var browsers = []discovery.Browser{{ID: "brave.desktop"}, {ID: "firefox.desktop"}, {ID: "zen.desktop"}}

func TestPreselect(t *testing.T) {
	for _, tc := range []struct {
		st   prefs.State
		want int
	}{
		{prefs.State{LastUsed: "zen.desktop", PreviousDefault: "brave.desktop"}, 2},
		{prefs.State{LastUsed: "gone.desktop", PreviousDefault: "firefox.desktop"}, 1},
		{prefs.State{LastUsed: "gone.desktop", PreviousDefault: "also-gone.desktop"}, 0},
		{prefs.State{}, 0},
	} {
		if got := Preselect(browsers, tc.st); got != tc.want {
			t.Errorf("%+v: got %d, want %d", tc.st, got, tc.want)
		}
	}
	if got := Preselect(nil, prefs.State{LastUsed: "zen.desktop"}); got != -1 {
		t.Errorf("empty list: got %d", got)
	}
}

func analysis(t *testing.T, raw string) *clean.Analysis {
	rules, err := clean.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	return clean.Analyse(raw, rules)
}

func TestNeedWindow(t *testing.T) {
	always := prefs.Config{Window: prefs.WindowAlways}
	when := prefs.Config{Window: prefs.WindowWhenSuggestions}
	clean := analysis(t, "https://example.com/")
	tracked := analysis(t, "https://example.com/?fbclid=x")
	for name, tc := range map[string]struct {
		s    Situation
		want bool
	}{
		"always, clean":        {Situation{Config: always, Analysis: clean, Browsers: browsers}, true},
		"when, clean":          {Situation{Config: when, Analysis: clean, Browsers: browsers}, false},
		"when, tracked":        {Situation{Config: when, Analysis: tracked, Browsers: browsers}, true},
		"when, invalid url":    {Situation{Config: when, ValidationErr: errors.New("x"), Browsers: browsers}, true},
		"when, no browsers":    {Situation{Config: when, Analysis: clean}, true},
		"when, config problem": {Situation{Config: when, Analysis: clean, Browsers: browsers, Problems: []error{errors.New("bad")}}, true},
	} {
		if got := NeedWindow(tc.s); got != tc.want {
			t.Errorf("%s: got %v, want %v", name, got, tc.want)
		}
	}
}

func TestOpenRecordsLastUsedOnSuccess(t *testing.T) {
	dir := t.TempDir()
	st := prefs.State{PreviousDefault: "brave.desktop"}
	if err := Open(dir, st, browsers[2], "https://example.com/", func(discovery.Browser, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if got := prefs.LoadState(dir); got.LastUsed != "zen.desktop" || got.PreviousDefault != "brave.desktop" {
		t.Fatalf("state = %+v", got)
	}
	if got := Preselect(browsers, prefs.LoadState(dir)); got != 2 {
		t.Fatalf("next invocation pre-selects %d", got)
	}
}

func TestOpenFailureLeavesStateUnchanged(t *testing.T) {
	dir := t.TempDir()
	if err := prefs.SaveState(dir, prefs.State{LastUsed: "brave.desktop"}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(dir, "state.toml"))
	err := Open(dir, prefs.LoadState(dir), browsers[2], "https://example.com/",
		func(discovery.Browser, string) error { return errors.New("exec failed") })
	if err == nil {
		t.Fatal("expected error")
	}
	after, _ := os.ReadFile(filepath.Join(dir, "state.toml"))
	if string(before) != string(after) {
		t.Fatalf("state changed: %q -> %q", before, after)
	}
}

func TestVisibleOrderAndHidden(t *testing.T) {
	all := []discovery.Browser{{ID: "brave.desktop"}, {ID: "firefox.desktop"}, {ID: "zen.desktop"}}
	cfg := prefs.Config{Browsers: prefs.BrowsersConfig{
		Order:  []string{"gone.desktop", "zen.desktop"},
		Hidden: []string{"firefox.desktop"},
	}}
	got := Visible(all, cfg)
	if len(got) != 2 || got[0].ID != "zen.desktop" || got[1].ID != "brave.desktop" {
		t.Fatalf("got %v", got)
	}
	cfg = prefs.Config{Browsers: prefs.BrowsersConfig{Order: []string{"zen.desktop"}}}
	if got := Visible(all, cfg); got[0].ID != "zen.desktop" || got[1].ID != "brave.desktop" || got[2].ID != "firefox.desktop" {
		t.Fatalf("ordered then unordered: got %v", got)
	}
}

func TestPreselectSkipsHiddenLastUsed(t *testing.T) {
	all := []discovery.Browser{{ID: "brave.desktop"}, {ID: "firefox.desktop"}}
	cfg := prefs.Config{Browsers: prefs.BrowsersConfig{Hidden: []string{"firefox.desktop"}}}
	visible := Visible(all, cfg)
	st := prefs.State{LastUsed: "firefox.desktop", PreviousDefault: "brave.desktop"}
	if i := Preselect(visible, st); i != 0 || visible[i].ID != "brave.desktop" {
		t.Fatalf("pre-selected %d", i)
	}
}

func TestAllHiddenNeedsWindow(t *testing.T) {
	all := []discovery.Browser{{ID: "brave.desktop"}}
	cfg := prefs.Config{Window: prefs.WindowWhenSuggestions, Browsers: prefs.BrowsersConfig{Hidden: []string{"brave.desktop"}}}
	visible := Visible(all, cfg)
	if len(visible) != 0 || Preselect(visible, prefs.State{}) != -1 {
		t.Fatalf("visible %v", visible)
	}
	if !NeedWindow(Situation{Config: cfg, Analysis: analysis(t, "https://example.com/"), Browsers: visible}) {
		t.Fatal("all-hidden must show the window")
	}
}

func TestOrderedIncludesHidden(t *testing.T) {
	all := []discovery.Browser{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	cfg := prefs.Config{Browsers: prefs.BrowsersConfig{Order: []string{"c"}, Hidden: []string{"c"}}}
	got := Ordered(all, cfg)
	if len(got) != 3 || got[0].ID != "c" {
		t.Fatalf("got %v", got)
	}
}

func TestMoveOrder(t *testing.T) {
	ordered := []discovery.Browser{{ID: "brave"}, {ID: "zen"}, {ID: "ff"}}
	got := MoveOrder(ordered, []string{"gone", "zen"}, 1, -1)
	want := []string{"zen", "brave", "ff", "gone"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if got := MoveOrder(ordered, nil, 0, -1); got[0] != "brave" {
		t.Fatalf("moving the first up changed the order: %v", got)
	}
}

func userConfig(rules ...prefs.UserRule) prefs.Config {
	return prefs.Config{Rules: prefs.RulesConfig{User: rules}}
}

func TestUserRuleOnHost(t *testing.T) {
	builtin, _ := clean.Builtin()
	cfg := userConfig(prefs.UserRule{ID: "u-1", Kind: "tracking", Param: "ref", Hosts: []string{"news.example.com"}, Reason: "Referrer tracking"})
	a := clean.Analyse("https://news.example.com/a?ref=home", Rules(builtin, cfg, nil))
	if len(a.Suggestions) != 1 {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
	s := a.Suggestions[0]
	if s.Kind != clean.KindTracking || s.Source != clean.SourceUser || s.Reason != "Referrer tracking" || s.RuleID != "u-1" || s.Text != "ref=home" {
		t.Fatalf("suggestion %+v", s)
	}
	if a := clean.Analyse("https://other.example.com/a?ref=home", Rules(builtin, cfg, nil)); len(a.Suggestions) != 0 {
		t.Fatalf("scope ignored: %+v", a.Suggestions)
	}
}

func TestUserRuleOverridesBuiltin(t *testing.T) {
	builtin, _ := clean.Builtin()
	cfg := userConfig(prefs.UserRule{ID: "u-2", Kind: "tracking", Param: "fbclid", Reason: "Facebook tracking, always remove"})
	a := clean.Analyse("https://example.com/?fbclid=x", Rules(builtin, cfg, nil))
	if len(a.Suggestions) != 1 || a.Suggestions[0].Source != clean.SourceUser || a.Suggestions[0].Reason != "Facebook tracking, always remove" {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
}

func TestInvalidUserRuleNotApplied(t *testing.T) {
	builtin, _ := clean.Builtin()
	cfg := userConfig(prefs.UserRule{ID: "u-3", Kind: "tracking", Param: "ref", Reason: ""})
	if a := clean.Analyse("https://example.com/?ref=x", Rules(builtin, cfg, nil)); len(a.Suggestions) != 0 {
		t.Fatalf("invalid rule applied: %+v", a.Suggestions)
	}
}

func TestDisabledBuiltinDoesNotBlockUserRule(t *testing.T) {
	builtin, _ := clean.Builtin()
	cfg := userConfig(prefs.UserRule{ID: "u-4", Kind: "affiliate", Param: "utm_*", Reason: "mine"})
	cfg.Rules.Disabled = []string{"utm"}
	a := clean.Analyse("https://example.com/?utm_source=x", Rules(builtin, cfg, nil))
	if len(a.Suggestions) != 1 || a.Suggestions[0].Source != clean.SourceUser || a.Suggestions[0].Default {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
}
