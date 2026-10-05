package ui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
	"time"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
	"gioui.org/x/styledtext"

	"github.com/blackfyre/bopen/internal/clean"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	colTracking  = color.NRGBA{R: 0xC6, G: 0x28, B: 0x28, A: 0xFF}
	colAffiliate = color.NRGBA{R: 0xB2, G: 0x6A, B: 0x00, A: 0xFF}
	colRedirect  = color.NRGBA{R: 0x15, G: 0x65, B: 0xC0, A: 0xFF}
	colMuted     = color.NRGBA{R: 0x6B, G: 0x6B, B: 0x6B, A: 0xFF}
	colWarning   = color.NRGBA{R: 0xFF, G: 0xF4, B: 0xE5, A: 0xFF}
	colError     = color.NRGBA{R: 0xFD, G: 0xEC, B: 0xEA, A: 0xFF}
)

func kindColour(k clean.Kind) color.NRGBA {
	switch k {
	case clean.KindTracking:
		return colTracking
	case clean.KindAffiliate:
		return colAffiliate
	case clean.KindRedirect:
		return colRedirect
	}
	return color.NRGBA{A: 0xFF}
}

func kindLabel(k clean.Kind) string {
	switch k {
	case clean.KindTracking:
		return "tracking"
	case clean.KindAffiliate:
		return "affiliate"
	case clean.KindRedirect:
		return "redirect"
	}
	return string(k)
}

func sourceLabel(s clean.Source) string {
	switch s {
	case clean.SourceBuiltin:
		return "built-in rule"
	case clean.SourceClearURLs:
		return "ClearURLs list"
	}
	return string(s) + " rule"
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
	result   widget.Selectable
	settings settingsView
	form     ruleForm
	menus    inspectorMenus
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
	w := &window{m: m, env: env, th: th, fetches: make(chan fetchResult, 4)}
	w.list.Axis = layout.Vertical
	w.settings.list.Axis = layout.Vertical
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
	var ops op.Ops
	for {
		switch e := win.Event().(type) {
		case app.DestroyEvent:
			return e.Err
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
	if w.form.open {
		if r, ok := w.handleForm(gtx); ok {
			w.saveRule(r)
		}
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
			w.done = w.m.OpenSelected() || w.done
		case key.NameEscape:
			w.done = true
		case key.NameUpArrow:
			w.m.Move(-1)
		case key.NameDownArrow:
			w.m.Move(1)
		default:
			if d, err := strconv.Atoi(string(ke.Name)); err == nil && d <= len(w.m.Browsers) {
				w.m.Selected = d - 1
			}
		}
		w.browsers.Value = strconv.Itoa(w.m.Selected)
	}
	for i := range w.toggles {
		if w.toggles[i].Update(gtx) {
			w.m.Accepted[i] = w.toggles[i].Value
		}
	}
	if w.browsers.Update(gtx) {
		if i, err := strconv.Atoi(w.browsers.Value); err == nil {
			w.m.Selected = i
		}
	}
	if w.open.Clicked(gtx) {
		w.done = w.m.OpenSelected() || w.done
	}
	if w.cancel.Clicked(gtx) {
		w.done = true
	}
}

func (w *window) layout(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, w.th.Palette.Bg)
	var dims layout.Dimensions
	if w.view == viewSettings {
		dims = w.layoutSettings(gtx)
	} else {
		dims = w.layoutInspector(gtx)
	}
	if w.form.open {
		w.layoutForm(gtx)
	}
	return dims
}

