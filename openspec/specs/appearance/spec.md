# appearance Specification

## Purpose

Gives bopen's window the Blackfyre design and makes it follow the host system's colour mode and contrast preference, while keeping every text legible.

## Requirements

### Requirement: Colour mode follows the system
bopen SHALL use a dark palette when the system prefers dark, and a light palette otherwise:
- **Linux:** the desktop portal's `org.freedesktop.appearance` `color-scheme` is `1` for dark, `2` for light, and `0` for no preference.
- **Windows:** `AppsUseLightTheme` is `0` for dark.

When the setting cannot be read, bopen SHALL use the light palette without delaying the window by more than 250 ms.

#### Scenario: Dark desktop
- **WHEN** the portal reports `color-scheme` `1`
- **THEN** the window uses the dark palette

#### Scenario: No portal
- **WHEN** no desktop portal answers on the session bus
- **THEN** the window opens with the light palette

### Requirement: High contrast is honoured
When the system requests high contrast, bopen SHALL use a high-contrast variant of the current palette. In that variant, secondary text uses the primary text colour, and borders and focus marks are drawn solid in the primary text colour. The system requests high contrast through the portal `contrast` value `1` on Linux, or the Windows high-contrast setting.

#### Scenario: High contrast on
- **WHEN** the portal reports `contrast` `1`
- **THEN** no text is drawn in a reduced-contrast secondary colour

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
