// Package prefs reads and writes bopen's preferences (config.toml) and
// runtime state (state.toml) in the OS user configuration directory.
package prefs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configFile = "config.toml"
	stateFile  = "state.toml"
)

// Window controls when the inspector window is shown.
type Window string

const (
	WindowAlways          Window = "always"
	WindowWhenSuggestions Window = "when-suggestions"
)

// Config holds the user preferences.
type Config struct {
	Window Window `toml:"window"`
}

// DefaultConfig returns the preferences used when none are configured.
func DefaultConfig() Config {
	return Config{Window: WindowAlways}
}

// State holds runtime state that bopen writes itself.
type State struct {
	LastUsed        string `toml:"last_used,omitempty"`
	PreviousDefault string `toml:"previous_default,omitempty"`
}

// Dir returns the bopen directory inside the OS user configuration directory.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "bopen"), nil
}

// LoadConfig reads config.toml from dir. A missing file yields the defaults
// without problems. Parse errors and invalid values are returned as problems,
// and the affected settings keep their defaults. Unknown keys are ignored.
func LoadConfig(dir string) (Config, []error) {
	cfg := DefaultConfig()
	path := filepath.Join(dir, configFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, []error{fmt.Errorf("%s: %w", path, err)}
	}
	var raw Config
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return cfg, []error{fmt.Errorf("%s: %w", path, err)}
	}
	var problems []error
	switch raw.Window {
	case "":
	case WindowAlways, WindowWhenSuggestions:
		cfg.Window = raw.Window
	default:
		problems = append(problems, fmt.Errorf("%s: invalid window value %q (want %q or %q)",
			path, raw.Window, WindowAlways, WindowWhenSuggestions))
	}
	return cfg, problems
}

// LoadState reads state.toml from dir. A missing or unreadable file yields
// an empty state.
func LoadState(dir string) State {
	var st State
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if err != nil {
		return State{}
	}
	if _, err := toml.Decode(string(data), &st); err != nil {
		return State{}
	}
	return st
}

// SaveState writes state.toml to dir atomically, creating dir if needed.
func SaveState(dir string, st State) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(st)
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(dir, stateFile), data)
}

// writeFileAtomic writes data to a temporary file next to path and renames it
// over path, so an interrupted write leaves the previous contents intact.
func writeFileAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
