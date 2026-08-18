#!/bin/sh
set -e

REPO="rockpanel/rockpanel"
BINARY="rockpanel"
INSTALL_DIR="/usr/local/bin"

detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)  OS="linux" ;;
        darwin*) OS="darwin" ;;
        *)       echo "Unsupported OS: $OS"; exit 1 ;;
    esac
}

detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        armv7l|armhf)   ARCH="armv7" ;;
        *)              echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
}

detect_arch
detect_os

VERSION="${1:-latest}"
if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi

echo "Installing RockPanel $VERSION for $OS $ARCH"

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY}-${OS}-${ARCH}"
CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

echo "Downloading from $DOWNLOAD_URL"
curl -fsSL -o "$TEMP_DIR/$BINARY" "$DOWNLOAD_URL"
chmod +x "$TEMP_DIR/$BINARY"

if curl -fsSL -o "$TEMP_DIR/checksum.txt" "$CHECKSUM_URL" 2>/dev/null; then
    echo "Verifying checksum..."
    cd "$TEMP_DIR"
    if command -v sha256sum >/dev/null 2>&1; then
        echo "$(cat checksum.txt)  $BINARY" | sha256sum -c -
    elif command -v shasum >/dev/null 2>&1; then
        echo "$(cat checksum.txt)  $BINARY" | shasum -a 256 -c -
    fi
    cd -
fi

echo "Installing to $INSTALL_DIR/$BINARY"
if [ -w "$INSTALL_DIR" ]; then
    cp "$TEMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
    sudo cp "$TEMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "RockPanel installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Run: rockpanel init"
echo "  2. Run: rockpanel start"
echo "  3. Open: http://localhost:8080"