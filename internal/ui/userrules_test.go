package ui

import (
	"net/url"
	"strings"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/key"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

func ruleWindow(t *testing.T, raw string) (*window, *input.Router) {
	m := envModel(t, raw)
	m.Env.ConfigDir = t.TempDir()
	m.Open = func(discovery.Browser, string) error { return nil }
	w := newWindow(m, m.Env)
	router := new(input.Router)
	frame(w, router)
	return w, router
}

func paramIndex(t *testing.T, a *clean.Analysis, name string) int {
	for i, p := range a.Params {
		if p.Name == name {
			return i
		}
	}
	t.Fatalf("no param %q in %+v", name, a.Params)
	return -1
}

func TestFlagParameterCreatesRule(t *testing.T) {
	w, router := ruleWindow(t, "https://news.example.com/a?ref=home&fbclid=x")
	w.toggles[0].Value, w.m.Accepted[0] = false, false // keep fbclid
	w.flagParam(paramIndex(t, w.m.Analysis, "ref"))
	f := &w.form
	if !f.open || f.param.Text() != "ref" || f.scope.Value != scopeHost || f.thisHost != "news.example.com" {
		t.Fatalf("form %+v", f)
	}
	f.reason.SetText("Referrer tracking")
	r, err := f.rule()
	if err != nil {
		t.Fatal(err)
	}
	w.saveRule(r)
	frame(w, router)
	if f.open {
		t.Fatal("form still open")
	}
	var got *clean.Suggestion
	for i, s := range w.m.Analysis.Suggestions {
		if s.Text == "ref=home" {
			got = &w.m.Analysis.Suggestions[i]
		}
	}
	if got == nil || got.Source != clean.SourceUser || got.Reason != "Referrer tracking" {
		t.Fatalf("suggestions %+v", w.m.Analysis.Suggestions)
	}
	for i, s := range w.m.Analysis.Suggestions {
		if s.Text == "fbclid=x" && w.m.Accepted[i] {
			t.Fatal("fbclid toggle not carried over")
		}
	}
	cfg, _ := prefs.LoadConfig(w.env.ConfigDir)
	if len(cfg.Rules.User) != 1 || cfg.Rules.User[0].Hosts[0] != "news.example.com" {
		t.Fatalf("stored %+v", cfg.Rules.User)
	}
}

func TestFlagParameterInsideRedirectUsesTargetHost(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/item?ref=home")
	w, _ := ruleWindow(t, raw)
	w.flagParam(paramIndex(t, w.m.Analysis, "ref"))
	if w.form.thisHost != "shop.example.com" {
		t.Fatalf("host %q", w.form.thisHost)
	}
}

func TestReasonIsMandatory(t *testing.T) {
	w, _ := ruleWindow(t, "https://news.example.com/a?ref=home")
	w.flagParam(paramIndex(t, w.m.Analysis, "ref"))
	if _, err := w.form.rule(); err == nil || !strings.Contains(err.Error(), "reason is required") {
		t.Fatalf("err %v", err)
	}
}

func TestFormEscapeClosesOnlyTheForm(t *testing.T) {
	w, router := ruleWindow(t, "https://news.example.com/a?ref=home")
	w.flagParam(0)
	frame(w, router)
	press(w, router, key.NameEscape)
	if w.form.open || w.done {
		t.Fatalf("form open %v, window done %v", w.form.open, w.done)
	}
}

func TestDisableBuiltinFromInspector(t *testing.T) {
	w, _ := ruleWindow(t, "https://example.com/?utm_source=x&utm_medium=y&fbclid=z")
	w.disableRule("utm")
	if len(w.m.Analysis.Suggestions) != 1 || w.m.Analysis.Suggestions[0].RuleID != "fbclid" {
		t.Fatalf("suggestions %+v", w.m.Analysis.Suggestions)
	}
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); !cfg.IsDisabled("utm") {
		t.Fatal("not persisted")
	}
}

func TestEditUserRuleFromInspector(t *testing.T) {
	w, _ := ruleWindow(t, "https://example.com/?ref=x")
	_ = w.env.Update(func(c *prefs.Config) {
		c.AddUserRule(prefs.UserRule{Kind: "tracking", Param: "ref", Reason: "old"})
	})
	w.afterRuleChange()
	s := w.m.Analysis.Suggestions[0]
	w.editRule(s.RuleID)
	if !w.form.open || w.form.editID != s.RuleID || w.form.reason.Text() != "old" || !w.form.hostsMode {
		t.Fatalf("form %+v", w.form)
	}
	w.form.reason.SetText("new reason")
	r, err := w.form.rule()
	if err != nil {
		t.Fatal(err)
	}
	w.saveRule(r)
	if got := w.m.Analysis.Suggestions[0].Reason; got != "new reason" {
		t.Fatalf("reason %q", got)
	}
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); len(cfg.Rules.User) != 1 {
		t.Fatalf("edit duplicated the rule: %+v", cfg.Rules.User)
	}
}

func TestSettingsAddEditDeleteUserRule(t *testing.T) {
	w, _ := ruleWindow(t, "https://example.com/?ref=x")
	w.openSettings()
	w.form.add()
	w.form.param.SetText("ref")
	w.form.hosts.SetText("*.Example.com, example.com")
	w.form.reason.SetText("Referrer")
	r, err := w.form.rule()
	if err != nil || len(r.Hosts) != 2 || r.Hosts[0] != "*.example.com" {
		t.Fatalf("rule %+v %v", r, err)
	}
	w.saveRule(r)
	cfg, _ := prefs.LoadConfig(w.env.ConfigDir)
	if len(cfg.Rules.User) != 1 {
		t.Fatalf("stored %+v", cfg.Rules.User)
	}
	id := cfg.Rules.User[0].ID
	w.save(func(c *prefs.Config) { c.DeleteUserRule(id) })
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); len(cfg.Rules.User) != 0 {
		t.Fatalf("not deleted: %+v", cfg.Rules.User)
	}
	w.closeSettings()
	if len(w.m.Analysis.Suggestions) != 0 {
		t.Fatalf("deleted rule still applied: %+v", w.m.Analysis.Suggestions)
	}
}
