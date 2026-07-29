package agg

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPathLine(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 90})

	paint := &render.Paint{
		LineWidth: 2.0,
		Stroke:    render.Color{R: 0, G: 0, B: 0, A: 1},
	}

	// Should not panic
	r.Path(p, paint)
	_ = r.End()
}

func TestPathFill(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 10})
	p.LineTo(geom.Pt{X: 90, Y: 90})
	p.LineTo(geom.Pt{X: 10, Y: 90})
	p.Close()

	paint := &render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1},
	}

	r.Path(p, paint)
	_ = r.End()

	// Verify something was drawn (pixel at center should be red)
	img := r.Image()
	c := img.RGBAAt(50, 50)
	if c.R < 200 {
		t.Errorf("center pixel should be red, got R=%d", c.R)
	}
}

func TestPathFillUsesStraightAlphaColor(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2},
	})
	_ = r.End()

	got := r.Image().RGBAAt(50, 50)
	// Verified against Matplotlib 3.10.9 rendering the same fill over white.
	want := color.RGBA{R: 222, G: 232, B: 251, A: 255}
	if got != want {
		t.Fatalf("straight-alpha fill pixel = %+v, want %+v", got, want)
	}
}

func TestPathPipelineRemovesNonFiniteVertices(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: math.NaN(), Y: 10})
	p.LineTo(geom.Pt{X: 20, Y: 10})
	p.LineTo(geom.Pt{X: 20, Y: 20})

	got := removeNonFinitePathVertices(p)
	if len(got.C) != 3 {
		t.Fatalf("expected split two-command subpath after invalid segment, got commands %v", got.C)
	}
	if got.C[0] != geom.MoveTo || got.C[1] != geom.MoveTo || got.C[2] != geom.LineTo {
		t.Fatalf("unexpected commands after NaN cleanup: %v", got.C)
	}
	if got.V[1] != (geom.Pt{X: 20, Y: 10}) {
		t.Fatalf("expected next finite point to restart subpath, got vertices %v", got.V)
	}
}

func TestPathPipelineCullsOutsideVisibleArea(t *testing.T) {
	r := mustNew(t, 100, 100)
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}}); err != nil {
		t.Fatal(err)
	}
	r.ClipRect(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})

	var outside geom.Path
	outside.MoveTo(geom.Pt{X: 50, Y: 50})
	outside.LineTo(geom.Pt{X: 60, Y: 50})
	outside.LineTo(geom.Pt{X: 60, Y: 60})
	outside.Close()

	_, ok := r.preparePathForPaint(outside, &render.Paint{Fill: render.Color{A: 1}})
	if ok {
		t.Fatal("expected path entirely outside clip to be culled")
	}
	_ = r.End()
}

func TestPathPipelineSnapsWithLineWidthAwareAlignment(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 1.2, Y: 2.2})
	p.LineTo(geom.Pt{X: 3.2, Y: 2.2})

	if shouldSnapPath(p, &render.Paint{}) {
		t.Fatal("zero-value paint should preserve existing unsnapped behavior")
	}
	if !shouldSnapPath(p, &render.Paint{Snap: render.SnapAuto}) {
		t.Fatal("SnapAuto should snap simple horizontal paths")
	}

	odd := snapPath(p, &render.Paint{Snap: render.SnapOn, Stroke: render.Color{A: 1}, LineWidth: 1})
	if odd.V[0] != (geom.Pt{X: 1.5, Y: 2.5}) || odd.V[1] != (geom.Pt{X: 3.5, Y: 2.5}) {
		t.Fatalf("odd-width snap should center on half pixels, got %v", odd.V)
	}

	even := snapPath(p, &render.Paint{Snap: render.SnapOn, Stroke: render.Color{A: 1}, LineWidth: 2})
	if even.V[0] != (geom.Pt{X: 1, Y: 2}) || even.V[1] != (geom.Pt{X: 3, Y: 2}) {
		t.Fatalf("even-width snap should align to whole pixels, got %v", even.V)
	}
}

