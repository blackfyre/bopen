package discovery

import "github.com/blackfyre/bopen/internal/winreg"

// Discover lists the installed browsers.
func Discover() []Browser { return Expand(Windows(winreg.System{}), SystemPaths()) }
