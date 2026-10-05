# bopen

bopen registers itself as your default web browser. When you click a link in
any application, bopen shows it to you first:

- redirect wrappers (Google, Facebook, Outlook Safe Links, YouTube) and
  tracking parameters (`utm_*`, `fbclid`, `gclid`, …) are highlighted, each
  with the reason it is suggested for removal;
- every suggestion can be ticked or unticked, and the resulting link updates
  as you go (affiliate tags are shown but kept unless you untick them);
- you pick the browser to open it in; the browser you used last is
  pre-selected.

Installed browsers are discovered automatically: on Linux from desktop
entries (system packages, Flatpak, Snap and per-user installs), on Windows
from the registered browsers.

## Installing

### Linux

```
curl -fsSL https://raw.githubusercontent.com/blackfyre/bopen/main/install.sh | sh
```

The script downloads the latest release, verifies it against the release's
`checksums.txt`, and installs `bopen` into `~/.local/bin` without root. It
does not change your default browser; run `bopen register` afterwards.

To read the script before running it:

```
curl -fsSLO https://raw.githubusercontent.com/blackfyre/bopen/main/install.sh
less install.sh
sh install.sh
```

Options, as environment variables:

- `BOPEN_VERSION=v0.1.0` installs a specific release instead of the latest.
- `BOPEN_INSTALL_DIR=/some/dir` installs somewhere other than `~/.local/bin`.

Run the script again to upgrade.

Packages are published with each release as well, and pull in the runtime
libraries:

```
# Debian/Ubuntu
curl -fsSLO https://github.com/blackfyre/bopen/releases/latest/download/bopen_linux_amd64.deb
sudo apt install ./bopen_linux_amd64.deb

# Fedora/openSUSE
sudo dnf install https://github.com/blackfyre/bopen/releases/latest/download/bopen_linux_amd64.rpm
```

The Linux binary is built for x86-64 and needs glibc 2.35 or newer (Ubuntu
22.04, Debian 12, Fedora 36 and later) and the usual Wayland, X11, EGL and
xkbcommon libraries that every desktop already has. No `aarch64` binary is
published yet; build from source instead.

### Windows

Download `bopen_windows_amd64.zip` (or `bopen_windows_arm64.zip`) from the
[latest release](https://github.com/blackfyre/bopen/releases/latest), put
`bopen.exe` somewhere permanent, and run `bopen.exe register` from a terminal
in that folder. The executable is not code-signed yet, so SmartScreen may warn
on first start.

### Uninstalling

Run `bopen unregister` first, so your previous default browser is restored,
then delete the binary (`~/.local/bin/bopen` when installed by the script) or
remove the package. Preferences in `~/.config/bopen/` (`%AppData%\bopen\`
on Windows) can be deleted as well.

## Usage

```
bopen <url>       inspect a link and open it in a browser
bopen register    make bopen the default handler for web links
bopen unregister  undo 'bopen register'
```

Keyboard: `Enter` opens, `Esc` cancels, `Up`/`Down` or `1`–`9` choose the
browser.

### Registering

- **Linux:** `bopen register` installs `bopen.desktop` in
  `~/.local/share/applications` and makes it the handler for `http` and
  `https` with `xdg-mime` (from xdg-utils). The previous default browser is
  remembered, and `bopen unregister` restores it.
- **Windows:** `bopen register` registers bopen for the current user (no
  administrator rights needed) and opens Settings › Default apps. Windows does
  not let programs make themselves the default, so select bopen there.

Register the binary from its final location; the registration stores its
absolute path.

## Configuration

Preferences live in `config.toml` in the bopen directory of your user
configuration directory (`~/.config/bopen/` on Linux,
`%AppData%\bopen\` on Windows):

```toml
# "always" (default) shows the inspector for every link.
# "when-suggestions" opens links with nothing to remove directly in the
# last-used browser and shows the inspector only when there is something
# to suggest.
window = "always"
```

bopen keeps the last-used browser in `state.toml` next to it.

bopen makes no network requests.

## Building

bopen needs Go (see `mise.toml`).

On Linux the GUI toolkit ([Gio](https://gioui.org)) needs these development
packages. On Fedora:

```
sudo dnf install gcc pkgconf-pkg-config wayland-devel libX11-devel libxkbcommon-x11-devel \
    mesa-libGLES-devel mesa-libEGL-devel libXcursor-devel libXfixes-devel vulkan-headers
```

On Debian/Ubuntu:

```
sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev \
    libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
```

Then:

```
go build -o bopen ./cmd/bopen                                 # Linux
GOOS=windows go build -ldflags -H=windowsgui -o bopen.exe ./cmd/bopen   # Windows
go test ./...
```

The Windows build needs no C toolchain and can be cross-compiled from Linux.
`-H=windowsgui` stops a console window flashing up on every link.

`goreleaser release --snapshot --clean` builds every release asset into
`dist/` without publishing anything, and `scripts/test-install.sh` tests
`install.sh` against a local server. `mise install` provides GoReleaser,
ShellCheck and actionlint.

## Releasing

Releases are built by GitHub Actions when a version tag is pushed:

```
git tag v0.1.0
git push origin v0.1.0
```

The workflow runs the tests, then GoReleaser publishes the archives, `.deb`
and `.rpm` packages and `checksums.txt` to the GitHub release.

## Licence

MIT, see [LICENSE](LICENSE).
