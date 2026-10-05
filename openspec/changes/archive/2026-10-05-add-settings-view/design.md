# Design

## Context

This builds on `add-link-inspector-core` (see its design for package layout, the Gio window and `internal/prefs`). Until now `config.toml` is read-only for bopen. This change makes it UI-owned. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One window with two views (inspector and settings) and in-memory state preserved across view switches.
- Writing `config.toml` safely when several short-lived bopen instances may be open at once.

**Non-Goals:**
- User-defined rules and the right-click "flag this parameter" action (`add-user-rules`).
- The ClearURLs toggle (`add-clearurls-rules`).
- Preserving hand-written comments in `config.toml`.

## Decisions

### View switching in a single window

`internal/ui` gets a top-level `view` state (`inspector` | `settings`). The inspector model (analysis result, toggle states, selection) is held outside the view, so switching views doesn't rebuild it. When the user returns from settings, bopen re-reads the preferences, re-applies hidden/order filtering to the already-discovered browser list, and re-runs analysis with the current disabled-rule set. Toggles are carried over by suggestion span, so a suggestion that still exists keeps the user's choice. Browsers are not re-discovered, because discovery doesn't depend on any setting.

Opening settings in a separate window was rejected. Gio supports multiple windows, but that complicates focus and lifecycle for a short-lived process, and the agreed UX is a screen swap.

### Read-modify-write per change

Each settings change is expressed as a small mutation function on the decoded config (for example "set window", "toggle hidden X", "move X up"). On save, bopen:
1. re-reads `config.toml`;
2. applies the single mutation;
3. encodes the result;
4. writes it to a temporary file and renames it over the original.

This keeps a second open instance from overwriting unrelated changes, and avoids a lock file. Two instances changing the *same* setting at the same moment simply resolve as last-write-wins, which is acceptable.

Encoding uses `BurntSushi/toml` with a fixed struct, so key order is stable and comments are dropped (documented in the README).

### Config schema additions

```toml
window = "always"

[browsers]
hidden = ["firefox.desktop"]
order  = ["app.zen_browser.zen.desktop", "brave-browser.desktop"]

[rules]
disabled = ["fbclid"]
```

The lists hold identities, not indices, so uninstalling and reinstalling a browser keeps its settings. Moving a browser up or down operates on the visible order. Moving it writes the full current visible order into `order`, which keeps the semantics simple and the result predictable.

### Built-in rule identifiers

Every `[[rule]]` in `builtin.toml` gains an `id` (for example `utm`, `fbclid`, `redirect-google`, `affiliate-amazon-tag`). The existing loader test is extended to assert that IDs are unique and non-empty. IDs are part of the user's config, so a rule may be renamed only by keeping its old ID.

### Registration status

- **Linux.** `xdg-mime query default x-scheme-handler/https`, compared with `bopen.desktop`.
- **Windows.** The effective `UserChoice` ProgID for `https` under `HKCU\...\Shell\Associations\UrlAssociations\https\UserChoice` can be read (not written), and is compared with `bopenURL`.

Re-register calls the same `internal/register` function as the CLI. On Windows it therefore opens the Default Apps settings again.

### `bopen settings`

`cmd/bopen` dispatches the new subcommand to `ui.RunSettings`, which starts the window in the settings view with no inspector model, so the back control is not shown.

## Risks / Trade-offs

- **[Comments lost]** Users who hand-edited `config.toml` lose their comments on the first UI save. → Documented. The file stays plain and readable.
- **[Re-analysis on return changes the suggestion list]** Disabling a rule removes its suggestion. → Carry toggles over by span, so the remaining suggestions keep the user's choices.
- **[Registry read of UserChoice may be absent]** It is missing on fresh profiles. → Show the status as "unknown" rather than "not default".
