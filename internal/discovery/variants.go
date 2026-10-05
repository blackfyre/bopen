package discovery

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type profileKind int

const (
	firefoxProfiles profileKind = iota
	chromiumProfiles
)

// family describes where a browser keeps its profiles and how to open one,
// and on Windows its private-window flag.
type family struct {
	kind profileKind
	// ids are desktop file IDs; programs are Exec program base names.
	ids, programs []string
	// Profile directories relative to the home directory, per install kind.
	native, flatpak, snap []string
	// Windows: executable base name, profile directory under %APPDATA%
	// (roaming) or %LOCALAPPDATA% (local), and private-window flag.
	exe, roaming, local, privateFlag string
}

var families = []family{
	{kind: firefoxProfiles,
		ids:      []string{"firefox.desktop", "firefox-esr.desktop", "firefox_firefox.desktop", "org.mozilla.firefox.desktop"},
		programs: []string{"firefox", "firefox-esr"},
		native:   []string{".config/mozilla/firefox", ".mozilla/firefox"},
		flatpak:  []string{".var/app/org.mozilla.firefox/config/mozilla/firefox", ".var/app/org.mozilla.firefox/.mozilla/firefox"},
		snap:     []string{"snap/firefox/common/.config/mozilla/firefox", "snap/firefox/common/.mozilla/firefox"},
		exe:      "firefox.exe", roaming: `Mozilla\Firefox`, privateFlag: "-private-window"},
	{kind: firefoxProfiles,
		ids:      []string{"zen.desktop", "zen-browser.desktop", "app.zen_browser.zen.desktop"},
		programs: []string{"zen", "zen-browser"},
		native:   []string{".zen", ".config/zen"},
		flatpak:  []string{".var/app/app.zen_browser.zen/.zen"},
		exe:      "zen.exe", roaming: `zen`, privateFlag: "-private-window"},
	{kind: firefoxProfiles,
		ids:      []string{"librewolf.desktop", "io.gitlab.librewolf-community.desktop"},
		programs: []string{"librewolf"},
		native:   []string{".librewolf", ".config/librewolf"},
		flatpak:  []string{".var/app/io.gitlab.librewolf-community/.librewolf"},
		exe:      "librewolf.exe", roaming: `librewolf`, privateFlag: "-private-window"},
	{kind: chromiumProfiles,
		ids:      []string{"google-chrome.desktop", "com.google.Chrome.desktop"},
		programs: []string{"google-chrome", "google-chrome-stable"},
		native:   []string{".config/google-chrome"},
		flatpak:  []string{".var/app/com.google.Chrome/config/google-chrome"},
		exe:      "chrome.exe", local: `Google\Chrome\User Data`, privateFlag: "--incognito"},
	{kind: chromiumProfiles,
		ids:      []string{"chromium.desktop", "chromium-browser.desktop", "chromium_chromium.desktop", "org.chromium.Chromium.desktop"},
		programs: []string{"chromium", "chromium-browser"},
		native:   []string{".config/chromium"},
		flatpak:  []string{".var/app/org.chromium.Chromium/config/chromium"},
		snap:     []string{"snap/chromium/common/chromium"},
		exe:      "chromium.exe", local: `Chromium\User Data`, privateFlag: "--incognito"},
	{kind: chromiumProfiles,
		ids:      []string{"brave-browser.desktop", "com.brave.Browser.desktop"},
		programs: []string{"brave-browser", "brave-browser-stable", "brave"},
		native:   []string{".config/BraveSoftware/Brave-Browser"},
		flatpak:  []string{".var/app/com.brave.Browser/config/BraveSoftware/Brave-Browser"},
		exe:      "brave.exe", local: `BraveSoftware\Brave-Browser\User Data`, privateFlag: "--incognito"},
	{kind: chromiumProfiles,
		ids:      []string{"microsoft-edge.desktop", "com.microsoft.Edge.desktop"},
		programs: []string{"microsoft-edge", "microsoft-edge-stable"},
		native:   []string{".config/microsoft-edge"},
		flatpak:  []string{".var/app/com.microsoft.Edge/config/microsoft-edge"},
		exe:      "msedge.exe", local: `Microsoft\Edge\User Data`, privateFlag: "--inprivate"},
	{kind: chromiumProfiles,
		ids:      []string{"vivaldi-stable.desktop", "com.vivaldi.Vivaldi.desktop"},
		programs: []string{"vivaldi", "vivaldi-stable"},
		native:   []string{".config/vivaldi"},
		flatpak:  []string{".var/app/com.vivaldi.Vivaldi/config/vivaldi"},
		exe:      "vivaldi.exe", local: `Vivaldi\User Data`, privateFlag: "--incognito"},
}

// Paths locates the user's directories for profile discovery.
type Paths struct {
	Home, AppData, LocalAppData string
}

// SystemPaths returns the current user's directories.
func SystemPaths() Paths {
	home, _ := os.UserHomeDir()
	return Paths{Home: home, AppData: os.Getenv("APPDATA"), LocalAppData: os.Getenv("LOCALAPPDATA")}
}

type profile struct {
	key, name string
}

