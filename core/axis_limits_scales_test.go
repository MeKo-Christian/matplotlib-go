package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAxis_Draw(t *testing.T) {
	// Test drawing a basic X axis
	axis := NewXAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_YAxis(t *testing.T) {
	// Test drawing a basic Y axis
	axis := NewYAxis()

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_CustomSettings(t *testing.T) {
	// Test axis with custom settings
	axis := &Axis{
		Side:       AxisTop,
		Locator:    LinearLocator{},
		Formatter:  ScalarFormatter{Prec: 2},
		Color:      render.Color{R: 1, G: 0, B: 0, A: 1}, // red
		LineWidth:  2.0,
		TickSize:   10.0,
		ShowSpine:  true,
		ShowTicks:  true,
		ShowLabels: false,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxis_DisabledComponents(t *testing.T) {
	// Test axis with components disabled
	axis := NewXAxis()
	axis.ShowSpine = false
	axis.ShowTicks = false
	axis.ShowLabels = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic even with everything disabled
	axis.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestAxes_SetLimits(t *testing.T) {
	// Test the convenience methods for setting limits
	axes := &Axes{
		XScale:     nil,
		YScale:     nil,
		XAxis:      NewXAxis(),
		YAxis:      NewYAxis(),
		XAxisTop:   &Axis{Side: AxisTop},
		YAxisRight: &Axis{Side: AxisRight},
	}

	// Test SetXLim
	axes.SetXLim(-5, 10)
	xMin, xMax := axes.XScale.Domain()
	if xMin != -5 || xMax != 10 {
		t.Errorf("SetXLim failed: expected (-5, 10), got (%v, %v)", xMin, xMax)
	}

	// Test SetYLim
	axes.SetYLim(0, 100)
	yMin, yMax := axes.YScale.Domain()
	if yMin != 0 || yMax != 100 {
		t.Errorf("SetYLim failed: expected (0, 100), got (%v, %v)", yMin, yMax)
	}

	// Test SetXLimLog
	axes.SetXLimLog(1, 1000, 10)
	xMin, xMax = axes.XScale.Domain()
	if xMin != 1 || xMax != 1000 {
		t.Errorf("SetXLimLog failed: expected (1, 1000), got (%v, %v)", xMin, xMax)
	}

	// Check that locator was updated to logarithmic
	if logLoc, ok := axes.XAxis.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetXLimLog should update locator to LogLocator with base 10")
	}
	if logLoc, ok := axes.XAxisTop.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetXLimLog should update top axis locator to LogLocator with base 10")
	}

	axes.SetYLimLog(1, 100, 10)
	if logLoc, ok := axes.YAxis.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetYLimLog should update locator to LogLocator with base 10")
	}
	if logLoc, ok := axes.YAxisRight.Locator.(LogLocator); !ok || logLoc.Base != 10 {
		t.Errorf("SetYLimLog should update right axis locator to LogLocator with base 10")
	}
}

func TestAxes_SetScalePreservesDomainAndConfiguresLogDefaults(t *testing.T) {
	axes := &Axes{
		XScale:   transform.NewLinear(1, 1000),
		XAxis:    NewXAxis(),
		XAxisTop: &Axis{Side: AxisTop},
	}

	err := axes.SetXScale(
		"LOG",
		transform.WithScaleBase(10),
		transform.WithScaleSubs(2, 4, 5),
	)
	if err != nil {
		t.Fatalf("SetXScale(log): %v", err)
	}

	logScale, ok := axes.XScale.(transform.Log)
	if !ok {
		t.Fatalf("x scale type = %T, want transform.Log", axes.XScale)
	}
	xMin, xMax := logScale.Domain()
	if xMin != 1 || xMax != 1000 {
		t.Fatalf("x scale domain = (%v, %v), want (1, 1000)", xMin, xMax)
	}

	loc, ok := axes.XAxis.Locator.(LogLocator)
	if !ok || loc.Base != 10 {
		t.Fatalf("bottom locator = %#v, want log locator base 10", axes.XAxis.Locator)
	}
	minor, ok := axes.XAxis.MinorLocator.(LogLocator)
	if !ok || len(minor.Subs) != 3 {
		t.Fatalf("bottom minor locator = %#v, want log minor locator with subs", axes.XAxis.MinorLocator)
	}
	topLoc, ok := axes.XAxisTop.Locator.(LogLocator)
	if !ok || topLoc.Base != 10 {
		t.Fatalf("top locator = %#v, want log locator base 10", axes.XAxisTop.Locator)
	}
}

func TestAxesSemilogConvenienceWrappers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	if line := ax.SemilogX([]float64{1, 10, 100}, []float64{1, 2, 3}); line == nil {
		t.Fatal("SemilogX() returned nil")
	}
	if _, ok := ax.XScale.(transform.Log); !ok {
		t.Fatalf("SemilogX x scale = %T, want transform.Log", ax.XScale)
	}
	if _, ok := ax.YScale.(transform.Linear); !ok {
		t.Fatalf("SemilogX y scale = %T, want transform.Linear", ax.YScale)
	}

	fig = NewFigure(800, 600)
	ax = fig.AddAxes(unitRect())
	if line := ax.SemilogY([]float64{1, 2, 3}, []float64{1, 10, 100}); line == nil {
		t.Fatal("SemilogY() returned nil")
	}
	if _, ok := ax.YScale.(transform.Log); !ok {
		t.Fatalf("SemilogY y scale = %T, want transform.Log", ax.YScale)
	}

	fig = NewFigure(800, 600)
	ax = fig.AddAxes(unitRect())
	if line := ax.LogLog([]float64{1, 10, 100}, []float64{1, 10, 100}); line == nil {
		t.Fatal("LogLog() returned nil")
	}
	if _, ok := ax.XScale.(transform.Log); !ok {
		t.Fatalf("LogLog x scale = %T, want transform.Log", ax.XScale)
	}
	if _, ok := ax.YScale.(transform.Log); !ok {
		t.Fatalf("LogLog y scale = %T, want transform.Log", ax.YScale)
	}
}

func TestAxes_SetScaleUpdatesSharedRoot(t *testing.T) {
	root := &Axes{
		XScale: transform.NewLinear(-5, 15),
		XAxis:  NewXAxis(),
	}
	shared := &Axes{shareX: root}

	err := shared.SetXScale(
		"symlog",
		transform.WithScaleBase(10),
		transform.WithScaleLinThresh(2),
	)
	if err != nil {
		t.Fatalf("SetXScale(symlog): %v", err)
	}

	if _, ok := root.XScale.(transform.SymLog); !ok {
		t.Fatalf("shared root x scale type = %T, want transform.SymLog", root.XScale)
	}
	locator, ok := root.XAxis.Locator.(SymLogLocator)
	if !ok {
		t.Fatalf("shared root x locator = %T, want SymLogLocator", root.XAxis.Locator)
	}
	if locator.Base != 10 || locator.LinThresh != 2 {
		t.Fatalf("shared root x locator = %+v, want base=10 linthresh=2", locator)
	}
	if _, ok := root.XAxis.MinorLocator.(SymLogLocator); !ok {
		t.Fatalf("shared root x minor locator = %T, want SymLogLocator", root.XAxis.MinorLocator)
	}
	if formatter, ok := root.XAxis.Formatter.(LogFormatterMathText); !ok || !formatter.SciNotation {
		t.Fatalf("shared root x formatter = %#v, want scientific LogFormatterMathText", root.XAxis.Formatter)
	}
	xMin, xMax := root.XScale.Domain()
	if xMin != -5 || xMax != 15 {
		t.Fatalf("shared root x domain = (%v, %v), want (-5, 15)", xMin, xMax)
	}
}

func TestAxes_SetScaleUpdatesOverlayAxisDefaults(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	twinX := ax.TwinX()
	secondaryX, err := ax.SecondaryXAxis(
		AxisTop,
		func(x float64) float64 { return x * 2 },
		func(x float64) (float64, bool) { return x / 2, true },
	)
	if err != nil {
		t.Fatalf("SecondaryXAxis: %v", err)
	}

	if err := ax.SetXScale("log", transform.WithScaleBase(10), transform.WithScaleSubs(2, 5)); err != nil {
		t.Fatalf("SetXScale(log): %v", err)
	}

	if loc, ok := twinX.XAxis.Locator.(LogLocator); !ok || loc.Base != 10 {
		t.Fatalf("twin x locator = %#v, want log base 10", twinX.XAxis.Locator)
	}
	if loc, ok := secondaryX.XAxisTop.Locator.(LogLocator); !ok || loc.Base != 10 {
		t.Fatalf("secondary x top locator = %#v, want log base 10", secondaryX.XAxisTop.Locator)
	}
	if minor, ok := secondaryX.XAxisTop.MinorLocator.(LogLocator); !ok || len(minor.Subs) != 2 {
		t.Fatalf("secondary x top minor locator = %#v, want log minor subs", secondaryX.XAxisTop.MinorLocator)
	}
	if formatter, ok := secondaryX.XAxisTop.Formatter.(LogFormatterMathText); !ok || !formatter.SciNotation {
		t.Fatalf("secondary x top formatter = %#v, want scientific LogFormatterMathText", secondaryX.XAxisTop.Formatter)
	}

	twinY := ax.TwinY()
	secondaryY, err := ax.SecondaryYAxis(
		AxisRight,
		func(y float64) float64 { return y + 1 },
		func(y float64) (float64, bool) { return y - 1, true },
	)
	if err != nil {
		t.Fatalf("SecondaryYAxis: %v", err)
	}

	if err := ax.SetYScale("logit"); err != nil {
		t.Fatalf("SetYScale(logit): %v", err)
	}

	if _, ok := twinY.YAxis.Locator.(LogitLocator); !ok {
		t.Fatalf("twin y locator = %T, want LogitLocator", twinY.YAxis.Locator)
	}
	if _, ok := secondaryY.YAxisRight.Locator.(LogitLocator); !ok {
		t.Fatalf("secondary y right locator = %T, want LogitLocator", secondaryY.YAxisRight.Locator)
	}
	if formatter, ok := secondaryY.YAxisRight.MinorFormatter.(LogitFormatter); !ok || !formatter.Minor {
		t.Fatalf("secondary y right minor formatter = %#v, want minor LogitFormatter", secondaryY.YAxisRight.MinorFormatter)
	}
}

func TestAxes_SetScaleInstallsAsinhLocatorDefaults(t *testing.T) {
	axes := &Axes{
		XScale: transform.NewLinear(-5, 15),
		XAxis:  NewXAxis(),
	}

	err := axes.SetXScale(
		"asinh",
		transform.WithScaleBase(10),
		transform.WithScaleLinearWidth(2),
		transform.WithScaleSubs(2, 5),
	)
	if err != nil {
		t.Fatalf("SetXScale(asinh): %v", err)
	}

	locator, ok := axes.XAxis.Locator.(AsinhLocator)
	if !ok {
		t.Fatalf("x locator = %T, want AsinhLocator", axes.XAxis.Locator)
	}
	if locator.Base != 10 || locator.LinearWidth != 2 {
		t.Fatalf("x locator = %+v, want base=10 linear_width=2", locator)
	}
	minor, ok := axes.XAxis.MinorLocator.(AsinhLocator)
	if !ok {
		t.Fatalf("x minor locator = %T, want AsinhLocator", axes.XAxis.MinorLocator)
	}
	if len(minor.Subs) != 2 || minor.Subs[0] != 2 || minor.Subs[1] != 5 {
		t.Fatalf("x minor locator subs = %v, want [2 5]", minor.Subs)
	}
	if formatter, ok := axes.XAxis.Formatter.(LogFormatterMathText); !ok || !formatter.SciNotation {
		t.Fatalf("x formatter = %#v, want scientific LogFormatterMathText", axes.XAxis.Formatter)
	}
}

func TestAxes_SetScaleInstallsLogitLocatorDefaults(t *testing.T) {
	axes := &Axes{
		XScale: transform.NewLinear(0.01, 0.99),
		XAxis:  NewXAxis(),
	}

	if err := axes.SetXScale("logit"); err != nil {
		t.Fatalf("SetXScale(logit): %v", err)
	}

	if _, ok := axes.XAxis.Locator.(LogitLocator); !ok {
		t.Fatalf("x locator = %T, want LogitLocator", axes.XAxis.Locator)
	}
	if _, ok := axes.XAxis.Formatter.(LogitFormatter); !ok {
		t.Fatalf("x formatter = %T, want LogitFormatter", axes.XAxis.Formatter)
	}
	minor, ok := axes.XAxis.MinorLocator.(LogitLocator)
	if !ok {
		t.Fatalf("x minor locator = %T, want LogitLocator", axes.XAxis.MinorLocator)
	}
	if !minor.Minor {
		t.Fatalf("x minor locator = %+v, want minor=true", minor)
	}
	if _, ok := axes.XAxis.MinorFormatter.(LogitFormatter); !ok {
		t.Fatalf("x minor formatter = %T, want LogitFormatter", axes.XAxis.MinorFormatter)
	}
}

func TestAxes_SetScaleInstallsFunctionLogLocatorDefaults(t *testing.T) {
	axes := &Axes{
		XScale: transform.NewLinear(1, 100),
		XAxis:  NewXAxis(),
	}

	err := axes.SetXScale(
		"functionlog",
		transform.WithScaleBase(10),
		transform.WithScaleFunctions(
			func(x float64) float64 { return x * x },
			func(y float64) (float64, bool) { return math.Sqrt(y), true },
		),
	)
	if err != nil {
		t.Fatalf("SetXScale(functionlog): %v", err)
	}

	locator, ok := axes.XAxis.Locator.(LogLocator)
	if !ok {
		t.Fatalf("x locator = %T, want LogLocator", axes.XAxis.Locator)
	}
	if locator.Base != 10 {
		t.Fatalf("x locator base = %v, want 10", locator.Base)
	}
	if formatter, ok := axes.XAxis.Formatter.(LogFormatterMathText); !ok || !formatter.SciNotation {
		t.Fatalf("x formatter = %#v, want scientific LogFormatterMathText", axes.XAxis.Formatter)
	}
	minor, ok := axes.XAxis.MinorLocator.(LogLocator)
	if !ok {
		t.Fatalf("x minor locator = %T, want LogLocator", axes.XAxis.MinorLocator)
	}
	if minor.Base != 10 || minor.SubsMode != "auto" {
		t.Fatalf("x minor locator = %+v, want base 10 auto subs", minor)
	}
}

func TestAxes_SetLimPreservesScaleType(t *testing.T) {
	axes := &Axes{
		XScale: transform.NewSymLog(-10, 10, 10, 1, 1),
	}

	axes.SetXLim(-20, 30)

	symLog, ok := axes.XScale.(transform.SymLog)
	if !ok {
		t.Fatalf("x scale type after SetXLim = %T, want transform.SymLog", axes.XScale)
	}
	xMin, xMax := symLog.Domain()
	if xMin != -20 || xMax != 30 {
		t.Fatalf("x scale domain after SetXLim = (%v, %v), want (-20, 30)", xMin, xMax)
	}
}

func TestAxes_SetLimExpandsSingularLinearDomains(t *testing.T) {
	ax := &Axes{XScale: transform.NewLinear(0, 1), YScale: transform.NewLinear(0, 1)}

	ax.SetXLim(2, 2)
	xMin, xMax := ax.XScale.Domain()
	if math.Abs(xMin-1.9) > 1e-12 || math.Abs(xMax-2.1) > 1e-12 {
		t.Fatalf("singular x limits = (%v, %v), want (1.9, 2.1)", xMin, xMax)
	}

	ax.SetYLim(0, 0)
	yMin, yMax := ax.YScale.Domain()
	if math.Abs(yMin+0.05) > 1e-12 || math.Abs(yMax-0.05) > 1e-12 {
		t.Fatalf("singular y limits = (%v, %v), want (-0.05, 0.05)", yMin, yMax)
	}
}

func TestAxes_LogAutoscaleNormalizesNonPositiveData(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	if err := ax.SetXScale(
		"log",
		transform.WithScaleBase(10),
		transform.WithScaleNonPositive(transform.NonPositiveClip),
	); err != nil {
		t.Fatalf("SetXScale(log): %v", err)
	}

	ax.Plot([]float64{-10, 1, 100}, []float64{1, 2, 3})
	ax.AutoScale(0)

	logScale, ok := ax.XScale.(transform.Log)
	if !ok {
		t.Fatalf("x scale = %T, want transform.Log", ax.XScale)
	}
	minVal, maxVal := logScale.Domain()
	if minVal <= 0 || maxVal <= minVal {
		t.Fatalf("autoscaled log domain = (%v, %v), want positive increasing domain", minVal, maxVal)
	}
	if got := logScale.Fwd(-10); math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("clipped log forward after autoscale should stay finite, got %v", got)
	}

	if err := ax.SetXScale(
		"log",
		transform.WithScaleBase(10),
		transform.WithScaleNonPositive(transform.NonPositiveMask),
	); err != nil {
		t.Fatalf("SetXScale(log mask): %v", err)
	}
	if got := ax.XScale.Fwd(-10); !math.IsNaN(got) {
		t.Fatalf("masked log forward should be NaN for non-positive input, got %v", got)
	}
}

