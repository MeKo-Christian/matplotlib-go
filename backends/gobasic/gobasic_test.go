package gobasic

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestNew(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	if r == nil {
		t.Fatal("New returned nil")
	}

	img := r.GetImage()
	if img == nil {
		t.Fatal("GetImage returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 50 {
		t.Errorf("Expected dimensions 100x50, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Check that background color is set (sample a few pixels)
	expectedColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for _, pt := range []image.Point{{0, 0}, {50, 25}, {99, 49}} {
		if c := img.RGBAAt(pt.X, pt.Y); c != expectedColor {
			t.Errorf("Expected background color %v at %v, got %v", expectedColor, pt, c)
		}
	}
}

func TestImageScalingUsesPixelEdges(t *testing.T) {
	r := New(6, 1, render.Color{R: 1, G: 1, B: 1, A: 1})
	src := image.NewRGBA(image.Rect(0, 0, 3, 1))
	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	src.SetRGBA(0, 0, red)
	src.SetRGBA(1, 0, green)
	src.SetRGBA(2, 0, blue)

	r.Image(render.NewImageData(src), geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 6, Y: 1},
	})

	want := []color.RGBA{red, red, green, green, blue, blue}
	for x, expected := range want {
		if got := r.GetImage().RGBAAt(x, 0); got != expected {
			t.Fatalf("pixel %d = %#v, want %#v", x, got, expected)
		}
	}
}

func TestImageAppliesImageAlpha(t *testing.T) {
	r := New(1, 1, render.Color{R: 1, G: 1, B: 1, A: 1})
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})

	img := render.NewImageData(src)
	img.SetAlpha(0.25)
	r.Image(img, geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: 1, Y: 1},
	})

	got := r.GetImage().RGBAAt(0, 0)
	if got.G < 180 || got.B < 180 {
		t.Fatalf("image alpha was not applied before blending; got %+v", got)
	}
	if got.R < 250 {
		t.Fatalf("red source channel unexpectedly dimmed; got %+v", got)
	}
}

func TestImageTransformedAppliesAffineAndAlpha(t *testing.T) {
	r := New(12, 12, render.Color{})
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	img := render.NewImageData(src)
	img.SetAlpha(0.5)
	r.ImageTransformed(img, geom.Rect{}, geom.Affine{
		A: 2,
		D: 2,
		E: 3,
		F: 4,
	})

	// Display space is y-up: the backend composes a device y-flip into the
	// affine, so src row 0 (red) lands at the bottom of the placed image
	// (device y in {6,7}) rather than the top.
	got := r.GetImage().RGBAAt(3, 7)
	if got.A < 120 || got.A > 130 || got.R < 200 || got.G != 0 || got.B != 0 {
		t.Fatalf("transformed alpha image pixel = %+v, want half-alpha red", got)
	}
	if c := r.GetImage().RGBAAt(2, 7); c.A != 0 {
		t.Fatalf("pixel outside transformed image = %+v, want transparent", c)
	}
}

func TestBeginEnd(t *testing.T) {
	r := New(100, 50, render.Color{R: 0, G: 0, B: 0, A: 1})

	// Test Begin
	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Errorf("Begin failed: %v", err)
	}

	// Test double Begin should fail
	err = r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err == nil {
		t.Error("Expected Begin to fail when called twice")
	}

	// Test End
	err = r.End()
	if err != nil {
		t.Errorf("End failed: %v", err)
	}

	// Test End without Begin should fail
	err = r.End()
	if err == nil {
		t.Error("Expected End to fail when called without Begin")
	}
}

func TestSaveRestore(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Set a clip rect
	clipRect := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 90, Y: 40}}
	r.ClipRect(clipRect)

	// Save state
	r.Save()

	// Modify clip rect
	newClipRect := geom.Rect{Min: geom.Pt{X: 20, Y: 20}, Max: geom.Pt{X: 80, Y: 30}}
	r.ClipRect(newClipRect)

	// Restore should bring back original clip rect
	r.Restore()

	// Can't easily test the clip rect is restored without internal access,
	// but at least ensure Save/Restore don't crash
}

