# release-distribution Specification

## Purpose

Defines what each bopen release publishes and how the Linux install script obtains, verifies and installs it, so users can install bopen without building it.

## Requirements

### Requirement: Releases are built from version tags
Pushing a tag of the form `v<major>.<minor>.<patch>` (optionally with a pre-release suffix) SHALL produce a GitHub release for that tag, containing the release assets. No other event SHALL publish a release.

#### Scenario: Tag push publishes
- **WHEN** the maintainer pushes the tag `v0.1.0`
- **THEN** a GitHub release `v0.1.0` is created with the release assets attached

#### Scenario: Branch push does not publish
- **WHEN** a commit is pushed to `main` without a tag
- **THEN** no release is created

### Requirement: Release assets
Each release SHALL contain these assets:

| Asset | Contents |
|---|---|
| `bopen_linux_amd64.tar.gz` | the Linux binary, `LICENSE` and `README.md` |
| `bopen_windows_amd64.zip` and `bopen_windows_arm64.zip` | the Windows executable, `LICENSE` and `README.md` |
| `bopen_linux_amd64.deb` and `bopen_linux_amd64.rpm` | installing the binary as `/usr/bin/bopen` |
| `checksums.txt` | the SHA-256 digest of every other asset |

Asset names SHALL NOT contain the version, so `releases/latest/download/<asset>` URLs stay stable.

#### Scenario: Stable latest URL
- **WHEN** a new release is published
- **THEN** `https://github.com/blackfyre/bopen/releases/latest/download/bopen_linux_amd64.tar.gz` downloads that release's Linux archive

### Requirement: Release binaries report their version
A release binary SHALL embed the tag version. The version SHALL appear in the usage text that bopen prints.

#### Scenario: Usage shows version
- **WHEN** a binary built from tag `v0.1.0` is run without arguments
- **THEN** the usage text includes `0.1.0`

### Requirement: Linux binary portability
The Linux release binary SHALL run on distributions whose glibc is version 2.35 or newer and that provide the Wayland, X11, EGL and xkbcommon client libraries. The `.deb` and `.rpm` packages SHALL declare those libraries as dependencies.

#### Scenario: Package pulls runtime libraries
- **WHEN** `bopen_linux_amd64.deb` is installed with `apt` on a minimal Ubuntu 22.04 desktop
- **THEN** any missing runtime libraries are installed, and `bopen` runs

### Requirement: Windows executables are GUI applications
The Windows executables SHALL be built for the GUI subsystem, so opening a link does not show a console window.

#### Scenario: No console window
- **WHEN** Windows invokes `bopen.exe` for a clicked link
- **THEN** only the inspector window appears

### Requirement: Install script installs a verified binary without root
`install.sh` SHALL:
- run under POSIX `sh`;
- download the Linux archive and `checksums.txt` of the latest release, or of the release named by `BOPEN_VERSION`;
- verify the archive's SHA-256 digest;
- install the `bopen` binary into `BOPEN_INSTALL_DIR`, defaulting to `~/.local/bin`, without requiring root.

It SHALL replace an existing binary there, which makes running it again an upgrade.

#### Scenario: Fresh install
- **WHEN** a user runs `curl -fsSL https://raw.githubusercontent.com/blackfyre/bopen/main/install.sh | sh`
- **THEN** `~/.local/bin/bopen` exists and is executable, and the script prints the installed version

#### Scenario: Pinned version
- **WHEN** the script runs with `BOPEN_VERSION=v0.1.0`
- **THEN** the `v0.1.0` assets are downloaded, regardless of the latest release

#### Scenario: Checksum mismatch
- **WHEN** the downloaded archive does not match its entry in `checksums.txt`
- **THEN** the script exits with a non-zero status, and nothing is written to the install directory

### Requirement: Install script refuses unsupported systems clearly
The script SHALL exit with a non-zero status and an explanatory message, without downloading the archive, when any of these is true:
- the system is not Linux;
- the CPU architecture has no published binary;
- a required tool is missing. The required tools are `curl` or `wget`, plus `tar` and `sha256sum`.

#### Scenario: Unsupported architecture
- **WHEN** the script runs on `aarch64` Linux
- **THEN** it reports that no binary is published for `aarch64` yet, points to building from source, and exits non-zero

### Requirement: Install script does not change the default browser
The script SHALL NOT register bopen as the default browser. It SHALL print the `bopen register` command as the next step, and SHALL warn when the install directory is not on `PATH`.

#### Scenario: Next steps printed
- **WHEN** installation succeeds
- **THEN** the output tells the user to run `bopen register`, and the default browser is unchanged

#### Scenario: Install directory not on PATH
- **WHEN** `~/.local/bin` is not on `PATH`
- **THEN** the script warns about it and shows the full path to run `register`

### Requirement: Source installs report their module version
A binary built without a release version linked in SHALL report the module version recorded by the Go toolchain, without a leading `v`. It SHALL report `dev` only when no module version is recorded, such as in a build outside version control. Builds from a local checkout report the version Go derives from version control (for example `0.3.1-0.20261005173729-ca702f19e4c5+dirty`).

#### Scenario: go install of a tag
- **WHEN** bopen is installed with `go install github.com/blackfyre/bopen/cmd/bopen@v0.3.0` and run without arguments
- **THEN** the usage text starts with `bopen 0.3.0`
