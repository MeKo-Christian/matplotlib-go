package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type staticLocator []float64

func (s staticLocator) Ticks(min, max float64, count int) []float64 {
	return append([]float64(nil), s...)
}

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

func TestAxes_TickParamsLocatorParamsAndMinorTicks(t *testing.T) {
	axes := &Axes{
		XAxis:      NewXAxis(),
		YAxis:      NewYAxis(),
		XAxisTop:   NewXAxis(),
		YAxisRight: NewYAxis(),
	}
	axes.XAxisTop.Side = AxisTop
	axes.YAxisRight.Side = AxisRight

	if err := axes.MinorticksOn("x"); err != nil {
		t.Fatalf("MinorticksOn(x): %v", err)
	}
	if axes.XAxis.MinorLocator == nil || axes.XAxisTop.MinorLocator == nil {
		t.Fatal("MinorticksOn(x) should enable minor locators on both x axes")
	}

	if err := axes.LocatorParams(LocatorParams{Axis: "x", MajorCount: 8, MinorCount: 40}); err != nil {
		t.Fatalf("LocatorParams: %v", err)
	}
	if axes.XAxis.MajorTickCount != 8 || axes.XAxisTop.MinorTickCount != 40 {
		t.Fatalf("locator params not applied: bottom=%+v top=%+v", axes.XAxis, axes.XAxisTop)
	}

	length := 11.0
	width := 2.25
	minorWidth := 1.25
	showLabels := false
	color := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	if err := axes.TickParams(TickParams{
		Axis:       "right",
		Which:      "major",
		Color:      &color,
		Length:     &length,
		Width:      &width,
		ShowLabels: &showLabels,
	}); err != nil {
		t.Fatalf("TickParams: %v", err)
	}
	if err := axes.TickParams(TickParams{
		Axis:  "right",
		Which: "minor",
		Width: &minorWidth,
	}); err != nil {
		t.Fatalf("TickParams(minor width): %v", err)
	}
	if axes.YAxisRight.TickSize != length || axes.YAxisRight.ShowLabels {
		t.Fatalf("tick params not applied to right axis: %+v", axes.YAxisRight)
	}
	if axes.YAxisRight.Color == color {
		t.Fatalf("TickParams color should not recolor spine: %+v", axes.YAxisRight.Color)
	}

	axes.YAxisRight.Locator = staticLocator{5}
	axes.YAxisRight.MinorLocator = staticLocator{5.5}
	ctx := createTestDrawContext()
	r := &recordingRenderer{}
	axes.YAxisRight.Draw(r, ctx)
	axes.YAxisRight.DrawTicks(r, ctx)
	if len(r.pathCalls) != 3 {
		t.Fatalf("expected spine, minor tick, and major tick path calls, got %d", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.LineWidth; got != defaultAxisLineWidth {
		t.Fatalf("spine line width = %v, want default %v", got, defaultAxisLineWidth)
	}
	if got := r.pathCalls[1].paint.LineWidth; got != minorWidth {
		t.Fatalf("minor tick line width = %v, want %v", got, minorWidth)
	}
	if got := r.pathCalls[2].paint.LineWidth; got != width {
		t.Fatalf("major tick line width = %v, want %v", got, width)
	}

	if err := axes.MinorticksOff("x"); err != nil {
		t.Fatalf("MinorticksOff(x): %v", err)
	}
	if axes.XAxis.MinorLocator != nil || axes.XAxisTop.MinorLocator != nil {
		t.Fatal("MinorticksOff(x) should clear minor locators on both x axes")
	}
}

func TestAxisMajorTickTargetUsesMatplotlibTickSpaceHeuristic(t *testing.T) {
	ctx := &DrawContext{
		RC: style.RC{
			DPI:                100,
			XTickLabelFontSize: 10,
			YTickLabelFontSize: 10,
		},
		Clip: geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 320, Y: 460},
		},
	}

	axis := NewXAxis()
	if got, want := axis.majorTickTargetCountForContext(ctx, true), 7; got != want {
		t.Fatalf("x tick target = %v, want %v", got, want)
	}
	if got, want := axis.majorTickTargetCountForContext(ctx, false), 9; got != want {
		t.Fatalf("y tick target = %v, want %v", got, want)
	}

	shortY := *ctx
	shortY.Clip.Max.Y = 180
	if got, want := axis.majorTickTargetCountForContext(&shortY, false), 6; got != want {
		t.Fatalf("short y tick target = %v, want %v", got, want)
	}
}

func TestSpinePixelEndpointsRightBoundaryUsesMatplotlibPathSnapper(t *testing.T) {
	px := geom.Rect{
		Min: geom.Pt{X: 51.06977777777778, Y: 47.444777777777794},
		Max: geom.Pt{X: 846.4725805555556, Y: 673.4996666666666},
	}

	p1, p2 := spinePixelEndpoints(AxisRight, px)
	if p1.X != 846.5 || p2.X != 846.5 {
		t.Fatalf("right spine x = %v..%v, want Matplotlib pixel center 846.5", p1.X, p2.X)
	}
}

