# default-handler-registration Specification

## Purpose

Makes bopen the system's handler for web links, and lets the user revert that, on Linux and Windows without administrator rights.

## Requirements

### Requirement: Register command on Linux
`bopen register` on Linux SHALL:
- install a desktop entry with ID `bopen.desktop` in `$XDG_DATA_HOME/applications`, whose `Exec` uses the absolute path of the running bopen executable and whose `MimeType` lists `x-scheme-handler/http` and `x-scheme-handler/https`;
- record the current default handler for `x-scheme-handler/https`, unless it is already bopen;
- make bopen the default handler for `x-scheme-handler/http` and `x-scheme-handler/https`.

#### Scenario: Registration on Linux
- **WHEN** the user runs `bopen register` while Brave is the default browser
- **THEN** `xdg-mime query default x-scheme-handler/https` reports `bopen.desktop`, and Brave is recorded as the previous default

#### Scenario: Repeated registration keeps the original previous default
- **WHEN** the user runs `bopen register` a second time
- **THEN** the recorded previous default is still Brave

### Requirement: Register command on Windows
`bopen register` on Windows SHALL write per-user registration under `HKEY_CURRENT_USER` only:
- a URL-handler ProgID whose open command runs the absolute path of bopen with the URL as one argument;
- a `StartMenuInternet\bopen` client with `Capabilities\URLAssociations` mapping `http` and `https` to that ProgID;
- an entry under `RegisteredApplications`.

It SHALL then open the Windows Default Apps settings for bopen, and SHALL tell the user that they must confirm bopen as the default there.

#### Scenario: Registration on Windows
- **WHEN** the user runs `bopen register`
- **THEN** bopen appears as a selectable web browser in Windows Default Apps settings, the settings page is opened, and no administrator prompt is shown

### Requirement: Unregister command
`bopen unregister` SHALL remove everything `bopen register` created. On Linux it SHALL restore the recorded previous default handler for `x-scheme-handler/http` and `x-scheme-handler/https` when that handler is still installed.

#### Scenario: Unregister on Linux restores previous default
- **WHEN** the user runs `bopen unregister` after registering while Brave was the default
- **THEN** `bopen.desktop` no longer exists in `$XDG_DATA_HOME/applications`, and Brave is the default handler again

#### Scenario: Unregister on Windows
- **WHEN** the user runs `bopen unregister`
- **THEN** the `HKCU` keys created by registration are removed

### Requirement: Command-line usage
bopen SHALL accept one URL argument (open mode), or one of the subcommands `register`, `unregister` or `settings`. Any other invocation, including no arguments, SHALL print usage to standard error and exit with a non-zero status.

#### Scenario: No arguments
- **WHEN** bopen is run without arguments
- **THEN** usage is printed, and the exit status is non-zero

#### Scenario: Settings subcommand
- **WHEN** bopen is run as `bopen settings`
- **THEN** the settings view opens
