package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPatchAutoScaleIgnoresNonDataCoords(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&Rectangle{
		Patch:  Patch{FaceColor: render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}},
		XY:     geom.Pt{X: 2, Y: 3},
		Width:  2,
		Height: 3,
	})
	ax.AddPatch(&Circle{
		Patch:  Patch{FaceColor: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}},
		Center: geom.Pt{X: 0.5, Y: 0.5},
		Radius: 0.35,
		Coords: Coords(CoordAxes),
	})

	ax.AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if xMin != 2 || xMax != 4 {
		t.Fatalf("x domain = [%v, %v], want [2, 4]", xMin, xMax)
	}
	if yMin != 3 || yMax != 6 {
		t.Fatalf("y domain = [%v, %v], want [3, 6]", yMin, yMax)
	}
}

func TestPathPatchLegendEntryIncludesHatch(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&PathPatch{
		Patch: Patch{
			FaceColor:  render.Color{R: 0.7, G: 0.7, B: 0.2, A: 1},
			EdgeColor:  render.Color{R: 0.2, G: 0.2, B: 0.1, A: 1},
			EdgeWidth:  1,
			Hatch:      "x",
			HatchColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			Label:      "region",
		},
		Path: patchRectPath(geom.Rect{
			Min: geom.Pt{X: 1, Y: 1},
			Max: geom.Pt{X: 3, Y: 2},
		}),
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("expected one legend entry, got %d", len(entries))
	}
	if entries[0].Label != "region" || entries[0].kind != legendEntryPatch {
		t.Fatalf("unexpected legend entry: %+v", entries[0])
	}
	if entries[0].patchHatch != "x" {
		t.Fatalf("expected hatch metadata to be preserved, got %+v", entries[0])
	}
}

func TestAdditionalPatchClassesDrawExpectedPaths(t *testing.T) {
	ctx := createTestDrawContext()
	patches := []Artist{
		&RegularPolygon{
			Patch:       Patch{FaceColor: render.Color{A: 1}},
			Center:      geom.Pt{X: 2, Y: 2},
			NumVertices: 5,
			Radius:      1,
		},
		&CirclePolygon{
			Patch:      Patch{FaceColor: render.Color{A: 1}},
			Center:     geom.Pt{X: 2, Y: 2},
			Radius:     1,
			Resolution: 12,
		},
		&Arc{
			Patch:    Patch{EdgeColor: render.Color{A: 1}, EdgeWidth: 1},
			Center:   geom.Pt{X: 2, Y: 2},
			Width:    2,
			Height:   1,
			Theta1:   10,
			Theta2:   220,
			EdgeOnly: true,
		},
		&Annulus{
			Patch:   Patch{FaceColor: render.Color{A: 1}},
			Center:  geom.Pt{X: 2, Y: 2},
			RadiusA: 1.5,
			RadiusB: 1.0,
			Width:   0.3,
			Angle:   15,
		},
		&StepPatch{
			Patch:    Patch{FaceColor: render.Color{A: 1}, EdgeColor: render.Color{A: 1}, EdgeWidth: 1},
			Values:   []float64{1, 2, 1.5},
			Edges:    []float64{0, 1, 2, 3},
			Baseline: float64Ptr(0),
		},
	}

	for _, patch := range patches {
		r := &recordingRenderer{}
		patch.Draw(r, ctx)
		if len(r.pathCalls) == 0 {
			t.Fatalf("%T did not draw a path", patch)
		}
	}
}

func TestShadowOffsetsAndDarkensSourcePatch(t *testing.T) {
	source := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0.2, G: 0.1, B: 0.1, A: 1},
			EdgeWidth: 1,
		},
		XY:     geom.Pt{X: 1, Y: 2},
		Width:  2,
		Height: 1,
	}
	shadow := &Shadow{Patch: Patch{}, Source: source, Offset: geom.Pt{X: 0.5, Y: -0.25}, Shade: 0.7}

	r := &recordingRenderer{}
	shadow.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one shadow path, got %d", len(r.pathCalls))
	}
	paint := r.pathCalls[0].paint
	if paint.Fill.A != 0.5 {
		t.Fatalf("shadow alpha = %v, want 0.5", paint.Fill.A)
	}
	if paint.Fill.R >= source.FaceColor.R {
		t.Fatalf("shadow face color was not darkened: got %+v source %+v", paint.Fill, source.FaceColor)
	}
}

