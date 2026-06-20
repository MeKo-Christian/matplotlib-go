package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAddRadarAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"Speed", "Power", "Range", "Cost"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}

	if got := ax.ProjectionName(); got != "radar" {
		t.Fatalf("projection name = %q, want radar", got)
	}
	if ax.ShowFrame {
		t.Fatal("radar axes should disable rectangular frame fallback")
	}
	if !approx(ax.XAxis.TickSize, defaultTickSizePx, 1e-12) {
		t.Fatalf("radar theta tick size = %v, want Matplotlib default %v even though tick marks are hidden", ax.XAxis.TickSize, defaultTickSizePx)
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))
	top := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 1})
	wantTop := geom.Pt{X: center.X, Y: center.Y + radius}
	if !approx(top.X, wantTop.X, 1e-6) || !approx(top.Y, wantTop.Y, 1e-6) {
		t.Fatalf("theta=0 point = %+v, want %+v", top, wantTop)
	}
	left := ctx.DataToPixel.Apply(geom.Pt{X: math.Pi / 2, Y: 1})
	wantLeft := geom.Pt{X: center.X - radius, Y: center.Y}
	if !approx(left.X, wantLeft.X, 1e-6) || !approx(left.Y, wantLeft.Y, 1e-6) {
		t.Fatalf("theta=pi/2 point = %+v, want %+v", left, wantLeft)
	}

	r := &polarTextRenderer{}
	ax.XAxis.DrawTickLabels(r, ctx)
	if len(r.texts) != 4 {
		t.Fatalf("expected 4 radar spoke labels, got %d (%v)", len(r.texts), r.texts)
	}
	if r.texts[0] != "Speed" || r.texts[1] != "Power" || r.texts[2] != "Range" || r.texts[3] != "Cost" {
		t.Fatalf("radar spoke labels = %v", r.texts)
	}
}

func TestRadarFrameAndGridUsePolygonGeometry(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"A", "B", "C", "D", "E"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	grid := NewGrid(AxisLeft)
	grid.Locator = FixedLocator{TicksList: []float64{0.5}}
	grid.Minor = false

	r := &recordingRenderer{}
	grid.Draw(r, ctx)
	ax.XAxis.Draw(r, ctx)

	center, radius := polarCenterAndRadius(ax.adjustedLayout(fig))
	wantTop := geom.Pt{X: center.X, Y: center.Y + radius}
	var foundOuterPentagon bool
	for _, call := range r.pathCalls {
		if len(call.path.C) == 6 && call.path.C[len(call.path.C)-1] == geom.ClosePath && len(call.path.V) == 5 {
			if approx(call.path.V[0].X, wantTop.X, 1e-6) && approx(call.path.V[0].Y, wantTop.Y, 1e-6) {
				foundOuterPentagon = true
			}
		}
	}
	if !foundOuterPentagon {
		t.Fatal("expected radar outer spine to be a five-sided polygon with its first vertex at theta=0")
	}
}

func TestRadarRadialLabelsUseMatplotlibDefaultOffsetFromNorth(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"A", "B", "C", "D", "E"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ax.YAxis.Locator = FixedLocator{TicksList: []float64{0.5}}
	ax.YAxis.MinorLocator = nil
	ax.YAxis.Formatter = FuncFormatter(func(float64) string { return "radial" })

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	center, outerRadius := polarCenterAndRadius(ax.adjustedLayout(fig))
	r := &polarTextRenderer{}

	ax.YAxis.DrawTickLabels(r, ctx)

	if len(r.texts) != 1 {
		t.Fatalf("expected one radial tick label, got %d (%v)", len(r.texts), r.texts)
	}

	labelAngle := math.Pi/2 + defaultPolarRadialLabelAngle
	fontSize := tickLabelFontSize(ax.YAxis, ctx)
	layout := measureSingleLineTextLayout(r, "radial", fontSize, ctx.RC.FontKey)
	anchor := polarPixelPoint(center, outerRadius*0.5, labelAngle)
	// Full-circle polar axes anchor radial labels with ha='left', va='bottom'
	// (matplotlib PolarAxes.get_yaxis_text1_transform).
	hAlign, vAlign := polarRadialTickLabelAlignments(true, labelAngle)
	wantOrigin := alignedSingleLineOrigin(anchor, layout, hAlign, vAlign)

	if !approx(r.origins[0].X, wantOrigin.X, 1e-6) || !approx(r.origins[0].Y, wantOrigin.Y, 1e-6) {
		t.Fatalf("radial tick label origin = %+v, want %+v", r.origins[0], wantOrigin)
	}
}

