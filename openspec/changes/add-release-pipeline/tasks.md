# Tasks

## 1. Version and tooling

- [x] 1.1 Add `var version = "dev"` to `cmd/bopen` and print `bopen <version>` as the first usage line. Verify with a test that the usage text contains the version, and check that `go build -ldflags "-X main.version=0.1.0"` makes `./bopen` print `bopen 0.1.0`.
- [x] 1.2 Add `goreleaser` and `shellcheck` to `mise.toml`. Verify that `mise exec -- goreleaser --version` and `mise exec -- shellcheck --version` succeed.

## 2. GoReleaser

- [x] 2.1 Write `.goreleaser.yaml`: linux (cgo) and windows (no cgo, `-H=windowsgui`) builds with version ldflags; version-less archive names (`tar.gz`/`zip`) including `LICENSE` and `README.md`; nfpm `.deb`/`.rpm` with the dependency lists from design.md; `checksums.txt`; and the GitHub release for `blackfyre/bopen`. Verify that `goreleaser check` passes.
- [x] 2.2 Run `goreleaser release --snapshot --clean`. Verify that `dist/` contains exactly the assets named in the `release-distribution` spec, and that `checksums.txt` covers all of them (`sha256sum -c`). Inspect the `.deb` and `.rpm` metadata (`dpkg-deb -I`, `rpm -qpR`) for the declared dependencies and `/usr/bin/bopen`.

## 3. Install script

- [x] 3.1 Write `install.sh` per design.md: POSIX `sh`, `main` wrapper, OS/architecture/tool checks, `BOPEN_VERSION`, `BOPEN_INSTALL_DIR`, `BOPEN_BASE_URL`, checksum verification, atomic install, next-step and PATH messages, and no registration. Verify that `shellcheck install.sh` reports nothing.
- [x] 3.2 Add `scripts/test-install.sh`, which serves generated fake release assets (two versions plus a tampered copy) with `python3 -m http.server` in the layout `latest/download/` and `download/<tag>/`, and runs `install.sh` into temporary directories. It must cover: fresh install, pinned version, re-run upgrade, tampered archive (non-zero exit, nothing installed), simulated `aarch64` via a `uname` stub (non-zero exit, no download), missing `sha256sum`, and the PATH warning. Verify that the script passes all of its cases, and run `install.sh` once against the real snapshot `dist/` assets.

## 4. GitHub workflows

- [x] 4.1 Write `.github/workflows/ci.yml` (gofmt, vet, tests, Windows build, `goreleaser check`, shellcheck) on `ubuntu-22.04`, with the Gio apt packages. Verify that `mise exec -- actionlint` (or `npx actionlint`) reports no errors.
- [x] 4.2 Write `.github/workflows/release.yml` (tag `v*` trigger, `contents: write`, `fetch-depth: 0`, setup-go from `go.mod`, apt packages, tests, `goreleaser-action` v2 with `release --clean`). Verify with actionlint, and confirm that the trigger is limited to `v*` tags.

## 5. Documentation

- [x] 5.1 Update the README with:
  - the copy-and-paste install command;
  - a "download and inspect first" variant;
  - `BOPEN_VERSION` and `BOPEN_INSTALL_DIR`;
  - `.deb`/`.rpm` installation;
  - Windows zip download, with a SmartScreen note;
  - upgrading (re-run the script);
  - uninstalling (`bopen unregister`, then remove the binary);
  - the glibc 2.35 baseline, with the note that `aarch64` is not published yet;
  - the maintainer release procedure: tag and push.

  Verify that every command in the README matches the script's variables and the asset names in the spec.

## 6. Verification

- [x] 6.1 Run `gofmt -l .` (empty), `go vet ./...`, `go test ./...`, `GOOS=windows go build -ldflags -H=windowsgui ./cmd/bopen`, `goreleaser check`, `shellcheck install.sh` and `scripts/test-install.sh`, all passing. Publishing a real release (pushing a tag) is left to the maintainer.
