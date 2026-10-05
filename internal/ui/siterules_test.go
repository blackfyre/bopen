package ui

import (
	"net/url"
	"strings"
	"testing"

	"gioui.org/io/key"

	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

func TestSiteRulePreselects(t *testing.T) {
	m := envModel(t, "https://www.google.com/url?q="+url.QueryEscape("https://acme.atlassian.net/browse/X-1"))
	m.Env.State.LastUsed = "a"
	m.Env.Config.Sites = []prefs.SiteRule{{ID: "s-1", Hosts: []string{"*.atlassian.net"}, Browser: "c"}}
	m.Selected = -1
	m.Refresh()
	if m.Host != "acme.atlassian.net" || m.Site == nil || m.Browsers[m.Selected].ID != "c" {
		t.Fatalf("host %q site %+v selected %d", m.Host, m.Site, m.Selected)
	}
}

func TestRememberSiteOnOpen(t *testing.T) {
	w, _ := ruleWindow(t, "https://news.example.com/a?fbclid=x")
	var opened []string
	w.m.Open = func(b discovery.Browser, u string) error { opened = append(opened, b.ID); return nil }
	if w.m.Site != nil || w.m.Host != "news.example.com" {
		t.Fatalf("site %+v host %q", w.m.Site, w.m.Host)
	}
	w.m.Selected = 1
	w.m.Remember = true
	if !w.openSelected() {
		t.Fatal("not opened")
	}
	cfg, problems := prefs.LoadConfig(w.env.ConfigDir)
	if len(problems) != 0 || len(cfg.Sites) != 1 || cfg.Sites[0].Hosts[0] != "news.example.com" || cfg.Sites[0].Browser != "b" {
		t.Fatalf("sites %+v %v", cfg.Sites, problems)
	}
	// The next link to the same host pre-selects the remembered browser.
	next := envModel(t, "https://news.example.com/other")
	next.Env.Config = cfg
	next.Env.State.LastUsed = "c"
	next.Selected = -1
	next.Refresh()
	if next.Browsers[next.Selected].ID != "b" || next.Site == nil {
		t.Fatalf("selected %d site %+v", next.Selected, next.Site)
	}
}

func TestSiteRowText(t *testing.T) {
	w, _ := ruleWindow(t, "https://news.example.com/a")
	w.env.Config.Sites = []prefs.SiteRule{{ID: "s-1", Hosts: []string{"*.example.com"}, Browser: "b"}}
	w.m.Refresh()
	if w.m.Site == nil || !strings.Contains(strings.Join(w.m.Site.Hosts, ","), "*.example.com") {
		t.Fatalf("site %+v", w.m.Site)
	}
}

func TestSiteFormAddEditDelete(t *testing.T) {
	w, router := ruleWindow(t, "https://example.com/")
	w.openSettings()
	f := &w.siteForm
	f.add()
	if _, err := f.rule(); err == nil || !strings.Contains(err.Error(), "host pattern") {
		t.Fatalf("empty form accepted: %v", err)
	}
	f.hosts.SetText("*.Atlassian.net, jira.example.com")
	if _, err := f.rule(); err == nil || !strings.Contains(err.Error(), "browser") {
		t.Fatalf("browser not required: %v", err)
	}
	f.browser.Value = "c"
	f.direct.Value = true
	r, err := f.rule()
	if err != nil || r.Hosts[0] != "*.atlassian.net" || len(r.Hosts) != 2 {
		t.Fatalf("rule %+v %v", r, err)
	}
	w.saveSiteRule(r)
	frame(w, router)
	cfg, _ := prefs.LoadConfig(w.env.ConfigDir)
	if f.open || len(cfg.Sites) != 1 || !cfg.Sites[0].Direct || cfg.Sites[0].Browser != "c" {
		t.Fatalf("stored %+v", cfg.Sites)
	}
	f.edit(cfg.Sites[0])
	f.direct.Value = false
	r, _ = f.rule()
	w.saveSiteRule(r)
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); len(cfg.Sites) != 1 || cfg.Sites[0].Direct {
		t.Fatalf("edit duplicated or not saved: %+v", cfg.Sites)
	}
	id := cfg.Sites[0].ID
	w.save(func(c *prefs.Config) { c.DeleteSiteRule(id) })
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); len(cfg.Sites) != 0 {
		t.Fatalf("not deleted: %+v", cfg.Sites)
	}
}

func TestSiteFormEscapeClosesOnlyTheForm(t *testing.T) {
	w, router := ruleWindow(t, "https://example.com/")
	w.openSettings()
	w.siteForm.add()
	frame(w, router)
	press(w, router, key.NameEscape)
	if w.siteForm.open || w.view != viewSettings {
		t.Fatalf("form open %v view %v", w.siteForm.open, w.view)
	}
}
