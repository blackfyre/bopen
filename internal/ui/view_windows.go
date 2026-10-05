package ui

import (
	"unsafe"

	"gioui.org/app"
	"golang.org/x/sys/windows"
)

// viewEvent records the native window so the title bar can follow dark mode.
func (w *window) viewEvent(e app.ViewEvent) {
	if v, ok := e.(app.Win32ViewEvent); ok && v.HWND != 0 {
		w.hwnd = v.HWND
		setTitleBarDark(w.hwnd, w.look.Dark)
	}
}

// dwmwaUseImmersiveDarkMode is supported from Windows 10 20H1; older
// versions ignore the call.
const dwmwaUseImmersiveDarkMode = 20

var dwmSetWindowAttribute = windows.NewLazySystemDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")

func setTitleBarDark(hwnd uintptr, dark bool) {
	if hwnd == 0 || dwmSetWindowAttribute.Find() != nil {
		return
	}
	var v uint32
	if dark {
		v = 1
	}
	dwmSetWindowAttribute.Call(hwnd, dwmwaUseImmersiveDarkMode, uintptr(unsafe.Pointer(&v)), unsafe.Sizeof(v))
}
