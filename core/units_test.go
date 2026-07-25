package core

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/dates"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestAxesPlot_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	timestamps := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 4, 0, 0, 0, 0, time.UTC),
	}

	line, err := ax.Plot(timestamps, []float64{2, 3, 5})
	if err != nil {
		t.Fatalf("Plot returned error: %v", err)
	}
	if line == nil {
		t.Fatal("Plot returned nil line")
	}
	if got := line.XY[1].X; got != dates.Date2Num(timestamps[1]) {
		t.Fatalf("converted x[1] = %v, want %v", got, dates.Date2Num(timestamps[1]))
	}
	if _, ok := ax.XAxis.Locator.(dates.DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want dates.DateLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxis.Formatter.(dates.AutoDateFormatter); !ok {
		t.Fatalf("x-axis formatter = %T, want dates.AutoDateFormatter", ax.XAxis.Formatter)
	}

	ax.SetXLim(dates.Date2Num(timestamps[0]), dates.Date2Num(timestamps[len(timestamps)-1]))
	if _, ok := ax.XAxis.Formatter.(dates.AutoDateFormatter); !ok {
		t.Fatalf("x-axis formatter after SetXLim = %T, want dates.AutoDateFormatter", ax.XAxis.Formatter)
	}
}

func TestAxesPlot_RejectedInputIsTransactional(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	originalLocator := ax.XAxis.Locator
	originalFormatter := ax.XAxis.Formatter
	originalCycle := ax.ColorCycle
	originalCycleIndex := originalCycle.Index()
	timestamps := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC),
	}

	if line, err := ax.Plot(timestamps, []struct{ value int }{{1}, {2}}); err == nil || line != nil {
		t.Fatalf("Plot() = (%v, %v), want nil artist and conversion error", line, err)
	}
	if ax.xUnits != nil || ax.yUnits != nil {
		t.Fatalf("rejected Plot configured units: x=%v y=%v", ax.xUnits, ax.yUnits)
	}
	if !reflect.DeepEqual(ax.XAxis.Locator, originalLocator) || !reflect.DeepEqual(ax.XAxis.Formatter, originalFormatter) {
		t.Fatalf("rejected Plot changed x-axis locator/formatter to %T/%T", ax.XAxis.Locator, ax.XAxis.Formatter)
	}
	if len(ax.Artists) != 0 {
		t.Fatalf("rejected Plot added %d artists", len(ax.Artists))
	}
	if ax.ColorCycle != originalCycle || ax.ColorCycle.Index() != originalCycleIndex {
		t.Fatal("rejected Plot replaced or advanced the property cycle")
	}

	if line, err := ax.Plot([]float64{0, 1}, []float64{1}); err == nil || line != nil {
		t.Fatalf("mismatched Plot() = (%v, %v), want nil artist and length error", line, err)
	}
	if len(ax.Artists) != 0 || ax.ColorCycle != originalCycle || ax.ColorCycle.Index() != originalCycleIndex {
		t.Fatal("mismatched Plot changed artists or property cycle")
	}

	if line, err := ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{}, PlotOptions{}); err == nil || line != nil {
		t.Fatalf("multi-option Plot() = (%v, %v), want nil artist and option-count error", line, err)
	}
	if len(ax.Artists) != 0 || ax.ColorCycle != originalCycle || ax.ColorCycle.Index() != originalCycleIndex {
		t.Fatal("multi-option Plot changed artists or property cycle")
	}
}

func TestAxesScatter_ConfiguresCategoricalAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	scatter, err := ax.Scatter([]string{"draft", "review", "ship"}, []float64{0.3, 0.8, 0.5})
	if err != nil {
		t.Fatalf("Scatter returned error: %v", err)
	}
	if scatter == nil {
		t.Fatal("Scatter returned nil artist")
	}
	if got := scatter.XY[2].X; got != 2 {
		t.Fatalf("converted categorical x[2] = %v, want 2", got)
	}
	if _, ok := ax.XAxis.Locator.(ticker.FixedLocator); !ok {
		t.Fatalf("x-axis locator = %T, want ticker.FixedLocator", ax.XAxis.Locator)
	}
}

