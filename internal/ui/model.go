// Package ui implements the inspector window.
package ui

import (
	"github.com/blackfyre/bopen/internal/app"
	"github.com/blackfyre/bopen/internal/appearance"
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/clearurls"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

// Registrar makes bopen the default handler and reports whether it is.
type Registrar interface {
	Register() (string, error)
	IsDefault() (isDefault, known bool)
}

// Env is what the settings view reads and changes. It is shared by the
// inspector and the settings view.
type Env struct {
	// ConfigDir holds config.toml; "" when it is unavailable.
	ConfigDir string
	Config    prefs.Config
	State     prefs.State
	// AllBrowsers are all discovered browsers, hidden ones included.
	AllBrowsers []discovery.Browser
	Builtin     []clean.Rule
	// Registrar is nil when registration is unavailable.
	Registrar Registrar
	// ClearURLs is nil when the ClearURLs list is unavailable (no cache
	// directory).
	ClearURLs *ClearURLs
	// Appearance is the host appearance at start-up.
	Appearance appearance.Settings
}

// ClearURLs holds the ClearURLs list state.
type ClearURLs struct {
	Cache   clearurls.Cache
	Fetcher clearurls.Fetcher
	Update  clearurls.Updater
	Meta    clearurls.Meta
	// Rules are the rules from the cached list, loaded while enabled.
	Rules []clean.Rule
}

// Rules returns the rules analysis uses: the user's rules, then the enabled
// built-in rules.
func (e *Env) Rules() []clean.Rule {
	var extra []clean.Rule
	if e.ClearURLs != nil {
		extra = e.ClearURLs.Rules
	}
	return app.Rules(e.Builtin, e.Config, extra)
}

// Update saves one settings change and adopts the resulting preferences.
func (e *Env) Update(mutate func(*prefs.Config)) error {
	if e.ConfigDir == "" {
		mutate(&e.Config)
		return nil
	}
	cfg, err := prefs.UpdateConfig(e.ConfigDir, mutate)
	if err != nil {
		return err
	}
	e.Config = cfg
	return nil
}

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
	// Env provides the preferences; nil disables Refresh.
	Env *Env
	// Invalid is set when Input is not a usable web URL.
	Invalid bool
}

type spanKey struct {
	start, end int
	kind       clean.Kind
}

// Refresh re-applies the preferences in m.Env: browser visibility and order,
// and the enabled rules. Toggles of suggestions that still exist (same span)
// and the selected browser (when still offered) are kept.
func (m *Model) Refresh() {
	if m.Env == nil {
		return
	}
	selected := ""
	if m.Selected >= 0 && m.Selected < len(m.Browsers) {
		selected = m.Browsers[m.Selected].ID
	}
	m.Browsers = app.Visible(m.Env.AllBrowsers, m.Env.Config)
	m.Selected = app.Preselect(m.Browsers, m.Env.State)
	for i, b := range m.Browsers {
		if b.ID == selected {
			m.Selected = i
		}
	}
	if m.Analysis != nil {
		old := map[spanKey]bool{}
		for i, s := range m.Analysis.Suggestions {
			old[spanKey{s.Start, s.End, s.Kind}] = m.Accepted[i]
		}
		m.Analysis = clean.Analyse(m.Analysis.URL, m.Env.Rules())
		m.Accepted = m.Analysis.Defaults()
		for i, s := range m.Analysis.Suggestions {
			if v, ok := old[spanKey{s.Start, s.End, s.Kind}]; ok {
				m.Accepted[i] = v
			}
		}
	}
	m.Blocker = ""
	switch {
	case m.Invalid:
		m.Blocker = "This link cannot be opened: only http and https links are accepted."
	case len(m.Env.AllBrowsers) == 0:
		m.Blocker = "No web browsers were found on this system."
	case len(m.Browsers) == 0:
		m.Blocker = "All browsers are hidden. Open the settings (cog, top right) to show one."
	}
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

// Move moves the browser selection by delta, clamped to the list.
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
	// param is the index in Analysis.Params of the parameter the run
	// belongs to, or -1.
	param int
}

// runs splits the analysed URL into runs by the suggestion covering each
// byte. Parameter suggestions take precedence over the redirect wrappers
// containing them, and later (inner) suggestions over earlier ones.
func (m *Model) runs() []styleRun {
	if m.Analysis == nil {
		return []styleRun{{text: m.Input, param: -1}}
	}
	u := m.Analysis.URL
	owner := make([]int, len(u))
	param := make([]int, len(u))
	for i := range owner {
		owner[i], param[i] = -1, -1
	}
	// Parameters inside a redirect target come later and win over the
	// wrapper parameter that contains them.
	for i, p := range m.Analysis.Params {
		for b := p.Start; b < p.End && b < len(u); b++ {
			param[b] = i
		}
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
	style := func(b int) styleRun {
		o := owner[b]
		if o < 0 {
			return styleRun{param: param[b]}
		}
		return styleRun{
			kind:     m.Analysis.Suggestions[o].Kind,
			accepted: m.Accepted[o] && m.Analysis.Available(o, m.Accepted),
			param:    param[b],
		}
	}
	var out []styleRun
	start := 0
	for b := 1; b <= len(u); b++ {
		if b < len(u) && style(b) == style(start) {
			continue
		}
		r := style(start)
		r.text = u[start:b]
		out = append(out, r)
		start = b
	}
	return out
}
