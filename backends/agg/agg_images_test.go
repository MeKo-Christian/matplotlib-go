package agg

import (
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestGetImage(t *testing.T) {
	r := mustNew(t, 200, 150)
	img := r.Image()
	if img == nil {
		t.Fatal("Image returned nil")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 150 {
		t.Errorf("unexpected image dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageViewSharesBackingWhileGetImageCopies(t *testing.T) {
	r := mustNew(t, 4, 3)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 4, Y: 3}}); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.Path(aggTestRectPath(0, 0, 1, 1), &render.Paint{Fill: render.Color{R: 1, A: 1}})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	view := r.ImageView()
	copyImg := r.Image()
	if view == nil {
		t.Fatal("ImageView returned nil")
	}
	if len(view.Pix) == 0 || len(copyImg.Pix) == 0 {
		t.Fatal("expected non-empty images")
	}
	if &view.Pix[0] != &r.ctx.image.Data[0] {
		t.Fatal("ImageView does not share AGG backing storage")
	}
	if &copyImg.Pix[0] == &r.ctx.image.Data[0] {
		t.Fatal("Image returned shared storage; want owned copy")
	}

	view.Pix[0] = 123
	if got := r.ctx.image.Data[0]; got != 123 {
		t.Fatalf("ImageView mutation did not touch renderer storage, got %d", got)
	}
	if copyImg.Pix[0] == 123 {
		t.Fatal("Image copy changed after mutating ImageView")
	}
}

func TestClearResetsPixelsAndClipForRendererReuse(t *testing.T) {
	r := mustNew(t, 20, 20)
	vp := geom.Rect{Max: geom.Pt{X: 20, Y: 20}}
	if err := r.Begin(vp); err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 5, Y: 5}})
	r.Path(aggTestRectPath(0, 0, 20, 20), &render.Paint{Fill: render.Color{R: 1, A: 1}})
	if err := r.End(); err != nil {
		t.Fatalf("End failed: %v", err)
	}
	if got := r.Image().RGBAAt(10, 10); got != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("initial clipped draw touched unclipped pixel: %+v", got)
	}

	r.Clear(render.Color{G: 1, A: 1})
	if got := r.Image().RGBAAt(0, 0); got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("Clear pixel = %+v, want green", got)
	}

	if err := r.Begin(vp); err != nil {
		t.Fatalf("second Begin failed: %v", err)
	}
	r.Path(aggTestRectPath(0, 0, 20, 20), &render.Paint{Fill: render.Color{B: 1, A: 1}})
	if err := r.End(); err != nil {
		t.Fatalf("second End failed: %v", err)
	}
	if got := r.Image().RGBAAt(10, 10); got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("reuse draw pixel = %+v, want blue after clip reset", got)
	}
}

func TestRendererInterface(t *testing.T) {
	var _ render.Renderer = (*Renderer)(nil)
}