func TestAxesScatter_RejectedInputIsTransactional(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	originalLocator := ax.XAxis.Locator
	originalFormatter := ax.XAxis.Formatter
	originalCycle := ax.PatchColorCycle
	originalCycleIndex := originalCycle.Index()

	assertUnchanged := func(context string) {
		t.Helper()
		if ax.xUnits != nil || ax.yUnits != nil {
			t.Fatalf("%s configured units: x=%v y=%v", context, ax.xUnits, ax.yUnits)
		}
		if !reflect.DeepEqual(ax.XAxis.Locator, originalLocator) || !reflect.DeepEqual(ax.XAxis.Formatter, originalFormatter) {
			t.Fatalf("%s changed x-axis locator/formatter to %T/%T", context, ax.XAxis.Locator, ax.XAxis.Formatter)
		}
		if len(ax.Artists) != 0 {
			t.Fatalf("%s added %d artists", context, len(ax.Artists))
		}
		if ax.PatchColorCycle != originalCycle || ax.PatchColorCycle.Index() != originalCycleIndex {
			t.Fatalf("%s replaced or advanced the patch property cycle", context)
		}
	}

	if scatter, err := ax.Scatter(
		[]string{"draft", "review"},
		[]float64{0.3, 0.8},
		ScatterOptions{Sizes: []float64{4, 9, 16}},
	); err == nil || scatter != nil {
		t.Fatalf("shape-invalid Scatter() = (%v, %v), want nil artist and error", scatter, err)
	}
	assertUnchanged("shape-invalid Scatter")

	vmin := 0.0
	if scatter, err := ax.Scatter(
		[]string{"draft", "review"},
		[]float64{0.3, 0.8},
		ScatterOptions{
			ScalarValues: []float64{0.3, 0.8},
			Norm:         Normalize{VMin: 0, VMax: 1},
			VMin:         &vmin,
		},
	); err == nil || scatter != nil {
		t.Fatalf("scalar-map-invalid Scatter() = (%v, %v), want nil artist and error", scatter, err)
	}
	assertUnchanged("scalar-map-invalid Scatter")

	if scatter, err := ax.Scatter(
		[]float64{0, 1},
		[]float64{1, 2},
		ScatterOptions{},
		ScatterOptions{},
	); err == nil || scatter != nil {
		t.Fatalf("multi-option Scatter() = (%v, %v), want nil artist and error", scatter, err)
	}
	assertUnchanged("multi-option Scatter")
}

func TestAxesPlotDate_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	timestamps := []time.Time{
		time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 2, 0, 0, 0, 0, time.UTC),
	}

	line, err := ax.PlotDate(timestamps, []float64{1, 4})
	if err != nil {
		t.Fatalf("PlotDate() returned error: %v", err)
	}
	if line == nil {
		t.Fatal("PlotDate() returned nil")
	}
	if got := line.XY[0].X; got != dates.Date2Num(timestamps[0]) {
		t.Fatalf("PlotDate converted x[0] = %v, want %v", got, dates.Date2Num(timestamps[0]))
	}
	if _, ok := ax.XAxis.Locator.(dates.DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want dates.DateLocator", ax.XAxis.Locator)
	}
}

func TestAxesDateUnitsPreserveExplicitAxisInfoAfterAutoscale(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	timestamps := []time.Time{
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC),
	}
	if _, err := ax.Plot(timestamps, []float64{1, 3, 2}); err != nil {
		t.Fatalf("Plot returned error: %v", err)
	}

	ax.XAxis.Locator = dates.DayLocator{ByMonthDay: []int{5, 12, 19}, Location: time.UTC}
	ax.XAxis.Formatter = dates.DateFormatter{Layout: "02 Jan", Location: time.UTC}
	ax.AutoScale(0.06)

	if _, ok := ax.XAxis.Locator.(dates.DayLocator); !ok {
		t.Fatalf("x-axis locator after AutoScale = %T, want dates.DayLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxis.Formatter.(dates.DateFormatter); !ok {
		t.Fatalf("x-axis formatter after AutoScale = %T, want dates.DateFormatter", ax.XAxis.Formatter)
	}
}

