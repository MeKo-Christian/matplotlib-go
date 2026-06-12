package core

import (
	"fmt"
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

// TestLambertTransformProbe is an env-gated parity diagnostic for the
// geo_lambert_axes fixture. Compare its log output against Matplotlib's
// GeoAxes/LambertAxes transform stack in projections/geo.py.
func TestLambertTransformProbe(t *testing.T) {
	if os.Getenv("MPL_GO_GEO_PROBE") == "" {
		t.Skip("set MPL_GO_GEO_PROBE=1 to run")
	}
	fig := NewFigure(520, 520)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.10},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	}, "lambert")
	if err != nil {
		t.Fatal(err)
	}
	ax.SetTitle("Lambert Projection")
	ax.SetXLabel("longitude")
	ax.SetYLabel("latitude")
	ax.XAxis.Locator = FixedLocator{TicksList: []float64{
		-120 * math.Pi / 180,
		-90 * math.Pi / 180,
		-60 * math.Pi / 180,
		-30 * math.Pi / 180,
		0,
		30 * math.Pi / 180,
		60 * math.Pi / 180,
		90 * math.Pi / 180,
		120 * math.Pi / 180,
	}}
	ax.XAxis.Formatter = FuncFormatter(func(x float64) string {
		return fmt.Sprintf("%.0f", math.Round(x*180/math.Pi))
	})

	r := &axisLabelRecordingRenderer{}
	vp := geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 520, Y: 520}}
	prepareFigureLayout(fig, r, vp)
	px := ax.adjustedLayout(fig)
	ctx := newAxesDrawContext(ax, fig, vp, px)
	t.Logf("axes px rect: min=(%.4f, %.4f) size=(%.4f, %.4f)", px.Min.X, px.Min.Y, px.W(), px.H())
	for _, data := range []geom.Pt{
		{X: 0, Y: 0},
		{X: -math.Pi, Y: 0},
		{X: math.Pi, Y: 0},
		{X: 0, Y: math.Pi / 2},
		{X: 0, Y: -math.Pi / 2},
		{X: math.Pi / 2, Y: 0.35},
	} {
		pt := ctx.DataToPixel.Apply(data)
		t.Logf("transData (%.12g, %.12g) -> (%.4f, %.4f)", data.X, data.Y, pt.X, pt.Y)
	}
	for _, data := range []geom.Pt{{X: 0, Y: -geoLongitudeGridCap}, {X: 0, Y: geoLongitudeGridCap}} {
		pt := ctx.DataToPixel.Apply(data)
		t.Logf("xaxis grid cap (%.12g, %.12g) -> (%.4f, %.4f)", data.X, data.Y, pt.X, pt.Y)
	}
	anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, AxisBottom, figureTextAlignment{})
	t.Logf("xlabel anchor: (%.4f, %.4f) valign=%v", anchor.X, anchor.Y, vAlign)
}
