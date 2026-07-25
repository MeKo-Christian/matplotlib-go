package pyplot

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/dates"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestStatefulHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	if _, err := Plot([]float64{0, 1, 2}, []float64{1, 2, 3}, core.PlotOptions{Label: "line"}); err != nil {
		t.Fatalf("Plot() returned error: %v", err)
	}
	Title("Demo")
	XLabel("time")
	YLabel("value")
	Suptitle("Figure Demo")
	SupXLabel("shared time")
	SupYLabel("shared value")
	legend := Legend()
	if legend == nil {
		t.Fatal("Legend() returned nil")
	}
	figLegend := FigLegend()
	if figLegend == nil {
		t.Fatal("FigLegend() returned nil")
	}

	ax := GCA()
	if ax.Title != "Demo" {
		t.Fatalf("ax.Title = %q, want %q", ax.Title, "Demo")
	}
	if ax.XLabel != "time" || ax.YLabel != "value" {
		t.Fatalf("axis labels = (%q, %q), want (%q, %q)", ax.XLabel, ax.YLabel, "time", "value")
	}
	Box(false)
	if ax.ShowFrame {
		t.Fatal("Box(false) did not hide the current axes frame")
	}
	Box(true)
	if !ax.ShowFrame {
		t.Fatal("Box(true) did not show the current axes frame")
	}
	fig := GCF()
	if fig.SupTitle != "Figure Demo" || fig.SupXLabel != "shared time" || fig.SupYLabel != "shared value" {
		t.Fatalf("figure labels = (%q, %q, %q)", fig.SupTitle, fig.SupXLabel, fig.SupYLabel)
	}
	if figLegend.Figure != fig || figLegend.Axes != nil {
		t.Fatalf("figure legend ownership = figure %p axes %p, want figure %p axes nil", figLegend.Figure, figLegend.Axes, fig)
	}
	if len(fig.Artists) != 1 || fig.Artists[0] != figLegend {
		t.Fatalf("figure artists = %+v, want fig legend", fig.Artists)
	}
	if len(ax.Artists) != 2 {
		t.Fatalf("len(ax.Artists) = %d, want 2", len(ax.Artists))
	}
}

// TestPyplotWrappersShareCoreAxesPath proves the Phase 17.6.7 contract that
// stateful pyplot wrappers route through GCA()/GCF() to the same core
// implementation as the object-oriented API: the artist a wrapper returns is the
// exact artist the current axes appended, and a direct GCA() call produces an
// equivalent artist through the same path.
func TestPyplotWrappersShareCoreAxesPath(t *testing.T) {
	resetForTests()

	x := []float64{0, 1, 2}
	y := []float64{1, 2, 3}

	// The artist returned by the pyplot wrapper must be identical to the one the
	// current axes appended -- i.e. the wrapper added nothing of its own and
	// delegated straight to Axes.Plot.
	wrapperLine, err := Plot(x, y)
	if err != nil {
		t.Fatalf("Plot() returned error: %v", err)
	}
	ax := GCA()
	if wrapperLine == nil {
		t.Fatal("Plot() returned nil")
	}
	if len(ax.Artists) != 1 || ax.Artists[0] != wrapperLine {
		t.Fatalf("pyplot.Plot did not append its returned artist to GCA(): artists=%d", len(ax.Artists))
	}

	// Calling the object-oriented API on the same current axes must use the same
	// path and append an equivalent *core.Line2D, increasing the artist count by
	// exactly one.
	directLine, _ := ax.Plot(x, y)
	if directLine == nil {
		t.Fatal("GCA().Plot() returned nil")
	}
	if len(ax.Artists) != 2 || ax.Artists[1] != directLine {
		t.Fatalf("GCA().Plot did not append exactly one artist: artists=%d", len(ax.Artists))
	}
	if directLine == wrapperLine {
		t.Fatal("expected distinct Line2D artists for the two calls")
	}

	// State mutators must target the same GCA() fields as their OO counterparts.
	resetForTests()
	Title("via pyplot")
	if got := GCA().Title; got != "via pyplot" {
		t.Fatalf("pyplot.Title delegated to %q, want %q", got, "via pyplot")
	}
	GCA().SetTitle("via axes")
	if got := GCA().Title; got != "via axes" {
		t.Fatalf("GCA().SetTitle set %q, want %q; wrappers must share the same field", got, "via axes")
	}

	// Figure-level wrappers must target the current figure that GCF() returns.
	resetForTests()
	Suptitle("shared")
	if got := GCF().SupTitle; got != "shared" {
		t.Fatalf("pyplot.Suptitle delegated to %q, want %q", got, "shared")
	}
}

