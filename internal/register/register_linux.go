package register

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/prefs"
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
	exe, err := executable()
	if err != nil {
		return nil, err
	}
	stateDir, err := prefs.Dir()
	if err != nil {
		return nil, err
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return Linux{
		DataHome: dataHome,
		Exe:      exe,
		StateDir: stateDir,
		Run: func(name string, args ...string) (string, error) {
			out, err := exec.Command(name, args...).Output()
			return string(out), err
		},
		LookPath: exec.LookPath,
		Installed: func(id string) bool {
			for _, b := range discovery.Discover() {
				if b.ID == id {
					return true
				}
			}
			return false
		},
	}, nil
}

func executable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
