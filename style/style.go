package style

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/render"
)

// RC holds global rendering defaults (rc-like configuration).
// Fields are simple value types to keep configuration immutable-ish by copy.
type RC struct {
	DPI          float64
	FontKey      string
	FontSize     float64
	LineWidth    float64
	TextColor    [4]float64
	LineColor    [4]float64
	Background   [4]float64
	TickCountX   int
	TickCountY   int
	FigureWidth  float64
	FigureHeight float64
	UseTeX       bool

	PathSimplify          bool
	PathSimplifyThreshold float64
	AggPathChunkSize      int
	// PathSketch is the global sketch/xkcd path perturbation applied to every
	// drawn path. The zero value disables it (matplotlib's "path.sketch": None).
	PathSketch render.SketchParams

	AxesBackground     render.Color
	AxesEdgeColor      render.Color
	AxesTitleColor     render.Color
	AxesLabelColor     render.Color
	AxisLineWidth      float64
	XTickColor         render.Color
	YTickColor         render.Color
	TitleFontSize      float64
	AxisLabelFontSize  float64
	XTickLabelFontSize float64
	YTickLabelFontSize float64
	GridColor          render.Color
	MinorGridColor     render.Color
	GridLineWidth      float64
	MinorGridLineWidth float64
	GridDashes         []float64
	MinorGridDashes    []float64
	GridVisible        bool
	GridAxis           string
	GridWhich          string
	LegendBackground   render.Color
	LegendBorderColor  render.Color
	LegendTextColor    render.Color
	LegendFontSize     float64
	LegendFrameAlpha   float64
	LegendFrameOn      bool
	ColorCycle         color.Palette
	WidgetVisualStyle  WidgetVisualStyle

	// Image holds image.* rcParams (imshow defaults).
	Image ImageRC
	// Hatch holds hatch.* rcParams (hatch pattern defaults).
	Hatch HatchRC
	// Boxplot holds boxplot.* rcParams (boxplot artist defaults).
	Boxplot BoxplotRC
	// Mathtext holds mathtext.* rcParams (math rendering font defaults).
	Mathtext MathtextRC
	// Date holds date.* rcParams (date axis formatter defaults).
	Date DateRC
	// PDF holds pdf.* rcParams (PDF backend defaults).
	PDF PDFRC
	// PS holds ps.* rcParams (PostScript backend defaults).
	PS PSRC
	// SVG holds svg.* rcParams (SVG backend defaults).
	SVG SVGRC
	// Animation holds animation.* rcParams (animation writer defaults).
	Animation AnimationRC
	// Savefig holds savefig.* rcParams (save-time output defaults).
	Savefig SavefigRC
}

// ImageRC mirrors Matplotlib's image.* rcParams used as imshow defaults.
type ImageRC struct {
	// Cmap is the default colormap name (image.cmap).
	Cmap string
	// Interpolation is the default interpolation method (image.interpolation);
	// "auto" defers to the renderer's choice.
	Interpolation string
	// InterpolationStage selects when interpolation is applied
	// (image.interpolation_stage): "auto", "data", or "rgba". Stored only.
	InterpolationStage string
	// Origin places the [0,0] index at the upper or lower corner
	// (image.origin): "upper" or "lower". Stored only.
	Origin string
	// Aspect is the default axes aspect for images (image.aspect):
	// "equal", "auto", or a numeric ratio. Stored only.
	Aspect string
	// Resample enables resampling when scaling images (image.resample).
	// Stored only.
	Resample bool
	// LUT is the number of colors in the colormap lookup table (image.lut).
	// Stored only.
	LUT int
	// CompositeImage composites multiple images into one (image.composite_image).
	// Stored only.
	CompositeImage bool
}

// HatchRC mirrors Matplotlib's hatch.* rcParams.
type HatchRC struct {
	// Color is the default hatch line color (hatch.color).
	Color render.Color
	// LineWidth is the default hatch line width in points (hatch.linewidth).
	LineWidth float64
}

