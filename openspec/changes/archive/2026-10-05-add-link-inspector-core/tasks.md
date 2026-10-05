# Tasks

## 1. Project setup

- [x] 1.1 Initialise the Go module, add the MIT `LICENSE`, and create the package layout from design.md. Verify that `go build ./...` and `go vet ./...` succeed.
- [x] 1.2 Add a README covering purpose, the Linux build packages Gio needs (Wayland, X11, EGL, xkbcommon headers), and `register`/`unregister` usage. Verify the README exists and lists the build packages.

## 2. Preferences

- [x] 2.1 Implement `internal/prefs` config loading (`window`, defaults, collected errors for invalid TOML or values, unknown keys ignored) in `os.UserConfigDir()/bopen`. Verify with unit tests for the missing-file, malformed, invalid-value and unknown-key scenarios.
- [x] 2.2 Implement `state.toml` reading and atomic writing (`last_used`, `previous_default`; corrupt state treated as empty). Verify with unit tests that write via a temp file and rename, read a corrupt file as empty, and leave `config.toml` untouched.

## 3. Link cleaning

- [x] 3.1 Define the rule schema and loader, and author `internal/clean/rules/builtin.toml` with every rule required by the `link-cleaning` spec, each with a specific reason. Verify with a test that the embedded file parses, every rule has a non-empty reason and a known kind, and the required rules are present.
- [x] 3.2 Implement span-based analysis of raw query strings (case-insensitive and prefix matching, host-scoped rules, default acceptance per kind). Verify with table-driven tests covering the tracking, clean-URL, upper-case, Amazon `tag`, and non-Amazon `tag` scenarios.
- [x] 3.3 Implement recursive redirect unwrapping (depth limit 5, http/https target validation, dependent suggestions). Verify with tests for Google, Facebook, Outlook Safe Links and YouTube wrappers, tracking inside a wrapped target, a `javascript:` target, and the depth limit.
- [x] 3.4 Implement cleaned-URL composition from accepted suggestions (order and encoding preserved, fragment kept, empty `?` dropped, a rejected redirect disables its dependents). Verify with tests for every composition scenario in the spec.

## 4. Browser discovery

- [x] 4.1 Implement the Desktop Entry parser (`[Desktop Entry]` keys, localised `Name`, spec-compliant `Exec` tokenising and field-code expansion). Verify with tests over real entries in `testdata/` (Firefox, Chromium, Brave, Flatpak Zen, Snap Firefox), including quoting and escape cases.
- [x] 4.2 Implement Linux discovery over injected root directories (precedence order, desktop file ID deduplication including hiding overrides, `TryExec`, `Hidden`/`NoDisplay`, kind derivation, self-exclusion, sorting). Verify with tests over a fake directory tree covering every `browser-discovery` Linux scenario.
- [x] 4.3 Implement Windows discovery behind the registry interface (HKCU over HKLM, display-name fallback, self-exclusion) with a real `x/sys/windows/registry` implementation. Verify with fake-registry unit tests, and confirm `GOOS=windows go build ./...` succeeds.

## 5. Browser handoff

- [x] 5.1 Implement URL validation and canonicalisation (http/https only, non-empty host, `"` encoded). Verify with tests for option-like input, a missing host, `file:` and `javascript:` schemes, and an embedded quote.
- [x] 5.2 Implement Linux command building from the expanded `Exec` and a detached start (`Setsid`, `Start`, `Release`). Verify with unit tests on the built argv (Flatpak `@@u %u @@`, shell metacharacters, no field code), and with a test that launches a stub script and confirms bopen does not wait for it.
- [x] 5.3 Implement Windows command-line building from the registered command (`%1` with or without quotes, or no placeholder) and a detached start via `SysProcAttr.CmdLine`. Verify with unit tests for Chrome, Edge, Firefox, Brave and Opera templates and the embedded-quote case, and confirm the Windows cross-build succeeds.
- [x] 5.4 Record `last_used` only after a successful launch. Verify with a test that a failed start leaves `state.toml` unchanged.

## 6. Registration

- [x] 6.1 Implement Linux `register`/`unregister` (write `bopen.desktop` with an absolute `Exec`, record the previous default unless it is bopen, run `xdg-mime`, restore on unregister, and report a clear error when `xdg-mime` is missing). Verify with tests using a temporary `XDG_DATA_HOME` and a stub `xdg-mime` on `PATH`, covering repeated registration and restore.
- [x] 6.2 Implement Windows `register`/`unregister` (HKCU ProgID, StartMenuInternet capabilities, RegisteredApplications, opening `ms-settings:defaultapps?registeredAppUser=bopen`, user instructions, removal on unregister). Verify with fake-registry tests that assert the exact keys and values written and removed.

## 7. Inspector window

- [x] 7.1 Implement the Gio inspector: original URL with highlighted spans by kind, suggestion list with text, reason, source and toggles (dependents disabled), live result URL, and browser list with kind labels and resolved pre-selection. Verify by running `bopen 'https://www.google.com/url?q=https%3A%2F%2Fexample.com%2F%3Futm_source%3Dx%26fbclid%3Dy'` and checking that the highlights, reasons, toggles and live result behave as specified.
- [x] 7.2 Implement keyboard handling (Enter, Escape, Up/Down, 1–9) with mouse equivalents. Verify manually that Enter opens the result and closes the window, and that Escape closes without launching or changing `state.toml`.
- [x] 7.3 Implement error presentation for unusable input, no browsers found, launch failure (window stays open), and config errors. Verify manually with `file:///etc/passwd`, an empty discovery root, a stub browser that fails to start, and a malformed `config.toml`.

## 8. Entry point and integration

- [x] 8.1 Implement `cmd/bopen` argument parsing (URL, `register`, `unregister`, usage with a non-zero exit) and the process flow from design.md, including the `when-suggestions` silent path. On Windows, link with `-H windowsgui` and call `AttachConsole` for CLI output. Verify with tests for argument dispatch and the window-needed decision, and confirm that both platform builds succeed.
- [x] 8.2 End-to-end check on Linux. Run `bopen register`, click a tracking link in another application, confirm the inspector appears with the last-used browser pre-selected, open the link in a Flatpak browser and in a system browser, then run `bopen unregister` and confirm the previous default is restored.
- [x] 8.3 Run final verification: `gofmt -l .` (empty), `go vet ./...`, `go test ./...`, and `GOOS=windows go build ./...`, all passing.
