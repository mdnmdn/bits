#!/bin/sh
set -e

# CoinGecko CLI installer
# Usage: curl -sSfL https://raw.githubusercontent.com/coingecko/coingecko-cli/main/install.sh | sh

REPO="coingecko/coingecko-cli"
BINARY="cg"
INSTALL_DIR="/usr/local/bin"

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "unsupported"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported"; exit 1 ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

echo "Detected: ${OS}/${ARCH}"

# Get latest release tag
LATEST=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Error: could not determine latest release"
  exit 1
fi

echo "Latest release: ${LATEST}"

VERSION="${LATEST#v}"
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${URL}..."
curl -sSfL "$URL" -o "${TMPDIR}/${FILENAME}"

echo "Extracting..."
if [ "$EXT" = "tar.gz" ]; then
  tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"
else
  unzip -q "${TMPDIR}/${FILENAME}" -d "$TMPDIR"
fi

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"