func TestDateConverterRCSwitchesDefaultFormatter(t *testing.T) {
	t.Cleanup(style.ResetDefaults)

	state := &axisUnitsState{kind: unitAxisDate}
	info := state.axisInfo(0, 30)
	if _, ok := info.Formatter.(dates.AutoDateFormatter); !ok {
		t.Fatalf("default date formatter = %T, want dates.AutoDateFormatter", info.Formatter)
	}

	if _, err := style.UpdateParams(style.Params{"date.converter": "concise"}); err != nil {
		t.Fatalf("UpdateParams: %v", err)
	}
	info = state.axisInfo(0, 30)
	if _, ok := info.Formatter.(dates.ConciseDateFormatter); !ok {
		t.Fatalf("date.converter: concise formatter = %T, want dates.ConciseDateFormatter", info.Formatter)
	}
}

func TestAxesBar_ConfiguresCategoricalXAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	bar, err := ax.Bar([]string{"alpha", "beta", "gamma"}, []float64{1, 3, 2})
	if err != nil {
		t.Fatalf("Bar returned error: %v", err)
	}
	if bar == nil {
		t.Fatal("Bar returned nil bar")
	}

	wantX := []float64{0, 1, 2}
	for i, want := range wantX {
		if got := bar.X[i]; got != want {
			t.Fatalf("bar x[%d] = %v, want %v", i, got, want)
		}
	}

	loc, ok := ax.XAxis.Locator.(ticker.FixedLocator)
	if !ok {
		t.Fatalf("x-axis locator = %T, want FixedLocator", ax.XAxis.Locator)
	}
	if len(loc.TicksList) != 3 || loc.TicksList[2] != 2 {
		t.Fatalf("categorical ticks = %v, want [0 1 2]", loc.TicksList)
	}

	formatter, ok := ax.XAxis.Formatter.(ticker.FixedFormatter)
	if !ok {
		t.Fatalf("x-axis formatter = %T, want FixedFormatter", ax.XAxis.Formatter)
	}
	if got := formatter.FormatTick(0, 1, loc.TicksList); got != "beta" {
		t.Fatalf("categorical label = %q, want %q", got, "beta")
	}
}

func TestAxesBarH_ConfiguresCategoricalYAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	bar, err := ax.BarH([]string{"north", "south"}, []float64{4, 7})
	if err != nil {
		t.Fatalf("BarH returned error: %v", err)
	}
	if bar == nil {
		t.Fatal("BarH returned nil bar")
	}
	if got := bar.X[1]; got != 1 {
		t.Fatalf("horizontal categorical bar position = %v, want 1", got)
	}
	if _, ok := ax.YAxis.Locator.(ticker.FixedLocator); !ok {
		t.Fatalf("y-axis locator = %T, want FixedLocator", ax.YAxis.Locator)
	}
}