func TestAxes_TopAxisCreatesExplicitAxis(t *testing.T) {
	axes := &Axes{
		XAxis: NewXAxis(),
	}
	axes.XAxis.Color = render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	axes.XAxis.LineWidth = 2.5

	top := axes.TopAxis()
	if top == nil {
		t.Fatal("TopAxis() returned nil")
	}
	if top == axes.XAxis {
		t.Fatal("TopAxis() should create a distinct axis")
	}
	if top.Side != AxisTop {
		t.Fatalf("TopAxis side = %v, want %v", top.Side, AxisTop)
	}
	if !top.ShowSpine || !top.ShowTicks || !top.ShowLabels {
		t.Fatalf("TopAxis should default to visible components, got %+v", top)
	}
	if top.Color != axes.XAxis.Color || top.LineWidth != axes.XAxis.LineWidth {
		t.Fatalf("TopAxis should inherit x-axis style, got color=%+v width=%v", top.Color, top.LineWidth)
	}
	if axes.TopAxis() != top {
		t.Fatal("TopAxis() should return the existing explicit top axis")
	}
}

func TestAddAxesDefaultXAxisMatchesMatplotlibBottomOnly(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(unitRect())

	if ax.XAxis == nil {
		t.Fatal("default axes should create a bottom x-axis")
	}
	if ax.XAxis.Side != AxisBottom {
		t.Fatalf("default x-axis side = %v, want bottom", ax.XAxis.Side)
	}
	if !ax.XAxis.ShowTicks || !ax.XAxis.ShowLabels {
		t.Fatalf("default bottom x-axis should show ticks and labels: %+v", ax.XAxis)
	}
	if ax.XAxisTop != nil {
		t.Fatalf("default axes should not create a top x-axis, got %+v", ax.XAxisTop)
	}
	if ax.effectiveXLabelSide() != AxisBottom {
		t.Fatalf("default x-label side = %v, want bottom", ax.effectiveXLabelSide())
	}
}

