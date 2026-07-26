package core

import (
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
	if s, err := ax.Scatter([]float64{0, 1, 2}, []float64{0, 1}, ScatterOptions{}); err == nil || s != nil {
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
	if s, err := ax.Scatter([]float64{0, 1, 2}, []float64{0, 1, 2}, ScatterOptions{}); err != nil || s == nil {
		t.Fatalf("Scatter() = (%v, %v), want non-nil artist and nil error", s, err)
	}
	if len(*got) != 0 {
		t.Fatalf("valid Scatter should not warn, got %v", *got)
	}
}

// Every plotting entry point that rejects input must report a reason through
// the error channel and stay silent on the warning channel, which is reserved
// for artists that are accepted with a documented degradation.
func TestRejectedPlotInputReturnsErrorWithoutWarning(t *testing.T) {
	tests := []struct {
		name string
		call func(*Axes) (bool, error) // reports whether an artist came back
	}{
		{
			name: "hist mismatched weights",
			call: func(ax *Axes) (bool, error) {
				h, err := ax.Hist([]float64{1, 2, 3}, HistOptions{Weights: []float64{1, 1}})
				return h != nil, err
			},
		},
		{
			name: "hist empty data",
			call: func(ax *Axes) (bool, error) {
				h, err := ax.Hist(nil, HistOptions{})
				return h != nil, err
			},
		},
		{
			name: "fill between mismatched where",
			call: func(ax *Axes) (bool, error) {
				f, err := ax.FillBetweenPlot([]float64{0, 1, 2}, []float64{0, 0, 0}, []float64{1, 1, 1},
					FillOptions{Where: []bool{true, false}})
				return f != nil, err
			},
		},
		{
			name: "fill between x mismatched lengths",
			call: func(ax *Axes) (bool, error) {
				f, err := ax.FillBetweenX([]float64{0, 1, 2}, []float64{0, 0}, []float64{1, 1, 1}, FillOptions{})
				return f != nil, err
			},
		},
		{
			name: "errorbar negative errors",
			call: func(ax *Axes) (bool, error) {
				b, err := ax.ErrorBar([]float64{0, 1}, []float64{0, 1}, []float64{-1}, nil, ErrorBarOptions{})
				return b != nil, err
			},
		},
		{
			name: "errorbar invalid errorevery",
			call: func(ax *Axes) (bool, error) {
				b, err := ax.ErrorBar([]float64{0, 1}, []float64{0, 1}, nil, []float64{0.1},
					ErrorBarOptions{ErrorEvery: -1})
				return b != nil, err
			},
		},
		{
			name: "imshow rgb ragged rows",
			call: func(ax *Axes) (bool, error) {
				img, err := ax.ImShowRGB([][][]float64{
					{{0, 0, 0}, {1, 1, 1}},
					{{0, 0, 0}},
				}, ImShowRGBOptions{})
				return img != nil, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, restore := captureWarnings()
			defer restore()

			ax := newAlphaTestAxes()
			before := len(ax.Artists)
			added, err := tt.call(ax)
			if err == nil || added {
				t.Fatalf("call returned (artist=%v, %v), want no artist and an error", added, err)
			}
			if len(*warnings) != 0 {
				t.Fatalf("rejected input should return an error, not warn: %v", *warnings)
			}
			if len(ax.Artists) != before {
				t.Fatalf("artist count = %d, want %d — rejection must not mutate the axes", len(ax.Artists), before)
			}
		})
	}
}
