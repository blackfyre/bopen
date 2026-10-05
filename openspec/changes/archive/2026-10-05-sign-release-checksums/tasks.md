# Tasks

## 1. Signing

- [x] 1.1 Add a GoReleaser `signs` entry (cosign `sign-blob --bundle`, checksum artifact), plus `id-token: write` and `sigstore/cosign-installer` in `release.yml`, and `--skip=sign` for the CI snapshot. Verify with `goreleaser check` and actionlint. The signature itself can only be produced by the next tagged release.

## 2. Install script

- [x] 2.1 Verify the bundle with cosign when available (identity regexp for release tags, GitHub OIDC issuer), fail closed on rejection, and print a note without cosign. Verify with shellcheck and `scripts/test-install.sh` cases using a stub cosign (accepting, rejecting, absent).

## 3. Docs and verification

- [x] 3.1 Document manual verification in the README, run all gates, push, and confirm CI passes.
