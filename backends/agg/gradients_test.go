package agg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestSupportsGradientFillAdvertised(t *testing.T) {
	r := mustNew(t, 10, 10)
	if !r.SupportsGradientFill() {
		t.Fatal("AGG renderer should advertise SupportsGradientFill")
	}
}

func TestSupportsPatternFillAdvertised(t *testing.T) {
	r := mustNew(t, 10, 10)
	if !r.SupportsPatternFill() {
		t.Fatal("AGG renderer should advertise SupportsPatternFill")
	}
}

func TestLinearGradientFillProducesGradientAcrossX(t *testing.T) {
	r := mustNew(t, 40, 10)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 40, Y: 10}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 0, Y: 0})
	rect.LineTo(geom.Pt{X: 40, Y: 0})
	rect.LineTo(geom.Pt{X: 40, Y: 10})
	rect.LineTo(geom.Pt{X: 0, Y: 10})
	rect.Close()

	r.Path(rect, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 5},
			End:   geom.Pt{X: 40, Y: 5},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{B: 1, A: 1}},
			},
		},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	img := r.ctx.GetImage()
	if img == nil {
		t.Fatal("nil image after Path")
	}

	// Sample colors near the left edge (mostly red) and right edge (mostly
	// blue). Allow generous tolerances because gradient endpoints land inside
	// the path so the very edge pixel may be antialiased.
	leftR, _, leftB, _ := pixelAt(t, r, 2, 5)
	rightR, _, rightB, _ := pixelAt(t, r, 37, 5)

	if leftR < 200 {
		t.Fatalf("expected strong red on left edge, got R=%d", leftR)
	}
	if rightB < 200 {
		t.Fatalf("expected strong blue on right edge, got B=%d", rightB)
	}
	if rightR > 80 {
		t.Fatalf("right edge should not be red, got R=%d", rightR)
	}
	if leftB > 80 {
		t.Fatalf("left edge should not be blue, got B=%d", leftB)
	}
}

func TestRadialGradientFillProducesCenterToEdgeFalloff(t *testing.T) {
	r := mustNew(t, 40, 40)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 40, Y: 40}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 0, Y: 0})
	rect.LineTo(geom.Pt{X: 40, Y: 0})
	rect.LineTo(geom.Pt{X: 40, Y: 40})
	rect.LineTo(geom.Pt{X: 0, Y: 40})
	rect.Close()

	r.Path(rect, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 20, Y: 20},
			Radius: 20,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
				{Offset: 1, Color: render.Color{A: 1}},
			},
		},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	cR, cG, cB, _ := pixelAt(t, r, 20, 20)
	eR, eG, eB, _ := pixelAt(t, r, 0, 0)
	if cR < 200 || cG < 200 || cB < 200 {
		t.Fatalf("expected near-white center, got (%d,%d,%d)", cR, cG, cB)
	}
	if eR > 60 || eG > 60 || eB > 60 {
		t.Fatalf("expected near-black corner, got (%d,%d,%d)", eR, eG, eB)
	}
}

func TestSolidFillStillWorksAfterGradient(t *testing.T) {
	// Renderer must reset the fill source between draws so subsequent solid
	// fills are not painted through the gradient span generator.
	r := mustNew(t, 40, 20)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 40, Y: 20}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var leftHalf geom.Path
	leftHalf.MoveTo(geom.Pt{X: 0, Y: 0})
	leftHalf.LineTo(geom.Pt{X: 20, Y: 0})
	leftHalf.LineTo(geom.Pt{X: 20, Y: 20})
	leftHalf.LineTo(geom.Pt{X: 0, Y: 20})
	leftHalf.Close()

	r.Path(leftHalf, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 0, Y: 10},
			End:   geom.Pt{X: 20, Y: 10},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 1, A: 1}},
				{Offset: 1, Color: render.Color{G: 1, A: 1}},
			},
		},
	})

	var rightHalf geom.Path
	rightHalf.MoveTo(geom.Pt{X: 20, Y: 0})
	rightHalf.LineTo(geom.Pt{X: 40, Y: 0})
	rightHalf.LineTo(geom.Pt{X: 40, Y: 20})
	rightHalf.LineTo(geom.Pt{X: 20, Y: 20})
	rightHalf.Close()

	r.Path(rightHalf, &render.Paint{Fill: render.Color{B: 1, A: 1}})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	rR, _, rB, _ := pixelAt(t, r, 30, 10)
	if rB < 200 || rR > 60 {
		t.Fatalf("right half should be solid blue, got (R=%d,B=%d)", rR, rB)
	}
}

