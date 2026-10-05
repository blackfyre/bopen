package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
	"gioui.org/x/styledtext"

	"github.com/blackfyre/bopen/internal/appearance"
	"github.com/blackfyre/bopen/internal/clean"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func kindLabel(k clean.Kind) string {
	switch k {
	case clean.KindTracking:
		return "Tracking"
	case clean.KindAffiliate:
		return "Affiliate"
	case clean.KindRedirect:
		return "Redirect"
	}
	return string(k)
}

func sourceLabel(s clean.Source) string {
	switch s {
	case clean.SourceBuiltin:
		return "Built-in"
	case clean.SourceUser:
		return "Your rule"
	case clean.SourceClearURLs:
		return "ClearURLs list"
	}
	return string(s)
}

type view int

const (
	viewInspector view = iota
	viewSettings
)

type window struct {
	// m is nil when bopen runs as `bopen settings`.
	m        *Model
	env      *Env
	view     view
	th       *material.Theme
	list     widget.List
	urlText  richtext.InteractiveText
	toggles  []widget.Bool
	browsers widget.Enum
	open     widget.Clickable
	cancel   widget.Clickable
	cog      widget.Clickable
	copyURL  widget.Clickable
	// copied is the result last copied to the clipboard.
	copied   string
	remember widget.Bool
	private  widget.Bool
	result   widget.Selectable
	settings settingsView
	form     ruleForm
	siteForm siteForm
	// look is the host appearance; pal is derived from it.
	look appearance.Settings
	pal  Palette
	// looks delivers appearance changes to the UI goroutine.
	looks chan appearance.Settings
	// hwnd is the native window on Windows, for the title bar colour.
	hwnd  uintptr
	menus inspectorMenus
	// notice reports a failed inspector action.
	notice string
	done   bool
	// win is set while the window runs, so background work can request
	// a redraw.
	win *app.Window
	// fetches delivers finished ClearURLs downloads to the UI goroutine.
	fetches  chan fetchResult
	fetching bool
}

var (
	iconSettings = mustIcon(icons.ActionSettings)
	iconBack     = mustIcon(icons.NavigationArrowBack)
)

func mustIcon(data []byte) *widget.Icon {
	ic, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return ic
}

// Run shows the inspector for m and exits the process when it closes.
// It never returns.
func Run(m *Model) {
	w := newWindow(m, m.Env)
	w.startBackgroundRefresh(time.Now())
	run(w)
}

// RunSettings shows only the settings view and exits when it closes.
// It never returns.
func RunSettings(env *Env) {
	w := newWindow(nil, env)
	w.openSettings()
	run(w)
}

