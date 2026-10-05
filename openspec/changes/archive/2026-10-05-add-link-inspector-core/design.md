# Design

## Context

Greenfield Go project. The repository holds only `mise.toml` (Go `latest`) and the OpenSpec scaffold. bopen is a short-lived process started once per clicked link, so start-up latency and binary size matter more than long-running concerns. Requirements are in `specs/`; motivation and scope are in `proposal.md`.

Platform facts that shape the approach:
- **Linux.** Flatpak and Snap browsers are ordinary desktop entries in known export directories. Their `Exec` lines already wrap `flatpak run` and `snap run`, so no install-kind-specific launch path is needed.
- **Windows.** The default browser cannot be set programmatically (the `UserChoice` hash is protected). Registration plus a user confirmation in Settings is the only supported route.
- **Distribution.** bopen must run unsandboxed. As a Flatpak it could neither read host desktop entries nor launch host browsers without a sandbox escape.

## Goals / Non-Goals

**Goals:**
- A pure-Go core (discovery, analysis, launch-command building, preferences) that is testable without a display and without the real file system or registry.
- The inspector window appears quickly. The target is under 150 ms from process start to first frame on a typical desktop.
- One code base, with platform specifics isolated behind build tags.

**Non-Goals:**
- The settings view, user-defined rules, and the ClearURLs fetch (follow-up change).
- Browser profile selection, such as Firefox or Chrome profiles.
- macOS and Snap-packaged distribution of bopen itself.
- Intercepting links that bypass the default browser on Windows (Widgets, Search, `microsoft-edge:` URIs).

## Decisions

### Package layout

```
cmd/bopen/            main: argument parsing, mode dispatch
internal/discovery/   browser list (discovery_linux.go, discovery_windows.go)
internal/desktopentry/ Desktop Entry parsing and Exec expansion (Linux)
internal/clean/       rules, analysis, cleaned-URL composition
internal/clean/rules/ builtin.toml (go:embed)
internal/launch/      URL validation, command building, detached start
internal/register/    register/unregister (register_linux.go, register_windows.go)
internal/prefs/       config.toml and state.toml
internal/ui/          Gio inspector window
```

Only `internal/ui` and `cmd/bopen` import Gio. Everything else is plain Go with injected inputs: directory roots, a registry reader interface, and an exec starter. That's what makes the core unit-testable.

### GUI toolkit: Gio

Gio was chosen over Fyne and over a webview. Gio's immediate-mode rendering and `richtext` package make the inline highlighted URL straightforward. It starts quickly, and the Windows build needs no cgo. Fyne's `RichText` has weaker support for inline styled spans and starts more slowly. A webview (WebKitGTK or WebView2) would make the highlighting trivial, but adds a heavy runtime dependency on Linux and a slower cold start. The cost of Gio is cgo on Linux (Wayland, X11, EGL and xkbcommon headers at build time), which is acceptable for a native binary.

### Span-based analysis on the raw URL

Analysis works on the raw query string, split on `&`, recording byte offsets. It does not use `url.Values`, because that reorders parameters and re-encodes values. That would violate the requirement to keep remaining parameters unchanged, and the spans would no longer match what the user sees. Composition then rebuilds the URL from the kept raw segments. For redirect wrappers, the target value is percent-decoded, validated as an absolute `http(s)` URL, and analysed recursively. Its suggestions record the parent redirect suggestion as their dependency.

### Rule data: embedded TOML

Built-in rules live in `internal/clean/rules/builtin.toml`, embedded with `go:embed`, rather than in Go literals. The same schema will later hold user rules, so the follow-up settings change reuses the loader. One rule looks like this:

```toml
[[rule]]
kind   = "tracking"          # tracking | affiliate | redirect
param  = "utm_*"             # tracking/affiliate: name or prefix*
hosts  = []                  # optional host globs; empty = any host
reason = "Campaign attribution for analytics (Google Analytics and others); does not affect the page."

[[rule]]
kind   = "redirect"
hosts  = ["www.google.*"]
path   = "/url"
target = ["q", "url"]        # first present parameter holds the target
reason = "Google click-tracking redirect: Google logs the click before forwarding you."
```

A unit test validates the embedded file: it parses, every rule has a non-empty reason, and the kinds are known.

### Own Desktop Entry parser

bopen uses a minimal in-house parser for `[Desktop Entry]` keys, not a library. It handles `Name` (unlocalised, then `Name[<lang>]` from `LANG`), `Exec`, `TryExec`, `MimeType`, `Hidden`, `NoDisplay` and `Type`. `Exec` tokenising follows the specification's quoting and escape rules. Field codes are expanded as follows:
- `%u` and `%U` become the URL as a single argument;
- `%f`, `%F`, `%d`, `%D`, `%n`, `%N`, `%v` and `%m` are removed;
- `%i` becomes `--icon <Icon>` when `Icon` is set;
- `%c` becomes the name;
- `%k` becomes the file path;
- `%%` becomes `%`.

The specification is small, and the available Go libraries either pull in GTK or don't do `Exec` expansion. Calling `gio launch` was rejected because it adds a runtime dependency and gives no control over argument handling.

### Desktop file ID and kind

The desktop file ID is the path relative to the `applications` directory, with `/` replaced by `-`. The install kind is derived from which root directory the file came from (see the `browser-discovery` spec).

### Windows registry behind an interface

