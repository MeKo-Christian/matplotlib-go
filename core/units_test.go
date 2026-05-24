package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/internal/geom"
)

func TestAxesPlotUnits_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	dates := []time.Time{
		time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 4, 0, 0, 0, 0, time.UTC),
	}

	line, err := ax.PlotUnits(dates, []float64{2, 3, 5})
	if err != nil {
		t.Fatalf("PlotUnits returned error: %v", err)
	}
	if line == nil {
		t.Fatal("PlotUnits returned nil line")
	}
	if got := line.XY[1].X; got != timeToDateNumber(dates[1]) {
		t.Fatalf("converted x[1] = %v, want %v", got, timeToDateNumber(dates[1]))
	}
	if _, ok := ax.XAxis.Locator.(DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want DateLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxis.Formatter.(AutoDateFormatter); !ok {
		t.Fatalf("x-axis formatter = %T, want AutoDateFormatter", ax.XAxis.Formatter)
	}

	ax.SetXLim(timeToDateNumber(dates[0]), timeToDateNumber(dates[len(dates)-1]))
	if _, ok := ax.XAxis.Formatter.(AutoDateFormatter); !ok {
		t.Fatalf("x-axis formatter after SetXLim = %T, want AutoDateFormatter", ax.XAxis.Formatter)
	}
}

func TestAxesPlotDate_ConfiguresDateAxis(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	dates := []time.Time{
		time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.March, 2, 0, 0, 0, 0, time.UTC),
	}

	line := ax.PlotDate(dates, []float64{1, 4})
	if line == nil {
		t.Fatal("PlotDate() returned nil")
	}
	if got := line.XY[0].X; got != timeToDateNumber(dates[0]) {
		t.Fatalf("PlotDate converted x[0] = %v, want %v", got, timeToDateNumber(dates[0]))
	}
	if _, ok := ax.XAxis.Locator.(DateLocator); !ok {
		t.Fatalf("x-axis locator = %T, want DateLocator", ax.XAxis.Locator)
	}
}

func TestAxesDateUnitsPreserveExplicitAxisInfoAfterAutoscale(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	dates := []time.Time{
		time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC),
	}
	if _, err := ax.PlotUnits(dates, []float64{1, 3, 2}); err != nil {
		t.Fatalf("PlotUnits returned error: %v", err)
	}

	ax.XAxis.Locator = DayLocator{ByMonthDay: []int{5, 12, 19}, Location: time.UTC}
	ax.XAxis.Formatter = DateFormatter{Layout: "02 Jan", Location: time.UTC}
	ax.AutoScale(0.06)

	if _, ok := ax.XAxis.Locator.(DayLocator); !ok {
		t.Fatalf("x-axis locator after AutoScale = %T, want DayLocator", ax.XAxis.Locator)
	}
	if _, ok := ax.XAxis.Formatter.(DateFormatter); !ok {
		t.Fatalf("x-axis formatter after AutoScale = %T, want DateFormatter", ax.XAxis.Formatter)
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

	loc, ok := ax.XAxis.Locator.(FixedLocator)
	if !ok {
		t.Fatalf("x-axis locator = %T, want FixedLocator", ax.XAxis.Locator)
	}
	if len(loc.TicksList) != 3 || loc.TicksList[2] != 2 {
		t.Fatalf("categorical ticks = %v, want [0 1 2]", loc.TicksList)
	}

	formatter, ok := ax.XAxis.Formatter.(FixedFormatter)
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
	if _, ok := ax.YAxis.Locator.(FixedLocator); !ok {
		t.Fatalf("y-axis locator = %T, want FixedLocator", ax.YAxis.Locator)
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
		Formatter: FormatStrFormatter{Pattern: "%.1f km"},
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
	loc := DateLocator{Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 4)
	if len(ticks) < 3 {
		t.Fatalf("date tick count = %d, want at least 3", len(ticks))
	}

	formatter := AutoDateFormatter{Min: minVal, Max: maxVal, Location: time.UTC}
	if got := formatter.Format(ticks[0]); got == "" {
		t.Fatal("formatted date tick should not be empty")
	}
}

func TestDateLocatorUsesDailyTicksForCompactDateRange(t *testing.T) {
	loc := DateLocator{Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2023, time.December, 31, 13, 12, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.January, 10, 10, 48, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 5)
	if len(ticks) != 10 {
		t.Fatalf("date tick count = %d, want 10: %v", len(ticks), ticks)
	}
	for i, tick := range ticks {
		got := dateNumberToTime(tick, time.UTC)
		want := time.Date(2024, time.January, i+1, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("tick %d = %s, want %s", i, got, want)
		}
	}
}

func TestDayLocatorUsesRequestedMonthDays(t *testing.T) {
	loc := DayLocator{ByMonthDay: []int{5, 12, 19}, Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.February, 20, 0, 0, 0, 0, time.UTC))

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
		got := dateNumberToTime(tick, time.UTC)
		if !got.Equal(want[i]) {
			t.Fatalf("tick %d = %s, want %s", i, got, want[i])
		}
	}
}

