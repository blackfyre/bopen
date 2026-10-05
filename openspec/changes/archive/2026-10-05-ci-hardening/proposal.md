# Proposal

## Why

All Windows-specific code has only ever been cross-compiled: registry access, command-line building for real browsers, registration, the console attach and the appearance readers. The first real test would be a user's bug report. CI also warns that the pinned GitHub Actions target a deprecated Node runtime. bopen launches on every click, yet nothing guards its start-up cost.

## What Changes

- A `windows-latest` CI job runs the test suite on Windows, plus Windows-only integration tests:
  - real registry round trips under a dedicated test key;
  - launching a stub browser through the registered-command path and checking the exact arguments it receives;
  - registration and unregistration against the real HKCU (CI only).
- The GitHub Actions move to their current major versions (`actions/checkout@v7`, `actions/setup-go@v7`, `goreleaser/goreleaser-action@v7`).
- A start-up guard: a benchmark and a generous threshold test for the work bopen does before the window appears. Both run in CI.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. No product behaviour changes, so this change sets `skip_specs`.

## Impact

`.github/workflows/ci.yml` and `release.yml`, new `_windows_test.go` files, and a start-up test in `cmd/bopen`. Tests that write to the real registry run only when `BOPEN_WINDOWS_INTEGRATION=1`, so running the tests on a developer's Windows machine never touches their registration.
