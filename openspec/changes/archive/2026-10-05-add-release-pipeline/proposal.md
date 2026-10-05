# Proposal

## Why

bopen can currently only be installed by building it from source, which needs Go and, on Linux, the Gio development headers. Users should be able to install it with one copy-and-paste command, and that needs published release binaries. The repository now has its GitHub remote (`github.com/blackfyre/bopen`), so a release pipeline can be built.

## What Changes

- A GoReleaser configuration that builds release artefacts when a `v*` tag is pushed:
  - `linux/amd64` with cgo, for Gio;
  - `windows/amd64` and `windows/arm64` without cgo, linked as GUI-subsystem executables;
  - archives with stable, version-less names, `.deb` and `.rpm` packages that declare their runtime libraries, and a `checksums.txt`.
- A GitHub Actions workflow that runs GoReleaser on tag pushes, plus a CI workflow that runs the existing verification gates on pushes and pull requests.
- The release version injected into the binary at build time.
- An `install.sh` script in the repository root that:
  - downloads the latest (or a pinned) Linux release;
  - verifies its checksum;
  - installs the binary into `~/.local/bin` without root;
  - tells the user to run `bopen register`, rather than registering automatically.
- README: a copy-and-paste install command, package installation, upgrading and uninstalling.

## Capabilities

### New Capabilities
- `release-distribution`: What a release contains, and how the install script behaves.

### Modified Capabilities

None.

## Impact

- New files: `.goreleaser.yaml`, `.github/workflows/release.yml`, `.github/workflows/ci.yml`, `install.sh`.
- `cmd/bopen` gains a `version` variable set through `-ldflags`. No CLI behaviour changes.
- `mise.toml` gains `goreleaser` and `shellcheck` as development tools.
- Publishing a release requires pushing a tag. That stays a manual, deliberate step by the maintainer.
- Depends on nothing in the other open changes, and can be applied independently of them.
