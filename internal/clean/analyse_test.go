package clean

import (
	"net/url"
	"strings"
	"testing"
)

func analyse(t *testing.T, raw string) *Analysis {
	t.Helper()
	rules, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	return Analyse(raw, rules)
}

// spanText returns the text of the original URL covered by suggestion s.
func spanText(a *Analysis, s Suggestion) string {
	return a.URL[s.Start:s.End]
}

func TestTrackingParameterExplained(t *testing.T) {
	a := analyse(t, "https://example.com/a?id=7&utm_source=newsletter")
	if len(a.Suggestions) != 1 {
		t.Fatalf("suggestions = %+v, want one", a.Suggestions)
	}
	s := a.Suggestions[0]
	if s.Kind != KindTracking || spanText(a, s) != "utm_source=newsletter" || s.Text != "utm_source=newsletter" ||
		s.Reason == "" || s.Source != SourceBuiltin || !s.Default || s.DependsOn != -1 {
		t.Fatalf("unexpected suggestion %+v", s)
	}
}

func TestCleanURLNoSuggestions(t *testing.T) {
	if a := analyse(t, "https://example.com/a?id=7"); len(a.Suggestions) != 0 {
		t.Fatalf("suggestions = %+v", a.Suggestions)
	}
	if a := analyse(t, "https://example.com/plain"); len(a.Suggestions) != 0 {
		t.Fatalf("suggestions = %+v", a.Suggestions)
	}
}

func TestTrackingParameters(t *testing.T) {
	for _, tc := range []struct{ url, span string }{
		{"https://example.com/?msclkid=abc", "msclkid=abc"},
		{"https://example.com/?x=1&UTM_Campaign=spring", "UTM_Campaign=spring"},
		{"https://example.com/?fbclid=IwAR3x&b=2", "fbclid=IwAR3x"},
		{"https://example.com/?a=1&_hsenc=p2&_hsmi=9", "_hsenc=p2"},
	} {
		a := analyse(t, tc.url)
		if len(a.Suggestions) == 0 {
			t.Errorf("%s: no suggestions", tc.url)
			continue
		}
		s := a.Suggestions[0]
		if s.Kind != KindTracking || spanText(a, s) != tc.span {
			t.Errorf("%s: got %q (%s), want %q", tc.url, spanText(a, s), s.Kind, tc.span)
		}
	}
}

func TestAmazonAffiliateTag(t *testing.T) {
	for _, raw := range []string{"https://www.amazon.de/dp/B000?tag=someone-21", "https://amazon.com/dp/B000?tag=someone-21"} {
		a := analyse(t, raw)
		if len(a.Suggestions) != 1 {
			t.Fatalf("%s: suggestions = %+v", raw, a.Suggestions)
		}
		s := a.Suggestions[0]
		if s.Kind != KindAffiliate || spanText(a, s) != "tag=someone-21" || s.Default {
			t.Fatalf("%s: unexpected %+v", raw, s)
		}
	}
}

func TestHostScopeRespected(t *testing.T) {
	if a := analyse(t, "https://blog.example.com/post?tag=golang"); len(a.Suggestions) != 0 {
		t.Fatalf("suggestions = %+v", a.Suggestions)
	}
}

func TestRedirectWrappers(t *testing.T) {
	target := "https://shop.example.com/item?colour=red"
	enc := url.QueryEscape(target)
	for _, raw := range []string{
		"https://www.google.com/url?q=" + enc + "&sa=D&ust=123",
		"https://www.google.co.uk/url?sa=t&url=" + enc,
		"https://l.facebook.com/l.php?u=" + enc + "&h=AT0",
		"https://lm.facebook.com/l.php?u=" + enc,
		"https://eur01.safelinks.protection.outlook.com/?url=" + enc + "&data=05",
		"https://www.youtube.com/redirect?event=video&q=" + enc,
	} {
		a := analyse(t, raw)
		if len(a.Suggestions) == 0 || a.Suggestions[0].Kind != KindRedirect {
			t.Errorf("%s: no redirect suggestion: %+v", raw, a.Suggestions)
			continue
		}
		s := a.Suggestions[0]
		if s.Target != target || s.Start != 0 || s.End != len(raw) {
			t.Errorf("%s: target %q span [%d,%d)", raw, s.Target, s.Start, s.End)
		}
		if got := a.Clean(a.Defaults()); got != target {
			t.Errorf("%s: cleaned %q, want %q", raw, got, target)
		}
	}
}

func TestGoogleWrapperScenario(t *testing.T) {
	a := analyse(t, "https://www.google.com/url?q=https%3A%2F%2Fshop.example.com%2Fitem&sa=D")
	if got := a.Clean(a.Defaults()); got != "https://shop.example.com/item" {
		t.Fatalf("cleaned %q", got)
	}
}

