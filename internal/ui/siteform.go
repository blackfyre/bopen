package ui

import (
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/blackfyre/bopen/internal/app"
	"github.com/blackfyre/bopen/internal/prefs"
)

// siteForm edits one site rule as an overlay over the settings view.
type siteForm struct {
	open   bool
	editID string
	hosts  widget.Editor
	// browser holds the chosen browser identity.
	browser      widget.Enum
	direct       widget.Bool
	save, cancel widget.Clickable
	err          string
	scrim        int
}

func (f *siteForm) reset() {
	f.open, f.err, f.editID = true, "", ""
	f.hosts.SingleLine, f.hosts.Submit = true, true
	f.hosts.SetText("")
	f.browser.Value = ""
	f.direct.Value = false
}

func (f *siteForm) add() { f.reset() }

func (f *siteForm) edit(r prefs.SiteRule) {
	f.reset()
	f.editID = r.ID
	f.hosts.SetText(strings.Join(r.Hosts, ", "))
	f.browser.Value = r.Browser
	f.direct.Value = r.Direct
}

// rule builds the site rule from the form fields, validated.
func (f *siteForm) rule() (prefs.SiteRule, error) {
	r := prefs.SiteRule{ID: f.editID, Browser: f.browser.Value, Direct: f.direct.Value}
	for _, h := range strings.Split(f.hosts.Text(), ",") {
		if h = strings.ToLower(strings.TrimSpace(h)); h != "" {
			r.Hosts = append(r.Hosts, h)
		}
	}
	if err := r.Validate(); err != nil {
		msg := err.Error()
		return r, errorf("%s.", strings.ToUpper(msg[:1])+msg[1:])
	}
	return r, nil
}

func (w *window) handleSiteForm(gtx layout.Context) {
	f := &w.siteForm
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if ke, ok := ev.(key.Event); ok && ke.State == key.Press {
			f.open = false
			return
		}
	}
	submitted := false
	for {
		ev, ok := f.hosts.Update(gtx)
		if !ok {
			break
		}
		if _, ok := ev.(widget.SubmitEvent); ok {
			submitted = true
		}
	}
	f.browser.Update(gtx)
	f.direct.Update(gtx)
	if f.cancel.Clicked(gtx) {
		f.open = false
		return
	}
	if f.save.Clicked(gtx) || submitted {
		r, err := f.rule()
		if err != nil {
			f.err = err.Error()
			return
		}
		w.saveSiteRule(r)
	}
}

// saveSiteRule stores r (adding it when new) and closes the form.
func (w *window) saveSiteRule(r prefs.SiteRule) {
	err := w.env.Update(func(c *prefs.Config) {
		if r.ID == "" || !c.SetSiteRule(r) {
			c.AddSiteRule(r)
		}
	})
	if err != nil {
		w.siteForm.err = "Could not save the site rule: " + err.Error()
		return
	}
	w.siteForm.open = false
	w.syncSettings()
}

func (w *window) layoutSiteForm(gtx layout.Context) layout.Dimensions {
	f := &w.siteForm
	title := "Add site rule"
	if f.editID != "" {
		title = "Edit site rule"
	}
	return w.modal(gtx, &f.scrim, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, w.title(title))
			}),
			layout.Rigid(w.field("Hosts (comma-separated, * wildcards)", &f.hosts, "e.g. *.atlassian.net, jira.example.com")),
			layout.Rigid(w.label("Open in")),
		}
		for _, b := range app.Ordered(w.env.AllBrowsers, w.env.Config) {
			label := b.Name
			if w.env.Config.IsHidden(b.ID) {
				label += " (hidden: the rule is ignored while hidden)"
			}
			children = append(children, layout.Rigid(w.radio(&f.browser, b.ID, label)))
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(8)}.Layout(gtx,
				w.checkBox(&f.direct, "Open directly, without the inspector (tracking is still removed)", w.pal.Fg).Layout)
		}))
		if f.err != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, w.banner(f.err, bannerError))
			}))
		}
		children = append(children, layout.Rigid(w.formButtons(&f.save, &f.cancel)))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

// siteRulesCard lists the site rules for the settings view.
func (w *window) siteRulesCard() layout.Widget {
	s := &w.settings
	var rows []layout.Widget
	if len(w.env.Config.Sites) == 0 {
		rows = append(rows, w.muted("None yet. Tick “Always open … in this browser” in the inspector, or add one here."))
	}
	names := map[string]string{}
	for _, b := range w.env.AllBrowsers {
		names[b.ID] = b.Name
	}
	for _, r := range w.env.Config.Sites {
		row := s.siteRule(r.ID)
		browser, ok := names[r.Browser]
		if !ok {
			browser = r.Browser + " (not installed)"
		}
		patterns := strings.Join(r.Hosts, ", ")
		details := "→ " + browser
		if err := r.Validate(); err != nil {
			details += " (ignored: " + err.Error() + ")"
		}
		direct := r.Direct
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							children := []layout.FlexChild{layout.Rigid(w.mono(patterns))}
							if direct {
								children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
									layout.Rigid(w.chip("Opens directly", w.pal.Fg)))
							}
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
						}),
						layout.Rigid(w.muted(details)),
					)
				}),
				layout.Rigid(w.smallButton(&row.edit, "Edit", false)),
				layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
				layout.Rigid(w.smallButton(&row.del, "Delete", false)),
			)
		})
	}
	rows = append(rows, func(gtx layout.Context) layout.Dimensions {
		return layout.W.Layout(gtx, w.secondaryButton(&s.addSite, "Add site rule"))
	})
	return w.card("Site rules", rows...)
}

func (w *window) handleSiteRules(gtx layout.Context) bool {
	s := &w.settings
	if s.addSite.Clicked(gtx) {
		w.siteForm.add()
		return true
	}
	for _, r := range w.env.Config.Sites {
		row := s.siteRule(r.ID)
		if row.edit.Clicked(gtx) {
			w.siteForm.edit(r)
			return true
		}
		if row.del.Clicked(gtx) {
			id := r.ID
			w.save(func(c *prefs.Config) { c.DeleteSiteRule(id) })
			return true
		}
	}
	return false
}

func (s *settingsView) siteRule(id string) *userRuleRow {
	if s.siteRules == nil {
		s.siteRules = map[string]*userRuleRow{}
	}
	if s.siteRules[id] == nil {
		s.siteRules[id] = &userRuleRow{}
	}
	return s.siteRules[id]
}
