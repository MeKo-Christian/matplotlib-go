package agg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

// makeSimplifyPath builds a Path from alternating (x,y) pairs starting with MoveTo.
func makeSimplifyPath(pts ...float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: pts[0], Y: pts[1]})
	for i := 2; i < len(pts); i += 2 {
		p.LineTo(geom.Pt{X: pts[i], Y: pts[i+1]})
	}
	return p
}

func approxEqSimplify(a, b float64) bool { return a-b < 1e-9 && b-a < 1e-9 }

// TestRunningSegment_CollinearSimplification mirrors TestPathPipelineSimplifiesLinePaths:
// near-collinear 4-point path must collapse to just the two endpoints.
func TestRunningSegment_CollinearSimplification(t *testing.T) {
	p := makeSimplifyPath(0, 0, 1, 0.01, 2, -0.01, 3, 0)
	got := simplifyLinePath(p, 0.05)
	if len(got.V) != 2 {
		t.Fatalf("expected 2 vertices (MoveTo + LineTo endpoint), got %d: %v", len(got.V), got.V)
	}
	if got.V[0] != (geom.Pt{X: 0, Y: 0}) || got.V[1] != (geom.Pt{X: 3, Y: 0}) {
		t.Fatalf("expected [(0,0),(3,0)], got %v", got.V)
	}
}

// TestRunningSegment_AntiParallel_EndAtForwardMax ports
// test_antiparallel_simplification "test ending on a maximum" from
// third_party/matplotlib/lib/matplotlib/tests/test_simplification.py:93.
// Path oscillates up/down and the last extremum before the turn is the forward max (y=2).
// The simplifier must preserve the backward min (y=-1) and forward max (y=2).
func TestRunningSegment_AntiParallel_EndAtForwardMax(t *testing.T) {
	// x=[0,0,0,0,0,1], y=[.5,1,-1,1,2,.5]
	p := makeSimplifyPath(0, .5, 0, 1, 0, -1, 0, 1, 0, 2, 1, .5)
	got := simplifyLinePath(p, 0.5)
	want := [][2]float64{{0, .5}, {0, -1}, {0, 2}, {1, .5}}
	if len(got.V) != len(want) {
		t.Fatalf("expected %d vertices, got %d: %v", len(want), len(got.V), got.V)
	}
	for i, w := range want {
		if !approxEqSimplify(got.V[i].X, w[0]) || !approxEqSimplify(got.V[i].Y, w[1]) {
			t.Errorf("vertex[%d]: want %v, got %v", i, w, got.V[i])
		}
	}
}

// TestRunningSegment_AntiParallel_EndAtBackwardMax ports
// "test ending on a minimum" — last extremum before the turn is the backward max (y=-2).
func TestRunningSegment_AntiParallel_EndAtBackwardMax(t *testing.T) {
	// x=[0,0,0,0,0,1], y=[.5,1,-1,1,-2,.5]
	p := makeSimplifyPath(0, .5, 0, 1, 0, -1, 0, 1, 0, -2, 1, .5)
	got := simplifyLinePath(p, 0.5)
	want := [][2]float64{{0, .5}, {0, 1}, {0, -2}, {1, .5}}
	if len(got.V) != len(want) {
		t.Fatalf("expected %d vertices, got %d: %v", len(want), len(got.V), got.V)
	}
	for i, w := range want {
		if !approxEqSimplify(got.V[i].X, w[0]) || !approxEqSimplify(got.V[i].Y, w[1]) {
			t.Errorf("vertex[%d]: want %v, got %v", i, w, got.V[i])
		}
	}
}

// TestRunningSegment_AntiParallel_EndInMiddle ports
// "test ending in between" — last point before the turn is neither the forward
// nor backward max, so it must be emitted separately.
func TestRunningSegment_AntiParallel_EndInMiddle(t *testing.T) {
	// x=[0,0,0,0,0,1], y=[.5,1,-1,1,0,.5]
	p := makeSimplifyPath(0, .5, 0, 1, 0, -1, 0, 1, 0, 0, 1, .5)
	got := simplifyLinePath(p, 0.5)
	want := [][2]float64{{0, .5}, {0, 1}, {0, -1}, {0, 0}, {1, .5}}
	if len(got.V) != len(want) {
		t.Fatalf("expected %d vertices, got %d: %v", len(want), len(got.V), got.V)
	}
	for i, w := range want {
		if !approxEqSimplify(got.V[i].X, w[0]) || !approxEqSimplify(got.V[i].Y, w[1]) {
			t.Errorf("vertex[%d]: want %v, got %v", i, w, got.V[i])
		}
	}
}

