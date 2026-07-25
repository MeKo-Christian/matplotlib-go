package core

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

type paddedRGBAOutputRenderer struct {
	render.NullRenderer
	img *image.RGBA
}

func (r *paddedRGBAOutputRenderer) Image() *image.RGBA {
	return r.img
}

func TestFigureImageDetachesPaddedRGBAExporter(t *testing.T) {
	var source *image.RGBA
	factory := func(width, height int, _ render.Color) (render.Renderer, error) {
		parent := image.NewRGBA(image.Rect(0, 0, width+1, height))
		source = parent.SubImage(image.Rect(1, 0, width+1, height)).(*image.RGBA)
		source.SetRGBA(1, 0, color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff})
		source.SetRGBA(2, 0, color.RGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff})
		source.SetRGBA(1, 1, color.RGBA{R: 0x77, G: 0x88, B: 0x99, A: 0xff})
		source.SetRGBA(2, 1, color.RGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff})
		return &paddedRGBAOutputRenderer{img: source}, nil
	}

	figureOutputRenderers.Lock()
	previous, hadPrevious := figureOutputRenderers.factories[".png"]
	figureOutputRenderers.factories[".png"] = factory
	figureOutputRenderers.Unlock()
	t.Cleanup(func() {
		figureOutputRenderers.Lock()
		defer figureOutputRenderers.Unlock()
		if hadPrevious {
			figureOutputRenderers.factories[".png"] = previous
		} else {
			delete(figureOutputRenderers.factories, ".png")
		}
	})

	got, err := NewFigure(2, 2).Image()
	if err != nil {
		t.Fatalf("Image: %v", err)
	}
	want := map[image.Point]color.RGBA{
		{X: 1, Y: 0}: {R: 0x11, G: 0x22, B: 0x33, A: 0xff},
		{X: 2, Y: 0}: {R: 0x44, G: 0x55, B: 0x66, A: 0xff},
		{X: 1, Y: 1}: {R: 0x77, G: 0x88, B: 0x99, A: 0xff},
		{X: 2, Y: 1}: {R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
	}
	if got.Bounds() != source.Bounds() {
		t.Fatalf("bounds = %v, want %v", got.Bounds(), source.Bounds())
	}
	for point, wantPixel := range want {
		if gotPixel := got.RGBAAt(point.X, point.Y); gotPixel != wantPixel {
			t.Errorf("RGBAAt(%v) = %#v, want %#v", point, gotPixel, wantPixel)
		}
	}

	got.SetRGBA(1, 0, color.RGBA{})
	if source.RGBAAt(1, 0) == (color.RGBA{}) {
		t.Fatal("mutating returned image affected the renderer image")
	}
}