// BoxplotRC mirrors the wired subset of Matplotlib's boxplot.* rcParams.
type BoxplotRC struct {
	// Notch draws notched boxes (boxplot.notch). Stored only.
	Notch bool
	// Vertical orients boxes vertically (boxplot.vertical). Stored only.
	Vertical bool
	// Whiskers sets the whisker length convention (boxplot.whiskers). Stored only.
	Whiskers float64
	// PatchArtist draws boxes as filled patches (boxplot.patchartist). Stored only.
	PatchArtist bool
	// ShowMeans shows the mean marker/line (boxplot.showmeans).
	ShowMeans bool
	// ShowCaps shows the whisker caps (boxplot.showcaps).
	ShowCaps bool
	// ShowBox shows the central box (boxplot.showbox).
	ShowBox bool
	// ShowFliers shows outlier points (boxplot.showfliers).
	ShowFliers bool
	// MeanLine draws the mean as a line rather than a marker (boxplot.meanline).
	MeanLine bool
	// FlierColor is the outlier marker color (boxplot.flierprops.color).
	FlierColor render.Color
	// FlierMarkerSize is the outlier marker size in points
	// (boxplot.flierprops.markersize).
	FlierMarkerSize float64
	// FlierEdgeWidth is the outlier marker edge width in points
	// (boxplot.flierprops.markeredgewidth).
	FlierEdgeWidth float64
	// BoxLineWidth is the box edge width in points (boxplot.boxprops.linewidth).
	BoxLineWidth float64
	// WhiskerLineWidth is the whisker width in points
	// (boxplot.whiskerprops.linewidth).
	WhiskerLineWidth float64
	// CapLineWidth is the cap width in points (boxplot.capprops.linewidth).
	CapLineWidth float64
	// MedianLineWidth is the median line width in points
	// (boxplot.medianprops.linewidth).
	MedianLineWidth float64
	// MedianColor is the median line color (boxplot.medianprops.color).
	MedianColor render.Color
	// MeanColor is the mean marker/line color (boxplot.meanprops.color).
	MeanColor render.Color
}

// MathtextRC mirrors Matplotlib's mathtext.* rcParams.
type MathtextRC struct {
	// Fontset selects the math font set (mathtext.fontset): "dejavusans",
	// "dejavuserif", "cm", "stix", "stixsans", or "custom".
	Fontset string
	// Default is the default math font style (mathtext.default), e.g. "it".
	// Stored only.
	Default string
	// Fallback is the fallback font set (mathtext.fallback); empty means None.
	// Stored only.
	Fallback string
	// BF is the bold font pattern (mathtext.bf). Stored only.
	BF string
	// BFit is the bold-italic font pattern (mathtext.bfit). Stored only.
	BFit string
	// Cal is the calligraphic font pattern (mathtext.cal). Stored only.
	Cal string
	// It is the italic font pattern (mathtext.it). Stored only.
	It string
	// RM is the roman font pattern (mathtext.rm). Stored only.
	RM string
	// SF is the sans-serif font pattern (mathtext.sf). Stored only.
	SF string
	// TT is the typewriter font pattern (mathtext.tt). Stored only.
	TT string
}

// DateRC mirrors Matplotlib's date.* rcParams. Auto* formatters are stored as
// strftime strings (stored only pending a strftime-to-Go-layout converter).
type DateRC struct {
	// AutoYear formats year-scale ticks (date.autoformatter.year).
	AutoYear string
	// AutoMonth formats month-scale ticks (date.autoformatter.month).
	AutoMonth string
	// AutoDay formats day-scale ticks (date.autoformatter.day).
	AutoDay string
	// AutoHour formats hour-scale ticks (date.autoformatter.hour).
	AutoHour string
	// AutoMinute formats minute-scale ticks (date.autoformatter.minute).
	AutoMinute string
	// AutoSecond formats second-scale ticks (date.autoformatter.second).
	AutoSecond string
	// AutoMicrosecond formats microsecond-scale ticks
	// (date.autoformatter.microsecond).
	AutoMicrosecond string
	// Epoch is the date epoch (date.epoch). Stored only.
	Epoch string
	// Converter selects the date converter (date.converter): "auto" or
	// "concise". Stored only.
	Converter string
	// IntervalMultiples snaps ticks to interval multiples
	// (date.interval_multiples). Stored only.
	IntervalMultiples bool
}

// PDFRC mirrors Matplotlib's pdf.* rcParams seeding the PDF backend defaults.
type PDFRC struct {
	// FontType is the embedded font type (pdf.fonttype): 3 or 42.
	FontType int
	// Use14CoreFonts uses the base-14 core fonts (pdf.use14corefonts).
	Use14CoreFonts bool
	// Compression is the stream compression level (pdf.compression). Stored only.
	Compression int
	// InheritColor inherits the current color rather than re-emitting it
	// (pdf.inheritcolor). Stored only.
	InheritColor bool
}

