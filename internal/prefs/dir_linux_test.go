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