// TestPathSnapMatchesMatplotlibPathSnapperExactly pins snapPath to
// PathSnapper::vertex in matplotlib's path_converters.h, which is
// `floor(v + 0.5) + snap_value` with no tolerance around .5.
//
// The near-tie input is deliberate: units_categories' bar edge used to reach
// the snapper as 267.49999999999994 and an epsilon here rounded it up to match
// matplotlib's 268.5. That papered over the real divergence — matplotlib's own
// value is exactly 267.5, because it composes transLimits with transAxes into a
// single matrix before mapping the vertex. The composition is now done the same
// way (core.Axes.ensureTransforms), so the snapper must stay exact: an epsilon
// would move genuinely-below-tie coordinates a pixel right.
func TestPathSnapMatchesMatplotlibPathSnapperExactly(t *testing.T) {
	paint := render.Paint{Snap: render.SnapAuto, Stroke: render.Color{A: 1}, LineWidth: 1}
	cases := []struct {
		in   float64
		want float64
	}{
		{267.5, 268.5},              // matplotlib's composed value for units_categories
		{267.49999999999994, 267.5}, // one ULP below the tie stays below
		{267.2, 267.5},
		{267.8, 268.5},
	}
	for _, c := range cases {
		var p geom.Path
		p.MoveTo(geom.Pt{X: c.in, Y: 126.2})
		p.LineTo(geom.Pt{X: c.in, Y: 287.8})
		got := snapPath(p, &paint)
		if got.V[0].X != c.want || got.V[1].X != c.want {
			t.Errorf("snapPath(%.17g) X = %v, want %v", c.in, []float64{got.V[0].X, got.V[1].X}, c.want)
		}
	}
}

func TestDrawPathCollectionSingleUnsnappedPathUsesMarkerCachePlacement(t *testing.T) {
	r, err := New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New renderer: %v", err)
	}
	defer r.End()

	var p geom.Path
	p.MoveTo(geom.Pt{X: 115.2, Y: 144})
	p.LineTo(geom.Pt{X: 115.2, Y: 108})
	p.LineTo(geom.Pt{X: 320, Y: 108})
	p.LineTo(geom.Pt{X: 524.8, Y: 108})
	p.LineTo(geom.Pt{X: 524.8, Y: 165.6})
	p.LineTo(geom.Pt{X: 320, Y: 309.6})
	p.LineTo(geom.Pt{X: 115.2, Y: 144})
	p.Close()

	ok := r.DrawPathCollection(render.PathCollectionBatch{Items: []render.PathCollectionItem{{
		Path: p,
		Paint: render.Paint{
			Fill:      render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7},
			Stroke:    render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1},
			LineWidth: 2,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapButt,
			Snap:      render.SnapAuto,
		},
		Antialiased: true,
	}}})
	if !ok {
		t.Fatal("DrawPathCollection returned false")
	}

	got := r.Image().RGBAAt(320, 253)
	if got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("collection baseline edge pixel at y=253 stayed white; want Matplotlib draw_markers half-pixel coverage")
	}
}

