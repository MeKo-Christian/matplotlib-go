package color

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

type matplotlibGrayscaleSequenceCase struct {
	name        string
	source      string
	spec        any
	want        render.Color
	wantErr     bool
	goSupported bool
}

func TestMatplotlibGrayscaleSequenceAmbiguityInventory(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range matplotlibGrayscaleSequenceCases {
		if tc.name == "" {
			t.Fatalf("grayscale/sequence case has empty name")
		}
		if tc.source == "" {
			t.Fatalf("%s: missing upstream source reference", tc.name)
		}
		if seen[tc.name] {
			t.Fatalf("duplicate grayscale/sequence case %q", tc.name)
		}
		seen[tc.name] = true
	}

	required := []string{
		"colors.py:_to_rgba_no_colorcycle: string gray",
		"colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		"colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		"colors.py:_to_rgba_no_colorcycle: RGB/RGBA values must be real",
		"tests/test_colors.py:test_conversions: grayscale list belongs to to_rgba_array",
		"tests/test_colors.py:test_failed_conversions",
	}
	for _, source := range required {
		found := false
		for _, tc := range matplotlibGrayscaleSequenceCases {
			if tc.source == source {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing grayscale/sequence upstream source %q", source)
		}
	}
}

var matplotlibGrayscaleSequenceCases = []matplotlibGrayscaleSequenceCase{
	{
		name:        "string grayscale lower bound",
		source:      "colors.py:_to_rgba_no_colorcycle: string gray",
		spec:        "0",
		want:        render.Color{R: 0, G: 0, B: 0, A: 1},
		goSupported: true,
	},
	{
		name:        "string grayscale upper bound",
		source:      "colors.py:_to_rgba_no_colorcycle: string gray",
		spec:        "1",
		want:        render.Color{R: 1, G: 1, B: 1, A: 1},
		goSupported: true,
	},
	{
		name:        "string grayscale fractional value",
		source:      "colors.py:_to_rgba_no_colorcycle: string gray",
		spec:        ".5",
		want:        render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		goSupported: true,
	},
	{
		name:        "string grayscale above range errors",
		source:      "tests/test_colors.py:test_failed_conversions",
		spec:        "5",
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "string grayscale below range errors",
		source:      "tests/test_colors.py:test_failed_conversions",
		spec:        "-1",
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "string grayscale nan errors",
		source:      "tests/test_colors.py:test_failed_conversions",
		spec:        "nan",
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "numeric scalar is not grayscale",
		source:      "colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		spec:        0.4,
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "rgb numeric sequence defaults alpha",
		source:      "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		spec:        []float64{0.2, 0.3, 0.4},
		want:        render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1},
		goSupported: true,
	},
	{
		name:        "rgba numeric sequence keeps alpha",
		source:      "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		spec:        []float64{0.2, 0.3, 0.4, 0.5},
		want:        render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5},
		goSupported: true,
	},
	{
		name:        "short numeric sequence errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		spec:        []float64{0.2, 0.3},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "long numeric sequence errors",
		source:      "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		spec:        []float64{0.2, 0.3, 0.4, 0.5, 0.6},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "numeric sequence values must be real",
		source:      "colors.py:_to_rgba_no_colorcycle: RGB/RGBA values must be real",
		spec:        []any{0.2, "0.3", 0.4},
		wantErr:     true,
		goSupported: true,
	},
	{
		name:        "list of grayscale strings is array conversion input",
		source:      "tests/test_colors.py:test_conversions: grayscale list belongs to to_rgba_array",
		spec:        []string{".2", ".5", ".8"},
		wantErr:     true,
		goSupported: false,
	},
}
