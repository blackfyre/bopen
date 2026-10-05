package clean

import (
	"net/url"
	"strings"
)

// MaxRedirectDepth is the maximum number of nested redirect wrappers unwrapped.
const MaxRedirectDepth = 5

// Suggestion is one proposed change to a URL.
type Suggestion struct {
	Kind Kind
	// Start and End are byte offsets of the affected span in Analysis.URL.
	Start, End int
	// Text is the affected part as shown to the user: "name=value" for
	// parameters, the wrapper's host and path for redirects.
	Text   string
	Reason string
	Source Source
	// RuleID identifies the rule that produced the suggestion.
	RuleID  string
	Default bool
	// DependsOn is the index of the redirect suggestion that revealed this
	// suggestion, or -1.
	DependsOn int
	// Target is the unwrapped URL of a redirect suggestion.
	Target string
}

// Param is one query parameter found in the analysed URL or in an unwrapped
// redirect target.
type Param struct {
	// Start and End are byte offsets of the parameter in Analysis.URL.
	Start, End int
	// Name is the decoded parameter name.
	Name string
	// Host is the host of the URL that contains the parameter.
	Host string
	// Suggestion is the index of the suggestion covering it, or -1.
	Suggestion int
}

// Analysis is the result of analysing one URL.
type Analysis struct {
	URL         string
	Suggestions []Suggestion
	Params      []Param
	root        *node
}

// node is one URL in a chain of redirect wrappers. text is the URL as
// decoded from its parent; starts and ends map each byte of text back to its
// span in the analysed URL (nil for the root, which maps to itself).
type node struct {
	text         string
	starts, ends []int
	dep          int

	prefix, suffix string
	segs           []segment
	// raws are removable spans outside the query, in node text offsets.
	raws []rawSpan

	redirect int
	child    *node
}

type segment struct {
	raw        string
	start, end int
	sug        int
}

type rawSpan struct {
	start, end int
	sug        int
}

// scope decides, once per node, whether rules of each imported provider
// apply to the node's URL.
type scope struct {
	text      string
	providers map[*Provider]bool
}

func (sc *scope) applies(r Rule, host string) bool {
	if r.Pattern == nil {
		return r.matchesHost(host)
	}
	p := r.Pattern.Provider
	if v, ok := sc.providers[p]; ok {
		return v
	}
	v := p.URL.MatchString(sc.text)
	for _, ex := range p.Exceptions {
		if v && ex.MatchString(sc.text) {
			v = false
		}
	}
	sc.providers[p] = v
	return v
}

// Analyse analyses rawURL, which must be an absolute http or https URL,
// against rules. It performs no network access.
func Analyse(rawURL string, rules []Rule) *Analysis {
	a := &Analysis{URL: rawURL}
	a.root = &node{text: rawURL, dep: -1}
	a.analyse(a.root, rules, 0)
	return a
}

func (a *Analysis) analyse(n *node, rules []Rule, depth int) {
	n.redirect = -1
	n.split()
	u, err := url.Parse(n.text)
	if err != nil {
		return
	}
	host := strings.ToLower(u.Hostname())
	sc := &scope{text: n.text, providers: map[*Provider]bool{}}

	targetSeg := -1
	if depth < MaxRedirectDepth {
		targetSeg = a.unwrap(n, u, host, rules, sc)
	}

	for i := range n.segs {
		if i == targetSeg {
			continue
		}
		seg := &n.segs[i]
		name, _, _ := strings.Cut(seg.raw, "=")
		rawName := name
		if dec, err := url.QueryUnescape(name); err == nil {
			name = dec
		}
		if name == "" {
			continue
		}
		start, end := n.span(seg.start, seg.end)
		param := len(a.Params)
		a.Params = append(a.Params, Param{Start: start, End: end, Name: name, Host: host, Suggestion: -1})
		name = strings.ToLower(name)
		for _, r := range rules {
			if r.Kind == KindRedirect || !r.paramRule() || !sc.applies(r, host) || !r.matchesParamName(name, rawName) {
				continue
			}
			seg.sug = len(a.Suggestions)
			a.Params[param].Suggestion = seg.sug
			a.Suggestions = append(a.Suggestions, Suggestion{
				Kind:      r.Kind,
				Start:     start,
				End:       end,
				Text:      seg.raw,
				Reason:    r.Reason,
				Source:    r.Source,
				RuleID:    r.ID,
				Default:   r.Kind != KindAffiliate,
				DependsOn: n.dep,
			})
			break
		}
	}

	a.matchRaw(n, rules, host, sc)

	if n.child != nil {
		a.analyse(n.child, rules, depth+1)
	}
}

