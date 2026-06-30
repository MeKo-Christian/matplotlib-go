package style

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestParseMPLStyleSubset(t *testing.T) {
	src := `
font.size: 10.0
font.family: "DejaVu Sans"
lines.linewidth: 2.0
lines.color: C1
text.color: "#333333"
text.usetex: true
axes.facecolor: E5E5E5
axes.edgecolor: white
axes.linewidth: 1.0
axes.labelcolor: 555555
axes.prop_cycle: cycler('color', ['E24A33', '348ABD', '988ED5'])
xtick.color: 555555
ytick.color: 555555
grid.color: white
grid.alpha: 0.75
grid.linewidth: 0.5
legend.facecolor: inherit
legend.edgecolor: 0.5
legend.labelcolor: black
figure.facecolor: white
patch.facecolor: 348ABD
`

	theme, report, err := ParseMPLStyle("GGPlot.mplstyle", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}

	if theme.Name != "ggplot" {
		t.Fatalf("theme name = %q, want ggplot", theme.Name)
	}
	if len(report.Applied) != 20 {
		t.Fatalf("applied count = %d, want 20", len(report.Applied))
	}
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "patch.facecolor" {
		t.Fatalf("unexpected unsupported report: %+v", report.Unsupported)
	}

	if theme.RC.FontKey != "DejaVu Sans" || theme.RC.FontSize != 10 {
		t.Fatalf("unexpected font settings: %+v", theme.RC)
	}
	if got, want := theme.RC.LineWidth, 2.0; !almostEqual(got, want) {
		t.Fatalf("line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.AxisLineWidth, 1.0; !almostEqual(got, want) {
		t.Fatalf("axis line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.GridLineWidth, 0.5; !almostEqual(got, want) {
		t.Fatalf("grid line width = %v, want %v", got, want)
	}
	if got, want := theme.RC.MinorGridLineWidth, theme.RC.GridLineWidth; !almostEqual(got, want) {
		t.Fatalf("minor grid line width = %v, want %v", got, want)
	}
	if got := theme.RC.AxesBackground; !almostEqual(got.R, 0xE5/255.0) || !almostEqual(got.G, 0xE5/255.0) || !almostEqual(got.B, 0xE5/255.0) {
		t.Fatalf("axes background = %+v", got)
	}
	if got := theme.RC.AxesEdgeColor; got.R != 1 || got.G != 1 || got.B != 1 {
		t.Fatalf("axes edge color = %+v", got)
	}
	if got := theme.RC.XTickColor; !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("x tick color = %+v", got)
	}
	if got := theme.RC.YTickColor; !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("y tick color = %+v", got)
	}
	if got := theme.RC.DefaultTextColor(); !almostEqual(got.R, 0x33/255.0) || !almostEqual(got.G, 0x33/255.0) || !almostEqual(got.B, 0x33/255.0) {
		t.Fatalf("text color = %+v", got)
	}
	if !theme.RC.UseTeX {
		t.Fatal("expected UseTeX to be enabled from mplstyle")
	}
	if got := theme.RC.DefaultAxesLabelColor(); !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("axes label color = %+v", got)
	}
	if got := theme.RC.GridColor; !almostEqual(got.A, 0.75) {
		t.Fatalf("grid alpha = %v, want 0.75", got.A)
	}
	if got, want := theme.RC.LegendBackground, theme.RC.AxesBackground; got != want {
		t.Fatalf("legend background = %+v, want inherit %+v", got, want)
	}
	if got := theme.RC.LegendBorderColor; !almostEqual(got.R, 0.5) || !almostEqual(got.G, 0.5) || !almostEqual(got.B, 0.5) {
		t.Fatalf("legend border color = %+v", got)
	}
	if got, want := theme.RC.Palette()[0], mustParseTestColor(t, "E24A33"); got != want {
		t.Fatalf("palette[0] = %+v, want %+v", got, want)
	}
	if got, want := theme.RC.DefaultLineColor(), theme.RC.Palette()[1]; got != want {
		t.Fatalf("line color = %+v, want %+v", got, want)
	}
}

