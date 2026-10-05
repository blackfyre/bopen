# Proposal

## Why

The most common reason to use a browser picker is to send certain sites to a certain browser every time: work tools to the work browser, banking to a hardened profile. Today bopen only pre-selects the last-used browser, so every link to such a site needs a manual choice.

## What Changes

- **Site rules:** host patterns (for example `*.atlassian.net`) mapped to a browser. The first rule matching the link's destination host decides the pre-selected browser, ahead of the last-used browser.
- **Open directly:** a per-rule option that skips the inspector for matching links. They open at once in that browser with the default suggestions applied (tracking removed, affiliate tags kept).
- **From the inspector:** an "Always open `<host>` in this browser" checkbox creates a rule when the link is opened. A link that already matches a rule says which rule chose the browser.
- **In settings:** a "Site rules" card to add, edit and delete rules.

## Capabilities

### New Capabilities
- `site-rules`: Matching links to browsers by host, and opening them directly.

### Modified Capabilities
- `link-inspector`: Window display honours direct site rules; pre-selection puts the site rule first; the remember checkbox.
- `preferences`: Storage of site rules.
- `settings-view`: Site rule management.

## Impact

Touches `internal/prefs` (`[[sites]]`), `internal/app` (matching, pre-selection, window decision), `internal/ui` (Open-in card, settings card, site form) and `cmd/bopen` (silent path). No change to rules or link cleaning.
