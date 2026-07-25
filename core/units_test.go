package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/dates"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestAxesPlotUnits_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	timestamps := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 4, 0, 0, 0, 0, time.UTC),
	}

	line, err := ax.PlotUnits(timestamps, []float64{2, 3, 5})
	if err != nil {
		t.Fatalf("PlotUnits returned error: %v", err)
	}
	if line == nil {
		t.Fatal("PlotUnits returned nil line")
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

func TestAxesPlotDate_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	timestamps := []time.Time{
		time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 2, 0, 0, 0, 0, time.UTC),
	}

	line := ax.PlotDate(timestamps, []float64{1, 4})
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
	if _, err := ax.PlotUnits(timestamps, []float64{1, 3, 2}); err != nil {
		t.Fatalf("PlotUnits returned error: %v", err)
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

func TestAxesBarUnits_ConfiguresCategoricalXAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	bar, err := ax.BarUnits([]string{"alpha", "beta", "gamma"}, []float64{1, 3, 2})
	if err != nil {
		t.Fatalf("BarUnits returned error: %v", err)
	}
	if bar == nil {
		t.Fatal("BarUnits returned nil bar")
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

func TestAxesBarUnits_HorizontalConfiguresCategoricalYAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	orientation := BarHorizontal

	bar, err := ax.BarUnits([]string{"north", "south"}, []float64{4, 7}, BarOptions{
		Orientation: &orientation,
	})
	if err != nil {
		t.Fatalf("BarUnits returned error: %v", err)
	}
	if bar == nil {
		t.Fatal("BarUnits returned nil bar")
	}
	if got := bar.X[1]; got != 1 {
		t.Fatalf("horizontal categorical bar position = %v, want 1", got)
	}
	if _, ok := ax.YAxis.Locator.(ticker.FixedLocator); !ok {
		t.Fatalf("y-axis locator = %T, want FixedLocator", ax.YAxis.Locator)
	}
}

func TestAxesCategoryUnitsPreserveExplicitAxisInfoAfterRefresh(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	if _, err := ax.BarUnits([]string{"alpha", "beta"}, []float64{1, 2}); err != nil {
		t.Fatalf("BarUnits returned error: %v", err)
	}

	ax.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{10, 20}}
	ax.XAxis.Formatter = ticker.FormatStrFormatter{Pattern: "manual %.0f"}
	if _, err := ax.BarUnits([]string{"alpha", "beta", "gamma"}, []float64{1, 2, 3}); err != nil {
		t.Fatalf("second BarUnits returned error: %v", err)
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

func TestAxesPlotUnits_UsesRegisteredConverter(t *testing.T) {
	if err := RegisterUnitConverter(tripDistance(0), func() UnitsConverter { return tripDistanceConverter{} }); err != nil {
		t.Fatalf("RegisterUnitConverter: %v", err)
	}

	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	line, err := ax.PlotUnits([]tripDistance{1.5, 2.5, 4}, []float64{2, 4, 8})
	if err != nil {
		t.Fatalf("PlotUnits returned error: %v", err)
	}
	if line == nil {
		t.Fatal("PlotUnits returned nil line")
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
