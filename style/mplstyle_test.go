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

func TestParseMPLStyleTickGeometryAndVisibility(t *testing.T) {
	src := `
xtick.direction: in
xtick.alignment: right
xtick.bottom: true
xtick.top: true
xtick.labelbottom: false
xtick.labeltop: true
xtick.major.size: 7
xtick.major.width: 1.25
xtick.major.pad: 6
xtick.major.bottom: true
xtick.major.top: false
xtick.minor.size: 4
xtick.minor.width: 0.5
xtick.minor.pad: 2
xtick.minor.visible: true
xtick.minor.ndivs: 3
xtick.minor.bottom: false
xtick.minor.top: true
ytick.direction: inout
ytick.alignment: top
ytick.left: false
ytick.right: true
ytick.labelleft: false
ytick.labelright: true
ytick.major.size: 8
ytick.major.width: 1.5
ytick.major.pad: 5
ytick.major.left: true
ytick.major.right: true
ytick.minor.size: 3
ytick.minor.width: 0.75
ytick.minor.pad: 1
ytick.minor.visible: true
ytick.minor.ndivs: auto
ytick.minor.left: true
ytick.minor.right: false
`
	theme, report, err := ParseMPLStyle("ticks", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unsupported tick params: %+v", report.Unsupported)
	}
	if got, want := len(report.Applied), 36; got != want {
		t.Fatalf("applied tick params = %d, want %d", got, want)
	}

	x := theme.RC.XTick
	if x.Direction != "in" || x.Alignment != "right" ||
		!x.Primary || !x.Secondary || x.LabelPrimary || !x.LabelSecondary {
		t.Fatalf("x tick axis params = %+v", x)
	}
	if x.Major.Size != 7 || x.Major.Width != 1.25 || x.Major.Pad != 6 ||
		!x.Major.Primary || x.Major.Secondary {
		t.Fatalf("x major params = %+v", x.Major)
	}
	if x.Minor.Size != 4 || x.Minor.Width != 0.5 || x.Minor.Pad != 2 ||
		!x.Minor.Visible || x.Minor.NDivs != 3 || x.Minor.Primary || !x.Minor.Secondary {
		t.Fatalf("x minor params = %+v", x.Minor)
	}

	y := theme.RC.YTick
	if y.Direction != "inout" || y.Alignment != "top" ||
		y.Primary || !y.Secondary || y.LabelPrimary || !y.LabelSecondary {
		t.Fatalf("y tick axis params = %+v", y)
	}
	if y.Major.Size != 8 || y.Major.Width != 1.5 || y.Major.Pad != 5 ||
		!y.Major.Primary || !y.Major.Secondary {
		t.Fatalf("y major params = %+v", y.Major)
	}
	if y.Minor.Size != 3 || y.Minor.Width != 0.75 || y.Minor.Pad != 1 ||
		!y.Minor.Visible || y.Minor.NDivs != 0 || !y.Minor.Primary || y.Minor.Secondary {
		t.Fatalf("y minor params = %+v", y.Minor)
	}
}

