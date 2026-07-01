package style

import "testing"

func TestHatchRCParams(t *testing.T) {
	src := "hatch.color: red\nhatch.linewidth: 2.5\n"
	theme, report, err := ParseMPLStyle("hatch", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	if got := theme.RC.Hatch.Color; got.R != 1 || got.G != 0 || got.B != 0 {
		t.Fatalf("hatch color = %+v, want red", got)
	}
	if got := theme.RC.Hatch.LineWidth; !almostEqual(got, 2.5) {
		t.Fatalf("hatch linewidth = %v, want 2.5", got)
	}
}

func TestImageRCParams(t *testing.T) {
	src := "image.cmap: plasma\nimage.interpolation: nearest\nimage.origin: lower\nimage.aspect: auto\nimage.resample: False\nimage.lut: 128\n"
	theme, report, err := ParseMPLStyle("image", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	img := theme.RC.Image
	if img.Cmap != "plasma" {
		t.Fatalf("cmap = %q, want plasma", img.Cmap)
	}
	if img.Interpolation != "nearest" {
		t.Fatalf("interpolation = %q, want nearest", img.Interpolation)
	}
	if img.Origin != "lower" {
		t.Fatalf("origin = %q, want lower", img.Origin)
	}
	if img.Aspect != "auto" {
		t.Fatalf("aspect = %q, want auto", img.Aspect)
	}
	if img.Resample {
		t.Fatalf("resample = true, want false")
	}
	if img.LUT != 128 {
		t.Fatalf("lut = %d, want 128", img.LUT)
	}
}

func TestImageRCRejectsBadEnum(t *testing.T) {
	if _, _, err := ParseMPLStyle("image", "image.origin: sideways\n"); err == nil {
		t.Fatal("expected error for invalid image.origin")
	}
}

func TestMathtextRCParams(t *testing.T) {
	src := "mathtext.fontset: cm\nmathtext.default: rm\nmathtext.fallback: None\nmathtext.rm: serif\n"
	theme, report, err := ParseMPLStyle("mathtext", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	mtxt := theme.RC.Mathtext
	if mtxt.Fontset != "cm" {
		t.Fatalf("fontset = %q, want cm", mtxt.Fontset)
	}
	if mtxt.Default != "rm" {
		t.Fatalf("default = %q, want rm", mtxt.Default)
	}
	if mtxt.Fallback != "" {
		t.Fatalf("fallback = %q, want empty (None)", mtxt.Fallback)
	}
	if mtxt.RM != "serif" {
		t.Fatalf("rm = %q, want serif", mtxt.RM)
	}
}

func TestMathtextRCRejectsBadFontset(t *testing.T) {
	if _, _, err := ParseMPLStyle("mathtext", "mathtext.fontset: comicsans\n"); err == nil {
		t.Fatal("expected error for invalid mathtext.fontset")
	}
}

func TestAnimationRCParams(t *testing.T) {
	src := "animation.writer: imagemagick\nanimation.bitrate: 1800\nanimation.html: jshtml\nanimation.convert_args: -delay, 10\nanimation.embed_limit: 32.5\n"
	theme, report, err := ParseMPLStyle("animation", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	a := theme.RC.Animation
	if a.Writer != "imagemagick" || a.Bitrate != 1800 || a.HTML != "jshtml" {
		t.Fatalf("animation = %+v", a)
	}
	if len(a.ConvertArgs) != 2 || a.ConvertArgs[0] != "-delay" || a.ConvertArgs[1] != "10" {
		t.Fatalf("convert args = %v", a.ConvertArgs)
	}
	if !almostEqual(a.EmbedLimit, 32.5) {
		t.Fatalf("embed limit = %v, want 32.5", a.EmbedLimit)
	}
}

func TestAnimationRCArgsAreClonedByApply(t *testing.T) {
	base := Apply(Default)
	clone := Apply(base)
	if len(clone.Animation.ConvertArgs) == 0 {
		t.Fatal("expected default convert args")
	}
	clone.Animation.ConvertArgs[0] = "mutated"
	if base.Animation.ConvertArgs[0] == "mutated" {
		t.Fatal("Apply did not clone Animation.ConvertArgs; slices alias")
	}
}

func TestBackendRCParams(t *testing.T) {
	src := "pdf.fonttype: 42\npdf.use14corefonts: True\nps.fonttype: 42\nps.useafm: True\nps.papersize: a4\nsvg.fonttype: none\nsvg.hashsalt: salty\n"
	theme, report, err := ParseMPLStyle("backend", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	if theme.RC.PDF.FontType != 42 || !theme.RC.PDF.Use14CoreFonts {
		t.Fatalf("pdf = %+v", theme.RC.PDF)
	}
	if theme.RC.PS.FontType != 42 || !theme.RC.PS.UseAFM || theme.RC.PS.PaperSize != "a4" {
		t.Fatalf("ps = %+v", theme.RC.PS)
	}
	if theme.RC.SVG.FontType != "none" || theme.RC.SVG.HashSalt != "salty" {
		t.Fatalf("svg = %+v", theme.RC.SVG)
	}
}

func TestBackendRCRejectsBadFontType(t *testing.T) {
	if _, _, err := ParseMPLStyle("backend", "pdf.fonttype: 7\n"); err == nil {
		t.Fatal("expected error for invalid pdf.fonttype")
	}
}

func TestSavefigRCParams(t *testing.T) {
	src := "savefig.dpi: figure\nsavefig.facecolor: red\nsavefig.transparent: True\nsavefig.bbox: tight\nsavefig.pad_inches: 0.25\nsavefig.format: pdf\n"
	theme, report, err := ParseMPLStyle("savefig", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	sf := theme.RC.Savefig
	if sf.Dpi != 0 {
		t.Fatalf("dpi = %v, want 0 (figure)", sf.Dpi)
	}
	if sf.Facecolor != "red" || !sf.Transparent {
		t.Fatalf("savefig = %+v", sf)
	}
	if sf.BboxInches != "tight" || !almostEqual(sf.PadInches, 0.25) {
		t.Fatalf("bbox/pad = %q/%v", sf.BboxInches, sf.PadInches)
	}
	if sf.Format != "pdf" {
		t.Fatalf("format = %q, want pdf", sf.Format)
	}
}

func TestSavefigRCRejectsBadBbox(t *testing.T) {
	if _, _, err := ParseMPLStyle("savefig", "savefig.bbox: snug\n"); err == nil {
		t.Fatal("expected error for invalid savefig.bbox")
	}
}

func TestDateRCParams(t *testing.T) {
	src := "date.autoformatter.second: %H:%M:%S\ndate.converter: concise\ndate.interval_multiples: False\n"
	theme, report, err := ParseMPLStyle("date", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	d := theme.RC.Date
	if d.AutoSecond != "%H:%M:%S" {
		t.Fatalf("autoformatter.second = %q, want %%H:%%M:%%S", d.AutoSecond)
	}
	if d.Converter != "concise" {
		t.Fatalf("converter = %q, want concise", d.Converter)
	}
	if d.IntervalMultiples {
		t.Fatal("interval_multiples = true, want false")
	}
}

func TestBoxplotRCParams(t *testing.T) {
	src := "boxplot.showmeans: True\nboxplot.showcaps: False\nboxplot.boxprops.linewidth: 2.0\nboxplot.flierprops.markersize: 9\nboxplot.medianprops.color: red\n"
	theme, report, err := ParseMPLStyle("boxplot", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	b := theme.RC.Boxplot
	if !b.ShowMeans {
		t.Fatal("showmeans = false, want true")
	}
	if b.ShowCaps {
		t.Fatal("showcaps = true, want false")
	}
	if !almostEqual(b.BoxLineWidth, 2.0) {
		t.Fatalf("box linewidth = %v, want 2.0", b.BoxLineWidth)
	}
	if !almostEqual(b.FlierMarkerSize, 9) {
		t.Fatalf("flier markersize = %v, want 9", b.FlierMarkerSize)
	}
	if got := b.MedianColor; got.R != 1 || got.G != 0 || got.B != 0 {
		t.Fatalf("median color = %+v, want red", got)
	}
}

func TestAxesBehaviorRCParams(t *testing.T) {
	src := "axes.axisbelow: True\naxes.xmargin: 0.2\naxes.ymargin: 0\naxes.autolimit_mode: round_numbers\naxes.unicode_minus: False\n"
	theme, report, err := ParseMPLStyle("axesbehavior", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	a := theme.RC.Axes
	if a.AxisBelow != "True" {
		t.Fatalf("axisbelow = %q, want True", a.AxisBelow)
	}
	if a.XMargin != 0.2 || a.YMargin != 0 {
		t.Fatalf("margins = %v/%v, want 0.2/0", a.XMargin, a.YMargin)
	}
	if a.AutolimitMode != "round_numbers" {
		t.Fatalf("autolimit_mode = %q, want round_numbers", a.AutolimitMode)
	}
	if a.UnicodeMinus {
		t.Fatal("unicode_minus = true, want false")
	}
}

func TestAxesBehaviorRCParamsDefaults(t *testing.T) {
	a := Default.Axes
	if a.AxisBelow != "line" || a.XMargin != 0.05 || a.YMargin != 0.05 ||
		a.AutolimitMode != "data" || !a.UnicodeMinus {
		t.Fatalf("Default.Axes = %+v, want matplotlib defaults (line/0.05/0.05/data/true)", a)
	}
}

func TestAxesBehaviorRCParamsRejectInvalid(t *testing.T) {
	for _, src := range []string{
		"axes.axisbelow: sometimes\n",
		"axes.xmargin: -0.6\n",
		"axes.autolimit_mode: nearest\n",
	} {
		if _, _, err := ParseMPLStyle("bad", src); err == nil {
			t.Errorf("ParseMPLStyle(%q) succeeded, want error", src)
		}
	}
}

func TestLinesRCParams(t *testing.T) {
	src := "lines.linestyle: dashed\nlines.marker: o\nlines.markersize: 10\nlines.markeredgewidth: 2.5\n"
	theme, report, err := ParseMPLStyle("lines", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	l := theme.RC.Lines
	if l.LineStyle != "--" {
		t.Fatalf("linestyle = %q, want -- (canonical form of dashed)", l.LineStyle)
	}
	if l.Marker != "o" {
		t.Fatalf("marker = %q, want o", l.Marker)
	}
	if l.MarkerSize != 10 || l.MarkerEdgeWidth != 2.5 {
		t.Fatalf("marker size/edge = %v/%v, want 10/2.5", l.MarkerSize, l.MarkerEdgeWidth)
	}
}

func TestLinesRCParamsDefaults(t *testing.T) {
	l := Default.Lines
	if l.LineStyle != "-" || l.Marker != "None" || l.MarkerSize != 6 || l.MarkerEdgeWidth != 1 {
		t.Fatalf("Default.Lines = %+v, want matplotlib defaults (-/None/6/1)", l)
	}
}

func TestLinesRCParamsRejectInvalid(t *testing.T) {
	for _, src := range []string{
		"lines.linestyle: wavy\n",
		"lines.markersize: -1\n",
		"lines.markeredgewidth: -0.5\n",
	} {
		if _, _, err := ParseMPLStyle("bad", src); err == nil {
			t.Errorf("ParseMPLStyle(%q) succeeded, want error", src)
		}
	}
}

func TestArtistDefaultRCParams(t *testing.T) {
	src := "scatter.marker: s\nscatter.edgecolors: black\nerrorbar.capsize: 3\n"
	theme, report, err := ParseMPLStyle("artists", src)
	if err != nil {
		t.Fatalf("ParseMPLStyle() error = %v", err)
	}
	if len(report.Unsupported) != 0 {
		t.Fatalf("unexpected unsupported: %+v", report.Unsupported)
	}
	if theme.RC.Scatter.Marker != "s" {
		t.Fatalf("scatter.marker = %q, want s", theme.RC.Scatter.Marker)
	}
	if theme.RC.Scatter.EdgeColors != "black" {
		t.Fatalf("scatter.edgecolors = %q, want black", theme.RC.Scatter.EdgeColors)
	}
	if theme.RC.Errorbar.CapSize != 3 {
		t.Fatalf("errorbar.capsize = %v, want 3", theme.RC.Errorbar.CapSize)
	}
}

func TestArtistDefaultRCParamsDefaults(t *testing.T) {
	if Default.Scatter.Marker != "o" || Default.Scatter.EdgeColors != "face" || Default.Errorbar.CapSize != 0 {
		t.Fatalf("artist defaults = %+v/%+v, want o/face/0", Default.Scatter, Default.Errorbar)
	}
}

func TestArtistDefaultRCParamsRejectInvalid(t *testing.T) {
	for _, src := range []string{
		"scatter.edgecolors: notacolor\n",
		"errorbar.capsize: -1\n",
	} {
		if _, _, err := ParseMPLStyle("bad", src); err == nil {
			t.Errorf("ParseMPLStyle(%q) succeeded, want error", src)
		}
	}
}
