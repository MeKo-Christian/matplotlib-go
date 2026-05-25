package agg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// renderUpscaledImage builds a tiny 2x2 checkerboard, hands it to an AGG
// renderer with the given interpolation name, and returns the PNG bytes.
func renderUpscaledImage(t *testing.T, interp string, dstW, dstH int) []byte {
	t.Helper()
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	src.Set(0, 0, black)
	src.Set(1, 1, black)
	// Other pixels stay zero (transparent).

	raster := render.NewImageData(src)
	raster.SetInterpolation(interp)

	r, err := backends.Create(backends.AGG, backends.Config{Width: 64, Height: 64, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := r.Begin(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 64, Y: 64}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(raster, geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: float64(dstW), Y: float64(dstH)}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	exporter, ok := r.(render.PNGExporter)
	if !ok {
		t.Fatal("AGG renderer should implement render.PNGExporter")
	}
	tmp := t.TempDir() + "/out.png"
	if err := exporter.SavePNG(tmp); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PNG output is empty")
	}
	return data
}

func TestAggImage_NearestVsBilinear_DifferentBytes(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 64, 64)
	pngBilinear := renderUpscaledImage(t, "bilinear", 64, 64)
	if bytes.Equal(pngNearest, pngBilinear) {
		t.Fatal("expected different PNG bytes between nearest and bilinear; interpolation appears to be ignored")
	}
}

func TestAggImage_EmptyInterpolationMatchesNearest(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 64, 64)
	pngEmpty := renderUpscaledImage(t, "", 64, 64)

	decoded := func(data []byte) *image.RGBA {
		t.Helper()
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("png.Decode: %v", err)
		}
		if r, ok := img.(*image.RGBA); ok {
			return r
		}
		b := img.Bounds()
		rgba := image.NewRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				rgba.Set(x, y, img.At(x, y))
			}
		}
		return rgba
	}

	a := decoded(pngNearest)
	b := decoded(pngEmpty)
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("empty Interpolation should produce the same pixels as 'nearest'")
	}
}

func TestAggImage_AutoInterpolationMatchesNearestForIntegerScale(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 4, 4)
	pngAuto := renderUpscaledImage(t, "auto", 4, 4)
	if !bytes.Equal(pngNearest, pngAuto) {
		t.Fatal("auto should resolve to nearest on integer-scale transforms")
	}
}

func TestAggImage_AutoInterpolationUsesHanningForNonIntegerScale(t *testing.T) {
	pngNearest := renderUpscaledImage(t, "nearest", 3, 3)
	pngAuto := renderUpscaledImage(t, "auto", 3, 3)
	if bytes.Equal(pngNearest, pngAuto) {
		t.Fatal("auto should prefer hanning for non-integer small upscales")
	}
}

func TestAggImage_AllMatplotlibInterpolationNamesRender(t *testing.T) {
	names := []string{
		"nearest", "none", "bilinear", "bicubic", "hanning",
		"hamming", "lanczos", "spline16", "spline36", "kaiser",
		"quadric", "catrom", "gaussian", "bessel", "mitchell",
		"sinc", "blackman", "hermite", "antialiased", "auto",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			pngData := renderUpscaledImage(t, name, 17, 19)
			if _, err := png.Decode(bytes.NewReader(pngData)); err != nil {
				t.Fatalf("decode rendered PNG for %q: %v", name, err)
			}
		})
	}
}

