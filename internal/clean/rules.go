// Package clean analyses URLs for redirect wrappers and tracking parameters
// and composes cleaned URLs from the suggestions the user accepts.
package clean

import (
	_ "embed"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// Kind classifies a rule and the suggestions it produces.
type Kind string

const (
	KindTracking  Kind = "tracking"
	KindAffiliate Kind = "affiliate"
	KindRedirect  Kind = "redirect"
)

// Source names where a rule came from.
type Source string

const (
	SourceBuiltin Source = "builtin"
	SourceUser    Source = "user"
)

// Rule describes one tracking parameter, affiliate parameter or redirect wrapper.
type Rule struct {
	// ID identifies a built-in rule in the user's configuration.
	ID     string   `toml:"id"`
	Kind   Kind     `toml:"kind"`
	Param  string   `toml:"param"`
	Hosts  []string `toml:"hosts"`
	Path   string   `toml:"path"`
	Target []string `toml:"target"`
	Reason string   `toml:"reason"`
	Source Source   `toml:"-"`
}

// matchesHost reports whether host (lower case) matches the rule's host scope.
func (r Rule) matchesHost(host string) bool {
	if len(r.Hosts) == 0 {
		return true
	}
	for _, pattern := range r.Hosts {
		if ok, _ := path.Match(strings.ToLower(pattern), host); ok {
			return true
		}
	}
	return false
}

// matchesParam reports whether name (lower case) matches the rule's parameter.
func (r Rule) matchesParam(name string) bool {
	p := strings.ToLower(r.Param)
	if prefix, ok := strings.CutSuffix(p, "*"); ok {
		return strings.HasPrefix(name, prefix)
	}
	return name == p
}

//go:embed rules/builtin.toml
var builtinTOML string

var builtin = sync.OnceValues(func() ([]Rule, error) {
	return ParseRules(builtinTOML, SourceBuiltin)
})

// Builtin returns the embedded built-in rule set.
func Builtin() ([]Rule, error) {
	return builtin()
}

// ParseRules decodes and validates a TOML rule file.
func ParseRules(data string, source Source) ([]Rule, error) {
	var file struct {
		Rule []Rule `toml:"rule"`
	}
	if _, err := toml.Decode(data, &file); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for i := range file.Rule {
		r := &file.Rule[i]
		r.Source = source
		if err := r.validate(); err != nil {
			return nil, fmt.Errorf("rule %d: %w", i+1, err)
		}
		if seen[r.ID] {
			return nil, fmt.Errorf("rule %d: duplicate id %q", i+1, r.ID)
		}
		seen[r.ID] = true
	}
	return file.Rule, nil
}

// Combine concatenates rule sources in precedence order. Analyse applies the
// first matching rule, so a source listed earlier wins over later ones for
// the same parameter.
func Combine(sources ...[]Rule) []Rule {
	var out []Rule
	for _, src := range sources {
		out = append(out, src...)
	}
	return out
}

// Enabled returns rules without those whose ID is disabled.
func Enabled(rules []Rule, disabled func(id string) bool) []Rule {
	out := make([]Rule, 0, len(rules))
	for _, r := range rules {
		if !disabled(r.ID) {
			out = append(out, r)
		}
	}
	return out
}

func (r Rule) validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return fmt.Errorf("missing reason")
	}
	switch r.Kind {
	case KindTracking, KindAffiliate:
		if r.Param == "" || r.Param == "*" {
			return fmt.Errorf("%s rule needs a param", r.Kind)
		}
	case KindRedirect:
		if len(r.Hosts) == 0 || len(r.Target) == 0 {
			return fmt.Errorf("redirect rule needs hosts and target")
		}
	default:
		return fmt.Errorf("unknown kind %q", r.Kind)
	}
	for _, h := range r.Hosts {
		if _, err := path.Match(h, ""); err != nil {
			return fmt.Errorf("bad host pattern %q: %w", h, err)
		}
	}
	return nil
}
