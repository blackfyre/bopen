package clearurls

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	dataName = "data.minify.json"
	hashName = "rules.minify.hash"
	// MaxDataSize is the largest rule list accepted.
	MaxDataSize = 5 << 20
	// Timeout bounds one whole download attempt.
	Timeout = 15 * time.Second
)

// DefaultBases are the official ClearURLs rule mirrors, in order of preference.
var DefaultBases = []string{"https://rules2.clearurls.xyz", "https://rules1.clearurls.xyz"}

// Fetcher downloads and verifies the rule list.
type Fetcher struct {
	Client    *http.Client
	Bases     []string
	UserAgent string
}

// NewFetcher returns a fetcher for the official mirrors using the
// environment's proxy settings.
func NewFetcher(version string) Fetcher {
	return Fetcher{
		Client:    &http.Client{Timeout: Timeout, Transport: http.DefaultTransport},
		Bases:     DefaultBases,
		UserAgent: "bopen/" + version,
	}
}

// Fetch downloads the rule list from the first mirror that serves a list
// matching its published SHA-256 hash and parses as a ClearURLs file.
func (f Fetcher) Fetch(ctx context.Context) ([]byte, error) {
	var errs []error
	for _, base := range f.Bases {
		data, err := f.fetchFrom(ctx, base)
		if err == nil {
			return data, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", base, err))
	}
	return nil, errors.Join(errs...)
}

func (f Fetcher) fetchFrom(ctx context.Context, base string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	hash, err := f.get(ctx, base+"/"+hashName, 1024)
	if err != nil {
		return nil, err
	}
	data, err := f.get(ctx, base+"/"+dataName, MaxDataSize)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if want := strings.ToLower(strings.TrimSpace(string(hash))); hex.EncodeToString(sum[:]) != want {
		return nil, errors.New("rule list does not match its published SHA-256 hash")
	}
	if _, _, err := Rules(data); err != nil {
		return nil, fmt.Errorf("rule list is not a valid ClearURLs file: %w", err)
	}
	return data, nil
}

func (f Fetcher) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if f.UserAgent != "" {
		req.Header.Set("User-Agent", f.UserAgent)
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s: larger than %d bytes", url, limit)
	}
	return data, nil
}

// Update fetches the rule list into the cache and records the outcome in
// its metadata. On failure the cached list is left untouched.
func Update(ctx context.Context, c Cache, f Fetcher, now time.Time) (Meta, error) {
	m := c.Meta()
	data, err := f.Fetch(ctx)
	if err == nil {
		err = c.saveData(data)
	}
	if err != nil {
		m.LastError = err.Error()
		if saveErr := c.saveMeta(m); saveErr != nil {
			return m, errors.Join(err, saveErr)
		}
		return m, err
	}
	_, skipped, _ := Rules(data)
	m = Meta{LastSuccess: now, SkippedPatterns: skipped}
	return m, c.saveMeta(m)
}
