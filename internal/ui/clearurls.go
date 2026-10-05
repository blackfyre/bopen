package ui

import (
	"context"
	"fmt"
	"time"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/blackfyre/bopen/internal/clearurls"
	"github.com/blackfyre/bopen/internal/prefs"
)

const clearURLsProject = "https://github.com/ClearURLs/Rules"

type fetchResult struct {
	meta clearurls.Meta
	err  error
	// reload makes the new list apply to the current window; background
	// refreshes only take effect from the next link.
	reload bool
}

type clearURLsView struct {
	enabled widget.Bool
	update  widget.Clickable
	project widget.Selectable
}

// startBackgroundRefresh refreshes a stale cache while the inspector is
// open. It does nothing when ClearURLs is disabled.
func (w *window) startBackgroundRefresh(now time.Time) {
	cu := w.env.ClearURLs
	if cu == nil {
		return
	}
	w.fetching = clearurls.StartRefresh(w.env.Config.Rules.ClearURLs, cu.Meta, cu.Cache, cu.Fetcher, now, cu.Update,
		func(m clearurls.Meta, err error) { w.deliver(fetchResult{meta: m, err: err}) })
}

// startFetch downloads the list now, regardless of the cache age.
func (w *window) startFetch() {
	cu := w.env.ClearURLs
	if cu == nil || w.fetching {
		return
	}
	w.fetching = true
	go func() {
		m, err := cu.Update(context.Background(), cu.Cache, cu.Fetcher, time.Now())
		w.deliver(fetchResult{meta: m, err: err, reload: true})
	}()
}

func (w *window) deliver(r fetchResult) {
	w.fetches <- r
	if w.win != nil {
		w.win.Invalidate()
	}
}

// receiveFetches applies finished downloads; it runs on the UI goroutine.
func (w *window) receiveFetches() {
	for {
		select {
		case r := <-w.fetches:
			w.fetching = false
			cu := w.env.ClearURLs
			cu.Meta = r.meta
			if r.err == nil && r.reload {
				cu.Rules = clearurls.Load(cu.Cache)
			}
		default:
			return
		}
	}
}

func (w *window) handleClearURLs(gtx layout.Context) {
	s, cu := &w.settings, w.env.ClearURLs
	if cu == nil {
		return
	}
	if s.clearURLs.enabled.Update(gtx) {
		w.setClearURLs(s.clearURLs.enabled.Value)
	}
	if s.clearURLs.update.Clicked(gtx) && w.env.Config.Rules.ClearURLs {
		w.startFetch()
	}
}

// setClearURLs saves the ClearURLs preference; enabling it uses any cached
// list at once and starts a download.
func (w *window) setClearURLs(on bool) {
	cu := w.env.ClearURLs
	w.save(func(c *prefs.Config) { c.Rules.ClearURLs = on })
	if w.env.Config.Rules.ClearURLs {
		if cu.Rules == nil {
			cu.Rules = clearurls.Load(cu.Cache)
		}
		w.startFetch()
	}
}

// clearURLsStatus describes the cache state for the settings view.
func (w *window) clearURLsStatus() []string {
	cu := w.env.ClearURLs
	var lines []string
	switch {
	case w.fetching:
		lines = append(lines, "Downloading the rule list…")
	case cu.Meta.LastSuccess.IsZero():
		lines = append(lines, "Not downloaded yet.")
	default:
		lines = append(lines, "Last updated "+cu.Meta.LastSuccess.Local().Format("2 Jan 2006 15:04")+".")
	}
	if cu.Meta.LastError != "" && !w.fetching {
		lines = append(lines, "Last update failed: "+cu.Meta.LastError)
	}
	if cu.Meta.SkippedPatterns > 0 {
		lines = append(lines, fmt.Sprintf("%d patterns could not be used and were skipped.", cu.Meta.SkippedPatterns))
	}
	return lines
}

func (w *window) clearURLsCard() layout.Widget {
	s, cu := &w.settings, w.env.ClearURLs
	if cu == nil {
		return w.card("ClearURLs list", w.muted("Unavailable: no cache directory."))
	}
	rows := []layout.Widget{
		w.checkBox(&s.clearURLs.enabled,
			"Also use the ClearURLs list (downloads rule data from the ClearURLs project; updated daily)", w.pal.Fg).Layout,
	}
	if w.env.Config.Rules.ClearURLs {
		for _, line := range w.clearURLsStatus() {
			rows = append(rows, w.muted(line))
		}
		rows = append(rows, func(gtx layout.Context) layout.Dimensions {
			return layout.W.Layout(gtx, w.smallButton(&s.clearURLs.update, "Update now", w.fetching))
		})
	}
	rows = append(rows, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(w.muted("Rule data by the ClearURLs project, licensed under the LGPL-3.0. Suggestions from it are labelled “ClearURLs list”.")),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				l := material.Body2(w.th, clearURLsProject)
				l.Color = w.pal.Redirect
				l.State = &s.clearURLs.project
				return l.Layout(gtx)
			}),
		)
	})
	return w.card("ClearURLs list", rows...)
}
