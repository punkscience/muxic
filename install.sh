#!/usr/bin/env bash
# muxic one-liner install script.
# Detects the current platform and installs muxic:
#   - Linux:   adds the signed APT repo and runs `apt install muxic`
#   - macOS:   downloads the release archive into ~/.local/bin
#   - Windows: downloads the release archive into ~/.local/bin (or use choco)
#
# Usage:
#   curl -fsSL https://punkscience.github.io/muxic/install.sh | bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
NC='\033[0m' # No Color

REPO="punkscience/muxic"

# ---------------------------------------------------------------------------
# Platform detection
# ---------------------------------------------------------------------------
detect_platform() {
    local os
    os="$(uname -s)"
    case "$os" in
        Linux)  echo "linux" ;;
        Darwin) echo "macos" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)
            echo "unsupported: $os" >&2
            exit 1
            ;;
    esac
}

# Resolve the newest release tag (e.g. v1.2.3) from the GitHub API.
latest_tag() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep -m1 '"tag_name"' \
        | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# ---------------------------------------------------------------------------
# Linux: APT install (Debian/Ubuntu)
# ---------------------------------------------------------------------------
install_linux() {
    # Require root for package operations. Re-exec via sudo if not already root
    # and not running under sudo.
    if [ "$(id -u)" -ne 0 ] && [ -z "${SUDO_USER:-}" ]; then
        echo -e "${BOLD}📦 muxic installer — Linux (APT)${NC}"
        echo ""
        echo "  This will:"
        echo "    1. Install the muxic APT repository signing key"
        echo "    2. Add the muxic APT source"
        echo "    3. Run: apt install muxic"
        echo ""
        echo "  Root privileges are required for steps 1-3."
        echo ""
        # When piped via curl|bash or bash <(curl), $0 is /dev/fd/N which
        # sudo cannot read. Re-fetch from the canonical URL in that case.
        if [ ! -f "$0" ] || [ "${0#/dev/fd/}" != "$0" ]; then
            SCRIPT_URL="https://punkscience.github.io/muxic/install.sh"
            TMP_SCRIPT="$(mktemp /tmp/muxic-install.XXXXXX)"
            curl -fsSL "$SCRIPT_URL" -o "$TMP_SCRIPT"
            exec sudo bash "$TMP_SCRIPT"
        fi
        exec sudo bash "$0"
    fi

    echo -e "${BOLD}📦 muxic installer — Linux (APT)${NC}"

    # 1. Install GPG signing key (binary format, ready for apt)
    echo "  → Installing muxic APT signing key…"
    curl -fsSL https://punkscience.github.io/muxic/apt/muxic-archive-keyring.gpg \
        -o /usr/share/keyrings/muxic-archive-keyring.gpg

    # 2. Add apt source
    echo "  → Adding muxic APT source…"
    cat > /etc/apt/sources.list.d/muxic.list <<'SOURCELIST'
deb [arch=amd64,arm64 signed-by=/usr/share/keyrings/muxic-archive-keyring.gpg] https://punkscience.github.io/muxic/apt/ stable main
SOURCELIST

    # 3. Update and install
    echo "  → Updating package lists…"
    apt-get update -qq

    echo "  → Installing muxic…"
    apt-get install -y muxic

    echo ""
    echo -e "${GREEN}✓ muxic installed. Run: muxic --help${NC}"
}

# ---------------------------------------------------------------------------
# macOS: direct download into ~/.local/bin
# ---------------------------------------------------------------------------
install_macos() {
    echo -e "${BOLD}📦 muxic installer — macOS${NC}"
    echo ""

    local arch="amd64"
    if [ "$(uname -m)" = "arm64" ]; then
        arch="arm64"
    fi

    local tag version tarball url install_dir
    tag="$(latest_tag)"
    version="${tag#v}"
    tarball="muxic_${version}_darwin_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${tag}/${tarball}"
    install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"

    echo "  → Downloading muxic ${tag} (${arch})…"
    curl -fsSL "$url" -o "/tmp/${tarball}"

    echo "  → Extracting to $install_dir …"
    tar -xzf "/tmp/${tarball}" -C "$install_dir" muxic
    rm -f "/tmp/${tarball}"

    echo ""
    echo -e "${GREEN}✓ muxic installed to $install_dir/muxic${NC}"
    echo "  Make sure $install_dir is on your PATH."
}

# ---------------------------------------------------------------------------
# Windows: direct download or Chocolatey
# ---------------------------------------------------------------------------
install_windows() {
    echo -e "${BOLD}📦 muxic installer — Windows${NC}"
    echo ""
    echo "  Tip: on Windows you can also run:  choco install muxic"
    echo ""

    local arch="amd64"
    if [ "$(uname -m)" = "aarch64" ] || [ "${PROCESSOR_ARCHITECTURE:-}" = "ARM64" ]; then
        arch="arm64"
    fi

    local tag version zip url install_dir
    tag="$(latest_tag)"
    version="${tag#v}"
    zip="muxic_${version}_windows_${arch}.zip"
    url="https://github.com/${REPO}/releases/download/${tag}/${zip}"
    install_dir="$HOME/.local/bin"
    mkdir -p "$install_dir"

    echo "  → Downloading muxic ${tag} (${arch})…"
    curl -fsSL "$url" -o "${TEMP:-/tmp}/$zip"

    echo "  → Extracting to $install_dir …"
    unzip -o "${TEMP:-/tmp}/$zip" -d "$install_dir"
    rm -f "${TEMP:-/tmp}/$zip"

    echo ""
    echo -e "${GREEN}✓ muxic installed to $install_dir/muxic.exe${NC}"
    echo "  Make sure $install_dir is on your PATH."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
PLATFORM="$(detect_platform)"

case "$PLATFORM" in
    linux)   install_linux ;;
    macos)   install_macos ;;
    windows) install_windows ;;
esac
