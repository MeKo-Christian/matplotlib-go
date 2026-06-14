package render

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"golang.org/x/image/font/sfnt"
)

// TextGlyph is one positioned glyph from a laid-out text run.
type TextGlyph struct {
	Rune       rune
	Face       FontFace
	GlyphIndex sfnt.GlyphIndex
	Origin     geom.Pt
	Advance    float64
	Kern       float64
}

// TextGlyphLayout is a renderer-neutral, glyph-positioned single-line text
// layout. Size is interpreted as pixels-per-em by this low-level helper.
type TextGlyphLayout struct {
	Glyphs  []TextGlyph
	Advance float64
	Bounds  TextBounds
}

// LayoutTextGlyphs resolves text to font runs and applies pair kerning with
// the same ppem used for glyph advances. It intentionally avoids
// opentype.Face.Kern because that API has historically exposed adjustments in
// font-unit space rather than active-size pixel space.
func LayoutTextGlyphs(text string, origin geom.Pt, size float64, fontKey string) (TextGlyphLayout, bool) {
	shaped, ok := ShapeText(text, origin, size, TextShapingOptions{FontKey: fontKey})
	return glyphLayoutFromShapedText(shaped), ok
}

// LayoutTextGlyphRuns lays out already-resolved font runs.
func LayoutTextGlyphRuns(runs []FontRun, origin geom.Pt, size float64) (TextGlyphLayout, bool) {
	shaped, ok := ShapeTextRuns(runs, origin, size, TextShapingOptions{})
	return glyphLayoutFromShapedText(shaped), ok
}

func glyphLayoutFromShapedText(shaped ShapedText) TextGlyphLayout {
	layout := TextGlyphLayout{
		Advance: shaped.Advance.X,
		Bounds:  shaped.Bounds,
	}
	for _, glyph := range shaped.Glyphs {
		layout.Glyphs = append(layout.Glyphs, TextGlyph{
			Rune:       glyph.Rune,
			Face:       glyph.Face,
			GlyphIndex: glyph.GlyphIndex,
			Origin:     glyph.Origin,
			Advance:    glyph.Advance.X,
			Kern:       glyph.Kern,
		})
	}
	return layout
}
