package shortlinks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// shortener is a fake shortener that records the requests it receives.
type shortener struct {
	mu       sync.Mutex
	requests []string
}

func (s *shortener) log(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, r.Method+" "+r.URL.Path)
}

// expander treats the test server's host as the only shortener.
func expander(t *testing.T, srv *httptest.Server) Expander {
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	e := New("test")
	e.IsShortener = func(h string) bool { return h == u.Hostname() }
	return e
}

func TestChainStopsBeforeDestination(t *testing.T) {
	s := &shortener{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log(r)
		switch r.URL.Path {
		case "/a":
			http.Redirect(w, r, "/b", http.StatusMovedPermanently)
		case "/b":
			http.Redirect(w, r, "https://example.org/article?utm_source=x", http.StatusFound)
		}
	}))
	defer srv.Close()
	got, err := expander(t, srv).Expand(context.Background(), srv.URL+"/a")
	if err != nil || got != "https://example.org/article?utm_source=x" {
		t.Fatalf("got %q %v", got, err)
	}
	if strings.Join(s.requests, ",") != "HEAD /a,HEAD /b" {
		t.Fatalf("requests %v", s.requests)
	}
}

func TestHeadRefusedFallsBackToGet(t *testing.T) {
	s := &shortener{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.log(r)
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		http.Redirect(w, r, "https://example.org/", http.StatusMovedPermanently)
	}))
	defer srv.Close()
	if got, err := expander(t, srv).Expand(context.Background(), srv.URL+"/x"); err != nil || got != "https://example.org/" {
		t.Fatalf("got %q %v", got, err)
	}
	if strings.Join(s.requests, ",") != "HEAD /x,GET /x" {
		t.Fatalf("requests %v", s.requests)
	}
}

func TestFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/loop":
			http.Redirect(w, r, "/loop", http.StatusFound)
		case "/js":
			w.Header().Set("Location", "javascript:alert(1)")
			w.WriteHeader(http.StatusFound)
		case "/slow":
			time.Sleep(300 * time.Millisecond)
			http.Redirect(w, r, "https://example.org/", http.StatusFound)
		}
	}))
	defer srv.Close()
	e := expander(t, srv)
	for path, want := range map[string]string{
		"/ok":   "instead of a redirect",
		"/loop": "more than 5 redirects",
		"/js":   "invalid location",
	} {
		if _, err := e.Expand(context.Background(), srv.URL+path); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err %v, want %q", path, err, want)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := e.Expand(ctx, srv.URL+"/slow"); err == nil {
		t.Error("timeout not honoured")
	}
	if _, err := New("t").Expand(context.Background(), "https://example.com/x"); err == nil {
		t.Error("expanded a link that is not on a shortener")
	}
}

func TestIsShortener(t *testing.T) {
	for host, want := range map[string]bool{"bit.ly": true, "www.bit.ly": true, "T.CO": true, "example.com": false, "bit.ly.evil.com": false} {
		if IsShortener(host) != want {
			t.Errorf("%s: want %v", host, want)
		}
	}
}
