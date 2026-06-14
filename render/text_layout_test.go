package render

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

type textLayoutRenderer struct {
	metrics     TextMetrics
	bounds      TextBounds
	haveBounds  bool
	fontHeights FontHeightMetrics
	haveHeights bool
}

func (r textLayoutRenderer) Begin(geom.Rect) error    { return nil }
func (r textLayoutRenderer) End() error               { return nil }
func (r textLayoutRenderer) Save()                    {}
func (r textLayoutRenderer) Restore()                 {}
func (r textLayoutRenderer) ClipRect(geom.Rect)       {}
func (r textLayoutRenderer) ClipPath(geom.Path)       {}
func (r textLayoutRenderer) Path(geom.Path, *Paint)   {}
func (r textLayoutRenderer) Image(Image, geom.Rect)   {}
func (r textLayoutRenderer) GlyphRun(GlyphRun, Color) {}

func (r textLayoutRenderer) MeasureText(string, float64, string) TextMetrics {
	return r.metrics
}

func (r textLayoutRenderer) MeasureTextBounds(string, float64, string) (TextBounds, bool) {
	return r.bounds, r.haveBounds
}

func (r textLayoutRenderer) MeasureFontHeights(float64, string) (FontHeightMetrics, bool) {
	return r.fontHeights, r.haveHeights
}

func TestMeasureTextLineLayoutCombinesInkAndFontMetrics(t *testing.T) {
	layout := MeasureTextLineLayout(textLayoutRenderer{
		metrics:     TextMetrics{W: 30, H: 9, Ascent: 7, Descent: 2},
		bounds:      TextBounds{X: -1, Y: -6, W: 29, H: 8},
		haveBounds:  true,
		fontHeights: FontHeightMetrics{Ascent: 9, Descent: 3, LineGap: 2},
		haveHeights: true,
	}, "Ag", 12, "DejaVu Sans")

	if !layout.HaveInkBounds || layout.Width != 30 {
		t.Fatalf("layout bounds/width = %+v", layout)
	}
	if layout.RunAscent != 7 || layout.RunDescent != 2 {
		t.Fatalf("run extents = %v/%v, want 7/2", layout.RunAscent, layout.RunDescent)
	}
	if layout.Ascent != 9 || layout.Descent != 3 || layout.Height != 12 || layout.LineGap != 2 {
		t.Fatalf("font extents = %+v", layout)
	}
}

func TestMeasureTextLineLayoutFallsBackToMetrics(t *testing.T) {
	layout := MeasureTextLineLayout(textLayoutRenderer{
		metrics: TextMetrics{W: 24, H: 10, Ascent: 8, Descent: 2},
	}, "text", 12, "")

	if layout.HaveInkBounds {
		t.Fatalf("unexpected ink bounds: %+v", layout)
	}
	if layout.Ascent != 8 || layout.Descent != 2 || layout.Height != 10 {
		t.Fatalf("metric fallback = %+v", layout)
	}
}
