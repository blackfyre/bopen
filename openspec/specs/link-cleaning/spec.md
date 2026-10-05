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

### Requirement: Built-in rules have stable identifiers
Every built-in rule SHALL have a unique identifier that does not change between releases while the rule exists.

#### Scenario: Identifier uniqueness
- **WHEN** the built-in rule set is loaded
- **THEN** no two rules share an identifier

### Requirement: Disabled rules are not applied
A built-in rule whose identifier is listed as disabled SHALL NOT produce suggestions.

#### Scenario: Disabled tracking rule
- **WHEN** the `fbclid` rule is disabled and `https://example.com/?fbclid=x` is analysed
- **THEN** no suggestion is produced

#### Scenario: Disabled redirect rule
- **WHEN** the Google redirect rule is disabled and a Google wrapper URL is analysed
- **THEN** no redirect suggestion is produced

### Requirement: User rules produce suggestions
User rules of kind `tracking` or `affiliate` SHALL be applied with the same matching semantics as built-in parameter rules: case-insensitive names, trailing `*` as a prefix match, and an optional host scope. Their suggestions SHALL carry source `user` and the user's reason. User rules of kind `redirect` SHALL NOT be supported.

#### Scenario: User rule on a specific host
- **WHEN** a user rule flags `ref` on host `news.example.com` with the reason "Referrer tracking", and `https://news.example.com/a?ref=home` is analysed
- **THEN** a `tracking` suggestion with source `user` and reason "Referrer tracking" covers `ref=home`

#### Scenario: Host scope respected
- **WHEN** the same rule exists and `https://other.example.com/a?ref=home` is analysed
- **THEN** no suggestion covers `ref=home`

### Requirement: Invalid user rules are ignored and reported
A user rule with an empty reason, an empty parameter, or an unsupported kind SHALL NOT be applied. It SHALL be reported as a configuration problem in the same way as other invalid preferences.

#### Scenario: Rule without reason
- **WHEN** `config.toml` contains a user rule with an empty reason
- **THEN** the rule is not applied, and the inspector shows a configuration problem naming that rule

### Requirement: One suggestion per span with source precedence
When rules from several sources match the same span, exactly one suggestion SHALL be produced, taken from the highest-precedence source: `user` before `builtin`. A disabled built-in rule SHALL NOT take part in precedence.

#### Scenario: User rule overrides built-in reason
- **WHEN** a user rule flags `fbclid` with the reason "Facebook tracking, always remove" and a link contains `fbclid=x`
- **THEN** a single suggestion covers `fbclid=x`, with source `user` and the user's reason

### Requirement: ClearURLs rules as a rule source
When ClearURLs is enabled and a cache exists, its providers SHALL be applied to URLs that match the provider's `urlPattern` and none of its `exceptions`. Matching follows the ClearURLs semantics, applied case-insensitively:

| ClearURLs field | Applied as | Suggestion kind |
|---|---|---|
| `rules` | regular expressions matched against the whole query parameter name | `tracking` |
| `referralMarketing` | the same, matched against parameter names | `affiliate` |
| `rawRules` | regular expressions matched against the raw URL text | `tracking` |
| `redirections` | regular expressions whose first capture group, percent-decoded, is the target | `redirect` (same validation and recursion as built-in redirects) |

`completeProvider` and `forceRedirection` SHALL be ignored. A pattern that fails to compile SHALL be skipped, and the skip count SHALL be reported in settings.

A `rawRules` match before the query (in the path) or after it (in the fragment) SHALL become its own removable span. A match inside the query SHALL count only when it covers exactly one parameter that no other rule flags, plus at most its separators. Any other match SHALL be ignored, so suggestions never overlap.

#### Scenario: Provider-specific parameter
- **WHEN** ClearURLs is enabled and `https://www.amazon.de/dp/B000?crid=XYZ` is analysed
- **THEN** a `tracking` suggestion with source `clearurls` covers `crid=XYZ`

#### Scenario: Provider exception honoured
- **WHEN** a URL matches a provider's `urlPattern` and one of its `exceptions`
- **THEN** that provider produces no suggestions for the URL

#### Scenario: Raw rule in the path
- **WHEN** ClearURLs is enabled and `https://www.amazon.de/dp/B000/ref=sr_1_3?keywords=x` is analysed
- **THEN** a `tracking` suggestion covers `/ref=sr_1_3`, and accepting it yields `https://www.amazon.de/dp/B000?keywords=x`

### Requirement: ClearURLs suggestions have a generic reason
Suggestions from ClearURLs SHALL carry source `clearurls` and a reason of the form "Listed by ClearURLs as tracking for <provider>". Global rules SHALL use the provider name "all sites".

#### Scenario: Reason text
- **WHEN** the Amazon provider flags `crid`
- **THEN** the suggestion's reason names ClearURLs and the provider `amazon`

### Requirement: ClearURLs has the lowest precedence
When a ClearURLs match covers the same span as a user or built-in match, the user or built-in suggestion SHALL be used. Disabling a built-in rule SHALL NOT suppress a ClearURLs match for the same parameter.

#### Scenario: Built-in reason wins
- **WHEN** both the built-in `fbclid` rule and ClearURLs global rules match `fbclid=x`
- **THEN** a single suggestion with source `builtin` and the built-in reason is produced

### Requirement: Disabled or missing ClearURLs data has no effect
When ClearURLs is disabled, or enabled with no valid cache, analysis SHALL produce the same suggestions as without ClearURLs.

#### Scenario: Enabled before first download
- **WHEN** ClearURLs has just been enabled and no download has completed yet
- **THEN** only user and built-in rules produce suggestions
