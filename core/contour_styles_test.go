package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxesContourNegativeDashedLines(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	// Single-color (monochrome) contour with mixed-sign levels: negative
	// isolines should be dashed, non-negative ones solid.
	black := render.Color{A: 1}
	cs := ax.Contour([][]float64{
		{-2, -1, 0},
		{-1, 0, 1},
		{0, 1, 2},
	}, ContourOptions{
		Levels: []float64{-1, 1},
		Color:  optional.Of(black),
	})
	if cs == nil || cs.Lines == nil {
		t.Fatal("expected contour lines")
	}
	if len(cs.Lines.DashPatterns) != len(cs.Lines.Segments) {
		t.Fatalf("dash patterns = %d, want one per segment (%d)", len(cs.Lines.DashPatterns), len(cs.Lines.Segments))
	}
	for i, level := range cs.lineLevels {
		dash := cs.Lines.DashPatterns[i]
		if level < 0 && len(dash) == 0 {
			t.Fatalf("negative level %v should be dashed, got solid", level)
		}
		if level >= 0 && len(dash) != 0 {
			t.Fatalf("non-negative level %v should be solid, got %v", level, dash)
		}
	}
}

func TestAxesContourfExtendAddsBands(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	grid := [][]float64{
		{0, 1, 2, 3},
		{1, 2, 3, 4},
		{2, 3, 4, 5},
		{3, 4, 5, 6},
	}
	levels := []float64{2, 3, 4}

	base := ax.Contourf(grid, ContourOptions{Levels: levels})
	ext := ax.Contourf(grid, ContourOptions{Levels: levels, Extend: "both"})
	if base == nil || ext == nil || base.Fills == nil || ext.Fills == nil {
		t.Fatal("expected filled contours")
	}
	if len(ext.Fills.Polygons) <= len(base.Fills.Polygons) {
		t.Fatalf("extend=both should add under/over bands: base=%d ext=%d",
			len(base.Fills.Polygons), len(ext.Fills.Polygons))
	}
	// Public levels stay at the user-supplied range.
	if len(ext.Levels) != len(levels) {
		t.Fatalf("extend should not change public levels: got %v", ext.Levels)
	}
}

func TestAxesContourfHatchesCyclePerBand(t *testing.T) {
	fig := NewFigure(320, 240)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	cs := ax.Contourf([][]float64{
		{0, 1, 2, 3},
		{1, 2, 3, 4},
		{2, 3, 4, 5},
	}, ContourOptions{
		Levels:  []float64{0, 1, 2, 3, 4},
		Hatches: []string{"/", "\\"},
	})
	if cs == nil || cs.Fills == nil {
		t.Fatal("expected filled contours")
	}
	// Hatched contourf merges each band into one compound path so the hatch
	// tiles continuously; there is one path (and one hatch) per band.
	if len(cs.Fills.Polygons) != 0 {
		t.Fatalf("hatched contourf should use compound paths, got %d polygons", len(cs.Fills.Polygons))
	}
	if len(cs.Fills.Hatches) != len(cs.Fills.Paths) {
		t.Fatalf("hatches = %d, want one per band path (%d)", len(cs.Fills.Hatches), len(cs.Fills.Paths))
	}
	if cs.Fills.HatchColor.A <= 0 {
		t.Fatal("expected a default hatch color to be set")
	}
	seen := map[string]bool{}
	for _, h := range cs.Fills.Hatches {
		seen[h] = true
	}
	if !seen["/"] || !seen["\\"] {
		t.Fatalf("expected both hatch patterns cycled, saw %v", seen)
	}
}

func TestContourLineStyleDashes(t *testing.T) {
	if got := lineStyleToDashes("solid", 1); got != nil {
		t.Fatalf("solid dashes = %v, want nil", got)
	}
	if got := lineStyleToDashes("-", 2); got != nil {
		t.Fatalf("\"-\" dashes = %v, want nil", got)
	}
	// dashed base [3.7, 1.6] scaled by linewidth 2.
	got := lineStyleToDashes("dashed", 2)
	want := []float64{7.4, 3.2}
	if len(got) != len(want) {
		t.Fatalf("dashed dashes = %v, want %v", got, want)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("dashed dashes = %v, want %v", got, want)
		}
	}
	if got := lineStyleToDashes("--", 1); len(got) != 2 {
		t.Fatalf("\"--\" dashes = %v, want 2-element pattern", got)
	}
	if got := lineStyleToDashes("dashdot", 1); len(got) != 4 {
		t.Fatalf("dashdot dashes = %v, want 4-element pattern", got)
	}
	if got := lineStyleToDashes("dotted", 1); len(got) != 2 {
		t.Fatalf("dotted dashes = %v, want 2-element pattern", got)
	}
}