func TestPathPipelineAutoDoesNotSnapCurvedPathsLikeMatplotlib(t *testing.T) {
	var rounded geom.Path
	rounded.MoveTo(geom.Pt{X: 1.2, Y: 2.2})
	rounded.LineTo(geom.Pt{X: 8.8, Y: 2.2})
	rounded.QuadTo(geom.Pt{X: 10.2, Y: 2.2}, geom.Pt{X: 10.2, Y: 3.6})
	rounded.LineTo(geom.Pt{X: 10.2, Y: 8.8})
	rounded.QuadTo(geom.Pt{X: 10.2, Y: 10.2}, geom.Pt{X: 8.8, Y: 10.2})
	rounded.LineTo(geom.Pt{X: 1.2, Y: 10.2})
	rounded.QuadTo(geom.Pt{X: -0.2, Y: 10.2}, geom.Pt{X: -0.2, Y: 8.8})
	rounded.LineTo(geom.Pt{X: -0.2, Y: 3.6})
	rounded.QuadTo(geom.Pt{X: -0.2, Y: 2.2}, geom.Pt{X: 1.2, Y: 2.2})
	rounded.Close()

	if shouldSnapPath(rounded, &render.Paint{Snap: render.SnapAuto}) {
		t.Fatal("SnapAuto should not snap curved paths; Matplotlib PathSnapper returns false for curve3/curve4")
	}
	if !shouldSnapPath(rounded, &render.Paint{Snap: render.SnapOn}) {
		t.Fatal("SnapOn should still force snapping curved paths")
	}
	snapped := snapPath(rounded, &render.Paint{Snap: render.SnapOn, Stroke: render.Color{A: 1}, LineWidth: 1})
	for _, pt := range snapped.V {
		if math.Abs(pt.X-math.Floor(pt.X)-0.5) > 1e-9 || math.Abs(pt.Y-math.Floor(pt.Y)-0.5) > 1e-9 {
			t.Fatalf("rounded rectangle vertex was not snapped to half-pixel center: %+v", pt)
		}
	}

	var arbitrary geom.Path
	arbitrary.MoveTo(geom.Pt{X: 1, Y: 2})
	arbitrary.QuadTo(geom.Pt{X: 3, Y: 4}, geom.Pt{X: 5, Y: 6})
	if shouldSnapPath(arbitrary, &render.Paint{Snap: render.SnapAuto}) {
		t.Fatal("SnapAuto should not snap arbitrary curved paths")
	}
}

func TestPathPipelineSimplifiesLinePaths(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 1, Y: 0.01})
	p.LineTo(geom.Pt{X: 2, Y: -0.01})
	p.LineTo(geom.Pt{X: 3, Y: 0})

	got := simplifyLinePath(p, 0.05)
	if len(got.V) != 2 {
		t.Fatalf("expected near-collinear path to simplify to endpoints, got %v", got.V)
	}
	if got.V[0] != (geom.Pt{X: 0, Y: 0}) || got.V[1] != (geom.Pt{X: 3, Y: 0}) {
		t.Fatalf("simplification should preserve endpoints, got %v", got.V)
	}
}

func TestPathPipelineChunksLargeOpenLinePaths(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{})
	for i := 1; i < 7; i++ {
		p.LineTo(geom.Pt{X: float64(i)})
	}

	chunks := chunkStrokePath(p, 4)
	if len(chunks) != 2 {
		t.Fatalf("expected two chunks, got %d: %v", len(chunks), chunks)
	}
	first, second := chunks[0].Path, chunks[1].Path
	if second.V[0] != first.V[len(first.V)-1] {
		t.Fatalf("second chunk should restart at previous endpoint, got %v then %v", first.V, second.V)
	}
	// The split is artificial, so the dash pattern has to carry across it: the
	// second chunk resumes at the arc length already covered by the first.
	if chunks[0].DashPhase != 0 {
		t.Fatalf("first chunk should start at phase 0, got %v", chunks[0].DashPhase)
	}
	if want := second.V[0].X; chunks[1].DashPhase != want {
		t.Fatalf("second chunk dash phase = %v, want %v", chunks[1].DashPhase, want)
	}
}

