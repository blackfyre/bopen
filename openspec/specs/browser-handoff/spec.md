# browser-handoff Specification

## Purpose

Launches the chosen browser with the final URL so that the URL can never be interpreted as anything other than a single web address.

## Requirements

### Requirement: Only web URLs are handed off
The system SHALL hand off only absolute URLs with scheme `http` or `https` and a non-empty host. The handed-off text SHALL be the link as received, with only these changes:
- surrounding whitespace is trimmed;
- the scheme is lower-cased;
- spaces, tabs and double quotes are percent-encoded;
- suggestions the user accepted are removed.

The argument therefore always begins with `http://` or `https://`. A malformed percent-escape SHALL NOT cause a link to be rejected.

#### Scenario: Option-like input rejected
- **WHEN** the input is `--gpu-launcher=calc.exe`
- **THEN** no browser is launched

#### Scenario: Missing host rejected
- **WHEN** the input is `https:///path`
- **THEN** no browser is launched

#### Scenario: Unchanged link passes through byte for byte
- **WHEN** the input is `https://example.com/café/ü?q=ö#` and no suggestion applies
- **THEN** the browser receives exactly `https://example.com/café/ü?q=ö#`

#### Scenario: Malformed escape still opens
- **WHEN** the input is `https://example.com/a%zz?utm_source=x`
- **THEN** the link is accepted, and accepting the suggestion hands off `https://example.com/a%zz`

### Requirement: Linux launch follows the desktop entry Exec line
On Linux the system SHALL build the command from the browser's `Exec` key as defined by the Desktop Entry specification, without invoking a shell. The URL SHALL replace `%u` or `%U` as exactly one argument, and other field codes SHALL be expanded or removed per the specification. When `Exec` contains no URL field code, the URL SHALL be appended as the final argument.

#### Scenario: Flatpak Exec line
- **WHEN** the entry's Exec is `/usr/bin/flatpak run --branch=stable --arch=x86_64 --command=launch-script.sh --file-forwarding app.zen_browser.zen @@u %u @@`
- **THEN** the process is started with those arguments and the URL as the single argument between `@@u` and `@@`

#### Scenario: Shell metacharacters stay literal
- **WHEN** the URL is `https://example.com/?a=$(id)&b=;ls`
- **THEN** the browser receives the URL as one unmodified argument, and no shell interprets it

### Requirement: Windows launch substitutes the registered command safely
On Windows the system SHALL build the command from the browser's registered `https` URL handler command (the ProgID named in its `Capabilities\URLAssociations`), falling back to the browser's `shell\open\command`. The URL SHALL be inserted where the command places its argument placeholder (`%1` or `%L`), keeping the template's own quoting, or appended when there is none. The browser SHALL receive the URL as exactly one argument. Because a handed-off URL contains no whitespace and no double quotes (they are percent-encoded), it can neither be split nor close a quote.

#### Scenario: Chrome single-argument command
- **WHEN** the command is `"C:\Program Files\Google\Chrome\Application\chrome.exe" --single-argument %1`
- **THEN** Chrome is started with `--single-argument` followed by the URL as one argument

#### Scenario: Embedded quote cannot break out
- **WHEN** the URL contains a `"` character
- **THEN** the character reaches the browser percent-encoded inside the single URL argument

### Requirement: Launched browser is detached
The launched browser process SHALL NOT be a child that bopen waits for. bopen SHALL exit after a successful launch, and closing bopen SHALL NOT terminate the browser.

#### Scenario: bopen exits promptly
- **WHEN** a browser is launched successfully
- **THEN** bopen terminates without waiting for the browser to exit

### Requirement: Last-used browser recorded on success only
After a successful launch the system SHALL record the selected browser's identity as last used. A failed launch or a cancelled inspector SHALL NOT change it.

#### Scenario: Successful launch updates state
- **WHEN** the user opens a link in Zen
- **THEN** the next invocation pre-selects Zen
