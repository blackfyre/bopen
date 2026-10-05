# Proposal

## Why

Shortened links (`bit.ly`, `t.co`, `lnkd.in`, …) hide their destination. The inspector can show neither where they go nor what tracking the destination carries. Resolving them needs a request to the shortener. That is a privacy trade-off, so it must be the user's explicit choice each time.

## What Changes

- **Opt-in preference:** "Expand short links on request", off by default.
- **On demand:** when enabled and the link's host is a known URL shortener, the inspector offers **Expand**. Pressing it asks the shortener where the link goes:
  - a `HEAD` request (`GET` if `HEAD` is refused), without cookies, reading no body;
  - following redirects only while they stay on known shorteners;
  - stopping at the first other host without contacting it;
  - at most 5 hops and 5 seconds.
- **Result:** the destination replaces the inspected link and is analysed as usual, and the inspector notes the short link it came from. Failures are reported, and the original link is kept.

## Capabilities

### New Capabilities
- `short-links`: Known shorteners and on-demand expansion.

### Modified Capabilities
- `preferences`: The `expand_short_links` key.
- `settings-view`: The toggle.

## Impact

New package `internal/shortlinks`; additions to `internal/prefs` and `internal/ui`. This is bopen's second kind of network access (after ClearURLs), only on an explicit click.
