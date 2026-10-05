# link-inspector Specification

## Purpose

The window that lets the user inspect an incoming link, review and toggle the suggested removals with their reasons, and choose which browser opens the result.

## Requirements

### Requirement: Window display follows the window preference
When bopen is invoked with a URL, it SHALL show the inspector window if the `window` preference is `always`, or if it is `when-suggestions` and analysis produced at least one suggestion. Otherwise it SHALL hand the unchanged URL to the pre-selected browser without showing a window.

#### Scenario: Always mode with clean link
- **WHEN** `window` is `always` and the link has no suggestions
- **THEN** the inspector window is shown

#### Scenario: When-suggestions mode with clean link
- **WHEN** `window` is `when-suggestions` and the link has no suggestions
- **THEN** no window is shown and the link opens in the pre-selected browser

#### Scenario: When-suggestions mode with tracking link
- **WHEN** `window` is `when-suggestions` and the link has a tracking parameter
- **THEN** the inspector window is shown

### Requirement: Original link is shown with highlighted suggestions
The inspector SHALL display the full original URL, visually highlighting each span covered by a suggestion and distinguishing the suggestion kinds.

#### Scenario: Highlighted parameter
- **WHEN** the inspector shows `https://example.com/?utm_source=x`
- **THEN** `utm_source=x` is visually highlighted within the displayed URL

### Requirement: Suggestions are listed with reasons and toggles
The inspector SHALL list every suggestion with its affected text, its reason and its source, each with a toggle initialised to its default acceptance state. A dependent suggestion SHALL be shown as unavailable while the redirect suggestion it depends on is rejected.

#### Scenario: Reason is visible
- **WHEN** the link contains `fbclid=abc`
- **THEN** the list shows `fbclid=abc` together with its reason

#### Scenario: Dependent suggestion disabled
- **WHEN** the user rejects a redirect suggestion
- **THEN** the suggestions found inside its target are shown as unavailable

### Requirement: Result URL updates live
The inspector SHALL display the cleaned URL and update it immediately whenever a toggle changes.

#### Scenario: Toggling updates the result
- **WHEN** the user rejects the `utm_source` suggestion
- **THEN** the displayed result URL includes `utm_source` again without any further action

### Requirement: Browser choice with last-used pre-selection
The inspector SHALL list the discovered browsers with their display names and install kinds. The pre-selected browser SHALL be the first of the following that is currently discovered:
1. the last-used browser;
2. the browser that was the system default before bopen registered itself;
3. the first browser in the list.

#### Scenario: Last-used pre-selected
- **WHEN** the last-used browser is Zen and Zen is still installed
- **THEN** Zen is pre-selected

#### Scenario: Last-used uninstalled
- **WHEN** the last-used browser is no longer discovered and the previous system default was Brave
- **THEN** Brave is pre-selected

### Requirement: Keyboard operation
The inspector SHALL support these keys:
- `Enter` opens the result URL in the selected browser.
- `Escape` closes the window without opening anything.
- `Up` and `Down` move the browser selection.
- `1` to `9` select the corresponding browser in the list.

Every action SHALL also be available with the mouse.

#### Scenario: Enter opens immediately
- **WHEN** the inspector opens and the user presses `Enter`
- **THEN** the result URL opens in the pre-selected browser and the window closes

#### Scenario: Escape cancels
- **WHEN** the user presses `Escape`
- **THEN** the window closes, no browser is launched, and the last-used state is unchanged

### Requirement: Unusable input and environment are reported
The inspector SHALL be shown with an explanatory message, and SHALL NOT offer to open a browser, when:
- the input is not an absolute `http` or `https` URL;
- no browsers were discovered;
- launching the selected browser fails (in this case the window stays open so another browser can be chosen).

This SHALL apply regardless of the `window` preference.

#### Scenario: Non-web scheme
- **WHEN** bopen is invoked with `file:///etc/passwd`
- **THEN** the window explains that only http and https links are accepted, and no browser is launched

#### Scenario: Launch failure in silent mode
- **WHEN** `window` is `when-suggestions`, the link is clean, and launching the pre-selected browser fails
- **THEN** the inspector window is shown with the error

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
