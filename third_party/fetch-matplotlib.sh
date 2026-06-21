#!/usr/bin/env bash
# Provision the matplotlib source tree used as the parity source-of-truth.
#
# Many parity tests read matplotlib's own sources and bundled fonts directly,
# e.g. third_party/matplotlib/lib/matplotlib/colorbar.py and
# third_party/matplotlib/lib/matplotlib/mpl-data/fonts/ttf/DejaVuSans.ttf. The
# directory is gitignored (it is large vendored source), so CI must fetch it.
#
# This downloads the pinned matplotlib 3.10.9 sdist from PyPI, verifies its
# sha256, and extracts it to third_party/matplotlib. It is idempotent: if the
# tree already exists at the pinned version it does nothing.
set -euo pipefail

MPL_VERSION="3.10.9"
MPL_SDIST_URL="https://files.pythonhosted.org/packages/63/1b/4be5be87d43d327a0cf4de1a56e86f7f84c89312452406cf122efe2839e6/matplotlib-${MPL_VERSION}.tar.gz"
MPL_SDIST_SHA256="fd66508e8c6877d98e586654b608a0456db8d7e8a546eb1e2600efd957302358"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dest="${script_dir}/matplotlib"

if [ -f "${dest}/PKG-INFO" ] && grep -qx "Version: ${MPL_VERSION}" "${dest}/PKG-INFO"; then
  echo "third_party/matplotlib already at ${MPL_VERSION}; skipping."
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

tarball="${tmp}/matplotlib-${MPL_VERSION}.tar.gz"
echo "Downloading matplotlib ${MPL_VERSION} sdist..."
curl -fsSL "${MPL_SDIST_URL}" -o "${tarball}"

echo "Verifying sha256..."
echo "${MPL_SDIST_SHA256}  ${tarball}" | sha256sum --check --status

echo "Extracting..."
tar -xzf "${tarball}" -C "${tmp}"

rm -rf "${dest}"
mv "${tmp}/matplotlib-${MPL_VERSION}" "${dest}"
echo "Provisioned third_party/matplotlib at ${MPL_VERSION}."