func TestSpinePixelEndpointsRightBoundaryRoundsPastHalfPixel(t *testing.T) {
	px := geom.Rect{
		Min: geom.Pt{X: 76.8, Y: 57.6},
		Max: geom.Pt{X: 588.8, Y: 316.8},
	}

	p1, p2 := spinePixelEndpoints(AxisRight, px)
	if p1.X != 589.5 || p2.X != 589.5 {
		t.Fatalf("right spine x = %v..%v, want Matplotlib pixel center 589.5", p1.X, p2.X)
	}
}

func TestAxesLocatorParamsMajorCountBypassesAdaptiveTickCapacity(t *testing.T) {
	axes := &Axes{
		XAxis:    NewXAxis(),
		XAxisTop: NewXAxis(),
	}
	axes.XAxisTop.Side = AxisTop

	if err := axes.LocatorParams(LocatorParams{Axis: "x", MajorCount: 6}); err != nil {
		t.Fatalf("LocatorParams: %v", err)
	}

	ctx := &DrawContext{
		RC: style.RC{
			DPI:                100,
			XTickLabelFontSize: 10,
		},
		Clip: geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 240, Y: 320},
		},
	}
	if got, want := axes.XAxisTop.majorTickTargetCountForContext(ctx, true), 6; got != want {
		t.Fatalf("explicit x tick target = %v, want %v", got, want)
	}

	ticks := axes.XAxisTop.Locator.Ticks(-1, 5, axes.XAxisTop.majorTickTargetCountForContext(ctx, true))
	want := []float64{-1, 0, 1, 2, 3, 4, 5}
	if len(ticks) != len(want) {
		t.Fatalf("tick count = %d, want %d: %v", len(ticks), len(want), ticks)
	}
	for i := range want {
		if ticks[i] != want[i] {
			t.Fatalf("ticks = %v, want %v", ticks, want)
		}
	}
}

func TestAxes_TickParamsAppliesLabelStyle(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis()}

	rotation := 45.0
	labelSize := 13.0
	pad := 9.0
	hAlign := TextAlignRight
	vAlign := TextVAlignTop
	showMinorLabels := true

	if err := axes.TickParams(TickParams{
		Axis:          "bottom",
		Which:         "minor",
		ShowLabels:    &showMinorLabels,
		LabelRotation: &rotation,
		LabelSize:     &labelSize,
		LabelPad:      &pad,
		LabelHAlign:   &hAlign,
		LabelVAlign:   &vAlign,
	}); err != nil {
		t.Fatalf("TickParams(minor label style): %v", err)
	}

	if !axes.XAxis.ShowMinorLabels {
		t.Fatal("TickParams should enable minor labels for minor selection")
	}
	if axes.XAxis.MinorLabelStyle.Rotation != rotation || axes.XAxis.MinorLabelStyle.Pad != pad || axes.XAxis.MinorLabelStyle.FontSize != labelSize {
		t.Fatalf("minor label style mismatch: %+v", axes.XAxis.MinorLabelStyle)
	}
	if axes.XAxis.MinorLabelStyle.HAlign != hAlign || axes.XAxis.MinorLabelStyle.VAlign != vAlign || axes.XAxis.MinorLabelStyle.AutoAlign {
		t.Fatalf("minor label alignment mismatch: %+v", axes.XAxis.MinorLabelStyle)
	}
}

func TestAxes_TickParamsAppliesDirection(t *testing.T) {
	axes := &Axes{YAxis: NewYAxis()}
	direction := "inout"

	if err := axes.TickParams(TickParams{
		Axis:      "left",
		Which:     "major",
		Direction: &direction,
	}); err != nil {
		t.Fatalf("TickParams(direction): %v", err)
	}

	if axes.YAxis.TickDirection != TickDirectionInOut {
		t.Fatalf("tick direction = %v, want %v", axes.YAxis.TickDirection, TickDirectionInOut)
	}
}

