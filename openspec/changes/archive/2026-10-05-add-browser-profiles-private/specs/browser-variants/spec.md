# Spec Delta

## Purpose

Lets the user open links in a specific browser profile or in a private window, without leaving bopen.

## ADDED Requirements

### Requirement: Profiles are offered as browsers
For a discovered browser of a known family with two or more profiles, bopen SHALL add one browser entry per profile, after the browser's own entry. Each profile entry SHALL have:
- the display name `<browser> · <profile name>`;
- the same install kind as the browser;
- the identity `<browser identity>#profile=<profile key>`, stable across runs.

Opening a profile entry SHALL start the browser in that profile. Browsers with a single profile SHALL get no extra entries.

- **Firefox family (Firefox, Zen, LibreWolf):** profiles come from `profiles.ini` in the profile directory matching the install kind. The key is the profile `Path`, and the browser is launched with `-P <name>`.
- **Chromium family (Chrome, Chromium, Brave, Edge, Vivaldi):** profiles come from `profile.info_cache` in `Local State`. The key is the profile directory, and the browser is launched with `--profile-directory=<dir>`.

#### Scenario: Two Zen profiles
- **WHEN** the Flatpak Zen's `profiles.ini` lists the profiles "Default Profile" and "Default (release)"
- **THEN** the browser list contains "Zen Browser", "Zen Browser · Default (release)" and "Zen Browser · Default Profile"

#### Scenario: Single Chrome profile
- **WHEN** Chrome's `Local State` lists one profile
- **THEN** only "Google Chrome" is listed

#### Scenario: Profile arguments in a Flatpak launch
- **WHEN** a Flatpak Zen profile entry opens a link
- **THEN** `-P <name>` is passed before the `@@u` URL wrapper, so it is not mistaken for a URL

### Requirement: Private windows
A browser SHALL support private windows when, on Linux, its desktop entry declares a private-window action (an action identifier containing `private` or `incognito`), or, on Windows, its executable is a known browser with a private-window flag. Opening privately SHALL launch the browser's private window with the link, as exactly one URL argument. For a profile entry, it SHALL do so in that profile.

#### Scenario: Firefox private window on Linux
- **WHEN** Firefox's entry declares `new-private-window` with `Exec=firefox --private-window %u` and the user opens privately
- **THEN** Firefox is started with `--private-window` and the link

#### Scenario: Edge on Windows
- **WHEN** the browser executable is `msedge.exe` and the user opens privately
- **THEN** Edge is started with `--inprivate` before its own arguments
