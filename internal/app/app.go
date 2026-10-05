// Package app holds bopen's platform-independent decisions: which browser is
// pre-selected, whether the inspector window is needed, and opening a link.
package app

import (
	"sort"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

// UserRules converts the valid user rules in cfg into analysis rules.
func UserRules(cfg prefs.Config) []clean.Rule {
	var out []clean.Rule
	for _, r := range cfg.ValidUserRules() {
		out = append(out, clean.Rule{
			ID:     r.ID,
			Kind:   clean.Kind(r.Kind),
			Param:  r.Param,
			Hosts:  r.Hosts,
			Reason: r.Reason,
			Source: clean.SourceUser,
		})
	}
	return out
}

// Rules returns the rules analysis uses, in precedence order: the user's
// rules, then the enabled built-in rules.
func Rules(builtin []clean.Rule, cfg prefs.Config) []clean.Rule {
	return clean.Combine(UserRules(cfg), clean.Enabled(builtin, cfg.IsDisabled))
}

// Visible returns the browsers the inspector offers: Ordered without the
// hidden ones.
func Visible(browsers []discovery.Browser, cfg prefs.Config) []discovery.Browser {
	var out []discovery.Browser
	for _, b := range Ordered(browsers, cfg) {
		if !cfg.IsHidden(b.ID) {
			out = append(out, b)
		}
	}
	return out
}

// Ordered returns all browsers with the configured ones first, in the
// configured order, and the rest in their discovery order.
func Ordered(browsers []discovery.Browser, cfg prefs.Config) []discovery.Browser {
	rank := map[string]int{}
	for i, id := range cfg.Browsers.Order {
		if _, ok := rank[id]; !ok {
			rank[id] = i
		}
	}
	var ordered, rest []discovery.Browser
	for _, b := range browsers {
		switch {
		case hasRank(rank, b.ID):
			ordered = append(ordered, b)
		default:
			rest = append(rest, b)
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool { return rank[ordered[i].ID] < rank[ordered[j].ID] })
	return append(ordered, rest...)
}

func hasRank(rank map[string]int, id string) bool {
	_, ok := rank[id]
	return ok
}

// MoveOrder returns the browser order after moving the browser at index i of
// ordered by delta. Identities in the previous order that are not among
// ordered (uninstalled browsers) are kept at the end.
func MoveOrder(ordered []discovery.Browser, previous []string, i, delta int) []string {
	j := i + delta
	ids := make([]string, len(ordered))
	for k, b := range ordered {
		ids[k] = b.ID
	}
	if i < 0 || i >= len(ids) || j < 0 || j >= len(ids) {
		return append(ids, missing(previous, ids)...)
	}
	ids[i], ids[j] = ids[j], ids[i]
	return append(ids, missing(previous, ids)...)
}

func missing(previous, present []string) []string {
	have := map[string]bool{}
	for _, id := range present {
		have[id] = true
	}
	var out []string
	for _, id := range previous {
		if !have[id] {
			out = append(out, id)
			have[id] = true
		}
	}
	return out
}

// Preselect returns the index in browsers of the browser to pre-select: the
// last-used browser, else the previous system default, else the first. It
// returns -1 when browsers is empty. Pass the visible browsers, so hidden
// candidates are skipped.
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
