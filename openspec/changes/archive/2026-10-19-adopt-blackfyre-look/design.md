# Design

## Context

Spooder's desktop calibrator (`desktop/internal/ui`) implements the Blackfyre look in Gio: `theme.go` (tokens), `widgets.go` (`tracked` labels, `tactileButton`) and `fonts/` (Archivo SemiBold, JetBrains Mono Regular). Its `theme.go` was itself derived from bopen's. The site's stylesheet (`legendary-winner/assets/css/styles.css`) is the source of the tokens.

## Decisions

### Tokens

As in the calibrator:
- light: `Bg` paper `#e6e4df`, `Surface` `#f3f2ee`, `SurfaceAlt` `#dedbd4`, `Fg` ink `#1b1b19`, `Muted` `#5e5c56`, `Border` `#d3d0c9`;
- dark: `Bg` `#1b1b19`, `Surface` `#2c2c29`, `SurfaceAlt` and `Border` `#3a3935`, `Fg` `#f3f2ee`, `Muted` `#a19e96`;
- accent `#e2551d`, pressed `#a93a0f`, text on it ink (white is 3.8:1).

bopen's own roles, from the site where it has them:
- **Tracking:** the accent as text, `#a3390c` light, `#ff8a4c` dark.
- **Affiliate:** `#6e5300` (the site's review-note heading) light, `#e0af68` dark.
- **Redirect:** the calibrator's blue `#7aa2f7` dark; a darker blue light, since the site has none for text on paper.
- **Warning banner:** the site's note paper `#fff6c9` with `#6e5300`; a dark amber well in dark mode.
- **Error banner:** an orange-tinted paper with `#7f2c09`; a dark orange well in dark mode.

The exact light blue and dark banner shades are chosen by the legibility test, not by the site.

### High contrast

The calibrator has no high-contrast mode; bopen keeps its own: black or white backgrounds, `Muted` and `Border` equal to `Fg`. The accent stays orange; it reaches 3:1 on both black and white.

### Type and buttons

The fonts package and `tracked` are copied from the calibrator. Gio has no letter-spacing, so `tracked` lays a label out a glyph at a time, 0.08 em apart. Button kinds map bopen's three buttons: primary (Open, Save) is the accent; secondary (Cancel, Add rule) is the neutral ink button; small (Copy, Edit, Up) is the quiet button. A disabled button has no shadow and muted text on `SurfaceAlt`.

Only JetBrains Mono Regular is embedded, so accepted parts of the link are no longer bold; their kind colour marks them.

## Risks / Trade-offs

- [Users lose their accent colour] → Deliberate: the family shares one accent. Light/dark and high contrast still follow the system.
- [Binary size] → About 450 KB of fonts.
- [Glyph-by-glyph labels lose kerning] → Only short uppercase labels are tracked; body text and links use normal shaping.
