#!/usr/bin/env bash
# Build a static FreeType 2.6.1 for matplotlib pixel parity.
#
# matplotlib generates its reference images with FreeType 2.6.1 (its pinned
# version, exposed as matplotlib.ft2font.__freetype_version__). The AGG backend
# links FreeType via `#cgo pkg-config: freetype2`; building 2.6.1 here and
# pointing PKG_CONFIG_PATH at this prefix makes our glyph rasterization match
# matplotlib's reference output. The pinned tarball/sha mirror matplotlib's
# third_party/matplotlib/subprojects/freetype-2.6.1.wrap.
#
# Output (all gitignored): third_party/freetype/prefix/{lib,include}
#   - lib/libfreetype.a
#   - lib/pkgconfig/freetype2.pc
#
# Usage: bash third_party/freetype/build.sh   (idempotent; safe to re-run)
set -euo pipefail

VERSION="2.6.1"
SHA256="0a3c7dfbda6da1e8fce29232e8e96d987ababbbf71ebc8c75659e4132c367014"
URL="https://download.savannah.nongnu.org/releases/freetype/freetype-old/freetype-${VERSION}.tar.gz"
FALLBACK_URL="https://downloads.sourceforge.net/project/freetype/freetype2/${VERSION}/freetype-${VERSION}.tar.gz"

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARBALL="${HERE}/freetype-${VERSION}.tar.gz"
SRC="${HERE}/src/freetype-${VERSION}"
PREFIX="${HERE}/prefix"
PC="${PREFIX}/lib/pkgconfig/freetype2.pc"

log() { printf '[freetype-2.6.1] %s\n' "$*" >&2; }

# 1. Idempotency: skip if the static lib + pkg-config file already exist.
#    The .pc "Version" field carries FreeType's libtool number (18.1.12 == the
#    2.6.1 release), not the release string, so match that.
PC_LIBTOOL_VERSION="18.1.12"
if [ -f "${PREFIX}/lib/libfreetype.a" ] && [ -f "${PC}" ]; then
  if grep -q "Version: ${PC_LIBTOOL_VERSION}" "${PC}" 2>/dev/null; then
    log "already built (${PREFIX}); nothing to do"
    exit 0
  fi
fi

verify_sha() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    echo "${SHA256}  ${file}" | sha256sum --check --status
  elif command -v shasum >/dev/null 2>&1; then
    echo "${SHA256}  ${file}" | shasum -a 256 --check --status
  else
    log "WARNING: no sha256sum/shasum available; skipping checksum verification"
  fi
}

# 2. Download (primary, then fallback) and verify the checksum.
if [ ! -f "${TARBALL}" ] || ! verify_sha "${TARBALL}" 2>/dev/null; then
  log "downloading ${URL}"
  if ! curl -fSL --retry 3 -o "${TARBALL}" "${URL}"; then
    log "primary download failed; trying fallback ${FALLBACK_URL}"
    curl -fSL --retry 3 -o "${TARBALL}" "${FALLBACK_URL}"
  fi
  if ! verify_sha "${TARBALL}"; then
    log "ERROR: sha256 mismatch for ${TARBALL} (expected ${SHA256})"
    rm -f "${TARBALL}"
    exit 1
  fi
fi

# 3. Extract a clean source tree.
rm -rf "${SRC}"
mkdir -p "${HERE}/src"
tar -xzf "${TARBALL}" -C "${HERE}/src"

# 4. Configure as a dependency-free static library.
#    -fPIC is mandatory: the .a is linked into a cgo-produced shared object.
#    Subpixel rendering stays OFF (stock 2.6.1 default; matches matplotlib's
#    ftoption policy). Disable all optional external deps for reproducibility.
log "configuring (static, -fPIC, no external deps)"
(
  cd "${SRC}"
  CFLAGS="-fPIC -O2" ./configure \
    --prefix="${PREFIX}" \
    --enable-static --disable-shared \
    --with-zlib=no --with-bzip2=no --with-png=no \
    --with-harfbuzz=no --with-brotli=no \
    --without-old-mac-fonts
  make "-j$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 2)"
  make install
)

# 5. Sanity check.
if [ ! -f "${PREFIX}/lib/libfreetype.a" ]; then
  log "ERROR: build did not produce ${PREFIX}/lib/libfreetype.a"
  exit 1
fi
if [ ! -f "${PC}" ]; then
  log "ERROR: build did not produce ${PC}"
  exit 1
fi
log "built static libfreetype.a + freetype2.pc in ${PREFIX}"
log "consume via: PKG_CONFIG_PATH=\"${PREFIX}/lib/pkgconfig:\$PKG_CONFIG_PATH\""
