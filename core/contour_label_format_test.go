package core

import (
	"math"
	"testing"
)

func TestContourLabelFormatString(t *testing.T) {
	f := newContourLabelFormatter(ClabelOptions{FormatString: "%1.2f"}, nil)
	if got := f.Format(1.5); got != "1.50" {
		t.Fatalf("format string label = %q, want %q", got, "1.50")
	}
}

func TestContourLabelFormatDict(t *testing.T) {
	f := newContourLabelFormatter(ClabelOptions{FormatDict: map[float64]string{1: "low", 2: "high"}}, nil)
	if got := f.Format(1); got != "low" {
		t.Fatalf("dict label = %q, want low", got)
	}
	if got := f.Format(2); got != "high" {
		t.Fatalf("dict label = %q, want high", got)
	}
	// Missing key falls back to %1.3f.
	if got := f.Format(3); got != "3.000" {
		t.Fatalf("dict fallback label = %q, want 3.000", got)
	}
}

func TestContourLabelFormatterCallable(t *testing.T) {
	f := newContourLabelFormatter(ClabelOptions{Formatter: FuncFormatter(func(x float64) string {
		return "v=" + ScalarFormatter{Prec: 0}.Format(x)
	})}, nil)
	if got := f.Format(4); got != "v=4" {
		t.Fatalf("callable label = %q, want v=4", got)
	}
}

func TestApplyRightSideUp(t *testing.T) {
	upward := math.Pi // 180°, would render upside-down
	if got := applyRightSideUp(upward, true); math.Abs(got) > math.Pi/2+1e-9 {
		t.Fatalf("rightSideUp angle = %v, want within [-pi/2, pi/2]", got)
	}
	if got := applyRightSideUp(upward, false); math.Abs(got-upward) > 1e-9 {
		t.Fatalf("raw angle = %v, want unchanged %v", got, upward)
	}
}
