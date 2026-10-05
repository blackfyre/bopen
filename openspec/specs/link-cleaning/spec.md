# link-cleaning Specification

## Purpose

Analyses a URL for redirect wrappers and tracking parameters, explains each suggested removal, and produces the cleaned URL from the suggestions the user accepts.

## Requirements

### Requirement: Analysis produces explained suggestions
Analysing a URL SHALL produce an ordered list of suggestions. Each suggestion SHALL have a kind (`redirect`, `tracking` or `affiliate`), the exact text span of the original URL it affects, a human-readable reason, a source (`builtin` in this change), and a default acceptance state. Analysis SHALL be deterministic and SHALL NOT perform network access.

#### Scenario: Tracking parameter is explained
- **WHEN** the URL `https://example.com/a?id=7&utm_source=newsletter` is analysed
- **THEN** exactly one suggestion is produced, with kind `tracking`, span `utm_source=newsletter`, a non-empty reason and source `builtin`

#### Scenario: Clean URL yields no suggestions
- **WHEN** the URL `https://example.com/a?id=7` is analysed
- **THEN** no suggestions are produced

### Requirement: Built-in rules cover common trackers and wrappers
The built-in rule set SHALL include at least the following.

Tracking parameters on any host:
- `utm_*`
- `fbclid`
- `gclid`
- `dclid`
- `gbraid`
- `wbraid`
- `msclkid`
- `yclid`
- `mc_eid`
- `igshid`
- `_hsenc`
- `_hsmi`
- `mkt_tok`

Redirect wrappers:
- `www.google.<tld>/url` (target in `q` or `url`)
- `l.facebook.com/l.php` and `lm.facebook.com/l.php` (target in `u`)
- `*.safelinks.protection.outlook.com` (target in `url`)
- `www.youtube.com/redirect` (target in `q`)

Affiliate parameters:
- `tag` on Amazon hosts

#### Scenario: Microsoft click identifier
- **WHEN** a URL with the query parameter `msclkid=abc` is analysed
- **THEN** a `tracking` suggestion covers `msclkid=abc`

#### Scenario: Google redirect wrapper
- **WHEN** `https://www.google.com/url?q=https%3A%2F%2Fshop.example.com%2Fitem&sa=D` is analysed
- **THEN** a `redirect` suggestion is produced whose result is `https://shop.example.com/item`

### Requirement: Parameter names match case-insensitively
Parameter rules SHALL match query parameter names case-insensitively. A rule ending in `*` SHALL match any parameter name with that prefix.

#### Scenario: Upper-case UTM parameter
- **WHEN** a URL containing `UTM_Campaign=spring` is analysed
- **THEN** a `tracking` suggestion covers `UTM_Campaign=spring`

### Requirement: Host-scoped rules apply only to their hosts
A rule scoped to a host pattern SHALL produce suggestions only for URLs whose host matches that pattern.

#### Scenario: Amazon affiliate tag
- **WHEN** `https://www.amazon.de/dp/B000?tag=someone-21` is analysed
- **THEN** an `affiliate` suggestion covers `tag=someone-21`

#### Scenario: Same parameter elsewhere is untouched
- **WHEN** `https://blog.example.com/post?tag=golang` is analysed
- **THEN** no suggestion covers `tag=golang`

### Requirement: Redirect wrappers are unwrapped recursively
When a redirect wrapper is found, its target SHALL be analysed as well, and nested wrappers SHALL be unwrapped up to a depth of 5. A wrapper whose target is not an absolute `http` or `https` URL SHALL NOT produce a redirect suggestion. Suggestions found in the target SHALL depend on the redirect suggestion that revealed them.

#### Scenario: Tracking inside wrapped target
- **WHEN** a Google redirect wrapper's target is `https://shop.example.com/item?fbclid=x`
- **THEN** a `redirect` suggestion and a dependent `tracking` suggestion for `fbclid=x` are produced

#### Scenario: Wrapper with non-web target
- **WHEN** a Google redirect wrapper's `q` parameter is `javascript:alert(1)`
- **THEN** no `redirect` suggestion is produced

### Requirement: Default acceptance depends on kind
`redirect` and `tracking` suggestions SHALL be accepted by default. `affiliate` suggestions SHALL NOT be accepted by default.

#### Scenario: Affiliate tag kept by default
- **WHEN** an Amazon URL with `tag=someone-21` is analysed and no suggestion is toggled
- **THEN** the cleaned URL still contains `tag=someone-21`

### Requirement: Cleaned URL reflects accepted suggestions only
The cleaned URL SHALL apply exactly the accepted suggestions. Rejecting a redirect suggestion SHALL also leave its dependent suggestions unapplied. Removing parameters SHALL preserve the order and the original encoding of the remaining parameters and SHALL preserve the fragment. When all query parameters are removed, the `?` SHALL be removed as well.

#### Scenario: Remaining parameters untouched
- **WHEN** `https://example.com/a?b=1&utm_source=x&c=%20d#top` is cleaned with all suggestions accepted
- **THEN** the result is `https://example.com/a?b=1&c=%20d#top`

#### Scenario: Rejected suggestion is kept
- **WHEN** the `utm_source` suggestion for `https://example.com/a?utm_source=x` is rejected
- **THEN** the cleaned URL equals the original URL

#### Scenario: Rejected redirect keeps wrapper
- **WHEN** the redirect suggestion for a Google wrapper is rejected
- **THEN** the cleaned URL is the wrapper URL, and the dependent suggestions in its target are not applied

#### Scenario: Empty query removed
- **WHEN** `https://example.com/a?fbclid=x` is cleaned with all suggestions accepted
- **THEN** the result is `https://example.com/a`
