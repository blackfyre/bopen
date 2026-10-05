# Tasks

## 1. Build

- [x] 1.1 Add `scripts/setup-arm64-cross.sh` and the GoReleaser `linux/arm64` target with its cross-compiler override. Verify with `goreleaser check` and a snapshot build in an Ubuntu 22.04 container that produces an aarch64 binary, `.deb` (arm64) and `.rpm` (aarch64), with the same library dependencies and a glibc ≤ 2.35 requirement.

## 2. Install script

- [x] 2.1 Map `x86_64`/`amd64` and `aarch64`/`arm64` to their archives, and extend `scripts/test-install.sh` with an arm64 install (stubbed `uname`) and a `riscv64` refusal. Verify with shellcheck and the test script.

## 3. Workflows and docs

- [x] 3.1 Run the setup script in `release.yml`, and add a CI job that runs the setup script and `goreleaser release --snapshot --clean`. Update the README. Verify with actionlint, then push and confirm all CI jobs pass, including the snapshot job.
