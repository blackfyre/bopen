# Spec Delta

## ADDED Requirements

### Requirement: Flag a parameter from the inspector
Right-clicking a query parameter in the displayed URL, including parameters inside an unwrapped redirect target, SHALL offer "Always flag this parameter…". Choosing it SHALL open a form with these fields:
- **Parameter:** pre-filled with the parameter name and editable, so a trailing `*` can make it a prefix.
- **Scope:** "this host" (the host of the URL containing the parameter) or "any host".
- **Kind:** `tracking` (default) or `affiliate`.
- **Reason:** required.

Saving SHALL store the rule and re-analyse the link immediately, preserving toggles of suggestions that still exist.

#### Scenario: Flagging an unknown parameter
- **WHEN** the user right-clicks `ref=home` in `https://news.example.com/a?ref=home`, chooses "Always flag this parameter…", enters a reason and saves
- **THEN** a user rule for `ref` scoped to `news.example.com` is stored, and `ref=home` immediately appears as a highlighted suggestion

#### Scenario: Reason is mandatory
- **WHEN** the user tries to save the form with an empty reason
- **THEN** the rule is not saved, and the form indicates that a reason is required

### Requirement: Act on an existing suggestion
Right-clicking a suggestion SHALL offer "Disable this rule" for a built-in suggestion and "Edit this rule…" for a user suggestion. Disabling SHALL record the built-in rule as disabled and re-analyse the link.

#### Scenario: Disabling a built-in rule from the inspector
- **WHEN** the user right-clicks the `utm_source` suggestion and chooses "Disable this rule"
- **THEN** the `utm_*` rule is recorded as disabled and its suggestions disappear from the inspector
