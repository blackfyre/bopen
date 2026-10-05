# Proposal

## Why

Links clicked in other applications often carry tracking parameters (`utm_*`, `fbclid`, `gclid`) or are wrapped in redirectors (`google.com/url?q=`) that log the click. The user also wants to choose per link which installed browser opens it. No small, open-source, cross-platform tool does both: it should show the link for inspection, explain what it suggests removing and why, and then hand the link to the browser the user picks.

## What Changes

- New Go application `bopen` (MIT licence) that registers as the system's default web browser and is launched once per link.
- Discovers installed browsers automatically. On Linux this covers system, Flatpak (system and user) and Snap installs through their desktop entries; on Windows it reads the registered `StartMenuInternet` clients.
- Analyses the incoming URL against a curated built-in rule set. It unwraps known redirect wrappers and flags tracking parameters, and every suggestion carries a human-readable reason.
- Shows an inspector window that presents the link with the suggested removals highlighted and their reasons, lets the user toggle each suggestion, shows the resulting URL live, and lists the discovered browsers with the last-used one pre-selected.
- A `window` preference (`always` | `when-suggestions`) controls whether links with nothing to suggest skip the window and go straight to the last-used browser.
- Hands the final URL to the chosen browser safely: only `http`/`https`, URL passed as a single argument and never interpretable as a command-line option, and the browser process is detached.
- Self-registration as the default handler. On Linux bopen sets itself as default. On Windows it registers its capabilities and opens the system Default Apps settings, because Windows does not allow setting the default programmatically.
- Preferences in `config.toml` and the last-used browser in `state.toml`, both under the OS user configuration directory.

Explicitly deferred to a follow-up change: the in-app settings view (cog icon), user-defined rules, right-click "flag this parameter", and the opt-in ClearURLs rule fetch. Until then `config.toml` is edited by hand.

## Capabilities

### New Capabilities
- `browser-discovery`: Enumerating installed web browsers on Linux and Windows, with stable identities.
- `link-cleaning`: Detecting redirect wrappers and tracking parameters in a URL, with a reason for each suggestion, and producing the cleaned URL from the accepted suggestions.
- `link-inspector`: The window that presents the link, highlights suggestions with reasons, allows toggling, and selects the target browser.
- `browser-handoff`: Validating the final URL and launching the selected browser with it safely.
- `default-handler-registration`: Registering bopen as the default web browser on Linux and Windows.
- `preferences`: Location, format and semantics of the user preferences and the last-used state.

### Modified Capabilities

None. This is a greenfield project with no existing specs.

## Impact

- New Go module; the repository currently contains only `mise.toml` and the OpenSpec scaffold.
- New dependencies: a GUI toolkit (Gio), a TOML library, and Windows registry access (`golang.org/x/sys/windows/registry`).
- Linux runtime touches the user's XDG MIME associations; Windows runtime writes per-user registry keys under `HKCU`. No administrator rights are required on either platform.
- Distribution as a native binary. A Flatpak build of bopen itself is out of scope because the sandbox prevents launching host browsers.
