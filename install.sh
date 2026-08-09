#!/bin/sh
set -e

REPO="toshalkumbhar8979/agent-grep"
BINARY="agtgrep"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

LATEST_TAG=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Failed to fetch latest release tag."
    exit 1
fi

TARBALL="${BINARY}_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TARBALL}"

echo "Downloading ${BINARY} ${LATEST_TAG} for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o "$TARBALL"

tar -xzf "$TARBALL" "$BINARY"
sudo mv "$BINARY" /usr/local/bin/
rm -f "$TARBALL"

echo "Successfully installed ${BINARY} to /usr/local/bin/${BINARY}"
