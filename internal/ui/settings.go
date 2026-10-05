package ui

import (
	"image/color"
	"slices"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/app"
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/prefs"
)

type browserRow struct {
	show     widget.Bool
	up, down widget.Clickable
}

type userRuleRow struct {
	edit, del widget.Clickable
}

type settingsView struct {
	list       widget.List
	userRules  map[string]*userRuleRow
	addRule    widget.Clickable
	windowMode widget.Enum
	browsers   map[string]*browserRow
	rules      []widget.Bool
	back       widget.Clickable
	close      widget.Clickable
	reregister widget.Clickable
	// message reports the outcome of the last action; isError marks failures.
	message   string
	isError   bool
	status    string
	clearURLs clearURLsView
}

// openSettings switches to the settings view with widgets reflecting the
// current preferences.
func (w *window) openSettings() {
	w.view = viewSettings
	s := &w.settings
	s.message, s.isError = "", false
	if s.browsers == nil {
		s.browsers = map[string]*browserRow{}
	}
	for _, b := range w.env.AllBrowsers {
		if s.browsers[b.ID] == nil {
			s.browsers[b.ID] = &browserRow{}
		}
	}
	if len(s.rules) != len(w.env.Builtin) {
		s.rules = make([]widget.Bool, len(w.env.Builtin))
	}
	w.syncSettings()
	w.refreshStatus()
}

// syncSettings sets the widgets from the saved preferences, so a failed save
// never leaves a control showing a value that was not stored.
func (w *window) syncSettings() {
	s, cfg := &w.settings, w.env.Config
	s.windowMode.Value = string(cfg.Window)
	for id, row := range s.browsers {
		row.show.Value = !cfg.IsHidden(id)
	}
	for i, r := range w.env.Builtin {
		s.rules[i].Value = !cfg.IsDisabled(r.ID)
	}
	s.clearURLs.enabled.Value = cfg.Rules.ClearURLs
}

func (w *window) refreshStatus() {
	s := &w.settings
	switch {
	case w.env.Registrar == nil:
		s.status = "Registration is not available."
	default:
		isDefault, known := w.env.Registrar.IsDefault()
		switch {
		case !known:
			s.status = "Unknown whether bopen is the default browser."
		case isDefault:
			s.status = "bopen is the default handler for https links."
		default:
			s.status = "bopen is not the default handler for https links."
		}
	}
}

// save applies one preference change and reports a failure in the view.
func (w *window) save(mutate func(*prefs.Config)) {
	s := &w.settings
	if err := w.env.Update(mutate); err != nil {
		s.message, s.isError = "Could not save settings: "+err.Error(), true
	} else if s.isError {
		s.message, s.isError = "", false
	}
	w.syncSettings()
}

func toggle(list []string, id string, present bool) []string {
	list = slices.DeleteFunc(slices.Clone(list), func(v string) bool { return v == id })
	if present {
		list = append(list, id)
	}
	return list
}

func (w *window) handleSettings(gtx layout.Context) {
	s := &w.settings
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if ke, ok := ev.(key.Event); ok && ke.State == key.Press {
			w.closeSettings()
			return
		}
	}
	if s.back.Clicked(gtx) || s.close.Clicked(gtx) {
		w.closeSettings()
		return
	}
	if s.windowMode.Update(gtx) {
		mode := prefs.Window(s.windowMode.Value)
		w.save(func(c *prefs.Config) { c.Window = mode })
	}
	ordered := app.Ordered(w.env.AllBrowsers, w.env.Config)
	for i, b := range ordered {
		row := s.browsers[b.ID]
		if row.show.Update(gtx) {
			id, hide := b.ID, !row.show.Value
			w.save(func(c *prefs.Config) { c.Browsers.Hidden = toggle(c.Browsers.Hidden, id, hide) })
		}
		for _, mv := range []struct {
			btn   *widget.Clickable
			delta int
		}{{&row.up, -1}, {&row.down, 1}} {
			if mv.btn.Clicked(gtx) {
				order := app.MoveOrder(ordered, w.env.Config.Browsers.Order, i, mv.delta)
				w.save(func(c *prefs.Config) { c.Browsers.Order = order })
			}
		}
	}
	w.handleClearURLs(gtx)
	if s.addRule.Clicked(gtx) {
		w.form.add()
		return
	}
	for _, r := range w.env.Config.Rules.User {
		row := s.userRule(r.ID)
		if row.edit.Clicked(gtx) {
			w.form.edit(r)
			return
		}
		if row.del.Clicked(gtx) {
			id := r.ID
			w.save(func(c *prefs.Config) { c.DeleteUserRule(id) })
			return
		}
	}
	for i, r := range w.env.Builtin {
		if s.rules[i].Update(gtx) {
			id, disable := r.ID, !s.rules[i].Value
			w.save(func(c *prefs.Config) { c.Rules.Disabled = toggle(c.Rules.Disabled, id, disable) })
		}
	}
	if s.reregister.Clicked(gtx) && w.env.Registrar != nil {
		msg, err := w.env.Registrar.Register()
		if err != nil {
			s.message, s.isError = "Registration failed: "+err.Error(), true
		} else {
			s.message, s.isError = msg, false
		}
		w.refreshStatus()
	}
}

