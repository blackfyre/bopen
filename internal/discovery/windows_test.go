package discovery

import (
	"testing"

	"github.com/blackfyre/bopen/internal/winreg"
)

const smi = `SOFTWARE\Clients\StartMenuInternet`

func TestWindowsDiscovery(t *testing.T) {
	reg := winreg.NewFake()
	chrome := smi + `\Google Chrome`
	reg.SetString(winreg.LocalMachine, chrome, "", "Google Chrome")
	reg.SetString(winreg.LocalMachine, chrome+`\shell\open\command`, "", `"C:\Program Files\Google\Chrome\Application\chrome.exe"`)
	reg.SetString(winreg.LocalMachine, chrome+`\Capabilities\URLAssociations`, "https", "ChromeHTML")
	reg.SetString(winreg.LocalMachine, `SOFTWARE\Classes\ChromeHTML\shell\open\command`, "",
		`"C:\Program Files\Google\Chrome\Application\chrome.exe" --single-argument %1`)

	// Per-user Firefox in both hives: HKCU wins.
	reg.SetString(winreg.CurrentUser, smi+`\Firefox-308046B0AF4A39CB`, "", "Firefox (user)")
	reg.SetString(winreg.CurrentUser, smi+`\Firefox-308046B0AF4A39CB\shell\open\command`, "", `"C:\Users\u\Firefox\firefox.exe"`)
	reg.SetString(winreg.LocalMachine, smi+`\Firefox-308046B0AF4A39CB`, "", "Firefox (machine)")
	reg.SetString(winreg.LocalMachine, smi+`\Firefox-308046B0AF4A39CB\shell\open\command`, "", `"C:\Program Files\Mozilla Firefox\firefox.exe"`)

	// No default value: name falls back to the key.
	reg.SetString(winreg.LocalMachine, smi+`\Opera\shell\open\command`, "", `"C:\Opera\launcher.exe"`)
	// No command: not a browser.
	reg.SetString(winreg.LocalMachine, smi+`\Broken`, "", "Broken")
	// bopen itself.
	reg.SetString(winreg.CurrentUser, smi+`\bopen`, "", "bopen")
	reg.SetString(winreg.CurrentUser, smi+`\bopen\shell\open\command`, "", `"C:\bopen.exe"`)

	bs := Windows(reg)
	if len(bs) != 3 {
		t.Fatalf("got %+v", bs)
	}
	byID := map[string]Browser{}
	for _, b := range bs {
		byID[b.ID] = b
	}
	if b := byID["Google Chrome"]; b.Kind != KindSystem || b.Name != "Google Chrome" ||
		b.Command != `"C:\Program Files\Google\Chrome\Application\chrome.exe" --single-argument %1` {
		t.Errorf("chrome: %+v", b)
	}
	if b := byID["Firefox-308046B0AF4A39CB"]; b.Kind != KindUser || b.Name != "Firefox (user)" ||
		b.Command != `"C:\Users\u\Firefox\firefox.exe"` {
		t.Errorf("firefox: %+v", b)
	}
	if b := byID["Opera"]; b.Name != "Opera" {
		t.Errorf("opera: %+v", b)
	}
	if bs[0].ID != "Firefox-308046B0AF4A39CB" || bs[1].ID != "Google Chrome" || bs[2].ID != "Opera" {
		t.Errorf("not sorted by name: %v", ids(bs))
	}
}

func TestWindowsEmptyRegistry(t *testing.T) {
	if bs := Windows(winreg.NewFake()); len(bs) != 0 {
		t.Fatalf("got %+v", bs)
	}
}
