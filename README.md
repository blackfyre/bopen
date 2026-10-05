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
sudo apt install ./bopen_linux_amd64.deb      # bopen_linux_arm64.deb on arm64

# Fedora/openSUSE
sudo dnf install https://github.com/blackfyre/bopen/releases/latest/download/bopen_linux_amd64.rpm
```

Linux binaries are built for x86-64 and arm64 (aarch64) and need glibc 2.35
or newer (Ubuntu 22.04, Debian 12, Fedora 36 and later) and the usual
Wayland, X11, EGL and xkbcommon libraries that every desktop already has. The
install script picks the right one; packages are named
`bopen_linux_<amd64|arm64>.<deb|rpm>`.

### Windows

Download `bopen_windows_amd64.zip` (or `bopen_windows_arm64.zip`) from the
[latest release](https://github.com/blackfyre/bopen/releases/latest), put
`bopen.exe` somewhere permanent, and run `bopen.exe register` from a terminal
in that folder. The executable is not code-signed yet, so SmartScreen may warn
on first start.

### From source with `go install`

Handy for testing a branch or installing on a platform without a published
binary. It needs Go 1.27.1 or newer (the version in `go.mod`) and, on Linux, the Gio development
packages listed under [Building](#building):

```
go install github.com/blackfyre/bopen/cmd/bopen@latest    # or @v0.1.0, @main, @<commit>
```

The binary lands in `$(go env GOBIN)`, or `$(go env GOPATH)/bin`
(usually `~/go/bin`) when `GOBIN` is unset; register it from there with
`bopen register`. On Windows, add `-ldflags=-H=windowsgui` so no console
window appears on every link. Binaries built this way report their version as
`dev`.

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
bopen settings    change bopen's preferences
bopen clean [--explain] [url...]
                  print links without tracking parts (reads stdin
                  when no url is given)
```

bopen follows your desktop's appearance: light or dark mode, the accent
colour and the high-contrast setting, read from the XDG desktop portal on
Linux (COSMIC, GNOME, KDE and others) and from the personalisation settings
on Windows. Changes apply while the window is open. Every text colour keeps a
contrast ratio of at least 4.5:1 in every mode and with any accent colour.

Keyboard: `Enter` opens, `Esc` cancels, `Up`/`Down` or `1`–`9` choose the
browser, `P` toggles a private window, `C` copies the cleaned link.

### Cleaning without opening

`bopen clean` prints links with the default suggestions applied (tracking
parameters and redirect wrappers removed, affiliate tags kept), using the
same rules as the inspector and without any network access. It takes links
as arguments or, one per line, on standard input, so it works in pipes and
editor integrations; `--explain` lists what was removed, and why, on standard
error:

```
$ bopen clean 'https://example.com/a?utm_source=news&id=7'
https://example.com/a?id=7
$ wl-paste | bopen clean | wl-copy        # clean the clipboard (Wayland)
```

In the inspector, **Copy** (or `C`) puts the cleaned link on the clipboard
without opening it.

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

Open the settings with the cog icon in the inspector, or run
`bopen settings`. There you can choose when the inspector appears, hide and
reorder browsers, switch individual built-in rules off, and register bopen
again if another browser has taken over. Changes are saved immediately.

