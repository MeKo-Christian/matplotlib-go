package render

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestTextPathBuildsValidGlyphOutlines(t *testing.T) {
	withDejaVuFontManager(t)

	textPath, ok := TextPath("Ag", geom.Pt{X: 10, Y: 20}, 14, "DejaVu Sans")
	if !ok {
		t.Fatal("TextPath returned !ok")
	}
	if len(textPath.C) == 0 || len(textPath.V) == 0 {
		t.Fatalf("expected path commands and vertices, got %+v", textPath)
	}
	if !textPath.Validate() {
		t.Fatalf("invalid path: %+v", textPath)
	}
}

func TestTextGlyphLayoutAppliesKerningAtActiveSize(t *testing.T) {
	withDejaVuFontManager(t)

	const size = 72.0
	layout, ok := LayoutTextGlyphs("Tr", geom.Pt{}, size, "DejaVu Sans")
	if !ok || len(layout.Glyphs) != 2 {
		t.Fatalf("LayoutTextGlyphs(Tr) = %+v, %v; want two glyphs", layout, ok)
	}
	separateT, okT := LayoutTextGlyphs("T", geom.Pt{}, size, "DejaVu Sans")
	separateR, okR := LayoutTextGlyphs("r", geom.Pt{}, size, "DejaVu Sans")
	if !okT || !okR {
		t.Fatalf("single-glyph layouts failed: T=%+v/%v r=%+v/%v", separateT, okT, separateR, okR)
	}

	unkerned := separateT.Advance + separateR.Advance
	if layout.Advance >= unkerned-2 {
		t.Fatalf("expected visible negative kerning for Tr at %vpx: kerned=%v unkerned=%v glyphs=%+v", size, layout.Advance, unkerned, layout.Glyphs)
	}
	if math.Abs((layout.Glyphs[1].Origin.X-separateT.Advance)-layout.Glyphs[1].Kern) > 1e-6 {
		t.Fatalf("second glyph origin does not include reported kern: layout=%+v separateT=%+v", layout, separateT)
	}
}

func TestTextGlyphLayoutKerningScalesWithSize(t *testing.T) {
	withDejaVuFontManager(t)

	small, okSmall := LayoutTextGlyphs("Te", geom.Pt{}, 12, "DejaVu Sans")
	large, okLarge := LayoutTextGlyphs("Te", geom.Pt{}, 72, "DejaVu Sans")
	if !okSmall || !okLarge || len(small.Glyphs) != 2 || len(large.Glyphs) != 2 {
		t.Fatalf("unexpected layouts: small=%+v/%v large=%+v/%v", small, okSmall, large, okLarge)
	}
	if small.Glyphs[1].Kern >= 0 || large.Glyphs[1].Kern >= 0 {
		t.Fatalf("expected negative kerning for Te: small=%+v large=%+v", small.Glyphs, large.Glyphs)
	}
	gotRatio := large.Glyphs[1].Kern / small.Glyphs[1].Kern
	if math.Abs(gotRatio-6) > 0.75 {
		t.Fatalf("kerning should scale with font size: small=%v large=%v ratio=%v", small.Glyphs[1].Kern, large.Glyphs[1].Kern, gotRatio)
	}
}

func TestShapeTextProducesGlyphRunsWithClustersAndBounds(t *testing.T) {
	withDejaVuFontManager(t)

	shaped, ok := ShapeText("Te", geom.Pt{X: 3, Y: 5}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(shaped.Runs) != 1 || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText = %+v, %v; want one two-glyph run", shaped, ok)
	}
	if shaped.Glyphs[0].Cluster != 0 || shaped.Glyphs[1].Cluster != 1 {
		t.Fatalf("clusters = [%d %d], want byte offsets [0 1]", shaped.Glyphs[0].Cluster, shaped.Glyphs[1].Cluster)
	}
	if shaped.Glyphs[1].Origin.X >= 3+shaped.Glyphs[0].Advance.X {
		t.Fatalf("second glyph origin should include negative kerning: shaped=%+v", shaped.Glyphs)
	}
	if shaped.Advance.X <= 0 || shaped.Advance.Y != 0 {
		t.Fatalf("advance = %+v, want positive horizontal advance", shaped.Advance)
	}
	if shaped.Bounds.W <= 0 || shaped.Bounds.H <= 0 {
		t.Fatalf("bounds = %+v, want non-empty ink bounds", shaped.Bounds)
	}
}

