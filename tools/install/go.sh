#!/bin/bash
# Core CLI installer - Go development variant
# Usage: curl -fsSL https://core.io.in/go.sh | bash
set -eo pipefail

VERSION="${VERSION:-${1:-latest}}"
REPO="host-uk/core"
BINARY="core"
VARIANT="go"

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

echo "Installing ${BINARY} ${VERSION} (${VARIANT} variant) for ${OS}/${ARCH}..."

# Download variant-specific archive if available, else fall back to full
ARCHIVE="${BINARY}-${VARIANT}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

if ! curl -fsSLI "$URL" 2>/dev/null | grep -qE "^HTTP/.* (200|302)"; then
  # Fall back to full variant
  ARCHIVE="${BINARY}-${OS}-${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
  echo "Using full variant (${VARIANT} variant not available for ${VERSION})"
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if ! curl -fsSL "$URL" -o "$TMPDIR/$ARCHIVE"; then
  echo "Failed to download ${ARCHIVE}" >&2
  exit 1
fi
tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"

[ -w /usr/local/bin ] && mv "$TMPDIR/${BINARY}" /usr/local/bin/ || sudo mv "$TMPDIR/${BINARY}" /usr/local/bin/

${BINARY} --version
