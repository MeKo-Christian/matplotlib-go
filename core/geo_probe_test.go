package core

import (
	"math"
	"os"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TestMollweideGridRowProbe is a temporary diagnostic (env-gated): it prints
// the raw display-y of each latitude gridline for the geo_mollweide_axes
// fixture geometry, for comparison against matplotlib's transData output.
func TestMollweideGridRowProbe(t *testing.T) {
	if os.Getenv("MPL_GO_GEO_PROBE") == "" {
		t.Skip("set MPL_GO_GEO_PROBE=1 to run")
	}
	fig := NewFigure(720, 420)
	ax, err := fig.AddAxesProjection(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.92, Y: 0.86}}, "mollweide")
	if err != nil {
		t.Fatal(err)
	}
	_ = ax
	r := &render.NullRenderer{}
	vp := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 720, Y: 420}}
	prepareFigureLayout(fig, r, vp)
	px := ax.adjustedLayout(fig)
	ctx := newAxesDrawContext(ax, fig, vp, px)
	t.Logf("axes px rect: %+v", px)
	for latDeg := -75; latDeg <= 75; latDeg += 15 {
		lat := float64(latDeg) * math.Pi / 180
		pt := ctx.DataToPixel.Apply(geom.Pt{X: 0.5, Y: lat})
		t.Logf("lat %4d: image_y=%.4f x=%.4f", latDeg, pt.Y, pt.X)
	}
}
