package discovery

import (
	"strings"

	"github.com/blackfyre/bopen/internal/winreg"
)

const startMenuInternet = `SOFTWARE\Clients\StartMenuInternet`

// Windows discovers browsers registered under StartMenuInternet in HKCU and
// HKLM. A client registered in both hives is taken from HKCU.
//
// The launch command is the browser's registered https URL handler command
// (via Capabilities\URLAssociations) when available, because that is the
// command Windows itself uses for links; otherwise the client's own
// shell\open\command is used and the URL is appended to it.
func Windows(reg winreg.Registry) []Browser {
	seen := map[string]bool{}
	var browsers []Browser
	for _, hive := range []winreg.Hive{winreg.CurrentUser, winreg.LocalMachine} {
		names, err := reg.SubKeys(hive, startMenuInternet)
		if err != nil {
			continue
		}
		for _, name := range names {
			key := strings.ToLower(name)
			if seen[key] {
				continue
			}
			seen[key] = true
			if strings.EqualFold(name, SelfClientKey) {
				continue
			}
			client := startMenuInternet + `\` + name
			command, err := reg.String(hive, client+`\shell\open\command`, "")
			if err != nil || strings.TrimSpace(command) == "" {
				continue
			}
			if handler := urlHandlerCommand(reg, hive, client); handler != "" {
				command = handler
			}
			display, err := reg.String(hive, client, "")
			if err != nil || display == "" {
				display = name
			}
			kind := KindSystem
			if hive == winreg.CurrentUser {
				kind = KindUser
			}
			browsers = append(browsers, Browser{ID: name, Name: display, Kind: kind, Command: command})
		}
	}
	sortBrowsers(browsers)
	return browsers
}

// urlHandlerCommand returns the open command of the ProgID the client
// associates with https, or "".
func urlHandlerCommand(reg winreg.Registry, hive winreg.Hive, client string) string {
	progID, err := reg.String(hive, client+`\Capabilities\URLAssociations`, "https")
	if err != nil || progID == "" {
		return ""
	}
	for _, h := range []winreg.Hive{hive, winreg.CurrentUser, winreg.LocalMachine} {
		cmd, err := reg.String(h, `SOFTWARE\Classes\`+progID+`\shell\open\command`, "")
		if err == nil && strings.TrimSpace(cmd) != "" {
			return cmd
		}
	}
	return ""
}