// PSRC mirrors Matplotlib's ps.* rcParams seeding the PostScript backend defaults.
type PSRC struct {
	// FontType is the embedded font type (ps.fonttype): 3 or 42.
	FontType int
	// UseAFM uses AFM (base-14) fonts (ps.useafm).
	UseAFM bool
	// PaperSize is the output paper size (ps.papersize), e.g. "letter".
	PaperSize string
	// UseDistiller post-processes output with a distiller (ps.usedistiller).
	// Stored only.
	UseDistiller bool
	// DistillerRes is the distiller resolution (ps.distiller.res). Stored only.
	DistillerRes int
}

// SVGRC mirrors Matplotlib's svg.* rcParams seeding the SVG backend defaults.
type SVGRC struct {
	// FontType selects text rendering (svg.fonttype): "none" or "path".
	FontType string
	// ImageInline inlines images as data URIs (svg.image_inline). Stored only.
	ImageInline bool
	// HashSalt salts deterministic element ids (svg.hashsalt); empty means None.
	HashSalt string
	// ID is the root element id (svg.id); empty means None. Stored only.
	ID string
}

// AnimationRC mirrors Matplotlib's animation.* rcParams (stored only; not yet
// consumed by the animation writers).
type AnimationRC struct {
	// HTML is the HTML representation (animation.html): "html5", "jshtml", or
	// "none".
	HTML string
	// Writer is the default writer name (animation.writer).
	Writer string
	// Codec is the default video codec (animation.codec).
	Codec string
	// Bitrate is the encoder bitrate (animation.bitrate); -1 means automatic.
	Bitrate int
	// FPS is the frames-per-second hint (animation.fps); 0 derives it from the
	// interval.
	FPS int
	// FrameFormat is the temporary frame image format (animation.frame_format).
	FrameFormat string
	// FFmpegPath is the ffmpeg executable path (animation.ffmpeg_path).
	FFmpegPath string
	// FFmpegArgs are extra ffmpeg arguments (animation.ffmpeg_args).
	FFmpegArgs []string
	// ConvertPath is the ImageMagick convert path (animation.convert_path).
	ConvertPath string
	// ConvertArgs are extra convert arguments (animation.convert_args).
	ConvertArgs []string
	// EmbedLimit is the embedded animation size limit in MB
	// (animation.embed_limit).
	EmbedLimit float64
}

// SavefigRC mirrors Matplotlib's savefig.* rcParams applied at save time.
type SavefigRC struct {
	// Dpi is the save-time resolution (savefig.dpi); 0 means use the figure DPI.
	Dpi float64
	// Facecolor is the figure face color override (savefig.facecolor); "auto"
	// inherits the figure background.
	Facecolor string
	// Edgecolor is the figure edge color override (savefig.edgecolor); "auto"
	// inherits the figure background.
	Edgecolor string
	// Transparent renders with a transparent figure/axes background
	// (savefig.transparent).
	Transparent bool
	// BboxInches selects the saved bounding box (savefig.bbox); "" is standard,
	// "tight" crops to content.
	BboxInches string
	// PadInches is the padding around a tight bbox in inches (savefig.pad_inches).
	PadInches float64
	// Format is the default output format (savefig.format); empty infers from the
	// path extension.
	Format string
}

// WidgetVisualStyle selects the visual defaults used by widget artists.
type WidgetVisualStyle string

const (
	// WidgetVisualGo uses matplotlib-go's native widget appearance.
	WidgetVisualGo WidgetVisualStyle = "go"
	// WidgetVisualMatplotlib uses source-backed Matplotlib-like widget defaults.
	WidgetVisualMatplotlib WidgetVisualStyle = "matplotlib"
)

