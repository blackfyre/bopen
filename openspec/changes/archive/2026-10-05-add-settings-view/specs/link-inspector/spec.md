# Spec Delta

## ADDED Requirements

### Requirement: Settings entry point
The inspector window SHALL show a cog control that opens the settings view.

#### Scenario: Cog opens settings
- **WHEN** the user activates the cog control
- **THEN** the settings view is shown in the same window

### Requirement: Browser list honours visibility and order
The inspector SHALL NOT offer hidden browsers, and SHALL order the remaining browsers by the configured order. Pre-selection SHALL skip any candidate that is hidden. When all discovered browsers are hidden, the inspector SHALL report that no browsers are available and offer the settings view.

#### Scenario: Last-used browser hidden
- **WHEN** the last-used browser is Firefox and Firefox is hidden
- **THEN** Firefox is not offered, and pre-selection continues with the next candidate

#### Scenario: All browsers hidden
- **WHEN** every discovered browser is hidden
- **THEN** the inspector reports that no browsers are available and does not offer to open the link