func TestPatternFillRepeatsTileAcrossPath(t *testing.T) {
	r := mustNew(t, 24, 12)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 24, Y: 12}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 0, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 12})
	rect.LineTo(geom.Pt{X: 0, Y: 12})
	rect.Close()

	var stripe geom.Path
	stripe.MoveTo(geom.Pt{X: 0, Y: 0})
	stripe.LineTo(geom.Pt{X: 4, Y: 0})
	stripe.LineTo(geom.Pt{X: 4, Y: 12})
	stripe.LineTo(geom.Pt{X: 0, Y: 12})
	stripe.Close()

	r.Path(rect, &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "stripe",
			Cell:       geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:       stripe,
			Foreground: render.Color{R: 1, A: 1},
			Background: render.Color{B: 1, A: 1},
		},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	leftR, _, leftB, _ := pixelAt(t, r, 2, 6)
	gapR, _, gapB, _ := pixelAt(t, r, 6, 6)
	repeatR, _, repeatB, _ := pixelAt(t, r, 10, 6)
	if leftR < 200 || leftB > 80 {
		t.Fatalf("first pattern stripe should be red, got R=%d B=%d", leftR, leftB)
	}
	if gapB < 200 || gapR > 80 {
		t.Fatalf("pattern background should be blue, got R=%d B=%d", gapR, gapB)
	}
	if repeatR < 200 || repeatB > 80 {
		t.Fatalf("repeated pattern stripe should be red, got R=%d B=%d", repeatR, repeatB)
	}
}

func TestPatternFillAppliesTileTransform(t *testing.T) {
	r := mustNew(t, 24, 12)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 24, Y: 12}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 0, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 12})
	rect.LineTo(geom.Pt{X: 0, Y: 12})
	rect.Close()

	var stripe geom.Path
	stripe.MoveTo(geom.Pt{X: 0, Y: 0})
	stripe.LineTo(geom.Pt{X: 4, Y: 0})
	stripe.LineTo(geom.Pt{X: 4, Y: 12})
	stripe.LineTo(geom.Pt{X: 0, Y: 12})
	stripe.Close()

	r.Path(rect, &render.Paint{
		FillPattern: render.PatternFill{
			ID:           "shifted-stripe",
			Cell:         geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:         stripe,
			Foreground:   render.Color{R: 1, A: 1},
			Background:   render.Color{B: 1, A: 1},
			Transform:    geom.Affine{A: 1, D: 1, E: 4},
			HasTransform: true,
		},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	gapR, _, gapB, _ := pixelAt(t, r, 2, 6)
	shiftR, _, shiftB, _ := pixelAt(t, r, 6, 6)
	if gapB < 200 || gapR > 80 {
		t.Fatalf("unshifted stripe location should be background blue, got R=%d B=%d", gapR, gapB)
	}
	if shiftR < 200 || shiftB > 80 {
		t.Fatalf("transformed stripe location should be red, got R=%d B=%d", shiftR, shiftB)
	}
}

func TestHatchTakesPrecedenceOverPatternFill(t *testing.T) {
	r := mustNew(t, 24, 12)
	viewport := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 24, Y: 12}}
	if err := r.Begin(viewport); err != nil {
		t.Fatalf("Begin: %v", err)
	}

	var rect geom.Path
	rect.MoveTo(geom.Pt{X: 0, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 0})
	rect.LineTo(geom.Pt{X: 24, Y: 12})
	rect.LineTo(geom.Pt{X: 0, Y: 12})
	rect.Close()

	var fullTile geom.Path
	fullTile.MoveTo(geom.Pt{X: 0, Y: 0})
	fullTile.LineTo(geom.Pt{X: 8, Y: 0})
	fullTile.LineTo(geom.Pt{X: 8, Y: 12})
	fullTile.LineTo(geom.Pt{X: 0, Y: 12})
	fullTile.Close()

	r.Path(rect, &render.Paint{
		Hatch:          "|",
		HatchColor:     render.Color{G: 1, A: 1},
		HatchLineWidth: 2,
		HatchSpacing:   6,
		FillPattern: render.PatternFill{
			ID:         "red-tile",
			Cell:       geom.Rect{Max: geom.Pt{X: 8, Y: 12}},
			Path:       fullTile,
			Foreground: render.Color{R: 1, A: 1},
		},
	})

	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	redFound := false
	greenFound := false
	for y := 1; y < 11; y++ {
		for x := 1; x < 23; x++ {
			rv, gv, bv, _ := pixelAt(t, r, x, y)
			if rv > 180 && gv < 80 && bv < 80 {
				redFound = true
			}
			if gv > 180 && rv < 120 && bv < 120 {
				greenFound = true
			}
		}
	}
	if redFound {
		t.Fatal("pattern fill painted red pixels even though hatch should take precedence")
	}
	if !greenFound {
		t.Fatal("expected hatch to paint green pixels")
	}
}

func pixelAt(t *testing.T, r *Renderer, x, y int) (uint8, uint8, uint8, uint8) {
	t.Helper()
	rgba := r.GetImage()
	if rgba == nil {
		t.Fatal("nil image")
	}
	c := rgba.RGBAAt(x, y)
	return c.R, c.G, c.B, c.A
}