func TestResolveContourLineStylesExplicitCycles(t *testing.T) {
	levels := []float64{1, 2, 3, 4, 5}
	styles := resolveContourLineStyles(levels, ContourOptions{LineStyles: []string{"solid", "dashed"}}, false)
	want := []string{"solid", "dashed", "solid", "dashed", "solid"}
	for i := range want {
		if styles[i] != want[i] {
			t.Fatalf("styles = %v, want %v", styles, want)
		}
	}
}

func TestResolveContourLineStylesNegativeMonochrome(t *testing.T) {
	levels := []float64{-2, -1, 0, 1, 2}
	// Monochrome with no explicit styles: negative levels become dashed.
	styles := resolveContourLineStyles(levels, ContourOptions{}, true)
	want := []string{"dashed", "dashed", "solid", "solid", "solid"}
	for i := range want {
		if styles[i] != want[i] {
			t.Fatalf("monochrome styles = %v, want %v", styles, want)
		}
	}

	// Non-monochrome: no negative dashing.
	styles = resolveContourLineStyles(levels, ContourOptions{}, false)
	for i, s := range styles {
		if s != "solid" {
			t.Fatalf("non-monochrome style[%d] = %q, want solid", i, s)
		}
	}

	// Custom negative style override.
	dotted := "dotted"
	styles = resolveContourLineStyles(levels, ContourOptions{NegativeLineStyles: optional.Of(dotted)}, true)
	if styles[0] != "dotted" || styles[1] != "dotted" {
		t.Fatalf("override negative styles = %v, want leading dotted", styles)
	}
}

func TestContourMonochrome(t *testing.T) {
	c := render.Color{A: 1}
	if !contourMonochrome(ContourOptions{Color: optional.Of(c)}) {
		t.Fatal("single Color should be monochrome")
	}
	if !contourMonochrome(ContourOptions{Colors: []render.Color{c}}) {
		t.Fatal("single Colors entry should be monochrome")
	}
	if contourMonochrome(ContourOptions{Colors: []render.Color{c, c}}) {
		t.Fatal("multiple explicit colors should not be monochrome")
	}
	if contourMonochrome(ContourOptions{}) {
		t.Fatal("default colormap should not be monochrome")
	}
}

func TestContourExtendedLevels(t *testing.T) {
	levels := []float64{0, 1, 2}
	if got := contourExtendedLevels(levels, "neither"); len(got) != 3 {
		t.Fatalf("neither = %v, want unchanged", got)
	}
	both := contourExtendedLevels(levels, "both")
	if len(both) != 5 || both[0] > -1e200 || both[len(both)-1] < 1e200 {
		t.Fatalf("both = %v, want sentinel-extended on both ends", both)
	}
	minOnly := contourExtendedLevels(levels, "min")
	if len(minOnly) != 4 || minOnly[0] > -1e200 || minOnly[len(minOnly)-1] != 2 {
		t.Fatalf("min = %v, want low sentinel only", minOnly)
	}
	maxOnly := contourExtendedLevels(levels, "max")
	if len(maxOnly) != 4 || maxOnly[0] != 0 || maxOnly[len(maxOnly)-1] < 1e200 {
		t.Fatalf("max = %v, want high sentinel only", maxOnly)
	}
}

func TestContourBandHatchCycles(t *testing.T) {
	hatches := []string{"/", "\\", "x"}
	for idx, want := range []string{"/", "\\", "x", "/", "\\"} {
		if got := contourBandHatch(hatches, idx); got != want {
			t.Fatalf("hatch[%d] = %q, want %q", idx, got, want)
		}
	}
	if got := contourBandHatch(nil, 0); got != "" {
		t.Fatalf("nil hatch = %q, want empty", got)
	}
}