func TestParseMPLStyleCyclerKeywordForm(t *testing.T) {
	src := `
axes.prop_cycle: cycler(color=['003FFF', '03ED3A'])
`

	theme, report, err := ParseMPLStyle("custom", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}
	if got, want := theme.RC.Palette()[0], mustParseTestColor(t, "003FFF"); got != want {
		t.Fatalf("palette[0] = %+v, want %+v", got, want)
	}
	if got, want := theme.RC.Palette()[1], mustParseTestColor(t, "03ED3A"); got != want {
		t.Fatalf("palette[1] = %+v, want %+v", got, want)
	}
	if theme.RC.PropCycle != nil {
		t.Fatalf("color-only cycle should leave PropCycle nil, got %+v", theme.RC.PropCycle)
	}
}

func TestParseMPLStyleMultiPropertyCycleConcat(t *testing.T) {
	src := `
axes.prop_cycle: cycler('color', ['FF0000', '00FF00']) + cycler('linestyle', ['-', '--'])
`
	theme, _, err := ParseMPLStyle("custom", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	pc := theme.RC.PropCycle
	if pc == nil {
		t.Fatal("expected non-nil PropCycle for multi-property cycle")
	}
	if pc.Len() != 2 {
		t.Fatalf("cycle len = %d, want 2", pc.Len())
	}
	if got, ok := pc.Row(0)["linestyle"].(string); !ok || got != "-" {
		t.Fatalf("row0 linestyle = %v, want -", pc.Row(0)["linestyle"])
	}
	if got, ok := pc.Row(1)["linestyle"].(string); !ok || got != "--" {
		t.Fatalf("row1 linestyle = %v, want --", pc.Row(1)["linestyle"])
	}
	// The color column still drives the palette.
	if got, want := theme.RC.Palette()[0], mustParseTestColor(t, "FF0000"); got != want {
		t.Fatalf("palette[0] = %+v, want %+v", got, want)
	}
}

func TestParseMPLStyleMultiPropertyCycleProduct(t *testing.T) {
	src := `
axes.prop_cycle: cycler('color', ['FF0000', '00FF00']) * cycler('linewidth', [1.0, 2.0])
`
	theme, _, err := ParseMPLStyle("custom", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	pc := theme.RC.PropCycle
	if pc == nil || pc.Len() != 4 {
		t.Fatalf("product cycle len = %d, want 4", pc.Len())
	}
	// Left (color) varies slowest: (red,1),(red,2),(green,1),(green,2).
	if got := pc.Row(0)["linewidth"].(float64); got != 1.0 {
		t.Fatalf("row0 linewidth = %v, want 1.0", got)
	}
	if got := pc.Row(1)["linewidth"].(float64); got != 2.0 {
		t.Fatalf("row1 linewidth = %v, want 2.0", got)
	}
	palette := theme.RC.Palette()
	if got, want := palette[0], mustParseTestColor(t, "FF0000"); got != want {
		t.Fatalf("palette[0] = %+v, want red", got)
	}
	if got, want := palette[2], mustParseTestColor(t, "00FF00"); got != want {
		t.Fatalf("palette[2] = %+v, want green", got)
	}
}

func TestParseMPLStyleBroaderCoverage(t *testing.T) {
	src := `
font.size: 12
figure.figsize: 7.5, 4.25
axes.grid: true
axes.grid.axis: y
axes.grid.which: both
axes.titlecolor: "#221144"
axes.titlesize: large
axes.labelcolor: "#334455"
axes.labelsize: small
xtick.color: "#445566"
xtick.labelsize: x-small
ytick.labelcolor: "#556677"
ytick.labelsize: 9
grid.linestyle: --
grid.minor.linestyle: :
legend.fontsize: medium
legend.framealpha: 0.4
legend.frameon: false
`

	theme, report, err := ParseMPLStyle("broader", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}

	if got, want := theme.RC.FigureWidth, 7.5; !almostEqual(got, want) {
		t.Fatalf("figure width = %v, want %v", got, want)
	}
	if got, want := theme.RC.FigureHeight, 4.25; !almostEqual(got, want) {
		t.Fatalf("figure height = %v, want %v", got, want)
	}
	if !theme.RC.GridVisible || theme.RC.GridAxis != "y" || theme.RC.GridWhich != "both" {
		t.Fatalf("unexpected grid defaults: visible=%v axis=%q which=%q", theme.RC.GridVisible, theme.RC.GridAxis, theme.RC.GridWhich)
	}
	if got, want := theme.RC.TitleSize(), 12*1.2; !almostEqual(got, want) {
		t.Fatalf("title size = %v, want %v", got, want)
	}
	if got, want := theme.RC.AxisLabelSize(), 10.0; !almostEqual(got, want) {
		t.Fatalf("label size = %v, want %v", got, want)
	}
	if got, want := theme.RC.TickLabelSize("x"), 8.33; math.Abs(got-want) > 0.02 {
		t.Fatalf("x tick size = %v, want about %v", got, want)
	}
	if got, want := theme.RC.TickLabelSize("y"), 9.0; !almostEqual(got, want) {
		t.Fatalf("y tick size = %v, want %v", got, want)
	}
	if got := theme.RC.DefaultAxesTitleColor(); !almostEqual(got.R, 0x22/255.0) || !almostEqual(got.G, 0x11/255.0) || !almostEqual(got.B, 0x44/255.0) {
		t.Fatalf("title color = %+v", got)
	}
	if got := theme.RC.DefaultAxesLabelColor(); !almostEqual(got.R, 0x33/255.0) || !almostEqual(got.G, 0x44/255.0) || !almostEqual(got.B, 0x55/255.0) {
		t.Fatalf("label color = %+v", got)
	}
	if got := theme.RC.XTickColor; !almostEqual(got.R, 0x44/255.0) || !almostEqual(got.G, 0x55/255.0) || !almostEqual(got.B, 0x66/255.0) {
		t.Fatalf("x tick color = %+v", got)
	}
	if got := theme.RC.YTickColor; !almostEqual(got.R, 0x55/255.0) || !almostEqual(got.G, 0x66/255.0) || !almostEqual(got.B, 0x77/255.0) {
		t.Fatalf("y tick color = %+v", got)
	}
	if got, want := theme.RC.GridDashes, []float64{6, 6}; !equalFloatSlices(got, want) {
		t.Fatalf("grid dashes = %v, want %v", got, want)
	}
	if got, want := theme.RC.MinorGridDashes, []float64{1.2, 2.4}; !equalFloatSlices(got, want) {
		t.Fatalf("minor grid dashes = %v, want %v", got, want)
	}
	if got, want := theme.RC.LegendSize(), 12.0; !almostEqual(got, want) {
		t.Fatalf("legend size = %v, want %v", got, want)
	}
	if got, want := theme.RC.LegendFrameAlpha, 0.4; !almostEqual(got, want) {
		t.Fatalf("legend frame alpha = %v, want %v", got, want)
	}
	if theme.RC.LegendFrameOn {
		t.Fatal("legend frame should be disabled")
	}
	if got := theme.RC.LegendBackground.A; got != 0 {
		t.Fatalf("legend background alpha = %v, want 0", got)
	}
}

func TestParseMPLStyleInvalidValue(t *testing.T) {
	_, _, err := ParseMPLStyle("broken", "lines.linewidth: nope\n")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadMPLStyleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dark_background.mplstyle")
	if err := os.WriteFile(path, []byte("figure.facecolor: black\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	theme, report, err := LoadMPLStyleFile(path)
	if err != nil {
		t.Fatalf("LoadMPLStyleFile() error = %v", err)
	}
	if theme.Name != "dark_background" {
		t.Fatalf("theme name = %q, want dark_background", theme.Name)
	}
	if len(report.Applied) != 1 {
		t.Fatalf("applied count = %d, want 1", len(report.Applied))
	}
	if got := theme.RC.FigureBackground(); got.R != 0 || got.G != 0 || got.B != 0 || got.A != 1 {
		t.Fatalf("figure background = %+v, want black", got)
	}
}

func TestSupportedMPLStyleKeysSorted(t *testing.T) {
	keys := SupportedMPLStyleKeys()
	if len(keys) == 0 {
		t.Fatal("expected supported keys")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("supported keys not sorted: %v", keys)
		}
	}
}

func mustParseTestColor(t *testing.T, value string) render.Color {
	t.Helper()
	parsed, err := parseMPLColor(value, Default)
	if err != nil {
		t.Fatalf("parseMPLColor(%q) error = %v", value, err)
	}
	return parsed
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func equalFloatSlices(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !almostEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}
