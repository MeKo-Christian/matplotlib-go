//go:build cgo && !purego && !systemfreetype

package agg

import "testing"

// TestNativeFreetypeIsPinned261 guards the default cgo build: the AGG backend
// statically links the vendored FreeType 2.6.1 (matplotlib's pinned version,
// used to generate the reference images). A mismatch means the cgo flags in
// freetype_native.go resolved a different library — rebuild the prefix with
// `just freetype261-build`. Under -tags systemfreetype this guard is skipped
// (that compile fallback intentionally links the system FreeType and is not
// parity-exact).
func TestNativeFreetypeIsPinned261(t *testing.T) {
	if got := NativeFreetypeVersion(); got != "2.6.1" {
		t.Fatalf("default build linked FreeType %q, want 2.6.1; run `just freetype261-build` to (re)build third_party/freetype/prefix", got)
	}
}