func (w *window) layoutInspector(gtx layout.Context) layout.Dimensions {
	var sections []layout.Widget
	for _, p := range w.m.Problems {
		sections = append(sections, w.banner(p, colWarning))
	}
	if w.m.Blocker != "" {
		sections = append(sections, w.banner(w.m.Blocker, colError))
	}
	if w.m.LaunchError != "" {
		sections = append(sections, w.banner(w.m.LaunchError, colError))
	}
	if w.notice != "" {
		sections = append(sections, w.banner(w.notice, colError))
	}
	sections = append(sections, w.headerRow("Link", &w.cog, iconSettings, "Settings"), w.link)
	if w.m.Analysis != nil {
		sections = append(sections, w.heading("Suggested changes"))
		if len(w.m.Analysis.Suggestions) == 0 {
			sections = append(sections, w.muted("Nothing to remove."))
		}
		for i := range w.m.Analysis.Suggestions {
			sections = append(sections, w.withRowMenu(i, w.suggestion(i)))
		}
		sections = append(sections, w.heading("Result"), w.resultURL)
	}
	if w.m.Blocker == "" && len(w.m.Browsers) > 0 {
		sections = append(sections, w.heading("Open in"))
		for i := range w.m.Browsers {
			sections = append(sections, w.browser(i))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(w.th, &w.list).Layout(gtx, len(sections), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(4), Bottom: unit.Dp(4)}.Layout(gtx, sections[i])
			})
		}),
		layout.Rigid(w.buttons),
	)
}

// headerRow is a heading with an icon button at the right edge.
func (w *window) headerRow(title string, btn *widget.Clickable, icon *widget.Icon, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, w.heading(title)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				b := material.IconButton(w.th, btn, icon, description)
				b.Size = unit.Dp(20)
				b.Inset = layout.UniformInset(unit.Dp(6))
				b.Background = colMuted
				return b.Layout(gtx)
			}),
		)
	}
}

func (w *window) heading(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			l := material.Subtitle1(w.th, s)
			l.Font.Weight = font.Bold
			return l.Layout(gtx)
		})
	}
}

func (w *window) muted(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Body2(w.th, s)
		l.Color = colMuted
		return l.Layout(gtx)
	}
}

func (w *window) banner(s string, bg color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		macro := op.Record(gtx.Ops)
		dims := layout.UniformInset(unit.Dp(8)).Layout(gtx, material.Body2(w.th, s).Layout)
		call := macro.Stop()
		rect := clip.UniformRRect(image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)}, gtx.Dp(4))
		paint.FillShape(gtx.Ops, bg, rect.Op(gtx.Ops))
		call.Add(gtx.Ops)
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, dims.Size.Y)}
	}
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
		s := richtext.SpanStyle{Content: r.text, Size: unit.Sp(15), Color: kindColour(r.kind), Font: font.Font{Typeface: "Go Mono"}}
		if r.param >= 0 {
			// Parameters are interactive so their hover state tells the
			// right-click menu which parameter it is for.
			s.Interactive = true
			s.Set("param", r.param)
		}
		if r.kind != "" {
			if r.accepted {
				s.Font.Weight = font.Bold
			} else {
				s.Color = colMuted
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
		cb.Color = kindColour(s.Kind)
		cb.Font.Typeface = "Go Mono"
		details := kindLabel(s.Kind) + " · " + sourceLabel(s.Source) + " · " + s.Reason
		if s.Kind == clean.KindRedirect {
			details += "\nTarget: " + s.Target
		}
		if !available {
			details += "\n(only applies when the redirect above is unwrapped)"
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(cb.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(32)}.Layout(gtx, w.muted(details))
			}),
		)
	}
}

func (w *window) resultURL(gtx layout.Context) layout.Dimensions {
	l := material.Body1(w.th, w.m.Result())
	l.Font.Typeface = "Go Mono"
	l.State = &w.result
	l.WrapPolicy = text.WrapGraphemes
	return l.Layout(gtx)
}

func (w *window) browser(i int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := w.m.Browsers[i]
		label := fmt.Sprintf("%s  (%s)", b.Name, b.Kind)
		if i < 9 {
			label = fmt.Sprintf("%d.  %s", i+1, label)
		}
		return material.RadioButton(w.th, &w.browsers, strconv.Itoa(i), label).Layout(gtx)
	}
}

func (w *window) buttons(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		hint := "Enter opens · Esc cancels · ↑/↓ or 1–9 choose the browser"
		if !w.m.CanOpen() {
			hint = "Esc closes"
		}
		children := []layout.FlexChild{layout.Flexed(1, w.muted(hint))}
		cancel := "Cancel"
		if !w.m.CanOpen() {
			cancel = "Close"
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			b := material.Button(w.th, &w.cancel, cancel)
			b.Background = colMuted
			return b.Layout(gtx)
		}))
		if w.m.CanOpen() {
			children = append(children,
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(material.Button(w.th, &w.open, "Open").Layout))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
	})
}
