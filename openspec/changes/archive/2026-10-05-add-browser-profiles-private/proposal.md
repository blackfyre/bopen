# Proposal

## Why

Many people keep separate browser profiles (work and personal) or want a single link in a private window. bopen can only pick a browser, so these choices still require copying the link by hand.

## What Changes

- **Private windows:** an "Open in a private window" option in the inspector, with `P` as its shortcut, offered when the selected browser supports it.
  - **Linux:** uses the browser's own `new-private-window` desktop action.
  - **Windows:** uses the browser's known flag (`--incognito`, `--inprivate`, `-private-window`).
- **Profiles:** browsers with two or more profiles get one extra entry per profile ("Zen Browser · Work"). The entry opens that profile.
  - **Firefox family:** read from `profiles.ini`, launched with `-P <name>`.
  - **Chromium family:** read from `Local State`, launched with `--profile-directory=<dir>`.

  Profile entries are ordinary browsers with stable identities, so site rules, hiding, ordering and last-used all work per profile.

## Capabilities

### New Capabilities
- `browser-variants`: Profile entries and private-window launching.

### Modified Capabilities
- `link-inspector`: The private-window option.

## Impact

Touches `internal/desktopentry` (desktop actions), `internal/discovery` (browser families, profile discovery, private launch data), `internal/launch` (extra arguments, private launch on both platforms) and `internal/ui` (option and shortcut). No configuration changes.
