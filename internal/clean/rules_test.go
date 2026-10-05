package clean

import (
	"strings"
	"testing"
)

func TestBuiltinRulesValid(t *testing.T) {
	rules, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range rules {
		if strings.TrimSpace(r.Reason) == "" {
			t.Errorf("rule %d has no reason", i)
		}
		if r.ID == "" {
			t.Errorf("rule %d has no id", i)
		}
		if r.Source != SourceBuiltin {
			t.Errorf("rule %d source = %q", i, r.Source)
		}
	}
}

func TestBuiltinRulesRequiredPresent(t *testing.T) {
	rules, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	params := map[string]Rule{}
	redirectHosts := map[string]Rule{}
	for _, r := range rules {
		if r.Kind == KindRedirect {
			for _, h := range r.Hosts {
				redirectHosts[h] = r
			}
			continue
		}
		params[r.Param] = r
	}
	for _, p := range []string{"utm_*", "fbclid", "gclid", "dclid", "gbraid", "wbraid", "msclkid",
		"yclid", "mc_eid", "igshid", "_hsenc", "_hsmi", "mkt_tok"} {
		r, ok := params[p]
		if !ok {
			t.Errorf("missing tracking rule %q", p)
			continue
		}
		if r.Kind != KindTracking || len(r.Hosts) != 0 {
			t.Errorf("rule %q: kind %q hosts %v, want tracking on any host", p, r.Kind, r.Hosts)
		}
	}
	for _, h := range []string{"www.google.*", "l.facebook.com", "lm.facebook.com",
		"*.safelinks.protection.outlook.com", "www.youtube.com"} {
		if _, ok := redirectHosts[h]; !ok {
			t.Errorf("missing redirect rule for host %q", h)
		}
	}
	if r, ok := params["tag"]; !ok || r.Kind != KindAffiliate || len(r.Hosts) == 0 {
		t.Errorf("missing host-scoped affiliate rule for Amazon tag: %+v", r)
	}
}

func TestParseRulesRejectsInvalid(t *testing.T) {
	for name, data := range map[string]string{
		"no id":        "[[rule]]\nkind = \"tracking\"\nparam = \"x\"\nreason = \"r\"\n",
		"no reason":    "[[rule]]\nid = \"a\"\nkind = \"tracking\"\nparam = \"x\"\n",
		"no param":     "[[rule]]\nid = \"a\"\nkind = \"tracking\"\nreason = \"r\"\n",
		"unknown kind": "[[rule]]\nid = \"a\"\nkind = \"other\"\nparam = \"x\"\nreason = \"r\"\n",
		"no target":    "[[rule]]\nid = \"a\"\nkind = \"redirect\"\nhosts = [\"a.com\"]\nreason = \"r\"\n",
		"bad glob":     "[[rule]]\nid = \"a\"\nkind = \"tracking\"\nparam = \"x\"\nhosts = [\"[\"]\nreason = \"r\"\n",
		"duplicate id": "[[rule]]\nid = \"a\"\nkind = \"tracking\"\nparam = \"x\"\nreason = \"r\"\n[[rule]]\nid = \"a\"\nkind = \"tracking\"\nparam = \"y\"\nreason = \"r\"\n",
	} {
		if _, err := ParseRules(data, SourceBuiltin); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
