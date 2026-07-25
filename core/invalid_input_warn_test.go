package core

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// captureWarnings installs a diag handler that records every warning and
// returns both the slice and a restore func to defer.
func captureWarnings() (*[]string, func()) {
	var got []string
	restore := diag.SetHandler(func(m string) { got = append(got, m) })
	return &got, restore
}

// Matplotlib raises "x and y must be the same size" when scatter inputs differ.
// Rejected Go plot input returns an error rather than warning and skipping.
func TestScatterLengthMismatchReturnsError(t *testing.T) {
	got, restore := captureWarnings()
	defer restore()

	ax := newAlphaTestAxes()
	if s, err := ax.Scatter([]float64{0, 1, 2}, []float64{0, 1}); err == nil || s != nil {
		t.Fatalf("Scatter() = (%v, %v), want nil artist and length error", s, err)
	}
	if len(*got) != 0 {
		t.Fatalf("rejected Scatter should return an error, not warn: %v", *got)
	}
}

func TestScatterValidInputDoesNotWarn(t *testing.T) {
	got, restore := captureWarnings()
	defer restore()

	ax := newAlphaTestAxes()
	if s, err := ax.Scatter([]float64{0, 1, 2}, []float64{0, 1, 2}); err != nil || s == nil {
		t.Fatalf("Scatter() = (%v, %v), want non-nil artist and nil error", s, err)
	}
	if len(*got) != 0 {
		t.Fatalf("valid Scatter should not warn, got %v", *got)
	}
}

// Hist with a weights slice whose length disagrees with the data is invalid
// input in Matplotlib; surface a reason instead of silently dropping the artist.
func TestHistWeightsMismatchWarns(t *testing.T) {
	got, restore := captureWarnings()
	defer restore()

	ax := newAlphaTestAxes()
	weights := []float64{1, 1}
	if h := ax.Hist([]float64{1, 2, 3}, HistOptions{Weights: weights}); h != nil {
		t.Fatal("expected nil hist for mismatched weights")
	}
	if len(*got) == 0 {
		t.Fatal("Hist weights mismatch produced no diagnostic")
	}
	if !strings.Contains(strings.ToLower((*got)[0]), "hist") {
		t.Fatalf("warning %q should name the Hist call", (*got)[0])
	}
}
