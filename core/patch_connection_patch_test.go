package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestConnectionPatchShrinkUsesPointUnits(t *testing.T) {
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
			ShrinkA:         1.5,
			ShrinkB:         3,
		},
		XYA:     geom.Pt{X: 0.25, Y: 0.5},
		XYB:     geom.Pt{X: 0.75, Y: 0.5},
		CoordsA: Coords(CoordAxes),
		CoordsB: Coords(CoordAxes),
	}
	ctx := createTestDrawContext()
	ctx.RC.DPI = 144

	path := patch.connectionDisplayPath(ctx)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) != 3 {
		t.Fatalf("connection path = commands %+v vertices %+v, want quadratic path", path.C, path.V)
	}
	start := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYA)
	end := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYB)
	shrinkA := pointsToPixels(ctx.RC, patch.ShrinkA)
	shrinkB := pointsToPixels(ctx.RC, patch.ShrinkB)
	if !approx(path.V[0].X, start.X+shrinkA, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("connection start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + shrinkA, Y: start.Y})
	}
	if !approx(path.V[2].X, end.X-shrinkB, 1e-9) || !approx(path.V[2].Y, end.Y, 1e-9) {
		t.Fatalf("connection end = %+v, want %+v", path.V[2], geom.Pt{X: end.X - shrinkB, Y: end.Y})
	}
}

func TestConnectionPatchClipsEndpointsToPatches(t *testing.T) {
	patchA := &Rectangle{
		XY:     geom.Pt{X: 0.5, Y: 4.5},
		Width:  1,
		Height: 1,
		Coords: Coords(CoordData),
	}
	patchB := &Rectangle{
		XY:     geom.Pt{X: 8.5, Y: 4.5},
		Width:  1,
		Height: 1,
		Coords: Coords(CoordData),
	}
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
			PatchA:          patchA,
			PatchB:          patchB,
		},
		XYA:     geom.Pt{X: 1, Y: 5},
		XYB:     geom.Pt{X: 9, Y: 5},
		CoordsA: Coords(CoordData),
		CoordsB: Coords(CoordData),
	}
	ctx := createTestDrawContext()

	path := patch.connectionDisplayPath(ctx)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) != 3 {
		t.Fatalf("connection path = commands %+v vertices %+v, want quadratic path", path.C, path.V)
	}
	wantStart := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 1.5, Y: 5})
	wantEnd := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 8.5, Y: 5})
	if !approxPt(path.V[0], wantStart, 1e-9) {
		t.Fatalf("patch-clipped start = %+v, want %+v", path.V[0], wantStart)
	}
	if !approxPt(path.V[2], wantEnd, 1e-9) {
		t.Fatalf("patch-clipped end = %+v, want %+v", path.V[2], wantEnd)
	}
}

func TestConnectionPatchResolvesIndependentCoordinateSpaces(t *testing.T) {
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			Patch: Patch{
				EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
				EdgeWidth: 1,
			},
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
		},
		XYA:     geom.Pt{X: 0.25, Y: 0.75},
		XYB:     geom.Pt{X: 0.5, Y: 0.5},
		CoordsA: Coords(CoordData),
		CoordsB: Coords(CoordAxes),
	}

	ctx := createTestDrawContext()
	r := &recordingRenderer{}
	patch.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one connection path, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].path
	wantA := ctx.TransformFor(Coords(CoordData)).Apply(patch.XYA)
	wantB := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYB)
	if len(got.C) != 2 || got.C[0] != geom.MoveTo || got.C[1] != geom.QuadTo || len(got.V) != 3 {
		t.Fatalf("expected straight quadratic arc3 path, got commands %+v vertices %+v", got.C, got.V)
	}
	if !approx(got.V[0].X, wantA.X, 1e-9) || !approx(got.V[0].Y, wantA.Y, 1e-9) {
		t.Fatalf("connection start = %+v, want %+v", got.V[0], wantA)
	}
	if !approx(got.V[2].X, wantB.X, 1e-9) || !approx(got.V[2].Y, wantB.Y, 1e-9) {
		t.Fatalf("connection end = %+v, want %+v", got.V[2], wantB)
	}
}
