#!/bin/sh
# Prepares an Ubuntu amd64 host to cross-compile bopen for linux/arm64:
# enables the arm64 architecture from Ubuntu Ports and installs the cross
# compiler and the arm64 Gio development libraries next to the amd64 ones.
# Used by the release workflow and for local checks in an Ubuntu container.
set -eu

codename=$(sed -n "s/^VERSION_CODENAME=//p" /etc/os-release)
sudo=""
[ "$(id -u)" -eq 0 ] || sudo=sudo

$sudo dpkg --add-architecture arm64

# Existing sources serve amd64 only; arm64 packages come from Ubuntu Ports.
for f in /etc/apt/sources.list /etc/apt/sources.list.d/*.list; do
	[ -f "$f" ] || continue
	$sudo sed -i -E 's/^deb ([^[])/deb [arch=amd64] \1/' "$f"
done
for f in /etc/apt/sources.list.d/*.sources; do
	[ -f "$f" ] || continue
	grep -q '^Architectures:' "$f" || $sudo sed -i -E 's/^(Types:.*)$/\1\nArchitectures: amd64/' "$f"
done
printf '%s\n' \
	"deb [arch=arm64] http://ports.ubuntu.com/ubuntu-ports $codename main universe" \
	"deb [arch=arm64] http://ports.ubuntu.com/ubuntu-ports $codename-updates main universe" \
	"deb [arch=arm64] http://ports.ubuntu.com/ubuntu-ports $codename-security main universe" |
	$sudo tee /etc/apt/sources.list.d/arm64-ports.list >/dev/null

$sudo apt-get update
$sudo apt-get install -y --no-install-recommends \
	gcc-aarch64-linux-gnu libc6-dev-arm64-cross \
	libwayland-dev:arm64 libx11-dev:arm64 libx11-xcb-dev:arm64 libxkbcommon-x11-dev:arm64 \
	libgles2-mesa-dev:arm64 libegl1-mesa-dev:arm64 libffi-dev:arm64 libxcursor-dev:arm64 libvulkan-dev:arm64
