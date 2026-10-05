package launch

import (
	"errors"
	"os/exec"

	"github.com/blackfyre/bopen/internal/discovery"
)

// LinuxArgs builds the argument vector that opens url in browser b.
func LinuxArgs(b discovery.Browser, url string) ([]string, error) {
	if b.Entry == nil {
		return nil, errors.New("browser has no desktop entry")
	}
	return b.Entry.Args(url)
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
