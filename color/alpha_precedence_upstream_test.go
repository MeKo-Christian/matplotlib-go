package color

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

type matplotlibAlphaPrecedenceCase struct {
	name        string
	source      string
	spec        any
	alpha       *float64
	want        render.Color
	goSupported bool
}

func TestMatplotlibAlphaPrecedenceUpstreamCaseInventory(t *testing.T) {
	seen := map[string]bool{}
	for _, tc := range matplotlibAlphaPrecedenceCases {
		if tc.name == "" {
			t.Fatalf("alpha-precedence case has empty name")
		}
		if tc.source == "" {
			t.Fatalf("%s: missing upstream source reference", tc.name)
		}
		if seen[tc.name] {
			t.Fatalf("duplicate alpha-precedence case %q", tc.name)
		}
		seen[tc.name] = true
	}

	required := []string{
		"colors.py:to_rgba: explicit alpha forces the returned alpha",
		"colors.py:_to_rgba_no_colorcycle: none ignores explicit alpha",
		"colors.py:_to_rgba_no_colorcycle: invalid explicit alpha errors",
		"tests/test_colors.py:test_to_rgba_accepts_color_alpha_tuple",
		"tests/test_colors.py:test_to_rgba_explicit_alpha_overrides_tuple_alpha",
		"tests/test_colors.py:test_to_rgba_error_with_color_invalid_alpha_tuple",
		"tests/test_colors.py:test_to_rgba_array_alpha_array",
	}
	for _, source := range required {
		found := false
		for _, tc := range matplotlibAlphaPrecedenceCases {
			if tc.source == source {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing alpha-precedence upstream source %q", source)
		}
	}
}

func alphaPtr(alpha float64) *float64 {
	return &alpha
}

var matplotlibAlphaPrecedenceCases = []matplotlibAlphaPrecedenceCase{
	{
		name:        "explicit alpha overrides opaque named color",
		source:      "colors.py:to_rgba: explicit alpha forces the returned alpha",
		spec:        "red",
		alpha:       alphaPtr(0.5),
		want:        render.Color{R: 1, G: 0, B: 0, A: 0.5},
		goSupported: true,
	},
	{
		name:        "explicit alpha overrides embedded hex alpha",
		source:      "colors.py:to_rgba: explicit alpha forces the returned alpha",
		spec:        "#ffffff00",
		alpha:       alphaPtr(0.5),
		want:        render.Color{R: 1, G: 1, B: 1, A: 0.5},
		goSupported: true,
	},
	{
		name:        "embedded hex alpha is used without explicit alpha",
		source:      "tests/test_colors.py:test_to_rgba_accepts_color_alpha_tuple",
		spec:        "#ffffff00",
		alpha:       nil,
		want:        render.Color{R: 1, G: 1, B: 1, A: 0},
		goSupported: true,
	},
	{
		name:        "rgb sequence defaults alpha to opaque",
		source:      "colors.py:to_rgba: alpha defaults to 1 without alpha channel",
		spec:        []float64{1, 1, 1},
		alpha:       nil,
		want:        render.Color{R: 1, G: 1, B: 1, A: 1},
		goSupported: true,
	},
	{
		name:        "explicit alpha overrides rgb sequence default alpha",
		source:      "tests/test_colors.py:test_conversions",
		spec:        []float64{1, 1, 1},
		alpha:       alphaPtr(0.5),
		want:        render.Color{R: 1, G: 1, B: 1, A: 0.5},
		goSupported: true,
	},
	{
		name:        "grayscale string accepts explicit alpha",
		source:      "tests/test_colors.py:test_conversions",
		spec:        ".1",
		alpha:       alphaPtr(0.5),
		want:        render.Color{R: 0.1, G: 0.1, B: 0.1, A: 0.5},
		goSupported: true,
	},
	{
		name:        "none ignores explicit alpha",
		source:      "colors.py:_to_rgba_no_colorcycle: none ignores explicit alpha",
		spec:        "none",
		alpha:       alphaPtr(0.5),
		want:        render.Color{A: 0},
		goSupported: true,
	},
	{
		name:        "invalid explicit alpha errors",
		source:      "colors.py:_to_rgba_no_colorcycle: invalid explicit alpha errors",
		spec:        "blue",
		alpha:       alphaPtr(2),
		goSupported: true,
	},
	{
		name:        "python color-alpha tuple supplies alpha",
		source:      "tests/test_colors.py:test_to_rgba_accepts_color_alpha_tuple",
		spec:        []any{"white", 0.5},
		alpha:       nil,
		want:        render.Color{R: 1, G: 1, B: 1, A: 0.5},
		goSupported: false,
	},
	{
		name:        "explicit alpha overrides python tuple alpha",
		source:      "tests/test_colors.py:test_to_rgba_explicit_alpha_overrides_tuple_alpha",
		spec:        []any{"red", 0.1},
		alpha:       alphaPtr(0.9),
		want:        render.Color{R: 1, G: 0, B: 0, A: 0.9},
		goSupported: false,
	},
	{
		name:        "invalid python tuple alpha errors",
		source:      "tests/test_colors.py:test_to_rgba_error_with_color_invalid_alpha_tuple",
		spec:        []any{"blue", 2.0},
		alpha:       nil,
		goSupported: false,
	},
	{
		name:        "to_rgba_array alpha sequence overrides per-color alpha",
		source:      "tests/test_colors.py:test_to_rgba_array_alpha_array",
		spec:        []string{"r", "g"},
		alpha:       nil,
		goSupported: false,
	},
}
