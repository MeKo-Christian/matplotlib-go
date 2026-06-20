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
	img := r.GetImage()
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

	got := r.GetImage().RGBAAt(50, 50)
	want := color.RGBA{R: 222, G: 233, B: 251, A: 255}
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

func TestPathPipelineSnapsNearHalfPixelTransformTiesLikeMatplotlib(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 267.49999999999994, Y: 126.2})
	p.LineTo(geom.Pt{X: 267.49999999999994, Y: 287.8})

	got := snapPath(p, &render.Paint{Snap: render.SnapAuto, Stroke: render.Color{A: 1}, LineWidth: 1})
	if got.V[0].X != 268.5 || got.V[1].X != 268.5 {
		t.Fatalf("near-half transform tie snapped to X vertices %v, want 268.5 like Matplotlib PathSnapper", got.V)
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

	got := r.GetImage().RGBAAt(320, 253)
	if got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("collection baseline edge pixel at y=253 stayed white; want Matplotlib draw_markers half-pixel coverage")
	}
}

func TestPathPipelineAutoSnapsRoundedRectCorners(t *testing.T) {
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

	if !shouldSnapPath(rounded, &render.Paint{Snap: render.SnapAuto}) {
		t.Fatal("SnapAuto should snap rounded rectangle paths with axis-aligned corner curves")
	}
	snapped := snapPath(rounded, &render.Paint{Snap: render.SnapAuto, Stroke: render.Color{A: 1}, LineWidth: 1})
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
	if chunks[1].V[0] != chunks[0].V[len(chunks[0].V)-1] {
		t.Fatalf("second chunk should restart at previous endpoint, got %v then %v", chunks[0].V, chunks[1].V)
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

	if got := r.GetImage().RGBAAt(10, 10); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
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

	bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
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

	slope, ok := darkPixelSlope(r.GetImage(), image.Rect(24, 24, 116, 76))
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

	runs := darkRunsOnScanline(r.GetImage(), 50, 25, 115)
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

			bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
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

	bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
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

	diffPixels, maxChannelDiff := hatchImageResidual(native.GetImage(), fallback.GetImage())
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
