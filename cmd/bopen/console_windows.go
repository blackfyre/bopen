package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// attachConsole connects a GUI-subsystem bopen to the console of the shell
// that started it, so command-line output is visible.
func attachConsole() {
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
	os.Stdout, os.Stderr = f, f
}
