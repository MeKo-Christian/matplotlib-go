package color

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestGetColormap_UnknownFallsBackToViridis(t *testing.T) {
	c := GetColormap("does-not-exist")
	if c.Name() != "viridis" {
		t.Fatalf("expected fallback colormap viridis, got %q", c.Name())
	}
}

func TestGetColormap_PlasmaRegistered(t *testing.T) {
	c := GetColormap("plasma")
	if c.Name() != "plasma" {
		t.Fatalf("expected plasma colormap, got %q", c.Name())
	}
	if got := c.At(0); got.B < got.R || got.B < got.G {
		t.Fatalf("expected plasma low end to be purple, got %#v", got)
	}
	if got := c.At(1); got.R < 0.9 || got.G < 0.9 {
		t.Fatalf("expected plasma high end to be yellow, got %#v", got)
	}
}

func TestListedColormapMatchesMatplotlibViridisBytes(t *testing.T) {
	c := GetColormap("viridis")
	tests := []struct {
		t          float64
		r, g, b, a uint8
	}{
		{0.000, 68, 1, 84, 255},
		{0.125, 71, 44, 123, 255},
		{0.250, 58, 82, 139, 255},
		{0.375, 44, 114, 142, 255},
		{0.500, 32, 144, 140, 255},
		{0.625, 40, 174, 127, 255},
		{0.750, 94, 201, 97, 255},
		{0.875, 173, 220, 48, 255},
		{1.000, 253, 231, 36, 255},
	}

	for _, tt := range tests {
		got := colorBytes(c.At(tt.t))
		want := [4]uint8{tt.r, tt.g, tt.b, tt.a}
		if got != want {
			t.Fatalf("viridis.At(%v) bytes = %v, want %v", tt.t, got, want)
		}
	}
}

func TestListedColormapRepresentativeBytes(t *testing.T) {
	tests := []struct {
		name string
		t    float64
		want [4]uint8
	}{
		{name: "inferno", t: 0.00, want: [4]uint8{0, 0, 3, 255}},
		{name: "inferno", t: 0.50, want: [4]uint8{187, 55, 84, 255}},
		{name: "inferno", t: 1.00, want: [4]uint8{252, 254, 164, 255}},
		{name: "magma", t: 0.00, want: [4]uint8{0, 0, 3, 255}},
		{name: "magma", t: 0.50, want: [4]uint8{182, 54, 121, 255}},
		{name: "magma", t: 1.00, want: [4]uint8{251, 252, 191, 255}},
		{name: "cividis", t: 0.00, want: [4]uint8{0, 34, 77, 255}},
		{name: "cividis", t: 0.50, want: [4]uint8{124, 123, 120, 255}},
		{name: "cividis", t: 1.00, want: [4]uint8{253, 231, 55, 255}},
	}

	for _, tt := range tests {
		got := colorBytes(GetColormap(tt.name).At(tt.t))
		if got != tt.want {
			t.Fatalf("%s.At(%v) bytes = %v, want %v", tt.name, tt.t, got, tt.want)
		}
	}
}

