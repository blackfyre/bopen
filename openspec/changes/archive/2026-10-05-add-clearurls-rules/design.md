# Design

## Context

This builds on the rule-source list introduced in `add-user-rules` (precedence `user`, then `builtin`) and on the settings view. It adds bopen's first network access. The ClearURLs data was checked on 2026-10-05:
- `data.minify.json` is about 36 KB and covers 206 providers;
- `rules.minify.hash` holds the hex SHA-256 digest of that file;
- all 1,095 patterns compile with Go's RE2 `regexp`.

See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Zero impact on link-handling latency and zero network access when disabled.
- Faithful mapping of ClearURLs semantics onto bopen's span-based suggestions.

**Non-Goals:**
- Per-provider enable or disable.
- Using ClearURLs `completeProvider` blocking.
- Bundling ClearURLs data in the binary.
- Signature verification beyond the published hash.

## Decisions

### Package `internal/clearurls`

The package has three parts:
- `fetch`: HTTP with timeout and size limit, mirror fallback, hash check, atomic cache write;
- `load`: parses the cache and compiles patterns, counting failures;
- `source`: implements the rule-source interface from `add-user-rules`.

Patterns are compiled when the cached list is loaded, which only happens when ClearURLs is enabled. Measured on the published list (2026-10-05: 810 rules, none skipped), compiling takes about 6 ms and analysing a heavily tracked URL with all rules about 0.1 ms, so lazy compilation is not needed.

### Mapping parameter rules onto spans

The ClearURLs extension builds `(?:&|[/?#&])(?:<rule>=[^&]*)` and matches it against the whole URL. bopen already splits the raw query into parameter segments with offsets, so it matches `^(?:<rule>)$` (case-insensitive) against each segment's raw name instead. That's equivalent for query parameters, and yields exact spans for highlighting. Patterns with a `(?:%3F)?` prefix still match, because the raw name is tested before decoding.

Raw rules are matched against the URL text. The published list has four: Amazon's `/ref=…` path segment (twice), a whole-query `?pc` and a `#lead…` fragment. A match in the path or fragment becomes its own removable span. A match inside the query counts only when it covers exactly one otherwise unflagged parameter (plus at most its separators); any other raw match is dropped, so highlights never nest or overlap.

Redirections use the first capture group, percent-decoded once, then pass through the existing redirect validation and recursion.

### Precedence via span collision

The analyser already keeps the highest-precedence source per span. A ClearURLs `rawRules` span can overlap a parameter span without being identical. When spans overlap partially, the higher-precedence span is kept and the overlapping lower-precedence span is dropped, so highlights never nest or overlap.

### Background refresh lifecycle

When the inspector shows, a goroutine checks the cache age and fetches if needed. The fetch writes to a temporary file and renames it only after verification, so abandoning it at process exit (the user pressed Enter) leaves the old cache intact. The new data takes effect from the next invocation. The current link is not re-analysed mid-inspection, so suggestions don't shift under the user. Concurrent refreshes from two open instances are harmless, because each rename is atomic and both write identical verified content.

The cache metadata (`last_success`, `last_error`, `skipped_patterns`) is stored in `clearurls.meta.toml` next to the data, so `state.toml` stays limited to browser state.

### HTTP client

bopen uses `net/http` with a 15-second total timeout, the proxy settings from the environment, and the user agent `bopen/<version>`. The 5 MiB limit is enforced with `io.LimitReader`, plus a check for an extra byte. Only the two fixed HTTPS URLs are ever requested.

## Risks / Trade-offs

- **[Hash is not a signature]** The hash comes from the same host as the data, so it detects corruption but not a compromised host. → Accepted. The data can only cause parameters to be *suggested* for removal, never code execution, and every suggestion is shown, labelled with its source, and toggleable.
- **[Upstream format changes]** → A parse failure rejects the download and keeps the previous cache. The error is shown in settings.
- **[Future pattern incompatible with RE2]** None today, but JavaScript regex features such as lookbehind could appear. → Skip and count; never fail the whole load.
- **[Generic reasons are less informative]** → The source label lets the user tell ClearURLs matches from curated ones. Built-in rules win on collisions, so curated reasons are preferred.