func TestAxes_TickParamsResetRestoresAxisOwnedDefaults(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis()}
	color := render.Color{R: 0.7, G: 0.2, B: 0.1, A: 1}
	length := 12.0
	width := 2.0
	rotation := 30.0
	labelSize := 14.0
	showLabels := true
	direction := "inout"

	if err := axes.TickParams(TickParams{
		Axis:          "bottom",
		Which:         "both",
		Color:         &color,
		Length:        &length,
		Width:         &width,
		Direction:     &direction,
		ShowLabels:    &showLabels,
		LabelRotation: &rotation,
		LabelSize:     &labelSize,
	}); err != nil {
		t.Fatalf("TickParams(style): %v", err)
	}
	if axes.XAxis.TickColor == nil || axes.XAxis.TickSize != length || axes.XAxis.MinorTickLineWidth != width {
		t.Fatalf("pre-reset tick params were not applied: %+v", axes.XAxis)
	}

	newLength := 6.0
	if err := axes.TickParams(TickParams{
		Axis:   "bottom",
		Which:  "major",
		Reset:  true,
		Length: &newLength,
	}); err != nil {
		t.Fatalf("TickParams(reset): %v", err)
	}
	if axes.XAxis.TickColor != nil || axes.XAxis.TickLabelColor != nil {
		t.Fatalf("reset tick colors = tick %+v label %+v, want nil overrides", axes.XAxis.TickColor, axes.XAxis.TickLabelColor)
	}
	if axes.XAxis.TickSize != newLength || axes.XAxis.MinorTickSize != 0 {
		t.Fatalf("reset tick sizes = major %v minor %v, want major override %v and default minor", axes.XAxis.TickSize, axes.XAxis.MinorTickSize, newLength)
	}
	if axes.XAxis.TickLineWidth != 0 || axes.XAxis.MinorTickLineWidth != 0 {
		t.Fatalf("reset tick widths = major %v minor %v, want defaults", axes.XAxis.TickLineWidth, axes.XAxis.MinorTickLineWidth)
	}
	if axes.XAxis.TickDirection != TickDirectionOut || !axes.XAxis.ShowLabels || axes.XAxis.ShowMinorLabels {
		t.Fatalf("reset visibility/direction mismatch: %+v", axes.XAxis)
	}
	if !axes.XAxis.MajorLabelStyle.AutoAlign || axes.XAxis.MajorLabelStyle.Rotation != 0 {
		t.Fatalf("reset major label style = %+v, want default auto alignment", axes.XAxis.MajorLabelStyle)
	}
	if axes.XAxis.MajorLabelStyle.FontSize != 0 {
		t.Fatalf("reset major label font size = %v, want default 0", axes.XAxis.MajorLabelStyle.FontSize)
	}
}

func TestAxes_TickParamsColorsTicksAndLabelsNotSpine(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{5}
	axis.Formatter = ScalarFormatter{Prec: 0}
	axes := &Axes{XAxis: axis}

	tickColor := render.Color{R: 0.18, G: 0.42, B: 0.55, A: 1}
	labelSize := 15.0
	if err := axes.TickParams(TickParams{
		Axis:      "bottom",
		Which:     "major",
		Color:     &tickColor,
		LabelSize: &labelSize,
	}); err != nil {
		t.Fatalf("TickParams(color): %v", err)
	}

	ctx := createTestDrawContext()
	r := &textRecordingRenderer{}
	axis.Draw(r, ctx)
	axis.DrawTicks(r, ctx)
	axis.DrawTickLabels(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected spine and tick path calls, got %d", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Stroke; got != (render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("spine color = %+v, want default black", got)
	}
	if got := r.pathCalls[1].paint.Stroke; got != tickColor {
		t.Fatalf("tick color = %+v, want %+v", got, tickColor)
	}
	if len(r.textColors) != 1 {
		t.Fatalf("tick label draw count = %d, want 1", len(r.textColors))
	}
	if got := r.textColors[0]; got != tickColor {
		t.Fatalf("tick label color = %+v, want %+v", got, tickColor)
	}
	if got := r.textSizes[0]; got != labelSize {
		t.Fatalf("tick label size = %v, want %v", got, labelSize)
	}
}

func TestAxes_TickParamsSideVisibility(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), YAxis: NewYAxis()}
	show := true
	hide := false

	if err := axes.TickParams(TickParams{
		Axis:        "x",
		Top:         &show,
		Bottom:      &hide,
		LabelTop:    &show,
		LabelBottom: &hide,
	}); err != nil {
		t.Fatalf("TickParams(x sides): %v", err)
	}
	if axes.XAxis == nil || axes.XAxis.ShowTicks || axes.XAxis.ShowLabels {
		t.Fatalf("bottom x-axis visibility = %+v, want hidden ticks/labels", axes.XAxis)
	}
	if axes.XAxisTop == nil || !axes.XAxisTop.ShowTicks || !axes.XAxisTop.ShowLabels {
		t.Fatalf("top x-axis visibility = %+v, want visible ticks/labels", axes.XAxisTop)
	}
	if axes.YAxis == nil || !axes.YAxis.ShowTicks || !axes.YAxis.ShowLabels {
		t.Fatalf("y-axis should be unchanged, got %+v", axes.YAxis)
	}

	if err := axes.TickParams(TickParams{
		Axis:       "y",
		Right:      &show,
		Left:       &hide,
		LabelRight: &show,
		LabelLeft:  &hide,
	}); err != nil {
		t.Fatalf("TickParams(y sides): %v", err)
	}
	if axes.YAxis == nil || axes.YAxis.ShowTicks || axes.YAxis.ShowLabels {
		t.Fatalf("left y-axis visibility = %+v, want hidden ticks/labels", axes.YAxis)
	}
	if axes.YAxisRight == nil || !axes.YAxisRight.ShowTicks || !axes.YAxisRight.ShowLabels {
		t.Fatalf("right y-axis visibility = %+v, want visible ticks/labels", axes.YAxisRight)
	}
}

