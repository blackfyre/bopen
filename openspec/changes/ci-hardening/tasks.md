# Tasks

## 1. Windows integration tests

- [x] 1.1 Add a stub browser built in `TestMain`, plus Windows launch tests through `start()` covering the Chrome, Firefox and no-placeholder templates and awkward URLs. Verify that `GOOS=windows go vet ./...` succeeds locally, and that the tests pass in the Windows CI job.
- [x] 1.2 Add guarded real-registry tests: a `winreg.System` round trip under `Software\bopen-test`, register/unregister with read-back, and `appearance.Read` not failing. Verify they skip without `BOPEN_WINDOWS_INTEGRATION=1` and pass in CI.

## 2. Start-up guard

- [x] 2.1 Add `BenchmarkPrepare` and `TestPrepareIsFast` (under 250 ms) in `cmd/bopen`, using a temporary home with stub browsers and ClearURLs enabled. Verify locally, and record the measured time.

## 3. Workflows

- [ ] 3.1 Add the `windows-latest` job, set `BOPEN_WINDOWS_INTEGRATION=1`, and bump all actions to v7 in `ci.yml` and `release.yml`. Verify with actionlint, then push and confirm both CI jobs pass.
