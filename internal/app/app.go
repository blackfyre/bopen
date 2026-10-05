// Package app holds bopen's platform-independent decisions: which browser is
// pre-selected, whether the inspector window is needed, and opening a link.
package app

import (
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

// Preselect returns the index in browsers of the browser to pre-select: the
// last-used browser, else the previous system default, else the first. It
// returns -1 when browsers is empty.
func Preselect(browsers []discovery.Browser, st prefs.State) int {
	for _, id := range []string{st.LastUsed, st.PreviousDefault} {
		if id == "" {
			continue
		}
		for i, b := range browsers {
			if b.ID == id {
				return i
			}
		}
	}
	if len(browsers) == 0 {
		return -1
	}
	return 0
}

// Situation is everything known about an invocation before showing a window.
type Situation struct {
	Config        prefs.Config
	Problems      []error
	ValidationErr error
	Analysis      *clean.Analysis
	Browsers      []discovery.Browser
}

// NeedWindow reports whether the inspector window must be shown. Errors and
// configuration problems always need it; otherwise it follows the window
// preference.
func NeedWindow(s Situation) bool {
	if s.ValidationErr != nil || len(s.Problems) > 0 || len(s.Browsers) == 0 || s.Analysis == nil {
		return true
	}
	if s.Config.Window == prefs.WindowWhenSuggestions {
		return len(s.Analysis.Suggestions) > 0
	}
	return true
}

// Opener launches a browser with a URL.
type Opener func(b discovery.Browser, url string) error

// Open launches b with url and, only when that succeeds, records b as the
// last-used browser in the state file in dir.
func Open(dir string, st prefs.State, b discovery.Browser, url string, open Opener) error {
	if err := open(b, url); err != nil {
		return err
	}
	st.LastUsed = b.ID
	return prefs.SaveState(dir, st)
}
