package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/x/component"
	"gioui.org/x/richtext"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/prefs"
)

// inspectorMenus holds the right-click menus of the inspector.
type inspectorMenus struct {
	// link is the context area over the rendered link. Rich-text spans
	// only report hover and primary clicks, so the parameter under the
	// pointer when the menu opens is the one it acts on.
	link       component.ContextArea
	linkMenu   component.MenuState
	hoverParam int
	menuParam  int
	flag       widget.Clickable
	disable    widget.Clickable
	edit       widget.Clickable
	rows       []suggestionMenu
}

type suggestionMenu struct {
	area   component.ContextArea
	menu   component.MenuState
	action widget.Clickable
}

func (mm *inspectorMenus) reset(suggestions int) {
	mm.hoverParam, mm.menuParam = -1, -1
	mm.link.Dismiss()
	mm.rows = make([]suggestionMenu, suggestions)
}

// menuParam returns the parameter the open link menu acts on.
func (w *window) menuParam() (clean.Param, bool) {
	i := w.menus.menuParam
	if w.m == nil || w.m.Analysis == nil || i < 0 || i >= len(w.m.Analysis.Params) {
		return clean.Param{}, false
	}
	return w.m.Analysis.Params[i], true
}

// paramSuggestion returns the suggestion covering parameter p.
func (w *window) paramSuggestion(p clean.Param) (clean.Suggestion, bool) {
	if p.Suggestion < 0 {
		return clean.Suggestion{}, false
	}
	return w.m.Analysis.Suggestions[p.Suggestion], true
}

func (w *window) handleMenus(gtx layout.Context) {
	mm := &w.menus
	for {
		span, ev, ok := w.urlText.Update(gtx)
		if !ok {
			break
		}
		idx, isParam := span.Get("param").(int)
		if !isParam {
			continue
		}
		switch ev.Type {
		case richtext.Hover:
			mm.hoverParam = idx
		case richtext.Unhover:
			if mm.hoverParam == idx {
				mm.hoverParam = -1
			}
		}
	}
	mm.link.Update(gtx)
	if mm.link.Activated() {
		mm.menuParam = mm.hoverParam
		if mm.menuParam < 0 {
			mm.link.Dismiss()
		}
	}
	if p, ok := w.menuParam(); ok {
		s, hasSuggestion := w.paramSuggestion(p)
		switch {
		case mm.flag.Clicked(gtx):
			mm.link.Dismiss()
			w.flagParam(mm.menuParam)
		case hasSuggestion && mm.disable.Clicked(gtx):
			mm.link.Dismiss()
			w.disableRule(s.RuleID)
			return
		case hasSuggestion && mm.edit.Clicked(gtx):
			mm.link.Dismiss()
			w.editRule(s.RuleID)
		}
	}
	for i := range mm.rows {
		row := &mm.rows[i]
		row.area.Update(gtx)
		s := w.m.Analysis.Suggestions[i]
		if row.area.Activated() && s.Source != clean.SourceBuiltin && s.Source != clean.SourceUser {
			row.area.Dismiss()
		}
		if row.action.Clicked(gtx) {
			row.area.Dismiss()
			if s.Source == clean.SourceBuiltin {
				w.disableRule(s.RuleID)
				return
			}
			w.editRule(s.RuleID)
		}
	}
}

// flagParam opens the rule form for parameter i of the analysis, scoped to
// the host of the URL that contains it.
func (w *window) flagParam(i int) {
	p := w.m.Analysis.Params[i]
	w.form.flag(p.Name, p.Host)
}

// disableRule records a built-in rule as disabled and re-analyses the link.
func (w *window) disableRule(id string) {
	if err := w.env.Update(func(c *prefs.Config) { c.DisableRule(id) }); err != nil {
		w.notice = "Could not disable the rule: " + err.Error()
		return
	}
	w.afterRuleChange()
}

// editRule opens the form for the user rule id.
func (w *window) editRule(id string) {
	if r, ok := w.env.Config.UserRule(id); ok {
		w.form.edit(r)
	}
}

func (w *window) menu(state *component.MenuState, items ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		state.Options = state.Options[:0]
		for _, it := range items {
			state.Options = append(state.Options, it)
		}
		gtx.Constraints.Min = image.Point{}
		return component.Menu(w.th, state).Layout(gtx)
	}
}

// linkMenu lists the actions for the parameter under the pointer.
func (w *window) linkMenu(gtx layout.Context) layout.Dimensions {
	mm := &w.menus
	p, ok := w.menuParam()
	if !ok {
		return layout.Dimensions{}
	}
	items := []layout.Widget{component.MenuItem(w.th, &mm.flag, "Always flag “"+p.Name+"”…").Layout}
	if s, ok := w.paramSuggestion(p); ok {
		switch s.Source {
		case clean.SourceBuiltin:
			items = append(items, component.MenuItem(w.th, &mm.disable, "Disable this rule").Layout)
		case clean.SourceUser:
			items = append(items, component.MenuItem(w.th, &mm.edit, "Edit this rule…").Layout)
		}
	}
	return w.menu(&mm.linkMenu, items...)(gtx)
}

// withRowMenu adds the right-click menu of suggestion i around content.
func (w *window) withRowMenu(i int, content layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		row := &w.menus.rows[i]
		s := w.m.Analysis.Suggestions[i]
		label := "Edit this rule…"
		if s.Source == clean.SourceBuiltin {
			label = "Disable this rule"
		}
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(content),
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return row.area.Layout(gtx, w.menu(&row.menu, component.MenuItem(w.th, &row.action, label).Layout))
			}),
		)
	}
}