// TestRunningSegment_NoAntiParallel_EndAtMax ports
// "test no anti-parallel ending at max" — all parallel; furthest point is the winner.
func TestRunningSegment_NoAntiParallel_EndAtMax(t *testing.T) {
	// x=[0,0,0,0,0,1], y=[.5,1,2,1,3,.5]
	p := makeSimplifyPath(0, .5, 0, 1, 0, 2, 0, 1, 0, 3, 1, .5)
	got := simplifyLinePath(p, 0.5)
	want := [][2]float64{{0, .5}, {0, 3}, {1, .5}}
	if len(got.V) != len(want) {
		t.Fatalf("expected %d vertices, got %d: %v", len(want), len(got.V), got.V)
	}
	for i, w := range want {
		if !approxEqSimplify(got.V[i].X, w[0]) || !approxEqSimplify(got.V[i].Y, w[1]) {
			t.Errorf("vertex[%d]: want %v, got %v", i, w, got.V[i])
		}
	}
}

// TestRunningSegment_NoAntiParallel_EndInMiddle ports
// "test no anti-parallel ending in middle" — forward max then retreat; the actual
// last point (not the max) must be emitted after the max.
func TestRunningSegment_NoAntiParallel_EndInMiddle(t *testing.T) {
	// x=[0,0,0,0,0,1], y=[.5,1,2,1,1,.5]
	p := makeSimplifyPath(0, .5, 0, 1, 0, 2, 0, 1, 0, 1, 1, .5)
	got := simplifyLinePath(p, 0.5)
	want := [][2]float64{{0, .5}, {0, 2}, {0, 1}, {1, .5}}
	if len(got.V) != len(want) {
		t.Fatalf("expected %d vertices, got %d: %v", len(want), len(got.V), got.V)
	}
	for i, w := range want {
		if !approxEqSimplify(got.V[i].X, w[0]) || !approxEqSimplify(got.V[i].Y, w[1]) {
			t.Errorf("vertex[%d]: want %v, got %v", i, w, got.V[i])
		}
	}
}

// TestRunningSegment_ZeroThreshold verifies that threshold=0 bypasses simplification.
func TestRunningSegment_ZeroThreshold(t *testing.T) {
	p := makeSimplifyPath(0, 0, 1, 0, 2, 0)
	got := simplifyLinePath(p, 0)
	if len(got.V) != 3 {
		t.Fatalf("zero threshold: expected unchanged path (3 vertices), got %d: %v", len(got.V), got.V)
	}
}

// TestRunningSegment_CurvesUnchanged verifies that paths with curves bypass simplification.
func TestRunningSegment_CurvesUnchanged(t *testing.T) {
	var p geom.Path
	p.MoveTo(geom.Pt{X: 0})
	p.QuadTo(geom.Pt{X: 1, Y: 1}, geom.Pt{X: 2})
	got := simplifyLinePath(p, 1.0)
	if len(got.V) != 3 {
		t.Fatalf("curves: expected unchanged path (3 vertices), got %d: %v", len(got.V), got.V)
	}
}

// TestRunningSegment_MultipleSubpaths verifies that each MoveTo segment is
// processed independently and both are simplified correctly.
func TestRunningSegment_MultipleSubpaths(t *testing.T) {
	var p geom.Path
	// Subpath 1: near-collinear, should collapse to 2 vertices
	p.MoveTo(geom.Pt{X: 0, Y: 0})
	p.LineTo(geom.Pt{X: 1, Y: 0.01})
	p.LineTo(geom.Pt{X: 2, Y: 0})
	// Subpath 2: near-collinear, should collapse to 2 vertices
	p.MoveTo(geom.Pt{X: 10, Y: 0})
	p.LineTo(geom.Pt{X: 11, Y: 0.01})
	p.LineTo(geom.Pt{X: 12, Y: 0})

	got := simplifyLinePath(p, 0.05)

	movetos := 0
	for _, c := range got.C {
		if c == geom.MoveTo {
			movetos++
		}
	}
	if movetos != 2 {
		t.Fatalf("expected 2 subpaths (MoveTos), got %d in commands %v", movetos, got.C)
	}
	if len(got.V) != 4 {
		t.Fatalf("expected 4 total vertices (2 per subpath), got %d: %v", len(got.V), got.V)
	}
}
