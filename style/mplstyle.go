package style

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/cycler"
	"github.com/cwbudde/matplotlib-go/render"
)

// MPLStyleIssue captures one ignored or unsupported rcParam entry.
type MPLStyleIssue struct {
	Line  int
	Key   string
	Value string
}

// MPLStyleReport describes how a .mplstyle file was applied.
type MPLStyleReport struct {
	Applied     []string
	Unsupported []MPLStyleIssue
}

var supportedMPLStyleKeys = []string{
	"agg.path.chunksize",
	"animation.bitrate",
	"animation.codec",
	"animation.convert_args",
	"animation.convert_path",
	"animation.embed_limit",
	"animation.ffmpeg_args",
	"animation.ffmpeg_path",
	"animation.frame_format",
	"animation.html",
	"animation.writer",
	"axes.autolimit_mode",
	"axes.axisbelow",
	"axes.edgecolor",
	"axes.facecolor",
	"axes.formatter.limits",
	"axes.formatter.min_exponent",
	"axes.formatter.offset_threshold",
	"axes.formatter.use_locale",
	"axes.formatter.use_mathtext",
	"axes.formatter.useoffset",
	"axes.grid",
	"axes.grid.axis",
	"axes.grid.which",
	"axes.labelcolor",
	"axes.labelpad",
	"axes.labelsize",
	"axes.labelweight",
	"axes.linewidth",
	"axes.prop_cycle",
	"axes.spines.bottom",
	"axes.spines.left",
	"axes.spines.right",
	"axes.spines.top",
	"axes.titlecolor",
	"axes.titlelocation",
	"axes.titlepad",
	"axes.titlesize",
	"axes.titleweight",
	"axes.titley",
	"axes.unicode_minus",
	"axes.xmargin",
	"axes.ymargin",
	"boxplot.boxprops.linewidth",
	"boxplot.capprops.linewidth",
	"boxplot.flierprops.color",
	"boxplot.flierprops.markeredgewidth",
	"boxplot.flierprops.markersize",
	"boxplot.meanline",
	"boxplot.meanprops.color",
	"boxplot.medianprops.color",
	"boxplot.medianprops.linewidth",
	"boxplot.notch",
	"boxplot.patchartist",
	"boxplot.showbox",
	"boxplot.showcaps",
	"boxplot.showfliers",
	"boxplot.showmeans",
	"boxplot.vertical",
	"boxplot.whiskerprops.linewidth",
	"boxplot.whiskers",
	"contour.algorithm",
	"contour.corner_mask",
	"contour.linewidth",
	"contour.negative_linestyle",
	"date.autoformatter.day",
	"date.autoformatter.hour",
	"date.autoformatter.microsecond",
	"date.autoformatter.minute",
	"date.autoformatter.month",
	"date.autoformatter.second",
	"date.autoformatter.year",
	"date.converter",
	"date.epoch",
	"date.interval_multiples",
	"errorbar.capsize",
	"figure.dpi",
	"figure.edgecolor",
	"figure.facecolor",
	"figure.figsize",
	"figure.frameon",
	"figure.autolayout",
	"figure.constrained_layout.use",
	"figure.constrained_layout.h_pad",
	"figure.constrained_layout.w_pad",
	"figure.constrained_layout.hspace",
	"figure.constrained_layout.wspace",
	"figure.subplot.left",
	"figure.subplot.right",
	"figure.subplot.bottom",
	"figure.subplot.top",
	"figure.subplot.wspace",
	"figure.subplot.hspace",
	"figure.titlesize",
	"figure.titleweight",
	"figure.labelsize",
	"figure.labelweight",
	"font.cursive",
	"font.family",
	"font.fantasy",
	"font.monospace",
	"font.sans-serif",
	"font.serif",
	"font.size",
	"font.stretch",
	"font.style",
	"font.variant",
	"font.weight",
	"grid.alpha",
	"grid.color",
	"grid.linewidth",
	"grid.linestyle",
	"grid.major.color",
	"grid.major.linestyle",
	"grid.minor.color",
	"grid.minor.linestyle",
	"hatch.color",
	"hatch.linewidth",
	"hist.bins",
	"image.aspect",
	"image.cmap",
	"image.composite_image",
	"image.interpolation",
	"image.interpolation_stage",
	"image.lut",
	"image.origin",
	"image.resample",
	"legend.borderaxespad",
	"legend.borderpad",
	"legend.columnspacing",
	"legend.edgecolor",
	"legend.facecolor",
	"legend.fancybox",
	"legend.framealpha",
	"legend.frameon",
	"legend.fontsize",
	"legend.handleheight",
	"legend.handlelength",
	"legend.handletextpad",
	"legend.labelcolor",
	"legend.labelspacing",
	"legend.loc",
	"legend.markerscale",
	"legend.numpoints",
	"legend.scatterpoints",
	"legend.shadow",
	"legend.title_fontsize",
	"lines.antialiased",
	"lines.color",
	"lines.dash_capstyle",
	"lines.dash_joinstyle",
	"lines.dashdot_pattern",
	"lines.dashed_pattern",
	"lines.dotted_pattern",
	"lines.linestyle",
	"lines.linewidth",
	"lines.marker",
	"lines.markeredgecolor",
	"lines.markeredgewidth",
	"lines.markerfacecolor",
	"lines.markersize",
	"lines.scale_dashes",
	"lines.solid_capstyle",
	"lines.solid_joinstyle",
	"markers.fillstyle",
	"patch.antialiased",
	"patch.edgecolor",
	"patch.facecolor",
	"patch.force_edgecolor",
	"patch.linewidth",
	"mathtext.bf",
	"mathtext.bfit",
	"mathtext.cal",
	"mathtext.default",
	"mathtext.fallback",
	"mathtext.fontset",
	"mathtext.it",
	"mathtext.rm",
	"mathtext.sf",
	"mathtext.tt",
	"path.simplify",
	"path.simplify_threshold",
	"path.sketch",
	"pdf.compression",
	"pdf.fonttype",
	"pdf.inheritcolor",
	"pdf.use14corefonts",
	"ps.distiller.res",
	"ps.fonttype",
	"ps.papersize",
	"ps.usedistiller",
	"ps.useafm",
	"savefig.bbox",
	"savefig.dpi",
	"savefig.edgecolor",
	"savefig.facecolor",
	"savefig.format",
	"savefig.pad_inches",
	"savefig.transparent",
	"scatter.edgecolors",
	"scatter.marker",
	"svg.fonttype",
	"svg.hashsalt",
	"svg.id",
	"svg.image_inline",
	"text.color",
	"text.usetex",
	"xtick.alignment",
	"xtick.bottom",
	"xtick.color",
	"xtick.direction",
	"xtick.labelbottom",
	"xtick.labelcolor",
	"xtick.labelsize",
	"xtick.labeltop",
	"xtick.major.bottom",
	"xtick.major.pad",
	"xtick.major.size",
	"xtick.major.top",
	"xtick.major.width",
	"xtick.minor.bottom",
	"xtick.minor.ndivs",
	"xtick.minor.pad",
	"xtick.minor.size",
	"xtick.minor.top",
	"xtick.minor.visible",
	"xtick.minor.width",
	"xtick.top",
	"ytick.alignment",
	"ytick.color",
	"ytick.direction",
	"ytick.labelleft",
	"ytick.labelcolor",
	"ytick.labelright",
	"ytick.labelsize",
	"ytick.left",
	"ytick.major.left",
	"ytick.major.pad",
	"ytick.major.right",
	"ytick.major.size",
	"ytick.major.width",
	"ytick.minor.left",
	"ytick.minor.ndivs",
	"ytick.minor.pad",
	"ytick.minor.right",
	"ytick.minor.size",
	"ytick.minor.visible",
	"ytick.minor.width",
	"ytick.right",
}

type mplStyleState struct {
	rc RC

	fontSizeSet bool
	fontSet     bool

	figureFaceValue      string
	figureFaceSet        bool
	figureEdgeValue      string
	figureEdgeSet        bool
	figureTitleSizeValue string
	figureTitleSizeSet   bool
	figureLabelSizeValue string
	figureLabelSizeSet   bool
	figureWidth          float64
	figureHeight         float64
	figureSizeSet        bool
	textColorValue       string
	textColorSet         bool
	textUseTeX           bool
	textUseTeXSet        bool
	lineColorValue       string
	lineColorSet         bool

	axesFaceValue   string
	axesFaceSet     bool
	axesEdgeValue   string
	axesEdgeSet     bool
	titleColorValue string
	titleColorSet   bool
	labelColorValue string
	labelColorSet   bool
	xTickColorValue string
	xTickColorSet   bool
	yTickColorValue string
	yTickColorSet   bool

	lineWidthPt      float64
	lineWidthSet     bool
	histBins         int
	histBinsSet      bool
	axisLineWidthPt  float64
	axisLineWidthSet bool
	titleFontSize    float64
	titleFontSizeSet bool
	labelFontSize    float64
	labelFontSizeSet bool
	xTickFontSize    float64
	xTickFontSizeSet bool
	yTickFontSize    float64
	yTickFontSizeSet bool
	gridLineWidthPt  float64
	gridLineWidthSet bool

	gridColorValue     string
	gridColorSet       bool
	gridMajorValue     string
	gridMajorSet       bool
	gridMinorValue     string
	gridMinorSet       bool
	gridAlpha          float64
	gridAlphaSet       bool
	gridVisible        bool
	gridVisibleSet     bool
	gridAxis           string
	gridAxisSet        bool
	gridWhich          string
	gridWhichSet       bool
	gridDashes         []float64
	gridDashesSet      bool
	gridMajorDashes    []float64
	gridMajorDashesSet bool
	gridMinorDashes    []float64
	gridMinorDashesSet bool

	legendFaceValue     string
	legendFaceSet       bool
	legendEdgeValue     string
	legendEdgeSet       bool
	legendTextValue     string
	legendTextSet       bool
	legendFontSize      float64
	legendFontSet       bool
	legendFrameAlpha    float64
	legendFrameAlphaSet bool
	legendFrameOn       bool
	legendFrameOnSet    bool
	legendTitleFont     string
	legendTitleFontSet  bool
}

