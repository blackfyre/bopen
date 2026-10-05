# Spec Delta

## ADDED Requirements

### Requirement: Browser visibility and order keys
`config.toml` SHALL support a `[browsers]` table with `hidden`, a list of browser identities not offered in the inspector, and `order`, a list of browser identities shown first in that order. Browsers not listed in `order` SHALL follow in the default sort order. Identities of browsers that are not currently discovered SHALL be kept in the file and ignored.

#### Scenario: Ordered and unordered browsers
- **WHEN** `order = ["app.zen_browser.zen.desktop"]` and Brave and Firefox are also discovered
- **THEN** the inspector lists Zen, then Brave, then Firefox

#### Scenario: Uninstalled browser stays configured
- **WHEN** a hidden browser is uninstalled and later reinstalled
- **THEN** it is hidden again without user action

### Requirement: Disabled rules key
`config.toml` SHALL support a `[rules]` table with `disabled`, a list of built-in rule identifiers that SHALL NOT produce suggestions. Unknown identifiers SHALL be ignored.

#### Scenario: Unknown rule identifier
- **WHEN** `disabled` contains an identifier that no built-in rule has
- **THEN** bopen runs normally and all built-in rules stay enabled

### Requirement: UI writes preserve concurrent changes
When the settings view saves, it SHALL re-read `config.toml`, apply only the changed setting, and write the result atomically, so changes saved by another running bopen instance are not lost.

#### Scenario: Two instances
- **WHEN** one bopen instance hides a browser and a second, already-open instance then changes the window mode
- **THEN** `config.toml` contains both changes