func run(win *window) {
	go func() {
		w := new(app.Window)
		w.Option(app.Title("bopen"), app.Size(unit.Dp(760), unit.Dp(600)), app.MinSize(unit.Dp(420), unit.Dp(320)))
		if err := win.loop(w); err != nil {
			fmt.Fprintln(os.Stderr, "bopen:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func newWindow(m *Model, env *Env) *window {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	w := &window{m: m, env: env, th: th, fetches: make(chan fetchResult, 4), looks: make(chan appearance.Settings, 4)}
	w.list.Axis = layout.Vertical
	w.settings.list.Axis = layout.Vertical
	var look appearance.Settings
	if env != nil {
		look = env.Appearance
	}
	w.setAppearance(look)
	w.syncInspector()
	return w
}

// syncInspector copies the model's toggles and selection into the widgets.
func (w *window) syncInspector() {
	if w.m == nil {
		return
	}
	w.toggles = make([]widget.Bool, len(w.m.Accepted))
	for i, a := range w.m.Accepted {
		w.toggles[i].Value = a
	}
	w.browsers.Value = strconv.Itoa(w.m.Selected)
	w.remember.Value = w.m.Remember
	w.private.Value = w.m.Private
	if w.m.Analysis != nil {
		w.menus.reset(len(w.m.Analysis.Suggestions))
	}
}

// closeSettings returns to the inspector, re-applying the preferences, or
// closes the window when there is no inspector.
func (w *window) closeSettings() {
	if w.m == nil {
		w.done = true
		return
	}
	w.m.Refresh()
	w.syncInspector()
	w.view = viewInspector
}

func (w *window) loop(win *app.Window) error {
	w.win = win
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	appearance.Watch(ctx, func(s appearance.Settings) {
		w.looks <- s
		win.Invalidate()
	})
	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.ConfigEvent:
			// Windows has no change notification bopen can receive, so the
			// appearance is read again whenever the window regains focus.
			if e.Config.Focused {
				w.applyLook(appearance.Read())
			}
		case app.ViewEvent:
			w.viewEvent(e)
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			w.handle(gtx)
			if w.done {
				win.Perform(system.ActionClose)
			}
			w.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

// handle processes keyboard shortcuts and widget state changes.
func (w *window) handle(gtx layout.Context) {
	w.receiveFetches()
	w.receiveLooks()
	if w.form.open {
		if r, ok := w.handleForm(gtx); ok {
			w.saveRule(r)
		}
		return
	}
	if w.siteForm.open {
		w.handleSiteForm(gtx)
		return
	}
	if w.view == viewSettings {
		w.handleSettings(gtx)
		return
	}
	if w.cog.Clicked(gtx) {
		w.openSettings()
		return
	}
	if w.m.Analysis != nil {
		w.handleMenus(gtx)
		if w.form.open {
			return
		}
	}
	filters := []event.Filter{
		key.Filter{Name: key.NameReturn}, key.Filter{Name: key.NameEnter},
		key.Filter{Name: key.NameEscape},
		key.Filter{Name: key.NameUpArrow}, key.Filter{Name: key.NameDownArrow},
	}
	for d := 1; d <= 9; d++ {
		filters = append(filters, key.Filter{Name: key.Name(strconv.Itoa(d))})
	}
	filters = append(filters, key.Filter{Name: "P", Optional: key.ModShift}, key.Filter{Name: "C", Optional: key.ModShift})
	for {
		ev, ok := gtx.Event(filters...)
		if !ok {
			break
		}
		ke, ok := ev.(key.Event)
		if !ok || ke.State != key.Press {
			continue
		}
		switch ke.Name {
		case key.NameReturn, key.NameEnter:
			w.done = w.openSelected() || w.done
		case key.NameEscape:
			w.done = true
		case key.NameUpArrow:
			w.m.Move(-1)
		case key.NameDownArrow:
			w.m.Move(1)
		case "P":
			w.m.SetPrivate(!w.m.Private)
		case "C":
			w.copyResult(gtx)
		default:
			if d, err := strconv.Atoi(string(ke.Name)); err == nil && d <= len(w.m.Browsers) {
				w.m.Selected = d - 1
				w.m.SetPrivate(w.m.Private)
			}
		}
		w.browsers.Value = strconv.Itoa(w.m.Selected)
		w.private.Value = w.m.Private
	}
	for i := range w.toggles {
		if w.toggles[i].Update(gtx) {
			w.m.Accepted[i] = w.toggles[i].Value
		}
	}
	if w.remember.Update(gtx) {
		w.m.Remember = w.remember.Value
	}
	if w.private.Update(gtx) {
		w.m.SetPrivate(w.private.Value)
	}
	if w.browsers.Update(gtx) {
		if i, err := strconv.Atoi(w.browsers.Value); err == nil {
			w.m.Selected = i
			w.m.SetPrivate(w.m.Private)
		}
	}
	w.private.Value = w.m.Private
	if w.copyURL.Clicked(gtx) {
		w.copyResult(gtx)
	}
	if w.open.Clicked(gtx) {
		w.done = w.openSelected() || w.done
	}
	if w.cancel.Clicked(gtx) {
		w.done = true
	}
}

func (w *window) layout(gtx layout.Context) layout.Dimensions {
	var dims layout.Dimensions
	if w.view == viewSettings {
		dims = w.layoutSettings(gtx)
	} else {
		dims = w.layoutInspector(gtx)
	}
	if w.form.open {
		w.layoutForm(gtx)
	}
	if w.siteForm.open {
		w.layoutSiteForm(gtx)
	}
	return dims
}

func (w *window) layoutInspector(gtx layout.Context) layout.Dimensions {
	var sections []layout.Widget
	for _, p := range w.m.Problems {
		sections = append(sections, w.banner(p, bannerWarning))
	}
	for _, e := range []string{w.m.Blocker, w.m.LaunchError, w.notice} {
		if e != "" {
			sections = append(sections, w.banner(e, bannerError))
		}
	}
	sections = append(sections, w.card("Link", w.link))
	if w.m.Analysis != nil {
		var rows []layout.Widget
		if len(w.m.Analysis.Suggestions) == 0 {
			rows = append(rows, w.muted("Nothing to remove: this link has no known tracking parts."))
		}
		for i := range w.m.Analysis.Suggestions {
			rows = append(rows, w.withRowMenu(i, w.suggestion(i)))
		}
		sections = append(sections, w.card("Suggested changes", rows...), w.card("Result", w.resultRow))
	}
	if w.m.Blocker == "" && len(w.m.Browsers) > 0 {
		var rows []layout.Widget
		for i := range w.m.Browsers {
			rows = append(rows, w.browser(i))
		}
		if w.m.CanOpenPrivate() {
			rows = append(rows, w.privateRow)
		}
		if w.m.Site != nil || w.m.Host != "" {
			rows = append(rows, w.siteRow)
		}
		sections = append(sections, w.card("Open in", rows...))
	}
	return w.page(gtx, &w.list, w.headerRow("bopen", &w.cog, iconSettings, "Settings"), sections, w.buttons)
}

func (w *window) link(gtx layout.Context) layout.Dimensions {
	if w.m.Analysis == nil {
		return w.linkText(gtx)
	}
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(w.linkText),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return w.menus.link.Layout(gtx, w.linkMenu)
		}),
	)
}

func (w *window) linkText(gtx layout.Context) layout.Dimensions {
	var spans []richtext.SpanStyle
	for _, r := range w.m.runs() {
		s := richtext.SpanStyle{Content: r.text, Size: unit.Sp(15), Color: w.pal.Fg, Font: font.Font{Typeface: "Go Mono"}}
		if r.param >= 0 {
			// Parameters are interactive so their hover state tells the
			// right-click menu which parameter it is for.
			s.Interactive = true
			s.Set("param", r.param)
		}
		if r.kind != "" {
			if r.accepted {
				s.Color = w.pal.kind(r.kind)
				s.Font.Weight = font.Bold
			} else {
				s.Color = w.pal.Muted
			}
		}
		spans = append(spans, s)
	}
	t := richtext.Text(&w.urlText, w.th.Shaper, spans...)
	t.WrapPolicy = styledtext.WrapGraphemes
	return t.Layout(gtx)
}

func (w *window) suggestion(i int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		s := w.m.Analysis.Suggestions[i]
		available := w.m.Analysis.Available(i, w.m.Accepted)
		if !available {
			gtx = gtx.Disabled()
		}
		label := s.Text
		if s.Kind == clean.KindRedirect {
			label = "Unwrap " + s.Text
		}
		cb := material.CheckBox(w.th, &w.toggles[i], label)
		cb.Color = w.pal.kind(s.Kind)
		cb.IconColor = w.pal.Accent
		cb.Font.Typeface = "Go Mono"
		details := s.Reason
		if s.Kind == clean.KindRedirect {
			details += "\nTarget: " + s.Target
		}
		if !available {
			details += "\nOnly applies when the redirect above is unwrapped."
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, cb.Layout),
					layout.Rigid(w.chip(kindLabel(s.Kind), w.pal.kind(s.Kind))),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(w.chip(sourceLabel(s.Source), w.pal.Muted)),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(36)}.Layout(gtx, w.muted(details))
			}),
		)
	}
}

