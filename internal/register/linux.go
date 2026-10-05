// Package register makes bopen the default handler for web links.
package register

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blackfyre/bopen/internal/desktopentry"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
)

var schemes = []string{"x-scheme-handler/http", "x-scheme-handler/https"}

// Linux registers bopen through a desktop entry and xdg-mime.
type Linux struct {
	// DataHome is $XDG_DATA_HOME (default ~/.local/share).
	DataHome string
	// Exe is the absolute path of the bopen executable.
	Exe string
	// StateDir holds state.toml, where the previous default is recorded.
	StateDir string
	// Run runs a program and returns its standard output.
	Run func(name string, args ...string) (string, error)
	// LookPath finds a program on PATH.
	LookPath func(string) (string, error)
	// Installed reports whether a desktop file ID is an installed browser.
	Installed func(id string) bool
}

func (l Linux) desktopFile() string {
	return filepath.Join(l.DataHome, "applications", discovery.SelfDesktopID)
}

func (l Linux) xdgMime(args ...string) (string, error) {
	if _, err := l.LookPath("xdg-mime"); err != nil {
		return "", errors.New("xdg-mime not found: install xdg-utils to register bopen")
	}
	out, err := l.Run("xdg-mime", args...)
	if err != nil {
		return "", fmt.Errorf("xdg-mime %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(out), nil
}

// DesktopEntry returns the contents of bopen.desktop.
func (l Linux) DesktopEntry() string {
	return "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=bopen\n" +
		"GenericName=Link inspector\n" +
		"Comment=Inspect links, remove tracking parameters and choose a browser\n" +
		"Exec=" + desktopentry.QuoteArg(l.Exe) + " %u\n" +
		"Terminal=false\n" +
		"NoDisplay=true\n" +
		"Categories=Network;WebBrowser;\n" +
		"MimeType=" + strings.Join(schemes, ";") + ";\n"
}

// Register installs bopen.desktop, records the current https handler as the
// previous default unless it is bopen, and makes bopen the default handler.
func (l Linux) Register() (string, error) {
	if _, err := l.LookPath("xdg-mime"); err != nil {
		return "", errors.New("xdg-mime not found: install xdg-utils to register bopen")
	}
	path := l.desktopFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(l.DesktopEntry()), 0o644); err != nil {
		return "", err
	}
	current, err := l.xdgMime("query", "default", "x-scheme-handler/https")
	if err != nil {
		return "", err
	}
	st := prefs.LoadState(l.StateDir)
	if current != "" && current != discovery.SelfDesktopID {
		st.PreviousDefault = current
		if err := prefs.SaveState(l.StateDir, st); err != nil {
			return "", err
		}
	}
	if _, err := l.xdgMime(append([]string{"default", discovery.SelfDesktopID}, schemes...)...); err != nil {
		return "", err
	}
	msg := "bopen is now the default handler for http and https links."
	if st.PreviousDefault != "" {
		msg += "\nPrevious default: " + st.PreviousDefault + " (restored by 'bopen unregister')."
	}
	return msg, nil
}

// Unregister removes bopen.desktop and restores the recorded previous
// default when it is still installed.
func (l Linux) Unregister() (string, error) {
	if err := os.Remove(l.desktopFile()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	msg := "bopen has been unregistered."
	st := prefs.LoadState(l.StateDir)
	switch {
	case st.PreviousDefault == "":
		msg += "\nNo previous default was recorded; choose a default browser in your desktop settings."
	case !l.Installed(st.PreviousDefault):
		msg += "\nThe previous default " + st.PreviousDefault + " is no longer installed; choose a default browser in your desktop settings."
	default:
		if _, err := l.xdgMime(append([]string{"default", st.PreviousDefault}, schemes...)...); err != nil {
			return "", err
		}
		msg += "\nRestored " + st.PreviousDefault + " as the default handler."
		st.PreviousDefault = ""
		if err := prefs.SaveState(l.StateDir, st); err != nil {
			return "", err
		}
	}
	return msg, nil
}
