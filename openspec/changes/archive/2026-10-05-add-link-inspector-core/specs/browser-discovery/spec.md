# Spec Delta

## Purpose

Builds the list of web browsers installed on the machine, so the user can choose a target for each link without configuring browsers by hand.

## ADDED Requirements

### Requirement: Linux browsers are discovered from desktop entries
On Linux the system SHALL discover browsers from desktop entry files (`*.desktop`) in these directories, in this precedence order:
1. `$XDG_DATA_HOME/applications` (default `~/.local/share/applications`)
2. `~/.local/share/flatpak/exports/share/applications`
3. each `<dir>/applications` for `<dir>` in `$XDG_DATA_DIRS` (default `/usr/local/share:/usr/share`)
4. `/var/lib/flatpak/exports/share/applications`
5. `/var/lib/snapd/desktop/applications`

An entry SHALL qualify as a browser when its `MimeType` key lists `x-scheme-handler/http` or `x-scheme-handler/https`, its `Type` is `Application`, and neither `Hidden` nor `NoDisplay` is `true`.

#### Scenario: System-installed browser is found
- **WHEN** `/usr/share/applications/brave-browser.desktop` declares `MimeType=...;x-scheme-handler/https;...`
- **THEN** the browser list contains an entry for it with the name from the entry's `Name` key

#### Scenario: Flatpak browser is found without XDG_DATA_DIRS
- **WHEN** `app.zen_browser.zen.desktop` exists in `/var/lib/flatpak/exports/share/applications` and `$XDG_DATA_DIRS` does not include the Flatpak export directory
- **THEN** the browser list still contains an entry for it

#### Scenario: Non-browser application is ignored
- **WHEN** a desktop entry declares no `x-scheme-handler/http` or `x-scheme-handler/https` MIME type
- **THEN** it does not appear in the browser list

#### Scenario: Hidden entry is ignored
- **WHEN** a qualifying desktop entry has `Hidden=true` or `NoDisplay=true`
- **THEN** it does not appear in the browser list

### Requirement: Linux entries are deduplicated by desktop file ID
The system SHALL identify a Linux browser by its desktop file ID, as defined by the Desktop Entry specification. When the same ID exists in several directories, only the entry from the highest-precedence directory SHALL be used, including when that entry disqualifies the browser.

#### Scenario: User override shadows system entry
- **WHEN** `firefox.desktop` exists in both `~/.local/share/applications` and `/usr/share/applications`
- **THEN** the browser list contains one Firefox entry, built from the file in `~/.local/share/applications`

#### Scenario: User override hides a browser
- **WHEN** `~/.local/share/applications/firefox.desktop` sets `Hidden=true` and `/usr/share/applications/firefox.desktop` qualifies
- **THEN** Firefox does not appear in the browser list

### Requirement: Entries whose executable is missing are excluded
On Linux the system SHALL exclude an entry whose `TryExec` key names a program that does not exist or cannot be found on `PATH`.

#### Scenario: Uninstalled leftover entry
- **WHEN** a qualifying entry has `TryExec=/opt/oldbrowser/browser` and that file does not exist
- **THEN** it does not appear in the browser list

### Requirement: Windows browsers are discovered from registered clients
On Windows the system SHALL discover browsers from the subkeys of `SOFTWARE\Clients\StartMenuInternet` under `HKEY_CURRENT_USER` and `HKEY_LOCAL_MACHINE`. A subkey SHALL qualify when it has a `shell\open\command` default value. The subkey name SHALL be the browser's identity; when it exists under both hives, the `HKEY_CURRENT_USER` subkey SHALL take precedence. The display name SHALL be the subkey's default value, falling back to the subkey name.

#### Scenario: Machine-wide browser is found
- **WHEN** `HKLM\SOFTWARE\Clients\StartMenuInternet\Google Chrome\shell\open\command` exists
- **THEN** the browser list contains an entry for Google Chrome

#### Scenario: Per-user install takes precedence
- **WHEN** the same subkey name exists under both `HKCU` and `HKLM`
- **THEN** the browser list contains one entry, built from the `HKCU` subkey

### Requirement: bopen never lists itself
The system SHALL exclude its own registration (the desktop file ID `bopen.desktop` on Linux, the `StartMenuInternet` subkey `bopen` on Windows) from the browser list.

#### Scenario: Registered bopen is not a target
- **WHEN** bopen is registered as a browser
- **THEN** bopen does not appear in its own browser list

### Requirement: Browser entries expose identity, name and install kind
Each browser entry SHALL expose a stable identity, a display name, and an install kind. On Linux the kind SHALL be `flatpak` for entries from a Flatpak export directory, `snap` for entries from the Snap directory, `user` for entries from `$XDG_DATA_HOME/applications`, and `system` otherwise. On Windows the kind SHALL be `user` for `HKCU` entries and `system` for `HKLM` entries. The list SHALL be sorted by display name, case-insensitively.

#### Scenario: Same browser from two sources is distinguishable
- **WHEN** Firefox is installed both as `firefox.desktop` (system) and as `org.mozilla.firefox.desktop` (Flatpak)
- **THEN** the browser list contains two entries with different identities and kinds `system` and `flatpak`
