package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/appearance"
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/ui/fonts"
)

// Label sizes and letter-spacing for the uppercase JetBrains Mono labels
// of the Blackfyre look (the site uses .06–.08em).
const (
	sizeLabel  unit.Sp = 11
	trackingEm         = 0.08
)

var (
	monoFont = font.Font{Typeface: fonts.Mono}
	textFont = font.Font{Typeface: fonts.Text, Weight: font.SemiBold}
)

// tracked lays out an uppercase JetBrains Mono label with wide
// letter-spacing, one glyph at a time (Gio's text layout has no tracking).
func tracked(th *material.Theme, size unit.Sp, c color.NRGBA, s string) layout.Widget {
	s = strings.ToUpper(s)
	return func(gtx layout.Context) layout.Dimensions {
		gap := int(float32(gtx.Sp(size)) * trackingEm)
		x, h, base := 0, 0, 0
		gtx.Constraints.Min = image.Point{}
		for i, r := range s {
			if i > 0 {
				x += gap
			}
			trans := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
			l := material.Label(th, size, string(r))
			l.Font, l.Color, l.MaxLines = monoFont, c, 1
			d := l.Layout(gtx)
			trans.Pop()
			x += d.Size.X
			h = max(h, d.Size.Y)
			base = max(base, d.Baseline)
		}
		return layout.Dimensions{Size: image.Pt(x, h), Baseline: base}
	}
}

// setAppearance re-themes the window for the host appearance.
func (w *window) setAppearance(s appearance.Settings) {
	w.look = s
	w.pal = NewPalette(s)
	w.pal.apply(w.th)
	setTitleBarDark(w.hwnd, s.Dark)
}

// applyLook re-themes the window when the appearance changed.
func (w *window) applyLook(s appearance.Settings) {
	if s == w.look {
		return
	}
	w.setAppearance(s)
	if w.win != nil {
		w.win.Invalidate()
	}
}

// receiveLooks applies appearance changes delivered by the watcher; it
// runs on the UI goroutine.
func (w *window) receiveLooks() {
	for {
		select {
		case s := <-w.looks:
			w.applyLook(s)
		default:
			return
		}
	}
}

func (p Palette) kind(k clean.Kind) color.NRGBA {
	switch k {
	case clean.KindTracking:
		return p.Tracking
	case clean.KindAffiliate:
		return p.Affiliate
	case clean.KindRedirect:
		return p.Redirect
	}
	return p.Fg
}

type bannerKind int

const (
	bannerWarning bannerKind = iota
	bannerError
)

// rounded fills a rounded rectangle of size with fill and an optional
// 1 dp border.
func rounded(gtx layout.Context, size image.Point, radius unit.Dp, fill, border color.NRGBA) {
	r := gtx.Dp(radius)
	rr := clip.UniformRRect(image.Rectangle{Max: size}, r)
	paint.FillShape(gtx.Ops, fill, rr.Op(gtx.Ops))
	if border.A != 0 {
		paint.FillShape(gtx.Ops, border, clip.Stroke{Path: rr.Path(gtx.Ops), Width: float32(gtx.Dp(1))}.Op())
	}
}

// backed lays out content with a rounded background behind it, stretched
// to the full width when wide is set.
func backed(gtx layout.Context, wide bool, radius unit.Dp, inset layout.Inset, fill, border color.NRGBA, content layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	inner := gtx
	if wide {
		inner.Constraints.Min.X = gtx.Constraints.Max.X
	}
	dims := inset.Layout(inner, content)
	call := macro.Stop()
	rounded(gtx, dims.Size, radius, fill, border)
	call.Add(gtx.Ops)
	return dims
}

// card groups content on a rounded surface with a small-caps title.
func (w *window) card(title string, rows ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return backed(gtx, true, 8, layout.UniformInset(unit.Dp(14)), w.pal.Surface, w.pal.Border,
			func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, 0, len(rows)+1)
				if title != "" {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, w.sectionTitle(title))
					}))
				}
				for i, row := range rows {
					top := unit.Dp(0)
					if i > 0 {
						top = unit.Dp(8)
					}
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: top}.Layout(gtx, row)
					}))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
	}
}

// sectionTitle is a card's uppercase label.
func (w *window) sectionTitle(s string) layout.Widget {
	return tracked(w.th, sizeLabel, w.pal.Muted, s)
}

func (w *window) title(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Label(w.th, unit.Sp(20), s)
		l.Font, l.Color = textFont, w.pal.Fg
		return l.Layout(gtx)
	}
}

func (w *window) muted(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Body2(w.th, s)
		l.Color = w.pal.Muted
		return l.Layout(gtx)
	}
}

func (w *window) banner(s string, kind bannerKind) layout.Widget {
	bg, fg := w.pal.WarnBg, w.pal.WarnFg
	if kind == bannerError {
		bg, fg = w.pal.ErrBg, w.pal.ErrFg
	}
	return func(gtx layout.Context) layout.Dimensions {
		return backed(gtx, true, 8, layout.UniformInset(unit.Dp(10)), bg, color.NRGBA{},
			func(gtx layout.Context) layout.Dimensions {
				l := material.Body2(w.th, s)
				l.Color = fg
				return l.Layout(gtx)
			})
	}
}

// chip is a small rounded label, such as a suggestion kind, recessed into
// the card in the page colour.
func (w *window) chip(s string, fg color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{Top: unit.Dp(3), Bottom: unit.Dp(3), Left: unit.Dp(8), Right: unit.Dp(8)}
		return backed(gtx, false, 6, inset, w.pal.Bg, color.NRGBA{}, tracked(w.th, 10, fg, s))
	}
}

// buttonKind selects a button's colours.
type buttonKind int

