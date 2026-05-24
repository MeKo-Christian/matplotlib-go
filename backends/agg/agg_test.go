package agg

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

var white = render.Color{R: 1, G: 1, B: 1, A: 1}

// mustNew creates a renderer or fails the test.
func mustNew(t *testing.T, w, h int) *Renderer {
	t.Helper()
	r, err := New(w, h, white)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestNew(t *testing.T) {
	r := mustNew(t, 200, 100)
	if r.width != 200 || r.height != 100 {
		t.Errorf("unexpected dimensions: %dx%d", r.width, r.height)
	}
}

func TestNewInvalidDimensions(t *testing.T) {
	cases := []struct {
		name string
		w, h int
	}{
		{"zero width", 0, 100},
		{"zero height", 100, 0},
		{"negative width", -1, 100},
		{"negative height", 100, -1},
		{"both zero", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.w, tc.h, white)
			if err == nil {
				t.Errorf("New(%d, %d) should return error", tc.w, tc.h)
			}
		})
	}
}

func TestBeginEnd(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}

	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if err := r.Begin(viewport); err == nil {
		t.Fatal("double Begin should fail")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if err := r.End(); err == nil {
		t.Fatal("End before Begin should fail")
	}
}

func TestSaveRestore(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.Save()
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 50, Y: 50}})
	if r.clipRect == nil {
		t.Fatal("clip should be set after ClipRect")
	}
	r.Restore()
	if r.clipRect != nil {
		t.Fatal("clip should be nil after Restore")
	}
	_ = r.End()
}

func TestCopyFromBBoxAndRestoreRegion(t *testing.T) {
	r := mustNew(t, 80, 80)
	viewport := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 80, Y: 80}}
	_ = r.Begin(viewport)

	var redRect geom.Path
	redRect.MoveTo(geom.Pt{X: 10, Y: 10})
	redRect.LineTo(geom.Pt{X: 40, Y: 10})
	redRect.LineTo(geom.Pt{X: 40, Y: 40})
	redRect.LineTo(geom.Pt{X: 10, Y: 40})
	redRect.Close()

	r.Path(redRect, &render.Paint{Fill: render.Color{R: 1, A: 1}})

	region := r.CopyFromBBox(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 40, Y: 40}})
	if region == nil || region.Image == nil {
		t.Fatal("CopyFromBBox returned nil")
	}

	r.Path(fullRectPath(80, 80), &render.Paint{Fill: render.Color{B: 1, A: 1}})
	r.RestoreRegion(region, nil, geom.Pt{})
	_ = r.End()

	if got := r.GetImage().RGBAAt(20, 20); got != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected restored center pixel to be red, got %+v", got)
	}
	if got := r.GetImage().RGBAAt(5, 5); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("expected untouched pixel outside region to stay blue, got %+v", got)
	}
}

func TestRestoreRegionWithBBoxAndOffset(t *testing.T) {
	r := mustNew(t, 90, 90)
	viewport := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 90, Y: 90}}
	_ = r.Begin(viewport)

	var srcRect geom.Path
	srcRect.MoveTo(geom.Pt{X: 10, Y: 10})
	srcRect.LineTo(geom.Pt{X: 30, Y: 10})
	srcRect.LineTo(geom.Pt{X: 30, Y: 30})
	srcRect.LineTo(geom.Pt{X: 10, Y: 30})
	srcRect.Close()

	r.Path(srcRect, &render.Paint{Fill: render.Color{R: 1, A: 1}})

	region := r.CopyFromBBox(geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 30, Y: 30}})
	if region == nil || region.Image == nil {
		t.Fatal("CopyFromBBox returned nil")
	}

	r.Path(fullRectPath(90, 90), &render.Paint{Fill: render.Color{B: 1, A: 1}})
	r.RestoreRegion(region, &geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 10, Y: 10},
	}, geom.Pt{X: 20, Y: 20})
	_ = r.End()

	if got := r.GetImage().RGBAAt(35, 35); got != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected partial restored pixel to be red, got %+v", got)
	}
	if got := r.GetImage().RGBAAt(5, 5); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("expected non-restored pixel to remain blue, got %+v", got)
	}
}

