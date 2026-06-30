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
	"axes.edgecolor",
	"axes.facecolor",
	"axes.grid",
	"axes.grid.axis",
	"axes.grid.which",
	"axes.labelcolor",
	"axes.labelsize",
	"axes.linewidth",
	"axes.prop_cycle",
	"axes.titlecolor",
	"axes.titlesize",
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
	"figure.dpi",
	"figure.facecolor",
	"figure.figsize",
	"font.family",
	"font.size",
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
	"image.aspect",
	"image.cmap",
	"image.composite_image",
	"image.interpolation",
	"image.interpolation_stage",
	"image.lut",
	"image.origin",
	"image.resample",
	"legend.edgecolor",
	"legend.facecolor",
	"legend.framealpha",
	"legend.frameon",
	"legend.fontsize",
	"legend.labelcolor",
	"lines.color",
	"lines.linewidth",
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
	"svg.fonttype",
	"svg.hashsalt",
	"svg.id",
	"svg.image_inline",
	"text.color",
	"text.usetex",
	"xtick.color",
	"xtick.labelcolor",
	"xtick.labelsize",
	"ytick.color",
	"ytick.labelcolor",
	"ytick.labelsize",
}

type mplStyleState struct {
	rc RC

	fontSizeSet bool

	figureFaceValue string
	figureFaceSet   bool
	figureWidth     float64
	figureHeight    float64
	figureSizeSet   bool
	textColorValue  string
	textColorSet    bool
	textUseTeX      bool
	textUseTeXSet   bool
	lineColorValue  string
	lineColorSet    bool

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
	case "figure.figsize":
		width, height, err := parseMPLFigureSize(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.figureWidth = width
		state.figureHeight = height
		state.figureSizeSet = true
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
		parsed, err := parseMPLFontFamily(value)
		if err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.rc.FontKey = parsed
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
	case "axes.titlecolor":
		if err := validateMPLColorValue(value, state.rc, false); err != nil {
			return fmt.Errorf("parse %s on line %d: %w", key, lineNo, err)
		}
		state.titleColorValue = normalizeMPLValue(value)
		state.titleColorSet = true
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
		report.Unsupported = append(report.Unsupported, MPLStyleIssue{
			Line:  lineNo,
			Key:   key,
			Value: strings.TrimSpace(value),
		})
		return nil
	}

	report.Applied = append(report.Applied, key)
	return nil
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
	if state.figureSizeSet {
		state.rc.FigureWidth = state.figureWidth
		state.rc.FigureHeight = state.figureHeight
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
	if state.axisLineWidthSet {
		state.rc.AxisLineWidth = state.axisLineWidthPt
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

func parseMPLFontFamily(value string) (string, error) {
	normalized := normalizeMPLValue(value)
	if normalized == "" {
		return "", errors.New("empty font family")
	}
	if strings.HasPrefix(normalized, "[") && strings.HasSuffix(normalized, "]") {
		items := splitOutsideQuotes(normalized[1:len(normalized)-1], ',')
		for _, item := range items {
			candidate := normalizeMPLValue(item)
			if candidate != "" {
				return candidate, nil
			}
		}
		return "", errors.New("empty font family list")
	}
	items := splitOutsideQuotes(normalized, ',')
	if len(items) > 0 {
		first := normalizeMPLValue(items[0])
		if first != "" {
			return first, nil
		}
	}
	return normalized, nil
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
