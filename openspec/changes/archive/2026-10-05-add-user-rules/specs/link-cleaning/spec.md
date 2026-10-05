# Spec Delta

## ADDED Requirements

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