// TestNormalForVectorIsYUpCounterClockwisePerpendicular locks the arrow-head
// normal to Matplotlib's y-up convention: the perpendicular is the +90° (CCW)
// rotation of the direction, so a rightward shaft (+x) yields an upward (+y)
// normal. Under the old y-down display space this normal pointed the opposite
// way and arrow heads splayed on the wrong side.
func TestNormalForVectorIsYUpCounterClockwisePerpendicular(t *testing.T) {
	cases := []struct {
		v, want geom.Pt
	}{
		{geom.Pt{X: 1, Y: 0}, geom.Pt{X: 0, Y: 1}},   // +x -> +y (up)
		{geom.Pt{X: 0, Y: 1}, geom.Pt{X: -1, Y: 0}},  // +y -> -x
		{geom.Pt{X: -1, Y: 0}, geom.Pt{X: 0, Y: -1}}, // -x -> -y
		{geom.Pt{X: 3, Y: 4}, geom.Pt{X: -4.0 / 5, Y: 3.0 / 5}},
		{geom.Pt{X: 0, Y: 0}, geom.Pt{X: 0, Y: 0}}, // degenerate
	}
	for _, c := range cases {
		if got := normalForVector(c.v); !approxPt(got, c.want, 1e-9) {
			t.Fatalf("normalForVector(%+v) = %+v, want CCW perpendicular %+v", c.v, got, c.want)
		}
	}
}

// TestArrowHeadPathPlacesBaseCornersPerpendicularInYUp asserts that a rightward
// arrow head keeps its base behind the tip along the shaft and straddles the
// shaft with the left corner above (+y) and the right corner below (-y), the
// y-up orientation Matplotlib draws.
func TestArrowHeadPathPlacesBaseCornersPerpendicularInYUp(t *testing.T) {
	const headLength, headWidth = 4.0, 4.0
	path := arrowHeadPath(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 10, Y: 0}, headLength, headWidth, true, 0)
	if len(path.V) < 3 {
		t.Fatalf("arrow head path vertices = %+v, want at least tip/left/right", path.V)
	}
	tip, left, right := path.V[0], path.V[1], path.V[2]
	if !approxPt(tip, geom.Pt{X: 10, Y: 0}, 1e-9) {
		t.Fatalf("arrow head tip = %+v, want {10,0}", tip)
	}
	if !approxPt(left, geom.Pt{X: 6, Y: 4}, 1e-9) {
		t.Fatalf("arrow head left base = %+v, want {6,4} (behind tip, above shaft)", left)
	}
	if !approxPt(right, geom.Pt{X: 6, Y: -4}, 1e-9) {
		t.Fatalf("arrow head right base = %+v, want {6,-4} (behind tip, below shaft)", right)
	}
}

// TestRotationAffineRotatesCounterClockwiseInYUp pins the rotation used for
// rotated text and its bbox to Matplotlib's CCW-positive convention in y-up
// display space: a positive angle moves a point on a box's right edge toward
// the top edge. Under y-down this rotation ran clockwise and rotated labels
// tilted the wrong way.
func TestRotationAffineRotatesCounterClockwiseInYUp(t *testing.T) {
	rot := rotationAffine(90)
	if got := rot.Apply(geom.Pt{X: 1, Y: 0}); !approxPt(got, geom.Pt{X: 0, Y: 1}, 1e-9) {
		t.Fatalf("rotationAffine(90).Apply({1,0}) = %+v, want {0,1} (CCW: right -> up)", got)
	}

	// Rotate the corners of a unit box about its center by +90°. The corner on
	// the right edge must land on the top edge (CCW), not the bottom.
	box := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 2, Y: 2}}
	center := rectCenter(box)
	toCenter := translateAffine(geom.Pt{X: -center.X, Y: -center.Y})
	fromCenter := translateAffine(center)
	aboutCenter := fromCenter.Mul(rotationAffine(90)).Mul(toCenter)

	rightEdgeMid := geom.Pt{X: box.Max.X, Y: center.Y}   // {2,1}
	wantTopEdgeMid := geom.Pt{X: center.X, Y: box.Max.Y} // {1,2}
	if got := aboutCenter.Apply(rightEdgeMid); !approxPt(got, wantTopEdgeMid, 1e-9) {
		t.Fatalf("CCW box rotation mapped right edge %+v to %+v, want top edge %+v", rightEdgeMid, got, wantTopEdgeMid)
	}
}
