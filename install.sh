#!/bin/sh
set -e

REPO="mdawess/amp"
INSTALL_DIR="${AMP_INSTALL_DIR:-/usr/local/bin}"
BIN="amp"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64)  arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$os" != "darwin" ] && [ "$os" != "linux" ]; then
  echo "unsupported OS: $os" >&2
  exit 1
fi

asset="${BIN}-${os}-${arch}"

echo "Fetching latest release from github.com/${REPO}..."
tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$tag" ]; then
  echo "could not determine latest release tag" >&2
  exit 1
fi

url="https://github.com/${REPO}/releases/download/${tag}/${asset}"
echo "Downloading ${tag}/${asset}..."

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

curl -fsSL "$url" -o "$tmp"
chmod +x "$tmp"

if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to ${INSTALL_DIR} (may require sudo)..."
  sudo mv "$tmp" "${INSTALL_DIR}/${BIN}"
else
  mv "$tmp" "${INSTALL_DIR}/${BIN}"
fi

echo "Installed amp ${tag} to ${INSTALL_DIR}/${BIN}"