// matchRaw applies raw-text rules. A match before the query (path) or after
// it (fragment) becomes its own span. A match inside the query counts only
// when it covers exactly one otherwise unflagged parameter (with at most its
// separators); other matches are dropped so highlights never overlap.
func (a *Analysis) matchRaw(n *node, rules []Rule, host string, sc *scope) {
	queryStart, queryEnd := len(n.prefix), len(n.text)-len(n.suffix)
	for _, r := range rules {
		if r.Pattern == nil || r.Pattern.Raw == nil || !sc.applies(r, host) {
			continue
		}
		for _, m := range r.Pattern.Raw.FindAllStringIndex(n.text, -1) {
			s, e := m[0], m[1]
			if s == e {
				continue
			}
			switch {
			case e <= queryStart || s >= queryEnd:
				if a.rawTaken(n, s, e) {
					continue
				}
				start, end := n.span(s, e)
				n.raws = append(n.raws, rawSpan{start: s, end: e, sug: len(a.Suggestions)})
				a.Suggestions = append(a.Suggestions, a.suggestion(r, n, start, end, n.text[s:e]))
			default:
				for i := range n.segs {
					seg := &n.segs[i]
					if seg.sug >= 0 || seg.start-s > 1 || seg.start < s || e < seg.end || e-seg.end > 1 {
						continue
					}
					start, end := n.span(seg.start, seg.end)
					seg.sug = len(a.Suggestions)
					a.Suggestions = append(a.Suggestions, a.suggestion(r, n, start, end, seg.raw))
					for p := range a.Params {
						if a.Params[p].Start == start && a.Params[p].End == end {
							a.Params[p].Suggestion = seg.sug
						}
					}
				}
			}
		}
	}
}

func (a *Analysis) rawTaken(n *node, s, e int) bool {
	for _, r := range n.raws {
		if s < r.end && r.start < e {
			return true
		}
	}
	return false
}

func (a *Analysis) suggestion(r Rule, n *node, start, end int, text string) Suggestion {
	return Suggestion{
		Kind:      r.Kind,
		Start:     start,
		End:       end,
		Text:      text,
		Reason:    r.Reason,
		Source:    r.Source,
		RuleID:    r.ID,
		Default:   r.Kind != KindAffiliate,
		DependsOn: n.dep,
	}
}

// unwrap looks for a redirect rule matching n and, when its target is a web
// URL, records the redirect suggestion and the child node. It returns the
// index of the segment holding the target, or -1.
func (a *Analysis) unwrap(n *node, u *url.URL, host string, rules []Rule, sc *scope) int {
	for _, r := range rules {
		if r.Kind != KindRedirect {
			continue
		}
		if r.Pattern != nil {
			if r.Pattern.Redirect == nil || !sc.applies(r, host) {
				continue
			}
			m := r.Pattern.Redirect.FindStringSubmatchIndex(n.text)
			if m == nil || len(m) < 4 || m[2] < 0 {
				continue
			}
			if seg, ok := a.redirectTo(n, u, host, r, m[2], m[3]); ok {
				return seg
			}
			continue
		}
		if !r.matchesHost(host) || (r.Path != "" && u.Path != r.Path) {
			continue
		}
		for _, param := range r.Target {
			for i, seg := range n.segs {
				name, value, ok := strings.Cut(seg.raw, "=")
				if !ok || !strings.EqualFold(name, param) {
					continue
				}
				valueStart := seg.start + len(name) + 1
				if _, ok := a.redirectTo(n, u, host, r, valueStart, valueStart+len(value)); !ok {
					return -1
				}
				return i
			}
		}
		return -1
	}
	return -1
}

// redirectTo records a redirect suggestion whose percent-encoded target is
// n.text[s:e]. It returns the index of the segment containing the target
// (or -1) and whether the target is a web URL.
func (a *Analysis) redirectTo(n *node, u *url.URL, host string, r Rule, s, e int) (int, bool) {
	text, starts, ends, ok := unescapeMapped(n.text[s:e], s)
	if !ok || !IsWebURL(text) {
		return -1, false
	}
	for j := range starts {
		starts[j], _ = n.span(starts[j], starts[j]+1)
		_, ends[j] = n.span(ends[j]-1, ends[j])
	}
	start, end := n.span(0, len(n.text))
	n.redirect = len(a.Suggestions)
	a.Suggestions = append(a.Suggestions, Suggestion{
		Kind:      KindRedirect,
		Start:     start,
		End:       end,
		Text:      host + u.EscapedPath(),
		Reason:    r.Reason,
		Source:    r.Source,
		RuleID:    r.ID,
		Default:   true,
		DependsOn: n.dep,
		Target:    text,
	})
	n.child = &node{text: text, starts: starts, ends: ends, dep: n.redirect}
	for i, seg := range n.segs {
		if seg.start <= s && e <= seg.end {
			return i, true
		}
	}
	return -1, true
}

