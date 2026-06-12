package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAddMollweideAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}

	if got := ax.ProjectionName(); got != "mollweide" {
		t.Fatalf("projection name = %q, want mollweide", got)
	}
	if ax.ShowFrame {
		t.Fatal("mollweide axes should disable rectangular frame fallback")
	}

	xMin, xMax := ax.XScale.Domain()
	if !approx(xMin, -math.Pi, 1e-9) || !approx(xMax, math.Pi, 1e-9) {
		t.Fatalf("longitude domain = (%v, %v), want (-pi, pi)", xMin, xMax)
	}
	yMin, yMax := ax.YScale.Domain()
	if !approx(yMin, -math.Pi/2, 1e-9) || !approx(yMax, math.Pi/2, 1e-9) {
		t.Fatalf("latitude domain = (%v, %v), want (-pi/2, pi/2)", yMin, yMax)
	}

	if !ax.ContainsDisplayPoint(geom.Pt{X: 200, Y: 120}) {
		t.Fatal("mollweide axes should contain its center")
	}
	if ax.ContainsDisplayPoint(geom.Pt{X: 10, Y: 10}) {
		t.Fatal("mollweide axes should reject layout corners outside the oval frame")
	}
}

func TestMollweideTransformRoundTrip(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	points := []geom.Pt{
		{X: 0, Y: 0},
		{X: math.Pi / 3, Y: math.Pi / 6},
		{X: -2 * math.Pi / 3, Y: -math.Pi / 4},
	}
	for _, pt := range points {
		pixel := ctx.DataToPixel.Apply(pt)
		got, ok := ax.PixelToData(pixel)
		if !ok {
			t.Fatalf("PixelToData(%+v) failed", pixel)
		}
		if !approx(got.X, pt.X, 1e-6) || !approx(got.Y, pt.Y, 1e-6) {
			t.Fatalf("round trip = %+v, want %+v", got, pt)
		}
	}
}

func TestMollweideFrameAndGridUseSampledCurves(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	longitudeGrid := NewGrid(AxisBottom)
	longitudeGrid.Locator = FixedLocator{TicksList: []float64{0}}
	longitudeGrid.Minor = false

	latitudeGrid := NewGrid(AxisLeft)
	latitudeGrid.Locator = FixedLocator{TicksList: []float64{math.Pi / 6}}
	latitudeGrid.Minor = false

	r := &recordingRenderer{}
	longitudeGrid.Draw(r, ctx)
	latitudeGrid.Draw(r, ctx)
	ax.XAxis.Draw(r, ctx)

	if len(r.pathCalls) < 3 {
		t.Fatalf("expected longitude, latitude, and frame paths, got %d", len(r.pathCalls))
	}

	var foundLongitude, foundLatitude, foundFrame bool
	for _, call := range r.pathCalls {
		if len(call.path.V) == geoGridSegments+1 {
			first := call.path.V[0]
			last := call.path.V[len(call.path.V)-1]
			if approx(first.X, last.X, 1e-6) && !approx(first.Y, last.Y, 1e-6) {
				foundLongitude = true
			}
			if !approx(first.X, last.X, 1e-6) && approx(first.Y, last.Y, 1e-6) {
				foundLatitude = true
			}
		}
		if len(call.path.C) > 0 && call.path.C[len(call.path.C)-1] == geom.ClosePath && len(call.path.V) >= geoFrameSegments {
			foundFrame = true
		}
	}

	if !foundLongitude {
		t.Fatal("expected sampled meridian grid path")
	}
	if !foundLatitude {
		t.Fatal("expected sampled parallel grid path")
	}
	if !foundFrame {
		t.Fatal("expected closed oval frame path")
	}
}

