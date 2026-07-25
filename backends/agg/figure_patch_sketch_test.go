package agg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// fullCanvasRectPath builds a closed full-canvas rectangle in y-up display coordinates,
// traversing corners (0,0)->(w,0)->(w,h)->(0,h) in the same order Matplotlib's
// figure.patch (a transformed unit rectangle) reaches RendererAgg.draw_path.
func fullCanvasRectPath(w, h float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: 0})
	p.LineTo(geom.Pt{X: w, Y: h})
	p.LineTo(geom.Pt{X: 0, Y: h})
	p.Close()
	return p
}

// TestSketchedFigurePatchBorderNotches locks in the figure-patch parity fix:
// a non-antialiased, sketch-perturbed full-canvas white fill over a transparent
// canvas must leave exactly the 42 fully-transparent border pixels Matplotlib's
// reference produces for sketch_xkcd (path.sketch=(1,100,2), 640x360 @ 100 DPI).
// AntialiasOff is essential — Matplotlib draws figure.patch with antialiased=False,
// so the wiggle yields hard binary-coverage notches, not a soft alpha gradient.
func TestSketchedFigurePatchBorderNotches(t *testing.T) {
	const W, H = 640, 360
	r, err := New(W, H, render.Color{R: 1, G: 1, B: 1, A: 0})
	if err != nil {
		t.Fatal(err)
	}
	r.SetResolution(100)
	r.SetDefaultSketch(render.SketchParams{Scale: 1, Length: 100, Randomness: 2})
	if err := r.Begin(geom.Rect{Max: geom.Pt{X: W, Y: H}}); err != nil {
		t.Fatal(err)
	}
	r.Path(fullCanvasRectPath(W, H), &render.Paint{
		Fill:      render.Color{R: 1, G: 1, B: 1, A: 1},
		Antialias: render.AntialiasOff,
	})
	if err := r.End(); err != nil {
		t.Fatal(err)
	}
	img := r.Image()

	// Matplotlib reference notch pixels (testdata/matplotlib_ref/sketch_xkcd.png),
	// all (255,255,255,0). Verified byte-exact during the parity work.
	want := map[[2]int]bool{
		{21, 0}: true, {22, 0}: true, {116, 0}: true, {117, 0}: true,
		{208, 0}: true, {209, 0}: true, {298, 0}: true, {392, 0}: true,
		{393, 0}: true, {485, 0}: true, {486, 0}: true, {576, 0}: true,
		{577, 0}: true, {578, 0}: true, {639, 22}: true, {639, 23}: true,
		{0, 68}: true, {0, 69}: true, {639, 115}: true, {0, 153}: true,
		{0, 154}: true, {639, 206}: true, {639, 207}: true, {639, 208}: true,
		{0, 246}: true, {639, 302}: true, {639, 303}: true, {0, 336}: true,
		{0, 337}: true, {66, 359}: true, {67, 359}: true, {68, 359}: true,
		{166, 359}: true, {167, 359}: true, {251, 359}: true, {252, 359}: true,
		{336, 359}: true, {426, 359}: true, {427, 359}: true, {511, 359}: true,
		{512, 359}: true, {602, 359}: true,
	}

	got := map[[2]int]bool{}
	var semi int
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			a := img.Pix[y*img.Stride+x*4+3]
			if a == 255 {
				continue
			}
			if a != 0 {
				semi++ // binary coverage: there must be NO partial-alpha pixels
				continue
			}
			got[[2]int{x, y}] = true
		}
	}
	if semi != 0 {
		t.Errorf("antialiased=False fill produced %d partial-alpha pixels; want 0 (binary coverage)", semi)
	}
	if len(got) != len(want) {
		t.Errorf("fully-transparent pixel count = %d, want %d", len(got), len(want))
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing expected transparent notch at (%d,%d)", k[0], k[1])
		}
	}
	for k := range got {
		if !want[k] {
			t.Errorf("unexpected transparent pixel at (%d,%d)", k[0], k[1])
		}
	}
}
