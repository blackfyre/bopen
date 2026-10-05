# settings-view Specification

## Purpose

Lets the user view and change every bopen preference from inside the application, so `config.toml` never has to be edited by hand.

## Requirements

### Requirement: Settings view is reachable
The settings view SHALL be reachable from a cog control in the inspector window and by running `bopen settings`. When opened from the inspector, a back control SHALL return to the inspector with the link, the suggestion toggles and the browser selection preserved. When opened with `bopen settings`, closing the view SHALL exit bopen without launching a browser.

#### Scenario: Opening from the inspector and returning
- **WHEN** the user has rejected one suggestion, opens settings with the cog, and then goes back
- **THEN** the inspector shows the same link with that suggestion still rejected

#### Scenario: Standalone settings
- **WHEN** the user runs `bopen settings`
- **THEN** the settings view opens without an inspector, and closing it launches no browser

### Requirement: Window mode setting
The settings view SHALL let the user choose the `window` preference between `always` and `when-suggestions`.

#### Scenario: Switching window mode
- **WHEN** the user selects "only when there are suggestions"
- **THEN** `config.toml` contains `window = "when-suggestions"`

### Requirement: Browser visibility and order settings
The settings view SHALL list every discovered browser with its display name and install kind, a control to hide or show it, and controls to move it up or down.

#### Scenario: Hiding a browser
- **WHEN** the user hides Firefox
- **THEN** Firefox is recorded as hidden and is no longer offered in the inspector

#### Scenario: Reordering
- **WHEN** the user moves Zen above Brave
- **THEN** the inspector lists Zen before Brave

### Requirement: Built-in rule settings
The settings view SHALL list every built-in rule with its kind, its parameter or host pattern and its reason, each with a control to enable or disable it.

#### Scenario: Disabling a rule
- **WHEN** the user disables the `utm_*` rule
- **THEN** links containing `utm_source` no longer produce a suggestion for it

### Requirement: Registration status
The settings view SHALL show whether bopen is currently the default handler for `https` links, where the platform allows bopen to determine it, and SHALL offer an action that performs the same registration as `bopen register`.

#### Scenario: Re-registering after another browser took over
- **WHEN** another browser has made itself the default and the user activates re-register on Linux
- **THEN** bopen is the default handler for `http` and `https` again

### Requirement: Changes are saved immediately
Every change made in the settings view SHALL be written to `config.toml` immediately, with no separate save action.

#### Scenario: Change survives closing
- **WHEN** the user changes a setting and closes bopen
- **THEN** the next invocation uses the changed setting

### Requirement: User rule management
The settings view SHALL list all user rules with their parameter, scope, kind and reason. It SHALL allow adding a rule, editing any field of an existing rule, and deleting a rule, using the same field validation as the inspector form.

#### Scenario: Deleting a user rule
- **WHEN** the user deletes the `ref` rule in settings
- **THEN** the rule is removed from `config.toml` and no longer produces suggestions

#### Scenario: Editing a reason
- **WHEN** the user changes a rule's reason in settings
- **THEN** the inspector shows the new reason for that rule's suggestions

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

### Requirement: Site rule management
The settings view SHALL list the site rules in matching order, with their host patterns, the browser name and whether they open directly. It SHALL allow adding, editing and deleting rules. The rule form SHALL require at least one valid host pattern and a browser, chosen from the discovered browsers.

#### Scenario: Adding a direct rule
- **WHEN** the user adds a rule for `*.atlassian.net` with Chrome and "Open directly" ticked
- **THEN** the rule is saved, and the next matching link opens in Chrome without the inspector

#### Scenario: Deleting a rule
- **WHEN** the user deletes a site rule
- **THEN** matching links no longer pre-select its browser
