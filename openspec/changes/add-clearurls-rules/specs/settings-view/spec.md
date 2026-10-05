# Spec Delta

## ADDED Requirements

### Requirement: ClearURLs settings
The settings view SHALL provide a ClearURLs section containing:
- an enable toggle, which states that enabling it downloads data from the ClearURLs project;
- the last successful update time;
- the last error, if any;
- the number of skipped patterns, if any;
- an "Update now" action, available while enabled;
- attribution: the ClearURLs project name, its LGPL-3.0 licence, and a link to the project.

Enabling the toggle SHALL start an immediate fetch.

#### Scenario: Enabling starts a download
- **WHEN** the user enables ClearURLs
- **THEN** `clearurls = true` is saved, a fetch starts, and the section shows its outcome when it completes

#### Scenario: Attribution visible
- **WHEN** the ClearURLs section is shown
- **THEN** it names the ClearURLs project and its licence, and links to the project