func TestPathFill(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Create a simple triangle path
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
		V: []geom.Pt{{X: 50, Y: 10}, {X: 30, Y: 40}, {X: 70, Y: 40}},
	}

	paint := render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1}, // Red fill
	}

	// Should not crash
	r.Path(path, &paint)

	// Check that some pixels changed from white background
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	// Check a few pixels around the triangle center
	for y := 20; y <= 30; y++ {
		for x := 45; x <= 55; x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Error("Expected some pixels to change from background color after drawing triangle")
	}
}

func TestClipPathMasksPathDrawing(t *testing.T) {
	r := New(100, 100, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	r.ClipPath(upperLeftTriangleClip())
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1},
	})

	// Display space is y-up, so the clip triangle (display verts (0,0),(70,0),
	// (0,70)) flips to the device buffer's lower-left; the clipped-in sample is
	// therefore near device (10,90) and the clipped-out corner stays at (90,90).
	if got := r.GetImage().RGBAAt(10, 90); got.R <= 200 || got.G >= 80 || got.B >= 80 {
		t.Fatalf("expected clipped-in pixel to be red, got %+v", got)
	}
	if got := r.GetImage().RGBAAt(90, 90); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("expected clipped-out pixel to remain white, got %+v", got)
	}
	if len(r.clipMaskMap) != 1 {
		t.Fatalf("expected one cached clip mask, got %d", len(r.clipMaskMap))
	}
}

func TestClipPathRestoreStopsMasking(t *testing.T) {
	r := New(100, 100, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	r.Save()
	r.ClipPath(upperLeftTriangleClip())
	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 1, G: 0, B: 0, A: 1},
	})
	r.Restore()

	r.Path(fullRectPath(100, 100), &render.Paint{
		Fill: render.Color{R: 0, G: 0, B: 1, A: 1},
	})
	if got := r.GetImage().RGBAAt(90, 90); got.B <= 200 || got.R >= 80 || got.G >= 80 {
		t.Fatalf("expected restore to remove path clipping, got %+v", got)
	}
}

func TestPathStroke(t *testing.T) {
	r := New(100, 50, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 100, Y: 50}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Create a simple line path
	path := geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 10, Y: 25}, {X: 90, Y: 25}},
	}

	paint := render.Paint{
		Stroke:    render.Color{R: 0, G: 0, B: 1, A: 1}, // Blue stroke
		LineWidth: 2.0,
	}

	// Should not crash
	r.Path(path, &paint)

	// Check that some pixels changed from white background
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	// Check pixels along the line
	for x := 20; x <= 80; x += 10 {
		if c := img.RGBAAt(x, 25); c != whiteColor {
			changed = true
			break
		}
	}

	if !changed {
		t.Error("Expected some pixels to change from background color after drawing line")
	}
}

func TestClipRectAllowsPathDrawingAcrossIndependentRegions(t *testing.T) {
	r := New(240, 240, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 240, Y: 240}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	quadrants := []geom.Rect{
		{Min: geom.Pt{X: 20, Y: 20}, Max: geom.Pt{X: 100, Y: 100}},
		{Min: geom.Pt{X: 140, Y: 20}, Max: geom.Pt{X: 220, Y: 100}},
		{Min: geom.Pt{X: 20, Y: 140}, Max: geom.Pt{X: 100, Y: 220}},
		{Min: geom.Pt{X: 140, Y: 140}, Max: geom.Pt{X: 220, Y: 220}},
	}
	colors := []render.Color{
		{R: 1, G: 0, B: 0, A: 1},
		{R: 0, G: 1, B: 0, A: 1},
		{R: 0, G: 0, B: 1, A: 1},
		{R: 1, G: 0.5, B: 0, A: 1},
	}

	for i, clip := range quadrants {
		r.Save()
		r.ClipRect(clip)
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.LineTo, geom.ClosePath},
			V: []geom.Pt{
				{X: clip.Min.X + 8, Y: clip.Min.Y + 8},
				{X: clip.Max.X - 8, Y: clip.Min.Y + 8},
				{X: clip.Max.X - 8, Y: clip.Max.Y - 8},
				{X: clip.Min.X + 8, Y: clip.Max.Y - 8},
			},
		}
		r.Path(path, &render.Paint{Fill: colors[i]})
		r.Restore()
	}

	img := r.GetImage()
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i, clip := range quadrants {
		center := image.Point{
			X: int((clip.Min.X + clip.Max.X) / 2),
			Y: int((clip.Min.Y + clip.Max.Y) / 2),
		}
		if got := img.RGBAAt(center.X, center.Y); got == white {
			t.Fatalf("quadrant %d center remained background at %v", i, center)
		}
	}
}

