# Spec Delta

## MODIFIED Requirements

### Requirement: Release assets
Each release SHALL contain these assets:

| Asset | Contents |
|---|---|
| `bopen_linux_amd64.tar.gz` and `bopen_linux_arm64.tar.gz` | the Linux binary, `LICENSE` and `README.md` |
| `bopen_windows_amd64.zip` and `bopen_windows_arm64.zip` | the Windows executable, `LICENSE` and `README.md` |
| `bopen_linux_amd64.deb`, `bopen_linux_amd64.rpm`, `bopen_linux_arm64.deb` and `bopen_linux_arm64.rpm` | installing the binary as `/usr/bin/bopen` |
| `checksums.txt` | the SHA-256 digest of every other asset |

Asset names SHALL NOT contain the version, so `releases/latest/download/<asset>` URLs stay stable.

#### Scenario: Stable latest URL
- **WHEN** a new release is published
- **THEN** `https://github.com/blackfyre/bopen/releases/latest/download/bopen_linux_amd64.tar.gz` downloads that release's Linux archive

#### Scenario: arm64 assets
- **WHEN** a new release is published
- **THEN** `bopen_linux_arm64.tar.gz` contains an aarch64 executable linked against the same libraries as the amd64 build

### Requirement: Install script installs a verified binary without root
`install.sh` SHALL:
- run under POSIX `sh`;
- download the Linux archive for the machine's architecture (`x86_64`/`amd64` or `aarch64`/`arm64`) and `checksums.txt`, of the latest release or of the release named by `BOPEN_VERSION`;
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

#### Scenario: arm64 machine
- **WHEN** the script runs on `aarch64` Linux
- **THEN** it downloads and installs `bopen_linux_arm64.tar.gz`

### Requirement: Install script refuses unsupported systems clearly
The script SHALL exit with a non-zero status and an explanatory message, without downloading the archive, when any of these is true:
- the system is not Linux;
- the CPU architecture has no published binary;
- a required tool is missing. The required tools are `curl` or `wget`, plus `tar` and `sha256sum`.

#### Scenario: Unsupported architecture
- **WHEN** the script runs on `riscv64` Linux
- **THEN** it reports that no binary is published for `riscv64` yet, points to building from source, and exits non-zero