func TestPathChunkDashPhaseResetsAtGenuineSubpaths(t *testing.T) {
	// A real MoveTo restarts the pattern in AGG, so a chunk that opens on one
	// must report phase 0 even though earlier subpaths covered ground.
	var p geom.Path
	p.MoveTo(geom.Pt{})
	for i := 1; i < 5; i++ {
		p.LineTo(geom.Pt{X: float64(i)})
	}
	p.MoveTo(geom.Pt{X: 100})
	for i := 1; i < 5; i++ {
		p.LineTo(geom.Pt{X: 100 + float64(i)})
	}

	chunks := chunkStrokePath(p, 5)
	if len(chunks) != 2 {
		t.Fatalf("expected two chunks, got %d", len(chunks))
	}
	if chunks[1].Path.V[0].X != 100 {
		t.Fatalf("second chunk should open on the second subpath, got %v", chunks[1].Path.V[0])
	}
	if chunks[1].DashPhase != 0 {
		t.Fatalf("a chunk opening on a genuine MoveTo should have phase 0, got %v", chunks[1].DashPhase)
	}
}

func TestParallelRowRangesSplitLargeRegions(t *testing.T) {
	region := pixelRegion{minY: 3, maxY: 103}
	ranges := parallelRowRanges(region, 4)
	if len(ranges) != 4 {
		t.Fatalf("expected four row ranges, got %d: %+v", len(ranges), ranges)
	}
	if ranges[0].minY != 3 || ranges[len(ranges)-1].maxY != 103 {
		t.Fatalf("ranges should cover the original region, got %+v", ranges)
	}
	for i := 1; i < len(ranges); i++ {
		if ranges[i-1].maxY != ranges[i].minY {
			t.Fatalf("ranges should be contiguous, got %+v", ranges)
		}
	}

	small := parallelRowRanges(region, 1)
	if len(small) != 1 || small[0] != region {
		t.Fatalf("single worker should preserve the original region, got %+v", small)
	}
}

func TestPathAntialiasModeRestoresGamma(t *testing.T) {
	r := mustNew(t, 60, 60)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 60, Y: 60}})
	r.ctx.SetAntiAliasGamma(2)

	r.Path(fullRectPath(40, 40), &render.Paint{
		Fill:      render.Color{R: 1, A: 1},
		Antialias: render.AntialiasOff,
	})

	if got := r.ctx.GetAntiAliasGamma(); got != 2 {
		t.Fatalf("antialias mode should restore previous gamma, got %v", got)
	}
	_ = r.End()
}

func TestPathForcedAlphaOverridesPaintAlpha(t *testing.T) {
	r := mustNew(t, 40, 40)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 40, Y: 40}})

	r.Path(fullRectPath(30, 30), &render.Paint{
		Fill:       render.Color{R: 1, A: 1},
		ForceAlpha: true,
		Alpha:      0,
	})
	_ = r.End()

	if got := r.Image().RGBAAt(10, 10); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("forced transparent alpha should leave background unchanged, got %+v", got)
	}
}

func TestNativeHatchDrawsWithinPathClip(t *testing.T) {
	r := mustNew(t, 80, 80)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 80}})

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 20, Y: 20})
	rect.LineTo(geom.Pt{X: 60, Y: 20})
	rect.LineTo(geom.Pt{X: 60, Y: 60})
	rect.LineTo(geom.Pt{X: 20, Y: 60})
	rect.Close()
	r.Path(rect, &render.Paint{
		Hatch:          "|",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	})
	_ = r.End()

	bounds, pixels, ok := inkBounds(r.Image(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok || pixels == 0 {
		t.Fatal("expected hatch pixels to be drawn")
	}
	if bounds.Min.X < 19 || bounds.Min.Y < 19 || bounds.Max.X > 61 || bounds.Max.Y > 61 {
		t.Fatalf("native hatch should be clipped to path bounds, got %+v", bounds)
	}
}

func TestNativeDiagonalHatchUsesDeviceSpaceOrientation(t *testing.T) {
	r := mustNew(t, 140, 100)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 140, Y: 100}})

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 20, Y: 20})
	rect.LineTo(geom.Pt{X: 120, Y: 20})
	rect.LineTo(geom.Pt{X: 120, Y: 80})
	rect.LineTo(geom.Pt{X: 20, Y: 80})
	rect.Close()
	r.Path(rect, &render.Paint{
		Hatch:          "/",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
	})
	_ = r.End()

	slope, ok := darkPixelSlope(r.Image(), image.Rect(24, 24, 116, 76))
	if !ok {
		t.Fatal("expected diagonal hatch pixels")
	}
	if slope >= 0 {
		t.Fatalf("/ hatch image-space slope = %.3f, want negative like Matplotlib's AGG hatch tile", slope)
	}
}

