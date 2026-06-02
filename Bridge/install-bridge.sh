#!/usr/bin/env bash
# install-bridge.sh — install thermalik-bridge on macOS and Linux
set -euo pipefail

REPO="saefy/TicketJS"
BINARY="thermalik-bridge"
INSTALL_DIR="/usr/local/bin"

# ── helpers ──────────────────────────────────────────────────────────────────

log()  { echo "  $*"; }
ok()   { echo "✓ $*"; }
err()  { echo "✗ $*" >&2; exit 1; }

need() {
  command -v "$1" &>/dev/null || err "Required tool not found: $1"
}

# ── detect platform ───────────────────────────────────────────────────────────

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) err "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) err "Unsupported OS: $OS. Use install-bridge.ps1 on Windows." ;;
esac

ASSET="${BINARY}_${OS}_${ARCH}.tar.gz"

# ── resolve install dir ───────────────────────────────────────────────────────

if [[ ! -w "$INSTALL_DIR" ]]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"

  # Ensure ~/.local/bin is in PATH for this session.
  export PATH="$INSTALL_DIR:$PATH"
fi

# ── fetch latest release URL ──────────────────────────────────────────────────

need curl
need tar

echo ""
echo "✦ Thermalik Bridge Installer"
echo "────────────────────────────────────────────────"
log "Platform: ${OS}/${ARCH}"
log "Install dir: ${INSTALL_DIR}"
echo ""

API_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_JSON="$(curl -fsSL "$API_URL")"

DOWNLOAD_URL="$(echo "$RELEASE_JSON" \
  | grep "browser_download_url" \
  | grep "$ASSET" \
  | head -1 \
  | cut -d '"' -f 4)"

[[ -n "$DOWNLOAD_URL" ]] || err "Could not find release asset: $ASSET"

log "Downloading $ASSET..."

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR/$ASSET"

# ── verify checksum ───────────────────────────────────────────────────────────

CHECKSUM_URL="$(echo "$RELEASE_JSON" \
  | grep "browser_download_url" \
  | grep "checksums.txt" \
  | head -1 \
  | cut -d '"' -f 4)"

if [[ -n "$CHECKSUM_URL" ]]; then
  curl -fsSL "$CHECKSUM_URL" -o "$TMPDIR/checksums.txt"
  cd "$TMPDIR"
  grep "$ASSET" checksums.txt | sha256sum --check --status \
    || err "Checksum verification failed. Download may be corrupted."
  ok "Checksum verified"
  cd - > /dev/null
fi

# ── install ───────────────────────────────────────────────────────────────────

tar -xzf "$TMPDIR/$ASSET" -C "$TMPDIR"
install -m 755 "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"

ok "Installed to $INSTALL_DIR/$BINARY"

VERSION="$("$INSTALL_DIR/$BINARY" version 2>/dev/null || echo "unknown")"
log "Version: $VERSION"

echo ""
echo "Run the setup wizard to configure your printers and install as a service:"
echo ""
echo "  thermalik-bridge setup"
echo ""