func TestPlotAcceptsUnitCapableValues(t *testing.T) {
	resetForTests()
	x := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
	}

	line, err := Plot(x, []float64{1, 2})
	if err != nil {
		t.Fatalf("Plot() returned error: %v", err)
	}
	if line == nil {
		t.Fatal("Plot() returned nil line")
	}
	if _, ok := GCA().XAxis.Locator.(dates.DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want dates.DateLocator", GCA().XAxis.Locator)
	}
}

func TestScatterAcceptsUnitCapableValuesAndPropagatesErrors(t *testing.T) {
	resetForTests()

	scatter, err := Scatter([]string{"draft", "review"}, []float64{0.3, 0.8})
	if err != nil {
		t.Fatalf("Scatter() returned error: %v", err)
	}
	if scatter == nil {
		t.Fatal("Scatter() returned nil artist")
	}
	if _, ok := GCA().XAxis.Locator.(ticker.FixedLocator); !ok {
		t.Fatalf("x-axis locator = %T, want ticker.FixedLocator", GCA().XAxis.Locator)
	}

	if rejected, err := Scatter([]float64{0, 1}, []float64{1}); err == nil || rejected != nil {
		t.Fatalf("mismatched Scatter() = (%v, %v), want nil artist and error", rejected, err)
	}
}

func TestBarAcceptsUnitCapableValuesAndPropagatesErrors(t *testing.T) {
	resetForTests()

	bar, err := Bar([]string{"draft", "review"}, []float64{1, 2})
	if err != nil {
		t.Fatalf("Bar() returned error: %v", err)
	}
	if bar == nil {
		t.Fatal("Bar() returned nil artist")
	}
	if _, ok := GCA().XAxis.Locator.(ticker.FixedLocator); !ok {
		t.Fatalf("x-axis locator = %T, want ticker.FixedLocator", GCA().XAxis.Locator)
	}

	if rejected, err := BarH([]string{"north", "south"}, []float64{4}); err == nil || rejected != nil {
		t.Fatalf("mismatched BarH() = (%v, %v), want nil artist and error", rejected, err)
	}
}

func TestTextAndAnnotateDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	text := Text(0.2, 0.8, "note", core.TextOptions{
		FontSize: 14,
		HAlign:   core.TextAlignCenter,
	})
	annotation := Annotate("peak", 0.7, 0.3, core.AnnotationOptions{
		OffsetX: 10,
		OffsetY: -12,
	})

	if text == nil {
		t.Fatal("Text() returned nil")
	}
	if annotation == nil {
		t.Fatal("Annotate() returned nil")
	}
	ax := GCA()
	if len(ax.Artists) != 2 {
		t.Fatalf("len(ax.Artists) = %d, want 2", len(ax.Artists))
	}
	if text.Content != "note" || text.Position != (geom.Pt{X: 0.2, Y: 0.8}) {
		t.Fatalf("Text() artist = %+v, want delegated core text", text)
	}
	if annotation.Content != "peak" || annotation.Point != (geom.Pt{X: 0.7, Y: 0.3}) {
		t.Fatalf("Annotate() artist = %+v, want delegated core annotation", annotation)
	}
}

func TestReferenceLineAndSpanHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	hLine := AxHLine(0.25)
	vLine := AxVLine(0.75)
	line := AxLine(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 1, Y: 1})
	slopeLine := AxLineSlope(geom.Pt{X: 0.5, Y: 0.5}, 2)
	hSpan := AxHSpan(0.1, 0.2)
	vSpan := AxVSpan(0.3, 0.4)

	if hLine == nil || vLine == nil || line == nil || slopeLine == nil || hSpan == nil || vSpan == nil {
		t.Fatalf("reference helpers returned nil: h=%v v=%v line=%v slope=%v hspan=%v vspan=%v", hLine, vLine, line, slopeLine, hSpan, vSpan)
	}
	ax := GCA()
	if len(ax.Artists) != 6 {
		t.Fatalf("len(ax.Artists) = %d, want 6", len(ax.Artists))
	}
	if hLine.Start.Y != 0.25 || vLine.Start.X != 0.75 {
		t.Fatalf("line endpoints not delegated through core helpers: h=%+v v=%+v", hLine, vLine)
	}
	if line.Direction != (geom.Pt{X: 1, Y: 1}) || slopeLine.Direction != (geom.Pt{X: 1, Y: 2}) {
		t.Fatalf("axline directions = %+v / %+v, want (1,1) / (1,2)", line.Direction, slopeLine.Direction)
	}
	if hSpan.Start.Y != 0.1 || hSpan.End.Y != 0.2 || vSpan.Start.X != 0.3 || vSpan.End.X != 0.4 {
		t.Fatalf("span endpoints not delegated through core helpers: h=%+v v=%+v", hSpan, vSpan)
	}
}

func TestAxisLimitAndScaleHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	XLim(1, 100)
	YLim(-9, 9)
	if err := XScale("log", transform.WithScaleBase(10)); err != nil {
		t.Fatalf("XScale(log): %v", err)
	}
	if err := YScale("symlog", transform.WithScaleLinThresh(2)); err != nil {
		t.Fatalf("YScale(symlog): %v", err)
	}

	ax := GCA()
	if xMin, xMax := ax.XScale.Domain(); xMin != 1 || xMax != 100 {
		t.Fatalf("x domain = (%v, %v), want (1, 100)", xMin, xMax)
	}
	if yMin, yMax := ax.YScale.Domain(); yMin != -9 || yMax != 9 {
		t.Fatalf("y domain = (%v, %v), want (-9, 9)", yMin, yMax)
	}
	if _, ok := ax.XScale.(transform.Log); !ok {
		t.Fatalf("x scale = %T, want transform.Log", ax.XScale)
	}
	if _, ok := ax.YScale.(transform.SymLog); !ok {
		t.Fatalf("y scale = %T, want transform.SymLog", ax.YScale)
	}
}

func TestAxisModeHelperDelegatesToCurrentAxes(t *testing.T) {
	resetForTests()

	ax := GCA()
	if err := Axis("off"); err != nil {
		t.Fatalf("Axis(off) error = %v", err)
	}
	if ax.ShowFrame || ax.XAxis.ShowTicks || ax.YAxis.ShowTicks || ax.XAxis.ShowLabels || ax.YAxis.ShowLabels {
		t.Fatalf("Axis(off) left visible frame/ticks/labels: frame=%v xTicks=%v yTicks=%v xLabels=%v yLabels=%v",
			ax.ShowFrame, ax.XAxis.ShowTicks, ax.YAxis.ShowTicks, ax.XAxis.ShowLabels, ax.YAxis.ShowLabels)
	}

	if err := Axis("on"); err != nil {
		t.Fatalf("Axis(on) error = %v", err)
	}
	if !ax.ShowFrame || !ax.XAxis.ShowTicks || !ax.YAxis.ShowTicks || !ax.XAxis.ShowLabels || !ax.YAxis.ShowLabels {
		t.Fatalf("Axis(on) did not restore frame/ticks/labels: frame=%v xTicks=%v yTicks=%v xLabels=%v yLabels=%v",
			ax.ShowFrame, ax.XAxis.ShowTicks, ax.YAxis.ShowTicks, ax.XAxis.ShowLabels, ax.YAxis.ShowLabels)
	}

	if err := Axis("equal"); err != nil {
		t.Fatalf("Axis(equal) error = %v", err)
	}
	if err := Axis("auto"); err != nil {
		t.Fatalf("Axis(auto) error = %v", err)
	}
	if err := Axis("image"); err == nil {
		t.Fatal("Axis(image) returned nil error, want unsupported mode error")
	}
}

func TestGridAndTickParamsDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	gridColor := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	gridWidth := 2.5
	grids, err := Grid(true, core.TickParams{
		Axis:          "both",
		Which:         "both",
		GridColor:     &gridColor,
		GridLineWidth: &gridWidth,
	})
	if err != nil {
		t.Fatalf("Grid() error = %v", err)
	}
	if len(grids) != 2 {
		t.Fatalf("Grid() returned %d grids, want 2", len(grids))
	}
	for _, grid := range grids {
		if !grid.Major || !grid.Minor {
			t.Fatalf("grid visibility = major:%v minor:%v, want both true", grid.Major, grid.Minor)
		}
		if grid.Color != gridColor || grid.MinorColor != gridColor {
			t.Fatalf("grid colors = %+v / %+v, want %+v", grid.Color, grid.MinorColor, gridColor)
		}
		if grid.LineWidth != gridWidth || grid.MinorLineWidth != gridWidth {
			t.Fatalf("grid widths = %v / %v, want %v", grid.LineWidth, grid.MinorLineWidth, gridWidth)
		}
	}

	showLabels := false
	tickLength := 7.25
	if err := TickParams(core.TickParams{
		Axis:       "x",
		ShowLabels: &showLabels,
		Length:     &tickLength,
	}); err != nil {
		t.Fatalf("TickParams() error = %v", err)
	}

	ax := GCA()
	if ax.XAxis.ShowLabels {
		t.Fatal("TickParams() did not update x tick label visibility")
	}
	if !ax.YAxis.ShowLabels {
		t.Fatal("TickParams(axis=x) unexpectedly changed y tick label visibility")
	}
	if ax.XAxis.TickSize != tickLength {
		t.Fatalf("x tick length = %v, want %v", ax.XAxis.TickSize, tickLength)
	}

	if _, err := Grid(true, core.TickParams{Axis: "diagonal"}); err == nil {
		t.Fatal("Grid() with unsupported axis returned nil error")
	}
}

func TestMinorTickAndLocatorWrappersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	if err := MinorTicksOn("x"); err != nil {
		t.Fatalf("MinorTicksOn(x) error = %v", err)
	}
	ax := GCA()
	if ax.XAxis.MinorLocator == nil {
		t.Fatal("MinorTicksOn(x) did not enable x minor locator")
	}
	if ax.YAxis.MinorLocator != nil {
		t.Fatal("MinorTicksOn(x) unexpectedly enabled y minor locator")
	}

	if err := LocatorParams(core.LocatorParams{Axis: "x", MajorCount: 6, MinorCount: 24}); err != nil {
		t.Fatalf("LocatorParams(x) error = %v", err)
	}
	if ax.XAxis.MajorTickCount != 6 || ax.XAxis.MinorTickCount != 24 {
		t.Fatalf("x locator counts = major:%d minor:%d, want 6/24", ax.XAxis.MajorTickCount, ax.XAxis.MinorTickCount)
	}
	if ax.YAxis.MajorTickCount == 6 || ax.YAxis.MinorTickCount == 24 {
		t.Fatalf("LocatorParams(x) unexpectedly changed y locator counts: major:%d minor:%d", ax.YAxis.MajorTickCount, ax.YAxis.MinorTickCount)
	}

	if err := MinorTicksOff("x"); err != nil {
		t.Fatalf("MinorTicksOff(x) error = %v", err)
	}
	if ax.XAxis.MinorLocator != nil {
		t.Fatal("MinorTicksOff(x) did not clear x minor locator")
	}
	if err := MinorTicksOn("diagonal"); err == nil {
		t.Fatal("MinorTicksOn(diagonal) returned nil error")
	}
	if err := LocatorParams(core.LocatorParams{Axis: "diagonal"}); err == nil {
		t.Fatal("LocatorParams(diagonal) returned nil error")
	}
}

func TestTickLocationWrappersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	if err := XTicks([]float64{3, 1, 2}, []string{"three", "one", "two"}); err != nil {
		t.Fatalf("XTicks() error = %v", err)
	}
	if err := YTicks([]float64{-1, 1}); err != nil {
		t.Fatalf("YTicks() error = %v", err)
	}

	ax := GCA()
	xLoc, ok := ax.XAxis.Locator.(ticker.FixedLocator)
	if !ok {
		t.Fatalf("x locator = %T, want ticker.FixedLocator", ax.XAxis.Locator)
	}
	if got := xLoc.TicksList; len(got) != 3 || got[0] != 3 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("x ticks = %v, want [3 1 2]", got)
	}
	xFmt, ok := ax.XAxis.Formatter.(ticker.FixedFormatter)
	if !ok {
		t.Fatalf("x formatter = %T, want ticker.FixedFormatter", ax.XAxis.Formatter)
	}
	if got := xFmt.Labels; len(got) != 3 || got[0] != "three" || got[1] != "one" || got[2] != "two" {
		t.Fatalf("x labels = %v, want [three one two]", got)
	}

	yLoc, ok := ax.YAxis.Locator.(ticker.FixedLocator)
	if !ok {
		t.Fatalf("y locator = %T, want ticker.FixedLocator", ax.YAxis.Locator)
	}
	if got := yLoc.TicksList; len(got) != 2 || got[0] != -1 || got[1] != 1 {
		t.Fatalf("y ticks = %v, want [-1 1]", got)
	}
	if _, ok := ax.YAxis.Formatter.(ticker.FixedFormatter); ok {
		t.Fatal("YTicks without labels unexpectedly installed FixedFormatter")
	}

	if err := XTicks([]float64{1}, []string{"one"}, []string{"extra"}); err == nil {
		t.Fatal("XTicks with multiple label sets returned nil error")
	}
	if err := YTicks([]float64{1, 2}, []string{"one"}); err == nil {
		t.Fatal("YTicks with mismatched labels returned nil error")
	}
}

