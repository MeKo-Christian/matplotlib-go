package agg

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

	// Display space is y-up: the red rect at display (10,10)-(40,40) occupies
	// device rows 40-70 (H-y), and the captured region restores there.
	if got := r.Image().RGBAAt(20, 55); got != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected restored center pixel to be red, got %+v", got)
	}
	if got := r.Image().RGBAAt(5, 5); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
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
	// bbox and offset are y-up display space. The crop selects the display
	// sub-rect (10,20)-(20,30) of the captured region; the y-up offset (20,20)
	// shifts it right by 20 and up by 20.
	r.RestoreRegion(region, &geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 20, Y: 30},
	}, geom.Pt{X: 20, Y: 20})
	_ = r.End()

	// Display space is y-up: the captured region sits at device rows 60-80. The
	// selected display sub-rect (10,20)-(20,30) is the region's device top-left
	// 10x10 (rows 60-70, cols 10-20). A y-up offset of (20,20) moves it right by
	// 20 (cols 30-40) and up by 20 (device rows decrease to 40-50).
	if got := r.Image().RGBAAt(35, 45); got != (color.RGBA{R: 255, G: 0, B: 0, A: 255}) {
		t.Fatalf("expected partial restored pixel to be red, got %+v", got)
	}
	// The original captured location (device rows 60-70) was overwritten by the
	// blue fill and not restored in place, so it stays blue.
	if got := r.Image().RGBAAt(15, 65); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("expected original region location to remain blue, got %+v", got)
	}
	if got := r.Image().RGBAAt(5, 5); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
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

	// Display space is y-up: the filtered rect at display (0,0)-(30,30) occupies
	// device rows 30-60, composited back at device offset (5,5) -> cols 5-35,
	// rows 35-60.
	if got := r.Image().RGBAAt(15, 45); got != (color.RGBA{R: 0, G: 0, B: 255, A: 255}) {
		t.Fatalf("expected filtered-stop pixel to be blue, got %+v", got)
	}
	if got := r.Image().RGBAAt(2, 2); got != (color.RGBA{R: 0, G: 255, B: 0, A: 255}) {
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

	if got := r.Image().RGBAAt(30, 30); got.R == 0 {
		t.Fatalf("expected filtered path center to contain red, got %+v", got)
	}
	if got := r.Image().RGBAAt(19, 30); got.R == 0 || got.G == 255 {
		t.Fatalf("expected blurred red edge over green background, got %+v", got)
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

	img := r.Image()
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

	got := r.Image().RGBAAt(50, 50)
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

	if got := r.Image().RGBAAt(90, 90); got.B < 200 {
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
	r.DrawImage(render.NewImageData(src), geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 100, Y: 100}})
	_ = r.End()

	img := r.Image()
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
