package color

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

type matplotlibBadValueCase struct {
	name        string
	source      string
	spec        any
	want        render.Color
	wantErr     bool
	goSupported bool
}

func TestMatplotlibBadValueSemanticsInventory(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range matplotlibBadValueCases {
		if tc.name == "" {
			t.Fatalf("bad-value case has empty name")
		}
		if tc.source == "" {
			t.Fatalf("%s: missing upstream source reference", tc.name)
		}
		if seen[tc.name] {
			t.Fatalf("duplicate bad-value case %q", tc.name)
		}
		seen[tc.name] = true
	}

	required := []string{
		"colors.py:_to_rgba_no_colorcycle: masked scalar is transparent",
		"colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		"colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		"colors.py:_to_rgba_no_colorcycle: NaN sequence components pass through",
		"tests/test_colors.py:test_conversions_masked",
		"tests/test_colors.py:test_failed_conversions",
	}
	for _, source := range required {
		found := false
		for _, tc := range matplotlibBadValueCases {
			if tc.source == source {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing bad-value upstream source %q", source)
		}
	}
}

var matplotlibBadValueCases = []matplotlibBadValueCase{
	{
		name:        "nil equivalent is invalid",
		source:      "colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		spec:        nil,
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "scalar nan is invalid",
		source:      "tests/test_colors.py:test_failed_conversions",
		spec:        math.NaN(),
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "scalar positive infinity is invalid",
		source:      "colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		spec:        math.Inf(1),
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "scalar negative infinity is invalid",
		source:      "colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		spec:        math.Inf(-1),
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "negative rgb component errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		spec:        []float64{-0.1, 0, 0},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "rgb component above one errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		spec:        []float64{1.2, 0, 0},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "positive infinity rgb component errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		spec:        []float64{math.Inf(1), 0, 0},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "negative infinity rgb component errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		spec:        []float64{math.Inf(-1), 0, 0},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "nan rgb component passes through upstream",
		source:      "colors.py:_to_rgba_no_colorcycle: NaN sequence components pass through",
		spec:        []float64{math.NaN(), 0, 0},
		want:        render.Color{R: math.NaN(), G: 0, B: 0, A: 1},
		goSupported: false,
	},
	{
		name:        "masked scalar maps to transparent upstream",
		source:      "colors.py:_to_rgba_no_colorcycle: masked scalar is transparent",
		spec:        "numpy.ma.masked",
		want:        render.Color{A: 0},
		goSupported: false,
	},
	{
		name:        "masked string color maps to transparent upstream",
		source:      "tests/test_colors.py:test_conversions_masked",
		spec:        "masked array element",
		want:        render.Color{A: 0},
		goSupported: false,
	},
}
