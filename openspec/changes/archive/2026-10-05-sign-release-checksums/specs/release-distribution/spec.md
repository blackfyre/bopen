# Spec Delta

## ADDED Requirements

### Requirement: Release checksums are signed
Each release SHALL contain `checksums.txt.sigstore.json`, a Sigstore bundle signing `checksums.txt`. The signature SHALL be made keylessly by the release workflow (`.github/workflows/release.yml` on a `v*` tag), with the GitHub Actions OIDC issuer, so that it can be verified against that identity without a published key.

#### Scenario: Verifying by hand
- **WHEN** a user runs `cosign verify-blob` on a release's `checksums.txt` with its bundle, certificate identity `https://github.com/blackfyre/bopen/.github/workflows/release.yml@refs/tags/<tag>` and issuer `https://token.actions.githubusercontent.com`
- **THEN** verification succeeds

### Requirement: Install script verifies the signature when possible
When `cosign` is available, `install.sh` SHALL download the bundle and verify `checksums.txt` against the release workflow's identity before using it, and SHALL exit with a non-zero status, installing nothing, when verification fails. When `cosign` is not available, it SHALL continue with checksum verification only and state that the signature was not checked.

#### Scenario: Signature check fails
- **WHEN** `cosign` is installed and rejects the bundle
- **THEN** the script exits non-zero, and nothing is written to the install directory

#### Scenario: No cosign
- **WHEN** `cosign` is not installed
- **THEN** the archive is checksum-verified and installed, and the output notes that the signature was not checked

## MODIFIED Requirements

### Requirement: Install script refuses unsupported systems clearly
The script SHALL exit with a non-zero status and an explanatory message, without downloading the archive, when any of these is true:
- the system is not Linux;
- the CPU architecture has no published binary;
- a required tool is missing. The required tools are `curl` or `wget`, plus `tar`, `gzip` and `sha256sum`.

#### Scenario: Unsupported architecture
- **WHEN** the script runs on `riscv64` Linux
- **THEN** it reports that no binary is published for `riscv64` yet, points to building from source, and exits non-zero