func TestShapeTextAppliesAndDisablesStandardLigatures(t *testing.T) {
	withDejaVuFontManager(t)

	shaped, ok := ShapeText("fi", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok {
		t.Fatal("ShapeText with default features failed")
	}
	withoutLiga, ok := ShapeText("fi", geom.Pt{}, 72, TextShapingOptions{
		FontKey:  "DejaVu Sans",
		Features: []TextFeature{{Tag: "liga", Value: 0}},
	})
	if !ok {
		t.Fatal("ShapeText with liga disabled failed")
	}

	if len(shaped.Glyphs) != 1 {
		t.Fatalf("default shaping glyph count = %d, want fi ligature as one glyph: %+v", len(shaped.Glyphs), shaped.Glyphs)
	}
	if len(withoutLiga.Glyphs) != 2 {
		t.Fatalf("liga=0 glyph count = %d, want separate f and i glyphs: %+v", len(withoutLiga.Glyphs), withoutLiga.Glyphs)
	}
	if shaped.Glyphs[0].Cluster != 0 {
		t.Fatalf("ligature cluster = %d, want first byte offset 0", shaped.Glyphs[0].Cluster)
	}
	if shaped.Glyphs[0].GlyphIndex == withoutLiga.Glyphs[0].GlyphIndex {
		t.Fatalf("ligature reused first component glyph index: shaped=%+v withoutLiga=%+v", shaped.Glyphs, withoutLiga.Glyphs)
	}
}

func TestShapeTextHonorsFeaturesFromStructuredFontPropertiesKey(t *testing.T) {
	withDejaVuFontManager(t)

	fontKey := FontPropertiesKey(FontProperties{
		Families: []string{"DejaVu Sans"},
		Features: []TextFeature{
			{Tag: "liga", Value: 0},
		},
	})
	shaped, ok := ShapeText("fi", geom.Pt{}, 72, TextShapingOptions{FontKey: fontKey})
	if !ok {
		t.Fatal("ShapeText with structured font features failed")
	}
	if len(shaped.Glyphs) != 2 {
		t.Fatalf("structured liga=0 glyph count = %d, want separate f and i glyphs: %+v", len(shaped.Glyphs), shaped.Glyphs)
	}
	if len(shaped.Runs) != 1 || len(shaped.Runs[0].Features) != 1 || shaped.Runs[0].Features[0] != (TextFeature{Tag: "liga", Value: 0}) {
		t.Fatalf("structured font features were not preserved on shaped run: %+v", shaped.Runs)
	}
}

func TestShapeTextHonorsKernFeatureDisable(t *testing.T) {
	withDejaVuFontManager(t)

	kerned, ok := ShapeText("Te", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(kerned.Glyphs) != 2 {
		t.Fatalf("kerned ShapeText = %+v, %v; want two glyphs", kerned, ok)
	}
	withoutKern, ok := ShapeText("Te", geom.Pt{}, 72, TextShapingOptions{
		FontKey:  "DejaVu Sans",
		Features: []TextFeature{{Tag: "kern", Value: 0}},
	})
	if !ok || len(withoutKern.Glyphs) != 2 {
		t.Fatalf("kern=0 ShapeText = %+v, %v; want two glyphs", withoutKern, ok)
	}

	if kerned.Glyphs[1].Kern >= 0 {
		t.Fatalf("default shaping should apply negative kerning for Te: %+v", kerned.Glyphs)
	}
	if withoutKern.Glyphs[1].Kern != 0 {
		t.Fatalf("kern=0 should disable kerning, got %+v", withoutKern.Glyphs)
	}
	if withoutKern.Glyphs[1].Origin.X <= kerned.Glyphs[1].Origin.X {
		t.Fatalf("kern=0 should move second glyph right: kerned=%+v without=%+v", kerned.Glyphs, withoutKern.Glyphs)
	}
}

func TestShapeTextComposesCombiningMarksWhenPrecomposedGlyphExists(t *testing.T) {
	withDejaVuFontManager(t)

	decomposed, ok := ShapeText("e\u0301", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok {
		t.Fatal("ShapeText decomposed e-acute failed")
	}
	precomposed, ok := ShapeText("\u00e9", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(precomposed.Glyphs) != 1 {
		t.Fatalf("ShapeText precomposed e-acute = %+v, %v; want one glyph", precomposed, ok)
	}

	if len(decomposed.Glyphs) != 1 {
		t.Fatalf("decomposed e-acute glyph count = %d, want one composed glyph: %+v", len(decomposed.Glyphs), decomposed.Glyphs)
	}
	if decomposed.Glyphs[0].GlyphIndex != precomposed.Glyphs[0].GlyphIndex {
		t.Fatalf("decomposed glyph index = %d, want precomposed glyph index %d", decomposed.Glyphs[0].GlyphIndex, precomposed.Glyphs[0].GlyphIndex)
	}
	if decomposed.Glyphs[0].Cluster != 0 {
		t.Fatalf("composed glyph cluster = %d, want original base byte offset 0", decomposed.Glyphs[0].Cluster)
	}
}

func TestShapeTextPlacesUncomposedCombiningMarksOnBaseGlyph(t *testing.T) {
	withDejaVuFontManager(t)

	base, ok := ShapeText("x", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(base.Glyphs) != 1 {
		t.Fatalf("ShapeText base x = %+v, %v; want one glyph", base, ok)
	}
	shaped, ok := ShapeText("x\u0301", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText x+combining acute = %+v, %v; want base plus mark glyph", shaped, ok)
	}

	mark := shaped.Glyphs[1]
	if mark.Rune != '\u0301' {
		t.Fatalf("second glyph rune = %q, want combining acute", mark.Rune)
	}
	if math.Abs(shaped.Advance.X-base.Advance.X) > 1e-6 {
		t.Fatalf("combining mark should not expand advance: base=%v shaped=%v glyphs=%+v", base.Advance.X, shaped.Advance.X, shaped.Glyphs)
	}
	if mark.Origin.X >= shaped.Glyphs[0].Origin.X+shaped.Glyphs[0].Advance.X {
		t.Fatalf("combining mark origin should stay over base glyph, got base=%+v mark=%+v", shaped.Glyphs[0], mark)
	}
	if mark.Advance.X != 0 {
		t.Fatalf("combining mark advance = %v, want zero", mark.Advance.X)
	}
}

func TestShapeTextAppliesGPOSMarkToBasePositioning(t *testing.T) {
	withDejaVuFontManager(t)

	shaped, ok := ShapeText("q\u0323", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText q+combining dot below = %+v, %v; want base plus mark glyph", shaped, ok)
	}

	mark := shaped.Glyphs[1]
	if mark.Rune != '\u0323' {
		t.Fatalf("second glyph rune = %q, want combining dot below", mark.Rune)
	}
	if mark.Offset == (geom.Pt{}) {
		t.Fatalf("combining mark should expose font-provided attachment offset: glyphs=%+v", shaped.Glyphs)
	}
	if math.Abs(mark.Offset.Y) < 1e-6 {
		t.Fatalf("font-provided mark attachment should include vertical placement, got offset=%+v glyphs=%+v", mark.Offset, shaped.Glyphs)
	}
	if math.Abs(mark.Origin.X-(shaped.Glyphs[0].Origin.X+mark.Offset.X)) > 1e-6 ||
		math.Abs(mark.Origin.Y-(shaped.Glyphs[0].Origin.Y+mark.Offset.Y)) > 1e-6 {
		t.Fatalf("mark origin should apply offset relative to base glyph: base=%+v mark=%+v", shaped.Glyphs[0], mark)
	}
}

func TestShapeTextPlacesExplicitRTLRunInVisualOrder(t *testing.T) {
	withDejaVuFontManager(t)

	shaped, ok := ShapeText("אב", geom.Pt{}, 72, TextShapingOptions{
		FontKey:   "DejaVu Sans",
		Direction: TextDirectionRTL,
	})
	if !ok || len(shaped.Runs) != 1 || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText RTL Hebrew = %+v, %v; want one two-glyph run", shaped, ok)
	}
	if shaped.Runs[0].Direction != TextDirectionRTL {
		t.Fatalf("run direction = %q, want rtl", shaped.Runs[0].Direction)
	}
	if shaped.Glyphs[0].Rune != 'ב' || shaped.Glyphs[1].Rune != 'א' {
		t.Fatalf("glyph visual rune order = %q %q, want bet then alef", shaped.Glyphs[0].Rune, shaped.Glyphs[1].Rune)
	}
	if shaped.Glyphs[0].Cluster != 2 || shaped.Glyphs[1].Cluster != 0 {
		t.Fatalf("glyph clusters = [%d %d], want original byte offsets [2 0]", shaped.Glyphs[0].Cluster, shaped.Glyphs[1].Cluster)
	}
	if shaped.Glyphs[1].Origin.X <= shaped.Glyphs[0].Origin.X {
		t.Fatalf("visual glyph origins should advance left-to-right after reordering: %+v", shaped.Glyphs)
	}
}

func TestShapeTextReordersMixedLTRAndRTLText(t *testing.T) {
	withDejaVuFontManager(t)

	shaped, ok := ShapeText("ab אב cd", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok {
		t.Fatal("ShapeText mixed bidi text failed")
	}
	var runes strings.Builder
	for _, glyph := range shaped.Glyphs {
		runes.WriteRune(glyph.Rune)
	}
	got := runes.String()
	if !strings.Contains(got, "בא") {
		t.Fatalf("mixed bidi glyph order = %q, want Hebrew run in visual order", got)
	}
	bet, alef := -1, -1
	for i, glyph := range shaped.Glyphs {
		switch glyph.Rune {
		case 'ב':
			bet = i
		case 'א':
			alef = i
		}
	}
	if bet < 0 || alef < 0 || bet > alef {
		t.Fatalf("Hebrew glyph indices in %q = bet %d alef %d, want bet before alef", got, bet, alef)
	}
	if shaped.Glyphs[bet].Cluster != 5 || shaped.Glyphs[alef].Cluster != 3 {
		t.Fatalf("Hebrew clusters = [%d %d], want original byte offsets [5 3]", shaped.Glyphs[bet].Cluster, shaped.Glyphs[alef].Cluster)
	}
}

func TestShapeTextUsesArabicGSUBPositionalSubstitutionsForJoiningLetters(t *testing.T) {
	withDejaVuFontManager(t)

	base, ok := ShapeText("م", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(base.Glyphs) != 1 {
		t.Fatalf("ShapeText Arabic base meem = %+v, %v; want one glyph", base, ok)
	}
	shaped, ok := ShapeText("مم", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText joined Arabic meem pair = %+v, %v; want two glyphs", shaped, ok)
	}

	if shaped.Glyphs[0].Cluster != 2 || shaped.Glyphs[1].Cluster != 0 {
		t.Fatalf("Arabic visual clusters = [%d %d], want original byte offsets [2 0]", shaped.Glyphs[0].Cluster, shaped.Glyphs[1].Cluster)
	}
	if shaped.Glyphs[0].GlyphIndex == base.Glyphs[0].GlyphIndex || shaped.Glyphs[1].GlyphIndex == base.Glyphs[0].GlyphIndex {
		t.Fatalf("joined Arabic glyphs should not reuse isolated base glyph: base=%+v joined=%+v", base.Glyphs, shaped.Glyphs)
	}
	if shaped.Glyphs[0].Rune != 'م' || shaped.Glyphs[1].Rune != 'م' {
		t.Fatalf("GSUB positional substitution should preserve logical runes, got %U %U", shaped.Glyphs[0].Rune, shaped.Glyphs[1].Rune)
	}
}

func TestShapeTextHonorsArabicPositionalFeatureDisable(t *testing.T) {
	withDejaVuFontManager(t)

	defaultShaped, ok := ShapeText("مم", geom.Pt{}, 72, TextShapingOptions{FontKey: "DejaVu Sans"})
	if !ok || len(defaultShaped.Glyphs) != 2 {
		t.Fatalf("default ShapeText joined Arabic meem pair = %+v, %v; want two glyphs", defaultShaped, ok)
	}
	shaped, ok := ShapeText("مم", geom.Pt{}, 72, TextShapingOptions{
		FontKey:  "DejaVu Sans",
		Features: []TextFeature{{Tag: "init", Value: 0}},
	})
	if !ok || len(shaped.Glyphs) != 2 {
		t.Fatalf("ShapeText joined Arabic meem pair with init=0 = %+v, %v; want two glyphs", shaped, ok)
	}

	if shaped.Glyphs[1].GlyphIndex == defaultShaped.Glyphs[1].GlyphIndex {
		t.Fatalf("init=0 should disable the default initial form, default=%+v disabled=%+v", defaultShaped.Glyphs, shaped.Glyphs)
	}
	if shaped.Glyphs[0].GlyphIndex != defaultShaped.Glyphs[0].GlyphIndex {
		t.Fatalf("init=0 should keep the final form for the trailing logical glyph, default=%+v disabled=%+v", defaultShaped.Glyphs, shaped.Glyphs)
	}
}

func TestTextPathBoundsUseSharedGlyphLayout(t *testing.T) {
	withDejaVuFontManager(t)

	for _, size := range []float64{12, 24, 72} {
		for _, text := range []string{"Tr", "Te", "AV", "fi", "x\u0301", "אב", "مم"} {
			path, ok := TextPath(text, geom.Pt{}, size, "DejaVu Sans")
			if !ok {
				t.Fatalf("TextPath(%q, %v) failed", text, size)
			}
			layout, ok := ShapeText(text, geom.Pt{}, size, TextShapingOptions{FontKey: "DejaVu Sans"})
			if !ok {
				t.Fatalf("ShapeText(%q, %v) failed", text, size)
			}
			bounds, ok := pathBounds(path)
			if !ok {
				t.Fatalf("TextPath(%q, %v) has no bounds", text, size)
			}
			if math.Abs(bounds.X-layout.Bounds.X) > 1e-6 ||
				math.Abs(bounds.Y-layout.Bounds.Y) > 1e-6 ||
				math.Abs(bounds.W-layout.Bounds.W) > 1e-6 ||
				math.Abs(bounds.H-layout.Bounds.H) > 1e-6 {
				t.Fatalf("path/layout bounds mismatch for %q at %vpx: path=%+v layout=%+v", text, size, bounds, layout.Bounds)
			}
		}
	}
}

func TestFontManagerResolveTextRunsUsesRequestedFace(t *testing.T) {
	manager := withDejaVuFontManager(t)

	runs, ok := manager.ResolveTextRuns("ABC", "DejaVu Sans")
	if !ok || len(runs) != 1 {
		t.Fatalf("ResolveTextRuns = %+v, %v; want one run", runs, ok)
	}
	if runs[0].Text != "ABC" || runs[0].Face.Path == "" {
		t.Fatalf("unexpected resolved run: %+v", runs[0])
	}
}

func TestTextPathAndLayoutUseEmbeddedFontFace(t *testing.T) {
	face, ok := embeddedFontFace("DejaVu Sans", FontProperties{Style: FontStyleNormal, Weight: 400})
	if !ok {
		t.Fatal("embeddedFontFace(DejaVu Sans) failed")
	}

	runs := []FontRun{{Text: "Ag", Face: face}}
	layout, ok := LayoutTextGlyphRuns(runs, geom.Pt{}, 14)
	if !ok {
		t.Fatal("LayoutTextGlyphRuns with embedded face failed")
	}
	if len(layout.Glyphs) != 2 {
		t.Fatalf("embedded layout glyph count = %d, want 2", len(layout.Glyphs))
	}
	if layout.Advance <= 0 {
		t.Fatalf("embedded layout advance = %v, want > 0", layout.Advance)
	}

	path, ok := textRunsPath(runs, geom.Pt{}, 14)
	if !ok {
		t.Fatal("textRunsPath with embedded face failed")
	}
	if len(path.C) == 0 || len(path.V) == 0 {
		t.Fatalf("embedded text path is empty: commands=%d vertices=%d", len(path.C), len(path.V))
	}
}

func TestTextPathSkipsMissingFontsAndEmptyInputs(t *testing.T) {
	if path, ok := TextPath("", geom.Pt{}, 12, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("empty text should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 0, "missing-font"); ok || len(path.C) != 0 {
		t.Fatalf("zero size should not build a path: %+v, %v", path, ok)
	}
	if path, ok := TextPath("A", geom.Pt{}, 12, filepath.Join(t.TempDir(), "missing.ttf")); ok || len(path.C) != 0 {
		t.Fatalf("missing font should not build a path: %+v, %v", path, ok)
	}
}

func withDejaVuFontManager(t *testing.T) *FontManager {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "DejaVuSans.ttf")
	if err := os.WriteFile(path, dejavusans.TTF, 0o644); err != nil {
		t.Fatalf("write test font: %v", err)
	}

	manager := NewFontManager()
	manager.AddFontDir(dir)
	previous := defaultFontManager
	defaultFontManager = manager
	t.Cleanup(func() { defaultFontManager = previous })
	return manager
}

func pathBounds(path geom.Path) (TextBounds, bool) {
	if len(path.V) == 0 {
		return TextBounds{}, false
	}
	minX, minY := path.V[0].X, path.V[0].Y
	maxX, maxY := minX, minY
	for _, pt := range path.V[1:] {
		minX = math.Min(minX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxX = math.Max(maxX, pt.X)
		maxY = math.Max(maxY, pt.Y)
	}
	return TextBounds{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}, true
}