func TestFilterStackStartStop(t *testing.T) {
	r := mustNew(t, 60, 60)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 60, Y: 60}})
	r.Path(fullRectPath(60, 60), &render.Paint{Fill: render.Color{G: 1, A: 1}})

	r.StartFilter()
	r.Path(fullRectPath(30, 30), &render.Paint{Fill: render.Color{R: 1, A: 1}})
	r.StopFilter(func(img *image.RGBA, _ float64) (*image.RGBA, geom.Pt) {
		out := &image.RGBA{
			Pix:    append([]uint8(nil), img.Pix...),
			Stride: img.Stride,
			Rect:   img.Rect,
		}
		for y := 0; y < out.Bounds().Dy(); y++ {
			for x := 0; x < out.Bounds().Dx(); x++ {
				off := out.PixOffset(x, y)
				if out.Pix[off+3] == 0 {
					continue
				}
				out.Pix[off+0] = 0
				out.Pix[off+1] = 0
				out.Pix[off+2] = 255
			}
		}
		return out, geom.Pt{X: 5, Y: 5}
	})
	_ = r.End()

	if got := r.GetImage().RGBAAt(15, 15); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("expected filtered-stop pixel to be blue, got %+v", got)
	}
	if got := r.GetImage().RGBAAt(2, 2); got != (color.RGBA{R: 0, G: 255, B: 0, A: 255}) {
		t.Fatalf("expected base green pixel to remain, got %+v", got)
	}
}

func TestPathEffectFilterUsesOffscreenSurface(t *testing.T) {
	r := mustNew(t, 60, 60)
	_ = r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 60, Y: 60}})
	r.Path(fullRectPath(60, 60), &render.Paint{Fill: render.Color{G: 1, A: 1}})

	var p geom.Path
	p.MoveTo(geom.Pt{X: 22, Y: 22})
	p.LineTo(geom.Pt{X: 38, Y: 22})
	p.LineTo(geom.Pt{X: 38, Y: 38})
	p.LineTo(geom.Pt{X: 22, Y: 38})
	p.Close()
	r.Path(p, &render.Paint{
		PathEffects: []render.PathEffect{
			render.FilterPathEffect(render.Color{R: 1, A: 1}, render.Color{}, 0, "blur", 4, geom.Pt{}),
		},
	})
	_ = r.End()

	if got := r.GetImage().RGBAAt(30, 30); got.R == 0 {
		t.Fatalf("expected filtered path center to contain red, got %+v", got)
	}
	if got := r.GetImage().RGBAAt(19, 30); got.R == 0 || got.G >= 255 {
		t.Fatalf("expected blurred red edge over green background, got %+v", got)
	}
}

func TestDrawTextWithFontDoesNotMutateLegacyFontState(t *testing.T) {
	r := mustNew(t, 120, 80)
	r.lastFontKey = "legacy-font"

	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 120, Y: 80}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.DrawTextWithFont("explicit", geom.Pt{X: 10, Y: 35}, 12, render.Color{A: 1}, "DejaVu Sans")
	r.DrawTextRotatedWithFont("rotated", geom.Pt{X: 60, Y: 50}, 12, math.Pi/8, render.Color{A: 1}, "DejaVu Sans")
	r.DrawTextVerticalWithFont("vertical", geom.Pt{X: 90, Y: 50}, 12, render.Color{A: 1}, "DejaVu Sans")
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if r.lastFontKey != "legacy-font" {
		t.Fatalf("lastFontKey = %q, want legacy-font", r.lastFontKey)
	}
}

