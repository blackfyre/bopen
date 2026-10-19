# Spec Delta

## ADDED Requirements

### Requirement: Blackfyre design
bopen SHALL use the Blackfyre design of blackfyre.ninja, as the Spooder desktop calibrator does:
- its light (paper `#e6e4df`, ink `#1b1b19`) and dark (`#1b1b19`) palettes;
- the orange accent `#e2551d` for primary actions, selection marks and focus highlights, with ink text on it;
- Archivo for text and JetBrains Mono for links, rule patterns, labels and buttons, embedded in bopen;
- tactile buttons: a flat face on a darker shadow that flattens while pressed.

bopen SHALL NOT use the system accent colour.

#### Scenario: Blue system accent
- **WHEN** the portal reports `accent-color` `(0.478, 0.635, 0.969)`
- **THEN** the Open button is drawn in `#e2551d` with ink text

## MODIFIED Requirements

### Requirement: Appearance changes apply while open
bopen SHALL apply a changed colour mode or contrast preference without being restarted:
- **Linux:** when the portal emits `SettingChanged` for `org.freedesktop.appearance`.
- **Windows:** when the window regains focus.

#### Scenario: Switching to dark while the inspector is open
- **WHEN** the inspector is open and the user switches the desktop to dark mode
- **THEN** the inspector redraws with the dark palette

### Requirement: Text stays legible
In every palette (light, dark, and the high-contrast variant of each), the contrast ratio SHALL be at least 4.5:1 for:
- every text colour against the background it is drawn on, including the suggestion-kind colours and banner text;
- the text of every button kind against its face.

#### Scenario: Kind colours on dark background
- **WHEN** the dark palette is used
- **THEN** tracking, affiliate and redirect highlight colours each have at least 4.5:1 contrast against the card background

## REMOVED Requirements

### Requirement: Accent colour follows the system
**Reason:** bopen uses the Blackfyre accent, shared with the Spooder calibrator and blackfyre.ninja.
**Migration:** None; the colour mode and contrast preference still follow the system.
