package clean

import (
	"net/url"
	"testing"
)

func TestCleanPreservesRemaining(t *testing.T) {
	a := analyse(t, "https://example.com/a?b=1&utm_source=x&c=%20d#top")
	if got := a.Clean(a.Defaults()); got != "https://example.com/a?b=1&c=%20d#top" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanRejectedSuggestionKept(t *testing.T) {
	raw := "https://example.com/a?utm_source=x"
	a := analyse(t, raw)
	if got := a.Clean([]bool{false}); got != raw {
		t.Fatalf("got %q", got)
	}
}

func TestCleanEmptyQueryRemoved(t *testing.T) {
	a := analyse(t, "https://example.com/a?fbclid=x")
	if got := a.Clean(a.Defaults()); got != "https://example.com/a" {
		t.Fatalf("got %q", got)
	}
	a = analyse(t, "https://example.com/a?fbclid=x#frag")
	if got := a.Clean(a.Defaults()); got != "https://example.com/a#frag" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanAffiliateKeptByDefault(t *testing.T) {
	raw := "https://www.amazon.de/dp/B000?tag=someone-21"
	a := analyse(t, raw)
	if got := a.Clean(a.Defaults()); got != raw {
		t.Fatalf("got %q", got)
	}
	if got := a.Clean([]bool{true}); got != "https://www.amazon.de/dp/B000" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanRejectedRedirectKeepsWrapper(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/?fbclid=x") + "&utm_source=y"
	a := analyse(t, raw)
	// redirect, wrapper's own utm_source, then the dependent fbclid.
	accepted := a.Defaults()
	redirect := -1
	for i, s := range a.Suggestions {
		if s.Kind == KindRedirect {
			redirect = i
		}
	}
	accepted[redirect] = false
	for i, s := range a.Suggestions {
		if s.DependsOn == redirect && a.Available(i, accepted) {
			t.Fatalf("dependent suggestion %d still available", i)
		}
	}
	want := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/?fbclid=x")
	if got := a.Clean(accepted); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAvailableWhenRedirectAccepted(t *testing.T) {
	raw := "https://www.google.com/url?q=" + url.QueryEscape("https://shop.example.com/?fbclid=x")
	a := analyse(t, raw)
	accepted := a.Defaults()
	for i := range a.Suggestions {
		if !a.Available(i, accepted) {
			t.Fatalf("suggestion %d unavailable with defaults", i)
		}
	}
}
