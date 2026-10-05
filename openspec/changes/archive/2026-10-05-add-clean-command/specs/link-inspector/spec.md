# Spec Delta

## ADDED Requirements

### Requirement: Copy the cleaned link
The inspector SHALL offer a Copy action next to the result, also triggered by the `C` key, which places the result URL on the clipboard and confirms this in the window, without opening a browser or closing the window.

#### Scenario: Copying
- **WHEN** the user presses `C`
- **THEN** the result URL is on the clipboard, the window shows "Copied", and no browser is launched
