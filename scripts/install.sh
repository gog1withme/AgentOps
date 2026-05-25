#!/usr/bin/env bash
set -euo pipefail

REPO="gog1withme/AgentOps"
VERSION="${AGENTOPS_VERSION:-1.0.0}"
INSTALL_DIR="${AGENTOPS_INSTALL_DIR:-$HOME/.agentops}"
SKIP_INIT="${AGENTOPS_SKIP_INIT:-0}"

detect_platform() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    linux) os="linux" ;;
    darwin) os="darwin" ;;
    *)
      echo "Unsupported OS: $os" >&2
      exit 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
      echo "Unsupported architecture: $arch" >&2
      exit 1
      ;;
  esac

  echo "${os}_${arch}"
}

download_release() {
  local platform="$1"
  local os="${platform%%_*}"
  local arch="${platform##*_}"
  local archive="agentops_${VERSION}_${os}_${arch}.tar.gz"
  local url="https://github.com/${REPO}/releases/download/v${VERSION}/${archive}"
  local tmpdir checksum_line expected actual

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  echo "Downloading ${url}..."
  curl -fsSL "$url" -o "${tmpdir}/${archive}"

  echo "Verifying checksum..."
  curl -fsSL "https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt" -o "${tmpdir}/checksums.txt"
  checksum_line="$(grep " ${archive}$" "${tmpdir}/checksums.txt" || true)"
  if [[ -z "$checksum_line" ]]; then
    echo "Checksum entry not found for ${archive}" >&2
    exit 1
  fi
  expected="${checksum_line%% *}"
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${tmpdir}/${archive}" | awk '{print $1}')"
  else
    actual="$(shasum -a 256 "${tmpdir}/${archive}" | awk '{print $1}')"
  fi
  if [[ "$expected" != "$actual" ]]; then
    echo "Checksum mismatch for ${archive}" >&2
    exit 1
  fi

  tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"
  echo "$tmpdir"
}

install_agentops() {
  local extract_dir="$1"
  local bin_dir="${INSTALL_DIR}/bin"
  local dash_dir="${INSTALL_DIR}/dashboard/out"

  mkdir -p "$bin_dir" "$dash_dir"
  install -m 755 "${extract_dir}/agentops" "${bin_dir}/agentops"
  rm -rf "${INSTALL_DIR}/dashboard/out"
  mkdir -p "${INSTALL_DIR}/dashboard"
  cp -R "${extract_dir}/dashboard/out" "${INSTALL_DIR}/dashboard/"

  echo ""
  echo "AgentOps ${VERSION} installed to ${INSTALL_DIR}"
  echo ""
  echo "Add to your PATH:"
  echo "  export PATH=\"${bin_dir}:\$PATH\""
  echo ""
  echo "Then run:"
  echo "  agentops init"
  echo "  agentops env    # run the printed command"
  echo "  agentops dev"
  echo ""

  if [[ "$SKIP_INIT" != "1" ]]; then
    if [[ ":$PATH:" != *":${bin_dir}:"* ]]; then
      export PATH="${bin_dir}:$PATH"
    fi
    "${bin_dir}/agentops" init || true
  fi
}

main() {
  local platform extract_dir
  platform="$(detect_platform)"
  extract_dir="$(download_release "$platform")"
  install_agentops "$extract_dir"
}

main "$@"
