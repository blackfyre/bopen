#!/bin/sh
# Install bopen from GitHub releases.
#
#   curl -fsSL https://raw.githubusercontent.com/blackfyre/bopen/main/install.sh | sh
#
# Environment:
#   BOPEN_VERSION      release tag to install (default: latest), e.g. v0.1.0
#   BOPEN_INSTALL_DIR  target directory (default: ~/.local/bin)
#   BOPEN_BASE_URL     releases URL (default: https://github.com/blackfyre/bopen/releases)
#
# The script never needs root and never changes your default browser; run
# `bopen register` afterwards to do that.

set -eu

say() {
	printf '%s\n' "$*"
}

fail() {
	printf 'bopen install: %s\n' "$*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "'$1' is required but was not found"
}

download() {
	# download URL FILE
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$2" "$1"
	else
		fail "either 'curl' or 'wget' is required"
	fi
}

main() {
	os=$(uname -s)
	[ "$os" = Linux ] || fail "this script installs the Linux build; $os is not supported (see the README for Windows)"
	arch=$(uname -m)
	case $arch in
	x86_64 | amd64) ASSET=bopen_linux_amd64.tar.gz ;;
	aarch64 | arm64) ASSET=bopen_linux_arm64.tar.gz ;;
	*) fail "no binary is published for $arch yet; build from source instead (see the README)" ;;
	esac

	if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
		fail "either 'curl' or 'wget' is required"
	fi
	need tar
	need sha256sum
	need mktemp

	base=${BOPEN_BASE_URL:-https://github.com/blackfyre/bopen/releases}
	if [ -n "${BOPEN_VERSION:-}" ]; then
		url=$base/download/$BOPEN_VERSION
	else
		url=$base/latest/download
	fi
	dir=${BOPEN_INSTALL_DIR:-$HOME/.local/bin}

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT INT TERM

	say "Downloading $url/$ASSET"
	download "$url/$ASSET" "$tmp/$ASSET" || fail "download of $ASSET failed"
	download "$url/checksums.txt" "$tmp/checksums.txt" || fail "download of checksums.txt failed"

	grep "  $ASSET\$" "$tmp/checksums.txt" >"$tmp/expected" || fail "checksums.txt has no entry for $ASSET"
	(cd "$tmp" && sha256sum -c expected >/dev/null 2>&1) || fail "checksum mismatch for $ASSET; nothing was installed"

	mkdir -p "$tmp/extract"
	tar -xzf "$tmp/$ASSET" -C "$tmp/extract" bopen || fail "could not extract $ASSET"

	mkdir -p "$dir"
	# Copy next to the target and rename, so a running bopen is replaced safely.
	cp "$tmp/extract/bopen" "$dir/.bopen.new"
	chmod 0755 "$dir/.bopen.new"
	mv -f "$dir/.bopen.new" "$dir/bopen"

	installed=$("$dir/bopen" 2>&1 | head -n 1 || true)
	case $installed in
	"bopen "*) say "Installed $installed to $dir/bopen" ;;
	*)
		say "Installed bopen to $dir/bopen, but it could not start:"
		say "  $installed"
		say "Install the Wayland, X11, EGL and xkbcommon libraries listed in the README."
		;;
	esac

	case ":${PATH:-}:" in
	*":$dir:"*) cmd=bopen ;;
	*)
		cmd=$dir/bopen
		say ""
		say "Warning: $dir is not on your PATH. Add it, or use the full path below."
		;;
	esac
	say ""
	say "To make bopen your default browser, run:"
	say "  $cmd register"
}

main "$@"
