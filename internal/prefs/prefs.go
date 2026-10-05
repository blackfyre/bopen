// Package prefs reads and writes bopen's preferences (config.toml) and
// runtime state (state.toml) in the OS user configuration directory.
package prefs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	configFile = "config.toml"
	stateFile  = "state.toml"
)

// Window controls when the inspector window is shown.
type Window string

const (
	WindowAlways          Window = "always"
	WindowWhenSuggestions Window = "when-suggestions"
)

// Config holds the user preferences.
type Config struct {
	Window   Window         `toml:"window"`
	Browsers BrowsersConfig `toml:"browsers"`
	Rules    RulesConfig    `toml:"rules"`
	// Sites are the site rules, in matching order.
	Sites []SiteRule `toml:"sites,omitempty"`
}

// SiteRule sends links for matching hosts to a browser.
type SiteRule struct {
	ID string `toml:"id"`
	// Hosts are host patterns; * matches any run of characters.
	Hosts []string `toml:"hosts"`
	// Browser is the identity of the browser to use.
	Browser string `toml:"browser"`
	// Direct opens matching links without the inspector.
	Direct bool `toml:"direct,omitempty"`
}

// Validate reports why the site rule cannot be applied, or nil.
func (r SiteRule) Validate() error {
	if len(r.Hosts) == 0 {
		return errors.New("at least one host pattern is required")
	}
	for _, h := range r.Hosts {
		if strings.TrimSpace(h) == "" {
			return errors.New("empty host pattern")
		}
		if _, err := path.Match(h, ""); err != nil {
			return fmt.Errorf("invalid host pattern %q", h)
		}
	}
	if strings.TrimSpace(r.Browser) == "" {
		return errors.New("a browser is required")
	}
	return nil
}

// SiteRule returns the site rule with identifier id.
func (c Config) SiteRule(id string) (SiteRule, bool) {
	for _, r := range c.Sites {
		if r.ID == id {
			return r, true
		}
	}
	return SiteRule{}, false
}

// AddSiteRule appends r with a new unique identifier and returns it.
func (c *Config) AddSiteRule(r SiteRule) string {
	for {
		r.ID = "s-" + randomHex()
		if _, taken := c.SiteRule(r.ID); !taken {
			break
		}
	}
	c.Sites = append(c.Sites, r)
	return r.ID
}

// SetSiteRule replaces the site rule with r's identifier. It reports
// whether the rule existed.
func (c *Config) SetSiteRule(r SiteRule) bool {
	for i := range c.Sites {
		if c.Sites[i].ID == r.ID {
			c.Sites[i] = r
			return true
		}
	}
	return false
}

// DeleteSiteRule removes the site rule with identifier id.
func (c *Config) DeleteSiteRule(id string) {
	c.Sites = slices.DeleteFunc(c.Sites, func(r SiteRule) bool { return r.ID == id })
}

func randomHex() string {
	var b [3]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// BrowsersConfig controls which browsers the inspector offers and in which
// order. Identities of browsers that are not currently installed are kept.
type BrowsersConfig struct {
	// Hidden lists browser identities not offered in the inspector.
	Hidden []string `toml:"hidden,omitempty"`
	// Order lists browser identities shown first, in this order.
	Order []string `toml:"order,omitempty"`
}

// RulesConfig controls the rule set.
type RulesConfig struct {
	// Disabled lists built-in rule identifiers that produce no suggestions.
	Disabled []string `toml:"disabled,omitempty"`
	// User holds the user's own rules.
	User []UserRule `toml:"user,omitempty"`
	// ClearURLs enables the ClearURLs rule list and its download.
	ClearURLs bool `toml:"clearurls,omitempty"`
}

// UserRule is a tracking or affiliate parameter rule written by the user.
type UserRule struct {
	ID     string   `toml:"id"`
	Kind   string   `toml:"kind"`
	Param  string   `toml:"param"`
	Hosts  []string `toml:"hosts,omitempty"`
	Reason string   `toml:"reason"`
}

// Validate reports why the rule cannot be applied, or nil.
func (r UserRule) Validate() error {
	switch {
	case strings.TrimSpace(r.Param) == "" || strings.TrimSpace(r.Param) == "*":
		return errors.New("a parameter name is required")
	case r.Kind != "tracking" && r.Kind != "affiliate":
		return fmt.Errorf("kind must be \"tracking\" or \"affiliate\", not %q", r.Kind)
	case strings.TrimSpace(r.Reason) == "":
		return errors.New("a reason is required")
	}
	for _, h := range r.Hosts {
		if _, err := path.Match(h, ""); err != nil || strings.TrimSpace(h) == "" {
			return fmt.Errorf("invalid host pattern %q", h)
		}
	}
	return nil
}

// ValidUserRules returns the user rules that pass Validate.
func (c Config) ValidUserRules() []UserRule {
	var out []UserRule
	for _, r := range c.Rules.User {
		if r.Validate() == nil {
			out = append(out, r)
		}
	}
	return out
}

// UserRule returns the user rule with identifier id.
func (c Config) UserRule(id string) (UserRule, bool) {
	for _, r := range c.Rules.User {
		if r.ID == id {
			return r, true
		}
	}
	return UserRule{}, false
}

// AddUserRule appends r with a new unique identifier and returns that identifier.
func (c *Config) AddUserRule(r UserRule) string {
	for {
		r.ID = "u-" + randomHex()
		if _, taken := c.UserRule(r.ID); !taken {
			break
		}
	}
	c.Rules.User = append(c.Rules.User, r)
	return r.ID
}

// SetUserRule replaces the user rule with r's identifier. It reports whether
// the rule existed.
func (c *Config) SetUserRule(r UserRule) bool {
	for i := range c.Rules.User {
		if c.Rules.User[i].ID == r.ID {
			c.Rules.User[i] = r
			return true
		}
	}
	return false
}

// DeleteUserRule removes the user rule with identifier id.
func (c *Config) DeleteUserRule(id string) {
	c.Rules.User = slices.DeleteFunc(c.Rules.User, func(r UserRule) bool { return r.ID == id })
}

// DisableRule records the built-in rule id as disabled.
func (c *Config) DisableRule(id string) {
	if !c.IsDisabled(id) {
		c.Rules.Disabled = append(c.Rules.Disabled, id)
	}
}

// IsHidden reports whether the browser with identity id is hidden.
func (c Config) IsHidden(id string) bool {
	return contains(c.Browsers.Hidden, id)
}

// IsDisabled reports whether the built-in rule with identifier id is disabled.
func (c Config) IsDisabled(id string) bool {
	return contains(c.Rules.Disabled, id)
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// DefaultConfig returns the preferences used when none are configured.
func DefaultConfig() Config {
	return Config{Window: WindowAlways}
}

// State holds runtime state that bopen writes itself.
type State struct {
	LastUsed        string `toml:"last_used,omitempty"`
	PreviousDefault string `toml:"previous_default,omitempty"`
}

// Dir returns the bopen directory inside the OS user configuration directory.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "bopen"), nil
}

// CacheDir returns the bopen directory inside the OS user cache directory,
// where downloaded rule data is kept.
func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "bopen"), nil
}