func TestAggImageExactSizeDrawPreservesBottomAndRightEdges(t *testing.T) {
	r, err := New(12, 12, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	black := color.RGBA{A: 255}
	for x := 0; x < 4; x++ {
		src.SetRGBA(x, 3, black)
	}
	for y := 0; y < 4; y++ {
		src.SetRGBA(3, y, black)
	}
	data := render.NewImageData(src)
	data.SetInterpolation("nearest")

	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 12, Y: 12}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(data, geom.Rect{Min: geom.Pt{X: 4, Y: 4}, Max: geom.Pt{X: 8, Y: 8}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	img := r.GetImage()
	if got := img.RGBAAt(5, 7); got != black {
		t.Fatalf("bottom edge pixel = %+v, want %+v", got, black)
	}
	if got := img.RGBAAt(7, 5); got != black {
		t.Fatalf("right edge pixel = %+v, want %+v", got, black)
	}
}

func TestImageTransformDisplaySpan(t *testing.T) {
	raster := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 2, 3)))
	spanX, spanY := imageTransformDisplaySpan(raster, geom.Affine{
		A: 2,
		B: 1,
		C: 1,
		D: 2,
		E: 0,
		F: 0,
	})
	if spanX != 7 || spanY != 8 {
		t.Fatalf("imageTransformDisplaySpan = (%g, %g), want (7, 8)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(render.NewImageData(image.NewRGBA(image.Rect(0, 0, 0, 0))), geom.Affine{A: 2})
	if spanX != 0 || spanY != 0 {
		t.Fatalf("empty image span should be zero, got (%g, %g)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(raster, geom.Affine{
		A: 1,
		B: 0,
		C: -1,
		D: 1,
	})
	if spanX != 5 || spanY != 3 {
		t.Fatalf("imageTransformDisplaySpan with opposing sign axes should be (5,3), got (%g, %g)", spanX, spanY)
	}

	spanX, spanY = imageTransformDisplaySpan(raster, geom.Affine{
		A: 0,
		B: 1,
		C: -1,
		D: 0,
		E: 0,
		F: 0,
	})
	if spanX != 3 || spanY != 2 {
		t.Fatalf("imageTransformDisplaySpan with 90° rotation should be (3,2), got (%g, %g)", spanX, spanY)
	}
}

func TestAggImageRespectsImageAlphaState(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	r, err := backends.Create(backends.AGG, backends.Config{Width: 10, Height: 10, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(data, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 10, Y: 10}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	aggR, ok := r.(*Renderer)
	if !ok {
		t.Fatalf("expected *agg.Renderer, got %T", r)
	}

	got := aggR.GetImage().RGBAAt(0, 0)
	if got.A != 255 {
		t.Fatalf("composited alpha = %d, want 255", got.A)
	}
	if got.R != 255 || math.Abs(float64(got.G)-128) > 2 || math.Abs(float64(got.B)-128) > 2 {
		t.Fatalf("expected red with 0.5 image alpha, got %+v", got)
	}
}

func TestAggImageAlphaPremultipliesSourceRGB(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.Set(0, 0, color.RGBA{R: 80, G: 120, B: 200, A: 255})
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	r, err := backends.Create(backends.AGG, backends.Config{Width: 4, Height: 4, DPI: 72})
	if err != nil {
		t.Fatalf("Create AGG: %v", err)
	}
	if err := r.Begin(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 4, Y: 4}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(data, geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 4, Y: 4}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	aggR, ok := r.(*Renderer)
	if !ok {
		t.Fatalf("expected *agg.Renderer, got %T", r)
	}
	got := aggR.GetImage().RGBAAt(0, 0)
	if math.Abs(float64(got.R)-168) > 2 || math.Abs(float64(got.G)-188) > 2 || math.Abs(float64(got.B)-228) > 2 {
		t.Fatalf("alpha-composited color = %+v, want approx {168 188 228 255}", got)
	}
}

func TestAggImageAlphaPremultipliesConvertedBuffer(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 80, G: 120, B: 200, A: 200})
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	converted, ok := renderImageToAGG(data)
	if !ok {
		t.Fatal("renderImageToAGG returned false")
	}
	rgba := converted.ToGoImage()
	got := rgba.RGBAAt(0, 0)
	if got.R != 31 || got.G != 47 || got.B != 78 {
		t.Fatalf("image alpha should premultiply RGB channels, got %+v", got)
	}
	if got.A != 100 {
		t.Fatalf("image alpha should scale source alpha, got %d want 100", got.A)
	}
	if src.RGBAAt(0, 0).A != 200 {
		t.Fatalf("renderImageToAGG mutated source alpha, got %d", src.RGBAAt(0, 0).A)
	}
}

func TestAggGetImageBufferIsRGBANotARGB(t *testing.T) {
	r, err := New(2, 1, render.Color{
		R: float64(0x12) / 255,
		G: float64(0x34) / 255,
		B: float64(0x56) / 255,
		A: 1,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got := r.GetImage()
	want := []uint8{0x12, 0x34, 0x56, 0xff}
	if !bytes.Equal(got.Pix[:4], want) {
		t.Fatalf("buffer bytes = %#v, want RGBA %#v", got.Pix[:4], want)
	}
}

func TestAggTransparentBackgroundRemainsTransparent(t *testing.T) {
	r, err := New(3, 2, render.Color{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	img := r.GetImage()
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			if got := img.RGBAAt(x, y); got != (color.RGBA{}) {
				t.Fatalf("pixel (%d,%d) = %+v, want transparent black", x, y, got)
			}
		}
	}
}

func TestAggSavePNGRoundTripsGetImageRGBA(t *testing.T) {
	r, err := New(2, 2, render.Color{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 2, Y: 2}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(render.NewImageData(src), geom.Rect{Max: geom.Pt{X: 2, Y: 2}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	path := t.TempDir() + "/out.png"
	if err := r.SavePNG(path); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open saved PNG: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("Decode saved PNG: %v", err)
	}

	got := r.GetImage()
	for y := 0; y < got.Bounds().Dy(); y++ {
		for x := 0; x < got.Bounds().Dx(); x++ {
			if decoded.At(x, y) != got.At(x, y) {
				t.Fatalf("decoded pixel (%d,%d) = %v, want %v", x, y, decoded.At(x, y), got.At(x, y))
			}
		}
	}
}

func TestAggTransformedImagePreservesSourceOrientation(t *testing.T) {
	r, err := New(20, 20, render.Color{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, A: 255})
	data := render.NewImageData(src)

	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 20, Y: 20}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.ImageTransformed(data, geom.Rect{Max: geom.Pt{X: 20, Y: 20}}, geom.Affine{
		A: 10,
		D: 10,
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	img := r.GetImage()
	samples := []struct {
		name string
		x, y int
		want color.RGBA
	}{
		{"top-left", 2, 2, color.RGBA{R: 255, A: 255}},
		{"top-right", 12, 2, color.RGBA{G: 255, A: 255}},
		{"bottom-left", 2, 12, color.RGBA{B: 255, A: 255}},
		{"bottom-right", 12, 12, color.RGBA{R: 255, G: 255, A: 255}},
	}
	for _, sample := range samples {
		if got := img.RGBAAt(sample.x, sample.y); got != sample.want {
			t.Fatalf("%s pixel = %+v, want %+v", sample.name, got, sample.want)
		}
	}
}