// Expand adds private-window support and one entry per profile for
// browsers of known families with two or more profiles.
func Expand(browsers []Browser, p Paths) []Browser {
	var out []Browser
	for _, b := range browsers {
		f, dirs := familyOf(b, p)
		if b.Entry != nil {
			for id, action := range b.Entry.Actions {
				lower := strings.ToLower(id)
				if strings.Contains(lower, "private") || strings.Contains(lower, "incognito") {
					a := action
					b.PrivateEntry = &a
					break
				}
			}
		} else if f != nil {
			b.PrivateFlag = f.privateFlag
		}
		out = append(out, b)
		if f == nil {
			continue
		}
		profiles := readProfiles(f.kind, dirs)
		if len(profiles) < 2 {
			continue
		}
		for _, pr := range profiles {
			v := b
			v.Base = b.ID
			v.ID = b.ID + "#profile=" + pr.key
			v.Name = b.Name + " · " + pr.name
			if f.kind == firefoxProfiles {
				v.ProfileArgs = []string{"-P", pr.name}
			} else {
				v.ProfileArgs = []string{"--profile-directory=" + pr.key}
			}
			out = append(out, v)
		}
	}
	sortBrowsers(out)
	return out
}

// familyOf returns the family of b and its candidate profile directories.
func familyOf(b Browser, p Paths) (*family, []string) {
	if b.Entry != nil {
		program := ""
		if args, err := b.Entry.Args("x"); err == nil && len(args) > 0 {
			program = filepath.Base(args[0])
		}
		for i := range families {
			f := &families[i]
			if !contains(f.ids, b.ID) && !contains(f.programs, program) {
				continue
			}
			rel := f.native
			switch b.Kind {
			case KindFlatpak:
				rel = f.flatpak
			case KindSnap:
				rel = f.snap
			}
			var dirs []string
			for _, r := range rel {
				dirs = append(dirs, filepath.Join(p.Home, filepath.FromSlash(r)))
			}
			return f, dirs
		}
		return nil, nil
	}
	exe := strings.ToLower(windowsExecutable(b.Command))
	for i := range families {
		f := &families[i]
		if exe != f.exe {
			continue
		}
		switch {
		case f.roaming != "" && p.AppData != "":
			return f, []string{filepath.Join(p.AppData, f.roaming)}
		case f.local != "" && p.LocalAppData != "":
			return f, []string{filepath.Join(p.LocalAppData, f.local)}
		}
		return f, nil
	}
	return nil, nil
}

// windowsExecutable returns the base name of a command's executable.
func windowsExecutable(command string) string {
	command = strings.TrimSpace(command)
	exe := command
	if strings.HasPrefix(command, `"`) {
		if end := strings.IndexByte(command[1:], '"'); end >= 0 {
			exe = command[1 : end+1]
		}
	} else if sp := strings.IndexAny(command, " \t"); sp >= 0 {
		exe = command[:sp]
	}
	if i := strings.LastIndexAny(exe, `\/`); i >= 0 {
		exe = exe[i+1:]
	}
	return exe
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// readProfiles reads the profiles from the first existing directory.
func readProfiles(kind profileKind, dirs []string) []profile {
	for _, dir := range dirs {
		var ps []profile
		var err error
		if kind == firefoxProfiles {
			ps, err = readProfilesINI(filepath.Join(dir, "profiles.ini"))
		} else {
			ps, err = readLocalState(filepath.Join(dir, "Local State"))
		}
		if err == nil {
			return ps
		}
	}
	return nil
}

// readProfilesINI lists the [ProfileN] sections of a Firefox-family
// profiles.ini, sorted by name.
func readProfilesINI(path string) ([]profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ps []profile
	var cur *profile
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[") {
			cur = nil
			if strings.HasPrefix(line, "[Profile") {
				ps = append(ps, profile{})
				cur = &ps[len(ps)-1]
			}
			continue
		}
		if cur == nil {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "Name":
			cur.name = strings.TrimSpace(v)
		case "Path":
			cur.key = strings.TrimSpace(v)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	ps = complete(ps)
	sort.Slice(ps, func(i, j int) bool { return strings.ToLower(ps[i].name) < strings.ToLower(ps[j].name) })
	return ps, nil
}

// readLocalState lists the profiles in a Chromium-family Local State file,
// sorted by display name.
func readLocalState(path string) ([]profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state struct {
		Profile struct {
			InfoCache map[string]struct {
				Name string `json:"name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	var ps []profile
	for dir, info := range state.Profile.InfoCache {
		name := info.Name
		if name == "" {
			name = dir
		}
		ps = append(ps, profile{key: dir, name: name})
	}
	ps = complete(ps)
	sort.Slice(ps, func(i, j int) bool {
		if a, b := strings.ToLower(ps[i].name), strings.ToLower(ps[j].name); a != b {
			return a < b
		}
		return ps[i].key < ps[j].key
	})
	return ps, nil
}

// complete drops profiles without a name or key.
func complete(ps []profile) []profile {
	out := ps[:0]
	for _, p := range ps {
		if p.name != "" && p.key != "" {
			out = append(out, p)
		}
	}
	return out
}
