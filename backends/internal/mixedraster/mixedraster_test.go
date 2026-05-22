package mixedraster

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestStartScalesRasterSurfaceByDPI(t *testing.T) {
	session, ok := Start(
		20,
		10,
		geom.Rect{Max: geom.Pt{X: 20, Y: 10}},
		render.Rasterization{Mode: render.RasterizeAlways, DPI: 144},
		72,
		nil,
		nil,
	)
	if !ok {
		t.Fatal("Start failed")
	}

	var path geom.Path
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 20, Y: 0})
	path.LineTo(geom.Pt{X: 20, Y: 10})
	path.LineTo(geom.Pt{X: 0, Y: 10})
	path.Close()
	session.Renderer().Path(path, &render.Paint{Fill: render.Color{R: 1, A: 1}})

	img, rect, ok := session.Stop()
	if !ok {
		t.Fatal("Stop failed")
	}
	if gotW, gotH := img.Size(); gotW != 40 || gotH != 20 {
		t.Fatalf("raster image size = %dx%d, want 40x20", gotW, gotH)
	}
	if rect.W() != 20 || rect.H() != 10 {
		t.Fatalf("placement rect = %v, want original 20x10 display units", rect)
	}
	alpha := img.RGBA().RGBAAt(30, 10).A
	if alpha == 0 {
		t.Fatal("scaled path did not fill high-DPI right half")
	}
}

func TestScaledRendererTransformsImagesAtRasterDPI(t *testing.T) {
	session, ok := Start(
		20,
		20,
		geom.Rect{Max: geom.Pt{X: 20, Y: 20}},
		render.Rasterization{Mode: render.RasterizeAlways, DPI: 144},
		72,
		nil,
		nil,
	)
	if !ok {
		t.Fatal("Start failed")
	}

	transformer, ok := session.Renderer().(render.ImageTransformer)
	if !ok {
		t.Fatal("mixed-raster replay renderer does not support transformed images")
	}

	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	transformer.ImageTransformed(render.NewImageData(src), geom.Rect{}, geom.Affine{
		A: 5,
		D: 5,
		E: 4,
		F: 6,
	})

	img, _, ok := session.Stop()
	if !ok {
		t.Fatal("Stop failed")
	}

	got := img.RGBA()
	if got.Bounds().Dx() != 40 || got.Bounds().Dy() != 40 {
		t.Fatalf("raster image size = %dx%d, want 40x40", got.Bounds().Dx(), got.Bounds().Dy())
	}
	if c := got.RGBAAt(9, 13); c.A == 0 || c.R < 200 {
		t.Fatalf("transformed image was not scaled into high-DPI raster surface; pixel = %+v", c)
	}
	if c := got.RGBAAt(4, 6); c.A != 0 {
		t.Fatalf("unscaled transform unexpectedly painted low-DPI position; pixel = %+v", c)
	}
}
