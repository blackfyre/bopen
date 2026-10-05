package desktopentry

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const testURL = "https://example.com/?a=1&b=$(id);ls"

func parse(t *testing.T, name, lang string) Entry {
	t.Helper()
	e, err := ParseFile(filepath.Join("testdata", name), lang)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestRealEntries(t *testing.T) {
	for _, tc := range []struct {
		file, name string
		args       []string
	}{
		{"org.mozilla.firefox.desktop", "Firefox", []string{"firefox", testURL}},
		{"google-chrome.desktop", "Google Chrome", []string{"/usr/bin/google-chrome-stable", testURL}},
		{"brave-browser.desktop", "Brave Web Browser", []string{"/usr/bin/brave-browser-stable", testURL}},
		{"app.zen_browser.zen.desktop", "Zen Browser", []string{"/usr/bin/flatpak", "run", "--branch=stable",
			"--arch=x86_64", "--command=launch-script.sh", "--file-forwarding", "app.zen_browser.zen", "@@u", testURL, "@@"}},
		{"firefox_firefox.desktop", "Firefox Web Browser", []string{"env",
			"BAMF_DESKTOP_FILE_HINT=/var/lib/snapd/desktop/applications/firefox_firefox.desktop", "/snap/bin/firefox", testURL}},
	} {
		e := parse(t, tc.file, "C")
		if e.Type != "Application" || e.Name != tc.name || !e.HasMimeType("x-scheme-handler/https") {
			t.Errorf("%s: unexpected entry %+v", tc.file, e)
		}
		args, err := e.Args(testURL)
		if err != nil {
			t.Errorf("%s: %v", tc.file, err)
			continue
		}
		if !reflect.DeepEqual(args, tc.args) {
			t.Errorf("%s:\n got %q\nwant %q", tc.file, args, tc.args)
		}
	}
}

func TestOnlyDesktopEntryGroupRead(t *testing.T) {
	// The [Desktop Action] groups carry their own Name and Exec keys.
	e := parse(t, "brave-browser.desktop", "C")
	if e.Exec != "/usr/bin/brave-browser-stable %U" {
		t.Fatalf("Exec = %q", e.Exec)
	}
}

func TestLocalisedName(t *testing.T) {
	for lang, want := range map[string]string{
		"en_GB.UTF-8":     "Firefox Web Browser (GB)",
		"en_GB.UTF-8@foo": "Firefox Web Browser (GB)",
		"de_DE.UTF-8":     "Firefox-Webbrowser",
		"fr_FR.UTF-8":     "Firefox Web Browser",
		"":                "Firefox Web Browser",
	} {
		if got := parse(t, "firefox_firefox.desktop", lang).Name; got != want {
			t.Errorf("lang %q: Name = %q, want %q", lang, got, want)
		}
	}
}

func TestQuotingAndFieldCodes(t *testing.T) {
	e := parse(t, "quoting.desktop", "C")
	if e.Name != "Quoting Test" {
		t.Fatalf("Name = %q", e.Name)
	}
	if !e.HasMimeType("x-scheme-handler/http;weird") || !e.HasMimeType("x-scheme-handler/https") || e.HasMimeType("x-scheme-handler/http") {
		t.Fatalf("MimeTypes = %q", e.MimeTypes)
	}
	args, err := e.Args(testURL)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/opt/My Browser/browser", "--name=Quoting Test", "--literal=%u", "--profile",
		`a "b" \c $HOME`, "--icon", "quoting-icon", "%", "--url=" + testURL}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("\n got %q\nwant %q", args, want)
	}
}

func TestNoURLFieldCodeAppends(t *testing.T) {
	e := Entry{Exec: "browser --new-window %f"}
	args, err := e.Args(testURL)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"browser", "--new-window", testURL}; !reflect.DeepEqual(args, want) {
		t.Fatalf("got %q", args)
	}
}

func TestExecErrors(t *testing.T) {
	for _, exec := range []string{"", "   ", `"/opt/browser %u`} {
		if _, err := (Entry{Exec: exec}).Args(testURL); err == nil {
			t.Errorf("Exec %q: expected error", exec)
		}
	}
}

func TestQuoteArgRoundTrip(t *testing.T) {
	path := `/home/u/my "odd" $dir/back\slash/b` + "`" + `open`
	content := "[Desktop Entry]\nType=Application\nExec=" + QuoteArg(path) + " %u\n"
	e, err := Parse(strings.NewReader(content), "C")
	if err != nil {
		t.Fatal(err)
	}
	args, err := e.Args(testURL)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{path, testURL}; !reflect.DeepEqual(args, want) {
		t.Fatalf("\n got %q\nwant %q", args, want)
	}
}

func TestParseRejectsNonDesktopEntry(t *testing.T) {
	if _, err := Parse(strings.NewReader("[Other]\nName=x\n"), "C"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDesktopActions(t *testing.T) {
	for _, tc := range []struct {
		file string
		args []string
	}{
		{"brave-browser.desktop", []string{"/usr/bin/brave-browser-stable", "--incognito", testURL}},
		{"org.mozilla.firefox.desktop", []string{"firefox", "--private-window", testURL}},
		{"app.zen_browser.zen.desktop", []string{"/usr/bin/flatpak", "run", "--branch=stable", "--arch=x86_64",
			"--command=launch-script.sh", "--file-forwarding", "app.zen_browser.zen", "--private-window", "@@u", testURL, "@@"}},
	} {
		e := parse(t, tc.file, "C")
		a, ok := e.Actions["new-private-window"]
		if !ok || a.Name == "" {
			t.Errorf("%s: actions %v", tc.file, e.Actions)
			continue
		}
		args, err := a.Args(testURL)
		if err != nil || !reflect.DeepEqual(args, tc.args) {
			t.Errorf("%s:\n got %q %v\nwant %q", tc.file, args, err, tc.args)
		}
	}
}

func TestActionsMustBeListed(t *testing.T) {
	e, err := Parse(strings.NewReader("[Desktop Entry]\nType=Application\nExec=b %u\nActions=a;\n[Desktop Action a]\nExec=b --a\n[Desktop Action unlisted]\nExec=b --x\n"), "C")
	if err != nil || len(e.Actions) != 1 || e.Actions["a"].Exec != "b --a" {
		t.Fatalf("actions %+v %v", e.Actions, err)
	}
}
