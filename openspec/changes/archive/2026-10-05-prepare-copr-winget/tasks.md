# Tasks

## 1. COPR

- [x] 1.1 Write the RPM spec (build from source, `%check`, licence and docs). Verify with `rpmbuild -ba` in a `fedora:latest` container using Fedora's Go (1.26.8). The package installs and runs, and `bopen clean` works.
- [x] 1.2 Lower the `go` directive to 1.26.0, with `toolchain go1.27.1`, as required by the Fedora build. Verify that the module builds locally and that Fedora's Go 1.26 builds and tests it.
- [x] 1.3 Add `.packit.yaml` (`copr_build` on release; Fedora x86_64 and aarch64; network enabled).

## 2. winget

- [x] 2.1 Add the GoReleaser `winget` section, publishing only when `WINGET_GITHUB_TOKEN` is set, and pass the secret in `release.yml`. Verify with `goreleaser check`, and with a snapshot build that writes the three manifests (installer, locale, version) with x64 and arm64 zip installers.

## 3. Docs and verification

- [x] 3.1 Document the activation steps for COPR (Packit app, project) and winget (fork, token, first manual submission), and update the Go requirement in the README. Run all gates, push, and confirm all CI jobs pass.
