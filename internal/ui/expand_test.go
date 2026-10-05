package ui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/blackfyre/bopen/internal/prefs"
)

func expandWindow(t *testing.T, raw string, enabled bool, result string, fail error) (*window, *int) {
	w, _ := ruleWindow(t, raw)
	calls := 0
	w.env.Expand = func(ctx context.Context, u string) (string, error) {
		calls++
		return result, fail
	}
	w.env.Config.ExpandShortLinks = enabled
	return w, &calls
}

func waitExpansion(t *testing.T, w *window) {
	t.Helper()
	select {
	case r := <-w.expansions:
		w.expansions <- r
		w.receiveExpansion()
	case <-time.After(2 * time.Second):
		t.Fatal("expansion not delivered")
	}
}

func TestExpandReplacesLinkAndPreselects(t *testing.T) {
	w, calls := expandWindow(t, "https://bit.ly/abc", true, "https://news.example.com/article?utm_source=x", nil)
	w.env.Config.Sites = []prefs.SiteRule{{ID: "s-1", Hosts: []string{"news.example.com"}, Browser: "c"}}
	if !w.m.CanExpand() {
		t.Fatal("expansion not offered")
	}
	w.startExpand()
	waitExpansion(t, w)
	if *calls != 1 || w.m.ExpandedFrom != "https://bit.ly/abc" || w.m.Analysis.URL != "https://news.example.com/article?utm_source=x" {
		t.Fatalf("calls %d from %q url %q", *calls, w.m.ExpandedFrom, w.m.Analysis.URL)
	}
	if len(w.m.Analysis.Suggestions) != 1 || w.m.Result() != "https://news.example.com/article" {
		t.Fatalf("suggestions %+v", w.m.Analysis.Suggestions)
	}
	if w.m.Site == nil || w.m.Browsers[w.m.Selected].ID != "c" || len(w.toggles) != 1 {
		t.Fatalf("site %+v selected %d", w.m.Site, w.m.Selected)
	}
}

func TestExpandFailureKeepsLink(t *testing.T) {
	w, _ := expandWindow(t, "https://t.co/xyz", true, "", errors.New("t.co answered 200 OK instead of a redirect"))
	w.startExpand()
	waitExpansion(t, w)
	if w.m.Analysis.URL != "https://t.co/xyz" || !strings.Contains(w.notice, "200 OK") || w.expanding {
		t.Fatalf("url %q notice %q", w.m.Analysis.URL, w.notice)
	}
}

func TestExpandNotOffered(t *testing.T) {
	if w, calls := expandWindow(t, "https://bit.ly/abc", false, "", nil); w.m.CanExpand() || func() bool { w.startExpand(); return *calls != 0 }() {
		t.Fatal("offered while disabled")
	}
	if w, _ := expandWindow(t, "https://example.com/abc", true, "", nil); w.m.CanExpand() {
		t.Fatal("offered for a non-shortener")
	}
}
