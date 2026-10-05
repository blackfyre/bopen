package register

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"

	"github.com/blackfyre/bopen/internal/winreg"
)

// Registrar registers and unregisters bopen and reports whether it is the
// default handler.
type Registrar interface {
	Register() (string, error)
	Unregister() (string, error)
	IsDefault() (isDefault, known bool)
}

// System returns the registrar for this machine.
func System() (Registrar, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return Windows{Reg: winreg.System{}, Exe: exe, OpenSettings: shellOpen}, nil
}

func shellOpen(uri string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(uri)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}
