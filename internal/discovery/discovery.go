// Package discovery lists the web browsers installed on the machine.
package discovery

import (
	"sort"
	"strings"

	"github.com/blackfyre/bopen/internal/desktopentry"
)

// Kind says how a browser is installed.
type Kind string

const (
	KindSystem  Kind = "system"
	KindUser    Kind = "user"
	KindFlatpak Kind = "flatpak"
	KindSnap    Kind = "snap"
)

// SelfID is bopen's own identity on every platform: the desktop file ID
// "bopen.desktop" on Linux and the StartMenuInternet key "bopen" on Windows.
const (
	SelfDesktopID = "bopen.desktop"
	SelfClientKey = "bopen"
)

// Browser is one discovered browser.
type Browser struct {
	// ID is stable across runs: the desktop file ID on Linux, the
	// StartMenuInternet key name on Windows.
	ID   string
	Name string
	Kind Kind
	// Entry is the desktop entry on Linux.
	Entry *desktopentry.Entry
	// Command is the registered open command on Windows.
	Command string
}

func sortBrowsers(bs []Browser) {
	sort.SliceStable(bs, func(i, j int) bool {
		a, b := strings.ToLower(bs[i].Name), strings.ToLower(bs[j].Name)
		if a != b {
			return a < b
		}
		return bs[i].ID < bs[j].ID
	})
}
