package style

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/cwbudde/matplotlib-go/render"
)

// Params holds rcParam-style key/value overrides using Matplotlib-style names.
type Params map[string]string

var runtimeDefaults = struct {
	mu      sync.Mutex
	current RC
	stack   []RC
}{
	current: Apply(Default),
}

// CurrentDefaults returns the active runtime defaults used for newly created figures.
func CurrentDefaults() RC {
	runtimeDefaults.mu.Lock()
	defer runtimeDefaults.mu.Unlock()
	return Apply(runtimeDefaults.current)
}

// CurrentParams returns the active runtime defaults serialized as rcParam values.
func CurrentParams() Params {
	return paramsFromRC(CurrentDefaults())
}

// ResetDefaults restores the active runtime defaults to the library baseline.
func ResetDefaults() {
	runtimeDefaults.mu.Lock()
	defer runtimeDefaults.mu.Unlock()
	runtimeDefaults.current = Apply(Default)
	runtimeDefaults.stack = nil
}

// UpdateParams applies rcParam-style overrides to the active runtime defaults.
func UpdateParams(params Params) (MPLStyleReport, error) {
	runtimeDefaults.mu.Lock()
	defer runtimeDefaults.mu.Unlock()

	next, report, err := applyMPLStyleParams(runtimeDefaults.current, params)
	if err != nil {
		return report, err
	}
	runtimeDefaults.current = next
	return report, nil
}

// PushContext applies temporary rcParam overrides and returns a restore function.
func PushContext(params Params) (func(), MPLStyleReport, error) {
	runtimeDefaults.mu.Lock()
	defer runtimeDefaults.mu.Unlock()

	next, report, err := applyMPLStyleParams(runtimeDefaults.current, params)
	if err != nil {
		return nil, report, err
	}

	previous := Apply(runtimeDefaults.current)
	runtimeDefaults.stack = append(runtimeDefaults.stack, previous)
	runtimeDefaults.current = next

	restore := func() {
		runtimeDefaults.mu.Lock()
		defer runtimeDefaults.mu.Unlock()
		n := len(runtimeDefaults.stack)
		if n == 0 {
			runtimeDefaults.current = Apply(Default)
			return
		}
		runtimeDefaults.current = runtimeDefaults.stack[n-1]
		runtimeDefaults.stack = runtimeDefaults.stack[:n-1]
	}
	return restore, report, nil
}

// LoadRCFile loads a Matplotlib-style rc file and replaces the active runtime defaults.
func LoadRCFile(path string) (MPLStyleReport, error) {
	resolved, err := resolveRCFilePath(path)
	if err != nil {
		return MPLStyleReport{}, err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return MPLStyleReport{}, err
	}

	rc, report, err := parseMPLStyleRC(Default, string(data))
	if err != nil {
		return report, err
	}

	runtimeDefaults.mu.Lock()
	defer runtimeDefaults.mu.Unlock()
	runtimeDefaults.current = rc
	runtimeDefaults.stack = nil
	return report, nil
}

// LoadDefaultRCFile searches the standard rc-file locations and applies the first match.
func LoadDefaultRCFile() (string, MPLStyleReport, error) {
	for _, path := range DefaultRCSearchPaths() {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			report, loadErr := LoadRCFile(path)
			return path, report, loadErr
		}
	}
	return "", MPLStyleReport{}, os.ErrNotExist
}

// DefaultRCSearchPaths returns the rc-file locations consulted by LoadDefaultRCFile.
func DefaultRCSearchPaths() []string {
	paths := make([]string, 0, 5)
	seen := make(map[string]struct{})
	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	if envPath := strings.TrimSpace(os.Getenv("MATPLOTLIBRC")); envPath != "" {
		add(normalizeRCEnvPath(envPath))
	}
	add("matplotlibrc")

	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		add(filepath.Join(xdg, "matplotlib", "matplotlibrc"))
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		add(filepath.Join(home, ".config", "matplotlib", "matplotlibrc"))
		add(filepath.Join(home, ".matplotlib", "matplotlibrc"))
	}

	return paths
}

func resolveRCFilePath(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", errors.New("style: empty rc file path")
	}
	info, err := os.Stat(trimmed)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return filepath.Join(trimmed, "matplotlibrc"), nil
	}
	return trimmed, nil
}

func normalizeRCEnvPath(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return path
	}
	if info.IsDir() {
		return filepath.Join(path, "matplotlibrc")
	}
	return path
}

