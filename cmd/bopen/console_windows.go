package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// attachConsole connects a GUI-subsystem bopen to the console of the shell
// that started it, so command-line output is visible. Standard handles that
// are already valid, such as pipes and files, are kept, so `bopen clean`
// works in pipelines.
func attachConsole() {
	stdoutOK := validStdHandle(windows.STD_OUTPUT_HANDLE)
	stderrOK := validStdHandle(windows.STD_ERROR_HANDLE)
	if stdoutOK && stderrOK {
		return
	}
	const attachParentProcess = ^uintptr(0)
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("AttachConsole")
	if ok, _, _ := proc.Call(attachParentProcess); ok == 0 {
		return
	}
	name, _ := windows.UTF16PtrFromString("CONOUT$")
	h, err := windows.CreateFile(name, windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return
	}
	f := os.NewFile(uintptr(h), "CONOUT$")
	if !stdoutOK {
		os.Stdout = f
	}
	if !stderrOK {
		os.Stderr = f
	}
}

func validStdHandle(std uint32) bool {
	h, err := windows.GetStdHandle(std)
	if err != nil || h == 0 || h == windows.InvalidHandle {
		return false
	}
	t, err := windows.GetFileType(h)
	return err == nil && t != windows.FILE_TYPE_UNKNOWN
}
