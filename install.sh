#!/usr/bin/env bash
set -euo pipefail

repo="SemihMutlu07/wrapminal"
version="${WRAPMINAL_VERSION:-latest}"
bin_dir="${WRAPMINAL_BIN_DIR:-$HOME/.local/bin}"
run_after_install=1

if [[ "${1:-}" == "--no-run" ]]; then
  run_after_install=0
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"

case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="wrapminal-${os}-${arch}"
if [[ "$version" == "latest" ]]; then
  url="https://github.com/${repo}/releases/latest/download/${asset}"
else
  url="https://github.com/${repo}/releases/download/${version}/${asset}"
fi

mkdir -p "$bin_dir"
target="$bin_dir/wrapminal"

echo "Downloading $asset..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$target"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$target"
else
  echo "curl or wget is required." >&2
  exit 1
fi

chmod +x "$target"
echo "Installed wrapminal to $target (prebuilt binary — no Go or Node required)"

ensure_on_path() {
  case ":$PATH:" in
    *":$bin_dir:"*) return 0 ;;
  esac

  local rc=""
  case "${SHELL:-}" in
    */zsh) rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
    */bash) rc="$HOME/.bashrc" ;;
    *) rc="$HOME/.profile" ;;
  esac

  local marker="# added by wrapminal installer"
  local line="export PATH=\"$bin_dir:\$PATH\"  $marker"

  if [[ -n "$rc" ]] && ! grep -qsF "$marker" "$rc"; then
    printf '\n%s\n' "$line" >> "$rc"
    echo "Added $bin_dir to PATH in $rc"
  fi
  echo "Run this to use 'wrapminal' in the current shell (or restart it):"
  echo "  export PATH=\"$bin_dir:\$PATH\""
}
ensure_on_path

if [[ "$run_after_install" == "1" ]]; then
  "$target"
fi
