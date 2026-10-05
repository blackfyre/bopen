# Proposal

## Why

bopen's window uses Gio's fixed light material theme with hard-coded colours. On a dark desktop it is a bright rectangle, it ignores the user's accent colour and high-contrast setting, and the layout is a plain list of text. bopen opens on every clicked link, so it should look native, follow the system's light/dark mode, accent colour and contrast preference, and present the link and its suggestions in a clearer, more polished layout.

## What Changes

- Read the host appearance at start-up:
  - **Linux:** the XDG desktop portal (`org.freedesktop.appearance`: `color-scheme`, `accent-color`, `contrast`) over the session D-Bus.
  - **Windows:** `AppsUseLightTheme`, the DWM accent colour and the high-contrast flag.
- Follow changes while the window is open. Linux uses the portal's `SettingChanged` signal. Windows re-reads the settings when the window regains focus.
- Light, dark and high-contrast palettes built from semantic colour tokens, with the system accent colour used for primary actions and selection. All text/background pairs meet WCAG AA contrast.
- A visual refresh of the inspector, settings and rule form:
  - content grouped on rounded surface cards;
  - suggestion kinds shown as coloured chips;
  - an accent-coloured primary button and a separated action bar;
  - consistent spacing and typography.
- On Windows, the native title bar follows dark mode.

## Capabilities

### New Capabilities
- `appearance`: How bopen derives its colours from the host system and keeps text legible.

### Modified Capabilities

None. No existing requirement changes; layout polish is not spec-level behaviour.

## Impact

- New package `internal/appearance`, with platform-specific readers.
- `github.com/godbus/dbus/v5` becomes a direct dependency. It is already in the module graph through Gio.
- `internal/winreg` gains DWORD reads.
- `internal/ui` replaces hard-coded colours with palette tokens and is restyled.
- No change to preferences, rules or link handling.
