# Spec Delta

## ADDED Requirements

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