// split separates n.text into prefix, query segments and fragment.
func (n *node) split() {
	before := n.text
	if h := strings.IndexByte(n.text, '#'); h >= 0 {
		before, n.suffix = n.text[:h], n.text[h:]
	}
	q := strings.IndexByte(before, '?')
	if q < 0 {
		n.prefix = before
		return
	}
	n.prefix = before[:q]
	offset := q + 1
	for _, raw := range strings.Split(before[q+1:], "&") {
		n.segs = append(n.segs, segment{raw: raw, start: offset, end: offset + len(raw), sug: -1})
		offset += len(raw) + 1
	}
}

// span maps the byte range [start, end) of n.text to the analysed URL.
func (n *node) span(start, end int) (int, int) {
	if n.starts == nil || end <= start {
		return start, end
	}
	return n.starts[start], n.ends[end-1]
}

// unescapeMapped decodes a query value like url.QueryUnescape and records,
// for every decoded byte, the span of the encoded input it came from.
func unescapeMapped(s string, offset int) (string, []int, []int, bool) {
	var b strings.Builder
	var starts, ends []int
	for i := 0; i < len(s); {
		switch c := s[i]; {
		case c == '%':
			if i+2 >= len(s) {
				return "", nil, nil, false
			}
			hi, ok1 := unhex(s[i+1])
			lo, ok2 := unhex(s[i+2])
			if !ok1 || !ok2 {
				return "", nil, nil, false
			}
			b.WriteByte(hi<<4 | lo)
			starts, ends = append(starts, offset+i), append(ends, offset+i+3)
			i += 3
		case c == '+':
			b.WriteByte(' ')
			starts, ends = append(starts, offset+i), append(ends, offset+i+1)
			i++
		default:
			b.WriteByte(c)
			starts, ends = append(starts, offset+i), append(ends, offset+i+1)
			i++
		}
	}
	return b.String(), starts, ends, true
}

func unhex(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// IsWebURL reports whether s is an absolute http or https URL with a host.
func IsWebURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return (scheme == "http" || scheme == "https") && u.Host != "" && u.Hostname() != ""
}

// Defaults returns the default acceptance state of every suggestion.
func (a *Analysis) Defaults() []bool {
	accepted := make([]bool, len(a.Suggestions))
	for i, s := range a.Suggestions {
		accepted[i] = s.Default
	}
	return accepted
}

// Available reports whether suggestion i can take effect given accepted:
// every redirect it depends on must be accepted.
func (a *Analysis) Available(i int, accepted []bool) bool {
	for dep := a.Suggestions[i].DependsOn; dep >= 0; dep = a.Suggestions[dep].DependsOn {
		if !accepted[dep] {
			return false
		}
	}
	return true
}

// Clean composes the URL that results from applying exactly the accepted
// suggestions.
func (a *Analysis) Clean(accepted []bool) string {
	return a.root.compose(accepted)
}

func (n *node) compose(accepted []bool) string {
	if n.redirect >= 0 && accepted[n.redirect] {
		return n.child.compose(accepted)
	}
	kept := make([]string, 0, len(n.segs))
	removed := false
	for _, seg := range n.segs {
		if seg.sug >= 0 && accepted[seg.sug] {
			removed = true
			continue
		}
		kept = append(kept, seg.raw)
	}
	prefix, suffix := n.prefix, n.suffix
	suffixStart := len(n.text) - len(n.suffix)
	// Remove accepted raw spans from the end backwards so offsets stay valid.
	for i := len(n.raws) - 1; i >= 0; i-- {
		r := n.raws[i]
		if !accepted[r.sug] {
			continue
		}
		removed = true
		if r.end <= len(n.prefix) {
			prefix = prefix[:r.start] + prefix[r.end:]
		} else {
			suffix = suffix[:r.start-suffixStart] + suffix[r.end-suffixStart:]
		}
	}
	if !removed {
		return n.text
	}
	var b strings.Builder
	b.WriteString(prefix)
	if len(kept) > 0 {
		b.WriteByte('?')
		b.WriteString(strings.Join(kept, "&"))
	}
	b.WriteString(suffix)
	return b.String()
}
