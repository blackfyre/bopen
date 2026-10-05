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

## Licence

MIT, see [LICENSE](LICENSE).