func TestClipPathMasksPathDrawing(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.ClipPath(upperLeftTriangleClip())
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 1, A: 1},
	})
	_ = r.End()

	img := r.GetImage()
	if got := img.RGBAAt(10, 10); got.R < 200 {
		t.Fatalf("expected clipped-in pixel to be red, got %+v", got)
	}
	if got := img.RGBAAt(90, 90); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("expected clipped-out pixel to remain white, got %+v", got)
	}
	if len(r.clipMaskMap) != 1 {
		t.Fatalf("expected one cached clip mask, got %d", len(r.clipMaskMap))
	}
}

func TestClipPathPreservesStraightAlphaFillColor(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.ClipPath(fullRectPath(100, 100))
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 0.36, G: 0.56, B: 0.92, A: 0.2},
	})
	_ = r.End()

	got := r.GetImage().RGBAAt(50, 50)
	want := color.RGBA{R: 222, G: 233, B: 251, A: 255}
	if got != want {
		t.Fatalf("clipped straight-alpha fill pixel = %+v, want %+v", got, want)
	}
}

func TestClipPathRestoreStopsMasking(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	r.Save()
	r.ClipPath(upperLeftTriangleClip())
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 1, A: 1},
	})
	r.Restore()
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{B: 1, A: 1},
	})
	_ = r.End()

	if got := r.GetImage().RGBAAt(90, 90); got.B < 200 {
		t.Fatalf("expected restored clip state to allow blue fill, got %+v", got)
	}
}

func TestClipPathMasksImageDrawing(t *testing.T) {
	r := mustNew(t, 100, 100)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 100}}
	_ = r.Begin(viewport)

	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.SetRGBA(x, y, color.RGBA{G: 255, A: 255})
		}
	}

	r.ClipPath(upperLeftTriangleClip())
	r.Image(render.NewImageData(src), geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}})
	_ = r.End()

	img := r.GetImage()
	if got := img.RGBAAt(10, 10); got.G < 200 {
		t.Fatalf("expected clipped-in image pixel to be green, got %+v", got)
	}
	if got := img.RGBAAt(90, 90); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("expected clipped-out image pixel to remain white, got %+v", got)
	}
}

func TestSaveRestoreUnderflow(t *testing.T) {
	r := mustNew(t, 100, 100)
	// Restore on empty stack should not panic
	r.Restore()
}

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

func upperLeftTriangleClip() geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 70, Y: 0})
	p.LineTo(geom.Pt{X: 0, Y: 70})
	p.Close()
	return p
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

func fullRectPath(w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: h})
	p.LineTo(geom.Pt{X: 0, Y: h})
	p.Close()
	return p
}

func TestMeasureText(t *testing.T) {
	r := mustNew(t, 100, 100)
	m := r.MeasureText("Hello", 12.0, "")
	if m.W <= 0 || m.H <= 0 {
		t.Errorf("text metrics should be positive: W=%f H=%f", m.W, m.H)
	}

	empty := r.MeasureText("", 12.0, "")
	if empty.W != 0 || empty.H != 0 {
		t.Errorf("empty text should have zero metrics")
	}
}

func TestGlyphRunRendersShapedGlyphs(t *testing.T) {
	r := mustNew(t, 220, 120)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 220, Y: 120}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	layout, ok := render.LayoutTextGlyphs("Ag", geom.Pt{}, 24, "DejaVu Sans")
	if !ok {
		t.Fatal("LayoutTextGlyphs(\"Ag\") failed")
	}
	run := render.GlyphRun{
		Size:    24,
		Origin:  geom.Pt{X: 40, Y: 80},
		FontKey: "DejaVu Sans",
		Glyphs:  make([]render.Glyph, 0, len(layout.Glyphs)),
	}
	for _, glyph := range layout.Glyphs {
		run.Glyphs = append(run.Glyphs, render.Glyph{
			ID:     uint32(glyph.GlyphIndex),
			Offset: glyph.Origin,
		})
	}

	r.GlyphRun(run, render.Color{A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if _, _, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255}); !ok {
		t.Fatal("GlyphRun should draw visible text from shaped glyph IDs")
	}
}