`discovery` and `register` depend on a small interface with three operations (list subkeys, read a value, write or delete keys), implemented with `golang.org/x/sys/windows/registry`. Tests use an in-memory fake, so Windows logic is testable on Linux CI.

### Launch: argv on Linux, raw command line on Windows

- **Linux.** `exec.Cmd` with the expanded argv; no shell. `SysProcAttr.Setsid = true` detaches the process; bopen calls `Start()`, then `Process.Release()`.
- **Windows.** Go's automatic argument quoting can't be trusted to reproduce the registered command template. bopen therefore takes the browser's https URL-handler command (the ProgID from `Capabilities\URLAssociations`, falling back to the client's `shell\open\command`), parses the executable from it (quoted first token), substitutes `%1`/`%L` in place (or appends the URL), and passes the full string via `SysProcAttr.CmdLine`. The process is started with `DETACHED_PROCESS | CREATE_NEW_PROCESS_GROUP`. The template's quoting is kept as it is: a quoted `"%1"` stays quoted, and an unquoted `%1` stays unquoted. Adding quotes would break Chrome-family `--single-argument %1`, which takes the rest of the line verbatim. Validated URLs contain no whitespace or `"`, so they are always one argument. Trailing backslashes are doubled inside quotes.
- **URL validation.** Before substitution, the URL is canonicalised with `net/url` (`URL.String()`), which percent-encodes spaces. Any remaining `"` is replaced by `%22`. Validation guarantees the argument starts with `http`, so it can never be parsed as an option.

### Registration mechanics

- **Linux.** bopen writes `bopen.desktop` (`Exec=<abs path> %u`, `NoDisplay=true`, MimeType for http and https). It reads the previous default with `xdg-mime query default x-scheme-handler/https`, stores it in `state.toml` unless it's `bopen.desktop`, and then runs `xdg-mime default bopen.desktop x-scheme-handler/http x-scheme-handler/https`. bopen doesn't take over `text/html`, because local files are rejected by the handoff rules anyway. `xdg-mime` (xdg-utils) is a runtime requirement for `register` and `unregister` only, and bopen reports a clear error if it's missing. `NoDisplay=true` keeps bopen out of application menus. Its own discovery excludes it by ID regardless.
- **Windows.** bopen writes `HKCU\Software\Classes\bopenURL` (URL handler ProgID, `shell\open\command` = `"<abs path>" "%1"`), `HKCU\Software\Clients\StartMenuInternet\bopen` with `Capabilities` (`ApplicationName`, `ApplicationDescription`, `URLAssociations\http|https = bopenURL`), and `HKCU\Software\RegisteredApplications\bopen`. It then opens `ms-settings:defaultapps?registeredAppUser=bopen` via `ShellExecute`. The previous default isn't recorded on Windows, because the protected `UserChoice` can't be restored by bopen anyway. Pre-selection then falls back to the first browser.

### Windows GUI subsystem and console output

The Windows binary is linked with `-H windowsgui`, so link clicks don't flash a console. For `register`, `unregister` and usage output, bopen calls `AttachConsole(ATTACH_PARENT_PROCESS)` so that messages reach the invoking terminal.

### Preferences

`internal/prefs` uses `github.com/BurntSushi/toml`, which is mature and decodes cleanly into structs, with metadata for unknown keys. `state.toml` is written to a temporary file in the same directory and renamed over the original. Schema:

```toml
# config.toml
window = "always"            # always | when-suggestions

# state.toml
last_used        = "app.zen_browser.zen.desktop"
previous_default = "brave-browser.desktop"
```

### Process flow

```
args --> parse
          |-- register / unregister --> internal/register --> exit
          '-- <url>
               |-- load prefs (errors collected, not fatal)
               |-- discover browsers
               |-- analyse URL (or record validation error)
               |-- resolve pre-selection
               |-- need window? (preference, suggestions, any error)
               |      no  --> launch --> ok: save state, exit
               |                     '-> fail: show window with error
               '      yes --> inspector --> Enter: launch --> ok: save state, exit
                                       '-> Esc: exit
```

## Risks / Trade-offs

- **[Desktop Entry edge cases]** Unusual `Exec` quoting or locale handling could produce wrong argv. → Table-driven tests from real entries (Firefox, Chromium, Brave, Flatpak Zen, Snap Firefox), stored under `testdata/`.
- **[Windows command templates vary]** Some browsers register `"%1"`, `%1`, `-- "%1"`, or no placeholder. → Treat the placeholder as optional, and keep the template's own quoting. Unit-test the known templates (Chrome, Edge, Firefox, Brave, Opera).
- **[Linux cgo build]** Gio needs system headers, which complicates CI and contributor set-up. → Document the build packages in the README, and keep the core build independent of `internal/ui`, so `go test ./internal/...` runs without them except for the UI package.
- **[Start-up latency]** Scanning every desktop entry on each click. → Only `[Desktop Entry]` groups are read, and parsing stops at the first non-matching group. Typical directories hold a few hundred small files. If measurement shows it matters, cache in `state.toml` keyed by directory modification times. Not planned initially.
- **[Rule false positives]** Removing a parameter can break a site. → Every suggestion can be toggled off, affiliate rules default to off, and reasons are shown so the user can judge.
- **[Windows default-app friction]** The user must confirm in Settings, and Windows may reset defaults after updates. → `register` prints clear instructions and can be re-run safely (idempotent).

## Migration Plan

Not applicable. This is the first release. Rollback for a user is `bopen unregister`, which removes the registration and, on Linux, restores the previous default.
