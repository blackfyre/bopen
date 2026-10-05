package launch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/blackfyre/bopen/internal/discovery"
)

var stubBrowser string

// TestMain builds the stub browser used by the real-launch tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "bopen-launch")
	if err != nil {
		panic(err)
	}
	stubBrowser = filepath.Join(dir, "argv.exe")
	build := exec.Command("go", "build", "-o", stubBrowser, "./testdata/argv")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// launchStub starts the stub through the real Windows launch path and
// returns the arguments it received.
func launchStub(t *testing.T, template, url string) []string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "argv.json")
	t.Setenv("BOPEN_ARGV_OUT", out)
	b := discovery.Browser{ID: "stub", Name: "Stub", Command: template}
	if err := Start(b, url); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		data, err := os.ReadFile(out)
		if err != nil {
			continue
		}
		var args []string
		if err := json.Unmarshal(data, &args); err != nil {
			t.Fatal(err)
		}
		return args
	}
	t.Fatal("stub browser did not run")
	return nil
}

func TestRealLaunchTemplates(t *testing.T) {
	exe := `"` + stubBrowser + `"`
	for _, url := range []string{
		"https://example.com/?a=1&b=2^3%25x",
		`https://example.com/dir\`,
		`https://example.com/?q="quoted" & more`,
	} {
		want, err := Validate(url)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			name, template string
			args           []string
		}{
			{"chrome", exe + " --single-argument %1", []string{"--single-argument", want}},
			{"firefox", exe + ` -osint -url "%1"`, []string{"-osint", "-url", want}},
			{"no placeholder", exe, []string{want}},
		} {
			if got := launchStub(t, tc.template, url); !reflect.DeepEqual(got, tc.args) {
				t.Errorf("%s %q:\n got %q\nwant %q", tc.name, url, got, tc.args)
			}
		}
	}
}

func TestRealLaunchPrivateProfile(t *testing.T) {
	url := "https://example.com/?a=1"
	b := discovery.Browser{ID: "stub", Name: "Stub", Command: `"` + stubBrowser + `" --single-argument %1`,
		PrivateFlag: "--incognito", ProfileArgs: []string{"--profile-directory=Profile 1"}, OpenPrivate: true}
	out := filepath.Join(t.TempDir(), "argv.json")
	t.Setenv("BOPEN_ARGV_OUT", out)
	if err := Start(b, url); err != nil {
		t.Fatal(err)
	}
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		data, err := os.ReadFile(out)
		if err != nil {
			continue
		}
		var args []string
		json.Unmarshal(data, &args)
		want := []string{"--incognito", "--profile-directory=Profile 1", "--single-argument", url}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("got %q, want %q", args, want)
		}
		return
	}
	t.Fatal("stub browser did not run")
}