func (w *window) layoutSettings(gtx layout.Context) layout.Dimensions {
	s := &w.settings
	header := w.title("bopen settings")
	if w.m != nil {
		header = w.headerRow("Settings", &s.back, iconBack, "Back")
	}
	var sections []layout.Widget
	if s.message != "" {
		kind := bannerWarning
		if s.isError {
			kind = bannerError
		}
		sections = append(sections, w.banner(s.message, kind))
	}

	sections = append(sections, w.card("Inspector window",
		w.radio(&s.windowMode, string(prefs.WindowAlways), "Always show it"),
		w.radio(&s.windowMode, string(prefs.WindowWhenSuggestions),
			"Only when there are suggestions (other links open directly in the last-used browser)")))

	var browserRows []layout.Widget
	ordered := app.Ordered(w.env.AllBrowsers, w.env.Config)
	if len(ordered) == 0 {
		browserRows = append(browserRows, w.muted("No web browsers were found on this system."))
	}
	for i, b := range ordered {
		row := s.browsers[b.ID]
		first, last := i == 0, i == len(ordered)-1
		cb := w.checkBox(&row.show, b.Name, w.pal.Fg)
		browserRows = append(browserRows, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, cb.Layout),
				layout.Rigid(w.chip(string(b.Kind), w.pal.Muted)),
				layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
				layout.Rigid(w.smallButton(&row.up, "Up", first)),
				layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
				layout.Rigid(w.smallButton(&row.down, "Down", last)),
			)
		})
	}
	sections = append(sections, w.card("Browsers", browserRows...))

	var ruleRows []layout.Widget
	if len(w.env.Config.Rules.User) == 0 {
		ruleRows = append(ruleRows, w.muted("None yet. Add one here, or right-click a parameter in the inspector."))
	}
	for _, r := range w.env.Config.Rules.User {
		row := s.userRule(r.ID)
		pattern := r.Param
		if len(r.Hosts) > 0 {
			pattern += "  on " + strings.Join(r.Hosts, ", ")
		}
		details := r.Reason
		if err := r.Validate(); err != nil {
			details += " (ignored: " + err.Error() + ")"
		}
		kind := clean.Kind(r.Kind)
		ruleRows = append(ruleRows, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									l := material.Body1(w.th, pattern)
									l.Font.Typeface = "Go Mono"
									l.Color = w.pal.kind(kind)
									return l.Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
								layout.Rigid(w.chip(kindLabel(kind), w.pal.kind(kind))),
							)
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
	ruleRows = append(ruleRows, func(gtx layout.Context) layout.Dimensions {
		return layout.W.Layout(gtx, w.secondaryButton(&s.addRule, "Add rule"))
	})
	sections = append(sections, w.card("Your rules", ruleRows...))

	var builtinRows []layout.Widget
	for i, r := range w.env.Builtin {
		cb := w.checkBox(&s.rules[i], rulePattern(r), w.pal.kind(r.Kind))
		cb.Font.Typeface = "Go Mono"
		details := r.Reason
		kind := r.Kind
		builtinRows = append(builtinRows, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, cb.Layout),
						layout.Rigid(w.chip(kindLabel(kind), w.pal.kind(kind))),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(36)}.Layout(gtx, w.muted(details))
				}),
			)
		})
	}
	sections = append(sections, w.card("Built-in rules", builtinRows...))

	sections = append(sections, w.clearURLsCard())

	defaultRows := []layout.Widget{w.muted(s.status)}
	if w.env.Registrar != nil {
		defaultRows = append(defaultRows, func(gtx layout.Context) layout.Dimensions {
			return layout.W.Layout(gtx, w.secondaryButton(&s.reregister, "Register bopen again"))
		})
	}
	sections = append(sections, w.card("Default browser", defaultRows...))

	hint, label := "Changes are saved immediately · Esc goes back", "Back"
	if w.m == nil {
		hint, label = "Changes are saved immediately · Esc closes", "Close"
	}
	return w.page(gtx, &s.list, header, sections, w.actionBar(hint, w.primaryButton(&s.close, label)))
}

func (w *window) radio(e *widget.Enum, key, label string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		rb := material.RadioButton(w.th, e, key, label)
		rb.IconColor, rb.Color = w.pal.Accent, w.pal.Fg
		return rb.Layout(gtx)
	}
}

func (w *window) checkBox(b *widget.Bool, label string, fg color.NRGBA) material.CheckBoxStyle {
	cb := material.CheckBox(w.th, b, label)
	cb.IconColor, cb.Color = w.pal.Accent, fg
	return cb
}

func (s *settingsView) userRule(id string) *userRuleRow {
	if s.userRules == nil {
		s.userRules = map[string]*userRuleRow{}
	}
	if s.userRules[id] == nil {
		s.userRules[id] = &userRuleRow{}
	}
	return s.userRules[id]
}

// rulePattern describes what a rule matches, for the settings list.
func rulePattern(r clean.Rule) string {
	if r.Kind == clean.KindRedirect {
		parts := make([]string, len(r.Hosts))
		for i, h := range r.Hosts {
			parts[i] = h + r.Path
		}
		return strings.Join(parts, ", ")
	}
	if len(r.Hosts) > 0 {
		return r.Param + "  on " + strings.Join(r.Hosts, ", ")
	}
	return r.Param
}