func TestAxes_RightAxisCreatesExplicitAxis(t *testing.T) {
	axes := &Axes{
		YAxis: NewYAxis(),
	}
	axes.YAxis.Color = render.Color{R: 0.4, G: 0.3, B: 0.2, A: 1}
	axes.YAxis.LineWidth = 1.75

	right := axes.RightAxis()
	if right == nil {
		t.Fatal("RightAxis() returned nil")
	}
	if right == axes.YAxis {
		t.Fatal("RightAxis() should create a distinct axis")
	}
	if right.Side != AxisRight {
		t.Fatalf("RightAxis side = %v, want %v", right.Side, AxisRight)
	}
	if !right.ShowSpine || !right.ShowTicks || !right.ShowLabels {
		t.Fatalf("RightAxis should default to visible components, got %+v", right)
	}
	if right.Color != axes.YAxis.Color || right.LineWidth != axes.YAxis.LineWidth {
		t.Fatalf("RightAxis should inherit y-axis style, got color=%+v width=%v", right.Color, right.LineWidth)
	}
	if axes.RightAxis() != right {
		t.Fatal("RightAxis() should return the existing explicit right axis")
	}
}

func TestAxes_TwinXRightAxisUsesTwinYDomain(t *testing.T) {
	fig := NewFigure(760, 360)
	host := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.58, Y: 0.14}, Max: geom.Pt{X: 0.95, Y: 0.78}})
	host.SetXLim(0, 10)
	host.SetYLim(0, 20)

	twin := host.TwinX()
	if twin == nil {
		t.Fatal("TwinX() returned nil")
	}
	twin.SetYLim(0, 100)

	hostYMin, hostYMax := host.effectiveYScale().Domain()
	if hostYMin != 0 || hostYMax != 20 {
		t.Fatalf("host y domain = (%v, %v), want (0, 20)", hostYMin, hostYMax)
	}
	twinYMin, twinYMax := twin.effectiveYScale().Domain()
	if twinYMin != 0 || twinYMax != 100 {
		t.Fatalf("twin y domain = (%v, %v), want (0, 100)", twinYMin, twinYMax)
	}

	ctx := newAxesDrawContext(twin, fig, fig.DisplayRect(), twin.adjustedLayout(fig))
	ctxYMin, ctxYMax := ctx.DataToPixel.YScale.Domain()
	if ctxYMin != 0 || ctxYMax != 100 {
		t.Fatalf("twin draw-context y domain = (%v, %v), want (0, 100)", ctxYMin, ctxYMax)
	}
	right := twin.RightAxis()
	ticks := visibleTicks(right.Locator.Ticks(ctxYMin, ctxYMax, right.majorTickTargetCountForContext(ctx, false)), ctxYMin, ctxYMax)
	if len(ticks) != 6 || ticks[0] != 0 || ticks[len(ticks)-1] != 100 {
		t.Fatalf("twin right-axis ticks = %v, want six ticks spanning 0..100", ticks)
	}
}

