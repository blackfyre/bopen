package register

import (
	"errors"

	"github.com/blackfyre/bopen/internal/winreg"
)

const (
	progID          = "bopenURL"
	progIDKey       = `Software\Classes\` + progID
	clientKey       = `Software\Clients\StartMenuInternet\bopen`
	capabilitiesKey = clientKey + `\Capabilities`
	registeredApps  = `Software\RegisteredApplications`
	settingsURI     = "ms-settings:defaultapps?registeredAppUser=bopen"
	description     = "Inspect links, remove tracking parameters and choose a browser"
)

// Windows registers bopen per user under HKEY_CURRENT_USER.
type Windows struct {
	Reg winreg.Registry
	// Exe is the absolute path of bopen.exe.
	Exe string
	// OpenSettings opens a ms-settings: URI.
	OpenSettings func(uri string) error
}

type value struct{ path, name, data string }

func (w Windows) values() []value {
	exe := `"` + w.Exe + `"`
	return []value{
		{progIDKey, "", "bopen URL"},
		{progIDKey, "FriendlyTypeName", "bopen URL"},
		{progIDKey + `\DefaultIcon`, "", exe + ",0"},
		{progIDKey + `\shell\open\command`, "", exe + ` "%1"`},
		{clientKey, "", "bopen"},
		{clientKey + `\DefaultIcon`, "", exe + ",0"},
		{clientKey + `\shell\open\command`, "", exe},
		{capabilitiesKey, "ApplicationName", "bopen"},
		{capabilitiesKey, "ApplicationDescription", description},
		{capabilitiesKey, "ApplicationIcon", exe + ",0"},
		{capabilitiesKey + `\StartMenu`, "StartMenuInternet", "bopen"},
		{capabilitiesKey + `\URLAssociations`, "http", progID},
		{capabilitiesKey + `\URLAssociations`, "https", progID},
		{registeredApps, "bopen", capabilitiesKey},
	}
}

// Register writes the per-user registration and opens Default Apps settings,
// where the user must confirm bopen as the default browser.
func (w Windows) Register() (string, error) {
	for _, v := range w.values() {
		if err := w.Reg.SetString(winreg.CurrentUser, v.path, v.name, v.data); err != nil {
			return "", err
		}
	}
	msg := "bopen is registered as a web browser.\n" +
		"Windows does not allow programs to make themselves the default: in the Default Apps " +
		"settings that are now opening, select bopen as the default for HTTP and HTTPS."
	if err := w.OpenSettings(settingsURI); err != nil {
		msg += "\nCould not open Settings (" + err.Error() + "); open Settings > Apps > Default apps > bopen."
	}
	return msg, nil
}

// Unregister removes everything Register created.
func (w Windows) Unregister() (string, error) {
	for _, key := range []string{progIDKey, clientKey} {
		if err := w.Reg.DeleteTree(winreg.CurrentUser, key); err != nil && !errors.Is(err, winreg.ErrNotExist) {
			return "", err
		}
	}
	if err := w.Reg.DeleteValue(winreg.CurrentUser, registeredApps, "bopen"); err != nil && !errors.Is(err, winreg.ErrNotExist) {
		return "", err
	}
	return "bopen has been unregistered. Choose a default browser in Settings > Apps > Default apps.", nil
}
