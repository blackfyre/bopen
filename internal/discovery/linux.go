package discovery

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blackfyre/bopen/internal/desktopentry"
)

// Root is a directory of desktop entries and the install kind of its entries.
type Root struct {
	Dir  string
	Kind Kind
}

// LinuxRoots returns the desktop entry directories in precedence order for
// the given environment lookup and home directory.
func LinuxRoots(getenv func(string) string, home string) []Root {
	dataHome := getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}
	dataDirs := getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = "/usr/local/share:/usr/share"
	}
	roots := []Root{
		{filepath.Join(dataHome, "applications"), KindUser},
		{filepath.Join(home, ".local", "share", "flatpak", "exports", "share", "applications"), KindFlatpak},
	}
	for _, d := range filepath.SplitList(dataDirs) {
		if d != "" {
			dir := filepath.Join(d, "applications")
			roots = append(roots, Root{dir, kindForDir(dir)})
		}
	}
	roots = append(roots,
		Root{"/var/lib/flatpak/exports/share/applications", KindFlatpak},
		Root{"/var/lib/snapd/desktop/applications", KindSnap},
	)
	// Keep the first occurrence of each directory.
	seen := map[string]bool{}
	out := roots[:0]
	for _, r := range roots {
		clean := filepath.Clean(r.Dir)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		r.Dir = clean
		out = append(out, r)
	}
	return out
}

func kindForDir(dir string) Kind {
	switch {
	case strings.Contains(dir, "/flatpak/exports/"):
		return KindFlatpak
	case strings.Contains(dir, "/snapd/desktop/"):
		return KindSnap
	}
	return KindSystem
}

// Linux discovers browsers from desktop entries under roots. lookPath
// resolves TryExec programs that are not absolute paths.
func Linux(roots []Root, lang string, lookPath func(string) (string, error)) []Browser {
	seen := map[string]bool{}
	var browsers []Browser
	for _, root := range roots {
		_ = filepath.WalkDir(root.Dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == root.Dir {
					return fs.SkipDir
				}
				return nil
			}
			if d.IsDir() || !strings.HasSuffix(path, ".desktop") {
				return nil
			}
			rel, err := filepath.Rel(root.Dir, path)
			if err != nil {
				return nil
			}
			id := strings.ReplaceAll(filepath.ToSlash(rel), "/", "-")
			if seen[id] {
				return nil
			}
			entry, err := desktopentry.ParseFile(path, lang)
			if err != nil {
				return nil
			}
			seen[id] = true
			if id == SelfDesktopID || !qualifies(entry, lookPath) {
				return nil
			}
			browsers = append(browsers, Browser{ID: id, Name: entry.Name, Kind: root.Kind, Entry: &entry})
			return nil
		})
	}
	sortBrowsers(browsers)
	return browsers
}

func qualifies(e desktopentry.Entry, lookPath func(string) (string, error)) bool {
	if e.Type != "Application" || e.Hidden || e.NoDisplay || e.Exec == "" {
		return false
	}
	if !e.HasMimeType("x-scheme-handler/http") && !e.HasMimeType("x-scheme-handler/https") {
		return false
	}
	if e.TryExec != "" {
		if filepath.IsAbs(e.TryExec) {
			if _, err := os.Stat(e.TryExec); err != nil {
				return false
			}
		} else if _, err := lookPath(e.TryExec); err != nil {
			return false
		}
	}
	return true
}

// DiscoverLinux discovers browsers using the current environment.
func DiscoverLinux() []Browser {
	home, _ := os.UserHomeDir()
	lang := os.Getenv("LC_ALL")
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	return Linux(LinuxRoots(os.Getenv, home), lang, exec.LookPath)
}
