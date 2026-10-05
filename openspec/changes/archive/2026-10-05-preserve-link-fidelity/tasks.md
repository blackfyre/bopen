# Tasks

## 1. Link fidelity

- [x] 1.1 Add `clean.ParseTolerant` and use it in analysis. Verify with tests for a stray `%` in the path and a malformed escape in the query that still produces suggestions.
- [x] 1.2 Rewrite `launch.Validate` to validate a tolerant copy and return the original with only the allowed changes. Verify with tests for byte-for-byte pass-through (non-ASCII, empty fragment, `+` and escapes kept), a malformed escape, scheme lower-casing, space and quote encoding, and the existing rejections. Confirm the Linux and Windows launch tests still pass.

## 2. Version

- [x] 2.1 Resolve the version from build info when none is linked in. Verify with unit tests for linked, module-versioned, `(devel)` and missing build info, and check what a local `go install` prints (Go records a VCS-derived pseudo-version, so it is not `dev`).

## 3. Verification

- [x] 3.1 Run `gofmt -l .` (empty), `go vet ./...` (Linux and Windows), `go test ./...` and the Windows GUI build, all passing. Then push and confirm both CI jobs pass.
