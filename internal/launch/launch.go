// Package launch validates URLs and starts browsers with them.
package launch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
)

// ErrNotWebURL is returned for input that is not an absolute http or https URL.
var ErrNotWebURL = errors.New("only http and https links can be opened")

// Validate checks that raw is an absolute http or https URL with a host and
// returns it as received, with only these changes: surrounding whitespace
// trimmed, the scheme lower-cased, and spaces, tabs and double quotes
// percent-encoded. The result always starts with "http://" or "https://"
// and contains no whitespace or double quotes, so it can never be mistaken
// for a command-line option or split into several arguments. A malformed
// percent-escape does not cause a rejection.
func Validate(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	u, err := clean.ParseTolerant(s)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrNotWebURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.Opaque != "" {
		return "", ErrNotWebURL
	}
	s = u.Scheme + s[len(u.Scheme):]
	s = strings.NewReplacer(" ", "%20", "\t", "%09", `"`, "%22").Replace(s)
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "", ErrNotWebURL
	}
	return s, nil
}

// Start validates url and launches browser b with it, detached from bopen.
func Start(b discovery.Browser, rawURL string) error {
	u, err := Validate(rawURL)
	if err != nil {
		return err
	}
	return start(b, u)
}
