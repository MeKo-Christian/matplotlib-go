package ps

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := New(200, 100, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

func TestImageEmitsPostScriptColorImage(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 0xff, A: 0xff})
	img.SetRGBA(1, 0, color.RGBA{G: 0xff, A: 0xff})

	r.Image(render.NewImageData(img), geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 50, Y: 40},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("2 1 8 [2 0 0 -1 0 1]")) {
		t.Fatalf("missing colorimage geometry in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("colorimage")) {
		t.Fatalf("missing colorimage operator in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("<ff000000ff00>")) {
		t.Fatalf("missing deterministic RGB image payload in\n%s", r.document)
	}
}

func TestImageTransformedEmitsConcatMatrix(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	transformer, ok := any(r).(render.ImageTransformer)
	if !ok {
		t.Fatal("PS renderer should implement render.ImageTransformer")
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{B: 0xff, A: 0xff})
	img.SetRGBA(1, 0, color.RGBA{R: 0xff, G: 0xff, A: 0xff})

	transformer.ImageTransformed(render.NewImageData(img), geom.Rect{
		Min: geom.Pt{X: 10, Y: 20},
		Max: geom.Pt{X: 12, Y: 21},
	}, geom.Affine{
		A: 2, B: 0.5,
		C: -0.25, D: 3,
		E: 10, F: 20,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("[4 1 -0.25 3 10 20] concat")) {
		t.Fatalf("missing transformed image concat matrix in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("<0000ffffff00>")) {
		t.Fatalf("missing transformed image payload in\n%s", r.document)
	}
}

func TestPathWithHatchEmitsClippedHatchLines(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	hatcher, ok := any(r).(render.NativeHatcher)
	if !ok {
		t.Fatal("PS renderer should implement render.NativeHatcher")
	}
	if !hatcher.SupportsNativeHatch() {
		t.Fatal("SupportsNativeHatch returned false")
	}

	var p geom.Path
	p.MoveTo(geom.Pt{X: 10, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 10})
	p.LineTo(geom.Pt{X: 50, Y: 50})
	p.Close()
	r.Path(p, &render.Paint{
		Fill:           render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1},
		Hatch:          "/",
		HatchColor:     render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
		HatchLineWidth: 1.5,
		HatchSpacing:   8,
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	if !bytes.Contains(r.document, []byte("clip newpath")) {
		t.Fatalf("missing hatch clip in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("0.1 0.2 0.3 setrgbcolor")) {
		t.Fatalf("missing hatch color in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("1.5 setlinewidth")) {
		t.Fatalf("missing hatch linewidth in\n%s", r.document)
	}
	if !bytes.Contains(r.document, []byte("lineto\nstroke")) {
		t.Fatalf("missing hatch stroke lines in\n%s", r.document)
	}
}
