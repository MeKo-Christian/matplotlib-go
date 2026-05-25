package style

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestDefaults(t *testing.T) {
	d := Default
	if d.DPI != 100 || d.FontKey == "" || d.FontSize <= 0 {
		t.Fatalf("unexpected defaults: %+v", d)
	}
	if d.TickCountX != 5 || d.TickCountY != 5 {
		t.Fatalf("unexpected tick defaults: %+v", d)
	}
	if len(d.ColorCycle) == 0 {
		t.Fatalf("expected default color cycle")
	}
	if got, want := d.GridColor, (render.Color{R: 0xb0 / 255.0, G: 0xb0 / 255.0, B: 0xb0 / 255.0, A: 1}); got != want {
		t.Fatalf("grid color = %+v, want Matplotlib default %+v", got, want)
	}
	if got, want := d.MinorGridColor, d.GridColor; got != want {
		t.Fatalf("minor grid color = %+v, want major grid color %+v", got, want)
	}
	if got, want := d.GridLineWidth, 0.8*d.DPI/72.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("grid line width = %v, want Matplotlib 0.8 pt at %v DPI = %v", got, d.DPI, want)
	}
	if got, want := d.MinorGridLineWidth, d.GridLineWidth; got != want {
		t.Fatalf("minor grid line width = %v, want major grid line width %v", got, want)
	}
}

func TestDefaultFontSizesMatchMatplotlib(t *testing.T) {
	d := Default
	if got, want := d.FontKey, "DejaVu Sans"; got != want {
		t.Fatalf("font family = %q, want Matplotlib sans-serif default resolved to %q", got, want)
	}
	if got, want := d.FontSize, 10.0; got != want {
		t.Fatalf("font.size = %v, want Matplotlib default %v pt", got, want)
	}
	if got, want := d.TitleSize(), 12.0; got != want {
		t.Fatalf("axes.titlesize = %v, want Matplotlib 'large' = %v pt", got, want)
	}
	if got, want := d.AxisLabelSize(), 10.0; got != want {
		t.Fatalf("axes.labelsize = %v, want Matplotlib 'medium' = %v pt", got, want)
	}
	if got, want := d.TickLabelSize("x"), 10.0; got != want {
		t.Fatalf("xtick.labelsize = %v, want Matplotlib 'medium' = %v pt", got, want)
	}
	if got, want := d.TickLabelSize("y"), 10.0; got != want {
		t.Fatalf("ytick.labelsize = %v, want Matplotlib 'medium' = %v pt", got, want)
	}
}

func TestOptionsApplyAndOrder(t *testing.T) {
	rc := Apply(
		Default,
		WithDPI(144),
		WithFont("TestFont", 14),
		WithLineWidth(2.0),
		WithTextColor(0.1, 0.2, 0.3, 0.4),
		WithLineColor(0.5, 0.6, 0.7, 0.8),
		WithBackground(0.9, 0.9, 0.9, 1.0),
		WithTickCounts(7, 9),
		WithAxesBackground(render.Color{R: 0.95, G: 0.95, B: 0.95, A: 1}),
		WithAxesEdgeColor(render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1}),
		WithAxisLineWidth(0.75),
		WithGridColors(render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}, render.Color{R: 0.9, G: 0.9, B: 0.9, A: 1}),
		WithGridLineWidths(1.1, 0.6),
		WithLegendColors(render.Color{R: 1, G: 1, B: 1, A: 1}, render.Color{R: 0, G: 0, B: 0, A: 0.2}, render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}),
	)
	if rc.DPI != 144 || rc.FontKey != "TestFont" || rc.FontSize != 14 {
		t.Fatalf("font/dpi options not applied: %+v", rc)
	}
	if rc.LineWidth != 2.0 || rc.TextColor != [4]float64{0.1, 0.2, 0.3, 0.4} {
		t.Fatalf("style color/width not applied: %+v", rc)
	}
	if rc.TickCountX != 7 || rc.TickCountY != 9 {
		t.Fatalf("tick counts not applied: %+v", rc)
	}
	if rc.AxisLineWidth != 0.75 || rc.GridLineWidth != 1.1 || rc.MinorGridLineWidth != 0.6 {
		t.Fatalf("theme widths not applied: %+v", rc)
	}

	// Order: last wins
	rc2 := Apply(Default, WithDPI(110), WithDPI(72))
	if rc2.DPI != 72 {
		t.Fatalf("expected last option to win, got %v", rc2.DPI)
	}
}

func TestPrecedence_SimulatedFigureAxes(t *testing.T) {
	// Simulate precedence: global(Default) -> figure overrides -> axes overrides
	figRC := Apply(Default, WithDPI(110), WithFont("FigFont", 11))
	axRC := Apply(figRC, WithFont("AxesFont", 9))

	// Axes font overrides figure/global
	if axRC.FontKey != "AxesFont" || axRC.FontSize != 9 {
		t.Fatalf("axes font override failed: %+v", axRC)
	}
	// Inherit figure DPI
	if axRC.DPI != 110 {
		t.Fatalf("expected DPI inherit from figure, got %v", axRC.DPI)
	}
	// Inherit defaults for untouched fields
	if axRC.LineWidth != Default.LineWidth {
		t.Fatalf("expected default line width inherit, got %v", axRC.LineWidth)
	}
}

func TestThemeLookupAndApply(t *testing.T) {
	theme, ok := GetTheme("publication")
	if !ok {
		t.Fatalf("expected publication theme to be registered")
	}
	if theme.Name != "publication" {
		t.Fatalf("unexpected theme name: %q", theme.Name)
	}

	rc := Apply(Default, WithTheme(theme), WithFont("Custom", 12))
	if rc.DPI != ThemePublication.RC.DPI {
		t.Fatalf("expected theme DPI, got %v", rc.DPI)
	}
	if rc.FontKey != "Custom" || rc.FontSize != 12 {
		t.Fatalf("expected explicit override after theme, got %+v", rc)
	}
	if got, want := rc.Palette()[0], ThemePublication.RC.Palette()[0]; got != want {
		t.Fatalf("unexpected palette head: got %+v want %+v", got, want)
	}
}

func TestAvailableThemesSorted(t *testing.T) {
	names := AvailableThemes()
	if len(names) < 3 {
		t.Fatalf("expected builtin themes, got %v", names)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Fatalf("themes not sorted: %v", names)
		}
	}
}