func TestMatplotlibPublicColormapCatalogRegistered(t *testing.T) {
	expected := []string{
		"magma", "inferno", "plasma", "viridis", "cividis",
		"twilight", "twilight_shifted", "turbo",
		"Blues", "BrBG", "BuGn", "BuPu", "CMRmap", "GnBu", "Greens", "Greys",
		"OrRd", "Oranges", "PRGn", "PiYG", "PuBu", "PuBuGn", "PuOr", "PuRd",
		"Purples", "RdBu", "RdGy", "RdPu", "RdYlBu", "RdYlGn", "Reds",
		"Spectral", "Wistia", "YlGn", "YlGnBu", "YlOrBr", "YlOrRd",
		"afmhot", "autumn", "binary", "bone", "brg", "bwr", "cool",
		"coolwarm", "copper", "cubehelix", "flag", "gist_earth", "gist_gray",
		"gist_heat", "gist_ncar", "gist_rainbow", "gist_stern", "gist_yarg",
		"gnuplot", "gnuplot2", "gray", "hot", "hsv", "jet", "nipy_spectral",
		"ocean", "pink", "prism", "rainbow", "seismic", "spring", "summer",
		"terrain", "winter", "Accent", "Dark2", "Paired", "Pastel1", "Pastel2",
		"Set1", "Set2", "Set3", "tab10", "tab20", "tab20b", "tab20c",
		"grey", "gist_grey", "gist_yerg", "Grays",
	}
	if len(matplotlibListedColormapNames) != len(expected) {
		t.Fatalf("Matplotlib colormap catalog length = %d, want %d", len(matplotlibListedColormapNames), len(expected))
	}
	for i, name := range expected {
		if matplotlibListedColormapNames[i] != name {
			t.Fatalf("Matplotlib colormap catalog[%d] = %q, want %q", i, matplotlibListedColormapNames[i], name)
		}
		cmap := GetColormap(name)
		if cmap.Name() != normalizeColormapName(name) {
			t.Fatalf("GetColormap(%q).Name() = %q, want %q", name, cmap.Name(), normalizeColormapName(name))
		}
	}
}

func TestGetColormap_ReversedVariantGeneratedFromBase(t *testing.T) {
	base := GetColormap("RdBu")
	reversed := GetColormap("RdBu_r")
	if reversed.Name() != "rdbu_r" {
		t.Fatalf("reversed colormap name = %q, want rdbu_r", reversed.Name())
	}
	if got, want := colorBytes(reversed.At(0)), colorBytes(base.At(1)); got != want {
		t.Fatalf("RdBu_r low color = %v, want base high color %v", got, want)
	}
	if got, want := colorBytes(reversed.At(1)), colorBytes(base.At(0)); got != want {
		t.Fatalf("RdBu_r high color = %v, want base low color %v", got, want)
	}
}

func TestColormapResampledCreatesListedLookup(t *testing.T) {
	base := GetColormap("viridis")
	resampled := base.Resampled(3)
	tests := []struct {
		t    float64
		want render.Color
	}{
		{t: 0, want: base.At(0)},
		{t: 0.5, want: base.At(0.5)},
		{t: 1, want: base.At(1)},
	}
	for _, tt := range tests {
		if got := resampled.At(tt.t); got != tt.want {
			t.Fatalf("resampled.At(%v) = %#v, want %#v", tt.t, got, tt.want)
		}
	}
}

func TestListedAndLinearSegmentedConstructors(t *testing.T) {
	listed := NewListedColormap("Listed Test", []render.Color{
		{R: 1, A: 1},
		{G: 1, A: 1},
		{B: 1, A: 1},
	})
	if listed.Name() != "listed test" {
		t.Fatalf("listed colormap name = %q, want listed test", listed.Name())
	}
	if got := listed.At(0.8); got != (render.Color{B: 1, A: 1}) {
		t.Fatalf("listed colormap high sample = %#v, want blue", got)
	}

	linear := NewLinearSegmentedColormap("Linear Test", []ColorStop{
		{Pos: 1, Color: render.Color{R: 1, A: 1}},
		{Pos: 0, Color: render.Color{B: 1, A: 1}},
	})
	if got := linear.At(0); got != (render.Color{B: 1, A: 1}) {
		t.Fatalf("linear colormap sorts low stop = %#v, want blue", got)
	}
	if got := linear.At(1); got != (render.Color{R: 1, A: 1}) {
		t.Fatalf("linear colormap sorts high stop = %#v, want red", got)
	}
}

func TestBinaryColormapMatchesMatplotlibSpyDefaults(t *testing.T) {
	cmap := GetColormap("binary")

	if got := cmap.At(0); got != (render.Color{R: 1, G: 1, B: 1, A: 1}) {
		t.Fatalf("binary at 0 = %+v, want white", got)
	}
	if got := cmap.At(1); got != (render.Color{R: 0, G: 0, B: 0, A: 1}) {
		t.Fatalf("binary at 1 = %+v, want black", got)
	}
}

