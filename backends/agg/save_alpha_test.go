package agg

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestAggSavePNGSemiTransparentOverTransparent guards PLAN.md D5: a
// semi-transparent fill over a transparent surface must save to PNG with the
// original straight-alpha RGB, only the alpha reduced. The AGG buffer holds
// straight (non-premultiplied) alpha; SavePNG must not let png.Encode apply
// premultiplied->straight conversion (which overflows when value>alpha and
// corrupts the color).
func TestAggSavePNGSemiTransparentOverTransparent(t *testing.T) {
	r, err := New(2, 2, render.Color{}) // fully transparent background
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// (200,120,40) @ a=0.5 -> straight RGBA (200,120,40,128).
	fill := render.Color{R: 200.0 / 255, G: 120.0 / 255, B: 40.0 / 255, A: 0.5}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 2, Y: 2}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Path(fullRectPath(2, 2), &render.Paint{Fill: fill})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	path := t.TempDir() + "/out.png"
	if err := r.SavePNG(path); err != nil {
		t.Fatalf("SavePNG: %v", err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	nr := image.NewNRGBA(decoded.Bounds())
	draw.Draw(nr, nr.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	got := nr.NRGBAAt(1, 1)

	want := color.NRGBA{R: 200, G: 120, B: 40, A: 128}
	const tol = 2
	if absDiff8(got.R, want.R) > tol || absDiff8(got.G, want.G) > tol ||
		absDiff8(got.B, want.B) > tol || absDiff8(got.A, want.A) > tol {
		t.Fatalf("saved PNG straight NRGBA = %+v, want approx %+v", got, want)
	}

	// The codebase's round-trip invariant must also hold for alpha<255:
	// decoded straight pixel equals the (straight) in-memory GetImage pixel.
	mem := r.GetImage().RGBAAt(1, 1)
	if absDiff8(got.R, mem.R) > tol || absDiff8(got.G, mem.G) > tol ||
		absDiff8(got.B, mem.B) > tol || absDiff8(got.A, mem.A) > tol {
		t.Fatalf("decoded PNG %+v != in-memory GetImage %+v", got, mem)
	}
}

// TestAggFillOverWhiteTransparentClear isolates the report-figure residual:
// drawing a semi-transparent fill over a (255,255,255,alpha=0) clear must yield
// the pure straight color, not the color matted toward white.
func TestAggFillOverWhiteTransparentClear(t *testing.T) {
	for _, bg := range []render.Color{
		{R: 0, G: 0, B: 0, A: 0}, // black-transparent
		{R: 1, G: 1, B: 1, A: 0}, // white-transparent (report config)
	} {
		r, err := New(2, 2, bg)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		fill := render.Color{R: 200.0 / 255, G: 120.0 / 255, B: 40.0 / 255, A: 0.5}
		if err := r.Begin(geom.Rect{Max: geom.Pt{X: 2, Y: 2}}); err != nil {
			t.Fatalf("Begin: %v", err)
		}
		r.Path(fullRectPath(2, 2), &render.Paint{Fill: fill})
		if err := r.End(); err != nil {
			t.Fatalf("End: %v", err)
		}
		c := r.GetImage().RGBAAt(1, 1)
		// A fully transparent clear (alpha 0) must not matte its RGB into the
		// fill, regardless of the clear's RGB. Pure straight target (200,120,40,128).
		want := color.NRGBA{R: 200, G: 120, B: 40, A: 128}
		const tol = 2
		if absDiff8(c.R, want.R) > tol || absDiff8(c.G, want.G) > tol ||
			absDiff8(c.B, want.B) > tol || absDiff8(c.A, want.A) > tol {
			t.Fatalf("clear bg=%+v matted the fill: got %+v, want approx %+v", bg, c, want)
		}
	}
}
