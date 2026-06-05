package color

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToRGBAErrorSurfaceMatchesUpstreamFailureCategories(t *testing.T) {
	tests := []struct {
		name     string
		spec     any
		opts     []ToRGBAOption
		category colorErrorCategory
		source   string
	}{
		{
			name:     "explicit alpha out of range",
			spec:     "red",
			opts:     []ToRGBAOption{WithAlpha(2)},
			category: colorErrorAlphaRange,
			source:   "colors.py:_to_rgba_no_colorcycle: alpha range check",
		},
		{
			name:     "malformed hex length",
			spec:     "#12",
			category: colorErrorHexFormat,
			source:   "colors.py:_to_rgba_no_colorcycle: invalid RGBA argument for malformed hex",
		},
		{
			name:     "malformed hex alpha suffix",
			spec:     "#1122334x",
			category: colorErrorHexFormat,
			source:   "colors.py:_to_rgba_no_colorcycle: invalid RGBA argument for malformed hex",
		},
		{
			name:     "grayscale string out of range",
			spec:     "5",
			category: colorErrorStringGrayscaleRange,
			source:   "colors.py:_to_rgba_no_colorcycle: string gray range check",
		},
		{
			name:     "unknown string color",
			spec:     "unknown_color",
			category: colorErrorInvalidArgument,
			source:   "tests/test_colors.py:test_failed_conversions",
		},
		{
			name:     "nil color",
			spec:     nil,
			category: colorErrorInvalidArgument,
			source:   "colors.py:_to_rgba_no_colorcycle: non-iterable values are invalid",
		},
		{
			name:     "numeric scalar is not grayscale",
			spec:     0.4,
			category: colorErrorInvalidArgument,
			source:   "tests/test_colors.py:test_failed_conversions",
		},
		{
			name:     "short sequence",
			spec:     []float64{0.2, 0.3},
			category: colorErrorSequenceLength,
			source:   "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		},
		{
			name:     "long sequence",
			spec:     []float64{0.2, 0.3, 0.4, 0.5, 0.6},
			category: colorErrorSequenceLength,
			source:   "colors.py:_to_rgba_no_colorcycle: RGB/RGBA sequence length",
		},
		{
			name:     "non-real sequence component",
			spec:     []any{0.2, "0.3", 0.4},
			category: colorErrorInvalidArgument,
			source:   "colors.py:_to_rgba_no_colorcycle: RGB/RGBA values must be real",
		},
		{
			name:     "channel below range",
			spec:     []float64{-0.1, 0, 0},
			category: colorErrorValueRange,
			source:   "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		},
		{
			name:     "channel above range",
			spec:     []float64{1.2, 0, 0},
			category: colorErrorValueRange,
			source:   "colors.py:_to_rgba_no_colorcycle: RGBA values must be within range",
		},
	}

	for _, tc := range tests {
		if tc.source == "" {
			t.Fatalf("%s: missing upstream source", tc.name)
		}
		got, err := ToRGBA(tc.spec, tc.opts...)
		if err == nil {
			t.Fatalf("%s: ToRGBA(%v) = %+v, want %s error", tc.name, tc.spec, got, tc.category)
		}
		if category := classifyColorError(err); category != tc.category {
			t.Fatalf("%s: ToRGBA(%v) error category = %s from %q, want %s",
				tc.name, tc.spec, category, err, tc.category)
		}
	}
}

func TestColorConversionMigrationNotesDocumentUnsupportedInputs(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	required := []string{
		"Phase 17.75.5 Color Conversion",
		"failure category rather than exact Python",
		"color-alpha tuples",
		"to_rgba_array",
		"NumPy masked values",
		"NaN RGB component",
		"explicit typed Go alpha",
	}
	for _, phrase := range required {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("migration notes missing %q", phrase)
		}
	}
}

type colorErrorCategory string

const (
	colorErrorAlphaRange           colorErrorCategory = "alpha-range"
	colorErrorHexFormat            colorErrorCategory = "hex-format"
	colorErrorStringGrayscaleRange colorErrorCategory = "string-grayscale-range"
	colorErrorInvalidArgument      colorErrorCategory = "invalid-rgba-argument"
	colorErrorSequenceLength       colorErrorCategory = "sequence-length"
	colorErrorValueRange           colorErrorCategory = "rgba-value-range"
	colorErrorUnknown              colorErrorCategory = "unknown"
)

func classifyColorError(err error) colorErrorCategory {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "alpha must be between"):
		return colorErrorAlphaRange
	case strings.Contains(msg, "invalid hex color specifier"):
		return colorErrorHexFormat
	case strings.Contains(msg, "invalid string grayscale value"):
		return colorErrorStringGrayscaleRange
	case strings.Contains(msg, "RGBA sequence should have length"):
		return colorErrorSequenceLength
	case strings.Contains(msg, "RGBA values should be within"):
		return colorErrorValueRange
	case strings.Contains(msg, "invalid RGBA argument"):
		return colorErrorInvalidArgument
	default:
		return colorErrorUnknown
	}
}