func TestGetColormap_ChannelMapsRegistered(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "red channel", want: "red channel"},
		{name: "green channel", want: "green channel"},
		{name: "blue channel", want: "blue channel"},
	}
	for _, tt := range tests {
		c := GetColormap(tt.name)
		if c.Name() != tt.want {
			t.Fatalf("GetColormap(%q).Name() = %q, want %q", tt.name, c.Name(), tt.want)
		}
	}
}

func TestRegisterColormap_NormalizesNameAndClampsStops(t *testing.T) {
	name := "Custom Test"
	RegisterColormap(name, NewColormap(name, []ColorStop{
		{Pos: -0.5, Color: render.Color{R: 1, G: 0, B: 0, A: 1}},
		{Pos: 1.4, Color: render.Color{R: 0, G: 0, B: 1, A: 1}},
	}))

	c := GetColormap("  CuStOm tEsT  ")
	if c.Name() != "custom test" {
		t.Fatalf("expected normalized colormap name %q, got %q", "custom test", c.Name())
	}

	mid := c.At(0.5)
	if math.Abs(mid.R-0.5) > 1e-9 || math.Abs(mid.G-0) > 1e-9 || math.Abs(mid.B-0.5) > 1e-9 {
		t.Fatalf("unexpected midpoint color: %#v", mid)
	}
}

func TestColormapCopyReversedAndMutatorsAreIndependent(t *testing.T) {
	low := render.Color{R: 0.1, A: 1}
	high := render.Color{B: 0.9, A: 1}
	base := NewColormap("copy-source", []ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	})

	copyMap := base.Copy("copy-target")
	copyMap.SetBad(render.Color{G: 1, A: 1})
	if copyMap.Name() != "copy-target" {
		t.Fatalf("copy name = %q, want copy-target", copyMap.Name())
	}
	if got := copyMap.AtValue(math.NaN()); got != (render.Color{G: 1, A: 1}) {
		t.Fatalf("copy bad color = %#v", got)
	}
	if got := base.AtValue(math.NaN()); got.A != 0 {
		t.Fatalf("base bad color changed through copy: %#v", got)
	}

	reversed := base.Reversed("copy-source_r")
	if reversed.Name() != "copy-source_r" {
		t.Fatalf("reversed name = %q, want copy-source_r", reversed.Name())
	}
	if got := reversed.At(0); got != high {
		t.Fatalf("reversed low endpoint = %#v, want original high %#v", got, high)
	}
	if got := reversed.At(1); got != low {
		t.Fatalf("reversed high endpoint = %#v, want original low %#v", got, low)
	}
}

func TestColormapAtValueUsesBadUnderAndOverColors(t *testing.T) {
	bad := render.Color{R: 0.7, G: 0.7, B: 0.7, A: 0.4}
	under := render.Color{R: 0.1, G: 0.2, B: 0.9, A: 1}
	over := render.Color{R: 0.9, G: 0.2, B: 0.1, A: 1}
	c := NewColormap("bounded", []ColorStop{
		{Pos: 0, Color: render.Color{R: 0, G: 0, B: 0, A: 1}},
		{Pos: 1, Color: render.Color{R: 1, G: 1, B: 1, A: 1}},
	}).WithBad(bad).WithUnder(under).WithOver(over)

	if got := c.AtValue(math.NaN()); got != bad {
		t.Fatalf("bad color = %#v, want %#v", got, bad)
	}
	if got := c.AtValue(-0.01); got != under {
		t.Fatalf("under color = %#v, want %#v", got, under)
	}
	if got := c.AtValue(1.01); got != over {
		t.Fatalf("over color = %#v, want %#v", got, over)
	}
	if got := c.AtValue(0.5); got.R < 0.49 || got.R > 0.51 {
		t.Fatalf("in-range color = %#v, want midpoint", got)
	}
}

