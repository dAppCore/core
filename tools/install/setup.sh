#!/bin/bash
# Core CLI installer for macOS and Linux
# Usage: curl -fsSL https://core.io.in/setup.sh | bash
#        curl -fsSL https://core.io.in/setup.sh | bash -s -- v1.0.0
set -eo pipefail

VERSION="${1:-latest}"
REPO="host-uk/core"
BINARY="core"
INSTALL_DIR="${CORE_INSTALL_DIR:-/usr/local/bin}"

# Colours
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
DIM='\033[2m'
NC='\033[0m'

info() { echo -e "${BLUE}>>>${NC} $1"; }
success() { echo -e "${GREEN}>>>${NC} $1"; }
error() { echo -e "${RED}>>>${NC} $1" >&2; exit 1; }

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) error "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) error "Unsupported OS: $OS (use setup.bat for Windows)" ;;
esac

# Resolve latest version
if [ "$VERSION" = "latest" ]; then
  info "Fetching latest version..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    error "Failed to fetch latest version"
  fi
fi

info "Installing ${BINARY} ${VERSION} for ${OS}/${ARCH}..."

# Download archive
ARCHIVE="${BINARY}-${OS}-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

info "Downloading ${ARCHIVE}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${ARCHIVE}"; then
  error "Failed to download ${DOWNLOAD_URL}"
fi

# Extract
info "Extracting..."
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"
chmod +x "${TMP_DIR}/${BINARY}"

# Install
info "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

# Verify
if command -v "$BINARY" &>/dev/null; then
  success "Installed successfully!"
  echo -e "${DIM}$($BINARY --version)${NC}"
else
  success "Installed to ${INSTALL_DIR}/${BINARY}"
  echo -e "${DIM}Add ${INSTALL_DIR} to your PATH if not already present${NC}"
fi
