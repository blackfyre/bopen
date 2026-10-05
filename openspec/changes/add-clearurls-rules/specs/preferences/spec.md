# Spec Delta

## ADDED Requirements

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
