//go:build linux

package launch

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/blackfyre/bopen/internal/desktopentry"
	"github.com/blackfyre/bopen/internal/discovery"
)

func TestLinuxArgs(t *testing.T) {
	for _, tc := range []struct {
		exec string
		want []string
	}{
		{"/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=launch-script.sh --file-forwarding app.zen_browser.zen @@u %u @@",
			[]string{"/usr/bin/flatpak", "run", "--branch=stable", "--arch=x86_64", "--command=launch-script.sh",
				"--file-forwarding", "app.zen_browser.zen", "@@u", "https://example.com/?a=$(id)&b=;ls", "@@"}},
		{"firefox", []string{"firefox", "https://example.com/?a=$(id)&b=;ls"}},
	} {
		b := discovery.Browser{Entry: &desktopentry.Entry{Exec: tc.exec}}
		got, err := LinuxArgs(b, "https://example.com/?a=$(id)&b=;ls")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("\n got %q\nwant %q", got, tc.want)
		}
	}
}

func TestStartIsDetachedAndPassesURLLiterally(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	script := filepath.Join(dir, "browser.sh")
	content := "#!/bin/sh\nprintf '%s|' \"$#\" \"$1\" > " + marker + ".tmp\nmv " + marker + ".tmp " + marker + "\nsleep 5\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	b := discovery.Browser{Entry: &desktopentry.Entry{Exec: desktopentry.QuoteArg(script) + " %u"}}
	url := "https://example.com/?a=$(touch%20pwned)&b=;ls"

	begin := time.Now()
	if err := Start(b, url); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(begin); elapsed > time.Second {
		t.Fatalf("Start waited %v for the browser", elapsed)
	}
	var data []byte
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if d, err := os.ReadFile(marker); err == nil {
			data = d
			break
		}
	}
	if got, want := string(data), "1|"+url+"|"; got != want {
		t.Fatalf("browser received %q, want %q", got, want)
	}
	if _, err := os.Stat("pwned"); err == nil || strings.Contains(string(data), "pwned\n") {
		t.Fatal("URL was interpreted by a shell")
	}
}

func TestStartRejectsInvalidURL(t *testing.T) {
	b := discovery.Browser{Entry: &desktopentry.Entry{Exec: "/bin/true %u"}}
	if err := Start(b, "--gpu-launcher=calc"); err == nil {
		t.Fatal("expected error")
	}
}

func TestStartMissingProgram(t *testing.T) {
	b := discovery.Browser{Entry: &desktopentry.Entry{Exec: "/nonexistent/browser %u"}}
	if err := Start(b, "https://example.com/"); err == nil {
		t.Fatal("expected error")
	}
}

func TestLinuxArgsPrivateAndProfile(t *testing.T) {
	url := "https://example.com/"
	zen := discovery.Browser{
		Entry:        &desktopentry.Entry{Exec: "/usr/bin/flatpak run app.zen_browser.zen @@u %u @@"},
		PrivateEntry: &desktopentry.Entry{Exec: "/usr/bin/flatpak run app.zen_browser.zen --private-window @@u %u @@"},
		ProfileArgs:  []string{"-P", "Default (release)"},
	}
	got, err := LinuxArgs(zen, url)
	if want := []string{"/usr/bin/flatpak", "run", "app.zen_browser.zen", "-P", "Default (release)", "@@u", url, "@@"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("profile:\n got %q %v\nwant %q", got, err, want)
	}
	zen.OpenPrivate = true
	got, err = LinuxArgs(zen, url)
	if want := []string{"/usr/bin/flatpak", "run", "app.zen_browser.zen", "--private-window", "-P", "Default (release)", "@@u", url, "@@"}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("private profile:\n got %q %v\nwant %q", got, err, want)
	}
	brave := discovery.Browser{
		Entry:        &desktopentry.Entry{Exec: "/usr/bin/brave-browser-stable %U"},
		PrivateEntry: &desktopentry.Entry{Exec: "/usr/bin/brave-browser-stable --incognito"},
		ProfileArgs:  []string{"--profile-directory=Profile 1"},
		OpenPrivate:  true,
	}
	got, err = LinuxArgs(brave, url)
	if want := []string{"/usr/bin/brave-browser-stable", "--incognito", "--profile-directory=Profile 1", url}; err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("brave:\n got %q %v\nwant %q", got, err, want)
	}
	if _, err := LinuxArgs(discovery.Browser{Name: "X", Entry: &desktopentry.Entry{Exec: "x %u"}, OpenPrivate: true}, url); err == nil {
		t.Fatal("private launch without support accepted")
	}
}