// copyResult puts the result URL on the clipboard.
func (w *window) copyResult(gtx layout.Context) {
	if w.m.Analysis == nil {
		return
	}
	r := w.m.Result()
	gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(r))})
	w.copied = r
}

func (w *window) resultRow(gtx layout.Context) layout.Dimensions {
	label := "Copy (C)"
	if w.copied != "" && w.copied == w.m.Result() {
		label = "Copied"
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, w.resultURL),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(w.smallButton(&w.copyURL, label, false)),
	)
}

func (w *window) resultURL(gtx layout.Context) layout.Dimensions {
	l := material.Body1(w.th, w.m.Result())
	l.Font.Typeface = "Go Mono"
	l.Color = w.pal.Fg
	l.State = &w.result
	l.WrapPolicy = text.WrapGraphemes
	return l.Layout(gtx)
}

func (w *window) browser(i int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := w.m.Browsers[i]
		label := b.Name
		if i < 9 {
			label = fmt.Sprintf("%d   %s", i+1, label)
		}
		rb := material.RadioButton(w.th, &w.browsers, strconv.Itoa(i), label)
		rb.IconColor = w.pal.Accent
		rb.Color = w.pal.Fg
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, rb.Layout),
			layout.Rigid(w.chip(string(b.Kind), w.pal.Muted)),
		)
	}
}

// openSelected opens the link and reports a failure to remember the site,
// which cannot be shown once the window closes.
func (w *window) openSelected() bool {
	ok := w.m.OpenSelected()
	if ok && w.m.RememberError != nil {
		fmt.Fprintln(os.Stderr, "bopen: could not save the site rule:", w.m.RememberError)
	}
	return ok
}

// privateRow offers a private window when the selected browser has one.
func (w *window) privateRow(gtx layout.Context) layout.Dimensions {
	if !w.m.CanOpenPrivate() {
		return layout.Dimensions{}
	}
	return w.checkBox(&w.private, "Open in a private window (P)", w.pal.Fg).Layout(gtx)
}

// siteRow names the site rule that chose the browser, or offers to
// remember the choice for the destination host.
func (w *window) siteRow(gtx layout.Context) layout.Dimensions {
	switch {
	case w.m.Site != nil:
		return w.muted("Chosen by your site rule for " + strings.Join(w.m.Site.Hosts, ", ") + " (change it in the settings).")(gtx)
	case w.m.Host != "":
		return w.checkBox(&w.remember, "Always open "+w.m.Host+" in this browser", w.pal.Fg).Layout(gtx)
	}
	return layout.Dimensions{}
}

func (w *window) buttons(gtx layout.Context) layout.Dimensions {
	if !w.m.CanOpen() {
		return w.actionBar("Esc closes", w.secondaryButton(&w.cancel, "Close"))(gtx)
	}
	return w.actionBar("Enter opens · Esc cancels · ↑/↓ 1–9 choose · P private · C copy",
		w.secondaryButton(&w.cancel, "Cancel"), w.primaryButton(&w.open, "Open"))(gtx)
}
