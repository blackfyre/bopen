// Package ui implements the inspector window.
package ui

import (
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
)

// Model is the state the inspector presents and edits.
type Model struct {
	// Input is the link as given to bopen.
	Input string
	// Analysis is nil when Input is not a usable web URL.
	Analysis *clean.Analysis
	Accepted []bool
	Browsers []discovery.Browser
	Selected int
	// Problems are shown as warnings above the link (configuration problems).
	Problems []string
	// Blocker, when set, explains why no browser can be opened.
	Blocker string
	// LaunchError is the error of the last failed launch attempt.
	LaunchError string
	// Open launches a browser with the final URL.
	Open func(b discovery.Browser, url string) error
}

// CanOpen reports whether a browser can be launched.
func (m *Model) CanOpen() bool {
	return m.Blocker == "" && m.Analysis != nil && m.Selected >= 0 && m.Selected < len(m.Browsers)
}

// Result is the URL that Open would launch.
func (m *Model) Result() string {
	if m.Analysis == nil {
		return ""
	}
	return m.Analysis.Clean(m.Accepted)
}

// Select moves the browser selection by delta, clamped to the list.
func (m *Model) Move(delta int) {
	if len(m.Browsers) == 0 {
		return
	}
	m.Selected = min(max(m.Selected+delta, 0), len(m.Browsers)-1)
}

// OpenSelected launches the selected browser. It reports whether the window
// should close.
func (m *Model) OpenSelected() bool {
	if !m.CanOpen() {
		return false
	}
	if err := m.Open(m.Browsers[m.Selected], m.Result()); err != nil {
		m.LaunchError = "Could not open " + m.Browsers[m.Selected].Name + ": " + err.Error()
		return false
	}
	return true
}

// styleRun is a run of the analysed URL drawn with one style.
type styleRun struct {
	text     string
	kind     clean.Kind // "" for unaffected text
	accepted bool
}

// runs splits the analysed URL into runs by the suggestion covering each
// byte. Parameter suggestions take precedence over the redirect wrappers
// containing them, and later (inner) suggestions over earlier ones.
func (m *Model) runs() []styleRun {
	if m.Analysis == nil {
		return []styleRun{{text: m.Input}}
	}
	u := m.Analysis.URL
	owner := make([]int, len(u))
	for i := range owner {
		owner[i] = -1
	}
	rank := func(i int) int {
		if i < 0 {
			return -1
		}
		r := i
		if m.Analysis.Suggestions[i].Kind != clean.KindRedirect {
			r += len(m.Analysis.Suggestions)
		}
		return r
	}
	for i, s := range m.Analysis.Suggestions {
		for b := s.Start; b < s.End && b < len(u); b++ {
			if rank(i) > rank(owner[b]) {
				owner[b] = i
			}
		}
	}
	style := func(o int) styleRun {
		if o < 0 {
			return styleRun{}
		}
		return styleRun{
			kind:     m.Analysis.Suggestions[o].Kind,
			accepted: m.Accepted[o] && m.Analysis.Available(o, m.Accepted),
		}
	}
	var out []styleRun
	start := 0
	for b := 1; b <= len(u); b++ {
		if b < len(u) && style(owner[b]) == style(owner[start]) {
			continue
		}
		r := style(owner[start])
		r.text = u[start:b]
		out = append(out, r)
		start = b
	}
	return out
}
