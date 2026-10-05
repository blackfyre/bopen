# Spec Delta

## Purpose

Defines where bopen keeps user preferences and runtime state, their format, and how missing or invalid files are handled.

## ADDED Requirements

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