func TestAxes_TickParamsGridStyling(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), YAxis: NewYAxis()}
	xGrid := axes.AddXGrid()
	yGrid := axes.AddYGrid()
	xGrid.Minor = true
	yGrid.Minor = true

	visible := false
	gridColor := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.35
	width := 2.5
	dashes := []float64{4, 2}
	if err := axes.TickParams(TickParams{
		Axis:          "x",
		Which:         "both",
		GridVisible:   &visible,
		GridColor:     &gridColor,
		GridAlpha:     &alpha,
		GridLineWidth: &width,
		GridDashes:    dashes,
	}); err != nil {
		t.Fatalf("TickParams(grid): %v", err)
	}

	if xGrid.Major || xGrid.Minor {
		t.Fatalf("x grid visibility = major %v minor %v, want both false", xGrid.Major, xGrid.Minor)
	}
	if xGrid.Color != gridColor || xGrid.MinorColor.R != gridColor.R || xGrid.MinorColor.A != alpha {
		t.Fatalf("x grid colors = major %+v minor %+v", xGrid.Color, xGrid.MinorColor)
	}
	if xGrid.Alpha != alpha || xGrid.LineWidth != width || xGrid.MinorLineWidth != width {
		t.Fatalf("x grid style = alpha %v width %v minor width %v", xGrid.Alpha, xGrid.LineWidth, xGrid.MinorLineWidth)
	}
	if len(xGrid.Dashes) != 2 || xGrid.Dashes[0] != 4 || len(xGrid.MinorDashes) != 2 || xGrid.MinorDashes[1] != 2 {
		t.Fatalf("x grid dashes = major %v minor %v", xGrid.Dashes, xGrid.MinorDashes)
	}
	if !yGrid.Major || !yGrid.Minor {
		t.Fatalf("y grid should be unchanged, got major %v minor %v", yGrid.Major, yGrid.Minor)
	}
}

func TestAxes_SetAxisLineStyleAppliesToSelectedAxes(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), XAxisTop: NewXAxis(), YAxis: NewYAxis()}

	if err := axes.SetAxisLineStyle("x", render.CapRound, render.JoinBevel, 3, 2); err != nil {
		t.Fatalf("SetAxisLineStyle(x): %v", err)
	}

	if axes.XAxis.LineCap != render.CapRound || axes.XAxis.LineJoin != render.JoinBevel {
		t.Fatalf("bottom axis style = cap %v join %v", axes.XAxis.LineCap, axes.XAxis.LineJoin)
	}
	if axes.XAxisTop.LineCap != render.CapRound || axes.XAxisTop.LineJoin != render.JoinBevel {
		t.Fatalf("top axis style = cap %v join %v", axes.XAxisTop.LineCap, axes.XAxisTop.LineJoin)
	}
	if len(axes.XAxis.Dashes) != 2 || axes.XAxis.Dashes[0] != 3 || axes.XAxis.Dashes[1] != 2 {
		t.Fatalf("bottom axis dashes = %v", axes.XAxis.Dashes)
	}
	if axes.YAxis.LineCap != render.CapSquare {
		t.Fatalf("y axis should be unchanged, got cap %v", axes.YAxis.LineCap)
	}
}

