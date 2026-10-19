# Proposal

## Why

bopen, the Spooder desktop calibrator and blackfyre.ninja are one family, but bopen still draws a generic material look with the system accent colour and the Go fonts. The calibrator already uses the Blackfyre design of blackfyre.ninja (Spooder's `add-desktop-calibrator`, design D7), and copied bopen's appearance code to do it. bopen should look the same.

## What Changes

- The Blackfyre palettes replace bopen's own: paper and ink in light mode, the dark surfaces of the site's dark themes, and the orange accent `#e2551d` with ink text. **BREAKING:** the system accent colour is no longer used.
- Archivo SemiBold for text and JetBrains Mono for links, patterns, labels and buttons, embedded with their OFL licences. Section titles, chips and button labels are uppercase JetBrains Mono with wide letter-spacing.
- Tactile buttons: a flat face on a darker 2 dp shadow that flattens on press.
- The suggestion-kind and banner colours are taken from the site's tokens and stay at 4.5:1 or more.
- Unchanged: following the system's light or dark mode, the high-contrast variants, live appearance changes and the Windows title bar.

## Capabilities

### Modified Capabilities
- `appearance`: the Blackfyre design replaces the system accent colour.

## Impact

`internal/ui` only (palette, styles, a new `internal/ui/fonts` package with two TTF files and their licences, about 450 KB). `internal/appearance` keeps reading the accent colour, unused, so its code stays shared with the calibrator. The README's appearance paragraph changes.
