#!/usr/bin/env bash
set -e

# asactl installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/emredurukn/Apple-Ads-CLI/main/install.sh | bash

REPO="emredurukn/Apple-Ads-CLI"
BINARY_NAME="asactl"

# Color helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BLUE}${BOLD}==>${NC} Installing ${BOLD}asactl${NC} (Apple Search Ads CLI)..."

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Darwin)
    TARGET_OS="darwin"
    ;;
  Linux)
    TARGET_OS="linux"
    ;;
  *)
    echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
    exit 1
    ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)
    TARGET_ARCH="amd64"
    ;;
  arm64|aarch64)
    TARGET_ARCH="arm64"
    ;;
  *)
    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
    exit 1
    ;;
esac

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
USE_SUDO=false

if [ ! -w "$INSTALL_DIR" ]; then
  if [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
  elif command -v sudo >/dev/null 2>&1 && [ -t 0 ]; then
    USE_SUDO=true
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# Fetch latest release version from GitHub
echo -e "${BLUE}${BOLD}==>${NC} Checking latest version..."
LATEST_TAG=$(curl -sSL -H "Accept: application/vnd.github.v3+json" "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)

if [ -z "$LATEST_TAG" ]; then
  # Fallback: check git tags
  LATEST_TAG="v0.1.0"
fi

VERSION="${LATEST_TAG#v}"
FILENAME="asactl_${VERSION}_${TARGET_OS}_${TARGET_ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${FILENAME}"

# Download and extract in temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo -e "${BLUE}${BOLD}==>${NC} Downloading ${BOLD}${LATEST_TAG}${NC} for ${TARGET_OS}/${TARGET_ARCH}..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$FILENAME"; then
  echo -e "${RED}Error: Failed to download release asset from ${DOWNLOAD_URL}${NC}"
  echo -e "You can build from source with: go install github.com/${REPO}@latest"
  exit 1
fi

tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
  echo -e "${RED}Error: Binary not found in release archive${NC}"
  exit 1
fi

# Move binary to target directory
echo -e "${BLUE}${BOLD}==>${NC} Installing to ${BOLD}${INSTALL_DIR}/${BINARY_NAME}${NC}..."
if [ "$USE_SUDO" = true ]; then
  sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

echo -e "${GREEN}${BOLD}✓ asactl installed successfully!${NC}"
echo ""

# Check if INSTALL_DIR is in PATH
if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  echo -e "${BLUE}Notice:${NC} ${INSTALL_DIR} is not in your PATH."
  echo -e "Add it to your shell configuration (e.g. ~/.zshrc or ~/.bashrc):"
  echo -e "  export PATH=\"\$PATH:${INSTALL_DIR}\""
  echo ""
fi

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
  "$BINARY_NAME" version
else
  "$INSTALL_DIR/$BINARY_NAME" version
fi

echo ""
echo -e "Get started by running:"
echo -e "  ${BOLD}asactl --help${NC}"
echo -e "  ${BOLD}asactl auth login --help${NC}"
