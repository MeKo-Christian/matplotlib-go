package color

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestNamedColorCatalogSizesMatchMatplotlib(t *testing.T) {
	if got, want := len(BaseColors()), 8; got != want {
		t.Fatalf("BaseColors length = %d, want %d", got, want)
	}
	if got, want := len(TableauColors()), 10; got != want {
		t.Fatalf("TableauColors length = %d, want %d", got, want)
	}
	if got, want := len(CSS4Colors()), 148; got != want {
		t.Fatalf("CSS4Colors length = %d, want %d", got, want)
	}
	if got, want := len(XKCDColors()), 949; got != want {
		t.Fatalf("XKCDColors length = %d, want %d", got, want)
	}
}

func TestToRGBAResolvesMatplotlibNamedColors(t *testing.T) {
	tests := []struct {
		name string
		want render.Color
	}{
		{"b", render.Color{R: 0, G: 0, B: 1, A: 1}},
		{"rebeccapurple", render.Color{R: 0x66 / 255.0, G: 0x33 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"tab:orange", Tab10[1]},
		{"tab:grey", Tab10[7]},
		{"xkcd:cloudy blue", render.Color{R: 0xac / 255.0, G: 0xc2 / 255.0, B: 0xd9 / 255.0, A: 1}},
		{"xkcd:warm gray", render.Color{R: 0x97 / 255.0, G: 0x8a / 255.0, B: 0x84 / 255.0, A: 1}},
	}
	for _, tc := range tests {
		got, err := ToRGBA(tc.name)
		if err != nil {
			t.Fatalf("ToRGBA(%q) error = %v", tc.name, err)
		}
		if !sameColor(got, tc.want) {
			t.Fatalf("ToRGBA(%q) = %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestToRGBAParsesHexGrayCycleAndTuples(t *testing.T) {
	palette := Palette{
		{R: 0.1, G: 0.2, B: 0.3, A: 1},
		{R: 0.4, G: 0.5, B: 0.6, A: 0.7},
	}
	tests := []struct {
		name string
		spec any
		opts []ToRGBAOption
		want render.Color
	}{
		{"long hex", "#336699", nil, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"short hex alpha", "#369c", nil, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 0xcc / 255.0}},
		{"bare hex opt in", "336699", []ToRGBAOption{WithBareHex()}, render.Color{R: 0x33 / 255.0, G: 0x66 / 255.0, B: 0x99 / 255.0, A: 1}},
		{"gray string", "0.25", nil, render.Color{R: 0.25, G: 0.25, B: 0.25, A: 1}},
		{"cycle", "C3", []ToRGBAOption{WithColorCycle(palette)}, palette[1]},
		{"tuple string", "(0.2, 0.3, 0.4, 0.5)", nil, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5}},
		{"float slice", []float64{0.2, 0.3, 0.4}, []ToRGBAOption{WithAlpha(0.6)}, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.6}},
		{"float array", [4]float64{0.2, 0.3, 0.4, 0.5}, nil, render.Color{R: 0.2, G: 0.3, B: 0.4, A: 0.5}},
	}
	for _, tc := range tests {
		got, err := ToRGBA(tc.spec, tc.opts...)
		if err != nil {
			t.Fatalf("%s: ToRGBA(%v) error = %v", tc.name, tc.spec, err)
		}
		if !sameColor(got, tc.want) {
			t.Fatalf("%s: ToRGBA(%v) = %+v, want %+v", tc.name, tc.spec, got, tc.want)
		}
	}
}

func TestToRGBARejectsInvalidSpecs(t *testing.T) {
	for _, spec := range []any{"B", "#12", "1.5", []float64{1, 0, 0, 0, 1}, []float64{1.2, 0, 0}} {
		if got, err := ToRGBA(spec); err == nil {
			t.Fatalf("ToRGBA(%v) = %+v, want error", spec, got)
		}
	}
	if got, err := ToRGBA("red", WithAlpha(2)); err == nil {
		t.Fatalf("ToRGBA(red, WithAlpha(2)) = %+v, want error", got)
	}
}

func sameColor(a, b render.Color) bool {
	const eps = 1e-12
	return math.Abs(a.R-b.R) < eps &&
		math.Abs(a.G-b.G) < eps &&
		math.Abs(a.B-b.B) < eps &&
		math.Abs(a.A-b.A) < eps
}