func TestMeasureTextScalesWithResolution(t *testing.T) {
	r := mustNew(t, 100, 100)

	r.SetResolution(72)
	width72 := r.MeasureText("Hello", 12, "").W

	r.SetResolution(100)
	width100 := r.MeasureText("Hello", 12, "").W

	if width72 <= 0 || width100 <= 0 {
		t.Fatalf("expected positive widths, got 72dpi=%v 100dpi=%v", width72, width100)
	}
	if width100 <= width72 {
		t.Fatalf("expected width to increase with DPI, got 72dpi=%v 100dpi=%v", width72, width100)
	}
}

func TestDrawTextRotatedMaintainsReadableFootprint(t *testing.T) {
	r := mustNew(t, 220, 220)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 220, Y: 220}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	const size = 24.0
	metrics := r.MeasureText("Amplitude", size, "")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Fatalf("expected positive text metrics, got %+v", metrics)
	}

	r.DrawTextRotated("Amplitude", geom.Pt{X: 72, Y: 160}, size, math.Pi/2, render.Color{R: 0, G: 0, B: 0, A: 1})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	bounds, pixels, ok := inkBounds(r.GetImage(), color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok {
		t.Fatal("expected rotated text to draw visible ink")
	}
	if bounds.W() < metrics.H*0.75 {
		t.Fatalf("rotated text width too small: got=%v want_at_least=%v bounds=%+v", bounds.W(), metrics.H*0.75, bounds)
	}
	if bounds.H() < metrics.W*0.65 {
		t.Fatalf("rotated text height too small: got=%v want_at_least=%v bounds=%+v", bounds.H(), metrics.W*0.65, bounds)
	}
	if pixels < 250 {
		t.Fatalf("rotated text ink coverage unexpectedly sparse: pixels=%d bounds=%+v", pixels, bounds)
	}
}

func TestGetImage(t *testing.T) {
	r := mustNew(t, 200, 150)
	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("unexpected image dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestRendererInterface(t *testing.T) {
	var _ render.Renderer = (*Renderer)(nil)
}

func TestBatchInterfaces(t *testing.T) {
	var _ render.MarkerDrawer = (*Renderer)(nil)
	var _ render.PathCollectionDrawer = (*Renderer)(nil)
	var _ render.QuadMeshDrawer = (*Renderer)(nil)
	var _ render.GouraudTriangleDrawer = (*Renderer)(nil)
}

func TestDrawMarkersBatchDrawsVisibleMarkers(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	marker := geom.Path{}
	marker.MoveTo(geom.Pt{X: -2, Y: -2})
	marker.LineTo(geom.Pt{X: 2, Y: -2})
	marker.LineTo(geom.Pt{X: 2, Y: 2})
	marker.LineTo(geom.Pt{X: -2, Y: 2})
	marker.Close()

	ok := r.DrawMarkers(render.MarkerBatch{
		Marker: marker,
		Items: []render.MarkerItem{
			{
				Offset: geom.Pt{X: 10, Y: 10},
				Transform: geom.Affine{
					A: 1,
					D: 1,
				},
				Paint: render.Paint{Fill: render.Color{R: 1, A: 1}},
			},
			{
				Offset: geom.Pt{X: 25, Y: 25},
				Transform: geom.Affine{
					A: 1,
					D: 1,
				},
				Paint: render.Paint{Fill: render.Color{B: 1, A: 1}},
			},
		},
	})
	if !ok {
		t.Fatal("DrawMarkers returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	img := r.GetImage()
	if got := img.RGBAAt(10, 10); got.R < 200 {
		t.Fatalf("first marker center = %+v, want red", got)
	}
	if got := img.RGBAAt(25, 25); got.B < 200 {
		t.Fatalf("second marker center = %+v, want blue", got)
	}
}

func TestDrawQuadMeshBatchDrawsCells(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawQuadMesh(render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 5, Y: 5},
				{X: 20, Y: 5},
				{X: 20, Y: 20},
				{X: 5, Y: 20},
			},
			Face: render.Color{G: 1, A: 1},
		},
	}})
	if !ok {
		t.Fatal("DrawQuadMesh returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if got := r.GetImage().RGBAAt(10, 10); got.G < 200 {
		t.Fatalf("quad mesh cell center = %+v, want green", got)
	}
}