func TestAxesBar_RejectedInputIsTransactional(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	originalXLocator := ax.XAxis.Locator
	originalXFormatter := ax.XAxis.Formatter
	originalYLocator := ax.YAxis.Locator
	originalYFormatter := ax.YAxis.Formatter
	originalCycle := ax.ColorCycle
	originalCycleIndex := originalCycle.Index()

	assertUnchanged := func(context string) {
		t.Helper()
		if ax.xUnits != nil || ax.yUnits != nil {
			t.Fatalf("%s configured units: x=%v y=%v", context, ax.xUnits, ax.yUnits)
		}
		if !reflect.DeepEqual(ax.XAxis.Locator, originalXLocator) ||
			!reflect.DeepEqual(ax.XAxis.Formatter, originalXFormatter) ||
			!reflect.DeepEqual(ax.YAxis.Locator, originalYLocator) ||
			!reflect.DeepEqual(ax.YAxis.Formatter, originalYFormatter) {
			t.Fatalf(
				"%s changed axis locator/formatter to x=%T/%T y=%T/%T",
				context,
				ax.XAxis.Locator,
				ax.XAxis.Formatter,
				ax.YAxis.Locator,
				ax.YAxis.Formatter,
			)
		}
		if len(ax.Artists) != 0 {
			t.Fatalf("%s added %d artists", context, len(ax.Artists))
		}
		if ax.ColorCycle != originalCycle || ax.ColorCycle.Index() != originalCycleIndex {
			t.Fatalf("%s replaced or advanced the property cycle", context)
		}
	}

	rejections := []struct {
		name string
		call func() (*Bar2D, error)
	}{
		{
			name: "vertical conversion",
			call: func() (*Bar2D, error) {
				return ax.Bar(
					[]string{"draft", "review"},
					[]struct{ value int }{{1}, {2}},
				)
			},
		},
		{
			name: "horizontal conversion",
			call: func() (*Bar2D, error) {
				return ax.BarH(
					[]string{"north", "south"},
					[]struct{ value int }{{4}, {7}},
				)
			},
		},
		{
			name: "mismatched shape",
			call: func() (*Bar2D, error) {
				return ax.Bar([]string{"draft", "review"}, []float64{1})
			},
		},
		{
			name: "per-bar option shape",
			call: func() (*Bar2D, error) {
				return ax.Bar(
					[]string{"draft", "review"},
					[]float64{1, 2},
					BarOptions{Widths: []float64{0.2, 0.3, 0.4}},
				)
			},
		},
		{
			name: "error-bar option shape",
			call: func() (*Bar2D, error) {
				return ax.Bar(
					[]string{"draft", "review"},
					[]float64{1, 2},
					BarOptions{YErr: []float64{0.1, 0.2, 0.3}},
				)
			},
		},
		{
			name: "invalid orientation",
			call: func() (*Bar2D, error) {
				orientation := BarOrientation(99)
				return ax.Bar(
					[]string{"draft", "review"},
					[]float64{1, 2},
					BarOptions{Orientation: &orientation},
				)
			},
		},
		{
			name: "multiple options",
			call: func() (*Bar2D, error) {
				return ax.Bar(
					[]string{"draft", "review"},
					[]float64{1, 2},
					BarOptions{},
					BarOptions{},
				)
			},
		},
	}
	for _, rejection := range rejections {
		bar, err := rejection.call()
		if err == nil || bar != nil {
			t.Fatalf("%s Bar() = (%v, %v), want nil artist and error", rejection.name, bar, err)
		}
		assertUnchanged(rejection.name)
	}
}

func TestAxesFillBetween_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	timestamps := []time.Time{
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 9, 0, 0, 0, 0, time.UTC),
	}

	fill, err := ax.FillBetween(timestamps, []float64{6, 7, 5}, []float64{10, 15, 13})
	if err != nil {
		t.Fatalf("FillBetween returned error: %v", err)
	}
	if fill == nil {
		t.Fatal("FillBetween returned nil fill")
	}
	if got := fill.X[1]; got != dates.Date2Num(timestamps[1]) {
		t.Fatalf("converted x[1] = %v, want %v", got, dates.Date2Num(timestamps[1]))
	}
	if _, ok := ax.XAxis.Locator.(dates.DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want dates.DateLocator", ax.XAxis.Locator)
	}
}

func TestAxesFillBetween_RejectedInputIsTransactional(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	originalXLocator := ax.XAxis.Locator
	originalXFormatter := ax.XAxis.Formatter
	originalCycle := ax.ColorCycle
	originalCycleIndex := originalCycle.Index()

	timestamps := []time.Time{
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC),
	}
	unsupported := []struct{ value int }{{1}, {2}}

	assertUnchanged := func(context string) {
		t.Helper()
		if ax.xUnits != nil || ax.yUnits != nil {
			t.Fatalf("%s configured units: x=%v y=%v", context, ax.xUnits, ax.yUnits)
		}
		if !reflect.DeepEqual(ax.XAxis.Locator, originalXLocator) ||
			!reflect.DeepEqual(ax.XAxis.Formatter, originalXFormatter) {
			t.Fatalf("%s changed x-axis locator/formatter to %T/%T", context, ax.XAxis.Locator, ax.XAxis.Formatter)
		}
		if len(ax.Artists) != 0 {
			t.Fatalf("%s added %d artists", context, len(ax.Artists))
		}
		if ax.ColorCycle != originalCycle || ax.ColorCycle.Index() != originalCycleIndex {
			t.Fatalf("%s replaced or advanced the property cycle", context)
		}
	}

	rejections := []struct {
		name string
		call func() (*Fill2D, error)
	}{
		{
			name: "x conversion",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(unsupported, []float64{6, 7}, []float64{10, 15})
			},
		},
		{
			name: "y1 conversion",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(timestamps, unsupported, []float64{10, 15})
			},
		},
		{
			// The x-axis date units configured before y2 fails must roll back.
			name: "y2 conversion",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(timestamps, []float64{6, 7}, unsupported)
			},
		},
		{
			name: "empty input",
			call: func() (*Fill2D, error) {
				return ax.FillBetween([]float64{}, []float64{}, []float64{})
			},
		},
		{
			name: "mismatched shape",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(timestamps, []float64{6, 7}, []float64{10})
			},
		},
		{
			name: "where shape",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(
					timestamps,
					[]float64{6, 7},
					[]float64{10, 15},
					FillOptions{Where: []bool{true, false, true}},
				)
			},
		},
		{
			name: "multiple options",
			call: func() (*Fill2D, error) {
				return ax.FillBetween(timestamps, []float64{6, 7}, []float64{10, 15}, FillOptions{}, FillOptions{})
			},
		},
	}
	for _, rejection := range rejections {
		fill, err := rejection.call()
		if err == nil || fill != nil {
			t.Fatalf("%s FillBetween() = (%v, %v), want nil artist and error", rejection.name, fill, err)
		}
		assertUnchanged(rejection.name)
	}
}

