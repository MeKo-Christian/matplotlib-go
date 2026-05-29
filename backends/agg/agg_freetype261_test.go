//go:build cgo && !purego && freetype261

package agg

import "testing"

// TestNativeFreetypeIsPinned261 guards the FreeType-parity build: when built
// with `-tags freetype261` (see the Justfile `*-parity` targets, which export
// PKG_CONFIG_PATH at third_party/freetype/prefix), the AGG backend must link
// FreeType 2.6.1 — matplotlib's pinned version used to generate the reference
// images. A mismatch means PKG_CONFIG_PATH did not resolve the vendored build
// (run `just freetype261-build` and ensure the prefix is on PKG_CONFIG_PATH),
// and parity output will silently diverge.
func TestNativeFreetypeIsPinned261(t *testing.T) {
	if got := NativeFreetypeVersion(); got != "2.6.1" {
		t.Fatalf("parity build linked FreeType %q, want 2.6.1; check PKG_CONFIG_PATH and third_party/freetype/prefix", got)
	}
}
