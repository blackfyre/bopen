package discovery

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func browserEntry(name, extra string) string {
	return "[Desktop Entry]\nType=Application\nName=" + name + "\nExec=" + name + " %u\n" +
		"MimeType=text/html;x-scheme-handler/http;x-scheme-handler/https;\n" + extra
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func noLookPath(string) (string, error) { return "", errors.New("not found") }

type tree struct {
	user, userFlatpak, system, flatpak, snap string
}

func newTree(t *testing.T) (tree, []Root) {
	base := t.TempDir()
	tr := tree{
		user:        filepath.Join(base, "home/.local/share/applications"),
		userFlatpak: filepath.Join(base, "home/.local/share/flatpak/exports/share/applications"),
		system:      filepath.Join(base, "usr/share/applications"),
		flatpak:     filepath.Join(base, "var/lib/flatpak/exports/share/applications"),
		snap:        filepath.Join(base, "var/lib/snapd/desktop/applications"),
	}
	return tr, []Root{
		{tr.user, KindUser}, {tr.userFlatpak, KindFlatpak}, {tr.system, KindSystem},
		{tr.flatpak, KindFlatpak}, {tr.snap, KindSnap},
	}
}

func ids(bs []Browser) []string {
	out := []string{}
	for _, b := range bs {
		out = append(out, b.ID+"/"+string(b.Kind))
	}
	return out
}

func TestLinuxDiscovery(t *testing.T) {
	tr, roots := newTree(t)
	write(t, filepath.Join(tr.system, "brave-browser.desktop"), browserEntry("Brave", ""))
	write(t, filepath.Join(tr.flatpak, "app.zen_browser.zen.desktop"), browserEntry("Zen", ""))
	write(t, filepath.Join(tr.snap, "firefox_firefox.desktop"), browserEntry("Firefox Snap", ""))
	write(t, filepath.Join(tr.userFlatpak, "org.mozilla.firefox.desktop"), browserEntry("firefox", ""))
	write(t, filepath.Join(tr.system, "firefox.desktop"), browserEntry("Firefox", ""))
	write(t, filepath.Join(tr.system, "vendor/sub.desktop"), browserEntry("Vendor", ""))
	write(t, filepath.Join(tr.system, "org.gnome.TextEditor.desktop"),
		"[Desktop Entry]\nType=Application\nName=Text\nExec=te %U\nMimeType=text/plain;\n")
	write(t, filepath.Join(tr.system, "hidden.desktop"), browserEntry("Hidden", "Hidden=true\n"))
	write(t, filepath.Join(tr.system, "nodisplay.desktop"), browserEntry("NoDisplay", "NoDisplay=true\n"))
	write(t, filepath.Join(tr.system, "link.desktop"), "[Desktop Entry]\nType=Link\nName=L\nURL=x\nMimeType=x-scheme-handler/https;\n")
	write(t, filepath.Join(tr.system, "old.desktop"), browserEntry("Old", "TryExec=/nonexistent/old-browser\n"))
	write(t, filepath.Join(tr.system, "missing.desktop"), browserEntry("Missing", "TryExec=no-such-browser\n"))
	write(t, filepath.Join(tr.system, "bopen.desktop"), browserEntry("bopen", ""))
	write(t, filepath.Join(tr.system, "notes.txt"), "not a desktop entry")

	got := ids(Linux(roots, "C", noLookPath))
	want := []string{
		"brave-browser.desktop/system",
		"firefox.desktop/system",
		"org.mozilla.firefox.desktop/flatpak",
		"firefox_firefox.desktop/snap",
		"vendor-sub.desktop/system",
		"app.zen_browser.zen.desktop/flatpak",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\n got %v\nwant %v", got, want)
	}
}

func TestLinuxUserOverrideShadows(t *testing.T) {
	tr, roots := newTree(t)
	write(t, filepath.Join(tr.user, "firefox.desktop"), browserEntry("Firefox (mine)", ""))
	write(t, filepath.Join(tr.system, "firefox.desktop"), browserEntry("Firefox", ""))
	bs := Linux(roots, "C", noLookPath)
	if len(bs) != 1 || bs[0].Name != "Firefox (mine)" || bs[0].Kind != KindUser {
		t.Fatalf("got %+v", bs)
	}
}

func TestLinuxUserOverrideHides(t *testing.T) {
	tr, roots := newTree(t)
	write(t, filepath.Join(tr.user, "firefox.desktop"), browserEntry("Firefox", "Hidden=true\n"))
	write(t, filepath.Join(tr.system, "firefox.desktop"), browserEntry("Firefox", ""))
	if bs := Linux(roots, "C", noLookPath); len(bs) != 0 {
		t.Fatalf("got %+v", bs)
	}
}

func TestLinuxTryExecFound(t *testing.T) {
	tr, roots := newTree(t)
	write(t, filepath.Join(tr.system, "a.desktop"), browserEntry("A", "TryExec=a-browser\n"))
	found := func(name string) (string, error) { return "/usr/bin/" + name, nil }
	if bs := Linux(roots, "C", found); len(bs) != 1 {
		t.Fatalf("got %+v", bs)
	}
}

func TestLinuxMissingRootsIgnored(t *testing.T) {
	if bs := Linux([]Root{{"/nonexistent/applications", KindSystem}}, "C", noLookPath); len(bs) != 0 {
		t.Fatalf("got %+v", bs)
	}
}

func TestLinuxRoots(t *testing.T) {
	env := map[string]string{
		"XDG_DATA_DIRS": "/home/u/.local/share/flatpak/exports/share:/var/lib/flatpak/exports/share:/usr/local/share:/usr/share",
	}
	got := LinuxRoots(func(k string) string { return env[k] }, "/home/u")
	want := []Root{
		{"/home/u/.local/share/applications", KindUser},
		{"/home/u/.local/share/flatpak/exports/share/applications", KindFlatpak},
		{"/var/lib/flatpak/exports/share/applications", KindFlatpak},
		{"/usr/local/share/applications", KindSystem},
		{"/usr/share/applications", KindSystem},
		{"/var/lib/snapd/desktop/applications", KindSnap},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\n got %v\nwant %v", got, want)
	}
}

func TestLinuxRootsDefaults(t *testing.T) {
	got := LinuxRoots(func(string) string { return "" }, "/home/u")
	want := []Root{
		{"/home/u/.local/share/applications", KindUser},
		{"/home/u/.local/share/flatpak/exports/share/applications", KindFlatpak},
		{"/usr/local/share/applications", KindSystem},
		{"/usr/share/applications", KindSystem},
		{"/var/lib/flatpak/exports/share/applications", KindFlatpak},
		{"/var/lib/snapd/desktop/applications", KindSnap},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\n got %v\nwant %v", got, want)
	}
}
