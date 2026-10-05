# preferences Specification

## Purpose

Defines where bopen keeps user preferences and runtime state, their format, and how missing or invalid files are handled.

## Requirements

### Requirement: Files live in the OS user configuration directory
The system SHALL store preferences in `config.toml` and runtime state in `state.toml`, both in a `bopen` directory inside the OS user configuration directory (`%AppData%` on Windows, `$XDG_CONFIG_HOME` or `~/.config` on Linux). The directory SHALL be created when bopen first writes to it.

#### Scenario: Linux location
- **WHEN** `XDG_CONFIG_HOME` is unset on Linux
- **THEN** preferences are read from `~/.config/bopen/config.toml`

#### Scenario: Windows location
- **WHEN** bopen runs on Windows
- **THEN** preferences are read from `%AppData%\bopen\config.toml`

### Requirement: Window preference
`config.toml` SHALL support a top-level key `window`, with values `always` or `when-suggestions`, defaulting to `always`.

#### Scenario: Default without config file
- **WHEN** `config.toml` does not exist
- **THEN** bopen behaves as if `window = "always"`

### Requirement: Invalid preferences fall back safely
When `config.toml` cannot be parsed, or contains an invalid value, bopen SHALL use the defaults for the affected settings and SHALL show the problem in the inspector window, showing the window even when it would otherwise be skipped. Unknown keys SHALL be ignored.

#### Scenario: Malformed config
- **WHEN** `config.toml` contains invalid TOML
- **THEN** the inspector is shown with a message naming the file and the parse error, and `window` is treated as `always`

#### Scenario: Unknown key
- **WHEN** `config.toml` contains `colour = "blue"`
- **THEN** bopen runs normally and ignores the key

### Requirement: State is separate and written atomically
The last-used browser identity and the previous system default SHALL be stored in `state.toml`, not in `config.toml`. Writes to `state.toml` SHALL be atomic, so that an interrupted write leaves the previous contents intact. An unreadable `state.toml` SHALL be treated as empty.

#### Scenario: Launch does not touch preferences
- **WHEN** a link is opened in a browser
- **THEN** `config.toml` is not modified

#### Scenario: Corrupt state
- **WHEN** `state.toml` contains invalid TOML
- **THEN** bopen runs with no last-used browser and overwrites the file on the next successful launch

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

### Requirement: User rules are stored in config
`config.toml` SHALL store user rules as an array of `[[rules.user]]` tables with the fields `id`, `kind`, `param`, `hosts` and `reason`. `id` SHALL be unique among user rules and assigned by bopen when a rule is created. Saving a rule change SHALL follow the same read-modify-write semantics as other settings.

#### Scenario: Rule persisted
- **WHEN** the user saves a new rule for `ref`
- **THEN** `config.toml` contains a `[[rules.user]]` entry with a new unique `id`, `param = "ref"`, and the entered reason

### Requirement: ClearURLs preference
`config.toml` SHALL support `[rules] clearurls`, a boolean defaulting to `false`, that enables the ClearURLs rule source and its network access.

#### Scenario: Default disabled
- **WHEN** `config.toml` has no `clearurls` key
- **THEN** ClearURLs is disabled

### Requirement: Cache directory
Downloaded rule data SHALL be stored in a `bopen` directory inside the OS user cache directory (`%LocalAppData%` on Windows, `$XDG_CACHE_HOME` or `~/.cache` on Linux), never in the configuration directory.

#### Scenario: Linux cache location
- **WHEN** `XDG_CACHE_HOME` is unset on Linux
- **THEN** the ClearURLs cache is `~/.cache/bopen/clearurls.json`

### Requirement: Site rules are stored in config
`config.toml` SHALL store site rules as an array of `[[sites]]` tables with the fields `id`, `hosts` (a list of patterns), `browser` (a browser identity) and `direct` (a boolean, default `false`), in matching order. `id` SHALL be assigned by bopen and unique among site rules. Saving SHALL follow the same read-modify-write semantics as other settings.

#### Scenario: Rule persisted
- **WHEN** the user saves a site rule for `*.atlassian.net` with Chrome
- **THEN** `config.toml` contains a `[[sites]]` entry with a new unique `id`, `hosts = ["*.atlassian.net"]` and Chrome's identity

### Requirement: Short-link expansion preference
`config.toml` SHALL support a top-level boolean `expand_short_links`, defaulting to `false`, that enables the Expand action for known shorteners.

#### Scenario: Default
- **WHEN** `config.toml` has no `expand_short_links` key
- **THEN** short-link expansion is disabled
