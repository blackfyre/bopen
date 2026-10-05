# Spec Delta

## MODIFIED Requirements

### Requirement: Only web URLs are handed off
The system SHALL hand off only absolute URLs with scheme `http` or `https` and a non-empty host. The handed-off text SHALL be the link as received, with only these changes:
- surrounding whitespace is trimmed;
- the scheme is lower-cased;
- spaces, tabs and double quotes are percent-encoded;
- suggestions the user accepted are removed.

The argument therefore always begins with `http://` or `https://`. A malformed percent-escape SHALL NOT cause a link to be rejected.

#### Scenario: Option-like input rejected
- **WHEN** the input is `--gpu-launcher=calc.exe`
- **THEN** no browser is launched

#### Scenario: Missing host rejected
- **WHEN** the input is `https:///path`
- **THEN** no browser is launched

#### Scenario: Unchanged link passes through byte for byte
- **WHEN** the input is `https://example.com/café/ü?q=ö#` and no suggestion applies
- **THEN** the browser receives exactly `https://example.com/café/ü?q=ö#`

#### Scenario: Malformed escape still opens
- **WHEN** the input is `https://example.com/a%zz?utm_source=x`
- **THEN** the link is accepted, and accepting the suggestion hands off `https://example.com/a%zz`
