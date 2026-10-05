# Proposal

## Why

The built-in rule set can never cover every tracker. Users need to flag parameters bopen does not know about, explain in their own words why, and do so at the moment they notice the parameter in the inspector, without leaving the window.

## What Changes

- User-defined tracking and affiliate rules, with a parameter name or prefix, an optional host scope and a required reason, stored in `config.toml`.
- A right-click action on any query parameter shown in the inspector: "Always flag this parameter…", which creates a user rule from a short form and re-analyses the link immediately.
- A right-click action on an existing suggestion: disable it (built-in rule) or edit it (user rule).
- A "Your rules" section in the settings view to add, edit and delete user rules.
- User rules take precedence over built-in rules when both match the same parameter.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `link-cleaning`: User rules as a rule source, their validation, and precedence between sources.
- `link-inspector`: Context actions on parameters and suggestions.
- `settings-view`: Management of user rules.
- `preferences`: Storage of user rules in `config.toml`.

## Impact

- Depends on `add-link-inspector-core` and `add-settings-view`, which must be archived first.
- Touches `internal/clean` (rule sources and precedence), `internal/prefs` (rule list and mutations), and `internal/ui` (context menus, rule form, settings section).
- Uses `gioui.org/x/component` for context menus; `gioui.org/x` is already a dependency (added by `add-link-inspector-core` for the highlighted link).
