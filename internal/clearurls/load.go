package clearurls

import (
	"encoding/json"
	"errors"
	"regexp"
	"sort"

	"github.com/blackfyre/bopen/internal/clean"
)

type file struct {
	Providers map[string]provider `json:"providers"`
}

type provider struct {
	URLPattern        string   `json:"urlPattern"`
	CompleteProvider  bool     `json:"completeProvider"`
	Rules             []string `json:"rules"`
	RawRules          []string `json:"rawRules"`
	ReferralMarketing []string `json:"referralMarketing"`
	Exceptions        []string `json:"exceptions"`
	Redirections      []string `json:"redirections"`
	ForceRedirection  bool     `json:"forceRedirection"`
}

const globalProvider = "globalRules"

// Rules converts a ClearURLs rule list into bopen rules. Patterns that do
// not compile as Go regular expressions are skipped and counted.
// completeProvider and forceRedirection are ignored. Global rules come last,
// so a site-specific provider's reason wins for the same parameter.
func Rules(data []byte) ([]clean.Rule, int, error) {
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, 0, err
	}
	if len(f.Providers) == 0 {
		return nil, 0, errors.New("rule list has no providers")
	}
	names := make([]string, 0, len(f.Providers))
	for name := range f.Providers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		gi, gj := names[i] == globalProvider, names[j] == globalProvider
		if gi != gj {
			return gj
		}
		return names[i] < names[j]
	})

	var rules []clean.Rule
	skipped := 0
	compile := func(expr string) *regexp.Regexp {
		re, err := regexp.Compile("(?i)" + expr)
		if err != nil {
			skipped++
			return nil
		}
		return re
	}
	for _, name := range names {
		p := f.Providers[name]
		urlRE := compile(p.URLPattern)
		if urlRE == nil {
			skipped += len(p.Rules) + len(p.RawRules) + len(p.ReferralMarketing) + len(p.Redirections)
			continue
		}
		prov := &clean.Provider{Name: displayName(name), URL: urlRE}
		for _, ex := range p.Exceptions {
			if re := compile(ex); re != nil {
				prov.Exceptions = append(prov.Exceptions, re)
			}
		}
		add := func(kind clean.Kind, pattern clean.Pattern, what string) {
			pattern.Provider = prov
			rules = append(rules, clean.Rule{
				ID:      "clearurls:" + name,
				Kind:    kind,
				Reason:  "Listed by ClearURLs as " + what + " for " + prov.Name + ".",
				Source:  clean.SourceClearURLs,
				Pattern: &pattern,
			})
		}
		for _, r := range p.Redirections {
			if re := compile(r); re != nil && re.NumSubexp() >= 1 {
				add(clean.KindRedirect, clean.Pattern{Redirect: re}, "a redirect wrapper")
			} else if re != nil {
				skipped++
			}
		}
		for _, r := range p.Rules {
			if re := compile("^(?:" + r + ")$"); re != nil {
				add(clean.KindTracking, clean.Pattern{Param: re}, "tracking")
			}
		}
		for _, r := range p.ReferralMarketing {
			if re := compile("^(?:" + r + ")$"); re != nil {
				add(clean.KindAffiliate, clean.Pattern{Param: re}, "referral marketing")
			}
		}
		for _, r := range p.RawRules {
			if re := compile(r); re != nil {
				add(clean.KindTracking, clean.Pattern{Raw: re}, "tracking")
			}
		}
	}
	return rules, skipped, nil
}

func displayName(provider string) string {
	if provider == globalProvider {
		return "all sites"
	}
	return provider
}

// Load returns the rules from the cached list. An absent or invalid cache
// yields no rules, so analysis behaves as if ClearURLs were disabled.
func Load(c Cache) []clean.Rule {
	data, err := c.Data()
	if err != nil {
		return nil
	}
	rules, _, err := Rules(data)
	if err != nil {
		return nil
	}
	return rules
}
