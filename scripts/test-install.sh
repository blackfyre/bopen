#!/bin/sh
# Tests install.sh against a local HTTP server serving fake release assets.
# Usage: scripts/test-install.sh
set -eu

repo=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
server_pid=
cleanup() {
	if [ -n "$server_pid" ]; then
		kill "$server_pid" 2>/dev/null || true
	fi
	rm -rf "$work"
}
trap cleanup EXIT INT TERM

failures=0
pass() { printf 'ok    %s\n' "$1"; }
fail() {
	printf 'FAIL  %s\n' "$1"
	failures=$((failures + 1))
}

# make_release DIR VERSION: fake bopen binaries for amd64 and arm64 in a
# release layout; each reports its version and architecture.
make_release() {
	for arch in amd64 arm64; do
		mkdir -p "$1/pkg"
		cat >"$1/pkg/bopen" <<SCRIPT
#!/bin/sh
if [ "\${1:-}" = register ]; then touch "$work/registered"; fi
echo "bopen $2 $arch" >&2
exit 2
SCRIPT
		chmod +x "$1/pkg/bopen"
		tar -czf "$1/bopen_linux_$arch.tar.gz" -C "$1/pkg" bopen
		rm -r "$1/pkg"
	done
	(cd "$1" && sha256sum bopen_linux_amd64.tar.gz bopen_linux_arm64.tar.gz >checksums.txt)
}

srv=$work/srv
make_release "$srv/latest/download" 0.2.0
make_release "$srv/download/v0.1.0" 0.1.0
make_release "$srv/tampered/latest/download" 9.9.9
printf 'tampered' >>"$srv/tampered/latest/download/bopen_linux_amd64.tar.gz"

port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')
python3 -m http.server --bind 127.0.0.1 --directory "$srv" "$port" >"$work/server.log" 2>&1 &
server_pid=$!
base=http://127.0.0.1:$port
for _ in 1 2 3 4 5 6 7 8 9 10; do
	curl -fs "$base/latest/download/checksums.txt" >/dev/null 2>&1 && break
	sleep 0.2
done

# run NAME [VAR=VALUE...]: run install.sh with a fresh install dir unless given.
run() {
	name=$1
	shift
	out=$work/$name.out
	set +e
	env BOPEN_BASE_URL="$base" "$@" sh "$repo/install.sh" >"$out" 2>&1
	status=$?
	set -e
}

dir=$work/fresh/bin
run fresh BOPEN_INSTALL_DIR="$dir" PATH="$dir:$PATH"
if [ "$status" -eq 0 ] && [ -x "$dir/bopen" ] && grep -q "Installed bopen 0.2.0 amd64" "$out" &&
	grep -q "  bopen register" "$out" && ! grep -q "not on your PATH" "$out"; then
	pass "fresh install of latest"
else
	fail "fresh install of latest (status $status): $(cat "$out")"
fi

dir=$work/pinned/bin
run pinned BOPEN_INSTALL_DIR="$dir" BOPEN_VERSION=v0.1.0
if [ "$status" -eq 0 ] && grep -q "Installed bopen 0.1.0 amd64" "$out"; then
	pass "pinned version"
else
	fail "pinned version (status $status): $(cat "$out")"
fi

run upgrade BOPEN_INSTALL_DIR="$dir"
if [ "$status" -eq 0 ] && "$dir/bopen" 2>&1 | grep -q "bopen 0.2.0"; then
	pass "re-run upgrades"
else
	fail "re-run upgrades (status $status): $(cat "$out")"
fi

if grep -q "not on your PATH" "$work/pinned.out" && grep -q "  $dir/bopen register" "$work/pinned.out"; then
	pass "PATH warning with full path"
else
	fail "PATH warning with full path: $(cat "$work/pinned.out")"
fi

dir=$work/tampered/bin
run tampered BOPEN_INSTALL_DIR="$dir" BOPEN_BASE_URL="$base/tampered"
if [ "$status" -ne 0 ] && grep -q "checksum mismatch" "$out" && [ ! -e "$dir" ]; then
	pass "tampered archive rejected"
else
	fail "tampered archive rejected (status $status): $(cat "$out")"
fi

stubs=$work/stubs
mkdir -p "$stubs"
# fake_arch ARCH: a uname stub reporting Linux on ARCH.
fake_arch() {
	cat >"$stubs/uname" <<SCRIPT
#!/bin/sh
case \$1 in -s) echo Linux ;; -m) echo $1 ;; esac
SCRIPT
	chmod +x "$stubs/uname"
}

fake_arch aarch64
dir=$work/arm/bin
run arm BOPEN_INSTALL_DIR="$dir" PATH="$stubs:$PATH"
if [ "$status" -eq 0 ] && grep -q "Installed bopen 0.2.0 arm64" "$out"; then
	pass "arm64 install"
else
	fail "arm64 install (status $status): $(cat "$out")"
fi

fake_arch riscv64
: >"$work/server.log"
dir=$work/riscv/bin
run riscv BOPEN_INSTALL_DIR="$dir" PATH="$stubs:$PATH"
if [ "$status" -ne 0 ] && grep -q "no binary is published for riscv64" "$out" &&
	! grep -q "bopen_linux_" "$work/server.log" && [ ! -e "$dir" ]; then
	pass "unsupported architecture"
else
	fail "unsupported architecture (status $status): $(cat "$out")"
fi

# A PATH with everything the script needs except sha256sum.
limited=$work/limited
mkdir -p "$limited"
for tool in sh uname curl tar mktemp grep cp chmod mv mkdir rm head cat dirname; do
	ln -s "$(command -v "$tool")" "$limited/$tool"
done
dir=$work/nosum/bin
run nosum BOPEN_INSTALL_DIR="$dir" PATH="$limited"
if [ "$status" -ne 0 ] && grep -q "'sha256sum' is required" "$out" && [ ! -e "$dir" ]; then
	pass "missing sha256sum"
else
	fail "missing sha256sum (status $status): $(cat "$out")"
fi

if [ ! -e "$work/registered" ]; then
	pass "never registers bopen"
else
	fail "never registers bopen"
fi

if [ "$failures" -ne 0 ]; then
	printf '%d test(s) failed\n' "$failures"
	exit 1
fi
echo "all install.sh tests passed"
