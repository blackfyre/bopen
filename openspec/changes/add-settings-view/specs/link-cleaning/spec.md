# Spec Delta

## ADDED Requirements

### Requirement: Built-in rules have stable identifiers
Every built-in rule SHALL have a unique identifier that does not change between releases while the rule exists.

#### Scenario: Identifier uniqueness
- **WHEN** the built-in rule set is loaded
- **THEN** no two rules share an identifier

### Requirement: Disabled rules are not applied
A built-in rule whose identifier is listed as disabled SHALL NOT produce suggestions.

#### Scenario: Disabled tracking rule
- **WHEN** the `fbclid` rule is disabled and `https://example.com/?fbclid=x` is analysed
- **THEN** no suggestion is produced

#### Scenario: Disabled redirect rule
- **WHEN** the Google redirect rule is disabled and a Google wrapper URL is analysed
- **THEN** no redirect suggestion is produced