// WriteFileAtomic writes data to path through a temporary file in the same
// directory, so readers never see a partial file.
func WriteFileAtomic(path string, data []byte) error {
	return writeFileAtomic(path, data)
}

// LoadConfig reads config.toml from dir. A missing file yields the defaults
// without problems. Parse errors and invalid values are returned as problems,
// and the affected settings keep their defaults. Unknown keys are ignored.
func LoadConfig(dir string) (Config, []error) {
	cfg := DefaultConfig()
	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, []error{fmt.Errorf("%s: %w", path, err)}
	}
	var raw Config
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return cfg, []error{fmt.Errorf("%s: %w", path, err)}
	}
	cfg.Browsers = raw.Browsers
	cfg.Rules = raw.Rules
	cfg.Sites = raw.Sites
	var problems []error
	for i, r := range raw.Sites {
		if err := r.Validate(); err != nil {
			name := r.ID
			if name == "" {
				name = fmt.Sprintf("#%d", i+1)
			}
			problems = append(problems, fmt.Errorf("%s: site rule %s (%s) is ignored: %w", path, name, strings.Join(r.Hosts, ", "), err))
		}
	}
	for i, r := range raw.Rules.User {
		if err := r.Validate(); err != nil {
			name := r.ID
			if name == "" {
				name = fmt.Sprintf("#%d", i+1)
			}
			problems = append(problems, fmt.Errorf("%s: user rule %s (%s) is ignored: %w", path, name, r.Param, err))
		}
	}
	switch raw.Window {
	case "":
	case WindowAlways, WindowWhenSuggestions:
		cfg.Window = raw.Window
	default:
		problems = append(problems, fmt.Errorf("%s: invalid window value %q (want %q or %q)",
			path, raw.Window, WindowAlways, WindowWhenSuggestions))
	}
	return cfg, problems
}

// UpdateConfig applies one change to config.toml in dir: it re-reads the
// file, applies mutate to the decoded preferences and writes the result
// atomically, so changes saved meanwhile by another bopen instance survive.
// It refuses to overwrite a file it cannot parse. Comments are not kept.
func UpdateConfig(dir string, mutate func(*Config)) (Config, error) {
	path := filepath.Join(dir, configFile)
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return cfg, err
	default:
		if _, err := toml.Decode(string(data), &cfg); err != nil {
			return cfg, fmt.Errorf("%s: %w; fix or remove the file before changing settings", path, err)
		}
		if cfg.Window == "" {
			cfg.Window = WindowAlways
		}
	}
	mutate(&cfg)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return cfg, err
	}
	out, err := toml.Marshal(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, writeFileAtomic(path, out)
}

// LoadState reads state.toml from dir. A missing or unreadable file yields
// an empty state.
func LoadState(dir string) State {
	var st State
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if err != nil {
		return State{}
	}
	if _, err := toml.Decode(string(data), &st); err != nil {
		return State{}
	}
	return st
}

// SaveState writes state.toml to dir atomically, creating dir if needed.
func SaveState(dir string, st State) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(st)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(dir, stateFile), data)
}

// writeFileAtomic writes data to a temporary file next to path and renames it
// over path, so an interrupted write leaves the previous contents intact.
func writeFileAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