func TestMollweideDefaultsUseMatplotlibGeoTickLabels(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}

	xTicks := ax.XAxis.Locator.Ticks(-math.Pi, math.Pi, 8)
	yTicks := ax.YAxis.Locator.Ticks(-math.Pi/2, math.Pi/2, 8)
	if got, want := len(xTicks), 11; got != want {
		t.Fatalf("longitude tick count = %d, want %d", got, want)
	}
	if got, want := len(yTicks), 11; got != want {
		t.Fatalf("latitude tick count = %d, want %d", got, want)
	}
	// The default geo formatter is matplotlib's ThetaFormatter: degrees with a
	// trailing degree sign (projections/geo.py).
	if got := formatTickLabel(ax.XAxis.Formatter, xTicks[0], 0, xTicks); got != "-150°" {
		t.Fatalf("first longitude label = %q, want -150°", got)
	}
	if got := formatTickLabel(ax.YAxis.Formatter, yTicks[len(yTicks)-1], len(yTicks)-1, yTicks); got != "75°" {
		t.Fatalf("last latitude label = %q, want 75°", got)
	}
}

func TestMollweideXAxisTickLabelsUseGeoTextTransform(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ax.XAxis.Locator = FixedLocator{TicksList: []float64{0}}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axisLabelRecordingRenderer{}

	ax.XAxis.DrawTickLabels(r, ctx)

	if got, want := r.texts, []string{"0°"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("x tick labels = %v, want %v", got, want)
	}
	pos := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 0})
	layout := measureSingleLineTextLayout(r, "0°", tickLabelFontSize(ax.XAxis, ctx), ctx.RC.FontKey)
	want := alignedSingleLineOrigin(
		geom.Pt{X: pos.X, Y: pos.Y + geoXAxisLabelPadPx},
		layout,
		TextAlignCenter,
		textLayoutVAlignBottom,
	)
	if got := r.origins[0]; !approx(got.X, want.X, 1e-9) || !approx(got.Y, want.Y, 1e-9) {
		t.Fatalf("center x tick label origin = %+v, want %+v", got, want)
	}
	if !(r.origins[0].Y > pos.Y) {
		t.Fatalf("geo x tick label should be above equator, got origin=%+v equator=%+v", r.origins[0], pos)
	}
}

func TestGeoXAxisLabelUsesFrameBottomNotEquatorTickBounds(t *testing.T) {
	fig := NewFigure(720, 420)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.86},
	}, "aitoff")
	if err != nil {
		t.Fatalf("AddAxesProjection(aitoff): %v", err)
	}
	ax.SetXLabel("longitude")
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	px := ax.adjustedLayout(fig)
	r := &axesLabelRecordingRenderer{}

	anchor, _ := xLabelAnchorPoint(ax, r, ctx, px, AxisBottom, figureTextAlignment{})
	wantY := px.Min.Y - axisLabelPadPx(ctx)
	if !approx(anchor.Y, wantY, 1e-9) {
		t.Fatalf("geo x-label anchor y = %v, want frame bottom minus pad %v", anchor.Y, wantY)
	}

	equatorY := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 0}).Y
	if !(anchor.Y < equatorY) {
		t.Fatalf("geo x-label anchor should be below equator tick labels, got anchor=%v equator=%v", anchor.Y, equatorY)
	}
}

func TestGeoProjection_CarriesNamedTransform(t *testing.T) {
	p := newMollweideProjection()
	if p.transform == nil {
		t.Fatal("expected non-nil transform on geoProjection")
	}
	if _, ok := p.transform.(mollweideDataTransform); !ok {
		t.Fatalf("transform = %T, want mollweideDataTransform", p.transform)
	}
}

func TestMollweideGridPreservesConfiguredRGBA(t *testing.T) {
	fig := NewFigure(400, 240)
	ax, err := fig.AddAxesProjection(unitRect(), "mollweide")
	if err != nil {
		t.Fatalf("AddAxesProjection(mollweide): %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	grid := NewGrid(AxisBottom)
	grid.Locator = FixedLocator{TicksList: []float64{0}}
	grid.Color = render.Color{R: 0.78, G: 0.80, B: 0.84, A: 0.55}
	grid.LineWidth = 0.8

	r := &recordingRenderer{}
	grid.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("grid path call count = %d, want 1", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Stroke, grid.Color; got != want {
		t.Fatalf("grid stroke color = %+v, want %+v", got, want)
	}
	if got, want := r.pathCalls[0].paint.LineWidth, grid.LineWidth; got != want {
		t.Fatalf("grid line width = %v, want %v", got, want)
	}
}
