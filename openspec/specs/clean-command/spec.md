# clean-command Specification

## Purpose

Cleans links outside the inspector, for sharing and scripting, with the same rules and no network access.

## Requirements

### Requirement: Clean links from arguments or standard input
`bopen clean` SHALL print one line per input link: the link with the default suggestions applied (tracking parameters and redirect wrappers removed, affiliate parameters kept). Input links SHALL come from the arguments, or, when there are none, from standard input, one per line, ignoring blank lines. It SHALL use the user's rules, the enabled built-in rules and, when enabled, the cached ClearURLs list, and SHALL NOT make network requests.

#### Scenario: Argument
- **WHEN** the user runs `bopen clean 'https://example.com/a?utm_source=x&id=1'`
- **THEN** `https://example.com/a?id=1` is printed, and the exit status is zero

#### Scenario: Pipe
- **WHEN** two links are piped into `bopen clean`
- **THEN** two cleaned links are printed, in the same order

### Requirement: Invalid input is reported per line
For input that is not an absolute http or https URL, `bopen clean` SHALL print the input unchanged to standard output, so output lines stay aligned with input lines, print an error naming it to standard error, and exit with a non-zero status after processing all input.

#### Scenario: Mixed input
- **WHEN** the input is `https://example.com/?fbclid=x` followed by `not a link`
- **THEN** standard output has `https://example.com/` and `not a link`, standard error names `not a link`, and the exit status is non-zero

### Requirement: Explanations
With `--explain`, `bopen clean` SHALL print each applied suggestion's affected text, kind, source and reason to standard error.

#### Scenario: Explain
- **WHEN** the user runs `bopen clean --explain 'https://example.com/?fbclid=x'`
- **THEN** standard error names `fbclid=x` with its reason, and standard output has only the cleaned link