func TestAxes_SetAxisSides(t *testing.T) {
	axes := &Axes{
		XAxis:      NewXAxis(),
		YAxis:      NewYAxis(),
		XAxisTop:   NewXAxis(),
		YAxisRight: NewYAxis(),
	}

	if err := axes.MoveXAxisToTop(); err != nil {
		t.Fatalf("MoveXAxisToTop: %v", err)
	}
	if axes.XAxis.Side != AxisTop {
		t.Fatalf("primary x-axis side = %v, want top", axes.XAxis.Side)
	}
	if axes.XAxisTop != nil {
		t.Fatal("moving primary x-axis to top should drop explicit top axis")
	}

	if err := axes.MoveYAxisToRight(); err != nil {
		t.Fatalf("MoveYAxisToRight: %v", err)
	}
	if axes.YAxis.Side != AxisRight {
		t.Fatalf("primary y-axis side = %v, want right", axes.YAxis.Side)
	}
	if axes.YAxisRight != nil {
		t.Fatal("moving primary y-axis to right should drop explicit right axis")
	}
}

func TestAxes_InvertXToggle(t *testing.T) {
	axes := &Axes{
		XScale: transform.NewLinear(0, 10),
	}

	if axes.XInverted() {
		t.Fatal("new linear x-axis should not be inverted")
	}

	axes.InvertX()
	if !axes.XInverted() {
		t.Fatal("InvertX() should mark the axis as inverted")
	}
	xMin, xMax := axes.XScale.Domain()
	if xMin != 10 || xMax != 0 {
		t.Fatalf("inverted x limits = (%v, %v), want (10, 0)", xMin, xMax)
	}
	if got := axes.XScale.Fwd(0); got != 1 {
		t.Fatalf("inverted x scale forward(0) = %v, want 1", got)
	}

	axes.InvertX()
	if axes.XInverted() {
		t.Fatal("second InvertX() should restore normal direction")
	}
	xMin, xMax = axes.XScale.Domain()
	if xMin != 0 || xMax != 10 {
		t.Fatalf("restored x limits = (%v, %v), want (0, 10)", xMin, xMax)
	}
}

