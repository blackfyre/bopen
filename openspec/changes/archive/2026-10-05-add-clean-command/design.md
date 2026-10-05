# Design

## Context

See proposal.md. `cmd/bopen` dispatches on `parseArgs`. `loadEnv` builds the rules (user, built-in and cached ClearURLs), and `launch.Validate` with `clean.Analyse` does the work. On Windows, `attachConsole` always replaces `os.Stdout` and `os.Stderr` with `CONOUT$`.

## Decisions

- **Dispatch.** `parseArgs` returns `modeClean` when the first argument is `clean`. `runClean(args, stdin, stdout, stderr) int` is a pure function over readers and writers, so it is unit-tested. It loads the environment (`loadEnv`), but never starts the background ClearURLs refresh, which lives only in the UI.
- **Output.** One line per input, in input order. Invalid lines are echoed unchanged so pipelines stay aligned, and they set the exit status to 1.
- **Windows console.** `attachConsole` keeps a standard handle that is already valid (`GetStdHandle` returns a handle whose `GetFileType` is not `FILE_TYPE_UNKNOWN`). This is the case for pipes and files. It attaches the console only for handles that are missing.
- **Copy.** `gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: …})` is issued from the inspector's handler for the Copy button and the `C` key. The window shows "Copied" until the result changes.

## Risks / Trade-offs

- **[Defaults hide choices]** `clean` cannot ask about affiliate tags. → It applies exactly the inspector defaults, and documents that.
