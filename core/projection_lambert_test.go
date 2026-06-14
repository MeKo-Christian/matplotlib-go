package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestLambertForward_AtCenter_IsOrigin(t *testing.T) {
	tr := lambertDataTransform{centerLon: 0.5, centerLat: 0.3}
	out := tr.Apply(geom.Pt{X: 0.5, Y: 0.3})
	if math.Abs(out.X-0.5) > 1e-9 || math.Abs(out.Y-0.5) > 1e-9 {
		t.Fatalf("center -> %v, want (0.5, 0.5)", out)
	}
}

func TestLambertRoundTrip(t *testing.T) {
	tr := lambertDataTransform{centerLon: 0, centerLat: 0}
	cases := []geom.Pt{
		{X: math.Pi / 4, Y: math.Pi / 6},
		{X: -math.Pi / 6, Y: -math.Pi / 8},
	}
	for _, in := range cases {
		out := tr.Apply(in)
		back, ok := tr.Invert(out)
		if !ok {
			t.Fatalf("invert failed for %v", in)
		}
		if math.Abs(back.X-in.X) > 1e-7 || math.Abs(back.Y-in.Y) > 1e-7 {
			t.Fatalf("round trip %v -> %v -> %v", in, out, back)
		}
	}
}

func TestLambertProjectionRegistered(t *testing.T) {
	if _, err := lookupProjection("lambert"); err != nil {
		t.Fatalf("lambert projection not registered: %v", err)
	}
}

func TestLambertAxesUseEqualBoxAspect(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "lambert")
	if err != nil {
		t.Fatalf("AddAxesProjection(lambert): %v", err)
	}

	if !approx(ax.boxAspect, 1, 1e-12) {
		t.Fatalf("lambert box aspect = %v, want 1", ax.boxAspect)
	}
}

func TestLambertDefaultsHideLatitudeTickLabels(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "lambert")
	if err != nil {
		t.Fatalf("AddAxesProjection(lambert): %v", err)
	}

	if _, ok := ax.YAxis.Formatter.(NullFormatter); !ok {
		t.Fatalf("lambert y-axis formatter = %T, want NullFormatter", ax.YAxis.Formatter)
	}
}
