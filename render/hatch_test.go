package render

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

type hatchRecordingRenderer struct {
	NullRenderer
	paths  []geom.Path
	paints []Paint
}

func (r *hatchRecordingRenderer) Path(p geom.Path, paint *Paint) {
	r.paths = append(r.paths, p)
	if paint == nil {
		r.paints = append(r.paints, Paint{})
		return
	}
	r.paints = append(r.paints, *paint)
}

func TestDrawHatchFallbackClipsToPolygon(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 10, Y: 0})
	clip.LineTo(geom.Pt{X: 0, Y: 10})
	clip.Close()

	r := &hatchRecordingRenderer{}
	ok := DrawHatchFallback(r, clip, Paint{
		Hatch:          "|",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   5,
	})
	if !ok {
		t.Fatal("DrawHatchFallback returned false")
	}
	if len(r.paths) == 0 {
		t.Fatal("expected clipped hatch paths")
	}

	for _, path := range r.paths {
		for _, pt := range path.V {
			if pt.X < -1e-9 || pt.Y < -1e-9 || pt.X+pt.Y > 10+1e-9 {
				t.Fatalf("hatch point %+v escaped triangular clip path in %+v", pt, path.V)
			}
		}
	}
}

func TestDrawHatchFallbackRepeatedPatternTightensSpacing(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 12, Y: 0})
	clip.LineTo(geom.Pt{X: 12, Y: 10})
	clip.LineTo(geom.Pt{X: 0, Y: 10})
	clip.Close()

	single := &hatchRecordingRenderer{}
	if !DrawHatchFallback(single, clip, Paint{
		Hatch:          "|",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	}) {
		t.Fatal("single hatch fallback returned false")
	}

	repeated := &hatchRecordingRenderer{}
	if !DrawHatchFallback(repeated, clip, Paint{
		Hatch:          "||",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	}) {
		t.Fatal("repeated hatch fallback returned false")
	}

	if got, want := hatchSegmentCount(repeated.paths), hatchSegmentCount(single.paths); got <= want {
		t.Fatalf("repeated hatch segment count = %d, want more than %d", got, want)
	}
}

func TestDrawHatchFallbackSupportsShapePatterns(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 20})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()

	for _, hatch := range []string{"o", "O", ".", "*"} {
		t.Run(hatch, func(t *testing.T) {
			r := &hatchRecordingRenderer{}
			if !DrawHatchFallback(r, clip, Paint{
				Hatch:          hatch,
				HatchColor:     Color{A: 1},
				HatchLineWidth: 1,
				HatchSpacing:   10,
			}) {
				t.Fatalf("DrawHatchFallback(%q) returned false", hatch)
			}
			if len(r.paths) == 0 {
				t.Fatalf("expected shape hatch paths for %q", hatch)
			}
			if hatch == "*" {
				if got := hatchSegmentCount(r.paths); got < 10 {
					t.Fatalf("star hatch line segments = %d, want visible star geometry", got)
				}
			} else if got := hatchCurveCount(r.paths); got == 0 {
				t.Fatalf("circle hatch %q had no curve commands: %+v", hatch, r.paths)
			}
			assertHatchPathsInsideRect(t, r.paths, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 20, Y: 20}})
		})
	}
}

func TestDrawHatchFallbackRepeatedShapePatternIncreasesDensity(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 20})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()

	single := &hatchRecordingRenderer{}
	if !DrawHatchFallback(single, clip, Paint{
		Hatch:          "o",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   10,
	}) {
		t.Fatal("single shape hatch fallback returned false")
	}
	repeated := &hatchRecordingRenderer{}
	if !DrawHatchFallback(repeated, clip, Paint{
		Hatch:          "oo",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   10,
	}) {
		t.Fatal("repeated shape hatch fallback returned false")
	}

	if got, want := hatchCurveCount(repeated.paths), hatchCurveCount(single.paths); got <= want {
		t.Fatalf("repeated shape curve count = %d, want more than %d", got, want)
	}
}

func TestDrawHatchFallbackShapeSizesFollowMatplotlibRatios(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 20})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()

	for _, tc := range []struct {
		hatch     string
		wantWidth float64
	}{
		{hatch: ".", wantWidth: 2},
		{hatch: "o", wantWidth: 4},
		{hatch: "O", wantWidth: 7},
		{hatch: "*", wantWidth: 6.3403767753},
	} {
		t.Run(tc.hatch, func(t *testing.T) {
			r := &hatchRecordingRenderer{}
			if !DrawHatchFallback(r, clip, Paint{
				Hatch:          tc.hatch,
				HatchColor:     Color{A: 1},
				HatchLineWidth: 1,
				HatchSpacing:   10,
			}) {
				t.Fatalf("DrawHatchFallback(%q) returned false", tc.hatch)
			}
			bounds, ok := pathBoundsForTest(r.paths[0])
			if !ok {
				t.Fatalf("missing hatch path bounds for %q", tc.hatch)
			}
			if math.Abs(bounds.W()-tc.wantWidth) > 1e-6 {
				t.Fatalf("shape hatch %q width = %g, want %g", tc.hatch, bounds.W(), tc.wantWidth)
			}
		})
	}
}

func TestDrawHatchFallbackUnfilledCircleHatchUsesMatplotlibRingContour(t *testing.T) {
	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 20})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()

	r := &hatchRecordingRenderer{}
	if !DrawHatchFallback(r, clip, Paint{
		Hatch:          "o",
		HatchColor:     Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   10,
	}) {
		t.Fatal("DrawHatchFallback returned false")
	}
	if len(r.paths) == 0 {
		t.Fatal("expected circle hatch paths")
	}
	if got := hatchMoveCount(r.paths[0]); got != 2 {
		t.Fatalf("unfilled circle hatch subpaths = %d, want outer and reversed inner contours", got)
	}
	if r.paints[0].Fill.A <= 0 {
		t.Fatalf("unfilled circle hatch should fill the annular contour like Matplotlib, got paint %+v", r.paints[0])
	}

	bounds, ok := pathBoundsForTest(r.paths[0])
	if !ok {
		t.Fatal("missing first hatch bounds")
	}
	if math.Abs(bounds.W()-4) > 1e-6 {
		t.Fatalf("outer circle hatch width = %g, want 4", bounds.W())
	}
}

func hatchSegmentCount(paths []geom.Path) int {
	count := 0
	for _, path := range paths {
		for _, cmd := range path.C {
			if cmd == geom.LineTo {
				count++
			}
		}
	}
	return count
}

func hatchMoveCount(path geom.Path) int {
	count := 0
	for _, cmd := range path.C {
		if cmd == geom.MoveTo {
			count++
		}
	}
	return count
}

func hatchCurveCount(paths []geom.Path) int {
	count := 0
	for _, path := range paths {
		for _, cmd := range path.C {
			if cmd == geom.QuadTo || cmd == geom.CubicTo {
				count++
			}
		}
	}
	return count
}

func pathBoundsForTest(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	minX, maxX := path.V[0].X, path.V[0].X
	minY, maxY := path.V[0].Y, path.V[0].Y
	for _, pt := range path.V[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}
	return geom.Rect{Min: geom.Pt{X: minX, Y: minY}, Max: geom.Pt{X: maxX, Y: maxY}}, true
}

func assertHatchPathsInsideRect(t *testing.T, paths []geom.Path, rect geom.Rect) {
	t.Helper()
	for _, path := range paths {
		for _, pt := range path.V {
			if !rect.ContainsInclusive(pt) {
				t.Fatalf("hatch point %+v escaped rect %+v in path %+v", pt, rect, path.V)
			}
		}
	}
}