func TestAxes_InvertYToggle(t *testing.T) {
	axes := &Axes{
		YScale: transform.NewLinear(-2, 8),
	}

	if axes.YInverted() {
		t.Fatal("new linear y-axis should not be inverted")
	}

	axes.InvertY()
	if !axes.YInverted() {
		t.Fatal("InvertY() should mark the axis as inverted")
	}
	yMin, yMax := axes.YScale.Domain()
	if yMin != 8 || yMax != -2 {
		t.Fatalf("inverted y limits = (%v, %v), want (8, -2)", yMin, yMax)
	}
	if got := axes.YScale.Fwd(-2); got != 1 {
		t.Fatalf("inverted y scale forward(-2) = %v, want 1", got)
	}

	axes.InvertY()
	if axes.YInverted() {
		t.Fatal("second InvertY() should restore normal direction")
	}
	yMin, yMax = axes.YScale.Domain()
	if yMin != -2 || yMax != 8 {
		t.Fatalf("restored y limits = (%v, %v), want (-2, 8)", yMin, yMax)
	}
}

func TestAxes_SetAspectAndBoxAspectAffectLayout(t *testing.T) {
	fig := NewFigure(400, 200)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	full := ax.adjustedLayout(fig)
	if full.W() != 400 || full.H() != 200 {
		t.Fatalf("default adjusted layout = %+v, want full figure rect", full)
	}

	if err := ax.SetAspect("equal"); err != nil {
		t.Fatalf("SetAspect(equal): %v", err)
	}
	equalRect := ax.adjustedLayout(fig)
	if equalRect.W() != equalRect.H() {
		t.Fatalf("equal aspect rect = %+v, want square", equalRect)
	}

	ax.SetAspect("auto")
	if err := ax.SetBoxAspect(2); err != nil {
		t.Fatalf("SetBoxAspect: %v", err)
	}
	boxRect := ax.adjustedLayout(fig)
	if got := boxRect.H() / boxRect.W(); got != 2 {
		t.Fatalf("box aspect ratio = %v, want 2", got)
	}
}

