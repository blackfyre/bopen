// Package shortlinks resolves shortened links by asking the shortener where
// they lead, contacting nothing but known shorteners.
package shortlinks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Hosts are the known URL shorteners. Only these are ever contacted.
var Hosts = map[string]bool{
	"bit.ly": true, "bitly.com": true, "t.co": true, "tinyurl.com": true, "tinyurl.is": true,
	"lnkd.in": true, "ow.ly": true, "buff.ly": true, "is.gd": true, "v.gd": true,
	"rebrand.ly": true, "cutt.ly": true, "rb.gy": true, "t.ly": true, "shorturl.at": true,
	"tiny.cc": true, "s.id": true, "amzn.to": true, "amzn.eu": true, "a.co": true,
	"aka.ms": true, "trib.al": true, "dlvr.it": true, "ift.tt": true, "fb.me": true,
	"spoti.fi": true, "apple.co": true, "bl.ink": true, "short.io": true, "qrco.de": true,
}

// IsShortener reports whether host (with or without "www.") is a known
// URL shortener.
func IsShortener(host string) bool {
	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	return Hosts[host]
}

const (
	// MaxHops is the most redirects followed.
	MaxHops = 5
	// Timeout bounds a whole expansion.
	Timeout = 5 * time.Second
)

// Expander follows shortener redirects.
type Expander struct {
	Client *http.Client
	// IsShortener decides which hosts may be contacted.
	IsShortener func(host string) bool
	UserAgent   string
}

// New returns an expander for the known shorteners.
func New(version string) Expander {
	return Expander{
		Client: &http.Client{
			Transport: http.DefaultTransport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		IsShortener: IsShortener,
		UserAgent:   "bopen/" + version,
	}
}

// Expand returns where the short link raw leads: the first location that
// is not on a known shortener. It contacts only shorteners.
func (e Expander) Expand(ctx context.Context, raw string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	cur, err := url.Parse(raw)
	if err != nil || !web(cur) {
		return "", errors.New("not a web link")
	}
	if !e.IsShortener(cur.Hostname()) {
		return "", fmt.Errorf("%s is not a known URL shortener", cur.Hostname())
	}
	for hop := 0; hop < MaxHops; hop++ {
		loc, err := e.next(ctx, cur)
		if err != nil {
			return "", err
		}
		if !e.IsShortener(loc.Hostname()) {
			return loc.String(), nil
		}
		cur = loc
	}
	return "", fmt.Errorf("more than %d redirects", MaxHops)
}

// next asks one shortener for its redirect target.
func (e Expander) next(ctx context.Context, cur *url.URL) (*url.URL, error) {
	resp, err := e.do(ctx, http.MethodHead, cur)
	if err == nil && (resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented) {
		resp.Body.Close()
		resp, err = e.do(ctx, http.MethodGet, cur)
	}
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return nil, fmt.Errorf("%s answered %s instead of a redirect", cur.Hostname(), resp.Status)
	}
	loc, err := cur.Parse(resp.Header.Get("Location"))
	if err != nil || resp.Header.Get("Location") == "" || !web(loc) {
		return nil, fmt.Errorf("%s redirected to an invalid location", cur.Hostname())
	}
	return loc, nil
}

func (e Expander) do(ctx context.Context, method string, u *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if e.UserAgent != "" {
		req.Header.Set("User-Agent", e.UserAgent)
	}
	return e.Client.Do(req)
}

func web(u *url.URL) bool {
	return (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}
