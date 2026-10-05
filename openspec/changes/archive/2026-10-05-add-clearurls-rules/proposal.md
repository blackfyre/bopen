# Proposal

## Why

The curated built-in rules are deliberately small, so they explain each suggestion well but cover only common trackers. The ClearURLs community rule set covers about 200 providers, including site-specific parameters and redirect wrappers. It is LGPL-3.0 licensed, so fetching it on demand (rather than bundling it) adds coverage without constraining bopen's MIT licence.

## What Changes

- An opt-in preference that enables the ClearURLs rule set. It is off by default, and no network access happens unless it is enabled.
- When enabled, bopen downloads the ClearURLs data file over HTTPS, verifies it against the published SHA-256 hash, and caches it in the OS user cache directory.
- The cache is refreshed in the background at most once a day while the inspector is open. Refreshing never delays a link, and an "Update now" action is available in settings.
- ClearURLs rules are mapped onto bopen suggestions:
  - parameter rules become `tracking` suggestions;
  - referral-marketing rules become `affiliate` suggestions;
  - redirections become `redirect` suggestions;
  - provider exceptions are honoured.
- These suggestions have source `clearurls`, a generic reason naming the provider, and the lowest precedence.
- The settings view shows the enable toggle, the last update time and any error, and gives attribution with a link to the ClearURLs project.

## Capabilities

### New Capabilities
- `clearurls-sync`: Fetching, verifying, caching and refreshing the ClearURLs rule data.

### Modified Capabilities
- `link-cleaning`: ClearURLs as a rule source, its mapping, and its precedence.
- `preferences`: The ClearURLs enable key and the cache location.
- `settings-view`: ClearURLs controls, status and attribution.

## Impact

- Depends on `add-link-inspector-core`, `add-settings-view` and `add-user-rules`. The rule-source list from `add-user-rules` is extended.
- First network access in bopen, limited to `rules2.clearurls.xyz` and `rules1.clearurls.xyz`, and only when enabled.
- New package `internal/clearurls`. No new third-party dependencies (`net/http`, `crypto/sha256`, `regexp`).
