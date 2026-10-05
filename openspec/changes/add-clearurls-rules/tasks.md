# Tasks

## 1. Preferences and cache

- [ ] 1.1 Add `[rules] clearurls` (default `false`) to the config model, and resolve the cache directory via `os.UserCacheDir()/bopen`. Verify with unit tests for the default value and the Linux cache path.
- [ ] 1.2 Implement `clearurls.meta.toml` (last success, last error, skipped patterns) with atomic writes. Verify with round-trip and corrupt-file unit tests.

## 2. Fetch

- [ ] 2.1 Implement the fetch (HTTPS, mirror fallback, 15-second timeout, 5 MiB limit, SHA-256 check against `rules.minify.hash`, parse check, atomic replace). Verify with `httptest` server tests for success, hash mismatch, oversize, timeout, unparsable data and mirror fallback, each asserting that the cache is replaced or unchanged as specified.
- [ ] 2.2 Verify with a test using a failing HTTP transport that no request is made while `clearurls` is disabled, both for the silent path and for the inspector.

## 3. Rule source

- [ ] 3.1 Implement loading and lazy compilation of providers (`urlPattern`, `exceptions`, `rules`, `referralMarketing`, `rawRules`, `redirections`; ignore `completeProvider` and `forceRedirection`; count failed patterns). Verify with tests against a fixture copy of `data.minify.json` in `testdata/`, asserting the provider count and zero skipped patterns, plus a synthetic failing pattern that is counted.
- [ ] 3.2 Implement span mapping (parameter names via anchored case-insensitive match, raw rules on the full URL, redirections via capture group) and register `clearurls` as the lowest-precedence source with partial-overlap resolution. Verify with tests for Amazon `crid`, provider exceptions, the Google redirection, built-in-wins on `fbclid`, the generic reason text, and enabled-without-cache behaving like disabled.

## 4. Refresh and settings

- [ ] 4.1 Start the background refresh when the inspector shows (enabled and cache older than 24 hours or missing) and never on the silent path. Verify with tests using an injected clock and a fake fetcher, and check manually that the window is usable while a slow fetch runs.
- [ ] 4.2 Add the ClearURLs section to settings (toggle with a download notice, immediate fetch on enable, last update, last error, skipped count, "Update now", attribution with licence and link). Verify manually by enabling it, confirming that the cache file appears and the status updates, and by simulating a failure with an unreachable proxy so the error is shown.

## 5. Verification

- [ ] 5.1 Document the ClearURLs option, its network behaviour and attribution in the README, then run `gofmt -l .` (empty), `go vet ./...`, `go test ./...` and `GOOS=windows go build ./...`, all passing.