func TestTrackingInsideWrappedTarget(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/item?fbclid=x&id=1")
	a := analyse(t, raw)
	if len(a.Suggestions) != 2 {
		t.Fatalf("suggestions = %+v", a.Suggestions)
	}
	r, tr := a.Suggestions[0], a.Suggestions[1]
	if r.Kind != KindRedirect || tr.Kind != KindTracking || tr.DependsOn != 0 || tr.Text != "fbclid=x" {
		t.Fatalf("unexpected suggestions %+v", a.Suggestions)
	}
	if got := spanText(a, tr); got != "fbclid%3Dx" {
		t.Fatalf("span in original = %q, want encoded fbclid", got)
	}
	if got := a.Clean(a.Defaults()); got != "https://shop.example.com/item?id=1" {
		t.Fatalf("cleaned %q", got)
	}
}

func TestWrapperWithNonWebTarget(t *testing.T) {
	a := analyse(t, "https://www.google.com/url?q=javascript%3Aalert(1)")
	for _, s := range a.Suggestions {
		if s.Kind == KindRedirect {
			t.Fatalf("unexpected redirect %+v", s)
		}
	}
}

func TestRedirectDepthLimit(t *testing.T) {
	raw := "https://example.com/final"
	for i := 0; i < MaxRedirectDepth+2; i++ {
		raw = "https://www.google.com/url?q=" + url.QueryEscape(raw)
	}
	a := analyse(t, raw)
	redirects := 0
	for _, s := range a.Suggestions {
		if s.Kind == KindRedirect {
			redirects++
		}
	}
	if redirects != MaxRedirectDepth {
		t.Fatalf("redirects = %d, want %d", redirects, MaxRedirectDepth)
	}
	if got := a.Clean(a.Defaults()); !strings.HasPrefix(got, "https://www.google.com/url?q=") {
		t.Fatalf("cleaned %q, want the innermost unexpanded wrapper", got)
	}
}

func analyseWithout(t *testing.T, raw string, disabled ...string) *Analysis {
	t.Helper()
	rules, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	off := func(id string) bool {
		for _, d := range disabled {
			if d == id {
				return true
			}
		}
		return false
	}
	return Analyse(raw, Enabled(rules, off))
}

func TestDisabledTrackingRule(t *testing.T) {
	if a := analyseWithout(t, "https://example.com/?fbclid=x", "fbclid"); len(a.Suggestions) != 0 {
		t.Fatalf("suggestions = %+v", a.Suggestions)
	}
	if a := analyseWithout(t, "https://example.com/?fbclid=x", "no-such-rule"); len(a.Suggestions) != 1 {
		t.Fatalf("unknown disabled id changed the result: %+v", a.Suggestions)
	}
}

func TestDisabledRedirectRule(t *testing.T) {
	a := analyseWithout(t, "https://www.google.com/url?q="+url.QueryEscape("https://example.com/"), "redirect-google")
	for _, s := range a.Suggestions {
		if s.Kind == KindRedirect {
			t.Fatalf("unexpected redirect %+v", s)
		}
	}
}

func TestParamsRecordedWithHosts(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/item?ref=home&fbclid=x") + "&sa=D"
	a := analyse(t, raw)
	byName := map[string]Param{}
	for _, p := range a.Params {
		byName[p.Name] = p
	}
	ref, ok := byName["ref"]
	if !ok || ref.Host != "shop.example.com" || ref.Suggestion != -1 || a.URL[ref.Start:ref.End] != "ref%3Dhome" {
		t.Fatalf("ref = %+v", ref)
	}
	if sa := byName["sa"]; sa.Host != "www.google.com" {
		t.Fatalf("sa = %+v", sa)
	}
	if fb := byName["fbclid"]; fb.Suggestion < 0 || a.Suggestions[fb.Suggestion].RuleID != "fbclid" {
		t.Fatalf("fbclid = %+v", fb)
	}
}

func TestMalformedEscapesStillAnalysed(t *testing.T) {
	a := analyse(t, "https://example.com/100%real?fbclid=x")
	if len(a.Suggestions) != 1 || a.Suggestions[0].Text != "fbclid=x" {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
	a = analyse(t, "https://example.com/a%zz?utm_source=x")
	if got := a.Clean(a.Defaults()); got != "https://example.com/a%zz" {
		t.Fatalf("cleaned %q", got)
	}
	a = analyse(t, "https://example.com/?q=50%&gclid=y")
	if len(a.Suggestions) != 1 || a.Suggestions[0].Text != "gclid=y" {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
}

func TestParseTolerant(t *testing.T) {
	for in, want := range map[string]string{
		"https://example.com/100%":      "/100%",
		"https://example.com/%zz/%41":   "/%zz/A",
		"https://example.com/caf%C3%A9": "/café",
	} {
		u, err := ParseTolerant(in)
		if err != nil || u.Path != want {
			t.Errorf("%q: path %q, %v; want %q", in, u.Path, err, want)
		}
	}
}
