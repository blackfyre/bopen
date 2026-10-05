# Design

## Context

bopen is a Go module (`github.com/blackfyre/bopen`, remote `git@github.com:blackfyre/bopen.git`) with one command at `./cmd/bopen`. See proposal.md for motivation. Facts about the binary that shape the pipeline:

- **Linux.** Gio links through cgo and pkg-config against `egl`, `wayland-egl`, `wayland-client`, `wayland-cursor`, `x11`, `xkbcommon`, `xkbcommon-x11`, `x11-xcb`, `xcursor` and `xfixes`. A local Fedora 44 build links these dynamically, plus `libffi`, and needs glibc symbols up to `GLIBC_2.34`.
- **Windows.** The build needs no cgo and already cross-compiles from Linux, with `-ldflags -H=windowsgui`.

## Goals / Non-Goals

**Goals:**
- One tag push produces every asset in the `release-distribution` spec, reproducibly.
- A copy-and-paste Linux install with checksum verification and no root.
- The existing gates (`gofmt`, `vet`, tests, Windows build) run in CI on every push.

**Non-Goals:**
- `linux/arm64` binaries. They need either a cross sysroot for the Gio libraries or a second native runner whose artefacts GoReleaser OSS cannot merge into the same release. Deferred; the install script reports it clearly.
- A Windows install script, winget or Scoop manifests; AUR, Flathub, Homebrew or apt/dnf repositories.
- Code signing (Authenticode) and signed checksums (cosign/GPG).
- Automatic changelog curation beyond GoReleaser's commit-based changelog.

## Decisions

### GoReleaser OSS (v2) on a single Linux runner

GoReleaser produces the archives, nfpm packages, checksums and GitHub release from one config, so there's no hand-written release logic. Two build IDs share the same `main`:

```yaml
version: 2
builds:
  - id: linux
    main: ./cmd/bopen
    goos: [linux]
    goarch: [amd64]
    env: [CGO_ENABLED=1]
    ldflags: ["-s -w -X main.version={{ .Version }}"]
  - id: windows
    main: ./cmd/bopen
    goos: [windows]
    goarch: [amd64, arm64]
    env: [CGO_ENABLED=0]
    ldflags: ["-s -w -H=windowsgui -X main.version={{ .Version }}"]
```

Archive and package `name_template` is `{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}`, without the version, so `releases/latest/download/...` URLs are stable (spec: Release assets). Archives use `tar.gz` on Linux and `zip` on Windows (`format_overrides`).

Alternative: a hand-rolled matrix workflow with `go build` and `softprops/action-gh-release`. That's more YAML to maintain, and nfpm packaging would have to be wired separately. Rejected.

### Runner: `ubuntu-22.04`

The glibc baseline is set by the build host. `ubuntu-22.04` ships glibc 2.35, which defines the portability requirement. The workflow installs the Gio build packages with `apt`:

```
libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev
libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
```

If GitHub retires that image, the fallback is to run the job in a `debian:bookworm` container (glibc 2.36) and raise the documented baseline accordingly.

### nfpm dependencies

The package dependencies are the runtime libraries behind the pkg-config modules above:

| Format | Dependencies |
|---|---|
| deb | `libegl1`, `libwayland-client0`, `libwayland-cursor0`, `libwayland-egl1`, `libx11-6`, `libx11-xcb1`, `libxkbcommon0`, `libxkbcommon-x11-0`, `libxcursor1`, `libxfixes3` |
| rpm | `libglvnd-egl`, `libwayland-client`, `libwayland-cursor`, `libwayland-egl`, `libX11`, `libX11-xcb`, `libxkbcommon`, `libxkbcommon-x11`, `libXcursor`, `libXfixes` |

Packages install `/usr/bin/bopen` plus `LICENSE` and `README.md` under `/usr/share/doc/bopen/`. They have no install scripts, so registration stays a per-user choice.

### Version injection

`cmd/bopen` gets `var version = "dev"`, set by `-X main.version=...`. The usage text's first line becomes `bopen <version>`. This is the only code change. It needs no new subcommand, so the `default-handler-registration` CLI requirement is untouched.

### Workflows

- **`release.yml`.** Triggered on `push: tags: ['v*']`. Steps:
  1. `actions/checkout` with `fetch-depth: 0`, which the changelog needs;
  2. `actions/setup-go` with `go-version-file: go.mod`;
  3. apt-install the Gio packages;
  4. `go test ./...`;
  5. `goreleaser/goreleaser-action` (`version: "~> v2"`, `args: release --clean`).

  Permissions are limited to `contents: write`, using `GITHUB_TOKEN` only, with no other secrets.
- **`ci.yml`.** Runs on push and pull request: `gofmt -l` (must be empty), `go vet ./...`, `go test ./...`, `GOOS=windows go build -ldflags -H=windowsgui ./cmd/bopen`, `goreleaser check`, and `shellcheck install.sh`.

### `install.sh`

The script is POSIX `sh`, uses `set -eu`, and wraps everything in a `main` function called on the last line, so a truncated `curl | sh` download can't run a partial script. Flow:

```
check uname -s == Linux, uname -m == x86_64      -> else explain, exit 1
pick downloader: curl -fsSL | wget -qO-          -> else exit 1
require tar, sha256sum                           -> else exit 1
base = ${BOPEN_BASE_URL:-https://github.com/blackfyre/bopen/releases}
url  = BOPEN_VERSION ? base/download/$BOPEN_VERSION : base/latest/download
tmp  = mktemp -d, trap cleanup
download bopen_linux_amd64.tar.gz and checksums.txt into tmp
grep the archive's line from checksums.txt | (cd tmp && sha256sum -c -)   -> mismatch: exit 1
tar -xzf into tmp; install -m 0755 tmp/bopen "$dir/bopen" (via tmp file + mv)
print "$dir/bopen" usage line (version), next step "bopen register",
PATH warning if $dir not in $PATH
```

`BOPEN_BASE_URL` exists so the script can be tested against a local HTTP server without GitHub. The binary is copied next to the destination and then renamed, so replacing a running binary is safe.

### Local verification

`goreleaser` and `shellcheck` are added to `mise.toml`. `goreleaser release --snapshot --clean` builds every asset locally into `dist/` (already git-ignored), without publishing. That's the gate for the configuration.

## Risks / Trade-offs

- **[glibc too new for older distributions]** Debian 11 and RHEL 8 users can't run the binary. → Documented baseline. Building from source remains possible, and the container fallback can lower the baseline if requested.
- **[`curl | sh` trust]** The script is fetched from `main`. → It verifies binaries against `checksums.txt` from the same release, and does nothing with root. The README also shows how to download and inspect the script before running it.
- **[Checksums aren't signatures]** A compromised release could replace both files. → Accepted for now. Signing is listed as a non-goal and can be added without changing the asset layout.
- **[Unsigned Windows executables trigger SmartScreen]** → Documented in the README. Signing is out of scope.
- **[Accidental release]** → Only tag pushes publish, and tags are created manually.
