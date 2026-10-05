# Spec Delta

## MODIFIED Requirements

### Requirement: Window display follows the window preference
When bopen is invoked with a URL, it SHALL open the link without showing a window when a direct site rule applies (see `site-rules`). Otherwise it SHALL show the inspector window if the `window` preference is `always`, or if it is `when-suggestions` and analysis produced at least one suggestion. In the remaining cases it SHALL hand the unchanged URL to the pre-selected browser without showing a window.

#### Scenario: Always mode with clean link
- **WHEN** `window` is `always` and the link has no suggestions
- **THEN** the inspector window is shown

#### Scenario: When-suggestions mode with clean link
- **WHEN** `window` is `when-suggestions` and the link has no suggestions
- **THEN** no window is shown and the link opens in the pre-selected browser

#### Scenario: When-suggestions mode with tracking link
- **WHEN** `window` is `when-suggestions` and the link has a tracking parameter
- **THEN** the inspector window is shown

#### Scenario: Direct site rule in always mode
- **WHEN** `window` is `always` and a direct site rule applies
- **THEN** no window is shown and the link opens in the rule's browser

### Requirement: Browser choice with last-used pre-selection
The inspector SHALL list the discovered browsers with their display names and install kinds. The pre-selected browser SHALL be the first of the following that is currently discovered:
1. the browser of the applying site rule;
2. the last-used browser;
3. the browser that was the system default before bopen registered itself;
4. the first browser in the list.

#### Scenario: Last-used pre-selected
- **WHEN** the last-used browser is Zen, Zen is still installed, and no site rule applies
- **THEN** Zen is pre-selected

#### Scenario: Last-used uninstalled
- **WHEN** the last-used browser is no longer discovered, no site rule applies, and the previous system default was Brave
- **THEN** Brave is pre-selected

#### Scenario: Site rule wins
- **WHEN** a site rule for the link's host maps to Chrome and the last-used browser is Zen
- **THEN** Chrome is pre-selected

## ADDED Requirements

### Requirement: Remember the browser for a site
When no site rule applies, the inspector SHALL offer an "Always open `<host>` in this browser" option, where `<host>` is the destination host. If the option is on when the link is opened, bopen SHALL save a site rule mapping that host to the browser used. When a rule applies, the inspector SHALL instead show which rule chose the browser.

#### Scenario: Remembering a site
- **WHEN** the user ticks "Always open news.example.com in this browser" and opens the link in Zen
- **THEN** a site rule mapping `news.example.com` to Zen is saved, and the next link to that host pre-selects Zen

#### Scenario: Rule shown
- **WHEN** a site rule for `*.atlassian.net` applies
- **THEN** the inspector states that the browser was chosen by the rule for `*.atlassian.net`
