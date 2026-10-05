# Tasks

## 1. Appearance detection

- [x] 1.1 Create `internal/appearance` with `Settings`, pure parsers for the portal values (scheme, contrast, `(ddd)` accent with a range check) and the Windows DWORDs (`AppsUseLightTheme`, `0xAABBGGRR` accent). Verify with unit tests for dark, light, no preference, an out-of-range accent and an ABGR decode.
- [x] 1.2 Implement the Linux reader (`ReadOne` with a timeout, `Read` fallback, concurrent keys) and `Watch` on `SettingChanged`. Verify with a test against a private D-Bus session (`dbus-daemon --session`), using a fake portal object that serves values and emits a change. Check manually that `go run` on this machine reports dark mode and the COSMIC accent colour.
- [x] 1.3 Add `Integer` to `internal/winreg` (fake and real), and implement the Windows reader with the high-contrast flag. Verify with fake-registry tests, and confirm `GOOS=windows go build ./...` succeeds.

## 2. Palette

- [x] 2.1 Implement `Palette`/`NewPalette` with light and dark bases, accent adjustment, `OnAccent`, high-contrast variants and `material.Theme` derivation. Verify with the legibility test enumerating modes, contrast and accents against the ratios in the spec.

## 3. UI

- [x] 3.1 Replace hard-coded colours in `internal/ui` with palette tokens, and restyle the inspector (cards, kind chips, action bar, outlined Cancel, accent Open). Verify with offscreen renders in light, dark and high contrast, and confirm the existing UI tests pass.
- [x] 3.2 Restyle the settings view, rule form and menus with the same tokens. Verify with offscreen renders in light and dark.
- [x] 3.3 Apply appearance at start-up and on change (Linux watch, Windows focus), plus the Windows dark title bar. Verify with a test that a delivered `Settings` change re-themes the window, and check manually that the real window opens dark on this desktop.

## 4. Verification

- [x] 4.1 Mention the system appearance support in the README, then run `gofmt -l .` (empty), `go vet ./...` (Linux and Windows), `go test ./...`, the Windows GUI build and `scripts/test-install.sh`, all passing.
