package render

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestGraphicsContextEffectivePaintAppliesAlpha(t *testing.T) {
	gc := NewGraphicsContext()
	gc.Alpha = 0.25
	gc.Paint = Paint{
		Stroke: Color{R: 1, A: 0.8},
		Fill:   Color{B: 1, A: 0.4},
		Dashes: []float64{1, 2},
	}

	paint := gc.EffectivePaint()
	if paint.Stroke.A != 0.2 || paint.Fill.A != 0.1 {
		t.Fatalf("effective alpha stroke=%v fill=%v, want 0.2/0.1", paint.Stroke.A, paint.Fill.A)
	}
	paint.Dashes[0] = 99
	if gc.Paint.Dashes[0] == 99 {
		t.Fatal("EffectivePaint reused mutable dash backing storage")
	}
}

func TestGraphicsContextEffectivePaintCarriesRendererState(t *testing.T) {
	gc := NewGraphicsContext().
		WithAntialias(AntialiasOff).
		WithSnap(SnapAuto).
		WithHatch("/", Color{R: 1, A: 0.5}, 2).
		WithHatchSpacing(12).
		WithSketch(SketchParams{Scale: 1, Length: 2, Randomness: 3}).
		WithForcedAlpha(0.25).
		WithClipPathTransform(geomIdentity())
	gc.Paint = Paint{
		Stroke: Color{A: 1},
		Fill:   Color{A: 1},
	}

	paint := gc.EffectivePaint()
	if paint.Antialias != AntialiasOff || paint.Snap != SnapAuto {
		t.Fatalf("effective paint lost antialias/snap state: %+v", paint)
	}
	if paint.Hatch != "/" || paint.HatchColor.R != 1 || paint.HatchLineWidth != 2 || paint.HatchSpacing != 12 {
		t.Fatalf("effective paint lost hatch state: %+v", paint)
	}
	if paint.Sketch != (SketchParams{Scale: 1, Length: 2, Randomness: 3}) {
		t.Fatalf("effective paint lost sketch state: %+v", paint.Sketch)
	}
	if !paint.ForceAlpha || paint.Stroke.A != 0.25 || paint.Fill.A != 0.25 {
		t.Fatalf("effective paint did not force alpha: %+v", paint)
	}
	if !paint.HasClipPathTrans {
		t.Fatalf("effective paint lost clip path transform: %+v", paint)
	}
}

func TestGraphicsContextEffectivePaintCombinesForcedAndContextAlpha(t *testing.T) {
	gc := NewGraphicsContext().
		WithAlpha(0.5).
		WithForcedAlpha(0.25)
	gc.Paint = Paint{
		Stroke: Color{A: 1},
		Fill:   Color{A: 1},
	}

	paint := gc.EffectivePaint()
	if !paint.ForceAlpha {
		t.Fatalf("effective paint lost force-alpha flag: %+v", paint)
	}
	if paint.Alpha != 0.125 {
		t.Fatalf("effective forced alpha = %v, want 0.125", paint.Alpha)
	}
	if paint.Stroke.A != 0.125 || paint.Fill.A != 0.125 {
		t.Fatalf("effective color alpha stroke=%v fill=%v, want 0.125/0.125", paint.Stroke.A, paint.Fill.A)
	}
}

func TestGraphicsContextEffectivePaintCarriesEffectsAndMixedOutputState(t *testing.T) {
	patternPath := geom.Path{
		V: []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
	}
	gc := NewGraphicsContext().
		WithCompositeMode(CompositeMultiply).
		WithRasterization(Rasterization{Mode: RasterizeAlways, DPI: 144})
	gc.Paint = Paint{
		FillPattern: PatternFill{
			ID:         "diag",
			Cell:       geom.Rect{Max: geom.Pt{X: 4, Y: 4}},
			Path:       patternPath,
			Foreground: Color{A: 1},
		},
		FillGradient: GradientFill{
			Kind:  LinearGradient,
			Start: geom.Pt{X: 0, Y: 0},
			End:   geom.Pt{X: 10, Y: 0},
			Stops: []GradientStop{
				{Offset: 0, Color: Color{R: 1, A: 1}},
				{Offset: 1, Color: Color{B: 1, A: 0.5}},
			},
		},
		PathEffects: []PathEffect{
			{Kind: PathEffectStroke, Stroke: Color{A: 1}, LineWidth: 4, Offset: geom.Pt{X: 1, Y: 2}},
			{Kind: PathEffectNormal},
		},
	}

	paint := gc.EffectivePaint()
	if paint.CompositeMode != CompositeMultiply {
		t.Fatalf("effective paint composite mode = %v, want %v", paint.CompositeMode, CompositeMultiply)
	}
	if paint.Rasterization.Mode != RasterizeAlways || paint.Rasterization.DPI != 144 {
		t.Fatalf("effective paint rasterization = %+v, want always at 144dpi", paint.Rasterization)
	}
	if paint.FillPattern.ID != "diag" || len(paint.FillPattern.Path.V) != 2 {
		t.Fatalf("effective paint lost fill pattern: %+v", paint.FillPattern)
	}
	if paint.FillGradient.Kind != LinearGradient || len(paint.FillGradient.Stops) != 2 {
		t.Fatalf("effective paint lost fill gradient: %+v", paint.FillGradient)
	}
	if len(paint.PathEffects) != 2 || paint.PathEffects[0].Kind != PathEffectStroke {
		t.Fatalf("effective paint lost path effects: %+v", paint.PathEffects)
	}

	paint.FillPattern.Path.V[0].X = 99
	paint.FillGradient.Stops[0].Offset = 0.25
	paint.PathEffects[0].Offset.X = 99
	if gc.Paint.FillPattern.Path.V[0].X == 99 {
		t.Fatal("EffectivePaint reused mutable pattern path backing storage")
	}
	if gc.Paint.FillGradient.Stops[0].Offset == 0.25 {
		t.Fatal("EffectivePaint reused mutable gradient stop backing storage")
	}
	if gc.Paint.PathEffects[0].Offset.X == 99 {
		t.Fatal("EffectivePaint reused mutable path-effect backing storage")
	}
}

func geomIdentity() geom.Affine {
	return geom.Affine{A: 1, D: 1}
}
