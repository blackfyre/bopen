# Tasks

## 1. Discovery

- [x] 1.1 Parse `[Desktop Action]` groups listed in `Actions`. Verify with tests on the real Brave, Firefox and Zen fixtures (action `Exec` and argv).
- [x] 1.2 Add the family table, private-window detection (Linux action, Windows flag), profile reading (`profiles.ini`, `Local State`) and `Expand`. Verify with fixture-tree tests for two Zen profiles (Flatpak path), one Chrome profile (no entries), a snap Firefox path, Windows paths and flags, and malformed stores.

## 2. Launch

- [x] 2.1 Linux: private action argv and profile-argument insertion before the URL or `@@u`. Windows: extra arguments after the executable, with quoting. Verify with unit tests, plus the Windows stub-browser test for a profile with spaces and a private flag.

## 3. Inspector

- [x] 3.1 Add the private option (checkbox, `P`, reset on an unsupported selection) and launch with `OpenPrivate`. Verify with UI tests using the router for `P`/Enter, and with a render.
- [x] 3.2 Use `Expand` in `cmd/bopen`'s environment and check the real browser list on this machine. Verify that the Zen and Firefox profile entries appear, then run all gates and confirm both CI jobs pass.
