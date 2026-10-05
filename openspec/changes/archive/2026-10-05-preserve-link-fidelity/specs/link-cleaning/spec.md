# Spec Delta

## ADDED Requirements

### Requirement: Analysis tolerates malformed escapes
Analysis SHALL work on links that contain malformed percent-escapes, treating a `%` that does not start a valid escape as a literal character. Such links SHALL produce the same suggestions as an equivalent well-formed link.

#### Scenario: Stray percent in the path
- **WHEN** `https://example.com/100%real?fbclid=x` is analysed
- **THEN** a `tracking` suggestion covers `fbclid=x`
