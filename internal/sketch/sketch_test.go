package sketch

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

// TestRNGMatchesMSVCConstants checks the LCG against an independent uint32
// accumulator, validating the constants and the implicit 2^32 modulus/divisor.
func TestRNGMatchesMSVCConstants(t *testing.T) {
	r := rng{seed: 0}
	var seed uint32
	for i := range 5 {
		seed = 214013*seed + 2531011
		want := float64(seed) / 4294967296.0
		got := r.double()
		if got != want {
			t.Fatalf("draw %d: got %v want %v", i, got, want)
		}
		if got < 0 || got >= 1 {
			t.Fatalf("draw %d out of [0,1): %v", i, got)
		}
	}
}

func horizontal(x0, x1, y float64) geom.Path {
	var p geom.Path
	p.MoveTo(geom.Pt{X: x0, Y: y})
	p.LineTo(geom.Pt{X: x1, Y: y})
	return p
}

func TestApplyScaleZeroIsNoOp(t *testing.T) {
	in := horizontal(0, 100, 5)
	out := Apply(in, 0, 100, 2)
	if len(out.C) != len(in.C) || len(out.V) != len(in.V) {
		t.Fatalf("scale 0 changed path shape: %d/%d cmds, %d/%d verts",
			len(out.C), len(in.C), len(out.V), len(in.V))
	}
	for i := range in.V {
		if out.V[i] != in.V[i] {
			t.Fatalf("scale 0 moved vertex %d: %v != %v", i, out.V[i], in.V[i])
		}
	}
}

// TestApplyZeroRandomnessNoDisplacement: with randomness 0 the wiggle amplitude
// collapses to zero (p_scale=0 -> sin=0), so the densified polyline stays on the
// original line.
func TestApplyZeroRandomnessNoDisplacement(t *testing.T) {
	out := Apply(horizontal(0, 100, 7), 10, 100, 0)
	if len(out.V) < 10 {
		t.Fatalf("expected dense resampling, got %d verts", len(out.V))
	}
	for i, v := range out.V {
		if math.Abs(v.Y-7) > 1e-9 {
			t.Fatalf("vertex %d displaced off line: y=%v", i, v.Y)
		}
	}
}

func TestApplyDeterministic(t *testing.T) {
	in := horizontal(0, 137, 3)
	a := Apply(in, 4, 80, 2)
	b := Apply(in, 4, 80, 2)
	if len(a.V) != len(b.V) {
		t.Fatalf("nondeterministic vertex count: %d != %d", len(a.V), len(b.V))
	}
	for i := range a.V {
		if a.V[i] != b.V[i] {
			t.Fatalf("nondeterministic vertex %d: %v != %v", i, a.V[i], b.V[i])
		}
	}
}

// TestApplyBoundedAndAnchored: perturbation magnitude never exceeds scale, the
// MoveTo point is left untouched, and a length-100 line yields ~100 samples.
func TestApplyBoundedAndAnchored(t *testing.T) {
	scale := 6.0
	out := Apply(horizontal(0, 100, 0), scale, 100, 2)
	if out.C[0] != geom.MoveTo || out.V[0] != (geom.Pt{X: 0, Y: 0}) {
		t.Fatalf("MoveTo anchor altered: %v", out.V[0])
	}
	for i, v := range out.V {
		if math.Abs(v.Y) > scale+1e-9 {
			t.Fatalf("vertex %d exceeds scale: y=%v > %v", i, v.Y, scale)
		}
	}
	if n := len(out.V); n < 90 || n > 110 {
		t.Fatalf("expected ~100 samples for a 100px line, got %d", n)
	}
}

// TestApplyFirstPerturbedVertex pins the integrated pipeline (segmentator +
// RNG + phase update + perpendicular offset) against an independent computation
// for the first interior vertex of a horizontal segment.
func TestApplyFirstPerturbedVertex(t *testing.T) {
	const scale, length, randomness = 10.0, 100.0, 2.0
	out := Apply(horizontal(0, 100, 0), scale, length, randomness)

	// First segmentator line vertex is at parameter ddl=0.01 -> x=1, y=0.
	// Independent expectation: rand0 is the first LCG draw; p=exp(rand0*2ln2);
	// the segment is horizontal so the offset is purely vertical:
	// y = sin(p * 2pi/(length*randomness)) * scale.
	rand0 := float64(uint32(214013*0+2531011)) / 4294967296.0
	p := math.Exp(rand0 * 2 * math.Log(randomness))
	wantY := math.Sin(p*2*math.Pi/(length*randomness)) * scale

	if math.Abs(out.V[1].X-1) > 1e-9 {
		t.Fatalf("first interior x: got %v want 1", out.V[1].X)
	}
	if math.Abs(out.V[1].Y-wantY) > 1e-9 {
		t.Fatalf("first interior y: got %v want %v", out.V[1].Y, wantY)
	}
}

// TestApplyFlattensCurves: a cubic is flattened to a polyline (no curve verbs
// survive) and every command is MoveTo/LineTo.
func TestApplyFlattensCurves(t *testing.T) {
	var in geom.Path
	in.MoveTo(geom.Pt{X: 0, Y: 0})
	in.CubicTo(geom.Pt{X: 20, Y: 40}, geom.Pt{X: 60, Y: -40}, geom.Pt{X: 100, Y: 0})
	out := Apply(in, 3, 100, 2)
	for i, c := range out.C {
		if c != geom.MoveTo && c != geom.LineTo {
			t.Fatalf("cmd %d is not a line segment: %v", i, c)
		}
	}
	if len(out.V) < 20 {
		t.Fatalf("curve undersampled: %d verts", len(out.V))
	}
}
