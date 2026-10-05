# Design

## Context

See proposal.md for motivation. The window is built with Gio's `material` widgets. Colours are hard-coded as package variables (`colTracking`, `colMuted` and others) plus `material.NewTheme()` defaults. Gio exposes no system colour scheme. The Windows view event carries the `HWND`. `godbus/dbus/v5` is already in the module graph through Gio. On the development machine (COSMIC), the portal (Settings version 2) reports `color-scheme` 1, `accent-color` (0.478, 0.635, 0.969) and `contrast` 0.

## Goals / Non-Goals

**Goals:**
- Look native on light and dark desktops, including the user's accent colour.
- Deterministic legibility: a unit test checks every colour pair in every palette.
- A cleaner visual hierarchy, without changing behaviour or layout semantics.

**Non-Goals:**
- Reading GTK/Qt/KDE theme files or system fonts. bopen keeps the bundled Go fonts.
- A user-selectable theme override.
- macOS.

## Decisions

### `internal/appearance`: platform readers returning `Settings`

```go
type Settings struct {
	Dark, HighContrast bool
	Accent             color.NRGBA
	HasAccent          bool
}
func Read() Settings                                  // platform specific, never fails
func Watch(ctx context.Context, changed func(Settings)) // Linux only; no-op elsewhere
```

- **Linux.** `dbus.SessionBus()`, then `org.freedesktop.portal.Settings.ReadOne` on `org.freedesktop.portal.Desktop` with a 200 ms context timeout per key. bopen falls back to the deprecated `Read` method, which returns a doubly wrapped variant, for portals older than version 2. Value parsing (`uint32` scheme and contrast, `(ddd)` accent) is a pure function with unit tests. `Watch` subscribes to `SettingChanged` with a match rule on the `org.freedesktop.appearance` namespace.
- **Windows.** The `internal/winreg` interface gains `Integer` (DWORD). bopen reads `HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize\AppsUseLightTheme` and `HKCU\Software\Microsoft\Windows\DWM\AccentColor` (`0xAABBGGRR`), and the high-contrast flag via `SystemParametersInfoW(SPI_GETHIGHCONTRAST)`. On focus regained (`app.ConfigEvent` with `Focused`), the window calls `Read` again.

The alternative was polling Linux settings. It was rejected: the portal signal is the standard mechanism, and COSMIC, GNOME and KDE all implement it.

### Semantic palette

`internal/ui/theme.go` defines `Palette`, which holds:

| Token | Purpose |
|---|---|
| `Bg` | Window background |
| `Surface` | Cards |
| `SurfaceAlt` | Hover and selected rows |
| `Fg`, `Muted` | Primary and secondary text |
| `Border` | Card and field outlines |
| `Accent`, `OnAccent` | Primary actions and selection |
| `Tracking`, `Affiliate`, `Redirect` | Suggestion kinds |
| `WarnBg`, `WarnFg`, `ErrBg`, `ErrFg` | Banners |
| `Scrim` | Modal overlay |

`NewPalette(Settings)` returns a light or dark base (neutral greys, kind colours tuned per mode), applies the accent, and for high contrast sets `Muted = Fg` and `Border = Fg`. `OnAccent` is chosen by WCAG relative luminance. Unless high contrast is on, the accent is adjusted in lightness until it reaches 4.5:1 against its own text colour, and 3:1 against the surface, so it also works as a selection mark. All drawing code takes colours from the palette, and the `material.Theme` palette (`Bg`, `Fg`, `ContrastBg`, `ContrastFg`) is derived from it, so the stock widgets match.

The legibility test enumerates light and dark, with and without high contrast, and a set of accents (the default, pure yellow, white, black, the COSMIC blue). It asserts the requirement's ratios for every text/background pair.

### Visual refresh

```
+--------------------------------------------------------------+
| bopen                                                [gear]  |
| +----------------------------------------------------------+ |
| | LINK                                                     | |
| | https://shop.example.com/item?utm_source=x&fbclid=y      | |
| +----------------------------------------------------------+ |
| +----------------------------------------------------------+ |
| | SUGGESTED CHANGES                                        | |
| | [x] utm_source=x   (Tracking) built-in                   | |
| |     Campaign attribution ...                             | |
| +----------------------------------------------------------+ |
| +----------------------------------------------------------+ |
| | RESULT   https://shop.example.com/item                   | |
| +----------------------------------------------------------+ |
| +----------------------------------------------------------+ |
| | OPEN IN   (o) 1 Zen Browser  flatpak                     | |
| +----------------------------------------------------------+ |
|--------------------------------------------------------------|
| Enter opens · Esc cancels              [Cancel]  [ Open ]    |
+--------------------------------------------------------------+
```

- Cards are rounded `Surface` rectangles with a 1 dp `Border`. Section titles are small capitals in `Muted`.
- The suggestion kind is a rounded chip in its kind colour, replacing the "tracking ·" prefix.
- The action bar is separated by a hairline. Cancel is an outlined secondary button, and Open uses the accent colour.
- The same card style is used in settings, and in the rule form, which keeps the scrim.

### Windows title bar

On `app.Win32ViewEvent`, bopen calls `DwmSetWindowAttribute(hwnd, DWMWA_USE_IMMERSIVE_DARK_MODE=20, &dark)`, and again on appearance changes. This is supported on Windows 10 20H1 and later; on older versions the call is ignored.

## Risks / Trade-offs

- **[Portal missing or slow]** → A 200 ms timeout per key and fall-back defaults. Read the three keys concurrently to keep the start-up cost under the 250 ms budget.
- **[Accent unreadable as a background]** → Lightness adjustment plus the computed `OnAccent`, covered by the legibility test.
- **[Windows code untested at runtime]** → Registry parsing goes through the fake registry, and the Win32 calls are thin. Verified by the cross-build only.
