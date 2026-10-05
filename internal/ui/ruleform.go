package ui

import (
	"errors"
	"strings"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/prefs"
)

const (
	scopeHost = "host"
	scopeAny  = "any"
)

// ruleForm edits one user rule. It is shown as an overlay over the
// inspector or the settings view.
type ruleForm struct {
	open bool
	// editID is the rule being edited, or "" for a new rule.
	editID string
	// hostsMode edits the host list directly instead of offering the
	// this-host / any-host choice.
	hostsMode bool
	thisHost  string

	param, reason, hosts widget.Editor
	scope, kind          widget.Enum
	save, cancel         widget.Clickable
	err                  string
	// scrim is the event tag that keeps clicks from reaching the view
	// underneath the form.
	scrim int
}

func (f *ruleForm) reset() {
	f.open, f.err = true, ""
	for _, e := range []*widget.Editor{&f.param, &f.reason, &f.hosts} {
		e.SingleLine, e.Submit = true, true
		e.SetText("")
	}
	f.kind.Value = "tracking"
	f.scope.Value = scopeAny
}

// flag opens the form for a new rule about parameter name seen on host.
func (f *ruleForm) flag(name, host string) {
	f.reset()
	f.editID, f.hostsMode, f.thisHost = "", false, host
	f.param.SetText(name)
	if host != "" {
		f.scope.Value = scopeHost
	}
}

// add opens the form for a new rule with an editable host list.
func (f *ruleForm) add() {
	f.reset()
	f.editID, f.hostsMode, f.thisHost = "", true, ""
}

// edit opens the form for an existing rule.
func (f *ruleForm) edit(r prefs.UserRule) {
	f.reset()
	f.editID, f.hostsMode, f.thisHost = r.ID, true, ""
	f.param.SetText(r.Param)
	f.reason.SetText(r.Reason)
	f.hosts.SetText(strings.Join(r.Hosts, ", "))
	f.kind.Value = r.Kind
}

// rule builds the rule from the form fields, validated.
func (f *ruleForm) rule() (prefs.UserRule, error) {
	r := prefs.UserRule{
		ID:     f.editID,
		Kind:   f.kind.Value,
		Param:  strings.TrimSpace(f.param.Text()),
		Reason: strings.TrimSpace(f.reason.Text()),
	}
	switch {
	case f.hostsMode:
		for _, h := range strings.Split(f.hosts.Text(), ",") {
			if h = strings.ToLower(strings.TrimSpace(h)); h != "" {
				r.Hosts = append(r.Hosts, h)
			}
		}
	case f.scope.Value == scopeHost && f.thisHost != "":
		r.Hosts = []string{f.thisHost}
	}
	if err := r.Validate(); err != nil {
		return r, errors.New(strings.ToUpper(err.Error()[:1]) + err.Error()[1:] + ".")
	}
	return r, nil
}

// handleForm processes the form's input. It returns the rule to save when
// the user saved a valid rule.
func (w *window) handleForm(gtx layout.Context) (prefs.UserRule, bool) {
	f := &w.form
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if ke, ok := ev.(key.Event); ok && ke.State == key.Press {
			f.open = false
			return prefs.UserRule{}, false
		}
	}
	submitted := false
	for _, e := range []*widget.Editor{&f.param, &f.reason, &f.hosts} {
		for {
			ev, ok := e.Update(gtx)
			if !ok {
				break
			}
			if _, ok := ev.(widget.SubmitEvent); ok {
				submitted = true
			}
		}
	}
	f.scope.Update(gtx)
	f.kind.Update(gtx)
	if f.cancel.Clicked(gtx) {
		f.open = false
		return prefs.UserRule{}, false
	}
	if f.save.Clicked(gtx) || submitted {
		r, err := f.rule()
		if err != nil {
			f.err = err.Error()
			return prefs.UserRule{}, false
		}
		return r, true
	}
	return prefs.UserRule{}, false
}

// saveRule stores r (adding it when new) and closes the form.
func (w *window) saveRule(r prefs.UserRule) {
	err := w.env.Update(func(c *prefs.Config) {
		if r.ID == "" || !c.SetUserRule(r) {
			c.AddUserRule(r)
		}
	})
	if err != nil {
		w.form.err = "Could not save the rule: " + err.Error()
		return
	}
	w.form.open = false
	w.afterRuleChange()
}

// afterRuleChange re-applies the rules to the current view.
func (w *window) afterRuleChange() {
	if w.view == viewSettings {
		w.syncSettings()
		return
	}
	if w.m != nil {
		w.m.Refresh()
		w.syncInspector()
	}
}

// modal draws content as a centred card over a scrim. The scrim's event
// tag keeps pointer input from reaching the view underneath.
func (w *window) modal(gtx layout.Context, scrim event.Tag, content layout.Widget) layout.Dimensions {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, scrim)
	paint.ColorOp{Color: w.pal.Scrim}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	area.Pop()
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X-gtx.Dp(32), gtx.Dp(560))
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return backed(gtx, true, 12, layout.UniformInset(unit.Dp(20)), w.pal.Surface, w.pal.Border, content)
	})
}

// formButtons is the Save/Cancel row of a modal form.
func (w *window) formButtons(save, cancel *widget.Clickable) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, w.muted("Enter saves · Esc cancels")),
				layout.Rigid(w.secondaryButton(cancel, "Cancel")),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(w.primaryButton(save, "Save")),
			)
		})
	}
}

func (w *window) layoutForm(gtx layout.Context) layout.Dimensions {
	f := &w.form
	return w.modal(gtx, &f.scrim, func(gtx layout.Context) layout.Dimensions {
		title := "Always flag this parameter"
		switch {
		case f.editID != "":
			title = "Edit rule"
		case f.hostsMode:
			title = "Add rule"
		}
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, w.title(title))
			}),
			layout.Rigid(w.field("Parameter (a trailing * matches a prefix)", &f.param, "e.g. ref or ref_*")),
		}
		if f.hostsMode {
			children = append(children, layout.Rigid(w.field("Hosts (comma-separated, * wildcards; empty for any host)", &f.hosts, "e.g. *.example.com")))
		} else {
			children = append(children, layout.Rigid(w.label("Applies to")))
			if f.thisHost != "" {
				children = append(children, layout.Rigid(w.radio(&f.scope, scopeHost, "This host only ("+f.thisHost+")")))
			}
			children = append(children, layout.Rigid(w.radio(&f.scope, scopeAny, "Any host")))
		}
		children = append(children,
			layout.Rigid(w.label("Kind")),
			layout.Rigid(w.radio(&f.kind, "tracking", "Tracking (removed by default)")),
			layout.Rigid(w.radio(&f.kind, "affiliate", "Affiliate (shown, kept by default)")),
			layout.Rigid(w.field("Reason (required, shown next to the suggestion)", &f.reason, "Why should this be removed?")),
		)
		if f.err != "" {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, w.banner(f.err, bannerError))
			}))
		}
		children = append(children, layout.Rigid(w.formButtons(&f.save, &f.cancel)))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}

func (w *window) label(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := material.Body2(w.th, s)
			l.Color = w.pal.Fg
			return l.Layout(gtx)
		})
	}
}

func (w *window) field(label string, e *widget.Editor, hint string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(w.label(label)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return border(gtx, w.pal.Border, 6, func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(w.th, e, hint)
						ed.Color, ed.HintColor, ed.SelectionColor = w.pal.Fg, w.pal.Muted, withAlpha(w.pal.Accent, 0x60)
						return layout.UniformInset(unit.Dp(8)).Layout(gtx, ed.Layout)
					})
				})
			}),
		)
	}
}