func TestAxes_EqualAspectLayoutMatchesMatplotlibFractionAnchoring(t *testing.T) {
	fig := NewFigure(1100, 720)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.66, Y: 0.34},
		Max: geom.Pt{X: 0.75, Y: 0.56},
	})
	ax.SetXLim(-0.5, 27.5)
	ax.SetYLim(27.5, -0.5)
	if err := ax.SetAspect("equal"); err != nil {
		t.Fatalf("SetAspect(equal): %v", err)
	}

	got := ax.adjustedLayout(fig)
	if got.Min.Y != 274.50000000000006 || got.Max.Y != 373.50000000000006 {
		t.Fatalf("equal-aspect y bounds = %.17g..%.17g, want Matplotlib fraction-anchored 274.50000000000006..373.50000000000006", got.Min.Y, got.Max.Y)
	}
}

func TestAxes_AspectAnchorPositionsShrunkBox(t *testing.T) {
	mk := func(anchor string) geom.Rect {
		fig := NewFigure(1100, 720)
		ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.66, Y: 0.34}, Max: geom.Pt{X: 0.75, Y: 0.56}})
		ax.SetXLim(-0.5, 27.5)
		ax.SetYLim(27.5, -0.5)
		_ = ax.SetAspect("equal")
		if err := ax.SetAnchor(anchor); err != nil {
			t.Fatalf("SetAnchor(%q): %v", anchor, err)
		}
		return ax.adjustedLayout(fig)
	}
	sw := mk("SW")
	c := mk("C")
	ne := mk("NE")
	// The box is shrunk vertically; SW pins it to the bottom, NE to the top.
	if !(sw.Min.Y < c.Min.Y && c.Min.Y < ne.Min.Y) {
		t.Fatalf("anchor ordering wrong: SW=%.3f C=%.3f NE=%.3f", sw.Min.Y, c.Min.Y, ne.Min.Y)
	}
	// Heights stay equal regardless of anchor.
	if !floatApprox(sw.H(), ne.H(), 1e-6) {
		t.Fatalf("anchor changed height: SW=%.3f NE=%.3f", sw.H(), ne.H())
	}
}

