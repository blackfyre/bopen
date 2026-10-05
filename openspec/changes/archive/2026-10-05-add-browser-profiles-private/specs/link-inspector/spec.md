# Spec Delta

## ADDED Requirements

### Requirement: Private-window option
When the selected browser supports private windows, the inspector SHALL offer an "Open in a private window" option, off by default, toggled by the `P` key. When the option is on, opening SHALL use the browser's private window. The option SHALL NOT change the last-used browser.

#### Scenario: Toggling with the keyboard
- **WHEN** the selected browser supports private windows and the user presses `P` and then `Enter`
- **THEN** the link opens in that browser's private window

#### Scenario: Unsupported browser
- **WHEN** the selected browser has no private-window support
- **THEN** the option is not shown, and `P` does nothing
