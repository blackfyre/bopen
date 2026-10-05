package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/blackfyre/bopen/internal/desktopentry"
)

func put(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const twoZenProfiles = `[Profile1]
Name=Default Profile
IsRelative=1
Path=zi88nqsi.Default Profile
Default=1

[Profile0]
Name=Default (release)
IsRelative=1
Path=79dznv8h.Default (release)

[General]
StartWithLastProfile=1

[Install2953CB39A2589173]
Default=79dznv8h.Default (release)
`

func entry(id, exec string, actions map[string]desktopentry.Entry) *desktopentry.Entry {
	return &desktopentry.Entry{Name: id, Exec: exec, Actions: actions}
}

func names(bs []Browser) []string {
	var out []string
	for _, b := range bs {
		out = append(out, b.Name)
	}
	return out
}

func TestExpandFlatpakZenProfiles(t *testing.T) {
	home := t.TempDir()
	put(t, filepath.Join(home, ".var/app/app.zen_browser.zen/.zen/profiles.ini"), twoZenProfiles)
	private := desktopentry.Entry{Exec: "flatpak run app.zen_browser.zen --private-window @@u %u @@"}
	zen := Browser{ID: "app.zen_browser.zen.desktop", Name: "Zen Browser", Kind: KindFlatpak,
		Entry: entry("zen", "/usr/bin/flatpak run app.zen_browser.zen @@u %u @@", map[string]desktopentry.Entry{"new-private-window": private})}
	got := Expand([]Browser{zen}, Paths{Home: home})
	want := []string{"Zen Browser", "Zen Browser · Default (release)", "Zen Browser · Default Profile"}
	if !reflect.DeepEqual(names(got), want) {
		t.Fatalf("got %v", names(got))
	}
	release := got[1]
	if release.ID != "app.zen_browser.zen.desktop#profile=79dznv8h.Default (release)" || release.Base != zen.ID ||
		!reflect.DeepEqual(release.ProfileArgs, []string{"-P", "Default (release)"}) || release.Kind != KindFlatpak {
		t.Fatalf("profile entry %+v", release)
	}
	for _, b := range got {
		if !b.SupportsPrivate() {
			t.Errorf("%s lost private support", b.Name)
		}
	}
}

func TestExpandSingleChromeProfile(t *testing.T) {
	home := t.TempDir()
	put(t, filepath.Join(home, ".config/google-chrome/Local State"), `{"profile":{"info_cache":{"Default":{"name":"Your Chrome"}}}}`)
	chrome := Browser{ID: "google-chrome.desktop", Name: "Google Chrome", Kind: KindSystem, Entry: entry("chrome", "/usr/bin/google-chrome-stable %U", nil)}
	if got := Expand([]Browser{chrome}, Paths{Home: home}); len(got) != 1 || got[0].SupportsPrivate() {
		t.Fatalf("got %+v", got)
	}
}

func TestExpandChromiumProfilesByProgram(t *testing.T) {
	home := t.TempDir()
	put(t, filepath.Join(home, ".config/BraveSoftware/Brave-Browser/Local State"),
		`{"profile":{"info_cache":{"Default":{"name":"Personal"},"Profile 1":{"name":"Work"}}}}`)
	// A renamed desktop file is still recognised by its program.
	brave := Browser{ID: "my-brave.desktop", Name: "Brave", Kind: KindUser, Entry: entry("brave", "/usr/bin/brave-browser-stable %U", nil)}
	got := Expand([]Browser{brave}, Paths{Home: home})
	if !reflect.DeepEqual(names(got), []string{"Brave", "Brave · Personal", "Brave · Work"}) {
		t.Fatalf("got %v", names(got))
	}
	if !reflect.DeepEqual(got[2].ProfileArgs, []string{"--profile-directory=Profile 1"}) {
		t.Fatalf("args %q", got[2].ProfileArgs)
	}
}

func TestExpandSnapFirefoxUsesSnapPath(t *testing.T) {
	home := t.TempDir()
	put(t, filepath.Join(home, "snap/firefox/common/.mozilla/firefox/profiles.ini"),
		"[Profile0]\nName=a\nPath=x.a\n[Profile1]\nName=b\nPath=y.b\n")
	// A native profile store must not be used for the snap.
	put(t, filepath.Join(home, ".mozilla/firefox/profiles.ini"), "[Profile0]\nName=native\nPath=n\n[Profile1]\nName=other\nPath=o\n")
	ff := Browser{ID: "firefox_firefox.desktop", Name: "Firefox", Kind: KindSnap, Entry: entry("ff", "/snap/bin/firefox %u", nil)}
	if got := names(Expand([]Browser{ff}, Paths{Home: home})); !reflect.DeepEqual(got, []string{"Firefox", "Firefox · a", "Firefox · b"}) {
		t.Fatalf("got %v", got)
	}
}

func TestExpandWindows(t *testing.T) {
	local, roaming := t.TempDir(), t.TempDir()
	put(t, filepath.Join(local, `Microsoft\Edge\User Data`, "Local State"),
		`{"profile":{"info_cache":{"Default":{"name":"Personal"},"Profile 2":{"name":"Work"}}}}`)
	edge := Browser{ID: "Microsoft Edge", Name: "Microsoft Edge", Kind: KindSystem,
		Command: `"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe" --single-argument %1`}
	ff := Browser{ID: "Firefox-308046B0AF4A39CB", Name: "Firefox", Kind: KindSystem, Command: `"C:\Program Files\Mozilla Firefox\firefox.exe" -osint -url "%1"`}
	unknown := Browser{ID: "Other", Name: "Other", Command: `"C:\Other\other.exe" "%1"`}
	got := Expand([]Browser{edge, ff, unknown}, Paths{AppData: roaming, LocalAppData: local})
	if !reflect.DeepEqual(names(got), []string{"Firefox", "Microsoft Edge", "Microsoft Edge · Personal", "Microsoft Edge · Work", "Other"}) {
		t.Fatalf("got %v", names(got))
	}
	if got[0].PrivateFlag != "-private-window" || got[1].PrivateFlag != "--inprivate" || got[4].SupportsPrivate() {
		t.Fatalf("flags %q %q %v", got[0].PrivateFlag, got[1].PrivateFlag, got[4].SupportsPrivate())
	}
}

func TestExpandIgnoresMalformedStores(t *testing.T) {
	home := t.TempDir()
	put(t, filepath.Join(home, ".config/google-chrome/Local State"), "{not json")
	put(t, filepath.Join(home, ".config/mozilla/firefox/profiles.ini"), "[Profile0]\nName=only-name\n")
	bs := []Browser{
		{ID: "google-chrome.desktop", Name: "Chrome", Entry: entry("c", "google-chrome %U", nil)},
		{ID: "firefox.desktop", Name: "Firefox", Entry: entry("f", "firefox %u", nil)},
	}
	if got := Expand(bs, Paths{Home: home}); len(got) != 2 {
		t.Fatalf("got %v", names(got))
	}
}
