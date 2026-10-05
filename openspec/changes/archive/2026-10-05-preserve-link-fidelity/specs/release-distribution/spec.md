# Spec Delta

## ADDED Requirements

### Requirement: Source installs report their module version
A binary built without a release version linked in SHALL report the module version recorded by the Go toolchain, without a leading `v`. It SHALL report `dev` only when no module version is recorded, such as in a build outside version control. Builds from a local checkout report the version Go derives from version control (for example `0.3.1-0.20261005173729-ca702f19e4c5+dirty`).

#### Scenario: go install of a tag
- **WHEN** bopen is installed with `go install github.com/blackfyre/bopen/cmd/bopen@v0.3.0` and run without arguments
- **THEN** the usage text starts with `bopen 0.3.0`
