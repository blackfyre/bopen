# Design

## Context

See proposal.md. On this machine:
- every browser's desktop entry declares a `new-private-window` action;
- Zen (Flatpak, `~/.var/app/app.zen_browser.zen/.zen/profiles.ini`) and Firefox (`~/.config/mozilla/firefox/profiles.ini`) each have two profiles;
- Chrome, Brave and Chromium each have one profile in `Local State`.

Desktop entry parsing currently stops at the first `[Desktop Action]` group.

## Decisions

- **Desktop actions.** `desktopentry.Entry` gains `Actions map[string]Entry` holding the `Name` and `Exec` of each action group listed in `Actions`. The parser keeps reading after `[Desktop Entry]`, for those groups only.
- **Families.** A table in `discovery` maps each family to its profile kind (firefox or chromium) and its locations:
  - **Linux:** by desktop ID or `Exec` program basename, with per-install-kind directories (native, Flatpak `~/.var/app/<id>/…`, Snap `~/snap/<name>/common/…`, and Firefox's XDG `~/.config/mozilla/firefox` before `~/.mozilla/firefox`). The first existing directory for the kind wins.
  - **Windows:** by executable basename (`chrome.exe`, `msedge.exe`, `brave.exe`, `vivaldi.exe`, `firefox.exe`, `zen.exe`, `librewolf.exe`) under `%LOCALAPPDATA%`/`%APPDATA%`. The same table holds the Windows private-window flag.
- **Browser fields.** `discovery.Browser` gains:
  - `PrivateEntry *desktopentry.Entry` (Linux) and `PrivateFlag string` (Windows);
  - `ProfileArgs []string` and `Base string` (the identity of the browser a profile entry belongs to);
  - `OpenPrivate bool`, set on a copy just before launching, so the `Open` function's signature doesn't change.
- **Expansion.** `discovery.Expand(browsers, fs)` runs after discovery. It reads profile stores through an injectable file system root, so it is testable with fixture trees. Profile names in `profiles.ini` are sorted by name, and Chromium profiles by display name.
- **Launch.**
  - **Linux:** build argv from `PrivateEntry` when opening privately (its `Exec` without a URL code gets the URL appended), else from `Entry`. Then insert `ProfileArgs` immediately before the URL argument, or before a preceding `@@u`.
  - **Windows:** `WindowsCommandLine` takes extra arguments (the private flag, then the profile arguments), inserted directly after the executable, each quoted when it contains spaces. They must precede a Chrome-style `--single-argument`, which takes the rest of the line.
- **Inspector.** `Model.Private` is reset whenever the selection changes to a browser without support. It is set by a checkbox in the Open-in card and by the `P` key. `OpenSelected` launches a copy of the browser with `OpenPrivate` set.

## Risks / Trade-offs

- **[Unknown browsers]** → Families not in the table get no profile entries. Private support on Linux still works through the desktop action.
- **[Profile stores change format]** → Parse errors yield no profile entries, never an error dialog.
