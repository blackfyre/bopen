# clearurls-sync Specification

## Purpose

Obtains and maintains a verified local copy of the ClearURLs rule data, only when the user has opted in, without ever delaying the handling of a link.

## Requirements

### Requirement: No network access unless enabled
bopen SHALL NOT make any network request when the ClearURLs preference is disabled.

#### Scenario: Default installation
- **WHEN** bopen handles links with default preferences
- **THEN** no network request is made

### Requirement: Download and verification
When fetching, bopen SHALL download `data.minify.json` and `rules.minify.hash` over HTTPS from `rules2.clearurls.xyz`, falling back to `rules1.clearurls.xyz` if that fails. It SHALL accept the data only when its SHA-256 digest matches the hash file, the response is no larger than 5 MiB, the request completes within 15 seconds, and the data parses as a ClearURLs rule file. Accepted data SHALL replace the cache atomically. Rejected data SHALL leave the existing cache untouched.

#### Scenario: Hash mismatch
- **WHEN** the downloaded data does not match the published hash
- **THEN** the cache is unchanged, and the failure is recorded as the last error

#### Scenario: Mirror fallback
- **WHEN** `rules2.clearurls.xyz` is unreachable and `rules1.clearurls.xyz` serves valid data
- **THEN** the cache is updated from `rules1.clearurls.xyz`

### Requirement: Cache location and use
The verified data SHALL be cached as `clearurls.json` in a `bopen` directory inside the OS user cache directory, together with the time of the last successful update and the last error. Analysis SHALL use only the cache and SHALL never wait for a download.

#### Scenario: Link handled while offline
- **WHEN** ClearURLs is enabled, a cache exists, and the network is unavailable
- **THEN** links are analysed with the cached ClearURLs rules without delay

### Requirement: Background refresh
While the inspector window is open, bopen SHALL start a refresh in the background if ClearURLs is enabled and the last successful update is older than 24 hours or missing. A refresh SHALL NOT delay showing the window or opening a link. A refresh still in progress when bopen exits SHALL be abandoned without affecting the cache. Links handled without a window SHALL NOT trigger a refresh.

#### Scenario: Stale cache refreshed
- **WHEN** ClearURLs is enabled, the cache is two days old, and the inspector is shown
- **THEN** a background refresh starts, and the window is fully usable while it runs

#### Scenario: Silent path
- **WHEN** a link is opened without showing the window
- **THEN** no refresh is started

### Requirement: Manual update
An explicit update request from settings SHALL fetch regardless of the cache age, and SHALL report success or the error when it completes.

#### Scenario: Update now
- **WHEN** the user activates "Update now"
- **THEN** a fetch runs, and the settings view then shows the new last-update time or the error
