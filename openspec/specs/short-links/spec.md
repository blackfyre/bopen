# short-links Specification

## Purpose

Reveals where shortened links lead, only when the user asks, and contacting only the shortener.

## Requirements

### Requirement: Expansion is offered only on request for known shorteners
When `expand_short_links` is enabled and the host of the inspected link is a known URL shortener, the inspector SHALL offer an Expand action. bopen SHALL NOT contact a shortener before the user activates Expand, and SHALL NOT offer Expand when the preference is disabled.

#### Scenario: Disabled by default
- **WHEN** the inspected link is `https://bit.ly/abc` and the preference is not set
- **THEN** no Expand action is offered and no request is made

#### Scenario: Offered for a shortener
- **WHEN** the preference is enabled and the link is `https://t.co/xyz`
- **THEN** the inspector offers Expand

### Requirement: Only shorteners are contacted
Expanding SHALL send a `HEAD` request, or a `GET` request when `HEAD` is refused, without cookies, and SHALL NOT read a response body. It SHALL follow a redirect only while the next location is on a known shortener. It SHALL stop, without requesting it, at the first location on another host, which is the result. It SHALL fail after 5 redirects, after 5 seconds, on a response that is not a redirect, or on a location that is not an absolute or relative http or https URL.

#### Scenario: Chain of shorteners
- **WHEN** `https://bit.ly/a` redirects to `https://ow.ly/b`, which redirects to `https://example.org/article?utm_source=x`
- **THEN** exactly `bit.ly` and `ow.ly` are contacted, and the result is `https://example.org/article?utm_source=x`

#### Scenario: Not a redirect
- **WHEN** the shortener answers `200 OK`
- **THEN** expansion fails with a message, and the inspected link is unchanged

### Requirement: The expanded link replaces the inspected link
On success, the destination SHALL become the inspected link: it is analysed, highlighted and matched against site rules as if it had been clicked. The inspector SHALL show the short link it was expanded from.

#### Scenario: Tracking revealed
- **WHEN** expansion yields `https://example.org/article?utm_source=x`
- **THEN** the inspector suggests removing `utm_source=x`, and states that the link was expanded from `https://bit.ly/a`