// Default contains the library defaults. Copy and apply options to customize.
var Default = RC{
	DPI:                   100,
	FontKey:               "DejaVu Sans",
	FontSize:              10,
	LineWidth:             1.25,
	TextColor:             [4]float64{0, 0, 0, 1},
	LineColor:             [4]float64{0, 0, 0, 1},
	Background:            [4]float64{1, 1, 1, 1},
	TickCountX:            5,
	TickCountY:            5,
	FigureWidth:           6.4,
	FigureHeight:          4.8,
	UseTeX:                false,
	PathSimplify:          false,
	PathSimplifyThreshold: 1.0 / 9.0,
	AggPathChunkSize:      0,
	AxesBackground:        render.Color{R: 1, G: 1, B: 1, A: 1},
	AxesEdgeColor:         render.Color{R: 0, G: 0, B: 0, A: 1},
	AxesTitleColor:        render.Color{R: 0, G: 0, B: 0, A: 1},
	AxesLabelColor:        render.Color{R: 0, G: 0, B: 0, A: 1},
	AxisLineWidth:         0.8 * 100.0 / 72.0,
	XTickColor:            render.Color{R: 0, G: 0, B: 0, A: 1},
	YTickColor:            render.Color{R: 0, G: 0, B: 0, A: 1},
	TitleFontSize:         12,
	AxisLabelFontSize:     10,
	XTickLabelFontSize:    10,
	YTickLabelFontSize:    10,
	GridColor:             render.Color{R: 0xb0 / 255.0, G: 0xb0 / 255.0, B: 0xb0 / 255.0, A: 1},
	MinorGridColor:        render.Color{R: 0xb0 / 255.0, G: 0xb0 / 255.0, B: 0xb0 / 255.0, A: 1},
	GridLineWidth:         0.8 * 100.0 / 72.0,
	MinorGridLineWidth:    0.8 * 100.0 / 72.0,
	GridVisible:           false,
	GridAxis:              "both",
	GridWhich:             "major",
	LegendBackground:      render.Color{R: 1, G: 1, B: 1, A: 0.8},
	LegendBorderColor:     render.Color{R: 0.8, G: 0.8, B: 0.8, A: 0.8},
	LegendTextColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
	LegendFontSize:        10,
	LegendFrameAlpha:      0.8,
	LegendFrameOn:         true,
	ColorCycle:            color.Tab10,
	WidgetVisualStyle:     WidgetVisualGo,
	Image: ImageRC{
		Cmap:               "viridis",
		Interpolation:      "auto",
		InterpolationStage: "auto",
		Origin:             "upper",
		Aspect:             "equal",
		Resample:           true,
		LUT:                256,
		CompositeImage:     true,
	},
	Hatch: HatchRC{
		Color:     render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 1.0,
	},
	Boxplot: BoxplotRC{
		Notch:            false,
		Vertical:         true,
		Whiskers:         1.5,
		PatchArtist:      false,
		ShowMeans:        false,
		ShowCaps:         true,
		ShowBox:          true,
		ShowFliers:       true,
		MeanLine:         false,
		FlierColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
		FlierMarkerSize:  6,
		FlierEdgeWidth:   1.0,
		BoxLineWidth:     1.0,
		WhiskerLineWidth: 1.0,
		CapLineWidth:     1.0,
		MedianLineWidth:  1.0,
		MedianColor:      color.Tab10[1],
		MeanColor:        color.Tab10[2],
	},
	Mathtext: MathtextRC{
		Fontset:  "dejavusans",
		Default:  "it",
		Fallback: "cm",
		BF:       "sans:bold",
		BFit:     "sans:italic:bold",
		Cal:      "cursive",
		It:       "sans:italic",
		RM:       "sans",
		SF:       "sans",
		TT:       "monospace",
	},
	Date: DateRC{
		AutoYear:          "%Y",
		AutoMonth:         "%Y-%m",
		AutoDay:           "%Y-%m-%d",
		AutoHour:          "%m-%d %H",
		AutoMinute:        "%d %H:%M",
		AutoSecond:        "%H:%M:%S",
		AutoMicrosecond:   "%M:%S.%f",
		Epoch:             "1970-01-01T00:00:00",
		Converter:         "auto",
		IntervalMultiples: true,
	},
	PDF: PDFRC{
		FontType:       3,
		Use14CoreFonts: false,
		Compression:    6,
		InheritColor:   false,
	},
	PS: PSRC{
		FontType:     3,
		UseAFM:       false,
		PaperSize:    "letter",
		DistillerRes: 6000,
	},
	SVG: SVGRC{
		FontType:    "path",
		ImageInline: true,
		HashSalt:    "",
		ID:          "",
	},
	Animation: AnimationRC{
		HTML:        "none",
		Writer:      "ffmpeg",
		Codec:       "h264",
		Bitrate:     -1,
		FPS:         0,
		FrameFormat: "png",
		FFmpegPath:  "ffmpeg",
		ConvertPath: "convert",
		ConvertArgs: []string{"-layers", "OptimizePlus"},
		EmbedLimit:  20.0,
	},
	Savefig: SavefigRC{
		Dpi:         0,
		Facecolor:   "auto",
		Edgecolor:   "auto",
		Transparent: false,
		BboxInches:  "",
		PadInches:   0.1,
		Format:      "",
	},
}

// Option mutates an RC. Options should be applied on a copy derived from Default.
type Option func(*RC)