func TestYearLocatorUsesBaseMonthAndDay(t *testing.T) {
	loc := YearLocator{Base: 2, Month: time.July, Day: 4, Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC))

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
	loc := MonthLocator{ByMonth: []time.Month{time.January, time.April, time.July, time.October}, ByMonthDay: 15, Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC))

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
	loc := WeekdayLocator{ByWeekday: []time.Weekday{time.Monday, time.Wednesday}, Location: time.UTC}
	minVal := timeToDateNumber(time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.January, 10, 12, 0, 0, 0, time.UTC))

	ticks := loc.Ticks(minVal, maxVal, 6)
	want := []time.Time{
		time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC),
	}
	assertDateTicks(t, ticks, want)
}

func TestClockLocatorsUseRequestedFields(t *testing.T) {
	minVal := timeToDateNumber(time.Date(2024, time.January, 1, 1, 15, 30, 0, time.UTC))
	maxVal := timeToDateNumber(time.Date(2024, time.January, 1, 6, 45, 30, 0, time.UTC))

	hours := (HourLocator{ByHour: []int{2, 4, 6}, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	assertDateTicks(t, hours, []time.Time{
		time.Date(2024, time.January, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 1, 4, 0, 0, 0, time.UTC),
		time.Date(2024, time.January, 1, 6, 0, 0, 0, time.UTC),
	})

	minutes := (MinuteLocator{ByMinute: []int{20, 40}, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	if len(minutes) == 0 {
		t.Fatal("MinuteLocator should produce ticks")
	}
	for _, tick := range minutes {
		minute := dateNumberToTime(tick, time.UTC).Minute()
		if minute != 20 && minute != 40 {
			t.Fatalf("minute tick = %d, want 20 or 40", minute)
		}
	}

	seconds := (SecondLocator{BySecond: []int{0, 30}, Interval: 30, Location: time.UTC}).Ticks(minVal, maxVal, 6)
	if len(seconds) == 0 {
		t.Fatal("SecondLocator should produce ticks")
	}
	for _, tick := range seconds[:min(4, len(seconds))] {
		second := dateNumberToTime(tick, time.UTC).Second()
		if second != 0 && second != 30 {
			t.Fatalf("second tick = %d, want 0 or 30", second)
		}
	}
}

func assertDateTicks(t *testing.T, ticks []float64, want []time.Time) {
	t.Helper()
	if len(ticks) != len(want) {
		t.Fatalf("tick count = %d, want %d: %v", len(ticks), len(want), ticks)
	}
	for i, tick := range ticks {
		got := dateNumberToTime(tick, want[i].Location())
		if !got.Equal(want[i]) {
			t.Fatalf("tick %d = %s, want %s", i, got, want[i])
		}
	}
}