func TestRadarThetaLabelsUseMatplotlibCenteredPadding(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"Top", "Left", "Bottom", "Right"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	center, outerRadius := polarCenterAndRadius(ax.adjustedLayout(fig))
	r := &polarTextRenderer{}

	ax.XAxis.DrawTickLabels(r, ctx)

	var origin geom.Pt
	found := false
	for i, label := range r.texts {
		if label == "Left" {
			origin = r.origins[i]
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected Left theta label, got %v", r.texts)
	}

	fontSize := tickLabelFontSize(ax.XAxis, ctx)
	layout := measureSingleLineTextLayout(r, "Left", fontSize, ctx.RC.FontKey)
	padPx := defaultTickSizePx + pointsToPixels(ctx.RC, defaultTickPadPt+7)
	anchor := polarPixelPoint(center, outerRadius+padPx, math.Pi)
	wantOrigin := alignedSingleLineOrigin(anchor, layout, TextAlignCenter, textLayoutVAlignCenter)

	if !approx(origin.X, wantOrigin.X, 1e-6) || !approx(origin.Y, wantOrigin.Y, 1e-6) {
		t.Fatalf("theta label origin = %+v, want centered Matplotlib padding origin %+v", origin, wantOrigin)
	}
}

func TestRadarHidesRadialSpineAndTicks(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"A", "B", "C", "D", "E"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ax.YAxis.Locator = FixedLocator{TicksList: []float64{0.5}}
	ax.YAxis.MinorLocator = nil

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &recordingRenderer{}

	ax.YAxis.Draw(r, ctx)

	if len(r.pathCalls) != 0 {
		t.Fatalf("radar radial axis drew %d paths, want none", len(r.pathCalls))
	}
}

func TestRadarTitleClearsTopThetaLabel(t *testing.T) {
	fig := NewFigure(400, 400)
	ax, err := fig.AddRadarAxes(unitRect(), []string{"Speed", "Power", "Range", "Handling", "Comfort"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ax.SetTitle("Radar Projection")

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &polarTextRenderer{}
	drawAxesLabels(ax, r, ctx, ax.adjustedLayout(fig), figureTextAlignment{})

	if len(r.texts) != 1 || r.texts[0] != "Radar Projection" {
		t.Fatalf("unexpected title draws: %v", r.texts)
	}
	titleLayout := measureSingleLineTextLayout(r, "Radar Projection", titleFontSize(ctx), ctx.RC.FontKey)
	titleBounds, ok := textInkRect(r.origins[0], titleLayout)
	if !ok {
		t.Fatal("expected title bounds")
	}
	thetaBounds, ok := axisTickLabelBounds(ax.XAxis, r, ctx)
	if !ok {
		t.Fatal("expected theta tick label bounds")
	}
	// Display space is y-up: the title clears the top theta label when its bottom
	// edge (Min.Y) sits at or above the theta labels' top edge (Max.Y).
	if titleBounds.Min.Y < thetaBounds.Max.Y {
		t.Fatalf("title overlaps top theta label: title=%+v theta=%+v", titleBounds, thetaBounds)
	}
}

func TestRadarTitleBaselineUsesUnsnappedMatplotlibTitlePad(t *testing.T) {
	fig := NewFigure(640, 640)
	ax, err := fig.AddRadarAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.12},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	}, []string{"Speed", "Power", "Range", "Handling", "Comfort"})
	if err != nil {
		t.Fatalf("AddRadarAxes: %v", err)
	}
	ax.SetTitle("Radar Projection")

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &polarTextRenderer{}
	drawAxesLabels(ax, r, ctx, ax.adjustedLayout(fig), figureTextAlignment{})

	if len(r.texts) != 1 || r.texts[0] != "Radar Projection" {
		t.Fatalf("unexpected title draws: %v", r.texts)
	}

	wantY := titleTopExtent(ax, r, ctx, ax.adjustedLayout(fig)) + pointsToPixels(ctx.RC, 6)
	if !approx(r.origins[0].Y, wantY, 1e-9) {
		t.Fatalf("radar title baseline Y = %v, want unsnapped top extent plus titlepad %v", r.origins[0].Y, wantY)
	}
}

func TestRadarVariableCountValidation(t *testing.T) {
	fig := NewFigure(400, 400)
	if _, err := fig.AddRadarAxes(unitRect(), []string{"A", "B"}); err == nil {
		t.Fatal("expected AddRadarAxes to reject fewer than 3 labels")
	}

	ax, err := fig.AddRadarAxes(unitRect(), nil)
	if err != nil {
		t.Fatalf("AddRadarAxes with defaults: %v", err)
	}
	if err := ax.SetRadarVariableCount(6); err != nil {
		t.Fatalf("SetRadarVariableCount: %v", err)
	}
	ticks := ax.XAxis.Locator.Ticks(0, 2*math.Pi, 10)
	if len(ticks) != 6 {
		t.Fatalf("radar tick count = %d, want 6", len(ticks))
	}
}
