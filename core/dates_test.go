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

func TestSetEpochLocksAfterUse(t *testing.T) {
	// A conversion has already happened in other tests, so SetEpoch must fail.
	_ = Date2Num(time.Now().UTC())
	if err := SetEpoch(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("SetEpoch should fail after a conversion has occurred")
	}
}
