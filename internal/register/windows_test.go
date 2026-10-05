package register

import (
	"strings"
	"testing"

	"github.com/blackfyre/bopen/internal/winreg"
)

func TestWindowsRegister(t *testing.T) {
	reg := winreg.NewFake()
	opened := ""
	w := Windows{Reg: reg, Exe: `C:\Tools\bopen.exe`, OpenSettings: func(uri string) error { opened = uri; return nil }}
	msg, err := w.Register()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ path, name, want string }{
		{`Software\Classes\bopenURL\shell\open\command`, "", `"C:\Tools\bopen.exe" "%1"`},
		{`Software\Clients\StartMenuInternet\bopen`, "", "bopen"},
		{`Software\Clients\StartMenuInternet\bopen\shell\open\command`, "", `"C:\Tools\bopen.exe"`},
		{`Software\Clients\StartMenuInternet\bopen\Capabilities`, "ApplicationName", "bopen"},
		{`Software\Clients\StartMenuInternet\bopen\Capabilities\URLAssociations`, "http", "bopenURL"},
		{`Software\Clients\StartMenuInternet\bopen\Capabilities\URLAssociations`, "https", "bopenURL"},
		{`Software\RegisteredApplications`, "bopen", `Software\Clients\StartMenuInternet\bopen\Capabilities`},
	} {
		got, err := reg.String(winreg.CurrentUser, tc.path, tc.name)
		if err != nil || got != tc.want {
			t.Errorf("%s [%s] = %q, %v; want %q", tc.path, tc.name, got, err, tc.want)
		}
	}
	if reg.Exists(winreg.LocalMachine, `Software\Clients\StartMenuInternet\bopen`) {
		t.Error("wrote to HKLM")
	}
	if opened != "ms-settings:defaultapps?registeredAppUser=bopen" {
		t.Errorf("opened %q", opened)
	}
	if !strings.Contains(msg, "Default Apps") {
		t.Errorf("message lacks instructions: %q", msg)
	}
}

func TestWindowsUnregister(t *testing.T) {
	reg := winreg.NewFake()
	reg.SetString(winreg.CurrentUser, `Software\RegisteredApplications`, "Other", "x")
	w := Windows{Reg: reg, Exe: `C:\bopen.exe`, OpenSettings: func(string) error { return nil }}
	if _, err := w.Register(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Unregister(); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`Software\Classes\bopenURL`, `Software\Clients\StartMenuInternet\bopen`} {
		if reg.Exists(winreg.CurrentUser, key) {
			t.Errorf("%s still exists", key)
		}
	}
	if _, err := reg.String(winreg.CurrentUser, `Software\RegisteredApplications`, "bopen"); err == nil {
		t.Error("RegisteredApplications value still exists")
	}
	if v, _ := reg.String(winreg.CurrentUser, `Software\RegisteredApplications`, "Other"); v != "x" {
		t.Error("unrelated RegisteredApplications value removed")
	}
	// Unregistering twice is harmless.
	if _, err := w.Unregister(); err != nil {
		t.Fatal(err)
	}
}
