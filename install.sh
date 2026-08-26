#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPOSITORY="${TASKAIO_REPOSITORY:-smilejk930/taskaio-cli}"
REQUESTED_VERSION="${TASKAIO_VERSION:-latest}"

command -v curl >/dev/null 2>&1 || { echo "Error: curl is required." >&2; exit 1; }
command -v tar >/dev/null 2>&1 || { echo "Error: tar is required." >&2; exit 1; }
command -v sha256sum >/dev/null 2>&1 || { echo "Error: sha256sum is required." >&2; exit 1; }

case "$(uname -s)" in
  Linux) os="linux" ;;
  *) echo "Error: only Linux and WSL are supported." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Error: unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

if [[ "$REQUESTED_VERSION" == "latest" ]]; then
  tag="$(curl -fsSL "https://api.github.com/repos/$REPOSITORY/releases/latest" \
    | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  [[ -n "$tag" ]] || { echo "Error: could not determine the latest release." >&2; exit 1; }
else
  tag="$REQUESTED_VERSION"
  [[ "$tag" == v* ]] || tag="v$tag"
fi

version="${tag#v}"
archive="taskaio-cli_${version}_${os}_${arch}.tar.gz"
base_url="https://github.com/$REPOSITORY/releases/download/$tag"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

echo "Installing taskAIO CLI $tag for $os/$arch..."
curl -fsSL "$base_url/$archive" -o "$tmp_dir/$archive"
curl -fsSL "$base_url/checksums.txt" -o "$tmp_dir/checksums.txt"
(
  cd "$tmp_dir"
  grep "  $archive\$" checksums.txt | sha256sum -c -
  tar -xzf "$archive" taskaio
)

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp_dir/taskaio" "$INSTALL_DIR/taskaio"
echo "Installed to $INSTALL_DIR/taskaio"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "Add $INSTALL_DIR to PATH to run taskaio from any directory."
fi
