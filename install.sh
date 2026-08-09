cat << 'EOF' > install.sh
#!/bin/sh
set -e

REPO="toshalkumbhar8979-design/Agent-grep"
BINARY="agtgrep"

echo "Installing ${BINARY}..."

LATEST_TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -n "$LATEST_TAG" ]; then
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    if [ "$ARCH" = "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        ARCH="arm64"
    fi

    VERSION="${LATEST_TAG#v}"
    TARBALL="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${TARBALL}"

    echo "Downloading pre-built binary ${LATEST_TAG} for ${OS}/${ARCH}..."
    if curl -fsSL "$URL" -o "$TARBALL"; then
        tar -xzf "$TARBALL" "$BINARY"
        INSTALL_DIR="/usr/local/bin"
        if [ ! -w "$INSTALL_DIR" ]; then
            sudo mv "$BINARY" "$INSTALL_DIR/"
        else
            mv "$BINARY" "$INSTALL_DIR/"
        fi
        rm -f "$TARBALL"
        echo "Successfully installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
        exit 0
    fi
fi

echo "No pre-compiled binary release found. Falling back to 'go install'..."
go install "github.com/${REPO}@latest"
echo "Successfully installed ${BINARY} via go install!"
EOF