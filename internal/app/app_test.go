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
