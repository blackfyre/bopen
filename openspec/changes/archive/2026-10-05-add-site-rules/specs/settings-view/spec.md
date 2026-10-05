# Spec Delta

## ADDED Requirements

### Requirement: Site rule management
The settings view SHALL list the site rules in matching order, with their host patterns, the browser name and whether they open directly. It SHALL allow adding, editing and deleting rules. The rule form SHALL require at least one valid host pattern and a browser, chosen from the discovered browsers.

#### Scenario: Adding a direct rule
- **WHEN** the user adds a rule for `*.atlassian.net` with Chrome and "Open directly" ticked
- **THEN** the rule is saved, and the next matching link opens in Chrome without the inspector

#### Scenario: Deleting a rule
- **WHEN** the user deletes a site rule
- **THEN** matching links no longer pre-select its browser
