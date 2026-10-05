package register

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/blackfyre/bopen/internal/desktopentry"
	"github.com/blackfyre/bopen/internal/prefs"
)

// fakeXDG emulates xdg-mime's default/query commands.
type fakeXDG struct{ defaults map[string]string }

func (f *fakeXDG) run(name string, args ...string) (string, error) {
	switch {
	case len(args) == 3 && args[0] == "query" && args[1] == "default":
		return f.defaults[args[2]] + "\n", nil
	case len(args) >= 3 && args[0] == "default":
		for _, mime := range args[2:] {
			f.defaults[mime] = args[1]
		}
		return "", nil
	}
	return "", errors.New("unexpected xdg-mime call: " + strings.Join(args, " "))
}

func newLinux(t *testing.T, installed ...string) (Linux, *fakeXDG) {
	base := t.TempDir()
	xdg := &fakeXDG{defaults: map[string]string{
		"x-scheme-handler/http":  "brave-browser.desktop",
		"x-scheme-handler/https": "brave-browser.desktop",
	}}
	return Linux{
		DataHome: filepath.Join(base, "data"),
		Exe:      "/home/u/My Apps/bopen",
		StateDir: filepath.Join(base, "config", "bopen"),
		Run:      xdg.run,
		LookPath: func(string) (string, error) { return "/usr/bin/xdg-mime", nil },
		Installed: func(id string) bool {
			for _, i := range installed {
				if i == id {
					return true
				}
			}
			return false
		},
	}, xdg
}

func TestLinuxRegister(t *testing.T) {
	l, xdg := newLinux(t)
	if _, err := l.Register(); err != nil {
		t.Fatal(err)
	}
	if xdg.defaults["x-scheme-handler/https"] != "bopen.desktop" || xdg.defaults["x-scheme-handler/http"] != "bopen.desktop" {
		t.Fatalf("defaults = %v", xdg.defaults)
	}
	if st := prefs.LoadState(l.StateDir); st.PreviousDefault != "brave-browser.desktop" {
		t.Fatalf("previous default = %q", st.PreviousDefault)
	}
	e, err := desktopentry.ParseFile(filepath.Join(l.DataHome, "applications", "bopen.desktop"), "C")
	if err != nil {
		t.Fatal(err)
	}
	if !e.NoDisplay || !e.HasMimeType("x-scheme-handler/http") || !e.HasMimeType("x-scheme-handler/https") {
		t.Fatalf("entry = %+v", e)
	}
	args, err := e.Args("https://example.com/")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/home/u/My Apps/bopen", "https://example.com/"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %q", args)
	}
}

func TestLinuxRegisterTwiceKeepsPreviousDefault(t *testing.T) {
	l, _ := newLinux(t)
	for i := 0; i < 2; i++ {
		if _, err := l.Register(); err != nil {
			t.Fatal(err)
		}
	}
	if st := prefs.LoadState(l.StateDir); st.PreviousDefault != "brave-browser.desktop" {
		t.Fatalf("previous default = %q", st.PreviousDefault)
	}
}

func TestLinuxUnregisterRestores(t *testing.T) {
	l, xdg := newLinux(t, "brave-browser.desktop")
	if _, err := l.Register(); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Unregister(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(l.DataHome, "applications", "bopen.desktop")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("bopen.desktop still present: %v", err)
	}
	if xdg.defaults["x-scheme-handler/https"] != "brave-browser.desktop" || xdg.defaults["x-scheme-handler/http"] != "brave-browser.desktop" {
		t.Fatalf("defaults = %v", xdg.defaults)
	}
}

func TestLinuxUnregisterPreviousUninstalled(t *testing.T) {
	l, xdg := newLinux(t)
	if _, err := l.Register(); err != nil {
		t.Fatal(err)
	}
	msg, err := l.Unregister()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, "no longer installed") || xdg.defaults["x-scheme-handler/https"] != "bopen.desktop" {
		t.Fatalf("msg %q defaults %v", msg, xdg.defaults)
	}
}

func TestLinuxRegisterWithoutXDGMime(t *testing.T) {
	l, _ := newLinux(t)
	l.LookPath = func(string) (string, error) { return "", errors.New("not found") }
	_, err := l.Register()
	if err == nil || !strings.Contains(err.Error(), "xdg-utils") {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(l.DataHome, "applications", "bopen.desktop")); statErr == nil {
		t.Fatal("desktop file written despite missing xdg-mime")
	}
}