// Apply copies base and applies the given options in order, returning the result.
//
//nolint:gocritic // RC is intentionally passed by value to preserve copy-on-apply semantics.
func Apply(base RC, opts ...Option) RC {
	rc := base
	rc.ColorCycle = clonePalette(rc.ColorCycle)
	rc.GridDashes = cloneDashes(rc.GridDashes)
	rc.MinorGridDashes = cloneDashes(rc.MinorGridDashes)
	rc.Animation.FFmpegArgs = cloneStrings(rc.Animation.FFmpegArgs)
	rc.Animation.ConvertArgs = cloneStrings(rc.Animation.ConvertArgs)
	for _, opt := range opts {
		if opt != nil {
			opt(&rc)
		}
	}
	return rc
}

// WithDPI sets the DPI.
func WithDPI(d float64) Option { return func(rc *RC) { rc.DPI = d } }

// WithFont sets the font key and size.
func WithFont(key string, size float64) Option {
	return func(rc *RC) {
		rc.FontKey, rc.FontSize = key, size
		rc.TitleFontSize = maxFloat(8, size*1.2)
		rc.AxisLabelFontSize = maxFloat(8, size)
		rc.XTickLabelFontSize = maxFloat(8, size)
		rc.YTickLabelFontSize = maxFloat(8, size)
		rc.LegendFontSize = maxFloat(8, size)
	}
}

// WithLineWidth sets the default line width.
func WithLineWidth(w float64) Option { return func(rc *RC) { rc.LineWidth = w } }

// WithPathSketch sets the global sketch/xkcd path perturbation (matplotlib's
// "path.sketch" rcParam). The zero value disables sketching.
func WithPathSketch(params render.SketchParams) Option {
	return func(rc *RC) { rc.PathSketch = params }
}

// WithXkcd enables matplotlib-style xkcd sketch rendering on every path, using
// the same defaults as pyplot.xkcd(): scale=1, length=100, randomness=2. Note
// the xkcd handwriting font is not bundled, so only the path wiggle is applied.
func WithXkcd() Option {
	return WithPathSketch(render.SketchParams{Scale: 1, Length: 100, Randomness: 2})
}

// WithTextColor sets the default text color as normalized sRGBA (0..1).
func WithTextColor(r, g, b, a float64) Option {
	return func(rc *RC) {
		color := render.Color{R: r, G: g, B: b, A: a}
		rc.TextColor = [4]float64{r, g, b, a}
		rc.AxesTitleColor = color
		rc.AxesLabelColor = color
		rc.LegendTextColor = color
	}
}

// WithLineColor sets the default stroke color as normalized sRGBA (0..1).
func WithLineColor(r, g, b, a float64) Option {
	return func(rc *RC) { rc.LineColor = [4]float64{r, g, b, a} }
}

// WithBackground sets the default background color as normalized sRGBA (0..1).
func WithBackground(r, g, b, a float64) Option {
	return func(rc *RC) { rc.Background = [4]float64{r, g, b, a} }
}

// WithTickCounts sets the target tick counts for X and Y.
func WithTickCounts(nx, ny int) Option { return func(rc *RC) { rc.TickCountX, rc.TickCountY = nx, ny } }

// WithAxesBackground sets the axes face color.
func WithAxesBackground(c render.Color) Option {
	return func(rc *RC) { rc.AxesBackground = c }
}

// WithAxesEdgeColor sets the axes spine and tick color.
func WithAxesEdgeColor(c render.Color) Option {
	return func(rc *RC) {
		rc.AxesEdgeColor = c
		rc.XTickColor = c
		rc.YTickColor = c
	}
}

// WithAxisLineWidth sets the default axes spine and tick width.
func WithAxisLineWidth(w float64) Option {
	return func(rc *RC) { rc.AxisLineWidth = w }
}

// WithGridColors sets the default major and minor grid colors.
func WithGridColors(major, minor render.Color) Option {
	return func(rc *RC) {
		rc.GridColor = major
		rc.MinorGridColor = minor
	}
}

// WithGridLineWidths sets the default major and minor grid widths.
func WithGridLineWidths(major, minor float64) Option {
	return func(rc *RC) {
		rc.GridLineWidth = major
		rc.MinorGridLineWidth = minor
	}
}

// WithLegendColors sets the legend box and text colors.
func WithLegendColors(background, border, text render.Color) Option {
	return func(rc *RC) {
		rc.LegendBackground = background
		rc.LegendBorderColor = border
		rc.LegendTextColor = text
	}
}

