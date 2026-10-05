# Spec Delta

## ADDED Requirements

### Requirement: Short-link expansion preference
`config.toml` SHALL support a top-level boolean `expand_short_links`, defaulting to `false`, that enables the Expand action for known shorteners.

#### Scenario: Default
- **WHEN** `config.toml` has no `expand_short_links` key
- **THEN** short-link expansion is disabled