func paramsFromRC(rc RC) Params {
	params := make(Params, len(supportedMPLStyleKeys))

	params["errorbar.capsize"] = formatMPLFloat(rc.Errorbar.CapSize)
	params["scatter.edgecolors"] = rc.Scatter.EdgeColors
	params["scatter.marker"] = rc.Scatter.Marker
	params["lines.linestyle"] = rc.Lines.LineStyle
	params["lines.marker"] = rc.Lines.Marker
	params["lines.markeredgewidth"] = formatMPLFloat(rc.Lines.MarkerEdgeWidth)
	params["lines.markersize"] = formatMPLFloat(rc.Lines.MarkerSize)
	params["axes.autolimit_mode"] = rc.Axes.AutolimitMode
	params["axes.axisbelow"] = rc.Axes.AxisBelow
	params["axes.unicode_minus"] = formatMPLBool(rc.Axes.UnicodeMinus)
	params["axes.xmargin"] = formatMPLFloat(rc.Axes.XMargin)
	params["axes.ymargin"] = formatMPLFloat(rc.Axes.YMargin)
	params["axes.grid"] = formatMPLBool(rc.GridVisible)
	params["axes.grid.axis"] = rc.GridAxis
	params["axes.grid.which"] = rc.GridWhich
	params["axes.edgecolor"] = formatMPLColor(rc.AxesEdgeColor)
	params["axes.facecolor"] = formatMPLColor(rc.AxesBackground)
	params["axes.labelcolor"] = formatMPLColor(rc.DefaultAxesLabelColor())
	params["axes.labelsize"] = formatMPLFloat(rc.AxisLabelSize())
	params["axes.linewidth"] = formatMPLFloat(rc.AxisLineWidth)
	params["axes.prop_cycle"] = formatMPLColorCycle(rc.Palette())
	params["axes.titlecolor"] = formatMPLColor(rc.DefaultAxesTitleColor())
	params["axes.titlesize"] = formatMPLFloat(rc.TitleSize())
	params["boxplot.boxprops.linewidth"] = formatMPLFloat(rc.Boxplot.BoxLineWidth)
	params["boxplot.capprops.linewidth"] = formatMPLFloat(rc.Boxplot.CapLineWidth)
	params["boxplot.flierprops.color"] = formatMPLColor(rc.Boxplot.FlierColor)
	params["boxplot.flierprops.markeredgewidth"] = formatMPLFloat(rc.Boxplot.FlierEdgeWidth)
	params["boxplot.flierprops.markersize"] = formatMPLFloat(rc.Boxplot.FlierMarkerSize)
	params["boxplot.meanline"] = formatMPLBool(rc.Boxplot.MeanLine)
	params["boxplot.meanprops.color"] = formatMPLColor(rc.Boxplot.MeanColor)
	params["boxplot.medianprops.color"] = formatMPLColor(rc.Boxplot.MedianColor)
	params["boxplot.medianprops.linewidth"] = formatMPLFloat(rc.Boxplot.MedianLineWidth)
	params["boxplot.notch"] = formatMPLBool(rc.Boxplot.Notch)
	params["boxplot.patchartist"] = formatMPLBool(rc.Boxplot.PatchArtist)
	params["boxplot.showbox"] = formatMPLBool(rc.Boxplot.ShowBox)
	params["boxplot.showcaps"] = formatMPLBool(rc.Boxplot.ShowCaps)
	params["boxplot.showfliers"] = formatMPLBool(rc.Boxplot.ShowFliers)
	params["boxplot.showmeans"] = formatMPLBool(rc.Boxplot.ShowMeans)
	params["boxplot.vertical"] = formatMPLBool(rc.Boxplot.Vertical)
	params["boxplot.whiskerprops.linewidth"] = formatMPLFloat(rc.Boxplot.WhiskerLineWidth)
	params["boxplot.whiskers"] = formatMPLFloat(rc.Boxplot.Whiskers)
	params["date.autoformatter.day"] = rc.Date.AutoDay
	params["date.autoformatter.hour"] = rc.Date.AutoHour
	params["date.autoformatter.microsecond"] = rc.Date.AutoMicrosecond
	params["date.autoformatter.minute"] = rc.Date.AutoMinute
	params["date.autoformatter.month"] = rc.Date.AutoMonth
	params["date.autoformatter.second"] = rc.Date.AutoSecond
	params["date.autoformatter.year"] = rc.Date.AutoYear
	params["date.converter"] = rc.Date.Converter
	params["date.epoch"] = rc.Date.Epoch
	params["date.interval_multiples"] = formatMPLBool(rc.Date.IntervalMultiples)
	params["agg.path.chunksize"] = strconv.Itoa(rc.AggPathChunkSize)
	params["animation.bitrate"] = strconv.Itoa(rc.Animation.Bitrate)
	params["animation.codec"] = rc.Animation.Codec
	params["animation.convert_args"] = strings.Join(rc.Animation.ConvertArgs, ", ")
	params["animation.convert_path"] = rc.Animation.ConvertPath
	params["animation.embed_limit"] = formatMPLFloat(rc.Animation.EmbedLimit)
	params["animation.ffmpeg_args"] = strings.Join(rc.Animation.FFmpegArgs, ", ")
	params["animation.ffmpeg_path"] = rc.Animation.FFmpegPath
	params["animation.frame_format"] = rc.Animation.FrameFormat
	params["animation.html"] = rc.Animation.HTML
	params["animation.writer"] = rc.Animation.Writer
	params["figure.dpi"] = formatMPLFloat(rc.DPI)
	params["figure.facecolor"] = formatMPLColor(rc.FigureBackground())
	params["figure.figsize"] = fmt.Sprintf("%s, %s", formatMPLFloat(rc.FigureWidth), formatMPLFloat(rc.FigureHeight))
	params["font.family"] = rc.FontKey
	params["font.size"] = formatMPLFloat(rc.FontSize)
	params["grid.alpha"] = formatMPLFloat(rc.GridColor.A)
	params["grid.color"] = formatMPLColor(rc.GridColor)
	params["grid.linewidth"] = formatMPLFloat(rc.GridLineWidth)
	params["grid.linestyle"] = formatMPLLineStyle(rc.GridDashes)
	params["grid.major.color"] = formatMPLColor(rc.GridColor)
	params["grid.major.linestyle"] = formatMPLLineStyle(rc.GridDashes)
	params["grid.minor.color"] = formatMPLColor(rc.MinorGridColor)
	params["grid.minor.linestyle"] = formatMPLLineStyle(rc.MinorGridDashes)
	params["hatch.color"] = formatMPLColor(rc.Hatch.Color)
	params["hatch.linewidth"] = formatMPLFloat(rc.Hatch.LineWidth)
	params["hist.bins"] = formatMPLHistBins(rc.HistBins)
	params["image.aspect"] = rc.Image.Aspect
	params["image.cmap"] = rc.Image.Cmap
	params["image.composite_image"] = formatMPLBool(rc.Image.CompositeImage)
	params["image.interpolation"] = rc.Image.Interpolation
	params["image.interpolation_stage"] = rc.Image.InterpolationStage
	params["image.lut"] = strconv.Itoa(rc.Image.LUT)
	params["image.origin"] = rc.Image.Origin
	params["image.resample"] = formatMPLBool(rc.Image.Resample)
	params["legend.edgecolor"] = formatMPLColor(rc.LegendBorderColor)
	params["legend.facecolor"] = formatMPLColor(rc.LegendBackground)
	params["legend.fancybox"] = formatMPLBool(rc.Legend.FancyBox)
	params["legend.fontsize"] = formatMPLFloat(rc.LegendSize())
	params["legend.framealpha"] = formatMPLFloat(rc.LegendFrameAlpha)
	params["legend.frameon"] = formatMPLBool(rc.LegendFrameOn)
	params["legend.borderaxespad"] = formatMPLFloat(rc.Legend.BorderAxesPad)
	params["legend.borderpad"] = formatMPLFloat(rc.Legend.BorderPad)
	params["legend.columnspacing"] = formatMPLFloat(rc.Legend.ColumnSpacing)
	params["legend.handleheight"] = formatMPLFloat(rc.Legend.HandleHeight)
	params["legend.handlelength"] = formatMPLFloat(rc.Legend.HandleLength)
	params["legend.handletextpad"] = formatMPLFloat(rc.Legend.HandleTextPad)
	params["legend.labelcolor"] = formatMPLColor(rc.LegendTextColor)
	params["legend.labelspacing"] = formatMPLFloat(rc.Legend.LabelSpacing)
	params["legend.loc"] = rc.Legend.Location
	params["legend.markerscale"] = formatMPLFloat(rc.Legend.MarkerScale)
	params["legend.numpoints"] = strconv.Itoa(rc.Legend.NumPoints)
	params["legend.scatterpoints"] = strconv.Itoa(rc.Legend.ScatterPoints)
	params["legend.shadow"] = formatMPLBool(rc.Legend.Shadow)
	params["legend.title_fontsize"] = formatMPLOptionalFontSize(rc.Legend.TitleFontSize)
	params["lines.color"] = formatMPLColor(rc.DefaultLineColor())
	params["lines.linewidth"] = formatMPLFloat(rc.LineWidth)
	params["mathtext.bf"] = rc.Mathtext.BF
	params["mathtext.bfit"] = rc.Mathtext.BFit
	params["mathtext.cal"] = rc.Mathtext.Cal
	params["mathtext.default"] = rc.Mathtext.Default
	params["mathtext.fallback"] = formatMPLMathtextFallback(rc.Mathtext.Fallback)
	params["mathtext.fontset"] = rc.Mathtext.Fontset
	params["mathtext.it"] = rc.Mathtext.It
	params["mathtext.rm"] = rc.Mathtext.RM
	params["mathtext.sf"] = rc.Mathtext.SF
	params["mathtext.tt"] = rc.Mathtext.TT
	params["path.simplify"] = formatMPLBool(rc.PathSimplify)
	params["path.simplify_threshold"] = formatMPLFloat(rc.PathSimplifyThreshold)
	params["pdf.compression"] = strconv.Itoa(rc.PDF.Compression)
	params["pdf.fonttype"] = strconv.Itoa(rc.PDF.FontType)
	params["pdf.inheritcolor"] = formatMPLBool(rc.PDF.InheritColor)
	params["pdf.use14corefonts"] = formatMPLBool(rc.PDF.Use14CoreFonts)
	params["ps.distiller.res"] = strconv.Itoa(rc.PS.DistillerRes)
	params["ps.fonttype"] = strconv.Itoa(rc.PS.FontType)
	params["ps.papersize"] = rc.PS.PaperSize
	params["ps.useafm"] = formatMPLBool(rc.PS.UseAFM)
	params["ps.usedistiller"] = formatMPLBool(rc.PS.UseDistiller)
	params["savefig.bbox"] = formatMPLSavefigBbox(rc.Savefig.BboxInches)
	params["savefig.dpi"] = formatMPLSavefigDPI(rc.Savefig.Dpi)
	params["savefig.edgecolor"] = rc.Savefig.Edgecolor
	params["savefig.facecolor"] = rc.Savefig.Facecolor
	params["savefig.format"] = rc.Savefig.Format
	params["savefig.pad_inches"] = formatMPLFloat(rc.Savefig.PadInches)
	params["savefig.transparent"] = formatMPLBool(rc.Savefig.Transparent)
	params["svg.fonttype"] = rc.SVG.FontType
	params["svg.hashsalt"] = formatMPLStringOrNone(rc.SVG.HashSalt)
	params["svg.id"] = formatMPLStringOrNone(rc.SVG.ID)
	params["svg.image_inline"] = formatMPLBool(rc.SVG.ImageInline)
	params["text.color"] = formatMPLColor(rc.DefaultTextColor())
	params["text.usetex"] = formatMPLBool(rc.UseTeX)
	params["xtick.alignment"] = rc.XTick.Alignment
	params["xtick.bottom"] = formatMPLBool(rc.XTick.Primary)
	params["xtick.color"] = formatMPLColor(rc.XTickColor)
	params["xtick.direction"] = rc.XTick.Direction
	params["xtick.labelbottom"] = formatMPLBool(rc.XTick.LabelPrimary)
	params["xtick.labelcolor"] = formatMPLColor(rc.XTickColor)
	params["xtick.labelsize"] = formatMPLFloat(rc.TickLabelSize("x"))
	params["xtick.labeltop"] = formatMPLBool(rc.XTick.LabelSecondary)
	params["xtick.major.bottom"] = formatMPLBool(rc.XTick.Major.Primary)
	params["xtick.major.pad"] = formatMPLFloat(rc.XTick.Major.Pad)
	params["xtick.major.size"] = formatMPLFloat(rc.XTick.Major.Size)
	params["xtick.major.top"] = formatMPLBool(rc.XTick.Major.Secondary)
	params["xtick.major.width"] = formatMPLFloat(rc.XTick.Major.Width)
	params["xtick.minor.bottom"] = formatMPLBool(rc.XTick.Minor.Primary)
	params["xtick.minor.ndivs"] = formatMPLMinorTickNDivs(rc.XTick.Minor.NDivs)
	params["xtick.minor.pad"] = formatMPLFloat(rc.XTick.Minor.Pad)
	params["xtick.minor.size"] = formatMPLFloat(rc.XTick.Minor.Size)
	params["xtick.minor.top"] = formatMPLBool(rc.XTick.Minor.Secondary)
	params["xtick.minor.visible"] = formatMPLBool(rc.XTick.Minor.Visible)
	params["xtick.minor.width"] = formatMPLFloat(rc.XTick.Minor.Width)
	params["xtick.top"] = formatMPLBool(rc.XTick.Secondary)
	params["ytick.alignment"] = rc.YTick.Alignment
	params["ytick.color"] = formatMPLColor(rc.YTickColor)
	params["ytick.direction"] = rc.YTick.Direction
	params["ytick.labelleft"] = formatMPLBool(rc.YTick.LabelPrimary)
	params["ytick.labelcolor"] = formatMPLColor(rc.YTickColor)
	params["ytick.labelright"] = formatMPLBool(rc.YTick.LabelSecondary)
	params["ytick.labelsize"] = formatMPLFloat(rc.TickLabelSize("y"))
	params["ytick.left"] = formatMPLBool(rc.YTick.Primary)
	params["ytick.major.left"] = formatMPLBool(rc.YTick.Major.Primary)
	params["ytick.major.pad"] = formatMPLFloat(rc.YTick.Major.Pad)
	params["ytick.major.right"] = formatMPLBool(rc.YTick.Major.Secondary)
	params["ytick.major.size"] = formatMPLFloat(rc.YTick.Major.Size)
	params["ytick.major.width"] = formatMPLFloat(rc.YTick.Major.Width)
	params["ytick.minor.left"] = formatMPLBool(rc.YTick.Minor.Primary)
	params["ytick.minor.ndivs"] = formatMPLMinorTickNDivs(rc.YTick.Minor.NDivs)
	params["ytick.minor.pad"] = formatMPLFloat(rc.YTick.Minor.Pad)
	params["ytick.minor.right"] = formatMPLBool(rc.YTick.Minor.Secondary)
	params["ytick.minor.size"] = formatMPLFloat(rc.YTick.Minor.Size)
	params["ytick.minor.visible"] = formatMPLBool(rc.YTick.Minor.Visible)
	params["ytick.minor.width"] = formatMPLFloat(rc.YTick.Minor.Width)
	params["ytick.right"] = formatMPLBool(rc.YTick.Secondary)

	return params
}

func formatMPLFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatMPLOptionalFontSize(value float64) string {
	if value <= 0 {
		return "None"
	}
	return formatMPLFloat(value)
}

func formatMPLMinorTickNDivs(ndivs int) string {
	if ndivs <= 0 {
		return "auto"
	}
	return strconv.Itoa(ndivs)
}

func formatMPLHistBins(bins int) string {
	if bins == HistBinsAuto {
		return "auto"
	}
	if bins == 0 {
		bins = 10
	}
	return strconv.Itoa(bins)
}

func formatMPLSavefigBbox(value string) string {
	if strings.TrimSpace(value) == "" {
		return "standard"
	}
	return value
}

func formatMPLSavefigDPI(dpi float64) string {
	if dpi <= 0 {
		return "figure"
	}
	return formatMPLFloat(dpi)
}

func formatMPLStringOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return value
}

func formatMPLMathtextFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return value
}

func formatMPLBool(value bool) string {
	if value {
		return "True"
	}
	return "False"
}

func formatMPLColor(color render.Color) string {
	r := colorChannelByte(color.R)
	g := colorChannelByte(color.G)
	b := colorChannelByte(color.B)
	a := colorChannelByte(color.A)
	if a == 0xFF {
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)
}

func formatMPLLineStyle(dashes []float64) string {
	switch {
	case len(dashes) == 0:
		return "-"
	case len(dashes) == 2 && almostEqualFloat(dashes[0], 6) && almostEqualFloat(dashes[1], 6):
		return "--"
	case len(dashes) == 4 && almostEqualFloat(dashes[0], 6) && almostEqualFloat(dashes[1], 3) &&
		almostEqualFloat(dashes[2], 1.5) && almostEqualFloat(dashes[3], 3):
		return "-."
	case len(dashes) == 2 && almostEqualFloat(dashes[0], 1.2) && almostEqualFloat(dashes[1], 2.4):
		return ":"
	default:
		return "-"
	}
}

func formatMPLColorCycle(palette []render.Color) string {
	if len(palette) == 0 {
		palette = Default.Palette()
	}
	parts := make([]string, len(palette))
	for i, color := range palette {
		parts[i] = fmt.Sprintf("'%s'", formatMPLColor(color))
	}
	return fmt.Sprintf("cycler('color', [%s])", strings.Join(parts, ", "))
}

func colorChannelByte(value float64) uint8 {
	switch {
	case value <= 0:
		return 0
	case value >= 1:
		return 0xFF
	default:
		return uint8(value*255 + 0.5)
	}
}

func almostEqualFloat(a, b float64) bool {
	const epsilon = 1e-9
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}