// SupportedMPLStyleKeys returns the subset of rcParams understood by the loader.
func SupportedMPLStyleKeys() []string {
	keys := make([]string, len(supportedMPLStyleKeys))
	copy(keys, supportedMPLStyleKeys)
	sort.Strings(keys)
	return keys
}

// ParseMPLStyle parses a Matplotlib .mplstyle payload into a theme.
//
// Only the rcParams returned by SupportedMPLStyleKeys are applied. Unknown keys
// are reported in the returned report and ignored.
func ParseMPLStyle(name, src string) (Theme, MPLStyleReport, error) {
	rc, report, err := parseMPLStyleRC(Default, src)
	if err != nil {
		return Theme{}, report, err
	}

	return Theme{
		Name: normalizeMPLStyleName(name),
		RC:   rc,
	}, report, nil
}

// LoadMPLStyleFile loads and parses a Matplotlib .mplstyle file.
func LoadMPLStyleFile(path string) (Theme, MPLStyleReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, MPLStyleReport{}, err
	}

	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return ParseMPLStyle(name, string(data))
}

func parseMPLStyleRC(base RC, src string) (RC, MPLStyleReport, error) {
	report := MPLStyleReport{}
	state := newMPLStyleState(base)

	lines := strings.Split(src, "\n")
	for i, rawLine := range lines {
		lineNo := i + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := splitMPLStyleLine(rawLine)
		if !ok {
			return RC{}, report, fmt.Errorf("parse .mplstyle line %d: expected key: value", lineNo)
		}

		normalizedKey := normalizeThemeName(key)
		if err := applyMPLStyleEntry(&state, normalizedKey, value, lineNo, &report); err != nil {
			return RC{}, report, err
		}
	}

	finalizeMPLStyleState(&state)
	return Apply(state.rc), report, nil
}

func applyMPLStyleParams(base RC, params Params) (RC, MPLStyleReport, error) {
	report := MPLStyleReport{}
	state := newMPLStyleState(base)

	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for i, key := range keys {
		normalizedKey := normalizeThemeName(key)
		if err := applyMPLStyleEntry(&state, normalizedKey, params[key], i+1, &report); err != nil {
			return RC{}, report, err
		}
	}

	finalizeMPLStyleState(&state)
	return Apply(state.rc), report, nil
}

func newMPLStyleState(base RC) mplStyleState {
	return mplStyleState{rc: Apply(base)}
}