func TestColormapAtValueDefaultsBadTransparentAndUnderOverEndpoints(t *testing.T) {
	c := NewColormap("defaults", []ColorStop{
		{Pos: 0, Color: render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}},
		{Pos: 1, Color: render.Color{R: 0.8, G: 0.7, B: 0.6, A: 1}},
	})

	if got := c.AtValue(math.NaN()); got.A != 0 {
		t.Fatalf("default bad color = %#v, want transparent", got)
	}
	if got, want := c.AtValue(-1), c.At(0); got != want {
		t.Fatalf("default under = %#v, want low endpoint %#v", got, want)
	}
	if got, want := c.AtValue(2), c.At(1); got != want {
		t.Fatalf("default over = %#v, want high endpoint %#v", got, want)
	}
}

func TestScalarColormapLookupCoversShapeAlphaAndBadValues(t *testing.T) {
	low := render.Color{R: 1, A: 0.25}
	mid := render.Color{G: 1, A: 0.5}
	high := render.Color{B: 1, A: 0.75}
	listed := NewListedColormap("lookup-shape", []render.Color{low, mid, high})

	for _, tt := range []struct {
		t    float64
		want render.Color
	}{
		{t: -0.1, want: low},
		{t: 0, want: low},
		{t: 0.2, want: low},
		{t: 1.0 / 3.0, want: mid},
		{t: 0.9, want: high},
		{t: 1, want: high},
		{t: math.NaN(), want: low},
	} {
		if got := listed.At(tt.t); got != tt.want {
			t.Fatalf("listed.At(%v) = %#v, want %#v", tt.t, got, tt.want)
		}
	}

	empty := NewListedColormap("empty lookup-shape", nil)
	if got := empty.At(0.99); got != (render.Color{A: 1}) {
		t.Fatalf("empty listed colormap lookup = %#v, want opaque black fallback", got)
	}

	linear := NewColormap("lookup-alpha", []ColorStop{
		{Pos: 0, Color: render.Color{R: 0.2, A: 0.2}},
		{Pos: 1, Color: render.Color{R: 0.8, A: 0.8}},
	})
	if got, want := linear.At(0.5), (render.Color{R: 0.5, A: 0.5}); got != want {
		t.Fatalf("linear alpha midpoint = %#v, want %#v", got, want)
	}

	bad := render.Color{R: 0.8, A: 0.1}
	under := render.Color{G: 0.7, A: 0.2}
	over := render.Color{B: 0.6, A: 0.3}
	bounded := listed.WithBad(bad).WithUnder(under).WithOver(over)
	for _, tt := range []struct {
		t    float64
		want render.Color
	}{
		{t: math.NaN(), want: bad},
		{t: math.Inf(1), want: bad},
		{t: math.Inf(-1), want: bad},
		{t: -0.01, want: under},
		{t: 1.01, want: over},
	} {
		if got := bounded.AtValue(tt.t); got != tt.want {
			t.Fatalf("bounded.AtValue(%v) = %#v, want %#v", tt.t, got, tt.want)
		}
	}
}

func TestRegisterColormap_IgnoreEmptyName(t *testing.T) {
	// Preserve the fallback behavior when name normalization would become empty.
	defaultBefore := DefaultColormap()
	RegisterColormap("   ", NewColormap("ignored", []ColorStop{}))
	got := GetColormap("ignored")
	if got.Name() != defaultBefore.Name() {
		t.Fatalf("empty name registration should be ignored, expected %q got %q", defaultBefore.Name(), got.Name())
	}
}

func colorBytes(c render.Color) [4]uint8 {
	return [4]uint8{
		uint8(c.R * 255),
		uint8(c.G * 255),
		uint8(c.B * 255),
		uint8(c.A * 255),
	}
}
