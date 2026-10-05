# Proposal

## Why

Tracking parameters spread when links are shared, but bopen only cleans links it is asked to open. Users also want the cleaned link itself, to paste into a chat or a document, and that needs a copy action and a way to use the cleaner outside the inspector.

## What Changes

- `bopen clean [URL…]` prints each link cleaned with the default suggestions, one per line. Without arguments it cleans the links read from standard input, one per line, so it works in pipes and editor integrations.
- It uses the same rules as the inspector (your rules, built-in rules, and the cached ClearURLs list when enabled), and never touches the network.
- `--explain` lists what was removed, and why, on standard error.
- The inspector gets a **Copy** button on the result (shortcut `C`), which copies the cleaned link to the clipboard and keeps the window open.
- On Windows, command-line output keeps an already redirected standard output instead of always attaching to the console, so `bopen clean … | …` works.

## Capabilities

### New Capabilities
- `clean-command`: The `bopen clean` command.

### Modified Capabilities
- `default-handler-registration`: Command-line usage gains `clean`.
- `link-inspector`: The copy action.

## Impact

Touches `cmd/bopen` (subcommand and Windows console handling) and `internal/ui` (copy button, shortcut and feedback). Both reuse the existing analysis.
