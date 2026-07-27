package dates

import (
	"testing"
	"time"

	"github.com/cwbudde/matplotlib-go/style"
)

// date.* rcParam consumption. date.epoch resolution is covered by
// TestParseDateEpochString only: the epoch is process-global and locks on the
// first conversion (matching matplotlib set_epoch/get_epoch), so an
// end-to-end rc test would poison every other date test in the package.

func TestParseDateEpochString(t *testing.T) {
	cases := []struct {
		value string
		want  time.Time
	}{
		{"1970-01-01T00:00:00", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2000-01-01T12:30:00", time.Date(2000, 1, 1, 12, 30, 0, 0, time.UTC)},
		{"1900-01-01", time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"0000-12-31T00:00:00", time.Date(0, 12, 31, 0, 0, 0, 0, time.UTC)},
		// Invalid values fall back to the matplotlib default epoch.
		{"not-a-date", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		if got := parseDateEpochString(tc.value); !got.Equal(tc.want) {
			t.Errorf("parseDateEpochString(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestDateLocatorIntervalMultiples(t *testing.T) {
	// A 25-day span selects a 4-day step with interval multiples (day table
	// {1,2,4,7,14}, ticks snapped to day 1,5,9,...) and a 3-day step without
	// (matplotlib's default table {1,2,3,7,14,21}, stepping from the first
	// day boundary).
	minVal := Date2Num(time.Date(2024, 3, 3, 12, 0, 0, 0, time.UTC))
	maxVal := Date2Num(time.Date(2024, 3, 28, 12, 0, 0, 0, time.UTC))

	daysOf := func(ticks []float64) []int {
		days := make([]int, 0, len(ticks))
		for _, v := range ticks {
			days = append(days, Num2Date(v, time.UTC).Day())
		}
		return days
	}

	defaultTicks := DateLocator{}.Ticks(minVal, maxVal, 6)
	got := daysOf(defaultTicks)
	want := []int{5, 9, 13, 17, 21, 25} // 4-day step on 1+4k days-of-month
	if len(got) != len(want) {
		t.Fatalf("interval-multiples ticks on days %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("interval-multiples ticks on days %v, want %v", got, want)
		}
	}

	noMultiples := false
	looseTicks := DateLocator{IntervalMultiples: &noMultiples}.Ticks(minVal, maxVal, 6)
	got = daysOf(looseTicks)
	// 3-day step from the truncated range start (Mar 3 00:00); the first
	// occurrence at or after vmin (Mar 3 12:00) is Mar 6, like rrule.
	want = []int{6, 9, 12, 15, 18, 21, 24, 27}
	if len(got) != len(want) {
		t.Fatalf("no-interval-multiples ticks on days %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("no-interval-multiples ticks on days %v, want %v", got, want)
		}
	}
}

func TestDateIntervalMultiplesRCReachesDateLocator(t *testing.T) {
	t.Cleanup(style.ResetDefaults)
	if _, err := style.UpdateParams(style.Params{"date.interval_multiples": "False"}); err != nil {
		t.Fatalf("UpdateParams: %v", err)
	}

	minVal := Date2Num(time.Date(2024, 3, 3, 12, 0, 0, 0, time.UTC))
	maxVal := Date2Num(time.Date(2024, 3, 28, 12, 0, 0, 0, time.UTC))
	ticks := DateLocator{}.Ticks(minVal, maxVal, 6)
	if len(ticks) == 0 {
		t.Fatal("no ticks")
	}
	// The rc value must select the 3-day step (8 ticks), not the snapped
	// 4-day step (6 ticks).
	if len(ticks) != 8 {
		t.Fatalf("got %d ticks, want 8 (3-day step from date.interval_multiples: False)", len(ticks))
	}
}

func TestDateAutoformatterRCReachesAutoDateFormatter(t *testing.T) {
	t.Cleanup(style.ResetDefaults)

	minVal := Date2Num(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC))
	maxVal := Date2Num(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	x := Date2Num(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	f := AutoDateFormatter{Min: minVal, Max: maxVal}

	if got := f.Format(x); got != "2020" {
		t.Fatalf("default year-scale label = %q, want %q", got, "2020")
	}

	if _, err := style.UpdateParams(style.Params{"date.autoformatter.year": "'%y"}); err != nil {
		t.Fatalf("UpdateParams: %v", err)
	}
	if got := f.Format(x); got != "'20" {
		t.Fatalf("rc-driven year-scale label = %q, want %q from date.autoformatter.year", got, "'20")
	}
}
