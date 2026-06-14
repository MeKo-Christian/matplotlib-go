package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAitoffForward_Origin(t *testing.T) {
	out := aitoffDataTransform{}.Apply(geom.Pt{X: 0, Y: 0})
	if math.Abs(out.X-0.5) > 1e-12 || math.Abs(out.Y-0.5) > 1e-12 {
		t.Fatalf("(0,0) -> %v, want (0.5, 0.5)", out)
	}
}

func TestAitoffForward_NonOrigin(t *testing.T) {
	out := aitoffDataTransform{}.Apply(geom.Pt{X: math.Pi / 2, Y: 0})
	if math.Abs(out.Y-0.5) > 1e-12 {
		t.Fatalf("y for (π/2, 0) = %v, want 0.5", out.Y)
	}
	if out.X <= 0.5 || out.X > 1.0 {
		t.Fatalf("x for (π/2, 0) = %v, want in (0.5, 1.0]", out.X)
	}
}

func TestAitoffInverseUnsupported(t *testing.T) {
	_, ok := aitoffDataTransform{}.Invert(geom.Pt{X: 0.5, Y: 0.5})
	if ok {
		t.Fatal("Aitoff inverse should report unsupported (parity with matplotlib)")
	}
}

func TestAitoffProjectionRegistered(t *testing.T) {
	if _, err := lookupProjection("aitoff"); err != nil {
		t.Fatalf("aitoff projection not registered: %v", err)
	}
}