func TestTickLabelFormatUpdatesCurrentScalarFormatters(t *testing.T) {
	resetForTests()

	ax := GCA()
	yBefore, ok := ax.YAxis.Formatter.(ticker.ScalarFormatter)
	if !ok {
		t.Fatalf("initial y formatter = %T, want ticker.ScalarFormatter", ax.YAxis.Formatter)
	}
	useMathText := true
	sciLimits := [2]int{-2, 3}
	if err := TickLabelFormat(TickLabelFormatOptions{
		Axis:        "x",
		Style:       "plain",
		SciLimits:   &sciLimits,
		UseMathText: &useMathText,
	}); err != nil {
		t.Fatalf("TickLabelFormat(x/plain) error = %v", err)
	}

	xFmt, ok := ax.XAxis.Formatter.(ticker.ScalarFormatter)
	if !ok {
		t.Fatalf("x formatter = %T, want ticker.ScalarFormatter", ax.XAxis.Formatter)
	}
	if !xFmt.DisableScientific || !xFmt.UseMathText || !xFmt.UsePowerLimits || xFmt.PowerLimits != sciLimits {
		t.Fatalf("x scalar formatter = %+v, want plain mathtext with limits %+v", xFmt, sciLimits)
	}

	yFmt, ok := ax.YAxis.Formatter.(ticker.ScalarFormatter)
	if !ok {
		t.Fatalf("y formatter = %T, want ticker.ScalarFormatter", ax.YAxis.Formatter)
	}
	if yFmt != yBefore {
		t.Fatalf("TickLabelFormat(x) changed y scalar formatter: got %+v, want %+v", yFmt, yBefore)
	}

	if err := TickLabelFormat(TickLabelFormatOptions{Axis: "both", Style: "scientific"}); err != nil {
		t.Fatalf("TickLabelFormat(both/scientific) error = %v", err)
	}
	xFmt = ax.XAxis.Formatter.(ticker.ScalarFormatter)
	yFmt = ax.YAxis.Formatter.(ticker.ScalarFormatter)
	if xFmt.DisableScientific || yFmt.DisableScientific {
		t.Fatalf("scientific style did not re-enable scientific formatting: x=%+v y=%+v", xFmt, yFmt)
	}

	if err := TickLabelFormat(TickLabelFormatOptions{Axis: "z"}); err == nil {
		t.Fatal("TickLabelFormat(z) returned nil error")
	}

	ax.XAxis.Formatter = ticker.FixedFormatter{Labels: []string{"fixed"}}
	if err := TickLabelFormat(TickLabelFormatOptions{Axis: "x", Style: "plain"}); err == nil {
		t.Fatal("TickLabelFormat on FixedFormatter returned nil error")
	}
}

func TestConveniencePlotHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	dateValues := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
	}
	if line, err := PlotDate(dateValues, []float64{1, 2}); err != nil || line == nil {
		t.Fatal("PlotDate() returned nil")
	}
	if _, ok := GCA().XAxis.Locator.(dates.DateLocator); !ok {
		t.Fatalf("PlotDate x-axis locator = %T, want DateLocator", GCA().XAxis.Locator)
	}

	resetForTests()
	if line := SemilogX([]float64{1, 10}, []float64{1, 2}); line == nil {
		t.Fatal("SemilogX() returned nil")
	}
	if _, ok := GCA().XScale.(transform.Log); !ok {
		t.Fatalf("SemilogX x scale = %T, want transform.Log", GCA().XScale)
	}

	resetForTests()
	if line := SemilogY([]float64{1, 2}, []float64{1, 10}); line == nil {
		t.Fatal("SemilogY() returned nil")
	}
	if _, ok := GCA().YScale.(transform.Log); !ok {
		t.Fatalf("SemilogY y scale = %T, want transform.Log", GCA().YScale)
	}

	resetForTests()
	if line := LogLog([]float64{1, 10}, []float64{1, 10}); line == nil {
		t.Fatal("LogLog() returned nil")
	}
	if _, ok := GCA().XScale.(transform.Log); !ok {
		t.Fatalf("LogLog x scale = %T, want transform.Log", GCA().XScale)
	}
	if _, ok := GCA().YScale.(transform.Log); !ok {
		t.Fatalf("LogLog y scale = %T, want transform.Log", GCA().YScale)
	}

	resetForTests()
	if bar, err := BarH([]float64{0, 1}, []float64{3, 4}); err != nil || bar == nil || bar.Orientation != core.BarHorizontal {
		t.Fatalf("BarH() = (%#v, %v), want horizontal bar", bar, err)
	}
	if fill := Fill([]float64{0, 1, 0}, []float64{0, 0, 1}); fill == nil {
		t.Fatal("Fill() returned nil")
	}
	if fill := FillBetweenX([]float64{0, 1, 2}, []float64{0, 1, 0}, []float64{1, 2, 1}); fill == nil || fill.Orientation != core.FillHorizontal {
		t.Fatalf("FillBetweenX() = %#v, want horizontal fill", fill)
	}
	arrow := Arrow(0.2, 0.3, 1.5, -0.5)
	if arrow == nil {
		t.Fatal("Arrow() returned nil")
	}
	if arrow.XY != (geom.Pt{X: 0.2, Y: 0.3}) || arrow.DX != 1.5 || arrow.DY != -0.5 {
		t.Fatalf("Arrow geometry = xy=%+v dx=%v dy=%v", arrow.XY, arrow.DX, arrow.DY)
	}
	hLines := HLines([]float64{1, 2}, []float64{0, 0.5}, []float64{3, 3.5})
	if hLines == nil || len(hLines.Segments) != 2 {
		t.Fatalf("HLines() = %#v, want two segments", hLines)
	}
	if hLines.Segments[1][0] != (geom.Pt{X: 0.5, Y: 2}) || hLines.Segments[1][1] != (geom.Pt{X: 3.5, Y: 2}) {
		t.Fatalf("HLines second segment = %+v", hLines.Segments[1])
	}
	if hLinesBroadcast := HLines([]float64{3, 4}, []float64{-1}, []float64{1}); hLinesBroadcast == nil || len(hLinesBroadcast.Segments) != 2 {
		t.Fatalf("HLines broadcast = %#v, want two segments", hLinesBroadcast)
	}
	vLines := VLines([]float64{1, 2}, []float64{-1, -2}, []float64{1, 2})
	if vLines == nil || len(vLines.Segments) != 2 {
		t.Fatalf("VLines() = %#v, want two segments", vLines)
	}
	if vLines.Segments[0][0] != (geom.Pt{X: 1, Y: -1}) || vLines.Segments[0][1] != (geom.Pt{X: 1, Y: 1}) {
		t.Fatalf("VLines first segment = %+v", vLines.Segments[0])
	}
	if vLinesBroadcast := VLines([]float64{3, 4}, []float64{-2}, []float64{2}); vLinesBroadcast == nil || len(vLinesBroadcast.Segments) != 2 {
		t.Fatalf("VLines broadcast = %#v, want two segments", vLinesBroadcast)
	}
	if step := Step([]float64{0, 1, 2}, []float64{1, 3, 2}); step == nil {
		t.Fatal("Step() returned nil")
	}
	if stairs := Stairs([]float64{1, 2}, []float64{0, 1, 2}); stairs == nil {
		t.Fatal("Stairs() returned nil")
	}
	broken := BrokenBarH([][2]float64{{1, 2}, {4, 1}}, [2]float64{0.5, 0.25})
	if broken == nil {
		t.Fatal("BrokenBarH() returned nil")
	}
	if labels := BarLabel(broken, []string{"one", "two"}); len(labels) != 2 {
		t.Fatalf("BarLabel() returned %d labels, want 2", len(labels))
	}
	if box := BoxPlot([]float64{1, 2, 3, 4}); box == nil {
		t.Fatal("BoxPlot() returned nil")
	}
	if bxp := Bxp([]core.BxpStat{{Med: 2, Q1: 1, Q3: 3, Whislo: 0, Whishi: 4}}); bxp == nil || len(bxp.Medians) != 1 {
		t.Fatalf("Bxp() = %#v, want one median", bxp)
	}
	if fills := StackPlot([]float64{0, 1}, [][]float64{{1, 2}, {2, 1}}); len(fills) != 2 {
		t.Fatalf("StackPlot() returned %d fills, want 2", len(fills))
	}
	if ecdf := ECDF([]float64{3, 1, 2}); ecdf == nil {
		t.Fatal("ECDF() returned nil")
	}
	pie := Pie([]float64{1, 2}, core.PieOptions{Labels: []string{"A", "B"}})
	if pie == nil {
		t.Fatal("Pie() returned nil")
	}
	if labels := PieLabel(pie, []string{"one", "two"}); len(labels) != 2 {
		t.Fatalf("PieLabel() returned %d labels, want 2", len(labels))
	}
	violin := Violin([]core.ViolinStat{{
		Coords: []float64{1, 2, 3},
		Vals:   []float64{0.2, 1, 0.2},
		Mean:   2,
		Median: 2,
		Min:    1,
		Max:    3,
	}})
	if violin == nil || violin.Bodies == nil || len(violin.Bodies.Polygons) != 1 {
		t.Fatalf("Violin() = %#v, want one body", violin)
	}

	contours := Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, core.ContourOptions{Levels: []float64{2}})
	if contours == nil {
		t.Fatal("Contour() returned nil")
	}
	if labels := Clabel(contours, core.ClabelOptions{Levels: []float64{2}}); len(labels) != 1 {
		t.Fatalf("Clabel() returned %d labels, want 1", len(labels))
	}

	ax := GCA()
	XLim(-10, 10)
	YLim(-10, 10)
	AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()
	if xMin > 0 || xMax < 4 || yMin > 0 || yMax < 4 {
		t.Fatalf("AutoScale() domains = x[%v,%v] y[%v,%v], want coverage for current artists", xMin, xMax, yMin, yMax)
	}
}