func TestNativeDiagonalHatchDensityMatchesMatplotlibReference(t *testing.T) {
	r := mustNew(t, 140, 100)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 140, Y: 100}})

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 20, Y: 20})
	rect.LineTo(geom.Pt{X: 120, Y: 20})
	rect.LineTo(geom.Pt{X: 120, Y: 80})
	rect.LineTo(geom.Pt{X: 20, Y: 80})
	rect.Close()
	r.Path(rect, &render.Paint{
		Hatch:          "///",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
	})
	_ = r.End()

	runs := darkRunsOnScanline(r.Image(), 50, 25, 115)
	if runs < 9 || runs > 13 {
		t.Fatalf("/// hatch drew %d dark runs across tile scanline, want Matplotlib-like density around 11", runs)
	}
}

func TestNativeHatchDrawsShapePatterns(t *testing.T) {
	for _, hatch := range []string{"o", "O", ".", "*"} {
		t.Run(hatch, func(t *testing.T) {
			r := mustNew(t, 80, 80)
			_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 80}})

			var rect geom.Path
			rect.MoveTo(geom.Pt{X: 20, Y: 20})
			rect.LineTo(geom.Pt{X: 60, Y: 20})
			rect.LineTo(geom.Pt{X: 60, Y: 60})
			rect.LineTo(geom.Pt{X: 20, Y: 60})
			rect.Close()
			r.Path(rect, &render.Paint{
				Hatch:          hatch,
				HatchColor:     render.Color{A: 1},
				HatchLineWidth: 1,
				HatchSpacing:   8,
			})
			_ = r.End()

			bounds, pixels, ok := inkBounds(r.Image(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
			if !ok || pixels == 0 {
				t.Fatalf("expected native hatch %q pixels to be drawn", hatch)
			}
			if bounds.Min.X < 19 || bounds.Min.Y < 19 || bounds.Max.X > 61 || bounds.Max.Y > 61 {
				t.Fatalf("native hatch %q should be clipped to path bounds, got %+v", hatch, bounds)
			}
		})
	}
}

func TestNativeShapeHatchUsesTilePhaseAtClipBoundary(t *testing.T) {
	r := mustNew(t, 100, 80)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 80}})

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 20, Y: 20})
	rect.LineTo(geom.Pt{X: 60, Y: 20})
	rect.LineTo(geom.Pt{X: 60, Y: 60})
	rect.LineTo(geom.Pt{X: 20, Y: 60})
	rect.Close()
	r.Path(rect, &render.Paint{
		Hatch:          "o",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
	})
	_ = r.End()

	bounds, pixels, ok := inkBounds(r.Image(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok || pixels == 0 {
		t.Fatal("expected native shape hatch pixels")
	}
	if bounds.Min.X > 21 || bounds.Min.Y > 21 {
		t.Fatalf("shape hatch ink bounds = %+v, want tile-phased glyphs clipped at the patch boundary", bounds)
	}
}

func darkRunsOnScanline(img *image.RGBA, y, minX, maxX int) int {
	runs := 0
	inRun := false
	for x := minX; x <= maxX; x++ {
		c := img.RGBAAt(x, y)
		dark := c.R < 64 && c.G < 64 && c.B < 64 && c.A > 0
		if dark && !inRun {
			runs++
		}
		inRun = dark
	}
	return runs
}

func darkPixelSlope(img *image.RGBA, bounds image.Rectangle) (float64, bool) {
	var n int
	var sumX, sumY float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R < 64 && c.G < 64 && c.B < 64 && c.A > 0 {
				n++
				sumX += float64(x)
				sumY += float64(y)
			}
		}
	}
	if n < 2 {
		return 0, false
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)
	var cov, varX float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R < 64 && c.G < 64 && c.B < 64 && c.A > 0 {
				dx := float64(x) - meanX
				cov += dx * (float64(y) - meanY)
				varX += dx * dx
			}
		}
	}
	if varX == 0 {
		return 0, false
	}
	return cov / varX, true
}

