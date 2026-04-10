#!/bin/sh
set -e

REPO="shinagawa-web/gomarklint"
BINARY="gomarklint"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Darwin) OS="Darwin" ;;
  Linux)  OS="Linux" ;;
  *)
    echo "Error: Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Resolve version: environment variable, GitHub API, or redirect fallback
fetch_version_from_api() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
    | grep '"tag_name"' \
    | sed -E 's/.*"v?([^"]+)".*/\1/' \
    | head -n 1
}

fetch_version_from_redirect() {
  curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null \
    | sed -E 's#.*/tag/v?([^/]+)$#\1#'
}

if [ -n "${GOMARKLINT_VERSION:-}" ]; then
  VERSION="$(printf '%s' "$GOMARKLINT_VERSION" | sed -E 's/^v//')"
else
  VERSION="$(fetch_version_from_api)"
  if [ -z "$VERSION" ]; then
    VERSION="$(fetch_version_from_redirect)"
  fi
fi

if [ -z "$VERSION" ]; then
  echo "Error: Failed to determine version. Set GOMARKLINT_VERSION to install a specific release." >&2
  exit 1
fi

# Determine install directory
if [ "$(id -u)" -eq 0 ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
fi

mkdir -p "$INSTALL_DIR"

# Download archive and checksums
ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
CHECKSUMS="${BINARY}_${VERSION}_checksums.txt"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

echo "Installing ${BINARY} v${VERSION} (${OS}/${ARCH})..."
echo "  From: ${BASE_URL}/${ARCHIVE}"
echo "  To:   ${INSTALL_DIR}/${BINARY}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${TMPDIR}/${ARCHIVE}"
curl -fsSL "${BASE_URL}/${CHECKSUMS}" -o "${TMPDIR}/${CHECKSUMS}"

# Verify SHA-256 checksum
EXPECTED="$(grep "${ARCHIVE}" "${TMPDIR}/${CHECKSUMS}" | awk '{print $1}')"
if [ -z "$EXPECTED" ]; then
  echo "Error: Checksum not found for ${ARCHIVE}" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
else
  echo "Warning: No sha256sum or shasum found, skipping checksum verification." >&2
  ACTUAL="$EXPECTED"
fi

if [ "$ACTUAL" != "$EXPECTED" ]; then
  echo "Error: Checksum mismatch!" >&2
  echo "  Expected: ${EXPECTED}" >&2
  echo "  Actual:   ${ACTUAL}" >&2
  exit 1
fi

echo "Checksum verified."

# Extract and install
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"
install -m 755 "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Successfully installed ${BINARY} v${VERSION} to ${INSTALL_DIR}/${BINARY}"

# Check if install dir is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Note: ${INSTALL_DIR} is not in your PATH."
    echo "Add it by running:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