// WithColorCycle sets the automatic series color palette.
func WithColorCycle(palette color.Palette) Option {
	return func(rc *RC) { rc.ColorCycle = clonePalette(palette) }
}

// WithTheme replaces the current RC with the named theme preset.
func WithTheme(theme Theme) Option {
	return func(rc *RC) { *rc = Apply(theme.RC) }
}

// WithWidgetVisualStyle sets the default visual style for widget artists.
func WithWidgetVisualStyle(widgetStyle WidgetVisualStyle) Option {
	return func(rc *RC) {
		switch widgetStyle {
		case WidgetVisualMatplotlib:
			rc.WidgetVisualStyle = widgetStyle
		default:
			rc.WidgetVisualStyle = WidgetVisualGo
		}
	}
}

// FigureBackground returns the figure face color as a renderer color.
func (rc RC) FigureBackground() render.Color {
	return render.Color{
		R: rc.Background[0],
		G: rc.Background[1],
		B: rc.Background[2],
		A: rc.Background[3],
	}
}

// DefaultTextColor returns the default text color as a renderer color.
func (rc RC) DefaultTextColor() render.Color {
	return render.Color{
		R: rc.TextColor[0],
		G: rc.TextColor[1],
		B: rc.TextColor[2],
		A: rc.TextColor[3],
	}
}

// DefaultLineColor returns the default line color as a renderer color.
func (rc RC) DefaultLineColor() render.Color {
	return render.Color{
		R: rc.LineColor[0],
		G: rc.LineColor[1],
		B: rc.LineColor[2],
		A: rc.LineColor[3],
	}
}

// DefaultAxesTitleColor returns the configured axes-title color.
func (rc RC) DefaultAxesTitleColor() render.Color {
	return rc.AxesTitleColor
}

// DefaultAxesLabelColor returns the configured axes-label color.
func (rc RC) DefaultAxesLabelColor() render.Color {
	return rc.AxesLabelColor
}

// TitleSize returns the configured title size with a minimum fallback.
func (rc RC) TitleSize() float64 {
	if rc.TitleFontSize >= 8 {
		return rc.TitleFontSize
	}
	if rc.FontSize > 0 {
		return maxFloat(8, rc.FontSize*1.2)
	}
	return 12
}

// AxisLabelSize returns the configured axis-label size with a minimum fallback.
func (rc RC) AxisLabelSize() float64 {
	if rc.AxisLabelFontSize >= 8 {
		return rc.AxisLabelFontSize
	}
	if rc.FontSize > 0 {
		return maxFloat(8, rc.FontSize)
	}
	return 8
}

// TickLabelSize returns the configured tick-label size for the requested axis.
func (rc RC) TickLabelSize(axis string) float64 {
	switch strings.ToLower(strings.TrimSpace(axis)) {
	case "y":
		if rc.YTickLabelFontSize >= 8 {
			return rc.YTickLabelFontSize
		}
	default:
		if rc.XTickLabelFontSize >= 8 {
			return rc.XTickLabelFontSize
		}
	}
	if rc.FontSize > 0 {
		return maxFloat(8, rc.FontSize)
	}
	return 8
}

// LegendSize returns the configured legend font size with a minimum fallback.
func (rc RC) LegendSize() float64 {
	if rc.LegendFontSize >= 8 {
		return rc.LegendFontSize
	}
	if rc.FontSize > 0 {
		return maxFloat(8, rc.FontSize)
	}
	return 8
}

// Palette returns a copy of the configured automatic color cycle.
func (rc RC) Palette() color.Palette {
	return clonePalette(rc.ColorCycle)
}

func clonePalette(palette color.Palette) color.Palette {
	if len(palette) == 0 {
		palette = color.Tab10
	}
	cloned := make(color.Palette, len(palette))
	copy(cloned, palette)
	return cloned
}

func cloneDashes(dashes []float64) []float64 {
	if len(dashes) == 0 {
		return nil
	}
	cloned := make([]float64, len(dashes))
	copy(cloned, dashes)
	return cloned
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// DefaultFigureSizePx returns the configured default figure size in pixels.
func (rc RC) DefaultFigureSizePx() (int, int) {
	dpi := rc.DPI
	if dpi <= 0 {
		dpi = Default.DPI
	}
	width := int(rc.FigureWidth*dpi + 0.5)
	height := int(rc.FigureHeight*dpi + 0.5)
	if width <= 0 {
		width = 640
	}
	if height <= 0 {
		height = 480
	}
	return width, height
}