func TestAxesCategoryUnitsPreserveExplicitAxisInfoAfterRefresh(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	if _, err := ax.Bar([]string{"alpha", "beta"}, []float64{1, 2}); err != nil {
		t.Fatalf("Bar returned error: %v", err)
	}

	ax.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{10, 20}}
	ax.XAxis.Formatter = ticker.FormatStrFormatter{Pattern: "manual %.0f"}
	if _, err := ax.Bar([]string{"alpha", "beta", "gamma"}, []float64{1, 2, 3}); err != nil {
		t.Fatalf("second Bar returned error: %v", err)
	}

	loc, ok := ax.XAxis.Locator.(ticker.FixedLocator)
	if !ok {
		t.Fatalf("x-axis locator after category refresh = %T, want FixedLocator", ax.XAxis.Locator)
	}
	if fmt.Sprint(loc.TicksList) != "[10 20]" {
		t.Fatalf("x-axis locator ticks after category refresh = %v, want [10 20]", loc.TicksList)
	}
	if _, ok := ax.XAxis.Formatter.(ticker.FormatStrFormatter); !ok {
		t.Fatalf("x-axis formatter after category refresh = %T, want FormatStrFormatter", ax.XAxis.Formatter)
	}
	if got := ax.XAxis.Formatter.Format(10); got != "manual 10" {
		t.Fatalf("manual category formatter output = %q, want manual 10", got)
	}
}

type tripDistance float64

type tripDistanceConverter struct{}

func (tripDistanceConverter) Convert(value any) (float64, error) {
	v, ok := value.(tripDistance)
	if !ok {
		return 0, fmt.Errorf("unexpected value %T", value)
	}
	return float64(v), nil
}

func (tripDistanceConverter) AxisInfo([]float64) AxisInfo {
	return AxisInfo{
		Formatter: ticker.FormatStrFormatter{Pattern: "%.1f km"},
	}
}

func TestAxesPlot_UsesRegisteredConverter(t *testing.T) {
	if err := RegisterUnitConverter(tripDistance(0), func() UnitsConverter { return tripDistanceConverter{} }); err != nil {
		t.Fatalf("RegisterUnitConverter: %v", err)
	}

	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	line, err := ax.Plot([]tripDistance{1.5, 2.5, 4}, []float64{2, 4, 8})
	if err != nil {
		t.Fatalf("Plot returned error: %v", err)
	}
	if line == nil {
		t.Fatal("Plot returned nil line")
	}
	if got := line.XY[1].X; got != 2.5 {
		t.Fatalf("converted custom x[1] = %v, want 2.5", got)
	}
	if got := ax.XAxis.Formatter.Format(2.5); got != "2.5 km" {
		t.Fatalf("custom formatter output = %q, want %q", got, "2.5 km")
	}
}

func TestDateLocatorAndFormatter(t *testing.T) {
	loc := dates.DateLocator{Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 4)
	if len(ticks) < 3 {
		t.Fatalf("date tick count = %d, want at least 3", len(ticks))
	}

	formatter := dates.AutoDateFormatter{Min: minVal, Max: maxVal, Location: time.UTC}
	if got := formatter.Format(ticks[0]); got == "" {
		t.Fatal("formatted date tick should not be empty")
	}
}

