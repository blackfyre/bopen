# site-rules Specification

## Purpose

Lets the user send links for chosen sites to a chosen browser automatically, optionally without showing the inspector.

## Requirements

### Requirement: Site rules choose the browser
A site rule SHALL consist of one or more host patterns, the identity of a browser, and a "direct" flag. Host patterns SHALL be matched case-insensitively, with `*` matching any run of characters. bopen SHALL match rules against the host of the URL that would be opened with the default suggestions applied, so that redirect wrappers are matched by their destination. The first rule in order that matches, and whose browser is currently offered (discovered and not hidden), SHALL apply.

#### Scenario: Subdomain pattern
- **WHEN** a rule maps `*.atlassian.net` to Chrome and the link is `https://acme.atlassian.net/browse/X-1`
- **THEN** Chrome is pre-selected, even though Zen was used last

#### Scenario: Matched through a redirect wrapper
- **WHEN** the same rule exists and the link is a Google redirect to `https://acme.atlassian.net/browse/X-1`
- **THEN** Chrome is pre-selected

#### Scenario: Rule for a hidden browser
- **WHEN** a rule's browser is hidden or no longer installed
- **THEN** the rule is ignored and pre-selection falls back as without rules

### Requirement: Direct site rules skip the inspector
When the applying rule is direct, bopen SHALL open the link in the rule's browser without showing the inspector, with the default suggestions applied. This SHALL apply regardless of the `window` preference. It SHALL NOT apply when the inspector must be shown to report a problem (invalid link, configuration problems, or a failed launch).

#### Scenario: Direct rule with tracking
- **WHEN** a direct rule maps `jira.example.com` to Chrome and the link is `https://jira.example.com/x?utm_source=mail`
- **THEN** Chrome opens `https://jira.example.com/x` and no window is shown

#### Scenario: Launch failure
- **WHEN** a direct rule applies and launching its browser fails
- **THEN** the inspector is shown with the error

### Requirement: Invalid site rules are ignored and reported
A site rule without host patterns, with an invalid pattern, or without a browser SHALL NOT be applied, and SHALL be reported as a configuration problem.

#### Scenario: Rule without browser
- **WHEN** `config.toml` contains a site rule with an empty browser
- **THEN** the rule is not applied, and the inspector shows a configuration problem naming it
