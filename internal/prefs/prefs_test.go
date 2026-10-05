package prefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, configFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	cfg, problems := LoadConfig(t.TempDir())
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if cfg.Window != WindowAlways {
		t.Fatalf("window = %q, want %q", cfg.Window, WindowAlways)
	}
}

func TestLoadConfigValid(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `window = "when-suggestions"`)
	cfg, problems := LoadConfig(dir)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if cfg.Window != WindowWhenSuggestions {
		t.Fatalf("window = %q", cfg.Window)
	}
}

func TestLoadConfigMalformed(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `window = "when-suggestions`)
	cfg, problems := LoadConfig(dir)
	if len(problems) != 1 {
		t.Fatalf("problems = %v, want one", problems)
	}
	if !strings.Contains(problems[0].Error(), configFile) {
		t.Fatalf("problem does not name the file: %v", problems[0])
	}
	if cfg.Window != WindowAlways {
		t.Fatalf("window = %q, want default", cfg.Window)
	}
}

func TestLoadConfigInvalidValue(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `window = "sometimes"`)
	cfg, problems := LoadConfig(dir)
	if len(problems) != 1 || !strings.Contains(problems[0].Error(), "sometimes") {
		t.Fatalf("problems = %v", problems)
	}
	if cfg.Window != WindowAlways {
		t.Fatalf("window = %q, want default", cfg.Window)
	}
}

func TestLoadConfigUnknownKey(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "colour = \"blue\"\nwindow = \"always\"\n")
	_, problems := LoadConfig(dir)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
}

func TestStateRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bopen")
	want := State{LastUsed: "app.zen_browser.zen.desktop", PreviousDefault: "brave-browser.desktop"}
	if err := SaveState(dir, want); err != nil {
		t.Fatal(err)
	}
	if got := LoadState(dir); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != stateFile {
		t.Fatalf("unexpected files left behind: %v", entries)
	}
}

func TestStateCorruptIsEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, stateFile), []byte("last_used = "), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := LoadState(dir); got != (State{}) {
		t.Fatalf("got %+v, want empty", got)
	}
	if err := SaveState(dir, State{LastUsed: "firefox.desktop"}); err != nil {
		t.Fatal(err)
	}
	if got := LoadState(dir); got.LastUsed != "firefox.desktop" {
		t.Fatalf("corrupt state not overwritten: %+v", got)
	}
}

func TestSaveStateLeavesConfigUntouched(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "# my comment\nwindow = \"always\"\n")
	if err := SaveState(dir, State{LastUsed: "x.desktop"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, configFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# my comment\nwindow = \"always\"\n" {
		t.Fatalf("config.toml modified: %q", data)
	}
}
