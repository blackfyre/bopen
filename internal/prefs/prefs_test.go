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

func TestLoadConfigBrowsersAndRules(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "[browsers]\nhidden = [\"firefox.desktop\"]\norder = [\"zen.desktop\", \"gone.desktop\"]\n\n[rules]\ndisabled = [\"fbclid\", \"no-such-rule\"]\n")
	cfg, problems := LoadConfig(dir)
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if !cfg.IsHidden("firefox.desktop") || cfg.IsHidden("zen.desktop") {
		t.Fatalf("hidden = %v", cfg.Browsers.Hidden)
	}
	if len(cfg.Browsers.Order) != 2 || cfg.Browsers.Order[1] != "gone.desktop" {
		t.Fatalf("order = %v", cfg.Browsers.Order)
	}
	if !cfg.IsDisabled("fbclid") || !cfg.IsDisabled("no-such-rule") {
		t.Fatalf("disabled = %v", cfg.Rules.Disabled)
	}
}

func TestUpdateConfigPreservesConcurrentChanges(t *testing.T) {
	dir := t.TempDir()
	// Each save is one instance changing one setting; neither may lose the other's.
	if _, err := UpdateConfig(dir, func(c *Config) { c.Browsers.Hidden = append(c.Browsers.Hidden, "firefox.desktop") }); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateConfig(dir, func(c *Config) { c.Window = WindowWhenSuggestions }); err != nil {
		t.Fatal(err)
	}
	got, problems := LoadConfig(dir)
	if len(problems) != 0 || got.Window != WindowWhenSuggestions || !got.IsHidden("firefox.desktop") {
		t.Fatalf("got %+v, problems %v", got, problems)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestUpdateConfigKeepsUnknownIdentities(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "[browsers]\nhidden = [\"uninstalled.desktop\"]\n")
	if _, err := UpdateConfig(dir, func(c *Config) { c.Window = WindowWhenSuggestions }); err != nil {
		t.Fatal(err)
	}
	if got, _ := LoadConfig(dir); !got.IsHidden("uninstalled.desktop") {
		t.Fatalf("hidden = %v", got.Browsers.Hidden)
	}
}

func TestUpdateConfigRefusesMalformedFile(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "window = ")
	if _, err := UpdateConfig(dir, func(c *Config) { c.Window = WindowAlways }); err == nil {
		t.Fatal("expected error")
	}
	data, _ := os.ReadFile(filepath.Join(dir, configFile))
	if string(data) != "window = " {
		t.Fatalf("malformed file overwritten: %q", data)
	}
}

func TestUserRulesValidation(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, `
[[rules.user]]
id = "u-good"
kind = "tracking"
param = "ref"
hosts = ["news.example.com"]
reason = "Referrer tracking"

[[rules.user]]
id = "u-noreason"
kind = "tracking"
param = "src"
reason = ""

[[rules.user]]
id = "u-noparam"
kind = "affiliate"
param = ""
reason = "x"

[[rules.user]]
id = "u-redirect"
kind = "redirect"
param = "q"
reason = "x"
`)
	cfg, problems := LoadConfig(dir)
	if len(problems) != 3 {
		t.Fatalf("problems = %v", problems)
	}
	for _, id := range []string{"u-noreason", "u-noparam", "u-redirect"} {
		found := false
		for _, p := range problems {
			found = found || strings.Contains(p.Error(), id)
		}
		if !found {
			t.Errorf("no problem names %s: %v", id, problems)
		}
	}
	valid := cfg.ValidUserRules()
	if len(valid) != 1 || valid[0].ID != "u-good" || valid[0].Hosts[0] != "news.example.com" {
		t.Fatalf("valid = %+v", valid)
	}
}