func TestDateLocatorUsesDailyTicksForCompactDateRange(t *testing.T) {
	loc := dates.DateLocator{Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2023, time.December, 31, 13, 12, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.January, 10, 10, 48, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 5)
	if len(ticks) != 10 {
		t.Fatalf("date tick count = %d, want 10: %v", len(ticks), ticks)
	}
	for i, tick := range ticks {
		got := dates.Num2Date(tick, time.UTC)
		want := time.Date(2024, time.January, i+1, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("tick %d = %s, want %s", i, got, want)
		}
	}
}

func TestDayLocatorUsesRequestedMonthDays(t *testing.T) {
	loc := dates.DayLocator{ByMonthDay: []int{5, 12, 19}, Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 6)
	want := []time.Time{
		time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 19, 0, 0, 0, 0, time.UTC),
	}
	if len(ticks) != len(want) {
		t.Fatalf("tick count = %d, want %d: %v", len(ticks), len(want), ticks)
	}
	for i, tick := range ticks {
		got := dates.Num2Date(tick, time.UTC)
		if !got.Equal(want[i]) {
			t.Fatalf("tick %d = %s, want %s", i, got, want[i])
		}
	}
}

func TestYearLocatorUsesBaseMonthAndDay(t *testing.T) {
	loc := dates.YearLocator{Base: 2, Month: time.July, Day: 4, Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 4)
	want := []time.Time{
		time.Date(2020, time.July, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.July, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.July, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.July, 4, 0, 0, 0, 0, time.UTC),
	}
	assertDateTicks(t, ticks, want)
}

func TestMonthLocatorUsesRequestedMonths(t *testing.T) {
	loc := dates.MonthLocator{ByMonth: []time.Month{time.January, time.April, time.July, time.October}, ByMonthDay: 15, Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 6)
	want := []time.Time{
		time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.July, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.October, 15, 0, 0, 0, 0, time.UTC),
	}
	assertDateTicks(t, ticks, want)
}

func TestWeekdayLocatorUsesRequestedWeekdays(t *testing.T) {
	loc := dates.WeekdayLocator{ByWeekday: []time.Weekday{time.Monday, time.Wednesday}, Location: time.UTC}
	minVal := dates.Date2Num(time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.January, 10, 12, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 6)
	want := []time.Time{
		time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
	}
	assertDateTicks(t, ticks, want)
}

