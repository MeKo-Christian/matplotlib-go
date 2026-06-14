package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestHammerForward_Origin(t *testing.T) {
	out := hammerDataTransform{}.Apply(geom.Pt{X: 0, Y: 0})
	if math.Abs(out.X-0.5) > 1e-12 || math.Abs(out.Y-0.5) > 1e-12 {
		t.Fatalf("(0,0) -> %v, want (0.5, 0.5)", out)
	}
}

func TestHammerRoundTrip(t *testing.T) {
	cases := []geom.Pt{
		{X: 0, Y: 0},
		{X: math.Pi / 3, Y: math.Pi / 6},
		{X: -math.Pi / 4, Y: -math.Pi / 5},
		{X: math.Pi / 2, Y: math.Pi / 4},
	}
	tr := hammerDataTransform{}
	for _, in := range cases {
		out := tr.Apply(in)
		back, ok := tr.Invert(out)
		if !ok {
			t.Fatalf("invert failed for %v", in)
		}
		if math.Abs(back.X-in.X) > 1e-9 || math.Abs(back.Y-in.Y) > 1e-9 {
			t.Fatalf("round trip %v -> %v -> %v", in, out, back)
		}
	}
}

func TestHammerProjectionRegistered(t *testing.T) {
	if _, err := lookupProjection("hammer"); err != nil {
		t.Fatalf("hammer projection not registered: %v", err)
	}
}