func TestUserRuleMutationsPersist(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "window = \"when-suggestions\"\n[browsers]\nhidden = [\"firefox.desktop\"]\n")
	var id string
	if _, err := UpdateConfig(dir, func(c *Config) {
		id = c.AddUserRule(UserRule{Kind: "tracking", Param: "ref", Hosts: []string{"news.example.com"}, Reason: "Referrer tracking"})
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id, "u-") || len(id) != 8 {
		t.Fatalf("id = %q", id)
	}
	cfg, _ := LoadConfig(dir)
	if r, ok := cfg.UserRule(id); !ok || r.Param != "ref" || r.Reason != "Referrer tracking" {
		t.Fatalf("added rule = %+v", cfg.Rules.User)
	}
	if _, err := UpdateConfig(dir, func(c *Config) {
		r, _ := c.UserRule(id)
		r.Reason = "Edited"
		c.SetUserRule(r)
		c.DisableRule("utm")
		c.DisableRule("utm")
	}); err != nil {
		t.Fatal(err)
	}
	cfg, _ = LoadConfig(dir)
	if r, _ := cfg.UserRule(id); r.Reason != "Edited" || len(cfg.Rules.Disabled) != 1 || !cfg.IsDisabled("utm") {
		t.Fatalf("after edit: %+v", cfg.Rules)
	}
	if _, err := UpdateConfig(dir, func(c *Config) { c.DeleteUserRule(id) }); err != nil {
		t.Fatal(err)
	}
	cfg, problems := LoadConfig(dir)
	if len(cfg.Rules.User) != 0 || len(problems) != 0 {
		t.Fatalf("after delete: %+v %v", cfg.Rules.User, problems)
	}
	if cfg.Window != WindowWhenSuggestions || !cfg.IsHidden("firefox.desktop") || !cfg.IsDisabled("utm") {
		t.Fatalf("unrelated settings lost: %+v", cfg)
	}
}

func TestAddUserRuleUniqueIDs(t *testing.T) {
	var c Config
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		id := c.AddUserRule(UserRule{Kind: "tracking", Param: "p", Reason: "r"})
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
}

func TestClearURLsDefaultsOff(t *testing.T) {
	cfg, _ := LoadConfig(t.TempDir())
	if cfg.Rules.ClearURLs {
		t.Fatal("ClearURLs enabled by default")
	}
	dir := t.TempDir()
	writeConfig(t, dir, "[rules]\nclearurls = true\n")
	if cfg, _ := LoadConfig(dir); !cfg.Rules.ClearURLs {
		t.Fatal("clearurls = true not read")
	}
}

func TestSiteRulesPersistAndValidate(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "[[sites]]\nid = \"s-bad\"\nhosts = [\"x.example.com\"]\nbrowser = \"\"\n")
	_, problems := LoadConfig(dir)
	if len(problems) != 1 || !strings.Contains(problems[0].Error(), "s-bad") {
		t.Fatalf("problems %v", problems)
	}
	var id string
	if _, err := UpdateConfig(dir, func(c *Config) {
		c.DeleteSiteRule("s-bad")
		id = c.AddSiteRule(SiteRule{Hosts: []string{"*.atlassian.net"}, Browser: "google-chrome.desktop", Direct: true})
	}); err != nil {
		t.Fatal(err)
	}
	cfg, problems := LoadConfig(dir)
	r, ok := cfg.SiteRule(id)
	if len(problems) != 0 || !ok || !strings.HasPrefix(id, "s-") || r.Hosts[0] != "*.atlassian.net" || !r.Direct || len(cfg.Sites) != 1 {
		t.Fatalf("cfg %+v problems %v", cfg.Sites, problems)
	}
	if _, err := UpdateConfig(dir, func(c *Config) {
		r.Direct = false
		c.SetSiteRule(r)
	}); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := LoadConfig(dir); cfg.Sites[0].Direct {
		t.Fatal("edit not saved")
	}
	for _, bad := range []SiteRule{{Browser: "b"}, {Hosts: []string{"["}, Browser: "b"}, {Hosts: []string{" "}, Browser: "b"}} {
		if bad.Validate() == nil {
			t.Errorf("%+v accepted", bad)
		}
	}
}