func applyMPLStyleEntry(state *mplStyleState, key, value string, lineNo int, report *MPLStyleReport) error {
	if state == nil || report == nil {
		return errors.New("nil mplstyle state")
	}
	if handled, err := applyMPLTickStyleEntry(&state.rc, key, value); handled {
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		report.Applied = append(report.Applied, key)
		return nil
	}

	switch key {
	case "figure.dpi":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.DPI = parsed
	case "figure.facecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.figureFaceValue = normalizeMPLValue(value)
		state.figureFaceSet = true
	case "figure.edgecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.figureEdgeValue = normalizeMPLValue(value)
		state.figureEdgeSet = true
	case "figure.figsize":
		width, height, err := parseMPLFigureSize(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.figureWidth = width
		state.figureHeight = height
		state.figureSizeSet = true
	case "figure.frameon":
		if err := parseMPLBoolInto(value, &state.rc.Figure.FrameOn); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "figure.autolayout":
		if err := parseMPLBoolInto(value, &state.rc.Figure.AutoLayout); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "figure.constrained_layout.use":
		if err := parseMPLBoolInto(value, &state.rc.Figure.Constrained.Use); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "figure.constrained_layout.h_pad", "figure.constrained_layout.w_pad",
		"figure.constrained_layout.hspace", "figure.constrained_layout.wspace",
		"figure.subplot.left", "figure.subplot.right", "figure.subplot.bottom",
		"figure.subplot.top", "figure.subplot.wspace", "figure.subplot.hspace":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		switch key {
		case "figure.constrained_layout.h_pad":
			state.rc.Figure.Constrained.HPad = parsed
		case "figure.constrained_layout.w_pad":
			state.rc.Figure.Constrained.WPad = parsed
		case "figure.constrained_layout.hspace":
			state.rc.Figure.Constrained.HSpace = parsed
		case "figure.constrained_layout.wspace":
			state.rc.Figure.Constrained.WSpace = parsed
		case "figure.subplot.left":
			state.rc.Figure.Subplot.Left = parsed
		case "figure.subplot.right":
			state.rc.Figure.Subplot.Right = parsed
		case "figure.subplot.bottom":
			state.rc.Figure.Subplot.Bottom = parsed
		case "figure.subplot.top":
			state.rc.Figure.Subplot.Top = parsed
		case "figure.subplot.wspace":
			state.rc.Figure.Subplot.WSpace = parsed
		case "figure.subplot.hspace":
			state.rc.Figure.Subplot.HSpace = parsed
		}
	case "figure.titlesize", "figure.labelsize":
		if _, err := parseMPLFontSize(value, state.rc.FontSize); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if key == "figure.titlesize" {
			state.figureTitleSizeValue, state.figureTitleSizeSet = normalizeMPLValue(value), true
		} else {
			state.figureLabelSizeValue, state.figureLabelSizeSet = normalizeMPLValue(value), true
		}
	case "figure.titleweight", "figure.labelweight":
		parsed, err := parseMPLFontWeight(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if key == "figure.titleweight" {
			state.rc.Figure.TitleWeight = parsed
		} else {
			state.rc.Figure.LabelWeight = parsed
		}
	case "path.simplify":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PathSimplify = parsed
	case "path.simplify_threshold":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PathSimplifyThreshold = parsed
	case "agg.path.chunksize":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.AggPathChunkSize = parsed
	case "path.sketch":
		parsed, err := parseMPLSketch(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PathSketch = parsed
	case "font.family":
		parsed, err := parseMPLFontFamilyList(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Font.Family = parsed
		state.fontSet = true
	case "font.serif", "font.sans-serif", "font.cursive", "font.fantasy", "font.monospace":
		parsed, err := parseMPLFontFamilyList(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		switch key {
		case "font.serif":
			state.rc.Font.Serif = parsed
		case "font.sans-serif":
			state.rc.Font.SansSerif = parsed
		case "font.cursive":
			state.rc.Font.Cursive = parsed
		case "font.fantasy":
			state.rc.Font.Fantasy = parsed
		case "font.monospace":
			state.rc.Font.Monospace = parsed
		}
		state.fontSet = true
	case "font.style":
		parsed, err := parseMPLFontStyle(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Font.Style = parsed
		state.fontSet = true
	case "font.variant":
		parsed, err := parseMPLFontVariant(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Font.Variant = parsed
		state.fontSet = true
	case "font.weight":
		parsed, err := parseMPLFontWeight(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Font.Weight = parsed
		state.fontSet = true
	case "font.stretch":
		parsed, err := parseMPLFontStretch(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Font.Stretch = parsed
		state.fontSet = true
	case "font.size":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.FontSize = parsed
		state.fontSizeSet = true
	case "text.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.textColorValue = normalizeMPLValue(value)
		state.textColorSet = true
	case "text.usetex":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.textUseTeX = parsed
		state.textUseTeXSet = true
	case "lines.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.lineColorValue = normalizeMPLValue(value)
		state.lineColorSet = true
	case "lines.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.lineWidthPt = parsed
		state.lineWidthSet = true
	case "lines.linestyle":
		parsed, err := parseMPLLineStyleName(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.LineStyle = parsed
	case "lines.marker":
		marker := normalizeMPLValue(value)
		if marker == "" {
			marker = "None"
		}
		state.rc.Lines.Marker = marker
	case "lines.markersize":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if parsed < 0 {
			return fmt.Errorf("parse %s on line %d: marker size must be non-negative, got %v", key, lineNo, parsed)
		}
		state.rc.Lines.MarkerSize = parsed
	case "lines.markeredgewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if parsed < 0 {
			return fmt.Errorf("parse %s on line %d: marker edge width must be non-negative, got %v", key, lineNo, parsed)
		}
		state.rc.Lines.MarkerEdgeWidth = parsed
	case "lines.dashed_pattern":
		parsed, err := parseMPLFloatList(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.DashedPattern = parsed
	case "lines.dashdot_pattern":
		parsed, err := parseMPLFloatList(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.DashDotPattern = parsed
	case "lines.dotted_pattern":
		parsed, err := parseMPLFloatList(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.DottedPattern = parsed
	case "lines.scale_dashes":
		if err := parseMPLBoolInto(value, &state.rc.Lines.ScaleDashes); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "lines.dash_capstyle":
		parsed, err := parseMPLLineCap(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.DashCap = parsed
	case "lines.solid_capstyle":
		parsed, err := parseMPLLineCap(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.SolidCap = parsed
	case "lines.dash_joinstyle":
		parsed, err := parseMPLLineJoin(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.DashJoin = parsed
	case "lines.solid_joinstyle":
		parsed, err := parseMPLLineJoin(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.SolidJoin = parsed
	case "lines.markerfacecolor":
		parsed, err := parseMPLMarkerColor(value, &state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.MarkerFaceColor = parsed
	case "lines.markeredgecolor":
		parsed, err := parseMPLMarkerColor(value, &state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.MarkerEdgeColor = parsed
	case "markers.fillstyle":
		parsed, err := parseMPLMarkerFillStyle(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Lines.MarkerFillStyle = parsed
	case "lines.antialiased":
		if err := parseMPLBoolInto(value, &state.rc.Lines.Antialiased); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "patch.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if parsed < 0 {
			return fmt.Errorf("parse %s on line %d: linewidth must be non-negative, got %v", key, lineNo, parsed)
		}
		state.rc.Patch.LineWidth = parsed
	case "patch.facecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Patch.FaceColorRaw = normalizeMPLValue(value)
	case "patch.edgecolor":
		parsed, err := parseMPLColor(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Patch.EdgeColor = parsed
	case "patch.force_edgecolor":
		if err := parseMPLBoolInto(value, &state.rc.Patch.ForceEdgeColor); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "patch.antialiased":
		if err := parseMPLBoolInto(value, &state.rc.Patch.Antialiased); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "scatter.marker":
		marker := normalizeMPLValue(value)
		if marker == "" {
			return fmt.Errorf("parse %s on line %d: empty marker", key, lineNo)
		}
		state.rc.Scatter.Marker = marker
	case "scatter.edgecolors":
		normalized := normalizeMPLValue(value)
		if !strings.EqualFold(normalized, "face") && !strings.EqualFold(normalized, "none") {
			if err := validateMPLColorValue(value, state.rc, false); err != nil {
				return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
			}
		}
		state.rc.Scatter.EdgeColors = normalized
	case "errorbar.capsize":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if parsed < 0 {
			return fmt.Errorf("parse %s on line %d: cap size must be non-negative, got %v", key, lineNo, parsed)
		}
		state.rc.Errorbar.CapSize = parsed
	case "hist.bins":
		parsed, err := parseMPLHistBins(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.histBins = parsed
		state.histBinsSet = true
	case "axes.facecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axesFaceValue = normalizeMPLValue(value)
		state.axesFaceSet = true
	case "axes.edgecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axesEdgeValue = normalizeMPLValue(value)
		state.axesEdgeSet = true
	case "axes.labelcolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.labelColorValue = normalizeMPLValue(value)
		state.labelColorSet = true
	case "axes.labelpad":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.LabelPad = parsed
	case "axes.labelweight":
		parsed, err := parseMPLFontWeight(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.LabelWeight = parsed
	case "axes.spines.top", "axes.spines.bottom", "axes.spines.left", "axes.spines.right":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		switch key {
		case "axes.spines.top":
			state.rc.Axes.Spines.Top = parsed
		case "axes.spines.bottom":
			state.rc.Axes.Spines.Bottom = parsed
		case "axes.spines.left":
			state.rc.Axes.Spines.Left = parsed
		case "axes.spines.right":
			state.rc.Axes.Spines.Right = parsed
		}
	case "axes.titlecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.titleColorValue = normalizeMPLValue(value)
		state.titleColorSet = true
	case "axes.titlelocation":
		normalized := strings.ToLower(normalizeMPLValue(value))
		switch normalized {
		case "left", "center", "right":
			state.rc.Axes.TitleLocation = normalized
		default:
			return fmt.Errorf("parse %s on line %d: invalid title location %q", key, lineNo, value)
		}
	case "axes.titlepad":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.TitlePad = parsed
	case "axes.titleweight":
		parsed, err := parseMPLFontWeight(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.TitleWeight = parsed
	case "axes.titley":
		if strings.EqualFold(normalizeMPLValue(value), "none") {
			state.rc.Axes.TitleY = 0
			state.rc.Axes.TitleYSet = false
			break
		}
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.TitleY = parsed
		state.rc.Axes.TitleYSet = true
	case "xtick.color", "xtick.labelcolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.xTickColorValue = normalizeMPLValue(value)
		state.xTickColorSet = true
	case "ytick.color", "ytick.labelcolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.yTickColorValue = normalizeMPLValue(value)
		state.yTickColorSet = true
	case "axes.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.axisLineWidthPt = parsed
		state.axisLineWidthSet = true
	case "axes.titlesize":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.titleFontSize = parsed
		state.titleFontSizeSet = true
	case "axes.labelsize":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.labelFontSize = parsed
		state.labelFontSizeSet = true
	case "axes.prop_cycle":
		parsed, err := parseMPLPropCycle(value, &state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if palette := propCycleColors(parsed); len(palette) > 0 {
			state.rc.ColorCycle = palette
		}
		// Store the full cycle only when it carries more than color, so
		// color-only sheets keep PropCycle nil (the historical behavior) and
		// avoid forcing non-color consumers through the prop-cycle path.
		if isColorOnlyCycle(parsed) {
			state.rc.PropCycle = nil
		} else {
			state.rc.PropCycle = parsed
		}
	case "axes.axisbelow":
		normalized := strings.ToLower(normalizeMPLValue(value))
		if normalized == "line" {
			state.rc.Axes.AxisBelow = "line"
		} else {
			parsed, err := parseMPLBool(value)
			if err != nil {
				return fmt.Errorf("parse %s on line %d: expected bool or 'line': %w", key, lineNo, err)
			}
			state.rc.Axes.AxisBelow = formatMPLBool(parsed)
		}
	case "axes.xmargin", "axes.ymargin":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		if parsed <= -0.5 {
			return fmt.Errorf("parse %s on line %d: margin must be greater than -0.5, got %v", key, lineNo, parsed)
		}
		if key == "axes.xmargin" {
			state.rc.Axes.XMargin = parsed
		} else {
			state.rc.Axes.YMargin = parsed
		}
	case "axes.autolimit_mode":
		normalized := strings.ToLower(normalizeMPLValue(value))
		if normalized != "data" && normalized != "round_numbers" {
			return fmt.Errorf("parse %s on line %d: expected 'data' or 'round_numbers', got %q", key, lineNo, value)
		}
		state.rc.Axes.AutolimitMode = normalized
	case "axes.unicode_minus":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.UnicodeMinus = parsed
	case "contour.algorithm":
		normalized := strings.ToLower(normalizeMPLValue(value))
		switch normalized {
		case "mpl2005", "mpl2014", "serial", "threaded":
			state.rc.Contour.Algorithm = normalized
		default:
			return fmt.Errorf("parse %s on line %d: invalid contour algorithm %q", key, lineNo, value)
		}
	case "contour.corner_mask":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Contour.CornerMask = parsed
	case "contour.linewidth":
		if strings.EqualFold(normalizeMPLValue(value), "none") {
			state.rc.Contour.LineWidth = 0
			state.rc.Contour.LineWidthSet = false
			break
		}
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Contour.LineWidth = parsed
		state.rc.Contour.LineWidthSet = true
	case "contour.negative_linestyle":
		normalized := strings.ToLower(normalizeMPLValue(value))
		switch normalized {
		case "-", "solid", "--", "dashed", "-.", "dashdot", ":", "dotted":
			state.rc.Contour.NegativeLineStyle = normalized
		default:
			return fmt.Errorf("parse %s on line %d: invalid contour linestyle %q", key, lineNo, value)
		}
	case "axes.formatter.limits":
		parsed, err := parseMPLIntPair(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.Formatter.Limits = parsed
	case "axes.formatter.min_exponent":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.Formatter.MinExponent = parsed
	case "axes.formatter.offset_threshold":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Axes.Formatter.OffsetThreshold = parsed
	case "axes.formatter.use_locale":
		if err := parseMPLBoolInto(value, &state.rc.Axes.Formatter.UseLocale); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "axes.formatter.use_mathtext":
		if err := parseMPLBoolInto(value, &state.rc.Axes.Formatter.UseMathText); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "axes.formatter.useoffset":
		if err := parseMPLBoolInto(value, &state.rc.Axes.Formatter.UseOffset); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "axes.grid":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridVisible = parsed
		state.gridVisibleSet = true
	case "axes.grid.axis":
		parsed, err := parseMPLGridAxis(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridAxis = parsed
		state.gridAxisSet = true
	case "axes.grid.which":
		parsed, err := parseMPLGridWhich(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridWhich = parsed
		state.gridWhichSet = true
	case "grid.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridColorValue = normalizeMPLValue(value)
		state.gridColorSet = true
	case "grid.major.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMajorValue = normalizeMPLValue(value)
		state.gridMajorSet = true
	case "grid.minor.color":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMinorValue = normalizeMPLValue(value)
		state.gridMinorSet = true
	case "grid.alpha":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridAlpha = parsed
		state.gridAlphaSet = true
	case "grid.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridLineWidthPt = parsed
		state.gridLineWidthSet = true
	case "grid.linestyle":
		parsed, err := parseMPLLineStyle(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridDashes = parsed
		state.gridDashesSet = true
	case "grid.major.linestyle":
		parsed, err := parseMPLLineStyle(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMajorDashes = parsed
		state.gridMajorDashesSet = true
	case "grid.minor.linestyle":
		parsed, err := parseMPLLineStyle(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.gridMinorDashes = parsed
		state.gridMinorDashesSet = true
	case "legend.facecolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendFaceValue = normalizeMPLValue(value)
		state.legendFaceSet = true
	case "legend.edgecolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendEdgeValue = normalizeMPLValue(value)
		state.legendEdgeSet = true
	case "legend.labelcolor":
		if err := validateMPLColorValue(value, state.rc, true); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendTextValue = normalizeMPLValue(value)
		state.legendTextSet = true
	case "legend.fontsize":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendFontSize = parsed
		state.legendFontSet = true
	case "legend.framealpha":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendFrameAlpha = parsed
		state.legendFrameAlphaSet = true
	case "legend.frameon":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendFrameOn = parsed
		state.legendFrameOnSet = true
	case "legend.loc":
		parsed, err := parseMPLLegendLocation(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Legend.Location = parsed
	case "legend.fancybox":
		if err := parseMPLBoolInto(value, &state.rc.Legend.FancyBox); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.shadow":
		if err := parseMPLBoolInto(value, &state.rc.Legend.Shadow); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.numpoints":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Legend.NumPoints = parsed
	case "legend.scatterpoints":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Legend.ScatterPoints = parsed
	case "legend.markerscale":
		if err := parseMPLFloatInto(value, &state.rc.Legend.MarkerScale); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.title_fontsize":
		if _, err := parseMPLFontSizeOrNone(value, state.rc.FontSize); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.legendTitleFont = normalizeMPLValue(value)
		state.legendTitleFontSet = true
	case "legend.borderpad":
		if err := parseMPLFloatInto(value, &state.rc.Legend.BorderPad); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.labelspacing":
		if err := parseMPLFloatInto(value, &state.rc.Legend.LabelSpacing); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.handlelength":
		if err := parseMPLFloatInto(value, &state.rc.Legend.HandleLength); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.handleheight":
		if err := parseMPLFloatInto(value, &state.rc.Legend.HandleHeight); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.handletextpad":
		if err := parseMPLFloatInto(value, &state.rc.Legend.HandleTextPad); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.borderaxespad":
		if err := parseMPLFloatInto(value, &state.rc.Legend.BorderAxesPad); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "legend.columnspacing":
		if err := parseMPLFloatInto(value, &state.rc.Legend.ColumnSpacing); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
	case "xtick.labelsize":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.xTickFontSize = parsed
		state.xTickFontSizeSet = true
	case "ytick.labelsize":
		parsed, err := parseMPLFontSize(value, state.rc.FontSize)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.yTickFontSize = parsed
		state.yTickFontSizeSet = true
	case "hatch.color":
		parsed, err := parseMPLColor(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Hatch.Color = parsed
	case "hatch.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Hatch.LineWidth = parsed
	case "image.cmap":
		normalized := normalizeMPLValue(value)
		if normalized == "" {
			return fmt.Errorf("parse %s on line %d: empty colormap", key, lineNo)
		}
		state.rc.Image.Cmap = normalized
	case "image.interpolation":
		state.rc.Image.Interpolation = normalizeMPLValue(value)
	case "image.interpolation_stage":
		parsed, err := parseMPLEnum(value, "auto", "data", "rgba")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.InterpolationStage = parsed
	case "image.origin":
		parsed, err := parseMPLEnum(value, "upper", "lower")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.Origin = parsed
	case "image.aspect":
		parsed, err := parseMPLAspect(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.Aspect = parsed
	case "image.resample":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.Resample = parsed
	case "image.composite_image":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.CompositeImage = parsed
	case "image.lut":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Image.LUT = parsed
	case "mathtext.fontset":
		parsed, err := parseMPLEnum(value, "dejavusans", "dejavuserif", "cm", "stix", "stixsans", "custom")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Mathtext.Fontset = parsed
	case "mathtext.default":
		parsed, err := parseMPLEnum(value, "rm", "cal", "bfit", "it", "tt", "sf", "bf", "default", "bb", "frak", "scr", "regular")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Mathtext.Default = parsed
	case "mathtext.fallback":
		parsed, err := parseMPLMathtextFallback(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Mathtext.Fallback = parsed
	case "mathtext.bf":
		state.rc.Mathtext.BF = normalizeMPLValue(value)
	case "mathtext.bfit":
		state.rc.Mathtext.BFit = normalizeMPLValue(value)
	case "mathtext.cal":
		state.rc.Mathtext.Cal = normalizeMPLValue(value)
	case "mathtext.it":
		state.rc.Mathtext.It = normalizeMPLValue(value)
	case "mathtext.rm":
		state.rc.Mathtext.RM = normalizeMPLValue(value)
	case "mathtext.sf":
		state.rc.Mathtext.SF = normalizeMPLValue(value)
	case "mathtext.tt":
		state.rc.Mathtext.TT = normalizeMPLValue(value)
	case "boxplot.notch":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.Notch = parsed
	case "boxplot.vertical":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.Vertical = parsed
	case "boxplot.patchartist":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.PatchArtist = parsed
	case "boxplot.meanline":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.MeanLine = parsed
	case "boxplot.showmeans":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.ShowMeans = parsed
	case "boxplot.showcaps":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.ShowCaps = parsed
	case "boxplot.showbox":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.ShowBox = parsed
	case "boxplot.showfliers":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.ShowFliers = parsed
	case "boxplot.whiskers":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.Whiskers = parsed
	case "boxplot.boxprops.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.BoxLineWidth = parsed
	case "boxplot.whiskerprops.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.WhiskerLineWidth = parsed
	case "boxplot.capprops.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.CapLineWidth = parsed
	case "boxplot.medianprops.linewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.MedianLineWidth = parsed
	case "boxplot.medianprops.color":
		parsed, err := parseMPLColor(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.MedianColor = parsed
	case "boxplot.meanprops.color":
		parsed, err := parseMPLColor(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.MeanColor = parsed
	case "boxplot.flierprops.color":
		parsed, err := parseMPLColor(value, state.rc)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.FlierColor = parsed
	case "boxplot.flierprops.markersize":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.FlierMarkerSize = parsed
	case "boxplot.flierprops.markeredgewidth":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Boxplot.FlierEdgeWidth = parsed
	case "date.autoformatter.year":
		state.rc.Date.AutoYear = normalizeMPLValue(value)
	case "date.autoformatter.month":
		state.rc.Date.AutoMonth = normalizeMPLValue(value)
	case "date.autoformatter.day":
		state.rc.Date.AutoDay = normalizeMPLValue(value)
	case "date.autoformatter.hour":
		state.rc.Date.AutoHour = normalizeMPLValue(value)
	case "date.autoformatter.minute":
		state.rc.Date.AutoMinute = normalizeMPLValue(value)
	case "date.autoformatter.second":
		state.rc.Date.AutoSecond = normalizeMPLValue(value)
	case "date.autoformatter.microsecond":
		state.rc.Date.AutoMicrosecond = normalizeMPLValue(value)
	case "date.epoch":
		state.rc.Date.Epoch = normalizeMPLValue(value)
	case "date.converter":
		parsed, err := parseMPLEnum(value, "auto", "concise")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Date.Converter = parsed
	case "date.interval_multiples":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Date.IntervalMultiples = parsed
	case "pdf.fonttype":
		parsed, err := parseMPLFontType(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PDF.FontType = parsed
	case "pdf.use14corefonts":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PDF.Use14CoreFonts = parsed
	case "pdf.compression":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PDF.Compression = parsed
	case "pdf.inheritcolor":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PDF.InheritColor = parsed
	case "ps.fonttype":
		parsed, err := parseMPLFontType(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PS.FontType = parsed
	case "ps.useafm":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PS.UseAFM = parsed
	case "ps.papersize":
		state.rc.PS.PaperSize = normalizeMPLValue(value)
	case "ps.usedistiller":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PS.UseDistiller = parsed
	case "ps.distiller.res":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.PS.DistillerRes = parsed
	case "svg.fonttype":
		parsed, err := parseMPLEnum(value, "none", "path")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.SVG.FontType = parsed
	case "svg.image_inline":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.SVG.ImageInline = parsed
	case "svg.hashsalt":
		state.rc.SVG.HashSalt = parseMPLStringOrNone(value)
	case "svg.id":
		state.rc.SVG.ID = parseMPLStringOrNone(value)
	case "animation.html":
		parsed, err := parseMPLEnum(value, "html5", "jshtml", "none")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Animation.HTML = parsed
	case "animation.writer":
		state.rc.Animation.Writer = normalizeMPLValue(value)
	case "animation.codec":
		state.rc.Animation.Codec = normalizeMPLValue(value)
	case "animation.bitrate":
		parsed, err := parseMPLInt(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Animation.Bitrate = parsed
	case "animation.frame_format":
		parsed, err := parseMPLEnum(value, "png", "jpeg", "tiff", "raw", "rgba", "ppm", "sgi", "bmp", "pbm", "svg")
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Animation.FrameFormat = parsed
	case "animation.ffmpeg_path":
		state.rc.Animation.FFmpegPath = normalizeMPLValue(value)
	case "animation.ffmpeg_args":
		state.rc.Animation.FFmpegArgs = parseMPLStringList(value)
	case "animation.convert_path":
		state.rc.Animation.ConvertPath = normalizeMPLValue(value)
	case "animation.convert_args":
		state.rc.Animation.ConvertArgs = parseMPLStringList(value)
	case "animation.embed_limit":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Animation.EmbedLimit = parsed
	case "savefig.dpi":
		parsed, err := parseMPLSavefigDPI(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.Dpi = parsed
	case "savefig.facecolor":
		if err := validateMPLSavefigColor(value, &state.rc); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.Facecolor = normalizeMPLValue(value)
	case "savefig.edgecolor":
		if err := validateMPLSavefigColor(value, &state.rc); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.Edgecolor = normalizeMPLValue(value)
	case "savefig.transparent":
		parsed, err := parseMPLBool(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.Transparent = parsed
	case "savefig.bbox":
		parsed, err := parseMPLSavefigBbox(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.BboxInches = parsed
	case "savefig.pad_inches":
		parsed, err := parseMPLFloat(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.Savefig.PadInches = parsed
	case "savefig.format":
		state.rc.Savefig.Format = strings.ToLower(normalizeMPLValue(value))
	default:
		maybeWarnUnparsedRCParam(key)
		report.Unsupported = append(report.Unsupported, MPLStyleIssue{
			Line:  lineNo,
			Key:   key,
			Value: strings.TrimSpace(value),
		})
		return nil
	}

	report.Applied = append(report.Applied, key)
	maybeWarnUnhonoredRCParam(key, &state.rc)
	return nil
}

func applyMPLTickStyleEntry(rc *RC, key, value string) (bool, error) {
	if rc == nil {
		return false, nil
	}
	var axis *TickAxisRC
	var suffix string
	switch {
	case strings.HasPrefix(key, "xtick."):
		axis = &rc.XTick
		suffix = strings.TrimPrefix(key, "xtick.")
	case strings.HasPrefix(key, "ytick."):
		axis = &rc.YTick
		suffix = strings.TrimPrefix(key, "ytick.")
	default:
		return false, nil
	}

	switch suffix {
	case "direction":
		parsed, err := parseMPLEnum(value, "in", "inout", "out")
		if err == nil {
			axis.Direction = parsed
		}
		return true, err
	case "alignment":
		values := []string{"center", "left", "right"}
		if axis == &rc.YTick {
			values = []string{"baseline", "bottom", "center", "center_baseline", "top"}
		}
		parsed, err := parseMPLEnum(value, values...)
		if err == nil {
			axis.Alignment = parsed
		}
		return true, err
	case "bottom", "left":
		return true, parseMPLBoolInto(value, &axis.Primary)
	case "top", "right":
		return true, parseMPLBoolInto(value, &axis.Secondary)
	case "labelbottom", "labelleft":
		return true, parseMPLBoolInto(value, &axis.LabelPrimary)
	case "labeltop", "labelright":
		return true, parseMPLBoolInto(value, &axis.LabelSecondary)
	case "major.bottom", "major.left":
		return true, parseMPLBoolInto(value, &axis.Major.Primary)
	case "major.top", "major.right":
		return true, parseMPLBoolInto(value, &axis.Major.Secondary)
	case "minor.bottom", "minor.left":
		return true, parseMPLBoolInto(value, &axis.Minor.Primary)
	case "minor.top", "minor.right":
		return true, parseMPLBoolInto(value, &axis.Minor.Secondary)
	case "minor.visible":
		return true, parseMPLBoolInto(value, &axis.Minor.Visible)
	case "major.size":
		return true, parseMPLFloatInto(value, &axis.Major.Size)
	case "major.width":
		return true, parseMPLFloatInto(value, &axis.Major.Width)
	case "major.pad":
		return true, parseMPLFloatInto(value, &axis.Major.Pad)
	case "minor.size":
		return true, parseMPLFloatInto(value, &axis.Minor.Size)
	case "minor.width":
		return true, parseMPLFloatInto(value, &axis.Minor.Width)
	case "minor.pad":
		return true, parseMPLFloatInto(value, &axis.Minor.Pad)
	case "minor.ndivs":
		parsed, err := parseMPLMinorTickNDivs(value)
		if err == nil {
			axis.Minor.NDivs = parsed
		}
		return true, err
	default:
		return false, nil
	}
}

func parseMPLBoolInto(value string, target *bool) error {
	parsed, err := parseMPLBool(value)
	if err != nil {
		return err
	}
	*target = parsed
	return nil
}

func parseMPLFloatInto(value string, target *float64) error {
	parsed, err := parseMPLFloat(value)
	if err != nil {
		return err
	}
	*target = parsed
	return nil
}

func parseMPLMinorTickNDivs(value string) (int, error) {
	if strings.EqualFold(normalizeMPLValue(value), "auto") {
		return 0, nil
	}
	parsed, err := parseMPLInt(value)
	if err != nil {
		return 0, errors.New(`expected "auto" or a non-negative integer`)
	}
	if parsed < 0 {
		return 0, errors.New(`expected "auto" or a non-negative integer`)
	}
	return parsed, nil
}

func finalizeMPLStyleState(state *mplStyleState) {
	if state == nil {
		return
	}

	if state.figureFaceSet {
		if parsed, err := parseMPLColor(state.figureFaceValue, state.rc); err == nil {
			state.rc.Background = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
		}
	}
	if state.rc.Patch.FaceColorRaw != "" {
		if parsed, err := parseMPLColor(state.rc.Patch.FaceColorRaw, state.rc); err == nil {
			state.rc.Patch.FaceColor = parsed
		}
	}
	if state.figureEdgeSet {
		if parsed, err := parseMPLColor(state.figureEdgeValue, state.rc); err == nil {
			state.rc.Figure.EdgeColor = parsed
		}
	}
	if state.figureSizeSet {
		state.rc.FigureWidth = state.figureWidth
		state.rc.FigureHeight = state.figureHeight
	}
	if state.fontSet {
		state.rc.FontKey = fontKeyFromRC(&state.rc.Font)
	}
	if state.textColorSet {
		if parsed, err := parseMPLColor(state.textColorValue, state.rc); err == nil {
			state.rc.TextColor = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
			if !state.titleColorSet {
				state.rc.AxesTitleColor = parsed
			}
			if !state.labelColorSet {
				state.rc.AxesLabelColor = parsed
			}
			if !state.legendTextSet {
				state.rc.LegendTextColor = parsed
			}
		}
	}
	if state.textUseTeXSet {
		state.rc.UseTeX = state.textUseTeX
	}
	if state.lineColorSet {
		if parsed, err := parseMPLColor(state.lineColorValue, state.rc); err == nil {
			state.rc.LineColor = [4]float64{parsed.R, parsed.G, parsed.B, parsed.A}
		}
	}
	if state.axesFaceSet {
		if parsed, err := parseMPLColor(state.axesFaceValue, state.rc); err == nil {
			state.rc.AxesBackground = parsed
		}
	}
	if state.axesEdgeSet {
		if parsed, err := parseMPLColor(state.axesEdgeValue, state.rc); err == nil {
			state.rc.AxesEdgeColor = parsed
			if !state.xTickColorSet {
				state.rc.XTickColor = parsed
			}
			if !state.yTickColorSet {
				state.rc.YTickColor = parsed
			}
		}
	}
	if state.titleColorSet {
		if parsed, err := parseMPLColor(state.titleColorValue, state.rc); err == nil {
			state.rc.AxesTitleColor = parsed
		}
	}
	if state.labelColorSet {
		if parsed, err := parseMPLColor(state.labelColorValue, state.rc); err == nil {
			state.rc.AxesLabelColor = parsed
		}
	}
	if state.xTickColorSet {
		if parsed, err := parseMPLColor(state.xTickColorValue, state.rc); err == nil {
			state.rc.XTickColor = parsed
		}
	}
	if state.yTickColorSet {
		if parsed, err := parseMPLColor(state.yTickColorValue, state.rc); err == nil {
			state.rc.YTickColor = parsed
		}
	}

	if state.lineWidthSet {
		state.rc.LineWidth = state.lineWidthPt
	}
	if state.histBinsSet {
		state.rc.HistBins = state.histBins
	}
	if state.axisLineWidthSet {
		state.rc.AxisLineWidth = state.axisLineWidthPt
	}
	if state.figureTitleSizeSet {
		if parsed, err := parseMPLFontSize(state.figureTitleSizeValue, state.rc.FontSize); err == nil {
			state.rc.Figure.TitleSize = parsed
		}
	} else if state.fontSizeSet {
		state.rc.Figure.TitleSize = state.rc.FontSize * 1.2
	}
	if state.figureLabelSizeSet {
		if parsed, err := parseMPLFontSize(state.figureLabelSizeValue, state.rc.FontSize); err == nil {
			state.rc.Figure.LabelSize = parsed
		}
	} else if state.fontSizeSet {
		state.rc.Figure.LabelSize = state.rc.FontSize * 1.2
	}
	if state.fontSizeSet || state.titleFontSizeSet {
		state.rc.TitleFontSize = maxFloat(8, state.rc.FontSize*1.2)
	}
	if state.titleFontSizeSet {
		state.rc.TitleFontSize = maxFloat(8, state.titleFontSize)
	}
	if state.fontSizeSet || state.labelFontSizeSet {
		state.rc.AxisLabelFontSize = maxFloat(8, state.rc.FontSize)
	}
	if state.labelFontSizeSet {
		state.rc.AxisLabelFontSize = maxFloat(8, state.labelFontSize)
	}
	if state.fontSizeSet || state.xTickFontSizeSet {
		state.rc.XTickLabelFontSize = maxFloat(8, state.rc.FontSize)
	}
	if state.xTickFontSizeSet {
		state.rc.XTickLabelFontSize = maxFloat(8, state.xTickFontSize)
	}
	if state.fontSizeSet || state.yTickFontSizeSet {
		state.rc.YTickLabelFontSize = maxFloat(8, state.rc.FontSize)
	}
	if state.yTickFontSizeSet {
		state.rc.YTickLabelFontSize = maxFloat(8, state.yTickFontSize)
	}
	if state.gridLineWidthSet {
		width := state.gridLineWidthPt
		state.rc.GridLineWidth = width
		state.rc.MinorGridLineWidth = width
	}

	major := state.rc.GridColor
	minor := state.rc.MinorGridColor
	if state.gridColorSet {
		if parsed, err := parseMPLColor(state.gridColorValue, state.rc); err == nil {
			major = parsed
			minor = parsed
		}
	}
	if state.gridMajorSet {
		if parsed, err := parseMPLColor(state.gridMajorValue, state.rc); err == nil {
			major = parsed
		}
	}
	if state.gridMinorSet {
		if parsed, err := parseMPLColor(state.gridMinorValue, state.rc); err == nil {
			minor = parsed
		}
	}
	if state.gridAlphaSet {
		major.A = state.gridAlpha
		minor.A = state.gridAlpha
	}
	state.rc.GridColor = major
	state.rc.MinorGridColor = minor
	if state.gridDashesSet {
		state.rc.GridDashes = cloneDashPattern(state.gridDashes)
	}
	if state.gridMajorDashesSet {
		state.rc.GridDashes = cloneDashPattern(state.gridMajorDashes)
	}
	if state.gridMinorDashesSet {
		state.rc.MinorGridDashes = cloneDashPattern(state.gridMinorDashes)
	} else if state.gridDashesSet {
		state.rc.MinorGridDashes = cloneDashPattern(state.gridDashes)
	}
	if state.gridVisibleSet {
		state.rc.GridVisible = state.gridVisible
	}
	if state.gridAxisSet {
		state.rc.GridAxis = state.gridAxis
	}
	if state.gridWhichSet {
		state.rc.GridWhich = state.gridWhich
	}

	if state.legendFaceSet {
		state.rc.LegendBackground = resolveMPLSpecialColor(state.legendFaceValue, state.rc, state.rc.AxesBackground)
	}
	if state.legendEdgeSet {
		state.rc.LegendBorderColor = resolveMPLSpecialColor(state.legendEdgeValue, state.rc, state.rc.AxesEdgeColor)
	}
	if state.legendTextSet {
		state.rc.LegendTextColor = resolveMPLSpecialColor(state.legendTextValue, state.rc, state.rc.DefaultTextColor())
	}
	if state.fontSizeSet || state.legendFontSet {
		state.rc.LegendFontSize = maxFloat(8, state.rc.FontSize)
	}
	if state.legendFontSet {
		state.rc.LegendFontSize = maxFloat(8, state.legendFontSize)
	}
	if state.legendFrameAlphaSet {
		state.rc.LegendFrameAlpha = state.legendFrameAlpha
		state.rc.LegendBackground.A = state.legendFrameAlpha
		state.rc.LegendBorderColor.A = state.legendFrameAlpha
	}
	if state.legendFrameOnSet {
		state.rc.LegendFrameOn = state.legendFrameOn
		if !state.legendFrameOn {
			state.rc.LegendBackground.A = 0
			state.rc.LegendBorderColor.A = 0
		}
	}
	if state.legendTitleFontSet {
		if parsed, err := parseMPLFontSizeOrNone(state.legendTitleFont, state.rc.FontSize); err == nil {
			state.rc.Legend.TitleFontSize = parsed
		}
	}
}

func splitMPLStyleLine(raw string) (string, string, bool) {
	noComment := stripMPLStyleComment(raw)
	if strings.TrimSpace(noComment) == "" {
		return "", "", false
	}

	idx := strings.Index(noComment, ":")
	if idx < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(noComment[:idx])
	value := strings.TrimSpace(noComment[idx+1:])
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

func stripMPLStyleComment(raw string) string {
	inQuote := rune(0)
	for i, r := range raw {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			inQuote = r
			continue
		}
		if r == '#' {
			if i == 0 {
				return ""
			}
			prev, _ := utf8LastRune(raw[:i])
			if unicode.IsSpace(prev) {
				return strings.TrimRightFunc(raw[:i], unicode.IsSpace)
			}
		}
	}
	return raw
}

func utf8LastRune(s string) (rune, int) {
	return utf8.DecodeLastRuneInString(s)
}

func parseMPLFloat(value string) (float64, error) {
	normalized := normalizeMPLValue(value)
	parsed, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %q", value)
	}
	return parsed, nil
}

func parseMPLInt(value string) (int, error) {
	normalized := normalizeMPLValue(value)
	parsed, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, fmt.Errorf("invalid int %q", value)
	}
	return parsed, nil
}

func parseMPLIntPair(value string) ([2]int, error) {
	var pair [2]int
	normalized := normalizeMPLValue(value)
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(normalized, "]"), "["))
	parts := splitOutsideQuotes(normalized, ',')
	if len(parts) != 2 {
		return pair, fmt.Errorf("expected two comma-separated integers, got %q", value)
	}
	for i, part := range parts {
		parsed, err := parseMPLInt(part)
		if err != nil {
			return pair, err
		}
		pair[i] = parsed
	}
	return pair, nil
}

func parseMPLFontWeight(value string) (int, error) {
	normalized := normalizeMPLValue(value)
	weights := map[string]int{
		"ultralight": 100,
		"light":      200,
		"lighter":    300,
		"normal":     400,
		"regular":    400,
		"book":       400,
		"medium":     500,
		"roman":      500,
		"semibold":   600,
		"demibold":   600,
		"demi":       600,
		"bold":       700,
		"bolder":     700,
		"heavy":      800,
		"extra bold": 800,
		"black":      900,
	}
	if weight, ok := weights[normalized]; ok {
		return weight, nil
	}
	weight, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, fmt.Errorf("invalid font weight %q", value)
	}
	if weight < 1 || weight > 1000 {
		return 0, fmt.Errorf("font weight %d outside [1, 1000]", weight)
	}
	return weight, nil
}

func parseMPLFontStyle(value string) (render.FontStyle, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "normal", "roman":
		return render.FontStyleNormal, nil
	case "italic":
		return render.FontStyleItalic, nil
	case "oblique":
		return render.FontStyleOblique, nil
	default:
		return "", fmt.Errorf("invalid font style %q", value)
	}
}

func parseMPLFontVariant(value string) (string, error) {
	switch normalized := strings.ToLower(normalizeMPLValue(value)); normalized {
	case "normal", "small-caps", "small_caps":
		if normalized == "small_caps" {
			return "small-caps", nil
		}
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid font variant %q", value)
	}
}

func parseMPLFontStretch(value string) (string, error) {
	normalized := strings.ToLower(normalizeMPLValue(value))
	valid := map[string]struct{}{
		"ultra-condensed": {}, "extra-condensed": {}, "condensed": {},
		"semi-condensed": {}, "normal": {}, "semi-expanded": {},
		"expanded": {}, "extra-expanded": {}, "ultra-expanded": {},
		"wider": {}, "narrower": {},
	}
	if _, ok := valid[normalized]; ok {
		return normalized, nil
	}
	numeric, err := strconv.Atoi(normalized)
	if err == nil && numeric >= 0 && numeric <= 1000 {
		return strconv.Itoa(numeric), nil
	}
	return "", fmt.Errorf("invalid font stretch %q", value)
}

func fontKeyFromRC(font *FontRC) string {
	if font == nil {
		return ""
	}
	families := expandRCFontFamilies(font)
	if (font.Style == "" || font.Style == render.FontStyleNormal) &&
		(font.Weight == 0 || font.Weight == 400) &&
		(font.Stretch == "" || font.Stretch == "normal") &&
		(font.Variant == "" || font.Variant == "normal") {
		return strings.Join(families, ", ")
	}
	props := render.FontProperties{
		Families: families,
		Style:    font.Style,
		Weight:   font.Weight,
		Stretch:  font.Stretch,
		Variant:  font.Variant,
	}
	return render.FontPropertiesKey(props)
}

func expandRCFontFamilies(font *FontRC) []string {
	if font == nil {
		return nil
	}
	var expanded []string
	seen := make(map[string]struct{})
	add := func(families []string) {
		for _, family := range families {
			family = strings.TrimSpace(family)
			if family == "" {
				continue
			}
			key := strings.ToLower(family)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			expanded = append(expanded, family)
		}
	}
	for _, family := range font.Family {
		switch strings.ToLower(strings.TrimSpace(family)) {
		case "serif":
			add(font.Serif)
		case "sans-serif", "sans serif", "sans":
			add(font.SansSerif)
		case "cursive":
			add(font.Cursive)
		case "fantasy":
			add(font.Fantasy)
		case "monospace", "mono":
			add(font.Monospace)
		default:
			add([]string{family})
		}
	}
	return expanded
}

func parseMPLLegendLocation(value string) (string, error) {
	normalized := strings.ToLower(normalizeMPLValue(value))
	names := [...]string{
		"best",
		"upper right",
		"upper left",
		"lower left",
		"lower right",
		"right",
		"center left",
		"center right",
		"lower center",
		"upper center",
		"center",
	}
	for _, name := range names {
		if normalized == name {
			return name, nil
		}
	}
	if code, err := strconv.Atoi(normalized); err == nil && code >= 0 && code < len(names) {
		return names[code], nil
	}
	return "", fmt.Errorf("invalid legend location %q", value)
}

func parseMPLFontSizeOrNone(value string, base float64) (float64, error) {
	if strings.EqualFold(normalizeMPLValue(value), "none") {
		return 0, nil
	}
	return parseMPLFontSize(value, base)
}

// parseMPLHistBins mirrors matplotlib's validate_hist_bins for the values this
// port supports: a positive bin count or the 'auto' estimator (mapped to
// HistBinsAuto). The other named estimators and explicit edge lists are
// rejected.
func parseMPLHistBins(value string) (int, error) {
	normalized := normalizeMPLValue(value)
	if strings.EqualFold(strings.TrimSpace(normalized), "auto") {
		return HistBinsAuto, nil
	}
	parsed, err := strconv.Atoi(normalized)
	if err != nil {
		return 0, fmt.Errorf("invalid hist.bins %q (want a positive int or 'auto')", value)
	}
	if parsed < 1 {
		return 0, fmt.Errorf("invalid hist.bins %q (want a positive int or 'auto')", value)
	}
	return parsed, nil
}

// parseMPLSketch mirrors matplotlib's validate_sketch: "none" disables the
// filter (zero value); otherwise the value is a (scale, length, randomness)
// triple, with or without surrounding parentheses.
func parseMPLSketch(value string) (render.SketchParams, error) {
	normalized := normalizeMPLValue(value)
	if strings.EqualFold(strings.TrimSpace(normalized), "none") || normalized == "" {
		return render.SketchParams{}, nil
	}
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(normalized, ")"), "("))
	parts := splitOutsideQuotes(normalized, ',')
	if len(parts) != 3 {
		return render.SketchParams{}, fmt.Errorf("expected (scale, length, randomness) tuple, got %q", value)
	}
	scale, err := parseMPLFloat(parts[0])
	if err != nil {
		return render.SketchParams{}, err
	}
	length, err := parseMPLFloat(parts[1])
	if err != nil {
		return render.SketchParams{}, err
	}
	randomness, err := parseMPLFloat(parts[2])
	if err != nil {
		return render.SketchParams{}, err
	}
	return render.SketchParams{Scale: scale, Length: length, Randomness: randomness}, nil
}

func parseMPLBool(value string) (bool, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", value)
	}
}

func parseMPLFigureSize(value string) (float64, float64, error) {
	normalized := normalizeMPLValue(value)
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(normalized, "]"), "["))
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(normalized, ")"), "("))
	parts := splitOutsideQuotes(normalized, ',')
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected width,height pair, got %q", value)
	}
	width, err := parseMPLFloat(parts[0])
	if err != nil {
		return 0, 0, err
	}
	height, err := parseMPLFloat(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func parseMPLFontSize(value string, base float64) (float64, error) {
	if parsed, err := parseMPLFloat(value); err == nil {
		return parsed, nil
	}

	if base <= 0 {
		base = Default.FontSize
	}

	switch strings.ToLower(normalizeMPLValue(value)) {
	case "xx-small":
		return base * (6.94 / 12.0), nil
	case "x-small":
		return base * (8.33 / 12.0), nil
	case "small":
		return base * (10.0 / 12.0), nil
	case "medium":
		return base, nil
	case "large":
		return base * 1.2, nil
	case "x-large":
		return base * 1.44, nil
	case "xx-large":
		return base * 1.728, nil
	case "smaller":
		return base * (10.0 / 12.0), nil
	case "larger":
		return base * 1.2, nil
	default:
		return 0, fmt.Errorf("invalid font size %q", value)
	}
}

// parseMPLEnum validates value against a fixed set of lowercase options,
// returning the normalized lowercase match.
func parseMPLEnum(value string, options ...string) (string, error) {
	normalized := strings.ToLower(normalizeMPLValue(value))
	if slices.Contains(options, normalized) {
		return normalized, nil
	}
	return "", fmt.Errorf("invalid value %q, want one of %s", value, strings.Join(options, ", "))
}

// parseMPLSavefigDPI parses a savefig.dpi value: a float, or "figure" (mapped
// to 0, meaning "use the figure DPI").
func parseMPLSavefigDPI(value string) (float64, error) {
	if strings.EqualFold(normalizeMPLValue(value), "figure") {
		return 0, nil
	}
	return parseMPLFloat(value)
}

// parseMPLSavefigBbox parses a savefig.bbox value: "standard"/"none" (stored as
// empty) or "tight".
func parseMPLSavefigBbox(value string) (string, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "standard", "none", "":
		return "", nil
	case "tight":
		return "tight", nil
	default:
		return "", fmt.Errorf("invalid savefig.bbox %q, want standard or tight", value)
	}
}

// validateMPLSavefigColor accepts the "auto" sentinel or any valid color value.
func validateMPLSavefigColor(value string, rc *RC) error {
	if strings.EqualFold(normalizeMPLValue(value), "auto") {
		return nil
	}
	return validateMPLColorValue(value, *rc, false)
}

// parseMPLStringList parses a comma-separated argument list (e.g. animation
// writer args), trimming surrounding brackets. Empty input yields nil.
func parseMPLStringList(value string) []string {
	normalized := normalizeMPLValue(value)
	normalized = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(normalized, "]"), "["))
	if strings.TrimSpace(normalized) == "" {
		return nil
	}
	parts := splitOutsideQuotes(normalized, ',')
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := normalizeMPLValue(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseMPLFontType validates a pdf/ps fonttype value (Type 3 or TrueType 42).
func parseMPLFontType(value string) (int, error) {
	parsed, err := parseMPLInt(value)
	if err != nil {
		return 0, err
	}
	if parsed != 3 && parsed != 42 {
		return 0, fmt.Errorf("invalid fonttype %d, want 3 or 42", parsed)
	}
	return parsed, nil
}

// parseMPLStringOrNone returns the empty string for a "None" value (Matplotlib's
// None sentinel) and the normalized string otherwise.
func parseMPLStringOrNone(value string) string {
	normalized := normalizeMPLValue(value)
	if strings.EqualFold(normalized, "none") {
		return ""
	}
	return normalized
}

// parseMPLMathtextFallback validates a mathtext.fallback value: "cm", "stix",
// "stixsans", or "none"/"None" (stored as empty string).
func parseMPLMathtextFallback(value string) (string, error) {
	normalized := strings.ToLower(normalizeMPLValue(value))
	switch normalized {
	case "none", "":
		return "", nil
	case "cm", "stix", "stixsans":
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid mathtext.fallback %q, want cm, stix, stixsans, or None", value)
	}
}

// parseMPLLineStyleName validates a lines.linestyle value and returns the
// canonical short form. Matplotlib additionally accepts on-off dash tuples,
// which are not supported here.
func parseMPLLineStyleName(value string) (string, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "-", "solid":
		return "-", nil
	case "--", "dashed":
		return "--", nil
	case "-.", "dashdot":
		return "-.", nil
	case ":", "dotted":
		return ":", nil
	case "", " ", "none":
		return "none", nil
	default:
		return "", fmt.Errorf("unsupported linestyle %q", value)
	}
}

// parseMPLAspect validates an image.aspect value: "equal", "auto", or a numeric
// ratio. The normalized string is returned for storage.
func parseMPLAspect(value string) (string, error) {
	normalized := strings.ToLower(normalizeMPLValue(value))
	switch normalized {
	case "equal", "auto":
		return normalized, nil
	}
	if _, err := strconv.ParseFloat(normalized, 64); err == nil {
		return normalized, nil
	}
	return "", fmt.Errorf("invalid aspect %q, want equal, auto, or a number", value)
}

func parseMPLGridAxis(value string) (string, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "both", "x", "y":
		return strings.ToLower(normalizeMPLValue(value)), nil
	default:
		return "", fmt.Errorf("invalid grid axis %q", value)
	}
}

func parseMPLGridWhich(value string) (string, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "major", "minor", "both":
		return strings.ToLower(normalizeMPLValue(value)), nil
	default:
		return "", fmt.Errorf("invalid grid which %q", value)
	}
}

func parseMPLLineStyle(value string) ([]float64, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "", "-", "solid":
		return nil, nil
	case "--", "dashed":
		return []float64{6, 6}, nil
	case "-.", "dashdot":
		return []float64{6, 3, 1.5, 3}, nil
	case ":", "dotted":
		return []float64{1.2, 2.4}, nil
	default:
		return nil, fmt.Errorf("unsupported line style %q", value)
	}
}

func parseMPLFontFamilyList(value string) ([]string, error) {
	normalized := normalizeMPLValue(value)
	if normalized == "" {
		return nil, errors.New("empty font family")
	}
	if strings.HasPrefix(normalized, "[") && strings.HasSuffix(normalized, "]") {
		normalized = normalized[1 : len(normalized)-1]
	}
	items := splitOutsideQuotes(normalized, ',')
	families := make([]string, 0, len(items))
	for _, item := range items {
		if family := normalizeMPLValue(item); family != "" {
			families = append(families, family)
		}
	}
	if len(families) == 0 {
		return nil, errors.New("empty font family list")
	}
	return families, nil
}

func parseMPLColor(value string, rc RC) (render.Color, error) {
	normalized := normalizeMPLValue(value)
	if normalized == "" {
		return render.Color{}, errors.New("empty color")
	}

	if strings.EqualFold(normalized, "inherit") {
		return render.Color{}, errors.New(`special value "inherit" requires contextual handling`)
	}

	return color.ToRGBA(normalized, color.WithColorCycle(rc.Palette()), color.WithBareHex())
}

func parseMPLFloatList(value string) ([]float64, error) {
	parts := splitOutsideQuotes(normalizeMPLValue(value), ',')
	if len(parts) == 0 {
		return nil, errors.New("empty float list")
	}
	out := make([]float64, 0, len(parts))
	for _, part := range parts {
		v, err := parseMPLFloat(part)
		if err != nil {
			return nil, fmt.Errorf("invalid float list: %w", err)
		}
		out = append(out, v)
	}
	return out, nil
}

func parseMPLLineCap(value string) (render.LineCap, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "butt":
		return render.CapButt, nil
	case "round":
		return render.CapRound, nil
	case "projecting", "square":
		return render.CapSquare, nil
	default:
		return 0, errors.New(`expected "butt", "round", or "projecting"`)
	}
}

func parseMPLLineJoin(value string) (render.LineJoin, error) {
	switch strings.ToLower(normalizeMPLValue(value)) {
	case "miter":
		return render.JoinMiter, nil
	case "round":
		return render.JoinRound, nil
	case "bevel":
		return render.JoinBevel, nil
	default:
		return 0, errors.New(`expected "miter", "round", or "bevel"`)
	}
}

func parseMPLMarkerColor(value string, rc *RC) (MarkerColorRC, error) {
	normalized := normalizeMPLValue(value)
	switch {
	case strings.EqualFold(normalized, "auto"):
		return MarkerColorRC{Mode: MarkerColorAuto}, nil
	case strings.EqualFold(normalized, "none"):
		return MarkerColorRC{Mode: MarkerColorNone}, nil
	default:
		if rc == nil {
			return MarkerColorRC{}, errors.New("nil rc")
		}
		parsed, err := parseMPLColor(normalized, *rc)
		if err != nil {
			return MarkerColorRC{}, err
		}
		return MarkerColorRC{Mode: MarkerColorExplicit, Color: parsed, Raw: normalized}, nil
	}
}

func parseMPLMarkerFillStyle(value string) (MarkerFillStyle, error) {
	switch normalized := MarkerFillStyle(strings.ToLower(normalizeMPLValue(value))); normalized {
	case MarkerFillFull, MarkerFillLeft, MarkerFillRight, MarkerFillBottom, MarkerFillTop, MarkerFillNone:
		return normalized, nil
	default:
		return "", errors.New(`expected "full", "left", "right", "bottom", "top", or "none"`)
	}
}

// parseMPLPropCycle parses an axes.prop_cycle expression into a Cycler,
// porting Matplotlib's cycler grammar: a sum (+) and product (*) of cycler(...)
// terms. "*" binds tighter than "+", matching Python operator precedence. Each
// term is cycler(key, [...]) (positional) or cycler(key=[...]) (keyword). Keys
// color/linestyle(ls)/marker/linewidth(lw) are normalized and typed; unknown
// keys are stored verbatim as strings so faithful round-tripping is preserved.
func parseMPLPropCycle(value string, rc *RC) (*cycler.Cycler, error) {
	expr := strings.TrimSpace(value)
	if expr == "" {
		return nil, errors.New("empty cycler")
	}

	// Split additive groups first; each group is a product of terms.
	var result *cycler.Cycler
	for _, group := range splitTopLevel(expr, '+') {
		group = strings.TrimSpace(group)
		if group == "" {
			return nil, fmt.Errorf("unsupported cycler syntax %q", value)
		}
		var groupCycle *cycler.Cycler
		for _, term := range splitTopLevel(group, '*') {
			parsed, err := parseMPLCyclerTerm(term, rc)
			if err != nil {
				return nil, err
			}
			if groupCycle == nil {
				groupCycle = parsed
				continue
			}
			groupCycle, err = groupCycle.Multiply(parsed)
			if err != nil {
				return nil, err
			}
		}
		if result == nil {
			result = groupCycle
			continue
		}
		var err error
		if result, err = result.Concat(groupCycle); err != nil {
			return nil, err
		}
	}
	if result == nil || result.Len() == 0 {
		return nil, errors.New("empty cycler")
	}
	return result, nil
}

// parseMPLCyclerTerm parses a single cycler(...) term into a one-key Cycler.
func parseMPLCyclerTerm(term string, rc *RC) (*cycler.Cycler, error) {
	term = strings.TrimSpace(term)
	lower := strings.ToLower(term)
	if !strings.HasPrefix(lower, "cycler(") || !strings.HasSuffix(term, ")") {
		return nil, fmt.Errorf("unsupported cycler syntax %q", term)
	}
	inner := strings.TrimSpace(term[len("cycler(") : len(term)-1])
	if inner == "" {
		return nil, errors.New("empty cycler")
	}

	var rawKey, rawList string
	switch inner[0] {
	case '\'', '"':
		// Positional: cycler('color', [...]).
		parts := splitTopLevel(inner, ',')
		if len(parts) < 2 {
			return nil, fmt.Errorf("unsupported cycler syntax %q", term)
		}
		rawKey = normalizeMPLValue(parts[0])
		rawList = strings.TrimSpace(strings.Join(parts[1:], ","))
	default:
		// Keyword: cycler(color=[...]).
		left, right, ok := strings.Cut(inner, "=")
		if !ok {
			return nil, fmt.Errorf("unsupported cycler syntax %q", term)
		}
		rawKey = strings.TrimSpace(left)
		rawList = strings.TrimSpace(right)
	}

	key := normalizeCyclerKey(rawKey)
	if key == "" {
		return nil, fmt.Errorf("unsupported cycler key in %q", term)
	}
	if !strings.HasPrefix(rawList, "[") || !strings.HasSuffix(rawList, "]") {
		return nil, fmt.Errorf("unsupported cycler value list %q", rawList)
	}
	items := splitOutsideQuotes(rawList[1:len(rawList)-1], ',')
	if len(items) == 0 || (len(items) == 1 && items[0] == "") {
		return nil, errors.New("empty cycler value list")
	}

	values := make([]any, 0, len(items))
	for _, item := range items {
		v, err := parseMPLCyclerValue(key, item, rc)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return cycler.New(key, values...), nil
}

// parseMPLCyclerValue parses one list element according to the property's type:
// colors via parseMPLColor, numeric widths/sizes via float, everything else as a
// dequoted string (linestyle, marker, and unknown keys).
func parseMPLCyclerValue(key, item string, rc *RC) (any, error) {
	switch key {
	case "color":
		return parseMPLColor(item, *rc)
	case "linewidth", "markersize", "markeredgewidth":
		return strconv.ParseFloat(normalizeMPLValue(item), 64)
	default:
		return normalizeMPLValue(item), nil
	}
}

// normalizeCyclerKey maps Matplotlib property aliases to canonical names.
func normalizeCyclerKey(key string) string {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "color", "c":
		return "color"
	case "linestyle", "ls":
		return "linestyle"
	case "marker":
		return "marker"
	case "linewidth", "lw":
		return "linewidth"
	case "markersize", "ms":
		return "markersize"
	case "markeredgewidth", "mew":
		return "markeredgewidth"
	case "":
		return ""
	default:
		return strings.ToLower(strings.TrimSpace(key))
	}
}

// isColorOnlyCycle reports whether the cycle carries no property beyond color.
func isColorOnlyCycle(c *cycler.Cycler) bool {
	keys := c.Keys()
	if len(keys) == 0 {
		return true
	}
	for _, k := range keys {
		if k != "color" {
			return false
		}
	}
	return true
}

// splitTopLevel splits s on sep, ignoring separators nested inside (), [], or
// quotes. It returns the single trimmed input when no top-level separator is
// found.
func splitTopLevel(s string, sep rune) []string {
	parts := make([]string, 0, 4)
	var current strings.Builder
	depth := 0
	inQuote := rune(0)
	for _, r := range s {
		switch {
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			}
		case r == '\'' || r == '"':
			inQuote = r
		case r == '(' || r == '[':
			depth++
		case r == ')' || r == ']':
			depth--
		case r == sep && depth == 0:
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	return parts
}

func splitOutsideQuotes(value string, sep rune) []string {
	parts := make([]string, 0, 4)
	var current strings.Builder
	inQuote := rune(0)
	for _, r := range value {
		if inQuote != 0 {
			current.WriteRune(r)
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '\'' || r == '"' {
			inQuote = r
			current.WriteRune(r)
			continue
		}
		if r == sep {
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	return parts
}

func normalizeMPLValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		if (trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') ||
			(trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') {
			return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
		}
	}
	return trimmed
}

func normalizeMPLStyleName(name string) string {
	normalized := normalizeThemeName(strings.TrimSuffix(name, ".mplstyle"))
	if normalized == "" {
		return "custom"
	}
	return normalized
}

func validateMPLColorValue(value string, rc RC, allowInherit bool) error {
	normalized := normalizeMPLValue(value)
	if allowInherit && strings.EqualFold(normalized, "inherit") {
		return nil
	}
	_, err := parseMPLColor(normalized, rc)
	return err
}

func resolveMPLSpecialColor(value string, rc RC, inherited render.Color) render.Color {
	switch strings.ToLower(value) {
	case "", "inherit":
		return inherited
	default:
		parsed, err := parseMPLColor(value, rc)
		if err != nil {
			return inherited
		}
		return parsed
	}
}

func cloneDashPattern(dashes []float64) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	cloned := make([]float64, len(dashes))
	copy(cloned, dashes)
	return cloned
}
