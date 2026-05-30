//go:build !cgo || purego

package agg

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) drawNativeFreetypeText(_ string, _ render.FontFace, _ geom.Pt, _ float64, _ render.Color) bool {
	return false
}

func (r *Renderer) drawNativeFreetypeRunText(_ string, _ render.FontFace, _ geom.Pt, _ float64, _ render.Color, _ int) bool {
	return false
}

func (r *Renderer) measureNativeFreetypeText(_ string, _ render.FontFace, _ float64, _ int) (render.TextMetrics, bool) {
	return render.TextMetrics{}, false
}

func (r *Renderer) measureNativeFreetypeGlyphRun(_ string, _ string, _ float64, _ int) ([]render.MathGlyphMetric, bool) {
	return nil, false
}

// DrawMathTextImage is unavailable without cgo FreeType; returning false makes
// core fall back to the pure-Go subpixel run-mask mathtext path.
func (r *Renderer) DrawMathTextImage(_ []render.MathGlyphPlacement, _ []render.MathRectPlacement, _ geom.Pt, _ render.Color) bool {
	return false
}

func (r *Renderer) measureNativeFreetypeTextBounds(_ string, _ render.FontFace, _ float64, _ int) (render.TextBounds, bool) {
	return render.TextBounds{}, false
}

func (r *Renderer) measureNativeFreetypeFontHeights(_ render.FontFace, _ float64, _ int) (render.FontHeightMetrics, bool) {
	return render.FontHeightMetrics{}, false
}

func nativeFreetypeVersion() string {
	return ""
}
