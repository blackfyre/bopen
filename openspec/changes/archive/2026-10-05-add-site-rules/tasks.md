# Tasks

## 1. Model and matching

- [x] 1.1 Add `SiteRule` with validation, mutations, `[[sites]]` storage, and problems for invalid rules. Verify with prefs tests for persistence, unique IDs and the invalid-rule report.
- [x] 1.2 Add `app.MatchSite`, preferred-identity pre-selection and the `Direct` window decision. Verify with app tests for subdomain matching, matching through a redirect wrapper, a hidden browser, first-match order, and direct versus problems.

## 2. Inspector

- [x] 2.1 Wire matching into `Model.Refresh` and `cmd/bopen` (`prepare`, silent path), and add the rule line or remember checkbox to the Open-in card, saving the rule on a successful open. Verify with UI tests for pre-selection, the remember flow (rule saved, next link pre-selects) and the direct path through `prepare`/`NeedWindow`.

## 3. Settings

- [x] 3.1 Add the "Site rules" card and the site form (patterns, browser, direct), with add, edit and delete. Verify with UI tests for add/edit/delete persistence and form validation, and with offscreen renders of the card and the form.

## 4. Verification

- [x] 4.1 Document site rules in the README. Run `gofmt -l .` (empty), `go vet ./...` (Linux and Windows), `go test ./...` and the Windows GUI build, all passing, then push and confirm both CI jobs pass.
