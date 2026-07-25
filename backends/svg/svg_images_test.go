package svg

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestImageTransformedEmitsMatrixAttribute(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))
		img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		// Affine: scale 1.5 in x, rotate-ish via b=0.5, translate (10, 20).
		r.ImageTransformed(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 2, Y: 2},
		}, geom.Affine{A: 1.5, B: 0.5, C: -0.25, D: 1, E: 10, F: 20})
	})

	for _, want := range []string{
		`<image x="0" y="0" width="2" height="2"`,
		`href="data:image/png;base64,`,
		`transform="matrix(1.5 -0.5 -0.25 -1 10 100)"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized transformed-image attribute %q in %q", want, content)
		}
	}
}

func TestImageTransformedHonorsClip(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ClipRect(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 50, Y: 50}})
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		r.ImageTransformed(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		}, geom.Identity())
	})

	if !strings.Contains(content, `<g clip-path="url(#clip1)"><image`) {
		t.Fatalf("transformed image should respect active clip, got %q", content)
	}
}

func TestImageTransformedSkipsUnsupportedImage(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.ImageTransformed(sizeOnlyImage{w: 10, h: 10}, geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		}, geom.Identity())
	})

	if strings.Contains(content, "<image") {
		t.Fatalf("unsupported image should not emit image node, got %q", content)
	}
}

func TestMatrixTransformFormat(t *testing.T) {
	got := matrixTransform(geom.Affine{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0})
	if got != "matrix(1 0 0 1 0 0)" {
		t.Errorf("identity matrix transform = %q, want %q", got, "matrix(1 0 0 1 0 0)")
	}

	got = matrixTransform(geom.Affine{A: 0.5, B: 0.25, C: -0.5, D: 1.25, E: 10, F: -20})
	if got != "matrix(0.5 0.25 -0.5 1.25 10 -20)" {
		t.Errorf("matrix transform = %q, want compact form", got)
	}
}

func TestImageSerializesEmbeddedPNGAndNormalizesDestinationRect(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		r.DrawImage(render.NewImageData(img), geom.Rect{
			Min: geom.Pt{X: 30, Y: 40},
			Max: geom.Pt{X: 10, Y: 20},
		})
	})

	for _, want := range []string{
		`<image x="10" y="80" width="20" height="20"`,
		`preserveAspectRatio="none"`,
		`href="data:image/png;base64,`,
		`xlink:href="data:image/png;base64,`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected serialized image attribute %q in %q", want, content)
		}
	}
}

func TestImageSkipsUnsupportedImageAndDegenerateRect(t *testing.T) {
	content := renderSVGDocument(t, func(r *Renderer) {
		r.DrawImage(sizeOnlyImage{w: 10, h: 10}, geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 10, Y: 10},
		})
		r.DrawImage(render.NewImageData(image.NewRGBA(image.Rect(0, 0, 1, 1))), geom.Rect{
			Min: geom.Pt{X: 10, Y: 10},
			Max: geom.Pt{X: 10, Y: 20},
		})
	})

	if strings.Contains(content, "<image") {
		t.Fatalf("unsupported images and degenerate rects should not emit image nodes, got %q", content)
	}
}

func TestHelperImageBranches(t *testing.T) {
	if _, err := encodeImage(nil); err == nil {
		t.Fatal("encoding nil image should fail")
	}
	if got := asRGBAImage(sizeOnlyImage{w: 3, h: 4}); got != nil {
		t.Fatalf("non-RGBA image should not convert, got %#v", got)
	}

	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	converted := asRGBAImage(render.NewImageData(img))
	if converted == nil || converted.Bounds() != img.Bounds() {
		t.Fatalf("expected RGBA image conversion to preserve bounds, got %#v", converted)
	}
}
