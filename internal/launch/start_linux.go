package launch

import (
	"os/exec"
	"syscall"

	"github.com/blackfyre/bopen/internal/discovery"
)

func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func start(b discovery.Browser, url string) error {
	args, err := LinuxArgs(b, url)
	if err != nil {
		return err
	}
	return startDetached(args)
}
