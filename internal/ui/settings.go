package ui

import (
	"fmt"
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
	message string
	isError bool
	status  string
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
	var sections []layout.Widget
	if w.m != nil {
		sections = append(sections, w.headerRow("Settings", &s.back, iconBack, "Back"))
	} else {
		sections = append(sections, w.heading("bopen settings"))
	}
	if s.message != "" {
		bg := colWarning
		if s.isError {
			bg = colError
		}
		sections = append(sections, w.banner(s.message, bg))
	}

	sections = append(sections, w.heading("Inspector window"),
		w.radio(&s.windowMode, string(prefs.WindowAlways), "Always show it"),
		w.radio(&s.windowMode, string(prefs.WindowWhenSuggestions),
			"Only when there are suggestions (other links open directly in the last-used browser)"))

	sections = append(sections, w.heading("Browsers"))
	ordered := app.Ordered(w.env.AllBrowsers, w.env.Config)
	if len(ordered) == 0 {
		sections = append(sections, w.muted("No web browsers were found on this system."))
	}
	for i, b := range ordered {
		row := s.browsers[b.ID]
		label := fmt.Sprintf("%s  (%s)", b.Name, b.Kind)
		first, last := i == 0, i == len(ordered)-1
		sections = append(sections, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, material.CheckBox(w.th, &row.show, label).Layout),
				layout.Rigid(w.smallButton(&row.up, "Up", first)),
				layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout),
				layout.Rigid(w.smallButton(&row.down, "Down", last)),
			)
		})
	}

	sections = append(sections, w.heading("Your rules"))
	if len(w.env.Config.Rules.User) == 0 {
		sections = append(sections, w.muted("None yet. Add one here, or right-click a parameter in the inspector."))
	}
	for _, r := range w.env.Config.Rules.User {
		row := s.userRule(r.ID)
		pattern := r.Param
		if len(r.Hosts) > 0 {
			pattern += "  on " + strings.Join(r.Hosts, ", ")
		}
		details := r.Kind + " · " + r.Reason
		if err := r.Validate(); err != nil {
			details += " (ignored: " + err.Error() + ")"
		}
		colour := kindColour(clean.Kind(r.Kind))
		sections = append(sections, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							l := material.Body1(w.th, pattern)
							l.Font.Typeface = "Go Mono"
							l.Color = colour
							return l.Layout(gtx)
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
	sections = append(sections, func(gtx layout.Context) layout.Dimensions {
		return layout.W.Layout(gtx, material.Button(w.th, &s.addRule, "Add rule").Layout)
	})

	sections = append(sections, w.heading("Built-in rules"))
	for i, r := range w.env.Builtin {
		cb := material.CheckBox(w.th, &s.rules[i], rulePattern(r))
		cb.Color = kindColour(r.Kind)
		cb.Font.Typeface = "Go Mono"
		details := kindLabel(r.Kind) + " · " + r.Reason
		sections = append(sections, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(cb.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(32)}.Layout(gtx, w.muted(details))
				}),
			)
		})
	}

	sections = append(sections, w.heading("Default browser"), w.muted(s.status))
	if w.env.Registrar != nil {
		sections = append(sections, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.W.Layout(gtx, material.Button(w.th, &s.reregister, "Register bopen again").Layout)
			})
		})
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(w.th, &s.list).Layout(gtx, len(sections), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, sections[i])
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				hint := "Changes are saved immediately · Esc goes back"
				label := "Back"
				if w.m == nil {
					hint, label = "Changes are saved immediately · Esc closes", "Close"
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, w.muted(hint)),
					layout.Rigid(material.Button(w.th, &s.close, label).Layout),
				)
			})
		}),
	)
}

func (w *window) radio(e *widget.Enum, key, label string) layout.Widget {
	return material.RadioButton(w.th, e, key, label).Layout
}

func (w *window) smallButton(btn *widget.Clickable, label string, disabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		b := material.Button(w.th, btn, label)
		b.TextSize = unit.Sp(12)
		b.Inset = layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4), Left: unit.Dp(8), Right: unit.Dp(8)}
		b.Background = colMuted
		return b.Layout(gtx)
	}
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
