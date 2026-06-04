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

func TestAggImageNearestNonIntegerUpscalePreservesSourcePalette(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 12, 10))
	dark := color.RGBA{R: 96, G: 150, B: 209, A: 255}
	light := color.RGBA{R: 229, G: 239, B: 255, A: 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 12; x++ {
			c := light
			if (x+y)%2 == 0 {
				c = dark
			}
			src.SetRGBA(x, y, c)
		}
	}
	raster := render.NewImageData(src)
	raster.SetInterpolation("nearest")

	r, err := New(64, 64, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 64, Y: 64}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(raster, geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 50, Y: 43.333333333333336}})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	got := r.GetImage()
	allowed := map[color.RGBA]bool{
		dark:                             true,
		light:                            true,
		{R: 255, G: 255, B: 255, A: 255}: true,
	}
	var blended []color.RGBA
	for y := 0; y < got.Bounds().Dy(); y++ {
		for x := 0; x < got.Bounds().Dx(); x++ {
			c := got.RGBAAt(x, y)
			if allowed[c] {
				continue
			}
			blended = append(blended, c)
			if len(blended) >= 5 {
				t.Fatalf("nearest non-integer upscale produced blended colors, first samples=%+v", blended)
			}
		}
	}
}

func TestAggImageNearestNonIntegerUpscaleAlignsTopEdgeLikeMatplotlib(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 14, 14))
	for y := 0; y < 14; y++ {
		for x := 0; x < 14; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	src.SetRGBA(0, 0, color.RGBA{A: 255})
	src.SetRGBA(0, 1, color.RGBA{A: 255})

	raster := render.NewImageData(src)
	raster.SetInterpolation("nearest")

	r, err := New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 640, Y: 360}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Image(raster, geom.Rect{
		Min: geom.Pt{X: 183.2, Y: 50.4},
		Max: geom.Pt{X: 456.8, Y: 324.0},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	got := r.GetImage()
	bounds, _, ok := inkBounds(got, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok {
		t.Fatal("expected rendered black image pixels")
	}
	// Matplotlib's AxesImage _make_image ceils this 273.6 px image to 274 px,
	// then RendererAgg places its top at device y=36 for this exact geometry.
	if bounds.Min.Y != 36 {
		t.Fatalf("nearest image top y = %v, want Matplotlib y=36 (bounds=%v)", bounds.Min.Y, bounds)
	}
}

func TestAggClipRectUsesMatplotlibHalfUpQuantizationForImages(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			src.SetRGBA(x, y, color.RGBA{G: 255, A: 255})
		}
	}

	r, err := New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 640, Y: 360}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.ClipRect(geom.Rect{
		Min: geom.Pt{X: 76.8, Y: 57.6},
		Max: geom.Pt{X: 588.8, Y: 316.8},
	})
	r.Image(render.NewImageData(src), geom.Rect{
		Min: geom.Pt{X: -179.2, Y: 57.6},
		Max: geom.Pt{X: 844.8, Y: 316.8},
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	got := r.GetImage()
	if px := got.RGBAAt(76, 100); px != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("clip column before rounded edge = %+v, want background", px)
	}
	if px := got.RGBAAt(77, 100); px.G < 200 {
		t.Fatalf("clip column at rounded edge = %+v, want image", px)
	}
}

func TestAggBboxImageNearestNonIntegerUpscaleUsesMatplotlibBboxPlacement(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 12, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 12; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	src.SetRGBA(0, 0, color.RGBA{A: 255})

	raster := render.NewImageData(src)
	r, err := New(720, 420, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 720, Y: 420}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if !r.DrawBboxImage(raster, geom.Rect{
		Min: geom.Pt{X: 522.64, Y: 115.63333333333334},
		Max: geom.Pt{X: 562.64, Y: 148.96666666666667},
	}) {
		t.Fatal("DrawBboxImage returned false")
	}
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	got := r.GetImage()
	bounds, _, ok := inkBounds(got, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !ok {
		t.Fatal("expected rendered black image pixels")
	}
	// Matplotlib BboxImage rounds this x and computes top as
	// int(canvasHeight - (bbox.y0 + ceil(imageHeight))).
	if bounds.Min.X != 523 || bounds.Min.Y != 270 {
		t.Fatalf("bbox image origin = %v, want Matplotlib top-left (523,270)", bounds.Min)
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
	// Display space is y-up: an upright placement maps image row 0 to the top
	// edge (display Max.Y), so the affine uses D=-sy with F=dst height. This
	// mirrors the matrix core.imageTransform builds for axis-aligned images.
	r.ImageTransformed(data, geom.Rect{Max: geom.Pt{X: 20, Y: 20}}, geom.Affine{
		A: 10,
		D: -10,
		F: 20,
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

func TestAggTransformedImageRespectsClipPathAndAlpha(t *testing.T) {
	r, err := New(20, 20, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			src.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	data := render.NewImageData(src)
	data.SetAlpha(0.5)

	var clip geom.Path
	clip.MoveTo(geom.Pt{X: 0, Y: 0})
	clip.LineTo(geom.Pt{X: 20, Y: 0})
	clip.LineTo(geom.Pt{X: 0, Y: 20})
	clip.Close()

	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 20, Y: 20}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.ClipPath(clip)
	r.ImageTransformed(data, geom.Rect{Max: geom.Pt{X: 20, Y: 20}}, geom.Affine{
		A: 10,
		D: 10,
	})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	img := r.GetImage()
	// Display space is y-up: the clip triangle (0,0)->(20,0)->(0,20) device-flips
	// to (0,20),(20,20),(0,0), so the kept region is y>=x. Sample inside (y>x) and
	// outside (y<x) accordingly.
	inside := img.RGBAAt(4, 15)
	if inside.R != 255 || math.Abs(float64(inside.G)-128) > 2 || math.Abs(float64(inside.B)-128) > 2 || inside.A != 255 {
		t.Fatalf("clipped transformed image inside pixel = %+v, want half-alpha red over white", inside)
	}
	outside := img.RGBAAt(15, 4)
	if outside != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Fatalf("transformed image escaped clip path: outside pixel = %+v", outside)
	}
}