func TestAxes_AspectDatalimExpandsDataLimits(t *testing.T) {
	fig := NewFigure(600, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 10}, []float64{0, 1})
	_ = ax.SetAspect("equal")
	if err := ax.SetAdjustable("datalim"); err != nil {
		t.Fatalf("SetAdjustable(datalim): %v", err)
	}
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	xspan := math.Abs(xMax - xMin)
	yspan := math.Abs(yMax - yMin)
	// Square pixel box + equal aspect => equal data spans (y expanded to match x).
	if !floatApprox(xspan, yspan, 1e-6) {
		t.Fatalf("datalim did not equalize spans: xspan=%.6f yspan=%.6f", xspan, yspan)
	}
	if yMin > 0 || yMax < 1 {
		t.Fatalf("datalim expansion must keep data visible: y=[%.3f,%.3f]", yMin, yMax)
	}
}

func TestAxes_AspectDatalimReappliedWhenAspectSetAfterAdjustable(t *testing.T) {
	fig := NewFigure(600, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 10}, []float64{0, 1})
	// Reverse order: adjustable selected before the aspect. SetAspect must
	// re-apply the datalim expansion; otherwise the data scale stays unequal.
	if err := ax.SetAdjustable("datalim"); err != nil {
		t.Fatalf("SetAdjustable(datalim): %v", err)
	}
	if err := ax.SetAspect("equal"); err != nil {
		t.Fatalf("SetAspect(equal): %v", err)
	}
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	xspan := math.Abs(xMax - xMin)
	yspan := math.Abs(yMax - yMin)
	if !floatApprox(xspan, yspan, 1e-6) {
		t.Fatalf("datalim not equalized when aspect set after adjustable: xspan=%.6f yspan=%.6f", xspan, yspan)
	}
}