const (
	kindNeutral buttonKind = iota // secondary actions
	kindPrimary                   // the accent
	kindQuiet                     // small buttons in rows
)

// Tactile button geometry: 7 dp radius, flat fill, and a 2 dp darker
// bottom shadow that flattens on press.
const (
	buttonRadius unit.Dp = 7
	buttonShadow unit.Dp = 2
)

// tactileButton is the Blackfyre button: a flat face on a darker 2 dp
// shadow; while pressed the face moves down onto it. A disabled button has
// no shadow and muted colours, and takes no clicks.
type tactileButton struct {
	click      *widget.Clickable
	kind       buttonKind
	label      string
	size       unit.Sp
	enabled    bool
	padX, padY unit.Dp
	pal        *Palette
	th         *material.Theme
}

func (b tactileButton) Layout(gtx layout.Context) layout.Dimensions {
	face, fg, shadow := b.pal.buttonColours(b.kind)
	if !b.enabled {
		face, fg, shadow = b.pal.SurfaceAlt, b.pal.Muted, color.NRGBA{}
		gtx = gtx.Disabled()
	}
	drop := gtx.Dp(buttonShadow)
	return b.click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pressed := b.enabled && b.click.Pressed()
		m := op.Record(gtx.Ops)
		inner := gtx
		inner.Constraints.Min = image.Point{}
		dims := layout.Inset{Top: b.padY, Bottom: b.padY, Left: b.padX, Right: b.padX}.Layout(inner, tracked(b.th, b.size, fg, b.label))
		call := m.Stop()
		r := image.Rectangle{Max: dims.Size}
		radius := gtx.Dp(buttonRadius)
		offset := 0
		if pressed {
			offset = drop
		} else if shadow.A != 0 {
			paint.FillShape(gtx.Ops, shadow, clip.UniformRRect(r.Add(image.Pt(0, drop)), radius).Op(gtx.Ops))
		}
		t := op.Offset(image.Pt(0, offset)).Push(gtx.Ops)
		paint.FillShape(gtx.Ops, face, clip.UniformRRect(r, radius).Op(gtx.Ops))
		call.Add(gtx.Ops)
		t.Pop()
		return layout.Dimensions{Size: image.Pt(dims.Size.X, dims.Size.Y+drop), Baseline: dims.Baseline}
	})
}

func (w *window) primaryButton(btn *widget.Clickable, label string) layout.Widget {
	return tactileButton{click: btn, kind: kindPrimary, label: label, size: 12, enabled: true, padX: 20, padY: 10, pal: &w.pal, th: w.th}.Layout
}

func (w *window) secondaryButton(btn *widget.Clickable, label string) layout.Widget {
	return tactileButton{click: btn, kind: kindNeutral, label: label, size: 12, enabled: true, padX: 18, padY: 10, pal: &w.pal, th: w.th}.Layout
}

func (w *window) smallButton(btn *widget.Clickable, label string, disabled bool) layout.Widget {
	return tactileButton{click: btn, kind: kindQuiet, label: label, size: sizeLabel, enabled: !disabled, padX: 10, padY: 6, pal: &w.pal, th: w.th}.Layout
}

func border(gtx layout.Context, c color.NRGBA, radius unit.Dp, content layout.Widget) layout.Dimensions {
	return widget.Border{Color: c, CornerRadius: radius, Width: unit.Dp(1)}.Layout(gtx, content)
}

func (w *window) iconButton(btn *widget.Clickable, icon *widget.Icon, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := material.IconButton(w.th, btn, icon, description)
		b.Size = unit.Dp(18)
		b.Inset = layout.UniformInset(unit.Dp(7))
		b.Background, b.Color = w.pal.QuietBg, w.pal.QuietFg
		return b.Layout(gtx)
	}
}

// actionBar is the bottom bar: a hairline, a hint and the buttons.
func (w *window) actionBar(hint string, buttons ...layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
				paint.FillShape(gtx.Ops, w.pal.Border, clip.Rect{Max: size}.Op())
				return layout.Dimensions{Size: size}
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return backed(gtx, true, 0, layout.Inset{Top: unit.Dp(12), Bottom: unit.Dp(12), Left: unit.Dp(16), Right: unit.Dp(16)},
					w.pal.Surface, color.NRGBA{}, func(gtx layout.Context) layout.Dimensions {
						children := []layout.FlexChild{layout.Flexed(1, w.muted(hint))}
						for _, b := range buttons {
							children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout), layout.Rigid(b))
						}
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
					})
			}),
		)
	}
}

// page lays out a header row and scrolling sections above an action bar.
func (w *window) page(gtx layout.Context, list *widget.List, header layout.Widget, sections []layout.Widget, bar layout.Widget) layout.Dimensions {
	paint.Fill(gtx.Ops, w.pal.Bg)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(12), Bottom: unit.Dp(4), Left: unit.Dp(16), Right: unit.Dp(16)}.Layout(gtx, header)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.List(w.th, list).Layout(gtx, len(sections), func(gtx layout.Context, i int) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(16), Right: unit.Dp(16), Top: unit.Dp(6), Bottom: unit.Dp(6)}.Layout(gtx, sections[i])
			})
		}),
		layout.Rigid(bar),
	)
}

// headerRow is a title with an icon button at the right edge.
func (w *window) headerRow(title string, btn *widget.Clickable, icon *widget.Icon, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, w.title(title)),
			layout.Rigid(w.iconButton(btn, icon, description)),
		)
	}
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

// mono is primary text in the monospace font, for patterns and URLs.
func (w *window) mono(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Body1(w.th, s)
		l.Font = monoFont
		l.Color = w.pal.Fg
		return l.Layout(gtx)
	}
}

func errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
