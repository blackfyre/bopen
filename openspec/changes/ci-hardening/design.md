# Design

## Context

See proposal.md. The Windows code paths are thin wrappers (`winreg.System`, `launch.start` with `SysProcAttr.CmdLine`, `register.Windows`), already unit-tested against fakes.

## Goals / Non-Goals

**Goals:**
- Exercise the real Win32 and registry calls on every push.
- Catch a start-up regression of the pre-window work.

**Non-Goals:**
- UI automation on Windows.
- Measuring time to first frame (CI has no display).

## Decisions

- **Stub browser.** `TestMain` in `internal/launch` builds a tiny Go program (`testdata/argv`) that writes its `os.Args` as JSON to a file. Building it at test time keeps binaries out of the repository. Real launches use Chrome-style (`--single-argument %1`), Firefox-style (`-osint -url "%1"`) and placeholder-less templates, with URLs containing `&`, `^`, `%`, `"` (after validation) and a trailing backslash. The test asserts exactly one URL argument, byte for byte.
- **Integration guard.** Tests that write to the real `HKCU` (registry round trip under `Software\bopen-test`, plus register and unregister) skip unless `BOPEN_WINDOWS_INTEGRATION=1`. The Windows CI job sets it.
- **Start-up guard.** `cmd/bopen` gains `BenchmarkPrepare` and `TestPrepareIsFast`. They run `prepare()` against a temporary home with stub browsers, the built-in rules and a ClearURLs fixture enabled, and assert it stays under 250 ms. That is generous next to the design target of 150 ms to first frame, so CI noise doesn't cause failures, but an order-of-magnitude regression is caught.
- **Windows CI job.** It runs `go vet`, `go test ./...` and the GUI build. The Linux job keeps the remaining gates.

## Risks / Trade-offs

- **[Flaky timing on shared runners]** → The threshold is about ten times the typical measurement. The benchmark output is logged for trend-watching.