func TestAxisSetLineStyleAffectsSpineAndTickPaint(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = nil
	axis.SetLineStyle(render.CapRound, render.JoinBevel, 4, 1)

	ctx := createTestDrawContext()
	r := &recordingRenderer{}

	axis.Draw(r, ctx)
	axis.DrawTicks(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected spine and tick path calls, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapRound || r.pathCalls[0].paint.LineJoin != render.JoinBevel {
		t.Fatalf("spine paint = %+v", r.pathCalls[0].paint)
	}
	if len(r.pathCalls[0].paint.Dashes) != 2 {
		t.Fatalf("spine dashes = %v", r.pathCalls[0].paint.Dashes)
	}
	if r.pathCalls[1].paint.LineCap != render.CapRound || r.pathCalls[1].paint.LineJoin != render.JoinBevel {
		t.Fatalf("tick paint = %+v", r.pathCalls[1].paint)
	}
}

func TestPolarAxisUsesConfiguredLineStyle(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ax.XAxis.SetLineStyle(render.CapRound, render.JoinBevel, 5, 2)

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &recordingRenderer{}

	ax.XAxis.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one polar spine path, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapRound || r.pathCalls[0].paint.LineJoin != render.JoinBevel {
		t.Fatalf("polar spine paint = %+v", r.pathCalls[0].paint)
	}
	if len(r.pathCalls[0].paint.Dashes) != 2 || r.pathCalls[0].paint.Dashes[0] != 5 || r.pathCalls[0].paint.Dashes[1] != 2 {
		t.Fatalf("polar spine dashes = %v", r.pathCalls[0].paint.Dashes)
	}
}

func TestAxes_TickLabelPositionHelpers(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), YAxis: NewYAxis()}

	if err := axes.SetXTickLabelPosition("top"); err != nil {
		t.Fatalf("SetXTickLabelPosition(top): %v", err)
	}
	if axes.XAxis.ShowLabels {
		t.Fatal("bottom x-axis labels should be hidden when top labels are requested")
	}
	if axes.XAxisTop == nil || !axes.XAxisTop.ShowLabels {
		t.Fatal("top x-axis labels should be visible after SetXTickLabelPosition(top)")
	}

	if err := axes.SetYTickLabelPosition("both"); err != nil {
		t.Fatalf("SetYTickLabelPosition(both): %v", err)
	}
	if !axes.YAxis.ShowLabels || axes.YAxisRight == nil || !axes.YAxisRight.ShowLabels {
		t.Fatal("both y-axis labels should be visible after SetYTickLabelPosition(both)")
	}
}

func TestAxes_LabelPositionHelpers(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	if err := ax.SetXLabelPosition("top"); err != nil {
		t.Fatalf("SetXLabelPosition(top): %v", err)
	}
	if err := ax.SetYLabelPosition("right"); err != nil {
		t.Fatalf("SetYLabelPosition(right): %v", err)
	}

	if ax.effectiveXLabelSide() != AxisTop {
		t.Fatalf("effective x label side = %v, want %v", ax.effectiveXLabelSide(), AxisTop)
	}
	if ax.effectiveYLabelSide() != AxisRight {
		t.Fatalf("effective y label side = %v, want %v", ax.effectiveYLabelSide(), AxisRight)
	}
}

func TestAxes_TwinAxes(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	twinX := ax.TwinX()
	if twinX == nil {
		t.Fatal("TwinX() returned nil")
	}
	if twinX.shareX != ax.xScaleRoot() {
		t.Fatal("TwinX() should share the x-scale root")
	}
	if twinX.YAxisRight == nil {
		t.Fatal("TwinX() should expose a right-side y-axis")
	}
	if twinX.XAxis.ShowTicks || twinX.XAxis.ShowLabels {
		t.Fatal("TwinX() should hide the duplicate primary x-axis")
	}

	twinY := ax.TwinY()
	if twinY == nil {
		t.Fatal("TwinY() returned nil")
	}
	if twinY.shareY != ax.yScaleRoot() {
		t.Fatal("TwinY() should share the y-scale root")
	}
	if twinY.XAxisTop == nil {
		t.Fatal("TwinY() should expose a top-side x-axis")
	}
	if twinY.YAxis.ShowTicks || twinY.YAxis.ShowLabels {
		t.Fatal("TwinY() should hide the duplicate primary y-axis")
	}
}

func TestAxes_SecondaryAxesUseLinkedScale(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 10)

	secondaryX, err := ax.SecondaryXAxis(
		AxisTop,
		func(v float64) float64 { return v*1.8 + 32 },
		func(v float64) (float64, bool) { return (v - 32) / 1.8, true },
	)
	if err != nil {
		t.Fatalf("SecondaryXAxis: %v", err)
	}
	dMin, dMax := secondaryX.XScale.Domain()
	if dMin != 32 || dMax != 212 {
		t.Fatalf("secondary x domain = (%v, %v), want (32, 212)", dMin, dMax)
	}
	if got := secondaryX.XScale.Fwd(122); got != ax.XScale.Fwd(50) {
		t.Fatalf("secondary x forward mapping mismatch: got %v want %v", got, ax.XScale.Fwd(50))
	}
	if secondaryX.XAxisTop == nil {
		t.Fatal("SecondaryXAxis should expose a top x-axis")
	}
	if secondaryX.XAxisTop.ShowSpine {
		t.Fatal("SecondaryXAxis should not draw an overlay spine over the parent frame")
	}

	secondaryY, err := ax.SecondaryYAxis(
		AxisRight,
		func(v float64) float64 { return v * 1000 },
		func(v float64) (float64, bool) { return v / 1000, true },
	)
	if err != nil {
		t.Fatalf("SecondaryYAxis: %v", err)
	}
	dMin, dMax = secondaryY.YScale.Domain()
	if dMin != 0 || dMax != 10000 {
		t.Fatalf("secondary y domain = (%v, %v), want (0, 10000)", dMin, dMax)
	}
	if secondaryY.YAxisRight == nil {
		t.Fatal("SecondaryYAxis should expose a right y-axis")
	}
	if secondaryY.YAxisRight.ShowSpine {
		t.Fatal("SecondaryYAxis should not draw an overlay spine over the parent frame")
	}
}

