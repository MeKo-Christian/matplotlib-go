package transform

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAffine2DRotationBuildersMatchMatplotlibComposition(t *testing.T) {
	translated := NewAffine(geom.Affine{A: 1, D: 1, E: 2, F: 3})
	rotated := translated.Rotate(math.Pi / 2)

	wantPt(t, rotated.Apply(geom.Pt{X: 1}), geom.Pt{X: -3, Y: 3}, 1e-12)
	wantPt(t, translated.RotateDegrees(90).Apply(geom.Pt{X: 1}), rotated.Apply(geom.Pt{X: 1}), 1e-12)

	center := geom.Pt{X: 4, Y: -2}
	around := NewAffine2D().RotateAround(center, math.Pi/2)
	wantPt(t, around.Apply(center), center, 1e-12)
	wantPt(t, around.Apply(geom.Pt{X: 5, Y: -2}), geom.Pt{X: 4, Y: -1}, 1e-12)
	wantPt(
		t,
		NewAffine2D().RotateAroundDegrees(center, 90).Apply(geom.Pt{X: 5, Y: -2}),
		geom.Pt{X: 4, Y: -1},
		1e-12,
	)
}

func TestAffine2DSkewBuildersMatchMatplotlibComposition(t *testing.T) {
	xShear := math.Atan(2)
	yShear := math.Atan(3)
	skewed := NewAffine2D().Skew(xShear, yShear)

	wantPt(t, skewed.Apply(geom.Pt{X: 4, Y: 5}), geom.Pt{X: 14, Y: 17}, 1e-12)
	wantPt(
		t,
		NewAffine2D().SkewDegrees(xShear*180/math.Pi, yShear*180/math.Pi).Apply(geom.Pt{X: 4, Y: 5}),
		geom.Pt{X: 14, Y: 17},
		1e-12,
	)

	matrix, ok := AsAffine(skewed)
	if !ok {
		t.Fatal("skewed AffineT must remain exactly affine")
	}
	wantPt(t, matrix.Apply(geom.Pt{X: 4, Y: 5}), skewed.Apply(geom.Pt{X: 4, Y: 5}), 1e-12)
}

func TestScaledTranslationTracksScaleTransformAndFreezes(t *testing.T) {
	box := NewBbox(rectFor(0, 0, 72, 144))
	scale := NewBboxTransformTo(box)
	scaled := NewScaledTranslation(geom.Pt{X: 1, Y: 0.5}, scale)
	var dependent TransformNode
	scaled.AddDependent(&dependent)

	p := geom.Pt{X: 10, Y: 20}
	wantPt(t, scaled.Apply(p), geom.Pt{X: 82, Y: 92}, 1e-12)
	back, ok := scaled.Invert(scaled.Apply(p))
	if !ok {
		t.Fatal("scaled translation must be invertible")
	}
	wantPt(t, back, p, 1e-12)

	frozen := Frozen(scaled)
	box.Set(rectFor(0, 0, 144, 288))
	if !dependent.Invalid().Has(InvalidAffine) {
		t.Fatal("scale-transform invalidation did not reach scaled-translation dependents")
	}
	wantPt(t, scaled.Apply(p), geom.Pt{X: 154, Y: 164}, 1e-12)
	wantPt(t, frozen.Apply(p), geom.Pt{X: 82, Y: 92}, 1e-12)

	matrix, affine := AsAffine(scaled)
	if !affine {
		t.Fatal("scaled translation must always be affine")
	}
	wantPt(t, matrix.Apply(p), scaled.Apply(p), 1e-12)
}

func TestTransformWrapperReplacementInvalidatesAndFreezes(t *testing.T) {
	wrapper := NewTransformWrapper(NewAffine(geom.Affine{A: 2, D: 3}))
	var dependent TransformNode
	wrapper.AddDependent(&dependent)

	p := geom.Pt{X: 2, Y: 4}
	wantPt(t, wrapper.Apply(p), geom.Pt{X: 4, Y: 12}, 1e-12)
	frozen := Frozen(wrapper)

	wrapper.Set(NewAffine(geom.Affine{A: 1, D: 1, E: 5, F: -2}))
	if !wrapper.Invalid().Has(InvalidAll) {
		t.Fatalf("wrapper invalidation = %v, want InvalidAll", wrapper.Invalid())
	}
	if !dependent.Invalid().Has(InvalidAll) {
		t.Fatalf("dependent invalidation = %v, want InvalidAll", dependent.Invalid())
	}
	wantPt(t, wrapper.Apply(p), geom.Pt{X: 7, Y: 2}, 1e-12)
	wantPt(t, frozen.Apply(p), geom.Pt{X: 4, Y: 12}, 1e-12)

	matrix, ok := AsAffine(wrapper)
	if !ok {
		t.Fatal("affine wrapper must flatten")
	}
	wantPt(t, matrix.Apply(p), wrapper.Apply(p), 1e-12)

	wrapper.Set(nil)
	wantPt(t, wrapper.Apply(p), p, 1e-12)
	back, ok := wrapper.Invert(p)
	if !ok {
		t.Fatal("nil-child wrapper must be invertible")
	}
	wantPt(t, back, p, 1e-12)
}

func TestTransformWrapperForwardsChildInvalidation(t *testing.T) {
	box := NewBbox(rectFor(0, 0, 10, 20))
	child := NewBboxTransformTo(box)
	wrapper := NewTransformWrapper(child)
	var dependent TransformNode
	wrapper.AddDependent(&dependent)

	box.Set(rectFor(0, 0, 20, 40))
	if !dependent.Invalid().Has(InvalidAffine) {
		t.Fatal("child invalidation did not propagate through wrapper")
	}
	wantPt(t, wrapper.Apply(geom.Pt{X: 0.5, Y: 0.5}), geom.Pt{X: 10, Y: 20}, 1e-12)
}
