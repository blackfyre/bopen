# Tasks

## 1. Rule model and matching

- [x] 1.1 Add `[[rules.user]]` to the config model with ID generation and validation (non-empty reason and param, kind `tracking`/`affiliate`), reporting invalid rules as config problems. Verify with unit tests for valid, reason-less, param-less and redirect-kind rules.
- [x] 1.2 Refactor analysis to take an ordered list of rule sources with per-span precedence, and add the `user` source ahead of `builtin`. Verify with tests for the host-scoped, scope-respected and user-overrides-built-in scenarios, and confirm the existing link-cleaning tests still pass.
- [x] 1.3 Add mutations to add, edit and delete user rules and to disable a built-in rule, reusing the read-modify-write save. Verify with unit tests that each mutation persists correctly and preserves unrelated settings.

## 2. Inspector actions

- [x] 2.1 Use `gioui.org/x/component` (already a dependency through `gioui.org/x`, pinned to the Gio release in use) and make every query parameter span in the rendered URL a right-click target. Verify that `go build ./...` succeeds, and check manually that right-clicking a parameter opens the menu.
- [x] 2.2 Implement the "Always flag this parameter…" overlay form (pre-filled param, scope defaulting to this host, kind, required reason) with immediate re-analysis and toggle carry-over. Verify manually with the flagging and mandatory-reason scenarios, and add a unit test for host selection when the parameter is inside a redirect target.
- [x] 2.3 Implement the suggestion context actions "Disable this rule" (built-in) and "Edit this rule…" (user). Verify manually that disabling `utm_*` removes its suggestions and that editing a user rule's reason updates the list.

## 3. Settings

- [x] 3.1 Add a "Your rules" section listing user rules, with add, edit (including wildcard hosts) and delete using the shared form. Verify manually with the delete and edit-reason scenarios, and check `config.toml` after each.

## 4. Verification

- [x] 4.1 Document user rules in the README, then run `gofmt -l .` (empty), `go vet ./...`, `go test ./...` and `GOOS=windows go build ./...`, all passing.
