package launch

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/blackfyre/bopen/internal/discovery"
)

const (
	detachedProcess       = 0x00000008
	createNewProcessGroup = 0x00000200
)

func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess | createNewProcessGroup}
}

func start(b discovery.Browser, url string) error {
	if b.Command == "" {
		return errors.New("browser has no registered command")
	}
	exe, line, err := WindowsCommandLine(b.Command, url)
	if err != nil {
		return err
	}
	cmd := exec.Command(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:       line,
		CreationFlags: detachedProcess | createNewProcessGroup,
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
