// Package clearurls downloads, verifies and caches the ClearURLs rule list
// (https://github.com/ClearURLs/Rules, LGPL-3.0) and turns it into bopen
// rules. The list is fetched on demand and never bundled with bopen.
package clearurls

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/blackfyre/bopen/internal/prefs"
)

const (
	dataFile = "clearurls.json"
	metaFile = "clearurls.meta.toml"
	// RefreshAfter is how old the cache may get before a background refresh.
	RefreshAfter = 24 * time.Hour
)

// Meta records the outcome of the last downloads.
type Meta struct {
	LastSuccess     time.Time `toml:"last_success,omitempty"`
	LastError       string    `toml:"last_error,omitempty"`
	SkippedPatterns int       `toml:"skipped_patterns,omitempty"`
}

// Cache is the directory holding the downloaded list and its metadata.
type Cache struct {
	Dir string
}

// SystemCache returns the cache in the OS user cache directory.
func SystemCache() (Cache, error) {
	dir, err := prefs.CacheDir()
	return Cache{Dir: dir}, err
}

// Data returns the cached rule list, or an error when there is none.
func (c Cache) Data() ([]byte, error) {
	return os.ReadFile(filepath.Join(c.Dir, dataFile))
}

// Meta returns the cache metadata; missing or unreadable metadata is empty.
func (c Cache) Meta() Meta {
	var m Meta
	data, err := os.ReadFile(filepath.Join(c.Dir, metaFile))
	if err != nil {
		return Meta{}
	}
	if _, err := toml.Decode(string(data), &m); err != nil {
		return Meta{}
	}
	return m
}

func (c Cache) saveMeta(m Meta) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(m)
	if err != nil {
		return err
	}
	return prefs.WriteFileAtomic(filepath.Join(c.Dir, metaFile), data)
}

func (c Cache) saveData(data []byte) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return err
	}
	return prefs.WriteFileAtomic(filepath.Join(c.Dir, dataFile), data)
}

// NeedsRefresh reports whether the cache is missing or older than
// RefreshAfter at now.
func (m Meta) NeedsRefresh(now time.Time) bool {
	return m.LastSuccess.IsZero() || now.Sub(m.LastSuccess) >= RefreshAfter
}
