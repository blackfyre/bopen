package ui

import (
	"context"
)

type expandResult struct {
	url string
	err error
}

// startExpand resolves the inspected short link in the background.
func (w *window) startExpand() {
	if w.expanding || !w.m.CanExpand() {
		return
	}
	w.expanding = true
	raw, expand := w.m.Analysis.URL, w.env.Expand
	go func() {
		url, err := expand(context.Background(), raw)
		w.expansions <- expandResult{url: url, err: err}
		if w.win != nil {
			w.win.Invalidate()
		}
	}()
}

// receiveExpansion applies a finished expansion on the UI goroutine.
func (w *window) receiveExpansion() {
	select {
	case r := <-w.expansions:
		w.expanding = false
		if r.err == nil {
			r.err = w.m.Replace(r.url)
		}
		if r.err != nil {
			w.notice = "Could not expand the short link: " + r.err.Error()
			return
		}
		w.notice = ""
		w.syncInspector()
	default:
	}
}