func TestImageIOWrappersDelegateToCoreHelpers(t *testing.T) {
	resetForTests()

	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 0x20, G: 0x40, B: 0x60, A: 0xff})
	src.SetRGBA(1, 0, color.RGBA{R: 0x80, G: 0xa0, B: 0xc0, A: 0xff})

	path := filepath.Join(t.TempDir(), "pyplot-image.png")
	if err := ImSave(path, render.NewImageData(src)); err != nil {
		t.Fatalf("ImSave() error = %v", err)
	}

	got, err := ImRead(path)
	if err != nil {
		t.Fatalf("ImRead() error = %v", err)
	}
	if got == nil || got.RGBA() == nil {
		t.Fatal("ImRead() returned nil image data")
	}
	if bounds := got.RGBA().Bounds(); bounds.Dx() != 2 || bounds.Dy() != 1 {
		t.Fatalf("read bounds = %v, want 2x1", bounds)
	}
	if px := got.RGBA().RGBAAt(1, 0); px.R != 0x80 || px.G != 0xa0 || px.B != 0xc0 || px.A != 0xff {
		t.Fatalf("read pixel = %#v, want source pixel", px)
	}
}

func TestGetCMapDelegatesToColorRegistry(t *testing.T) {
	if got := CMap("plasma").Name(); got != "plasma" {
		t.Fatalf("CMap(plasma).Name() = %q, want plasma", got)
	}
	if got := CMap("does-not-exist").Name(); got != "viridis" {
		t.Fatalf("CMap(unknown).Name() = %q, want viridis fallback", got)
	}
}

func TestVectorFieldHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	q := Quiver(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 0.5},
		[]float64{0.25, 0.75},
		core.QuiverOptions{Label: "q"},
	)
	if q == nil {
		t.Fatal("Quiver() returned nil")
	}
	key := QuiverKey(q, 0.8, 0.2, 1, "1 unit")
	if key == nil {
		t.Fatal("QuiverKey() returned nil")
	}
	barbs := Barbs(
		[]float64{0.5},
		[]float64{0.5},
		[]float64{12},
		[]float64{3},
	)
	if barbs == nil {
		t.Fatal("Barbs() returned nil")
	}
	stream := Streamplot(
		[]float64{0, 1, 2},
		[]float64{0, 1},
		[][]float64{{1, 1, 1}, {1, 1, 1}},
		[][]float64{{0, 0.2, 0.2}, {0, 0.2, 0.2}},
		core.StreamplotOptions{StartPoints: []geom.Pt{{X: 0.2, Y: 0.4}}},
	)
	if stream == nil {
		t.Fatal("Streamplot() returned nil")
	}

	ax := GCA()
	if len(ax.Artists) != 4 {
		t.Fatalf("len(ax.Artists) = %d, want 4", len(ax.Artists))
	}
}

