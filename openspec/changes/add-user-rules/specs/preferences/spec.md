# Spec Delta

## ADDED Requirements

### Requirement: User rules are stored in config
`config.toml` SHALL store user rules as an array of `[[rules.user]]` tables with the fields `id`, `kind`, `param`, `hosts` and `reason`. `id` SHALL be unique among user rules and assigned by bopen when a rule is created. Saving a rule change SHALL follow the same read-modify-write semantics as other settings.

#### Scenario: Rule persisted
- **WHEN** the user saves a new rule for `ref`
- **THEN** `config.toml` contains a `[[rules.user]]` entry with a new unique `id`, `param = "ref"`, and the entered reason
