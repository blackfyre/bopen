package ui

import (
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
)

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
		return backed(gtx, true, 10, layout.UniformInset(unit.Dp(14)), w.pal.Surface, w.pal.Border,
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

func (w *window) sectionTitle(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Caption(w.th, strings.ToUpper(s))
		l.Color = w.pal.Muted
		l.Font.Weight = font.SemiBold
		return l.Layout(gtx)
	}
}

func (w *window) title(s string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.H6(w.th, s)
		l.Color = w.pal.Fg
		l.Font.Weight = font.SemiBold
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

// chip is a small rounded label, such as a suggestion kind.
func (w *window) chip(s string, fg color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{Top: unit.Dp(2), Bottom: unit.Dp(2), Left: unit.Dp(8), Right: unit.Dp(8)}
		return backed(gtx, false, 10, inset, w.pal.SurfaceAlt, color.NRGBA{}, func(gtx layout.Context) layout.Dimensions {
			l := material.Caption(w.th, s)
			l.Color = fg
			l.Font.Weight = font.SemiBold
			return l.Layout(gtx)
		})
	}
}

func (w *window) primaryButton(btn *widget.Clickable, label string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := material.Button(w.th, btn, label)
		b.Background, b.Color = w.pal.Accent, w.pal.OnAccent
		b.CornerRadius = unit.Dp(8)
		b.Inset = layout.Inset{Top: unit.Dp(9), Bottom: unit.Dp(9), Left: unit.Dp(20), Right: unit.Dp(20)}
		return b.Layout(gtx)
	}
}

func (w *window) secondaryButton(btn *widget.Clickable, label string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := material.Button(w.th, btn, label)
		b.Background, b.Color = w.pal.SurfaceAlt, w.pal.Fg
		b.CornerRadius = unit.Dp(8)
		b.Inset = layout.Inset{Top: unit.Dp(9), Bottom: unit.Dp(9), Left: unit.Dp(18), Right: unit.Dp(18)}
		return border(gtx, w.pal.Border, 8, b.Layout)
	}
}

func (w *window) smallButton(btn *widget.Clickable, label string, disabled bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		if disabled {
			gtx = gtx.Disabled()
		}
		b := material.Button(w.th, btn, label)
		b.TextSize = unit.Sp(12)
		b.Background, b.Color = w.pal.SurfaceAlt, w.pal.Fg
		b.CornerRadius = unit.Dp(6)
		b.Inset = layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4), Left: unit.Dp(10), Right: unit.Dp(10)}
		return border(gtx, w.pal.Border, 6, b.Layout)
	}
}

func border(gtx layout.Context, c color.NRGBA, radius unit.Dp, content layout.Widget) layout.Dimensions {
	return widget.Border{Color: c, CornerRadius: radius, Width: unit.Dp(1)}.Layout(gtx, content)
}

func (w *window) iconButton(btn *widget.Clickable, icon *widget.Icon, description string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		b := material.IconButton(w.th, btn, icon, description)
		b.Size = unit.Dp(18)
		b.Inset = layout.UniformInset(unit.Dp(7))
		b.Background, b.Color = w.pal.SurfaceAlt, w.pal.Fg
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
