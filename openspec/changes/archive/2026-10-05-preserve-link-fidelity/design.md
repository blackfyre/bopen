# Design

## Context

See proposal.md. `launch.Validate` returns `url.Parse(raw).String()` with spaces and quotes encoded, and analysis runs on that string. `clean.Analyse` calls `url.Parse` on each node and returns early on error. The version comes only from `-X main.version`.

## Decisions

### Validate a tolerant copy, return the original

`launch.Validate` trims the input, and builds a validation copy where every `%` that doesn't start two hex digits becomes `%25`. It parses that copy, applies the existing checks (scheme, non-empty host, no opaque part), and then returns the original trimmed text with:
- the scheme lower-cased (the first `len(scheme)` bytes);
- spaces, tabs and `"` percent-encoded.

The `http://`/`https://` prefix check stays as a final guard. Control characters are still rejected, because the parser refuses them in the copy too.

Rejected alternative: keep re-serialising, but special-case non-ASCII text and empty fragments. That stays lossy, and still refuses links with malformed escapes.

### Tolerant parsing in analysis

The same escaping helper moves to `clean` as `ParseTolerant`, which `launch` reuses. `Analyse` uses it to find each node's host and path. Query segments are already handled raw.

### Version resolution

`cmd/bopen` resolves the version once. It uses the linked-in `version` when it isn't `dev`. Otherwise it uses `debug.ReadBuildInfo().Main.Version` when that is neither empty nor `(devel)`, with any leading `v` removed. Otherwise it reports `dev`. The resolver takes its build-info source as an argument, so it is unit-tested.

## Risks / Trade-offs

- **[Browsers see raw non-ASCII]** → That is what they received before bopen was installed, and every target browser accepts IRIs. On Windows the command line is UTF-16, so non-ASCII text survives.
