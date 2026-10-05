# Proposal

## Why

bopen re-serialises every link through Go's URL parser before handing it to a browser. That has three effects:
- **Rewriting:** it re-encodes non-ASCII paths (`/café` becomes `/caf%C3%A9`) and drops an empty fragment (`#`). The browser receives something other than what was clicked, even when nothing was cleaned.
- **Wrong refusal:** it rejects links with a malformed percent-escape (`%zz`), which every browser opens. The user gets the wrong message, "only http and https links are accepted".
- **No analysis:** the analyser gives up on such links entirely.

Separately, binaries installed with `go install` report their version as `dev`.

## What Changes

- **Fidelity:** a link is handed off exactly as received. The only changes are trimming surrounding whitespace, lower-casing the scheme, percent-encoding spaces, tabs and double quotes, and removing suggestions the user accepted. Validation (http/https, non-empty host, never option-like) is unchanged, but runs on a tolerant copy, so stray `%` signs no longer cause a rejection.
- **Tolerant analysis:** links containing malformed escapes are analysed like any other.
- **Version for source installs:** when no release version is linked in, bopen reports the module version Go records at build time (for example `0.3.0` for `go install …@v0.3.0`).

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `browser-handoff`: "Only web URLs are handed off" changes from canonical serialisation to fidelity.
- `link-cleaning`: Analysis tolerates malformed percent-escapes.
- `release-distribution`: Source installs report their module version.

## Impact

Touches `internal/launch` (`Validate`), `internal/clean` (URL parsing in analysis), and `cmd/bopen` (version resolution). The tests are extended. No configuration or UI changes.