func TestDrawQuadMeshSnapsFractionalRectilinearEdges(t *testing.T) {
	r := mustNew(t, 40, 40)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 40, Y: 40}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawQuadMesh(render.QuadMeshBatch{Cells: []render.QuadMeshCell{
		{
			Quad: [4]geom.Pt{
				{X: 10.2, Y: 10.2},
				{X: 25.2, Y: 10.2},
				{X: 25.2, Y: 25.2},
				{X: 10.2, Y: 25.2},
			},
			Edge:      render.Color{A: 1},
			LineWidth: 1,
		},
	}})
	if !ok {
		t.Fatal("DrawQuadMesh returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	if got := r.GetImage().RGBAAt(10, 15); got.R > 16 {
		t.Fatalf("snapped quad mesh edge pixel = %+v, want nearly black", got)
	}
}

func TestDrawGouraudTrianglesInterpolatesVertexColors(t *testing.T) {
	r := mustNew(t, 60, 60)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 60, Y: 60}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	ok := r.DrawGouraudTriangles(render.GouraudTriangleBatch{Triangles: []render.GouraudTriangle{
		{
			P: [3]geom.Pt{
				{X: 5, Y: 5},
				{X: 45, Y: 5},
				{X: 5, Y: 45},
			},
			Color: [3]render.Color{
				{R: 1, A: 1},
				{G: 1, A: 1},
				{B: 1, A: 1},
			},
		},
	}})
	if !ok {
		t.Fatal("DrawGouraudTriangles returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	got := r.GetImage().RGBAAt(10, 10)
	if got == (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatal("triangle sample remained background white")
	}
	if got.R <= got.G || got.R <= got.B {
		t.Fatalf("triangle sample near red vertex = %+v, want red-dominant interpolation", got)
	}
}

func TestQuantize(t *testing.T) {
	cases := []struct {
		in, want float64
	}{
		{0, 0},
		{1.0, 1.0},
		{0.1234567890123, 0.123457}, // rounded to grid
		{-3.14159265, -3.141593},    // negative
		{1e-7, 0},                   // below grid, rounds to 0
		{0.0000005, 0.000001},       // half grid rounds up
		{100.123456789, 100.123457}, // large value
	}
	for _, tc := range cases {
		got := quantize(tc.in)
		if math.Abs(got-tc.want) > quantizationGrid/2 {
			t.Errorf("quantize(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestQuantizePt(t *testing.T) {
	pt := geom.Pt{X: 1.23456789, Y: 9.87654321}
	q := quantizePt(pt)
	if math.Abs(q.X-1.234568) > quantizationGrid {
		t.Errorf("X not quantized: %v", q.X)
	}
	if math.Abs(q.Y-9.876543) > quantizationGrid {
		t.Errorf("Y not quantized: %v", q.Y)
	}
}

func TestQuantizeIdempotent(t *testing.T) {
	v := 3.141592653589793
	q1 := quantize(v)
	q2 := quantize(q1)
	if q1 != q2 {
		t.Errorf("quantize not idempotent: %v != %v", q1, q2)
	}
}

func inkBounds(img *image.RGBA, background color.RGBA) (geom.Rect, int, bool) {
	if img == nil {
		return geom.Rect{}, 0, false
	}

	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	pixels := 0

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y) == background {
				continue
			}
			pixels++
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x >= maxX {
				maxX = x + 1
			}
			if y >= maxY {
				maxY = y + 1
			}
		}
	}

	if pixels == 0 {
		return geom.Rect{}, 0, false
	}

	return geom.Rect{
		Min: geom.Pt{X: float64(minX), Y: float64(minY)},
		Max: geom.Pt{X: float64(maxX), Y: float64(maxY)},
	}, pixels, true
}
