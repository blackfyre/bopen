package clearurls

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blackfyre/bopen/internal/clean"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func hashOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]) + "\n"
}

// server serves a rule list and its hash; status overrides the response code.
func server(t *testing.T, data []byte, hash string, status int, delay time.Duration) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		switch r.URL.Path {
		case "/" + hashName:
			w.Write([]byte(hash))
		case "/" + dataName:
			w.Write(data)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func fetcher(timeout time.Duration, bases ...string) Fetcher {
	return Fetcher{Client: &http.Client{Timeout: timeout}, Bases: bases, UserAgent: "bopen/test"}
}

var now = time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)

func TestUpdateSuccess(t *testing.T) {
	data := fixture(t)
	srv := server(t, data, hashOf(data), 0, 0)
	c := Cache{Dir: t.TempDir()}
	m, err := Update(context.Background(), c, fetcher(time.Second, srv.URL), now)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := c.Data()
	if string(got) != string(data) || !m.LastSuccess.Equal(now) || m.LastError != "" {
		t.Fatalf("meta %+v", m)
	}
	if saved := c.Meta(); !saved.LastSuccess.Equal(now) {
		t.Fatalf("saved meta %+v", saved)
	}
}

// failingUpdate checks a failed update leaves a previous cache in place.
func failingUpdate(t *testing.T, f Fetcher, wantErr string) {
	t.Helper()
	c := Cache{Dir: t.TempDir()}
	if err := c.saveData([]byte("previous")); err != nil {
		t.Fatal(err)
	}
	if err := c.saveMeta(Meta{LastSuccess: now.Add(-48 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	m, err := Update(context.Background(), c, f, now)
	if err == nil || !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("err = %v, want %q", err, wantErr)
	}
	if got, _ := c.Data(); string(got) != "previous" {
		t.Fatalf("cache replaced: %q", got)
	}
	saved := c.Meta()
	if saved.LastError == "" || !saved.LastSuccess.Equal(now.Add(-48*time.Hour)) || m.LastError == "" {
		t.Fatalf("meta %+v", saved)
	}
}

func TestUpdateHashMismatch(t *testing.T) {
	data := fixture(t)
	srv := server(t, data, hashOf([]byte("other")), 0, 0)
	failingUpdate(t, fetcher(time.Second, srv.URL), "SHA-256")
}

func TestUpdateOversize(t *testing.T) {
	big := make([]byte, MaxDataSize+1)
	srv := server(t, big, hashOf(big), 0, 0)
	failingUpdate(t, fetcher(5*time.Second, srv.URL), "larger than")
}

func TestUpdateTimeout(t *testing.T) {
	data := fixture(t)
	srv := server(t, data, hashOf(data), 0, 300*time.Millisecond)
	failingUpdate(t, fetcher(50*time.Millisecond, srv.URL), "Timeout")
}

func TestUpdateUnparsable(t *testing.T) {
	data := []byte(`{"providers": "nope"}`)
	srv := server(t, data, hashOf(data), 0, 0)
	failingUpdate(t, fetcher(time.Second, srv.URL), "not a valid ClearURLs file")
}

func TestUpdateMirrorFallback(t *testing.T) {
	data := fixture(t)
	down := server(t, nil, "", http.StatusServiceUnavailable, 0)
	up := server(t, data, hashOf(data), 0, 0)
	c := Cache{Dir: t.TempDir()}
	if _, err := Update(context.Background(), c, fetcher(time.Second, down.URL, up.URL), now); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.Data(); string(got) != string(data) {
		t.Fatal("cache not filled from the mirror")
	}
}

type noNetwork struct{ t *testing.T }

func (n noNetwork) RoundTrip(r *http.Request) (*http.Response, error) {
	n.t.Errorf("unexpected request to %s", r.URL)
	return nil, errors.New("no network in this test")
}

func TestNoRequestWhenDisabledOrFresh(t *testing.T) {
	f := Fetcher{Client: &http.Client{Transport: noNetwork{t}}, Bases: DefaultBases}
	c := Cache{Dir: t.TempDir()}
	if StartRefresh(false, c.Meta(), c, f, now, Update, nil) {
		t.Fatal("refresh started while disabled")
	}
	if err := c.saveMeta(Meta{LastSuccess: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if StartRefresh(true, c.Meta(), c, f, now, Update, nil) {
		t.Fatal("refresh started with a fresh cache")
	}
}

func TestStartRefreshWhenStale(t *testing.T) {
	c := Cache{Dir: t.TempDir()}
	if err := c.saveMeta(Meta{LastSuccess: now.Add(-48 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	finished := make(chan Meta, 1)
	fake := func(ctx context.Context, c Cache, f Fetcher, at time.Time) (Meta, error) {
		return Meta{LastSuccess: at}, nil
	}
	if !StartRefresh(true, c.Meta(), c, Fetcher{}, now, fake, func(m Meta, err error) { finished <- m }) {
		t.Fatal("refresh not started for a stale cache")
	}
	select {
	case m := <-finished:
		if !m.LastSuccess.Equal(now) {
			t.Fatalf("meta %+v", m)
		}
	case <-time.After(time.Second):
		t.Fatal("refresh did not finish")
	}
}

func TestMetaCorruptIsEmpty(t *testing.T) {
	c := Cache{Dir: t.TempDir()}
	if err := os.WriteFile(filepath.Join(c.Dir, metaFile), []byte("last_success = "), 0o644); err != nil {
		t.Fatal(err)
	}
	if m := c.Meta(); !m.LastSuccess.IsZero() || !m.NeedsRefresh(now) {
		t.Fatalf("meta %+v", m)
	}
}

func TestRulesFromFixture(t *testing.T) {
	rules, skipped, err := Rules(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped %d", skipped)
	}
	providers := map[string]bool{}
	for _, r := range rules {
		providers[r.Pattern.Provider.Name] = true
	}
	if len(providers) != 6 || !providers["all sites"] {
		t.Fatalf("providers %v", providers)
	}
	if last := rules[len(rules)-1]; last.Pattern.Provider.Name != "all sites" {
		t.Fatalf("global rules are not last: %+v", last)
	}
}

func TestRulesCountsSkippedPatterns(t *testing.T) {
	data := []byte(`{"providers": {"bad": {"urlPattern": ".*", "rules": ["(?<=a)b", "ok"]}}}`)
	rules, skipped, err := Rules(data)
	if err != nil || skipped != 1 || len(rules) != 1 {
		t.Fatalf("rules %d skipped %d err %v", len(rules), skipped, err)
	}
}

func analyse(t *testing.T, raw string, disabled ...string) *clean.Analysis {
	t.Helper()
	builtin, err := clean.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	extra, _, err := Rules(fixture(t))
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
	return clean.Analyse(raw, clean.Combine(clean.Enabled(builtin, off), extra))
}

func find(a *clean.Analysis, text string) (clean.Suggestion, bool) {
	for _, s := range a.Suggestions {
		if s.Text == text {
			return s, true
		}
	}
	return clean.Suggestion{}, false
}

func TestProviderSpecificParameter(t *testing.T) {
	a := analyse(t, "https://www.amazon.de/dp/B000?crid=XYZ&tag=someone-21")
	s, ok := find(a, "crid=XYZ")
	if !ok || s.Source != clean.SourceClearURLs || s.Kind != clean.KindTracking || !s.Default {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
	if s.Reason != "Listed by ClearURLs as tracking for amazon." {
		t.Fatalf("reason %q", s.Reason)
	}
	// The built-in affiliate rule wins over ClearURLs' referral marketing.
	if tag, _ := find(a, "tag=someone-21"); tag.Source != clean.SourceBuiltin {
		t.Fatalf("tag suggestion %+v", tag)
	}
}

func TestProviderException(t *testing.T) {
	a := analyse(t, "https://www.amazon.de/gp/cart?crid=XYZ")
	if _, ok := find(a, "crid=XYZ"); ok {
		t.Fatalf("exception ignored: %+v", a.Suggestions)
	}
}

func TestBuiltinWinsOverGlobalRules(t *testing.T) {
	a := analyse(t, "https://example.com/?fbclid=x")
	if len(a.Suggestions) != 1 || a.Suggestions[0].Source != clean.SourceBuiltin {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
}

func TestDisabledBuiltinDoesNotSuppressClearURLs(t *testing.T) {
	a := analyse(t, "https://example.com/?fbclid=x", "fbclid")
	if len(a.Suggestions) != 1 || a.Suggestions[0].Source != clean.SourceClearURLs ||
		a.Suggestions[0].Reason != "Listed by ClearURLs as tracking for all sites." {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
}

func TestClearURLsRedirection(t *testing.T) {
	target := "https://shop.example.com/item?id=1&fbclid=y"
	a := analyse(t, "https://go.example.net/out?to="+url.QueryEscape(target)+"&src=mail")
	if len(a.Suggestions) == 0 || a.Suggestions[0].Kind != clean.KindRedirect || a.Suggestions[0].Source != clean.SourceClearURLs {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
	if got := a.Clean(a.Defaults()); got != "https://shop.example.com/item?id=1" {
		t.Fatalf("cleaned %q", got)
	}
}

func TestGoogleRedirectionWhenBuiltinDisabled(t *testing.T) {
	raw := "https://www.google.com/url?sa=t&url=" + url.QueryEscape("https://example.com/page") + "&ved=abc&usg=def"
	a := analyse(t, raw, "redirect-google")
	if len(a.Suggestions) == 0 || a.Suggestions[0].Source != clean.SourceClearURLs || a.Suggestions[0].Target != "https://example.com/page" {
		t.Fatalf("suggestions %+v", a.Suggestions)
	}
	if got := a.Clean(a.Defaults()); got != "https://example.com/page" {
		t.Fatalf("cleaned %q", got)
	}
	if a := analyse(t, raw); a.Suggestions[0].Source != clean.SourceBuiltin {
		t.Fatalf("built-in redirect does not win: %+v", a.Suggestions[0])
	}
}

func TestRawRules(t *testing.T) {
	for _, tc := range []struct{ raw, span, cleaned string }{
		{"https://www.amazon.de/Some-Product/dp/B000/ref=sr_1_3?keywords=x", "/ref=sr_1_3", "https://www.amazon.de/Some-Product/dp/B000?keywords=x"},
		{"https://frag.example.net/article#lead-from-newsletter", "#lead-from-newsletter", "https://frag.example.net/article"},
		{"https://q.example.net/page?pc", "pc", "https://q.example.net/page"},
	} {
		a := analyse(t, tc.raw)
		s, ok := find(a, tc.span)
		if !ok || s.Source != clean.SourceClearURLs || a.URL[s.Start:s.End] != tc.span {
			t.Errorf("%s: suggestions %+v", tc.raw, a.Suggestions)
			continue
		}
		if got := a.Clean(a.Defaults()); got != tc.cleaned {
			t.Errorf("%s: cleaned %q, want %q", tc.raw, got, tc.cleaned)
		}
		if got := a.Clean(make([]bool, len(a.Suggestions))); got != tc.raw {
			t.Errorf("%s: nothing accepted gave %q", tc.raw, got)
		}
	}
}

func TestLoadWithoutCacheYieldsNothing(t *testing.T) {
	if rules := Load(Cache{Dir: t.TempDir()}); rules != nil {
		t.Fatalf("rules %d", len(rules))
	}
}

// TestLiveList checks the published list compiles completely. It needs
// network access and runs only with BOPEN_LIVE_CLEARURLS=1.
func TestLiveList(t *testing.T) {
	if os.Getenv("BOPEN_LIVE_CLEARURLS") != "1" {
		t.Skip("set BOPEN_LIVE_CLEARURLS=1 to fetch the published list")
	}
	c := Cache{Dir: t.TempDir()}
	m, err := Update(context.Background(), c, NewFetcher("test"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	rules := Load(c)
	t.Logf("%d rules, %d skipped patterns", len(rules), m.SkippedPatterns)
	if m.SkippedPatterns != 0 || len(rules) == 0 {
		t.Fatalf("skipped %d, rules %d", m.SkippedPatterns, len(rules))
	}
}
