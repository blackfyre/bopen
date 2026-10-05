package register

import (
	"os"
	"strings"
	"testing"

	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/winreg"
)

func TestRealRegisterUnregister(t *testing.T) {
	if os.Getenv("BOPEN_WINDOWS_INTEGRATION") != "1" {
		t.Skip("set BOPEN_WINDOWS_INTEGRATION=1 to write bopen's registration to the real registry")
	}
	opened := ""
	w := Windows{Reg: winreg.System{}, Exe: `C:\bopen-test\bopen.exe`, OpenSettings: func(uri string) error { opened = uri; return nil }}
	t.Cleanup(func() { w.Unregister() })
	if _, err := w.Register(); err != nil {
		t.Fatal(err)
	}
	if opened == "" {
		t.Fatal("settings not opened")
	}
	cmd, err := w.Reg.String(winreg.CurrentUser, progIDKey+`\shell\open\command`, "")
	if err != nil || cmd != `"C:\bopen-test\bopen.exe" "%1"` {
		t.Fatalf("command = %q, %v", cmd, err)
	}
	// bopen's own client is excluded from discovery even when registered.
	for _, b := range discovery.Windows(winreg.System{}) {
		if strings.EqualFold(b.ID, discovery.SelfClientKey) {
			t.Fatal("bopen discovered itself")
		}
	}
	if _, known := w.IsDefault(); known {
		t.Log("UserChoice present on this runner")
	}
	if _, err := w.Unregister(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Reg.String(winreg.CurrentUser, progIDKey+`\shell\open\command`, ""); err == nil {
		t.Fatal("registration left behind")
	}
}
