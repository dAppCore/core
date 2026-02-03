#!/bin/bash
# Core CLI installer for CI environments
# Minimal, fast, no interactive prompts
# Usage: curl -fsSL https://core.io.in/ci.sh | bash
#        VERSION=v1.0.0 curl -fsSL https://core.io.in/ci.sh | bash
set -eo pipefail

VERSION="${VERSION:-${1:-latest}}"
REPO="host-uk/core"
BINARY="core"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Resolve latest
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    echo "Failed to fetch latest version from GitHub API" >&2
    exit 1
  fi
fi

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."

# Download and extract to secure temp dir
ARCHIVE="${BINARY}-${OS}-${ARCH}.tar.gz"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

if ! curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}" -o "$TMPDIR/$ARCHIVE"; then
  echo "Failed to download ${ARCHIVE}" >&2
  exit 1
fi
tar -xzf "$TMPDIR/$ARCHIVE" -C "$TMPDIR"

# Install (CI runners typically have write access to /usr/local/bin)
if [ -w /usr/local/bin ]; then
  mv "$TMPDIR/${BINARY}" /usr/local/bin/
else
  sudo mv "$TMPDIR/${BINARY}" /usr/local/bin/
fi

${BINARY} --version