func TestGrid_Draw(t *testing.T) {
	// Test grid drawing
	grid := NewGrid(AxisBottom)

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

type gridRecordingRenderer struct {
	render.NullRenderer
	paths []geom.Path
}

func (r *gridRecordingRenderer) Path(p geom.Path, _ *render.Paint) {
	r.paths = append(r.paths, p)
}

func TestGrid_UsesOwningAxisMajorLocator(t *testing.T) {
	grid := NewGrid(AxisLeft)
	ctx := createTestDrawContext()
	ctx.DataToPixel.YScale = transform.NewLinear(0, 80)
	ctx.Axes = &Axes{
		XAxis: NewXAxis(),
		YAxis: NewYAxis(),
	}
	ctx.Axes.YAxis.Locator = FixedLocator{TicksList: []float64{0, 20, 40, 60, 80}}

	renderer := &gridRecordingRenderer{}
	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()

	if got, want := len(renderer.paths), 5; got != want {
		t.Fatalf("grid path count = %d, want %d", got, want)
	}
}

func TestGrid_Disabled(t *testing.T) {
	// Test grid with major disabled
	grid := NewGrid(AxisLeft)
	grid.Major = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not draw anything
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_MinorDraw(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Minor = true
	grid.MinorDashes = []float64{2, 3}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	// Should not panic with minor grid enabled
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_MinorOnlyDraw(t *testing.T) {
	grid := NewGrid(AxisLeft)
	grid.Major = false
	grid.Minor = true

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_CustomLocator(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Locator = LogLocator{Base: 10}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestAxis_DrawTickLabels_UsesStepPrecisionForScalarFormatter(t *testing.T) {
	axis := NewYAxis()
	axis.Locator = LinearLocator{}
	axis.MajorTickCount = 6
	ctx := createTestDrawContext()
	ctx.DataToPixel.YScale = transform.NewLinear(0, 0.196)

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"0.00", "0.05", "0.10", "0.15"}
	if len(r.texts) != len(want) {
		t.Fatalf("unexpected tick label count: got %v want %v", r.texts, want)
	}
	for i := range want {
		if r.texts[i] != want[i] {
			t.Fatalf("tick label %d mismatch: got %q want %q", i, r.texts[i], want[i])
		}
	}
}

func TestAxis_DrawTickLabels_OmitsXLabelsOutsideViewLimits(t *testing.T) {
	axis := NewXAxis()
	ctx := createTestDrawContext()
	ctx.DataToPixel.XScale = transform.NewLinear(0.8, 10.8)

	var r axisLabelRecordingRenderer
	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	want := []string{"2", "4", "6", "8", "10"}
	if len(r.texts) != len(want) {
		t.Fatalf("unexpected tick label count: got %v want %v", r.texts, want)
	}
	for i := range want {
		if r.texts[i] != want[i] {
			t.Fatalf("tick label %d mismatch: got %q want %q", i, r.texts[i], want[i])
		}
	}
}

type axisLabelRecordingRenderer struct {
	render.NullRenderer
	texts          []string
	origins        []geom.Pt
	rotatedText    []string
	rotatedAnchors []geom.Pt
	textPathCalls  []string
	bounds         map[string]render.TextBounds
	useBounds      bool
	fontHeights    render.FontHeightMetrics
	useFontHeights bool
	pathCount      int
}

func (r *axisLabelRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *axisLabelRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	if !r.useBounds || r.bounds == nil {
		return render.TextBounds{}, false
	}
	b, ok := r.bounds[text]
	return b, ok
}

func (r *axisLabelRecordingRenderer) MeasureFontHeights(_ float64, _ string) (render.FontHeightMetrics, bool) {
	if !r.useFontHeights {
		return render.FontHeightMetrics{}, false
	}
	return r.fontHeights, true
}

func (r *axisLabelRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *axisLabelRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

func (r *axisLabelRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func (r *axisLabelRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
}

func TestTickLabelPositionUsesBoundsForBottomXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -8, W: 5, H: 10},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the bottom label sits below the tick (anchor
	// tickPos.Y-labelPad) with a Top vAlign baseline offset of -Ascent.
	want := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y - labelPad - 8,
	}
	if r.origins[0] != want {
		t.Fatalf("bottom x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPadMatchesMatplotlibOutsideTickPadding(t *testing.T) {
	axis := NewXAxis()
	axis.TickSize = 10

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if got, want := tickLabelPadPx(axis, ctx), 13.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("outward tick label pad = %v, want %v", got, want)
	}

	axis.TickDirection = TickDirectionInOut
	if got, want := tickLabelPadPx(axis, ctx), 8.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("inout tick label pad = %v, want %v", got, want)
	}

	axis.TickDirection = TickDirectionIn
	if got, want := tickLabelPadPx(axis, ctx), 3.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("inward tick label pad = %v, want %v", got, want)
	}
}

func TestTickLabelPositionUsesBoundsForLeftYAxis(t *testing.T) {
	axis := NewYAxis()
	axis.Locator = staticLocator{4}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"4": {X: 1, Y: -8, W: 5, H: 10},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(axis, ctx), Y: 4})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the y-axis label is center-baseline aligned, so the
	// baseline offset is -Ascent/2 = -4.
	want := geom.Pt{
		X: tickPos.X - labelPad - 5.0,
		Y: tickPos.Y - 4,
	}
	if r.origins[0] != want {
		t.Fatalf("left y tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesFontHeightMetricsForBottomXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	layout := measureSingleLineTextLayout(&r, "2", tickLabelFontSize(axis, ctx), ctx.RC.FontKey)
	// Display space is y-up: bottom label below the tick with Top vAlign (-Ascent).
	want := geom.Pt{
		X: tickPos.X - layout.Width/2,
		Y: tickPos.Y - labelPad - 8,
	}
	if r.origins[0] != want {
		t.Fatalf("bottom x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesBottomAlignmentForTopXAxis(t *testing.T) {
	axis := NewXAxis()
	axis.Side = AxisTop
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -6, W: 5, H: 8},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: the top label sits above the tick (anchor
	// tickPos.Y+labelPad) with a Bottom vAlign baseline offset of +Descent.
	want := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y + labelPad + 2,
	}
	if r.origins[0] != want {
		t.Fatalf("top x tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTickLabelPositionUsesCenterBaselineForRightYAxis(t *testing.T) {
	axis := NewYAxis()
	axis.Side = AxisRight
	axis.Locator = staticLocator{4}
	axis.Formatter = ScalarFormatter{Prec: 0}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"4": {X: 1, Y: -6, W: 5, H: 8},
	}
	r.useFontHeights = true
	r.fontHeights = render.FontHeightMetrics{Ascent: 8, Descent: 2}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.origins) != 1 {
		t.Fatalf("expected one tick label draw, got %d", len(r.origins))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: getSpinePosition(axis, ctx), Y: 4})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: center-baseline y label baseline offset is -Ascent/2 = -4.
	want := geom.Pt{
		X: tickPos.X + labelPad,
		Y: tickPos.Y - 4,
	}
	if r.origins[0] != want {
		t.Fatalf("right y tick origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestAxis_DrawTickLabels_UsesRotatedDrawerWhenRequested(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = ScalarFormatter{Prec: 0}
	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"2": {X: 1, Y: -8, W: 5, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 0 {
		t.Fatalf("expected rotated tick labels to bypass DrawText, got %v", r.texts)
	}
	if len(r.rotatedAnchors) != 1 {
		t.Fatalf("expected one rotated tick label draw, got %d", len(r.rotatedAnchors))
	}

	tickPos := ctx.DataToPixel.Apply(geom.Pt{X: 2, Y: getSpinePosition(axis, ctx)})
	labelPad := tickLabelPadPx(axis, ctx)
	// Display space is y-up: bottom label below the tick with Top vAlign (-Ascent).
	origin := geom.Pt{
		X: tickPos.X - 5.0/2.0,
		Y: tickPos.Y - labelPad - 8,
	}
	layout := measureSingleLineTextLayout(&r, "2", tickLabelFontSize(axis, ctx), ctx.RC.FontKey)
	want := tickLabelRotationAnchor(origin, layout, TextAlignCenter, textLayoutVAlignTop, math.Pi/4)
	if r.rotatedAnchors[0] != want {
		t.Fatalf("rotated tick label anchor = %+v, want %+v", r.rotatedAnchors[0], want)
	}
}

func TestAxis_TickLabelBoundsIncludeRotatedLayout(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{"rotated-label"}}

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"rotated-label": {X: 0, Y: -10, W: 80, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	unrotated, ok := axisTickLabelBounds(axis, &r, ctx)
	if !ok {
		t.Fatal("expected unrotated tick label bounds")
	}

	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}
	rotated, ok := axisTickLabelBounds(axis, &r, ctx)
	if !ok {
		t.Fatal("expected rotated tick label bounds")
	}

	// matplotlib's rotation_mode="default" with valign="top" (the AutoAlign
	// value for a bottom x-axis) pins the rotated bounding box's top edge to the
	// anchor and lets the label extend downward and sideways — it never grows
	// above the unrotated top. (Verified against matplotlib: a rotated bottom
	// tick label keeps its top y unchanged and only its bottom y increases.)
	// In this y-up coordinate the pinned top edge is Max.Y and the downward
	// extent is Min.Y.
	if rotated.Max.Y > unrotated.Max.Y+1 {
		t.Fatalf("rotated bounds extended above the pinned top edge: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
	if rotated.Min.Y >= unrotated.Min.Y-10 {
		t.Fatalf("rotated bounds did not extend below the unrotated layout: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
	if (rotated.Max.Y - rotated.Min.Y) <= (unrotated.Max.Y - unrotated.Min.Y) {
		t.Fatalf("rotated bounds were not taller than unrotated: rotated=%+v unrotated=%+v", rotated, unrotated)
	}
}

func TestAxis_DrawTickLabels_RendersFullMathAsPathsWhenRotated(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{`$\\frac{1}{2}$`}}
	axis.MajorLabelStyle = TickLabelStyle{Rotation: 45, AutoAlign: true}

	var r axisLabelRecordingRenderer
	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 0 {
		t.Fatalf("expected rotated math tick labels to bypass DrawText, got %v", r.texts)
	}
	if len(r.rotatedText) != 0 {
		t.Fatalf("expected rotated math tick labels to bypass DrawTextRotated, got %v", r.rotatedText)
	}
	if !containsString(r.textPathCalls, "1") || !containsString(r.textPathCalls, "2") {
		t.Fatalf("expected fraction runs to resolve through TextPath, got %v", r.textPathCalls)
	}
	if r.pathCount < 3 {
		t.Fatalf("expected fraction rule plus glyph paths, got %d paths", r.pathCount)
	}
}

func TestAxis_ExtraTickLevelsDrawAdditionalLabels(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = FixedFormatter{Labels: []string{"major"}}
	axis.ClearTickLevels()
	axis.AddTickLevel(TickLevel{
		Locator:    staticLocator{2},
		Formatter:  FixedFormatter{Labels: []string{"minor row"}},
		ShowLabels: true,
		LabelStyle: TickLabelStyle{Pad: 14, AutoAlign: true},
	})

	var r axisLabelRecordingRenderer
	r.useBounds = true
	r.bounds = map[string]render.TextBounds{
		"major":     {X: 1, Y: -8, W: 20, H: 10},
		"minor row": {X: 1, Y: -8, W: 35, H: 10},
	}

	ctx := createTestDrawContext()
	ctx.RC.DPI = 72

	if err := r.Begin(geom.Rect{}); err != nil {
		t.Fatalf("Begin: %v", err)
	}
	axis.DrawTickLabels(&r, ctx)
	if err := r.End(); err != nil {
		t.Fatalf("End: %v", err)
	}

	if len(r.texts) != 2 {
		t.Fatalf("expected major and extra tick labels, got %v", r.texts)
	}
	if r.texts[0] != "major" || r.texts[1] != "minor row" {
		t.Fatalf("unexpected tick label sequence: %v", r.texts)
	}
	// Display space is y-up: a bottom-axis extra level farther below sits at a
	// smaller Y than the major level.
	if !(r.origins[1].Y < r.origins[0].Y) {
		t.Fatalf("expected extra tick level to be farther from the axis: major=%+v extra=%+v", r.origins[0], r.origins[1])
	}
}

func TestAxes_AddGrid(t *testing.T) {
	// Test adding grids to axes
	axes := &Axes{
		Artists: []Artist{},
		XAxis:   NewXAxis(),
		YAxis:   NewYAxis(),
	}

	initialCount := len(axes.Artists)

	// Add X grid
	xGrid := axes.AddXGrid()
	if len(axes.Artists) != initialCount+1 {
		t.Errorf("AddXGrid should add one artist, got %d artists", len(axes.Artists))
	}
	if xGrid.Axis != AxisBottom {
		t.Errorf("AddXGrid should create grid for AxisBottom, got %v", xGrid.Axis)
	}

	// Add Y grid
	yGrid := axes.AddYGrid()
	if len(axes.Artists) != initialCount+2 {
		t.Errorf("AddYGrid should add second artist, got %d artists", len(axes.Artists))
	}
	if yGrid.Axis != AxisLeft {
		t.Errorf("AddYGrid should create grid for AxisLeft, got %v", yGrid.Axis)
	}
}

func TestDrawFigure_ExplicitTopRightAxesSuppressFrameFallback(t *testing.T) {
	fig := NewFigure(240, 180)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	top := ax.TopAxis()
	top.ShowTicks = false
	top.ShowLabels = false

	right := ax.RightAxis()
	right.ShowTicks = false
	right.ShowLabels = false

	r := &recordingRenderer{}
	DrawFigure(fig, r)

	if got := len(r.pathCalls); got != 4 {
		t.Fatalf("expected exactly four spine paths with explicit top/right axes, got %d", got)
	}
}
