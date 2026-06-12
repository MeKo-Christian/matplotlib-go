package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestAddSkewXAxesConfiguresProjection(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}

	if got := ax.ProjectionName(); got != "skewx" {
		t.Fatalf("projection name = %q, want skewx", got)
	}
	if ax.XAxisTop == nil {
		t.Fatal("skewx axes should configure a top x axis")
	}

	xMin, xMax := ax.XScale.Domain()
	if !approx(xMin, -40, 1e-9) || !approx(xMax, 50, 1e-9) {
		t.Fatalf("temperature domain = (%v, %v), want (-40, 50)", xMin, xMax)
	}
	yMin, yMax := ax.YScale.Domain()
	if !approx(yMin, 1050, 1e-9) || !approx(yMax, 100, 1e-9) {
		t.Fatalf("pressure domain = (%v, %v), want (1050, 100)", yMin, yMax)
	}
	if !ax.YInverted() {
		t.Fatal("skewx pressure axis should decrease upward")
	}
}

func TestSkewXTopAxisUsesSpineWithoutTickLabels(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	if ax.XAxisTop == nil {
		t.Fatal("skewx axes should configure a top spine axis")
	}

	if !ax.XAxisTop.ShowSpine {
		t.Fatal("skewx top axis should keep the top spine visible")
	}
	if ax.XAxisTop.ShowTicks {
		t.Fatal("skewx top axis should hide tick marks by default")
	}
	if ax.XAxisTop.ShowLabels {
		t.Fatal("skewx top axis should hide tick labels by default")
	}
}

func TestSkewXTransformRoundTrip(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	if err := ax.SetSkewXAngle(35); err != nil {
		t.Fatalf("SetSkewXAngle: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	points := []geom.Pt{
		{X: 0, Y: 1000},
		{X: 12, Y: 700},
		{X: -18, Y: 300},
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

func TestSkewXTransformKeepsLowerPressureEdgeUnshifted(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	bottom := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 1050})
	wantBottomX := ax.effectiveXScale().Fwd(0)
	if !approx(bottom.X, wantBottomX, 1e-9) {
		t.Fatalf("lower pressure edge x = %v, want unshifted %v", bottom.X, wantBottomX)
	}

	top := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 200})
	if !(top.X > bottom.X) {
		t.Fatalf("upper pressure levels should shift right: bottom=%+v top=%+v", bottom, top)
	}
}

func TestSkewXAngleValidationAndEffect(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}

	if err := ax.SetSkewXAngle(math.Inf(1)); err == nil {
		t.Fatal("expected infinite skewx angle to be rejected")
	}
	if err := ax.SetSkewXAngle(90); err == nil {
		t.Fatal("expected vertical skewx angle to be rejected")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	bottom := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 1000})
	top := ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 200})
	if !(top.X > bottom.X) {
		t.Fatalf("default skew should shift upper pressure levels right: bottom=%+v top=%+v", bottom, top)
	}

	if err := ax.SetSkewXAngle(-25); err != nil {
		t.Fatalf("SetSkewXAngle(-25): %v", err)
	}
	ctx = newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	bottom = ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 1000})
	top = ctx.TransProjection().Apply(geom.Pt{X: 0, Y: 200})
	if !(top.X < bottom.X) {
		t.Fatalf("negative skew should shift upper pressure levels left: bottom=%+v top=%+v", bottom, top)
	}
}

func TestSkewXProjectionRejectsAngleOnOtherAxes(t *testing.T) {
	ax := NewFigure(480, 360).AddAxes(unitRect())
	if err := ax.SetSkewXAngle(30); err == nil {
		t.Fatal("expected SetSkewXAngle on rectilinear axes to fail")
	}
}

func TestSkewXYGridSpansAxesWithoutSkew(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	ax.YAxis.Locator = FixedLocator{TicksList: []float64{300}}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	grid := NewGrid(AxisLeft)
	grid.Minor = false

	r := &recordingRenderer{}
	grid.Draw(r, ctx)

	if got := len(r.pathCalls); got != 1 {
		t.Fatalf("expected 1 major pressure grid path, got %d", got)
	}
	path := r.pathCalls[0].path
	if got := len(path.V); got != 2 {
		t.Fatalf("pressure grid vertex count = %d, want 2", got)
	}

	y := ctx.DataToPixel.YScale.Fwd(300)
	wantLeft := ctx.TransAxes().Apply(geom.Pt{X: 0, Y: y})
	wantRight := ctx.TransAxes().Apply(geom.Pt{X: 1, Y: y})
	if !approx(path.V[0].X, wantLeft.X, 1e-9) || !approx(path.V[0].Y, wantLeft.Y, 1e-9) {
		t.Fatalf("pressure grid start = %+v, want %+v", path.V[0], wantLeft)
	}
	if !approx(path.V[1].X, wantRight.X, 1e-9) || !approx(path.V[1].Y, wantRight.Y, 1e-9) {
		t.Fatalf("pressure grid end = %+v, want %+v", path.V[1], wantRight)
	}
}

func TestSkewXXGridUsesUpperViewInterval(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	grid := NewGrid(AxisBottom)
	grid.Minor = false

	r := &recordingRenderer{}
	grid.Draw(r, ctx)

	if got, want := len(r.pathCalls), 15; got != want {
		t.Fatalf("skewx x-grid path count = %d, want %d including top-edge ticks", got, want)
	}
}

func TestSkewXYTickLabelsStayOnLeftSpine(t *testing.T) {
	fig := NewFigure(480, 360)
	ax, err := fig.AddSkewXAxes(unitRect())
	if err != nil {
		t.Fatalf("AddSkewXAxes: %v", err)
	}
	ax.YAxis.Locator = FixedLocator{TicksList: []float64{200, 1000}}
	ax.YAxis.Formatter = ScalarFormatter{Prec: 0}
	ax.YAxis.MinorLocator = nil
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	r := &polarTextRenderer{}
	ax.YAxis.DrawTickLabels(r, ctx)

	if got := len(r.origins); got != 2 {
		t.Fatalf("skewx y tick labels = %d, want 2", got)
	}
	firstLayout := measureSingleLineTextLayout(r, r.texts[0], tickLabelFontSize(ax.YAxis, ctx), ctx.RC.FontKey)
	secondLayout := measureSingleLineTextLayout(r, r.texts[1], tickLabelFontSize(ax.YAxis, ctx), ctx.RC.FontKey)
	firstRight := r.origins[0].X + firstLayout.Width
	secondRight := r.origins[1].X + secondLayout.Width
	if !approx(firstRight, secondRight, 1e-9) {
		t.Fatalf("skewx y tick label right edges = %v and %v, want same left spine column", firstRight, secondRight)
	}
}