func TestMeasureText(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Test empty string
	metrics := r.MeasureText("", 12, "default")
	if metrics.W != 0 || metrics.H != 0 {
		t.Errorf("Expected zero metrics for empty string, got W=%v, H=%v", metrics.W, metrics.H)
	}

	// Test basic text
	metrics = r.MeasureText("Hello", 13, "default")
	if metrics.W <= 0 || metrics.H <= 0 {
		t.Errorf("Expected positive metrics for text, got W=%v, H=%v", metrics.W, metrics.H)
	}

	// Test scaling - larger size should give larger metrics
	metricsSmall := r.MeasureText("Test", 10, "default")
	metricsLarge := r.MeasureText("Test", 20, "default")
	if metricsLarge.W <= metricsSmall.W || metricsLarge.H <= metricsSmall.H {
		t.Errorf("Expected larger metrics for larger size, got small: W=%v,H=%v, large: W=%v,H=%v",
			metricsSmall.W, metricsSmall.H, metricsLarge.W, metricsLarge.H)
	}
}

func TestMeasureTextTracksRendererDPI(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	r.SetResolution(72)
	width72 := r.MeasureText("Basic Bars", 12, "default").W
	r.SetResolution(144)
	width144 := r.MeasureText("Basic Bars", 12, "default").W

	if width72 <= 0 || width144 <= 0 {
		t.Fatalf("expected positive text widths, got 72dpi=%v 144dpi=%v", width72, width144)
	}
	if width144 <= width72*1.8 {
		t.Fatalf("expected text width to scale with DPI, got 72dpi=%v 144dpi=%v", width72, width144)
	}
}

func TestDrawTextRenders(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Text should render visible glyphs in GoBasic.
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1} // black
	origin := geom.Pt{X: 10, Y: 50}
	r.DrawText("Hello, World!", origin, 13, textColor)

	// Verify that the image has changed from the white background.
	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Fatal("Expected at least one pixel to change after DrawText")
	}
}

func TestDrawTextRotatedRenders(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}
	r.DrawTextRotated("Hello", geom.Pt{X: 100, Y: 50}, 13, math.Pi/4, textColor)

	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}

	if !changed {
		t.Fatal("Expected at least one pixel to change after DrawTextRotated")
	}
}

func TestGlyphRun(t *testing.T) {
	r := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})

	err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 200, Y: 100}})
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer r.End()

	// Test GlyphRun - should not panic even with limited implementation
	glyphRun := render.GlyphRun{
		Glyphs:  []render.Glyph{{ID: 'H', Advance: 7, Offset: geom.Pt{}}},
		Origin:  geom.Pt{X: 10, Y: 50},
		Size:    13,
		FontKey: "default",
	}
	textColor := render.Color{R: 0, G: 0, B: 0, A: 1}

	// Should render a glyph when the ID maps to a visible rune.
	r.GlyphRun(glyphRun, textColor)

	img := r.GetImage()
	whiteColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	changed := false
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if c := img.RGBAAt(x, y); c != whiteColor {
				changed = true
				break
			}
		}
		if changed {
			break
		}
	}
	if !changed {
		t.Fatal("Expected DrawText to render at least one pixel")
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

func fullRectPath(w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: h})
	p.LineTo(geom.Pt{X: 0, Y: h})
	p.Close()
	return p
}
