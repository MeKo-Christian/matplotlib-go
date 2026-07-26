package pdf

import (
	"bytes"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRasterizedArtistEmbedsImageAndKeepsVectorContent(t *testing.T) {
	fig := core.NewFigure(200, 100)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.2, Y: 0.2}, Max: geom.Pt{X: 0.8, Y: 0.75}})
	line, _ := ax.Plot([]float64{0, 0.5, 1}, []float64{0, 1, 0}, core.PlotOptions{})
	line.SetRasterized(true)
	ax.SetTitle("Vector title")

	r := newTestRenderer(t)
	core.DrawFigure(fig, r)

	if !strings.Contains(r.content.String(), "/Im1 Do") {
		t.Fatalf("rasterized artist did not invoke a PDF image XObject:\n%s", r.content.String())
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.Contains(data, []byte("/Subtype /Image")) {
		t.Fatalf("rasterized artist did not serialize a PDF image XObject:\n%s", data)
	}
	if strings.Count(r.content.String(), " m\n") < 2 {
		t.Fatalf("surrounding vector path content was not preserved:\n%s", r.content.String())
	}
}

func TestRasterizedGroupReplaysActivePathClip(t *testing.T) {
	r := newTestRenderer(t)
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 20, Y: 20}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 10, Y: 0})
	clip.LineTo(geom.Pt{X: 10, Y: 20})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()
	r.ClipPath(clip)

	if !r.StartRasterized(render.Rasterization{Mode: render.RasterizeAlways, DPI: 72}) {
		t.Fatal("StartRasterized failed")
	}
	var fill geom.Path
	fill.MoveTo(geom.Pt{X: 0, Y: 0})
	fill.LineTo(geom.Pt{X: 20, Y: 0})
	fill.LineTo(geom.Pt{X: 20, Y: 20})
	fill.LineTo(geom.Pt{X: 0, Y: 20})
	fill.Close()
	r.Path(fill, &render.Paint{Fill: render.Color{R: 1, A: 1}})
	if !r.StopRasterized() {
		t.Fatal("StopRasterized failed")
	}

	if len(r.images) != 1 {
		t.Fatalf("embedded images = %d, want 1", len(r.images))
	}
	img := r.images[0]
	if !img.hasAlpha {
		t.Fatal("rasterized image should carry an alpha mask")
	}
	// Display space is y-up: the rasterized surface is the full 200x100 page, so
	// content drawn at display y in [0,20] (bottom-left) lands at device rows
	// [80,100). Sample a row inside that band; the clip keeps cols [0,10].
	left := img.alpha[90*img.width+5]
	right := img.alpha[90*img.width+15]
	if left == 0 {
		t.Fatalf("pixel inside path clip is transparent: alpha=%d", left)
	}
	if right != 0 {
		t.Fatalf("pixel outside path clip alpha=%d, want 0", right)
	}
}

func TestImageEmitsXObjectResourceAndDrawOperator(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{B: 255, A: 255})

	r.DrawImage(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	raw := r.content.String()
	if !strings.Contains(raw, "/Im1 Do") {
		t.Fatalf("expected image draw operator in content stream, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/XObject << /Im1", "/Subtype /Image", "/ColorSpace /DeviceRGB", "/Filter /FlateDecode"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestFlateImageEmitsPNGDecodeParms(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})

	r.DrawImage(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	body := pdfDocumentObjectBodyContaining(doc, "/Subtype /Image")
	for _, want := range []string{
		"/Filter /FlateDecode",
		"/DecodeParms << /Predictor 10 /Colors 3 /Columns 2 /BitsPerComponent 8 >>",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("image object missing %q:\n%s", want, body)
		}
	}
}

func TestJPEGImageEmitsDCTDecodeXObject(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := jpegTestImage{
		w:    2,
		h:    1,
		data: []byte{0xff, 0xd8, 0xff, 0xd9},
	}

	r.DrawImage(img, geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 50, Y: 40}})

	if !strings.Contains(r.content.String(), "/Im1 Do") {
		t.Fatalf("expected JPEG image draw operator, got %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/Subtype /Image", "/ColorSpace /DeviceRGB", "/Filter /DCTDecode"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestImageReusesXObjectForRepeatedImageData(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	img.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	data := render.NewImageData(img)

	r.DrawImage(data, geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 40}})
	r.DrawImage(data, geom.Rect{Min: geom.Pt{X: 40, Y: 20}, Max: geom.Pt{X: 60, Y: 40}})

	if got := strings.Count(r.content.String(), "/Im1 Do"); got != 2 {
		t.Fatalf("expected both draws to invoke reused Im1 XObject, got %d in %q", got, r.content.String())
	}
	if strings.Contains(r.content.String(), "/Im2 Do") {
		t.Fatalf("did not expect duplicate image XObject invocation in %q", r.content.String())
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if got := pdfDocumentObjectCountContaining(doc, "/Subtype /Image"); got != 1 {
		t.Fatalf("expected one image XObject for repeated image data, got %d; objects: %#v", got, doc.Objects)
	}
}

func TestImageWithAlphaEmitsSoftMask(t *testing.T) {
	r := newTestRenderer(t)
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.SetRGBA(0, 0, color.RGBA{R: 20, G: 40, B: 60, A: 128})

	r.DrawImage(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 40}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	data, err := r.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	for _, want := range []string{"/SMask", "/ColorSpace /DeviceGray"} {
		if !bytes.Contains(data, []byte(want)) {
			t.Fatalf("serialized PDF missing %q:\n%s", want, data)
		}
	}
}

func TestImageTransformedEmitsAffineImageMatrix(t *testing.T) {
	r := newTestRenderer(t)
	if _, ok := any(r).(render.ImageTransformer); !ok {
		t.Fatal("PDF renderer should implement render.ImageTransformer")
	}
	_ = r.Begin(geom.Rect{Max: geom.Pt{X: 200, Y: 100}})
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})

	r.ImageTransformed(render.NewImageData(img), geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 30, Y: 50}}, geom.Affine{
		A: 2,
		B: 0.5,
		C: -1,
		D: 3,
		E: 7,
		F: 11,
	})

	raw := r.content.String()
	if !strings.Contains(raw, "4 1 -3 9 7 11 cm") || !strings.Contains(raw, "/Im1 Do") {
		t.Fatalf("expected transformed image matrix and XObject invocation, got %q", raw)
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}
	doc := mustParsePDF(t, r)
	if !pdfDocumentBodyContains(doc, "/XObject << /Im1") {
		t.Fatalf("page resources should reference transformed image Im1; objects: %#v", doc.Objects)
	}
}
