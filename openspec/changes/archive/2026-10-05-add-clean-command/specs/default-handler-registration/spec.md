# Spec Delta

## MODIFIED Requirements

### Requirement: Command-line usage
bopen SHALL accept one URL argument (open mode), one of the subcommands `register`, `unregister` or `settings`, or the subcommand `clean` followed by optional `--explain` and URLs. Any other invocation, including no arguments, SHALL print usage to standard error and exit with a non-zero status.

#### Scenario: No arguments
- **WHEN** bopen is run without arguments
- **THEN** usage is printed, and the exit status is non-zero

#### Scenario: Settings subcommand
- **WHEN** bopen is run as `bopen settings`
- **THEN** the settings view opens

#### Scenario: Clean subcommand
- **WHEN** bopen is run as `bopen clean https://example.com/`
- **THEN** no window opens, and the cleaned link is printed
