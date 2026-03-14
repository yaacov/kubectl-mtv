#!/usr/bin/env bash
set -euo pipefail

REPO="yaacov/kubectl-mtv"
BINARY_NAME="kubectl-mtv"
ARCHIVE_PREFIX="kubectl-mtv"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

info()  { printf '\033[1;32m%s\033[0m\n' "$*"; }
warn()  { printf '\033[1;33m%s\033[0m\n' "$*" >&2; }
error() { printf '\033[1;31m%s\033[0m\n' "$*" >&2; exit 1; }

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       error "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             error "Unsupported architecture: $(uname -m). Only amd64 and arm64 are supported." ;;
  esac
}

latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | head -1 \
    | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/'
}

verify_checksum() {
  local file="$1" expected_file="$2"
  local expected actual
  expected="$(awk '{print $1}' "$expected_file")"
  if command -v sha256sum &>/dev/null; then
    actual="$(sha256sum "$file" | awk '{print $1}')"
  elif command -v shasum &>/dev/null; then
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
  else
    error "Neither sha256sum nor shasum found — cannot verify checksum."
  fi

  if [ "$expected" != "$actual" ]; then
    error "Checksum mismatch!\n  Expected: ${expected}\n  Got:      ${actual}"
  fi
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
VERSION="${VERSION:-$(latest_version)}"

[ -z "$VERSION" ] && error "Could not determine latest version. Set VERSION explicitly."

info "Installing ${ARCHIVE_PREFIX} ${VERSION} (${OS}/${ARCH})"

ARCHIVE="${ARCHIVE_PREFIX}-${VERSION}-${OS}-${ARCH}.tar.gz"
ARCHIVE_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
CHECKSUM_URL="${ARCHIVE_URL}.sha256sum"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading ${ARCHIVE}..."
curl -fSL -o "${TMPDIR}/${ARCHIVE}" "$ARCHIVE_URL"

info "Downloading checksum..."
curl -fSL -o "${TMPDIR}/${ARCHIVE}.sha256sum" "$CHECKSUM_URL"

info "Verifying checksum..."
verify_checksum "${TMPDIR}/${ARCHIVE}" "${TMPDIR}/${ARCHIVE}.sha256sum"

info "Extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

EXTRACTED_BINARY="${TMPDIR}/${ARCHIVE_PREFIX}-${OS}-${ARCH}"
[ -f "$EXTRACTED_BINARY" ] || error "Expected binary not found in archive: ${ARCHIVE_PREFIX}-${OS}-${ARCH}"

mkdir -p "$INSTALL_DIR"

install -m 0755 "$EXTRACTED_BINARY" "${INSTALL_DIR}/${BINARY_NAME}"

cat > "${INSTALL_DIR}/kubectl_complete-mtv" << 'SCRIPT'
#!/usr/bin/env bash
kubectl-mtv __complete "$@"
SCRIPT
chmod +x "${INSTALL_DIR}/kubectl_complete-mtv"

ln -sf "${INSTALL_DIR}/kubectl_complete-mtv" "${INSTALL_DIR}/oc_complete-mtv"

echo ""
info "Installed files:"
echo "  ${INSTALL_DIR}/${BINARY_NAME}"
echo "  ${INSTALL_DIR}/kubectl_complete-mtv"
echo "  ${INSTALL_DIR}/oc_complete-mtv"
echo ""

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    warn "${INSTALL_DIR} is not in your PATH."
    echo ""
    echo "Add it by running one of:"
    echo ""
    echo "  # bash"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
    echo ""
    echo "  # zsh"
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    echo ""
    ;;
esac

info "Done! Verify with:"
echo "  kubectl mtv --help"
