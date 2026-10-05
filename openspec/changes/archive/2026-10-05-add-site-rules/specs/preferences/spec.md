# Spec Delta

## ADDED Requirements

### Requirement: Site rules are stored in config
`config.toml` SHALL store site rules as an array of `[[sites]]` tables with the fields `id`, `hosts` (a list of patterns), `browser` (a browser identity) and `direct` (a boolean, default `false`), in matching order. `id` SHALL be assigned by bopen and unique among site rules. Saving SHALL follow the same read-modify-write semantics as other settings.

#### Scenario: Rule persisted
- **WHEN** the user saves a site rule for `*.atlassian.net` with Chrome
- **THEN** `config.toml` contains a `[[sites]]` entry with a new unique `id`, `hosts = ["*.atlassian.net"]` and Chrome's identity
