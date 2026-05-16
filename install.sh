#!/bin/bash
# install.sh - one-line installer
set -eu

REPO="nodora-org/nodora"
BINARY="nodora"
INSTALL_DIR="/usr/local/bin"

info()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
error() { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    error "need '$1' (command not found)"
  fi
}

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"  ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) error "Windows is not supported by this installer — download the latest release from https://github.com/${REPO}/releases" ;;
    *) error "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)       echo "amd64" ;;
    aarch64|arm64)      echo "arm64" ;;
    *) error "unsupported architecture: $(uname -m)" ;;
  esac
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd uname

  OS="$(detect_os)"
  ARCH="$(detect_arch)"

  info "Detected platform: ${OS}/${ARCH}"

  # fetch latest release
  info "Fetching latest release…"
  LATEST="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -n 1 | sed 's/.*"tag_name": *"//;s/".*//')"

  if [ -z "$LATEST" ]; then
    error "could not determine latest release"
  fi

  info "Latest version: ${LATEST}"

  FILENAME="${BINARY}-${LATEST}-${OS}-${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

  # download to a temp directory
  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  info "Downloading ${FILENAME}…"
  curl -fsSL -o "${TMPDIR}/${FILENAME}" "$DOWNLOAD_URL" \
    || error "download failed — check that a release exists for ${OS}/${ARCH}"

  # extract
  info "Extracting…"
  tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

  # find the binary
  BINARY_PATH="$(find "$TMPDIR" -name "$BINARY" -type f | head -n 1)"
  if [ -z "$BINARY_PATH" ]; then
    error "could not find '${BINARY}' binary in archive"
  fi
  chmod +x "$BINARY_PATH"

  # install
  if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
  else
    info "sudo required to install to ${INSTALL_DIR}"
    sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
  fi

  info "Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"
  info "Run '${BINARY} --help' to get started."
}

main "$@"