# Tasks

## 1. Expander

- [x] 1.1 Implement `internal/shortlinks` (host list, `IsShortener`, `Expander`). Verify with `httptest` tests for a chain that stops before the destination (request log), the `HEAD`→`GET` fallback, a `200` failure, too many hops, a relative `Location`, a non-http `Location` and the timeout.

## 2. Inspector and settings

- [x] 2.1 Add the `expand_short_links` preference and settings card. Verify with a prefs test for the default and a settings save test.
- [x] 2.2 Add the Expand button, background expansion, `Model.Replace` with `ExpandedFrom`, and error notices. Verify with UI tests using a fake expander (success re-analyses and pre-selects by site rule; failure keeps the link; no button when disabled or not a shortener) and a render.

## 3. Verification

- [x] 3.1 Document the option in the README. Run all gates, push, and confirm both CI jobs pass.