func TestParseMPLStyleRejectsInvalidTickParams(t *testing.T) {
	for _, src := range []string{
		"xtick.direction: sideways\n",
		"xtick.alignment: baseline\n",
		"ytick.alignment: right\n",
		"ytick.minor.ndivs: -1\n",
	} {
		if _, _, err := ParseMPLStyle("ticks", src); err == nil {
			t.Fatalf("ParseMPLStyle(%q) succeeded, want validation error", src)
		}
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

func TestParseMPLStyleLegendLayout(t *testing.T) {
	src := `
font.size: 10
legend.loc: lower left
legend.fancybox: false
legend.shadow: true
legend.numpoints: 3
legend.scatterpoints: 2
legend.markerscale: 1.75
legend.title_fontsize: large
legend.borderpad: 0.6
legend.labelspacing: 0.7
legend.handlelength: 2.5
legend.handleheight: 0.9
legend.handletextpad: 1.1
legend.borderaxespad: 0.8
legend.columnspacing: 2.4
`
	theme, report, err := ParseMPLStyle("legend-layout", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unsupported legend params: %+v", report.Unsupported)
	}

	got := theme.RC.Legend
	if got.Location != "lower left" || got.FancyBox || !got.Shadow {
		t.Fatalf("unexpected legend placement/frame defaults: %+v", got)
	}
	if got.NumPoints != 3 || got.ScatterPoints != 2 {
		t.Fatalf("unexpected legend point counts: %+v", got)
	}
	if !almostEqual(got.MarkerScale, 1.75) || !almostEqual(got.TitleFontSize, 12) {
		t.Fatalf("unexpected legend marker/title sizes: %+v", got)
	}
	wantDimensions := []float64{0.6, 0.7, 2.5, 0.9, 1.1, 0.8, 2.4}
	gotDimensions := []float64{
		got.BorderPad,
		got.LabelSpacing,
		got.HandleLength,
		got.HandleHeight,
		got.HandleTextPad,
		got.BorderAxesPad,
		got.ColumnSpacing,
	}
	if !equalFloatSlices(gotDimensions, wantDimensions) {
		t.Fatalf("legend dimensions = %v, want %v", gotDimensions, wantDimensions)
	}
}

func TestParseMPLStyleLegendTitleFontSizeNoneAndNumericLocation(t *testing.T) {
	theme, report, err := ParseMPLStyle("legend-special-values", `
legend.loc: 9
legend.title_fontsize: None
`)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unsupported legend params: %+v", report.Unsupported)
	}
	if got := theme.RC.Legend.Location; got != "upper center" {
		t.Fatalf("legend location = %q, want upper center", got)
	}
	if got := theme.RC.Legend.TitleFontSize; got != 0 {
		t.Fatalf("legend title fontsize = %v, want None sentinel 0", got)
	}
}

func TestParseMPLStyleLegendRelativeTitleSizeIsOrderIndependent(t *testing.T) {
	theme, _, err := ParseMPLStyle("legend-title-order", `
legend.title_fontsize: large
font.size: 20
`)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if got, want := theme.RC.Legend.TitleFontSize, 24.0; !almostEqual(got, want) {
		t.Fatalf("legend title fontsize = %v, want %v", got, want)
	}
}

func TestParseMPLStyleAxesTitleAndLabelPlacement(t *testing.T) {
	theme, report, err := ParseMPLStyle("axes-text-placement", `
axes.titlelocation: right
axes.titlepad: 9.5
axes.titleweight: bold
axes.titley: 0.83
axes.labelpad: 7
axes.labelweight: 650
`)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}
	got := theme.RC.Axes
	if got.TitleLocation != "right" || !almostEqual(got.TitlePad, 9.5) || got.TitleWeight != 700 {
		t.Fatalf("unexpected axes title defaults: %+v", got)
	}
	if !got.TitleYSet || !almostEqual(got.TitleY, 0.83) {
		t.Fatalf("unexpected axes title y default: value=%v set=%v", got.TitleY, got.TitleYSet)
	}
	if !almostEqual(got.LabelPad, 7) || got.LabelWeight != 650 {
		t.Fatalf("unexpected axes label defaults: %+v", got)
	}
	params := paramsFromRC(theme.RC)
	if params["axes.titlelocation"] != "right" || params["axes.titlepad"] != "9.5" ||
		params["axes.titleweight"] != "700" || params["axes.titley"] != "0.83" ||
		params["axes.labelpad"] != "7" || params["axes.labelweight"] != "650" {
		t.Fatalf("unexpected runtime axes text params: %+v", params)
	}

	theme, _, err = ParseMPLStyle("axes-title-auto", "axes.titley: None\n")
	if err != nil {
		t.Fatalf("ParseMPLStyle(titley=None) error = %v", err)
	}
	if theme.RC.Axes.TitleYSet {
		t.Fatalf("axes.titley None left explicit value enabled: %+v", theme.RC.Axes)
	}
	if got := paramsFromRC(theme.RC)["axes.titley"]; got != "None" {
		t.Fatalf("serialized axes.titley = %q, want None", got)
	}
}

func TestParseMPLStyleAxesFormatterDefaults(t *testing.T) {
	theme, report, err := ParseMPLStyle("axes-formatter", `
axes.formatter.limits: -3, 7
axes.formatter.min_exponent: 2
axes.formatter.offset_threshold: 5
axes.formatter.use_locale: True
axes.formatter.use_mathtext: True
axes.formatter.useoffset: False
`)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}
	got := theme.RC.Axes.Formatter
	if got.Limits != [2]int{-3, 7} || got.MinExponent != 2 || got.OffsetThreshold != 5 ||
		!got.UseLocale || !got.UseMathText || got.UseOffset {
		t.Fatalf("unexpected axes formatter defaults: %+v", got)
	}
	params := paramsFromRC(theme.RC)
	if params["axes.formatter.limits"] != "-3, 7" ||
		params["axes.formatter.min_exponent"] != "2" ||
		params["axes.formatter.offset_threshold"] != "5" ||
		params["axes.formatter.use_locale"] != "True" ||
		params["axes.formatter.use_mathtext"] != "True" ||
		params["axes.formatter.useoffset"] != "False" {
		t.Fatalf("unexpected runtime axes formatter params: %+v", params)
	}
}

func TestParseMPLStyleAxesFormatterLimitsRejectsWrongArity(t *testing.T) {
	if _, _, err := ParseMPLStyle("axes-formatter", "axes.formatter.limits: -5\n"); err == nil {
		t.Fatal("ParseMPLStyle accepted one formatter limit, want error")
	}
}

func TestParseMPLStyleAxesSpineVisibility(t *testing.T) {
	theme, report, err := ParseMPLStyle("axes-spines", `
axes.spines.top: False
axes.spines.bottom: true
axes.spines.left: no
axes.spines.right: yes
`)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported entries: %+v", report.Unsupported)
	}
	got := theme.RC.Axes.Spines
	want := SpineRC{Top: false, Bottom: true, Left: false, Right: true}
	if got != want {
		t.Fatalf("axes spine defaults = %+v, want %+v", got, want)
	}
	params := paramsFromRC(theme.RC)
	if params["axes.spines.top"] != "False" ||
		params["axes.spines.bottom"] != "True" ||
		params["axes.spines.left"] != "False" ||
		params["axes.spines.right"] != "True" {
		t.Fatalf("unexpected runtime axes spine params: %+v", params)
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