The settings are stored in `config.toml` in the bopen directory of your user
configuration directory (`~/.config/bopen/` on Linux, `%AppData%\bopen\`
on Windows):

```toml
# "always" (default) shows the inspector for every link.
# "when-suggestions" opens links with nothing to remove directly in the
# last-used browser and shows the inspector only when there is something
# to suggest.
window = "always"

[browsers]
hidden = ["firefox.desktop"]                              # not offered
order  = ["app.zen_browser.zen.desktop", "brave-browser.desktop"]  # shown first

[rules]
disabled = ["utm"]                                        # built-in rule ids
```

Browsers are identified by their desktop file ID on Linux and their
registered name on Windows; entries for browsers that are not installed are
kept, so a reinstalled browser keeps its settings. The file belongs to the
settings screen: it can be edited by hand, but comments are not kept when the
settings screen saves it.

### Profiles and private windows

Browsers with more than one profile (Firefox, Zen and LibreWolf from
`profiles.ini`; Chrome, Chromium, Brave, Edge and Vivaldi from `Local State`)
get one entry per profile, such as **Zen Browser · Work**, which opens that
profile. Profile entries work like any other browser: they can be hidden,
reordered and used in site rules.

Tick **Open in a private window** (or press `P`) to open the link in the
selected browser's private window. On Linux this uses the browser's own
private-window desktop action; on Windows, its known command-line flag.

### Site rules

To always open a site in a particular browser, tick **Always open <host> in
this browser** in the inspector before opening the link, or add a rule under
**Site rules** in the settings. A rule maps host patterns (such as
`*.atlassian.net`) to a browser and pre-selects that browser for matching
links, ahead of the last-used one. Rules match the link's destination, so a
Google or Outlook redirect to Jira counts as Jira. With **Open directly**
ticked, matching links skip the inspector and open straight away, with the
usual tracking parameters still removed. Rules are tried in order and stored
in `config.toml`:

```toml
[[sites]]
id      = "s-1a2b3c"                 # assigned by bopen
hosts   = ["*.atlassian.net"]
browser = "google-chrome.desktop"    # desktop file ID, or registry name on Windows
direct  = true
```

### Your own rules

Right-click any parameter in the inspector's link and choose
**Always flag "…"** to add a rule for it: the parameter name (a trailing `*`
matches a prefix), whether it applies to this host only or to any host,
whether it is tracking (removed by default) or affiliate (kept by default),
and a reason, which is required and shown next to every suggestion it makes.
Right-clicking a suggestion offers **Disable this rule** for built-in rules
and **Edit this rule…** for your own. The settings screen lists your rules
under **Your rules**, where they can be added, edited (including wildcard
hosts such as `*.example.com`) and deleted.

Your rules take precedence over the built-in ones for the same parameter.
They are stored in `config.toml`:

```toml
[[rules.user]]
id     = "u-3f9a1c"              # assigned by bopen
kind   = "tracking"              # or "affiliate"
param  = "ref"
hosts  = ["news.example.com"]    # omit for any host
reason = "Referrer tracking on this news site."
```

A rule without a reason or parameter, or with another kind, is ignored and
reported when bopen starts.

bopen keeps the last-used browser in `state.toml` next to it.

### Short links (optional)

Shortened links (`bit.ly`, `t.co`, `lnkd.in` and other known shorteners)
hide where they lead. Tick **Offer to expand short links** in the settings
(or set `expand_short_links = true` in `config.toml`), and the inspector
shows **Expand** for such links. Pressing it asks only the shortener where
the link leads: a `HEAD` request without cookies, following redirects only
between known shorteners, and stopping before the destination. The
destination then replaces the link and is cleaned as usual. Nothing is
contacted unless you press Expand.

### The ClearURLs list (optional)

The built-in rules are deliberately few, so each one can explain itself. For
wider coverage, tick **Also use the ClearURLs list** in the settings (or set
`clearurls = true` under `[rules]` in `config.toml`). bopen then also uses
the rule list of the [ClearURLs](https://github.com/ClearURLs/Rules) project,
about 200 sites' tracking parameters and redirect wrappers. Its suggestions
are labelled **ClearURLs list** and carry a generic reason naming the site;
the built-in rules and your own take precedence where they overlap.

Network use, only while this option is on:

- the list is downloaded from `rules2.clearurls.xyz` (or the mirror
  `rules1.clearurls.xyz`) over HTTPS and checked against its published
  SHA-256 hash before it replaces the cached copy;
- it is refreshed at most once a day, in the background while the
  inspector is open, so a link never waits for it; links opened without the
  inspector never trigger a download;
- **Update now** in the settings downloads it immediately.

The list is cached in `~/.cache/bopen/` (`%LocalAppData%\bopen\` on
Windows). It is not bundled with bopen: the ClearURLs rule data is licensed
under the LGPL-3.0 and is downloaded from its project.

With this option and short-link expansion off (the defaults), bopen makes
no network requests.

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
