# Proposal

## Why

After `add-link-inspector-core`, preferences can only be changed by hand-editing `config.toml`. The agreed interaction is that preferences are edited in the application itself, reached from the inspector, without the user ever needing to touch the file.

## What Changes

- A settings view inside the bopen window, opened from a cog icon in the inspector or directly with `bopen settings`.
- Settings for the `window` mode, browser visibility and ordering, and enabling or disabling individual built-in rules.
- A registration status row with a re-register action, which runs the same logic as `bopen register`.
- Changes are saved to `config.toml` immediately and take effect in the inspector when the user returns to it.
- Built-in rules gain stable identifiers so they can be disabled by ID.

## Capabilities

### New Capabilities
- `settings-view`: The settings screen, how it is reached, what it contains, and how changes are saved.

### Modified Capabilities
- `preferences`: New keys for hidden browsers, browser order and disabled built-in rules; saving semantics for UI-written config.
- `link-inspector`: Cog entry point; the browser list honours hidden and ordered browsers.
- `link-cleaning`: Built-in rules have stable identifiers; disabled rules produce no suggestions.
- `default-handler-registration`: The command line accepts the `settings` subcommand.

## Impact

- Depends on `add-link-inspector-core`, which must be applied and archived first.
- Touches `internal/ui` (new view), `internal/prefs` (new keys, writing `config.toml`), `internal/clean/rules/builtin.toml` (IDs), and `cmd/bopen` (subcommand).
- `config.toml` becomes UI-owned: hand-written comments are not preserved when the settings view saves.
