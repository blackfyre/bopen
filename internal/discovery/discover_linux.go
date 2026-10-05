package discovery

// Discover lists the installed browsers.
func Discover() []Browser { return Expand(DiscoverLinux(), SystemPaths()) }
