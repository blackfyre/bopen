# Spec Delta

## ADDED Requirements

### Requirement: Short-link expansion setting
The settings view SHALL offer a toggle for short-link expansion, stating that expanding contacts the shortener when the user presses Expand.

#### Scenario: Enabling
- **WHEN** the user enables the toggle
- **THEN** `config.toml` contains `expand_short_links = true`
