# Proposal

## Why

ARM64 Linux desktops (Raspberry Pi 5, Ampere workstations, Asahi and Snapdragon laptops) currently have to build bopen from source. The install script refuses them.

## What Changes

- Releases also contain `bopen_linux_arm64.tar.gz`, `.deb` and `.rpm`. They are cross-compiled on the amd64 release runner using Ubuntu's arm64 multiarch packages (`scripts/setup-arm64-cross.sh`), so a single GoReleaser run still produces every asset.
- `install.sh` installs the arm64 build on `aarch64`/`arm64` machines.
- CI runs a full GoReleaser snapshot build, including the arm64 cross-compile, on every push, so a broken release build is caught before tagging.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `release-distribution`: The release assets and the install script's supported architectures.

## Impact

Touches `.goreleaser.yaml` (arm64 target with a cross-compiler override), the workflows, `install.sh` and its tests, and the README.
