# Tasks

## 1. Preferences and rules

- [ ] 1.1 Add `[browsers]` (`hidden`, `order`) and `[rules]` (`disabled`) to the config model, with unknown identities ignored at use time but preserved in the file. Verify with unit tests for ordered and unordered browsers, a preserved unknown identity, and an unknown rule ID.
- [ ] 1.2 Implement the read-modify-write save (re-read, apply one mutation, atomic write). Verify with a test that simulates two instances saving different settings and asserts both survive.
- [ ] 1.3 Add a unique `id` to every rule in `builtin.toml`, and filter disabled IDs during analysis. Verify by extending the loader test (unique, non-empty IDs) and with tests for disabled tracking and redirect rules.

## 2. Inspector integration

- [ ] 2.1 Apply hidden/order filtering to the inspector browser list, skip hidden candidates in pre-selection, and handle the all-hidden case. Verify with unit tests on the list and pre-selection resolver for the last-used-hidden and all-hidden scenarios.
- [ ] 2.2 Add the cog control and view switching with preserved model, re-analysis on return, and toggles carried over by span. Verify manually: reject a suggestion, open settings, go back, and confirm the toggle is unchanged. A unit test covers the span-based toggle carry-over.

## 3. Settings view

- [ ] 3.1 Implement the settings view sections: window mode, browsers (hide/show, up/down) and built-in rules (enable/disable with kind, pattern and reason), each saving immediately. Verify manually that every control updates `config.toml` immediately and that the inspector reflects it after going back.
- [ ] 3.2 Implement the registration status (Linux `xdg-mime query`, Windows `UserChoice` read, "unknown" fallback) and the re-register action that reuses `internal/register`. Verify with unit tests using a stub `xdg-mime` and the fake registry, and check manually on Linux that re-register restores bopen as the default.
- [ ] 3.3 Add the `bopen settings` subcommand and update usage. Verify with an argument-dispatch test, and check manually that `bopen settings` opens the view without a back control and launches no browser when closed.

## 4. Verification

- [ ] 4.1 Document UI-owned `config.toml` (comments not preserved) and the new keys in the README, then run `gofmt -l .` (empty), `go vet ./...`, `go test ./...` and `GOOS=windows go build ./...`, all passing.
