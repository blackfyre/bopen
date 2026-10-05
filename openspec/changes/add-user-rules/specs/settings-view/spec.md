# Spec Delta

## ADDED Requirements

### Requirement: User rule management
The settings view SHALL list all user rules with their parameter, scope, kind and reason. It SHALL allow adding a rule, editing any field of an existing rule, and deleting a rule, using the same field validation as the inspector form.

#### Scenario: Deleting a user rule
- **WHEN** the user deletes the `ref` rule in settings
- **THEN** the rule is removed from `config.toml` and no longer produces suggestions

#### Scenario: Editing a reason
- **WHEN** the user changes a rule's reason in settings
- **THEN** the inspector shows the new reason for that rule's suggestions
