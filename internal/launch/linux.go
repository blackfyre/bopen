package launch

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/blackfyre/bopen/internal/discovery"
)

// LinuxArgs builds the argument vector that opens url in browser b: its
// private-window action when b.OpenPrivate is set, with the profile
// arguments placed just before the URL.
func LinuxArgs(b discovery.Browser, url string) ([]string, error) {
	entry := b.Entry
	if b.OpenPrivate {
		if b.PrivateEntry == nil {
			return nil, errors.New(b.Name + " has no private window")
		}
		entry = b.PrivateEntry
	}
	if entry == nil {
		return nil, errors.New("browser has no desktop entry")
	}
	args, err := entry.Args(url)
	if err != nil || len(b.ProfileArgs) == 0 {
		return args, err
	}
	// Insert before the URL argument, or before a Flatpak "@@u" that
	// introduces it, so the profile arguments are never taken for URLs.
	at := len(args)
	for i := 1; i < len(args); i++ {
		if strings.Contains(args[i], url) {
			at = i
			if args[i-1] == "@@u" {
				at = i - 1
			}
			break
		}
	}
	out := append([]string{}, args[:at]...)
	out = append(out, b.ProfileArgs...)
	return append(out, args[at:]...), nil
}

// startDetached starts args without a shell in its own session and does not
// wait for it.
func startDetached(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
