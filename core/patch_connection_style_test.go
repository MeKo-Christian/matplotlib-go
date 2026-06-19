package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestArrowAndConnectionStyleRegistriesParseMatplotlibNames(t *testing.T) {
	for _, name := range []string{"-", "->", "<-", "<->", "-[", "]-[", "|-|", "-|>", "<|-|>", "simple", "fancy", "wedge"} {
		style, ok := ArrowStyleFromString(name)
		if !ok {
			t.Fatalf("ArrowStyleFromString(%q) returned !ok", name)
		}
		if style.Name == "" {
			t.Fatalf("ArrowStyleFromString(%q) returned empty name", name)
		}
	}

	style, ok := ArrowStyleFromString("->,head_length=0.8,head_width=0.4")
	if !ok {
		t.Fatal("ArrowStyleFromString with parameters returned !ok")
	}
	if style.HeadLength != 0.8 || style.HeadWidth != 0.4 {
		t.Fatalf("style parameters not parsed: %+v", style)
	}

	for _, name := range []string{"arc3", "arc", "angle", "angle3", "bar"} {
		style, ok := ConnectionStyleFromString(name)
		if !ok {
			t.Fatalf("ConnectionStyleFromString(%q) returned !ok", name)
		}
		if style.Name == "" {
			t.Fatalf("ConnectionStyleFromString(%q) returned empty name", name)
		}
	}

	conn, ok := ConnectionStyleFromString("arc3,rad=0.35")
	if !ok {
		t.Fatal("ConnectionStyleFromString with parameters returned !ok")
	}
	if conn.Rad != 0.35 {
		t.Fatalf("connection parameters not parsed: %+v", conn)
	}
}

func TestConnectionStyleArc3ZeroRadKeepsQuadraticPath(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc3")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc3) returned !ok")
	}
	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo {
		t.Fatalf("arc3 zero-rad commands = %+v, want MoveTo/QuadTo", path.C)
	}
	want := []geom.Pt{{X: 0, Y: 0}, {X: 50, Y: 0}, {X: 100, Y: 0}}
	if len(path.V) != len(want) {
		t.Fatalf("arc3 zero-rad vertices = %+v, want %+v", path.V, want)
	}
	for i := range want {
		if !approxPt(path.V[i], want[i], 1e-9) {
			t.Fatalf("arc3 zero-rad vertex[%d] = %+v, want %+v", i, path.V[i], want[i])
		}
	}
}

func TestConnectionStyleArc3RadUsesDisplayYUpCoordinates(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc3,rad=0.25")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc3,rad=0.25) returned !ok")
	}

	start := geom.Pt{X: 100, Y: 120}
	end := geom.Pt{X: 200, Y: 170}
	path := style.connect(start, end, 0, 0)

	// Display space is y-up, so connect() uses the verbatim Matplotlib formula
	// (Arc3.connect): cx = x12 + rad*dy, cy = y12 - rad*dx.
	wantCtrl := geom.Pt{
		X: (start.X+end.X)/2 + style.Rad*(end.Y-start.Y),
		Y: (start.Y+end.Y)/2 - style.Rad*(end.X-start.X),
	}
	if len(path.V) != 3 || !approxPt(path.V[1], wantCtrl, 1e-9) {
		t.Fatalf("arc3 control = %+v, want y-up Matplotlib-equivalent control %+v", path.V, wantCtrl)
	}
}

func TestConnectionStyleArcUsesMatplotlibDefaultAngles(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc,armA=10,armB=5")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc) returned !ok")
	}
	if style.AngleA != 0 || style.AngleB != 0 {
		t.Fatalf("arc defaults = angleA %v angleB %v, want 0/0", style.AngleA, style.AngleB)
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.V) < 3 {
		t.Fatalf("arc path vertices = %+v, want start arm and end arm", path.V)
	}
	if !approx(path.V[1].X, 10, 1e-9) || !approx(path.V[1].Y, 0, 1e-9) {
		t.Fatalf("arc start arm = %+v, want horizontal +x arm at {10,0}", path.V[1])
	}
}

func TestConnectionStyleArcRadiusRoundsArmEndpointLikeMatplotlib(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc,armA=20,rad=5")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc) returned !ok")
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.V) < 5 {
		t.Fatalf("arc rounded path vertices = %+v, want start, rounded arm, and endpoint", path.V)
	}
	want := []geom.Pt{
		{X: 0, Y: 0},
		{X: 15, Y: 0},
		{X: 20, Y: 0},
		{X: 25, Y: 0},
		{X: 100, Y: 0},
	}
	for i := range want {
		if !approxPt(path.V[i], want[i], 1e-9) {
			t.Fatalf("arc rounded vertex %d = %+v, want %+v (all vertices %+v)", i, path.V[i], want[i], path.V)
		}
	}
	if path.C[2] != geom.QuadTo {
		t.Fatalf("arc rounded corner command = %v, want QuadTo in commands %+v", path.C[2], path.C)
	}
}

func TestConnectionStyleBarAngleProjectsEndpointLikeMatplotlib(t *testing.T) {
	style, ok := ConnectionStyleFromString("bar,angle=0,fraction=0.3")
	if !ok {
		t.Fatal("ConnectionStyleFromString(bar) returned !ok")
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 100}, 0, 0)
	if len(path.V) != 4 {
		t.Fatalf("bar path vertices = %+v, want 4 vertices", path.V)
	}
	wantY := -0.3 * math.Hypot(100, 100)
	if !approx(path.V[1].X, 0, 1e-9) || !approx(path.V[1].Y, wantY, 1e-9) {
		t.Fatalf("bar first arm = %+v, want {0,%v}", path.V[1], wantY)
	}
	if !approx(path.V[2].X, 100, 1e-9) || !approx(path.V[2].Y, wantY, 1e-9) {
		t.Fatalf("bar projected second arm = %+v, want {100,%v}", path.V[2], wantY)
	}
	if path.V[3] != (geom.Pt{X: 100, Y: 100}) {
		t.Fatalf("bar final endpoint = %+v, want original endpoint", path.V[3])
	}
}
