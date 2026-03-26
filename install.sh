#!/bin/sh
# Portable installer: latest onlycli release from GitHub.

set -eu

REPO="onlycli/onlycli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

fetch_stdout() {
	url=$1
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO- "$url"
	else
		echo "install.sh: need curl or wget" >&2
		exit 1
	fi
}

fetch_to_file() {
	url=$1
	dest=$2
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$dest"
	elif command -v wget >/dev/null 2>&1; then
		wget -q "$url" -O "$dest"
	else
		echo "install.sh: need curl or wget" >&2
		exit 1
	fi
}

get_latest_tag() {
	fetch_stdout "https://api.github.com/repos/${REPO}/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1
}

case "$(uname -s)" in
Linux) OS=linux ;;
Darwin) OS=darwin ;;
*)
	echo "install.sh: unsupported OS (need Linux or Darwin)" >&2
	exit 1
	;;
esac

ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
x86_64 | amd64) ARCH=amd64 ;;
arm64 | aarch64) ARCH=arm64 ;;
*)
	echo "install.sh: unsupported architecture: $ARCH_RAW (need amd64 or arm64)" >&2
	exit 1
	;;
esac

TAG=$(get_latest_tag)
if [ -z "$TAG" ]; then
	echo "install.sh: could not determine latest release tag" >&2
	exit 1
fi

ASSET="onlycli_${TAG}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT INT HUP

ARCHIVE="$TMPDIR/$ASSET"
fetch_to_file "$URL" "$ARCHIVE"

(
	cd "$TMPDIR" || exit 1
	tar -xzf "$ASSET"
)

BIN="$TMPDIR/onlycli"
if [ ! -f "$BIN" ]; then
	echo "install.sh: expected binary onlycli in archive, not found" >&2
	exit 1
fi

chmod +x "$BIN"
if ! "$BIN" version >/dev/null 2>&1; then
	echo "install.sh: downloaded binary failed 'onlycli version'" >&2
	exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
	echo "install.sh: INSTALL_DIR does not exist: $INSTALL_DIR" >&2
	exit 1
fi

if [ ! -w "$INSTALL_DIR" ]; then
	echo "install.sh: cannot write to $INSTALL_DIR (try sudo or set INSTALL_DIR to a writable directory)" >&2
	exit 1
fi

mv "$BIN" "$INSTALL_DIR/onlycli"

echo "Installed onlycli ${TAG} to ${INSTALL_DIR}/onlycli"
echo
echo "Usage:"
echo "  onlycli --help          Show commands"
echo "  onlycli version         Show version"
echo "  onlycli generate --help   Generate a CLI from an OpenAPI spec"
