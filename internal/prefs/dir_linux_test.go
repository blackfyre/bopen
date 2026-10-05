package prefs

import (
	"path/filepath"
	"testing"
)

func TestDirDefaultsToDotConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	got, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "bopen"); got != want {
		t.Fatalf("Dir() = %q, want %q", got, want)
	}
}

func TestCacheDirDefaultsToDotCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", "")
	got, err := CacheDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".cache", "bopen"); got != want {
		t.Fatalf("CacheDir() = %q, want %q", got, want)
	}
}
