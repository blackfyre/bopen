# Proposal

## Why

Fedora and Windows users expect to install bopen with their package manager (`dnf` via COPR, `winget`). Publishing to either needs maintainer accounts and tokens, but everything else can be prepared in the repository, so that activation is a few clicks.

## What Changes

- **COPR:** `packaging/fedora/bopen.spec` builds bopen from source with Fedora's own Go toolchain. `.packit.yaml` makes Packit build it in the COPR project `blackfyre/bopen` for every GitHub release, on Fedora x86_64 and aarch64. It is inactive until the Packit GitHub app is installed.
- **winget:** GoReleaser generates winget manifests (`blackfyre.bopen`, portable zip installers for x64 and arm64) on every release. It opens a pull request to `microsoft/winget-pkgs` only when the `WINGET_GITHUB_TOKEN` secret exists.
- **Go minimum:** the module's minimum Go version drops from 1.27.1 to 1.26.0, which is what the dependencies actually require. Fedora 44's Go 1.26 can then build it, and `go install` works with Go 1.26. `toolchain go1.27.1` keeps CI on the current release.
- **README:** maintainer activation steps for both channels.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. This is packaging and tooling only, so the change sets `skip_specs`.

## Impact

New `packaging/fedora/bopen.spec` and `.packit.yaml`; a `winget` section in `.goreleaser.yaml`; `release.yml` passes the optional token; `go.mod`; README.
