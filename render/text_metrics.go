package render

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// This file provides pure-Go text measurement shared by the vector backends
// (PDF/PS/PGF/SVG). They already draw glyphs through TextPath, which shapes text
// with the same sfnt-based stack used here, so routing MeasureText through these
// helpers makes rotated/vertical anchoring consistent with the rendered glyphs
// (and close to the AGG raster backend, which uses native FreeType metrics).
//
// Per-glyph mathtext metrics (MeasureMathGlyphRun / MathGlyphMeasurer) are
// deliberately NOT provided here: reproducing Matplotlib's _get_info hinting
// requires FreeType and stays AGG-only.

// MeasureTextMetrics returns single-line text metrics derived from the real
// resolved font: advance width plus per-string ink ascent/descent. The size is
// interpreted as pixels-per-em (the same convention TextPath/ShapeText use), so
// it is self-consistent with the glyph outlines the vector backends emit.
//
// On shaping failure with non-empty text it falls back to a crude proportional
// estimate so layout code never divides by zero in font-less environments.
func MeasureTextMetrics(text string, size float64, fontKey string) TextMetrics {
	if text == "" || size <= 0 {
		return TextMetrics{}
	}

	layout, ok := LayoutTextGlyphs(text, geom.Pt{}, size, fontKey)
	if !ok || layout.Advance <= 0 {
		// Crude fallback (legacy behavior) keeps layout finite without fonts.
		return TextMetrics{
			W:       0.6 * size * float64(len([]rune(text))),
			H:       size,
			Ascent:  0.8 * size,
			Descent: 0.2 * size,
		}
	}

	ascent := math.Max(0, -layout.Bounds.Y)
	descent := math.Max(0, layout.Bounds.Y+layout.Bounds.H)
	h := ascent + descent
	if h <= 0 {
		h = size
		ascent = size
		descent = 0
	}
	return TextMetrics{
		W:       layout.Advance,
		H:       h,
		Ascent:  ascent,
		Descent: descent,
	}
}

// MeasureTextInkBounds returns the ink bounds of text relative to the baseline
// origin (y-up: Y is negative above the baseline). Reports ok only when the
// bounds are non-degenerate.
func MeasureTextInkBounds(text string, size float64, fontKey string) (TextBounds, bool) {
	if text == "" || size <= 0 {
		return TextBounds{}, false
	}
	layout, ok := LayoutTextGlyphs(text, geom.Pt{}, size, fontKey)
	if !ok || layout.Bounds.W <= 0 || layout.Bounds.H <= 0 {
		return TextBounds{}, false
	}
	return layout.Bounds, true
}

// MeasureFontHeightMetrics returns font-wide vertical metrics (ascent, descent,
// line gap) for the primary resolved face at the given pixels-per-em size,
// read from the OS/2 (preferred) or hhea table.
func MeasureFontHeightMetrics(size float64, fontKey string) (FontHeightMetrics, bool) {
	if size <= 0 {
		return FontHeightMetrics{}, false
	}
	face, ok := resolvePrimaryFace(fontKey)
	if !ok {
		return FontHeightMetrics{}, false
	}
	data, err := loadFontFaceData(face)
	if err != nil {
		return FontHeightMetrics{}, false
	}
	head, hok := sfntTable(data, "head")
	if !hok || len(head) < 20 {
		return FontHeightMetrics{}, false
	}
	unitsPerEm := be16(head, 18)
	if unitsPerEm == 0 {
		return FontHeightMetrics{}, false
	}
	// Vector size is already on-page ppem (no dpi/72 factor, unlike AGG's
	// point-size raster path).
	scale := size / float64(unitsPerEm)
	ascent, descent, lineGap, mok := sfntHeightMetrics(data, scale)
	if !mok {
		return FontHeightMetrics{}, false
	}
	return FontHeightMetrics{Ascent: ascent, Descent: descent, LineGap: lineGap}, true
}

// resolvePrimaryFace resolves a font key to its primary (requested) face,
// mirroring ResolveTextRuns' primary-face selection (including the empty-key
// DejaVu Sans fallback).
func resolvePrimaryFace(fontKey string) (FontFace, bool) {
	props := ParseFontProperties(fontKey)
	face, ok := DefaultFontManager().FindFont(props)
	if !ok && fontKey == "" {
		face, ok = DefaultFontManager().FindFont(ParseFontProperties("DejaVu Sans"))
	}
	if !ok || fontFaceCacheKey(face) == "" {
		return FontFace{}, false
	}
	return face, true
}

// sfntHeightMetrics reads scaled ascent/descent/lineGap from the OS/2 table
// (preferred) or hhea table. Mirrors backends/agg/surface.go sfntTableHeightMetrics.
func sfntHeightMetrics(data []byte, scale float64) (ascent, descent, lineGap float64, ok bool) {
	if table, tok := sfntTable(data, "OS/2"); tok && len(table) >= 74 {
		return scale * float64(int16(be16(table, 68))),
			scale * float64(-int16(be16(table, 70))),
			scale * float64(int16(be16(table, 72))),
			true
	}
	if table, tok := sfntTable(data, "hhea"); tok && len(table) >= 10 {
		return scale * float64(int16(be16(table, 4))),
			scale * float64(-int16(be16(table, 6))),
			scale * float64(int16(be16(table, 8))),
			true
	}
	return 0, 0, 0, false
}