func TestNativeHatchResidualAgainstFallbackDiagnostic(t *testing.T) {
	clip := upperLeftTriangleClip()
	paint := render.Paint{
		Hatch:          "/",
		HatchColor:     render.Color{A: 1},
		HatchLineWidth: 1,
		HatchSpacing:   8,
	}

	native := mustNew(t, 80, 80)
	_ = native.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 80}})
	native.Path(clip, &paint)
	_ = native.End()

	fallback := mustNew(t, 80, 80)
	_ = fallback.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 80}})
	if !render.DrawHatchFallback(fallback, clip, paint) {
		t.Fatal("renderer-neutral hatch fallback drew no paths")
	}
	_ = fallback.End()

	diffPixels, maxChannelDiff := hatchImageResidual(native.Image(), fallback.Image())
	t.Logf("native hatch vs fallback residual: diffPixels=%d maxChannelDiff=%d", diffPixels, maxChannelDiff)
	if diffPixels > 1600 || maxChannelDiff > 240 {
		t.Fatalf("native hatch residual too large: diffPixels=%d maxChannelDiff=%d", diffPixels, maxChannelDiff)
	}
}

func hatchImageResidual(a, b *image.RGBA) (diffPixels int, maxChannelDiff uint8) {
	bounds := a.Bounds().Intersect(b.Bounds())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ac := a.RGBAAt(x, y)
			bc := b.RGBAAt(x, y)
			dr := byteAbsDiff(ac.R, bc.R)
			dg := byteAbsDiff(ac.G, bc.G)
			db := byteAbsDiff(ac.B, bc.B)
			da := byteAbsDiff(ac.A, bc.A)
			maxDiff := maxByteDiff(dr, dg, db, da)
			if maxDiff == 0 {
				continue
			}
			diffPixels++
			if maxDiff > maxChannelDiff {
				maxChannelDiff = maxDiff
			}
		}
	}
	return diffPixels, maxChannelDiff
}

func byteAbsDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxByteDiff(values ...uint8) uint8 {
	var out uint8
	for _, v := range values {
		if v > out {
			out = v
		}
	}
	return out
}

// TestPathDashOffsetShiftsPattern renders one horizontal dashed line twice and
// checks that a dash offset moves the pattern along the line. Without it the
// Matplotlib gapcolor pass, which paints the gaps by re-stroking the same path
// with the rotated pattern, cannot be expressed at all.
func TestPathDashOffsetShiftsPattern(t *testing.T) {
	inkColumns := func(offset float64) map[int]bool {
		r := mustNew(t, 60, 20)
		var p geom.Path
		p.MoveTo(geom.Pt{X: 0, Y: 10})
		p.LineTo(geom.Pt{X: 60, Y: 10})
		r.Path(p, &render.Paint{
			LineWidth:  2,
			Stroke:     render.Color{A: 1},
			LineCap:    render.CapButt,
			Dashes:     []float64{10, 10},
			DashOffset: offset,
		})
		img := r.Image()
		cols := map[int]bool{}
		for x := 0; x < 60; x++ {
			cr, _, _, _ := img.At(x, 10).RGBA()
			if cr>>8 < 128 {
				cols[x] = true
			}
		}
		return cols
	}

	base := inkColumns(0)
	shifted := inkColumns(10)
	if len(base) == 0 || len(shifted) == 0 {
		t.Fatalf("expected dashed ink, got %d and %d columns", len(base), len(shifted))
	}
	// A half-period offset inverts the pattern: every on run becomes a gap.
	for x := range base {
		if shifted[x] {
			t.Fatalf("column %d is inked at both offset 0 and offset 10; dash offset was ignored", x)
		}
	}
}

