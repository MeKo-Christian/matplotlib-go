package core

import (
	"image"
	"image/color"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func rasterizedRGBA(t *testing.T, img *Image2D) *image.RGBA {
	t.Helper()
	rendered, ok := img.rasterize()
	if !ok {
		t.Fatal("expected rasterization to succeed")
	}
	data, ok := rendered.(*render.ImageData)
	if !ok {
		t.Fatalf("expected render.ImageData, got %T", rendered)
	}
	pix := data.RGBA()
	if pix == nil {
		t.Fatal("expected RGBA pixels")
	}
	return pix
}

func TestNormalizeRGBArrayClassifiesChannels(t *testing.T) {
	cases := []struct {
		name     string
		data     [][][]float64
		wantKind rgbArrayKind
	}{
		{"rgb", [][][]float64{{{1, 0, 0}, {0, 1, 0}}}, rgbArrayRGB},
		{"rgba", [][][]float64{{{1, 0, 0, 0.5}}}, rgbArrayRGBA},
		{"scalar", [][][]float64{{{0.2}, {0.4}}}, rgbArrayScalar},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, kind, err := normalizeRGBArray(tc.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if kind != tc.wantKind {
				t.Fatalf("kind = %v, want %v", kind, tc.wantKind)
			}
		})
	}
}

func TestNormalizeRGBArrayRejectsBadShapes(t *testing.T) {
	cases := map[string][][][]float64{
		"empty":         {},
		"two-channel":   {{{0.1, 0.2}}},
		"ragged-rows":   {{{1, 0, 0}, {0, 1, 0}}, {{0, 0, 1}}},
		"ragged-pixels": {{{1, 0, 0}, {0, 1}}},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := normalizeRGBArray(data); err == nil {
				t.Fatalf("expected error for %s", name)
			}
		})
	}
}

func TestNormalizeRGBArrayClipsAndConverts(t *testing.T) {
	// Values outside [0,1] are clipped; in-range floats round to bytes.
	img, kind, err := normalizeRGBArray([][][]float64{{{1.5, -0.2, 0.5}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != rgbArrayRGB {
		t.Fatalf("kind = %v, want RGB", kind)
	}
	got := img.RGBAAt(0, 0)
	// 0.5*255 = 127.5 truncates to 127 (matplotlib's astype(uint8)).
	want := color.RGBA{R: 255, G: 0, B: 127, A: 255}
	if got != want {
		t.Fatalf("pixel = %+v, want %+v", got, want)
	}
}

func TestNormalizeRGBArrayPreservesStraightAlpha(t *testing.T) {
	// Alpha is stored straight (not premultiplied): the backend premultiplies.
	img, _, err := normalizeRGBArray([][][]float64{{{1, 1, 1, 0.5}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := img.RGBAAt(0, 0)
	if got.R != 255 || got.G != 255 || got.B != 255 {
		t.Fatalf("expected straight (un-premultiplied) RGB 255, got %+v", got)
	}
	if got.A != 127 {
		t.Fatalf("expected alpha 127 (0.5*255 truncated), got %d", got.A)
	}
}

func TestImShowRGBBypassesColormap(t *testing.T) {
	ax := &Axes{}
	img := ax.ImShowRGB([][][]float64{
		{{1, 0, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 1, 0}},
	})
	if img == nil {
		t.Fatal("expected image artist")
	}
	// True-color images carry no scalar mapping (no colorbar).
	if sm := img.ScalarMap(); sm.Colormap != "" || sm.Norm != nil {
		t.Fatalf("expected empty ScalarMap for true-color image, got %+v", sm)
	}
	// Default extent is centered on pixel grid like matplotlib imshow.
	if img.XMin != -0.5 || img.XMax != 1.5 || img.YMin != -0.5 || img.YMax != 1.5 {
		t.Fatalf("unexpected extent %v..%v / %v..%v", img.XMin, img.XMax, img.YMin, img.YMax)
	}

	pix := rasterizedRGBA(t, img)
	// Origin upper: array row 0 stays at the top row of the bitmap.
	if got := pix.RGBAAt(0, 0); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("top-left = %+v, want red", got)
	}
	if got := pix.RGBAAt(0, 1); got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("bottom-left = %+v, want blue", got)
	}
}

func TestImShowRGBOriginLowerFlips(t *testing.T) {
	ax := &Axes{}
	img := ax.ImShowRGB([][][]float64{
		{{1, 0, 0}},
		{{0, 0, 1}},
	}, ImShowRGBOptions{Origin: ImageOriginLower})
	if img == nil {
		t.Fatal("expected image artist")
	}
	pix := rasterizedRGBA(t, img)
	// Origin lower: array row 0 (red) moves to the bottom row of the bitmap.
	if got := pix.RGBAAt(0, 1); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("bottom row = %+v, want red (row 0)", got)
	}
	if got := pix.RGBAAt(0, 0); got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("top row = %+v, want blue (row 1)", got)
	}
}

func TestImShowRGBScalarSqueezeRoutesToColormap(t *testing.T) {
	ax := &Axes{}
	img := ax.ImShowRGB([][][]float64{
		{{0}, {1}},
	})
	if img == nil {
		t.Fatal("expected image artist")
	}
	// Squeezed (M,N,1) goes through the scalar colormap path.
	if img.rgba != nil {
		t.Fatal("expected scalar-backed image, got true-color")
	}
	if sm := img.ScalarMap(); sm.Colormap == "" {
		t.Fatal("expected scalar colormap mapping for squeezed (M,N,1) array")
	}
}

func TestImShowImageWrapsNativeImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	src.SetRGBA(1, 0, color.RGBA{R: 40, G: 50, B: 60, A: 255})

	ax := &Axes{}
	img := ax.ImShowImage(src)
	if img == nil {
		t.Fatal("expected image artist")
	}
	if img.rgba == nil {
		t.Fatal("expected true-color backing")
	}
	pix := rasterizedRGBA(t, img)
	if got := pix.RGBAAt(0, 0); got != (color.RGBA{R: 10, G: 20, B: 30, A: 255}) {
		t.Fatalf("pixel(0,0) = %+v", got)
	}
}

func TestImShowImageScalarAlphaApplies(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 1, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	ax := &Axes{}
	alpha := 0.5
	img := ax.ImShowImage(src, ImShowRGBOptions{Alpha: &alpha})
	if img == nil {
		t.Fatal("expected image artist")
	}
	rendered, _ := img.rasterize()
	data := rendered.(*render.ImageData)
	if got := data.Alpha(); got != 0.5 {
		t.Fatalf("image alpha = %v, want 0.5", got)
	}
}
