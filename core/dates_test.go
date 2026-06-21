package core

import (
	"math"
	"testing"
	"time"
)

func TestDate2NumMatchesMatplotlib(t *testing.T) {
	cases := []struct {
		t    time.Time
		want float64
	}{
		{time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), 18262.5},
		{time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC), 1.0},
		{time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 0.0},
	}
	for _, tc := range cases {
		if got := Date2Num(tc.t); math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("Date2Num(%v) = %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestNum2DateRoundTrip(t *testing.T) {
	orig := time.Date(2024, 6, 21, 13, 45, 30, 0, time.UTC)
	got := Num2Date(Date2Num(orig), time.UTC)
	if d := got.Sub(orig); d < -time.Millisecond || d > time.Millisecond {
		t.Fatalf("round trip drifted by %v: got %v want %v", d, got, orig)
	}
}

func TestDate2NumHandlesDatesBeyondDurationRange(t *testing.T) {
	// Dates more than ~292 years from the 1970 epoch overflow time.Duration, which
	// saturates at about ±106751.99 days. Verify such dates convert to the real
	// day count and round-trip back. (Default epoch: a conversion has occurred in
	// earlier tests, so the epoch stays 1970.)
	const saturationDays = 106751.99
	for _, orig := range []time.Time{
		time.Date(1600, 1, 1, 0, 0, 0, 0, time.UTC),   // ~370 years before epoch
		time.Date(2400, 7, 15, 6, 30, 0, 0, time.UTC), // ~430 years after epoch
	} {
		n := Date2Num(orig)
		if math.Abs(n) <= saturationDays {
			t.Fatalf("Date2Num(%v) = %v is within the duration-saturation band; overflow not avoided", orig, n)
		}
		got := Num2Date(n, time.UTC)
		if d := got.Sub(orig); d < -time.Second || d > time.Second {
			t.Fatalf("extreme-date round trip drifted by %v: got %v want %v", d, got, orig)
		}
	}
}

func TestSetEpochLocksAfterUse(t *testing.T) {
	// A conversion has already happened in other tests, so SetEpoch must fail.
	_ = Date2Num(time.Now().UTC())
	if err := SetEpoch(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("SetEpoch should fail after a conversion has occurred")
	}
}
