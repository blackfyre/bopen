# Proposal

## Why

The install script verifies archives against `checksums.txt`, but that file comes from the same release. Anyone able to replace release assets could replace both. Signing `checksums.txt` closes the gap. Sigstore's keyless signing does this with the release workflow's GitHub identity, so no keys, certificates or secrets have to be managed.

## What Changes

- **Signed checksums:** the release workflow signs `checksums.txt` with cosign (keyless, GitHub Actions OIDC) and publishes `checksums.txt.sigstore.json`. Because every asset is listed in the checksums, this covers all of them.
- **Install script:** when `cosign` is installed, `install.sh` verifies the signature against the release workflow's identity and aborts if it fails. Without cosign it says the signature was not checked.
- **README:** shows how to verify a download by hand.
- **Not included:** Authenticode signing of `bopen.exe` (needs a code-signing certificate) and GPG signing of `.deb`/`.rpm` (needs a maintained key). Both remain open.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `release-distribution`: Signed checksums, and signature verification in the install script.

## Impact

`.goreleaser.yaml` (`signs`), `release.yml` (`id-token: write`, cosign installer), `install.sh` and its tests, and the README. The CI snapshot job skips signing, because it has no release identity.