func TestMatrixAndSignalHelpersDelegateToCurrentAxes(t *testing.T) {
	resetForTests()

	mat := [][]float64{
		{0, 1},
		{2, 3},
	}
	bilinear := "bilinear"
	if img := ImShow(mat, core.ImShowOptions{Interpolation: &bilinear}); img == nil || img.Interpolation != "bilinear" {
		t.Fatalf("ImShow() = %#v, want image with bilinear interpolation", img)
	}
	if img := MatShow(mat); img == nil {
		t.Fatal("MatShow() returned nil")
	}
	if spy := Spy(mat); spy == nil {
		t.Fatal("Spy() returned nil")
	}
	if mesh := PColor(mat, core.MeshOptions{Label: "pcolor"}); mesh == nil || mesh.Label != "pcolor" {
		t.Fatalf("PColor() = %#v, want labelled mesh", mesh)
	}
	if mesh := PColorFast(mat, core.MeshOptions{Label: "fast"}); mesh == nil || mesh.Label != "fast" {
		t.Fatalf("PColorFast() = %#v, want labelled mesh", mesh)
	}
	if mesh := PColorMesh(mat, core.MeshOptions{Label: "mesh"}); mesh == nil || mesh.Label != "mesh" {
		t.Fatalf("PColorMesh() = %#v, want labelled mesh", mesh)
	}
	if spec := Specgram([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SpecgramOptions{NFFT: 4}); spec == nil {
		t.Fatal("Specgram() returned nil")
	}
	if psd := PSD([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); psd == nil {
		t.Fatal("PSD() returned nil")
	}
	if mag := MagnitudeSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); mag == nil {
		t.Fatal("MagnitudeSpectrum() returned nil")
	}
	if angle := AngleSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); angle == nil {
		t.Fatal("AngleSpectrum() returned nil")
	}
	if phase := PhaseSpectrum([]float64{0, 1, 0, -1, 0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); phase == nil {
		t.Fatal("PhaseSpectrum() returned nil")
	}
	if csd := CSD([]float64{0, 1, 0, -1}, []float64{0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); csd == nil {
		t.Fatal("CSD() returned nil")
	}
	if cohere := Cohere([]float64{0, 1, 0, -1}, []float64{0, 1, 0, -1}, core.SignalSpectrumOptions{NFFT: 4}); cohere == nil {
		t.Fatal("Cohere() returned nil")
	}
	if xcorr := XCorr([]float64{1, 2, 3}, []float64{1, 2, 3}); xcorr == nil {
		t.Fatal("XCorr() returned nil")
	}
	if acorr := ACorr([]float64{1, 2, 3}); acorr == nil {
		t.Fatal("ACorr() returned nil")
	}
	if heatmap := AnnotatedHeatmap(mat); heatmap == nil {
		t.Fatal("AnnotatedHeatmap() returned nil")
	}
}

func TestThreeDHelpersDelegateToCurrent3DAxes(t *testing.T) {
	resetForTests()

	ax := AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if ax == nil {
		t.Fatal("AddAxes3D() returned nil")
	}

	if line := Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1}); line == nil {
		t.Fatal("Plot3D() returned nil")
	}
	if scatter := Scatter3D([]float64{0, 1}, []float64{0, 1}, []float64{1, 2}); scatter == nil {
		t.Fatal("Scatter3D() returned nil")
	}
	if wire := Wireframe([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); wire == nil {
		t.Fatal("Wireframe() returned nil")
	}
	if surf := Surface([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); surf == nil {
		t.Fatal("Surface() returned nil")
	}
	if voxel := Voxel([]float64{0, 1}, []float64{0, 1}, []float64{0, 1}, []float64{1, 1}, []float64{1, 1}, []float64{1, 1}); voxel == nil {
		t.Fatal("Voxel() returned nil")
	}
	tri := core.Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}
	if tris := Trisurf(tri, []float64{0, 1, 2, 3}); tris == nil {
		t.Fatal("Trisurf() returned nil")
	}
	if contour := Contour3D([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); contour == nil {
		t.Fatal("Contour3D() returned nil")
	}
	if contourf := Contourf3D([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}); contourf == nil {
		t.Fatal("Contourf3D() returned nil")
	}
	if text := Text3D(0.2, 0.5, 0.9, "pt"); text == nil {
		t.Fatal("Text3D() returned nil")
	}
}

func TestParasiteAxesWrapperCreatesOverlayAxesAndSetsCurrent(t *testing.T) {
	resetForTests()

	fig := GCF()
	host := GCA()
	if fig == nil || host == nil {
		t.Fatal("initial figure/axes not available")
	}

	parasite := ParasiteAxes()
	if parasite == nil || parasite.Axes == nil {
		t.Fatal("ParasiteAxes() returned nil")
	}
	if got := GCA(); got != parasite.Axes {
		t.Fatalf("GCA() = %p, want %p", got, parasite.Axes)
	}
	if got := len(fig.Children); got != 2 {
		t.Fatalf("len(fig.Children) = %d, want 2", got)
	}
	if parasite.Host != host {
		t.Fatalf("parasite host = %p, want %p", parasite.Host, host)
	}
}
