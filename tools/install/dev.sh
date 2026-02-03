#!/bin/bash
# Core CLI installer - Multi-repo development variant (full)
# Usage: curl -fsSL https://core.io.in/dev.sh | bash
set -eo pipefail

VERSION="${VERSION:-${1:-latest}}"
REPO="host-uk/core"
BINARY="core"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version from GitHub API" >&2
    exit 1
  fi
fi

echo "Installing ${BINARY} ${VERSION} (full) for ${OS}/${ARCH}..."

ARCHIVE="${BINARY}-${OS}-${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if ! curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}" -o "$TMPDIR/$ARCHIVE"; then
  echo "Failed to download ${ARCHIVE}" >&2
  exit 1
fi
tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"

[ -w /usr/local/bin ] && mv "$TMPDIR/${BINARY}" /usr/local/bin/ || sudo mv "$TMPDIR/${BINARY}" /usr/local/bin/

${BINARY} --version

echo ""
echo "Full development variant installed. Available commands:"
echo "  core dev     - Multi-repo workflows"
echo "  core build   - Cross-platform builds"
echo "  core release - Build and publish releases"
