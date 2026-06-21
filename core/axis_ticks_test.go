package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

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

func TestAxisDefaultMinorTickLineWidthMatchesMatplotlib(t *testing.T) {
	want := 0.6 * 100.0 / 72.0

	for name, axis := range map[string]*Axis{
		"x": NewXAxis(),
		"y": NewYAxis(),
	} {
		if got := axis.MinorTickLineWidth; got != want {
			t.Fatalf("%s axis default minor tick width = %v, want %v", name, got, want)
		}
		if got := axis.minorTickLineWidth(); got != want {
			t.Fatalf("%s axis resolved minor tick width = %v, want %v", name, got, want)
		}
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
	if axes.XAxis.TickLineWidth != 0 || axes.XAxis.MinorTickLineWidth != 0.6*100.0/72.0 {
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
