//go:build !windows

package ui

import "gioui.org/app"

// viewEvent needs nothing outside Windows: Wayland and X11 compositors draw
// the window decorations in the system style.
func (w *window) viewEvent(app.ViewEvent) {}

func setTitleBarDark(uintptr, bool) {}