func TestClockLocatorsUseRequestedFields(t *testing.T) {
	minVal := dates.Date2Num(time.Date(2024, time.January, 1, 1, 15, 30, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.January, 1, 6, 45, 30, 0, time.UTC))

	hours := (dates.HourLocator{ByHour: []int{2, 4, 6}, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	assertDateTicks(t, hours, []time.Time{
		time.Date(2024, time.January, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 1, 4, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 1, 6, 0, 0, 0, time.UTC),
	})

	minutes := (dates.MinuteLocator{ByMinute: []int{20, 40}, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	if len(minutes) == 0 {
		t.Fatal("dates.MinuteLocator should produce ticks")
	}
	for _, tick := range minutes {
		minute := dates.Num2Date(tick, time.UTC).Minute()
		if minute != 20 && minute != 40 {
			t.Fatalf("minute tick = %d, want 20 or 40", minute)
		}
	}

	seconds := (dates.SecondLocator{BySecond: []int{0, 30}, Interval: 30, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	if len(seconds) == 0 {
		t.Fatal("dates.SecondLocator should produce ticks")
	}
	for _, tick := range seconds[:min(4, len(seconds))] {
		second := dates.Num2Date(tick, time.UTC).Second()
		if second != 0 && second != 30 {
			t.Fatalf("second tick = %d, want 0 or 30", second)
		}
	}
}

func TestMicrosecondLocatorUsesRequestedInterval(t *testing.T) {
	loc := dates.MicrosecondLocator{Interval: 500}
	minVal := dates.Date2Num(time.Unix(0, 100*time.Microsecond.Nanoseconds()).UTC())
	maxVal := dates.Date2Num(time.Unix(0, 1400*time.Microsecond.Nanoseconds()).UTC())

	ticks := loc.Ticks(minVal, maxVal, 6)
	assertDateTicks(t, ticks, []time.Time{
		time.Unix(0, 500*time.Microsecond.Nanoseconds()).UTC(),
		time.Unix(0, 1000*time.Microsecond.Nanoseconds()).UTC(),
	})
}

func TestDateLocatorUsesMicrosecondTicksForSubsecondRange(t *testing.T) {
	loc := dates.DateLocator{Location: time.UTC}
	minVal := dates.Date2Num(time.Unix(0, 0).UTC())
	maxVal := dates.Date2Num(time.Unix(0, 2500*time.Microsecond.Nanoseconds()).UTC())

	ticks := loc.Ticks(minVal, maxVal, 4)
	assertDateTicks(t, ticks, []time.Time{
		time.Unix(0, 0).UTC(),
		time.Unix(0, 500*time.Microsecond.Nanoseconds()).UTC(),
		time.Unix(0, 1000*time.Microsecond.Nanoseconds()).UTC(),
		time.Unix(0, 1500*time.Microsecond.Nanoseconds()).UTC(),
		time.Unix(0, 2000*time.Microsecond.Nanoseconds()).UTC(),
		time.Unix(0, 2500*time.Microsecond.Nanoseconds()).UTC(),
	})
}

func TestDateLocatorsUseNonUTCLocation(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	dayLoc := dates.DayLocator{Location: loc}
	minVal := dates.Date2Num(time.Date(2024, time.January, 1, 21, 30, 0, 0, time.UTC))
	maxVal := dates.Date2Num(time.Date(2024, time.January, 2, 23, 30, 0, 0, time.UTC))

	ticks := dayLoc.Ticks(minVal, maxVal, 4)
	assertDateTicks(t, ticks, []time.Time{
		time.Date(2024, time.January, 2, 0, 0, 0, 0, loc),
		time.Date(2024, time.January, 3, 0, 0, 0, 0, loc),
	})
}

func TestConciseDateFormatterUsesSharedTickLevel(t *testing.T) {
	formatter := dates.ConciseDateFormatter{Location: time.UTC}

	daily := []float64{
		dates.Date2Num(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)),
	}
	if got := labelsForTicks(formatter, daily); fmt.Sprint(got) != "[Jan 02 03]" {
		t.Fatalf("daily concise labels = %v, want [Jan 02 03]", got)
	}

	monthly := []float64{
		dates.Date2Num(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)),
	}
	if got := labelsForTicks(formatter, monthly); fmt.Sprint(got) != "[2024 Feb Mar]" {
		t.Fatalf("monthly concise labels = %v, want [2024 Feb Mar]", got)
	}

	subsecond := []float64{
		dates.Date2Num(time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 1, 12, 0, 0, int(500*time.Millisecond), time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 1, 12, 0, 1, 0, time.UTC)),
	}
	if got := labelsForTicks(formatter, subsecond); fmt.Sprint(got) != "[12:00 00.5 01]" {
		t.Fatalf("subsecond concise labels = %v, want [12:00 00.5 01]", got)
	}
}

func TestConciseDateFormatterOffsetTextUsesSharedDateContext(t *testing.T) {
	formatter := dates.ConciseDateFormatter{Location: time.UTC}
	ticks := []float64{
		dates.Date2Num(time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 2, 6, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC)),
		dates.Date2Num(time.Date(2024, time.January, 2, 18, 0, 0, 0, time.UTC)),
	}
	offsetter, ok := any(formatter).(ticker.OffsetFormatter)
	if !ok {
		t.Fatal("dates.ConciseDateFormatter should provide axis offset text")
	}
	if got := offsetter.OffsetText(ticks); got != "2024-Jan-02" {
		t.Fatalf("concise offset text = %q, want %q", got, "2024-Jan-02")
	}
}

func assertDateTicks(t *testing.T, ticks []float64, want []time.Time) {
	t.Helper()
	if len(ticks) != len(want) {
		t.Fatalf("tick count = %d, want %d: %v", len(ticks), len(want), ticks)
	}
	for i, tick := range ticks {
		got := dates.Num2Date(tick, want[i].Location())
		if !got.Equal(want[i]) {
			t.Fatalf("tick %d = %s, want %s", i, got, want[i])
		}
	}
}

func labelsForTicks(formatter ticker.Formatter, ticks []float64) []string {
	labels := make([]string, len(ticks))
	for i, tick := range ticks {
		labels[i] = ticker.FormatTick(formatter, tick, i, ticks)
	}
	return labels
}