// dashedLineColumns renders one long horizontal dashed line and returns the
// inked columns along it, so tests can compare rasterizations that should agree.
func dashedLineColumns(t *testing.T, width int, paint *render.Paint) map[int]bool {
	t.Helper()
	r := mustNew(t, width, 20)
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 10})
	for x := 1; x <= width; x++ {
		p.LineTo(geom.Pt{X: float64(x), Y: 10})
	}
	paint.LineWidth = 2
	paint.Stroke = render.Color{A: 1}
	paint.LineCap = render.CapButt
	r.Path(p, paint)

	img := r.Image()
	cols := map[int]bool{}
	for x := 0; x < width; x++ {
		cr, _, _, _ := img.At(x, 10).RGBA()
		if cr>>8 < 128 {
			cols[x] = true
		}
	}
	return cols
}

// TestPathDashPhaseSurvivesChunkBoundaries checks that splitting a long polyline
// into several Stroke calls does not restart the dash pattern. AGG's dash
// generator resets per stroke, so a chunked path has to re-enter the pattern at
// the arc length the previous chunk ended on.
func TestPathDashPhaseSurvivesChunkBoundaries(t *testing.T) {
	const width = 200
	dashes := []float64{7, 5} // a cycle that does not divide the chunk length

	whole := dashedLineColumns(t, width, &render.Paint{Dashes: dashes})
	chunked := dashedLineColumns(t, width, &render.Paint{Dashes: dashes, MaxChunkVertices: 16})

	if len(whole) == 0 {
		t.Fatal("expected dashed ink")
	}
	for x := 0; x < width; x++ {
		if whole[x] != chunked[x] {
			t.Fatalf("column %d: unchunked inked=%v, chunked inked=%v; dash phase restarted at a chunk boundary",
				x, whole[x], chunked[x])
		}
	}
}

// TestPathOddDashPatternCyclesInverted checks that an odd-length dash pattern
// alternates on/off across its second pass rather than repeating the first.
// Matplotlib walks such a sequence twice before handing it to AGG (the Dashes
// caster in src/_backend_agg_basic_types.h), and the pdf/ps/svg specs cycle odd
// arrays the same way.
func TestPathOddDashPatternCyclesInverted(t *testing.T) {
	const width = 200
	// [10, 5, 20] paints 10 on, 5 off, 20 on, 10 off, 5 on, 20 off, and only
	// then repeats — the same thing as the even pattern spelled out twice. The
	// old code dropped the unpaired trailing entry instead, collapsing this to
	// a 15px [10, 5] cycle.
	odd := dashedLineColumns(t, width, &render.Paint{Dashes: []float64{10, 5, 20}})
	doubled := dashedLineColumns(t, width, &render.Paint{Dashes: []float64{10, 5, 20, 10, 5, 20}})
	truncated := dashedLineColumns(t, width, &render.Paint{Dashes: []float64{10, 5}})

	if len(odd) == 0 {
		t.Fatal("odd-length dash pattern rendered nothing")
	}
	for x := 0; x < width; x++ {
		if odd[x] != doubled[x] {
			t.Fatalf("column %d: [10 5 20] inked=%v, [10 5 20 10 5 20] inked=%v; odd pattern did not cycle inverted",
				x, odd[x], doubled[x])
		}
	}
	same := true
	for x := 0; x < width; x++ {
		if odd[x] != truncated[x] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("[10 5 20] rendered as [10 5]; the unpaired entry was dropped rather than cycled")
	}
}
