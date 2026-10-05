package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/clearurls"
	"github.com/blackfyre/bopen/internal/prefs"
)

const fixtureList = `{"providers": {"example": {"urlPattern": "^https?://news\\.example\\.com", "rules": ["ref"]}}}`

// fakeUpdate writes fixtureList into the cache (or fails) and counts calls.
type fakeUpdate struct {
	calls atomic.Int32
	fail  bool
}

func (f *fakeUpdate) update(ctx context.Context, c clearurls.Cache, _ clearurls.Fetcher, now time.Time) (clearurls.Meta, error) {
	f.calls.Add(1)
	if f.fail {
		return clearurls.Meta{LastError: "network unreachable"}, errors.New("network unreachable")
	}
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return clearurls.Meta{}, err
	}
	if err := os.WriteFile(filepath.Join(c.Dir, "clearurls.json"), []byte(fixtureList), 0o644); err != nil {
		return clearurls.Meta{}, err
	}
	return clearurls.Meta{LastSuccess: now}, nil
}

func clearURLsWindow(t *testing.T, enabled bool, meta clearurls.Meta) (*window, *fakeUpdate) {
	w, _ := ruleWindow(t, "https://news.example.com/a?ref=home")
	f := &fakeUpdate{}
	w.env.ClearURLs = &ClearURLs{Cache: clearurls.Cache{Dir: t.TempDir()}, Update: f.update, Meta: meta}
	if enabled {
		_ = w.env.Update(func(c *prefs.Config) { c.Rules.ClearURLs = true })
	}
	return w, f
}

// waitFetch waits for a background download to be delivered and applied.
func waitFetch(t *testing.T, w *window) {
	t.Helper()
	select {
	case r := <-w.fetches:
		w.fetches <- r
		w.receiveFetches()
	case <-time.After(2 * time.Second):
		t.Fatal("download not delivered")
	}
}

func TestBackgroundRefreshOnlyWhenEnabledAndStale(t *testing.T) {
	w, f := clearURLsWindow(t, false, clearurls.Meta{})
	w.startBackgroundRefresh(time.Now())
	w, f2 := clearURLsWindow(t, true, clearurls.Meta{LastSuccess: time.Now().Add(-time.Hour)})
	w.startBackgroundRefresh(time.Now())
	time.Sleep(50 * time.Millisecond)
	if f.calls.Load() != 0 || f2.calls.Load() != 0 || w.fetching {
		t.Fatalf("refresh ran: disabled %d, fresh %d", f.calls.Load(), f2.calls.Load())
	}
}

func TestBackgroundRefreshDoesNotChangeCurrentLink(t *testing.T) {
	w, f := clearURLsWindow(t, true, clearurls.Meta{LastSuccess: time.Now().Add(-48 * time.Hour)})
	before := len(w.m.Analysis.Suggestions)
	w.startBackgroundRefresh(time.Now())
	if !w.fetching {
		t.Fatal("refresh not started")
	}
	waitFetch(t, w)
	if f.calls.Load() != 1 || w.fetching || w.env.ClearURLs.Meta.LastSuccess.IsZero() {
		t.Fatalf("calls %d fetching %v meta %+v", f.calls.Load(), w.fetching, w.env.ClearURLs.Meta)
	}
	if w.env.ClearURLs.Rules != nil || len(w.m.Analysis.Suggestions) != before {
		t.Fatal("background refresh changed the current analysis")
	}
}

func TestEnablingDownloadsAndApplies(t *testing.T) {
	w, f := clearURLsWindow(t, false, clearurls.Meta{})
	w.openSettings()
	w.setClearURLs(true)
	if cfg, _ := prefs.LoadConfig(w.env.ConfigDir); !cfg.Rules.ClearURLs {
		t.Fatal("preference not saved")
	}
	if !strings.Contains(strings.Join(w.clearURLsStatus(), " "), "Downloading") {
		t.Fatalf("status %v", w.clearURLsStatus())
	}
	waitFetch(t, w)
	if f.calls.Load() != 1 || len(w.env.ClearURLs.Rules) != 1 {
		t.Fatalf("calls %d rules %d", f.calls.Load(), len(w.env.ClearURLs.Rules))
	}
	if !strings.Contains(strings.Join(w.clearURLsStatus(), " "), "Last updated") {
		t.Fatalf("status %v", w.clearURLsStatus())
	}
	w.closeSettings()
	s := w.m.Analysis.Suggestions
	if len(s) != 1 || s[0].Source != clean.SourceClearURLs || sourceLabel(s[0].Source) != "ClearURLs list" {
		t.Fatalf("suggestions %+v", s)
	}
}

func TestUpdateNowFailureIsShown(t *testing.T) {
	w, f := clearURLsWindow(t, true, clearurls.Meta{LastSuccess: time.Now()})
	f.fail = true
	w.openSettings()
	w.startFetch()
	waitFetch(t, w)
	status := strings.Join(w.clearURLsStatus(), " ")
	if f.calls.Load() != 1 || !strings.Contains(status, "network unreachable") {
		t.Fatalf("status %q", status)
	}
}

func TestDisablingStopsUsingTheList(t *testing.T) {
	w, _ := clearURLsWindow(t, false, clearurls.Meta{})
	w.openSettings()
	w.setClearURLs(true)
	waitFetch(t, w)
	w.setClearURLs(false)
	w.closeSettings()
	if len(w.m.Analysis.Suggestions) != 0 {
		t.Fatalf("list still applied: %+v", w.m.Analysis.Suggestions)
	}
}
