package render

import (
	"testing"

	"golang.org/x/image/font/sfnt"
)

// TestGlyphIDToRuneRoundTrip verifies that a glyph index obtained from the
// forward cmap (rune -> glyph index) resolves back to the same rune. This is the
// contract the backends rely on: render.Glyph.ID is a glyph index, and the
// resolver inverts the cmap rather than casting the index to a rune.
func TestGlyphIDToRuneRoundTrip(t *testing.T) {
	withDejaVuFontManager(t)

	face, ok := DefaultFontManager().FindFont(ParseFontProperties("DejaVu Sans"))
	if !ok {
		t.Fatal("FindFont(DejaVu Sans) failed")
	}
	fontData, err := loadTextPathFontFace(face)
	if err != nil {
		t.Fatalf("load font: %v", err)
	}

	var buf sfnt.Buffer
	for _, want := range []rune{'A', 'g', '7', 'Z', '%'} {
		idx, err := fontData.GlyphIndex(&buf, want)
		if err != nil || idx == 0 {
			t.Fatalf("forward cmap for %q failed: idx=%d err=%v", want, idx, err)
		}
		got, ok := GlyphIDToRune("DejaVu Sans", uint32(idx))
		if !ok {
			t.Fatalf("GlyphIDToRune(%d) returned !ok for rune %q", idx, want)
		}
		if got != want {
			t.Fatalf("GlyphIDToRune(%d) = %q, want %q (direct rune(id) cast would give %q)", idx, got, want, rune(idx))
		}
	}

	// Glyph index 0 (.notdef) never resolves.
	if _, ok := GlyphIDToRune("DejaVu Sans", 0); ok {
		t.Fatal("GlyphIDToRune(0) should not resolve")
	}
}
