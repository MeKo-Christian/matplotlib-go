package agg

import (
	"image/png"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestAlphaOverTransparentReproD5 gathers evidence for PLAN.md D5: a
// semi-transparent fill over a transparent surface should save to PNG with the
// original (straight-alpha) RGB unchanged, only the alpha reduced. matplotlib
// stores #E91E63 at alpha 0.8 as straight RGBA (233,30,99,204).
func TestAlphaOverTransparentReproD5(t *testing.T) {
	r, err := New(4, 4, render.Color{}) // fully transparent background
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	fill := render.Color{R: 233.0 / 255, G: 30.0 / 255, B: 99.0 / 255, A: 0.8}
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: 4, Y: 4}}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	r.Path(fullRectPath(4, 4), &render.Paint{Fill: fill})
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	mem := r.GetImage().RGBAAt(2, 2)
	t.Logf("in-memory GetImage RGBA = %+v (premult-domain)", mem)

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
	dr, dg, db, da := decoded.At(2, 2).RGBA()
	t.Logf("decoded PNG straight RGBA(16) = (%d,%d,%d,%d) -> 8bit (%d,%d,%d,%d)",
		dr, dg, db, da, dr>>8, dg>>8, db>>8, da>>8)

	wantR, wantG, wantB, wantA := uint32(233), uint32(30), uint32(99), uint32(204)
	gotR, gotG, gotB, gotA := dr>>8, dg>>8, db>>8, da>>8
	if absDiff(gotR, wantR) > 2 || absDiff(gotG, wantG) > 2 || absDiff(gotB, wantB) > 2 || absDiff(gotA, wantA) > 2 {
		t.Fatalf("saved PNG straight RGBA = (%d,%d,%d,%d), want approx (%d,%d,%d,%d)",
			gotR, gotG, gotB, gotA, wantR, wantG, wantB, wantA)
	}
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
