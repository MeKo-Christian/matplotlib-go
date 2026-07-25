package style

import (
	"strings"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/cycler"
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
	// HistBins is the hist.bins rcParam: the default bin count for Hist when
	// no bins are given (matplotlib default 10). HistBinsAuto selects numpy's
	// 'auto' estimator; 0 means unset (callers fall back to the default 10).
	HistBins int

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
	// PropCycle holds the full axes.prop_cycle when it carries more than the
	// color column (linestyle/marker/linewidth). When nil, automatic styling
	// cycles ColorCycle only, preserving the historical color-only behavior.
	PropCycle         *cycler.Cycler
	WidgetVisualStyle WidgetVisualStyle

	// Axes holds behavior-related axes.* rcParams (grid/label styling has
	// dedicated flat RC fields above).
	Axes AxesRC
	// Lines holds line-artist lines.* rcParams (width/color live in the flat
	// LineWidth/LineColor fields above).
	Lines LinesRC
	// Patch holds patch-artist patch.* defaults.
	Patch PatchRC
	// Font holds the default font properties and ordered generic-family
	// fallback lists. FontKey is the renderer-facing serialization of these
	// values and remains available for backwards compatibility.
	Font FontRC
	// XTick and YTick hold tick geometry, placement, and visibility rcParams.
	XTick TickAxisRC
	YTick TickAxisRC
	// Legend holds legend placement, geometry, and visibility rcParams beyond
	// the flat color/font/frame fields above.
	Legend LegendRC
	// Scatter holds scatter.* rcParams (scatter artist defaults).
	Scatter ScatterRC
	// Errorbar holds errorbar.* rcParams (errorbar artist defaults).
	Errorbar ErrorbarRC
	// Image holds image.* rcParams (imshow defaults).
	Image ImageRC
	// Hatch holds hatch.* rcParams (hatch pattern defaults).
	Hatch HatchRC
	// Contour holds contour.* defaults for structured and triangular contours.
	Contour ContourRC
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
	// Figure holds figure patch, subplot, layout, and super-label defaults.
	Figure FigureRC
}

// FigureRC mirrors figure.* defaults consumed when a Figure and its managed
// subplot grids are created.
type FigureRC struct {
	EdgeColor   render.Color
	FrameOn     bool
	AutoLayout  bool
	TitleSize   float64
	TitleWeight int
	LabelSize   float64
	LabelWeight int
	Subplot     FigureSubplotRC
	Constrained FigureConstrainedLayoutRC
}

// FigureSubplotRC holds the initial Figure.subplot parameters. WSpace and
// HSpace use Matplotlib's units: fractions of the average Axes width/height.
type FigureSubplotRC struct {
	Left, Right, Bottom, Top float64
	WSpace, HSpace           float64
}

// FigureConstrainedLayoutRC holds constrained-layout defaults. Pad values are
// inches and spacing values are fractions, matching Matplotlib.
type FigureConstrainedLayoutRC struct {
	Use            bool
	HPad, WPad     float64
	HSpace, WSpace float64
}

// AxesRC mirrors behavior-related axes.* rcParams. Color/size styling for
// axes lives in the flat RC fields (AxesBackground, AxisLineWidth, ...).
type AxesRC struct {
	// AxisBelow places grid lines and ticks relative to artists
	// (axes.axisbelow): "line" (below lines, above patches — the default),
	// "True" (below everything), or "False" (above everything). Consumed at
	// axes creation (core.Axes.applyRCDefaults).
	AxisBelow string
	// XMargin and YMargin are the autoscale padding fractions
	// (axes.xmargin / axes.ymargin). Non-default values seed new axes'
	// margins (core.Axes.applyRCDefaults).
	XMargin float64
	YMargin float64
	// AutolimitMode finalizes autoscaled limits (axes.autolimit_mode):
	// "data" (tight to the padded range) or "round_numbers". Consumed at
	// axes creation (core.Axes.applyRCDefaults).
	AutolimitMode string
	// UnicodeMinus renders negative tick labels with U+2212 instead of a
	// hyphen (axes.unicode_minus). Consumed by the scalar tick formatters
	// (core/tick_formatters.go).
	UnicodeMinus bool
	// Formatter holds the axes.formatter.* defaults consumed by newly created
	// scalar and logarithmic axis formatters.
	Formatter FormatterRC
	// Spines controls the initial visibility of the four rectilinear axes
	// spines (axes.spines.{top,bottom,left,right}).
	Spines SpineRC
	// TitleLocation aligns the axes title at the left, center, or right edge
	// (axes.titlelocation).
	TitleLocation string
	// TitlePad is the display-space gap above the axes title anchor, in points
	// (axes.titlepad).
	TitlePad float64
	// TitleWeight is the numeric font weight used for axes titles
	// (axes.titleweight; 400 is normal and 700 is bold).
	TitleWeight int
	// TitleY is the title's axes-relative vertical coordinate when TitleYSet
	// is true (axes.titley). When false, title placement automatically clears
	// top-axis decorations.
	TitleY    float64
	TitleYSet bool
	// LabelPad is the gap between an axis label and its tick/spine extent, in
	// points (axes.labelpad).
	LabelPad float64
	// LabelWeight is the numeric font weight used for x/y axis labels
	// (axes.labelweight).
	LabelWeight int
}

// SpineRC mirrors Matplotlib's axes.spines.* visibility rcParams.
type SpineRC struct {
	Top    bool
	Bottom bool
	Left   bool
	Right  bool
}

// FormatterRC mirrors Matplotlib's axes.formatter.* rcParams.
type FormatterRC struct {
	// Limits are the inclusive powers at which ScalarFormatter switches to
	// scientific notation (axes.formatter.limits).
	Limits [2]int
	// MinExponent is the smallest absolute exponent displayed as a power by
	// LogFormatterMathText (axes.formatter.min_exponent).
	MinExponent int
	// OffsetThreshold is the number of leading digits an additive offset must
	// save before ScalarFormatter uses it (axes.formatter.offset_threshold).
	OffsetThreshold int
	// UseLocale localizes numeric separators (axes.formatter.use_locale).
	UseLocale bool
	// UseMathText wraps scalar labels and scientific notation in MathText
	// (axes.formatter.use_mathtext).
	UseMathText bool
	// UseOffset enables automatic additive offsets
	// (axes.formatter.useoffset).
	UseOffset bool
}

// LinesRC mirrors line-artist lines.* rcParams beyond the flat LineWidth and
// LineColor fields.
type LinesRC struct {
	// LineStyle is the default line style (lines.linestyle): "-", "--",
	// "-.", ":", or "none". Consumed by Axes.Plot for lines without an
	// explicit or cycled style.
	LineStyle string
	// Marker is the default marker (lines.marker); "None" draws no marker.
	// Consumed by Axes.Plot.
	Marker string
	// MarkerSize is the default marker size in points (lines.markersize).
	// Consumed by Axes.Plot.
	MarkerSize float64
	// MarkerEdgeWidth is the default marker edge width in points
	// (lines.markeredgewidth). Consumed by Axes.Plot. A zero rc value is
	// indistinguishable from Line2D's unset state (which falls back to the
	// 1 pt default) and is therefore ignored.
	MarkerEdgeWidth float64
	// DashedPattern, DashDotPattern, and DottedPattern are the unscaled
	// on/off sequences in points selected by the corresponding named
	// linestyle.
	DashedPattern  []float64
	DashDotPattern []float64
	DottedPattern  []float64
	// ScaleDashes scales named dash sequences by the line width.
	ScaleDashes bool
	// DashCap/Join and SolidCap/Join select stroke geometry according to
	// whether the resolved linestyle is dashed or solid.
	DashCap   render.LineCap
	DashJoin  render.LineJoin
	SolidCap  render.LineCap
	SolidJoin render.LineJoin
	// MarkerFaceColor and MarkerEdgeColor are the default marker colors.
	MarkerFaceColor MarkerColorRC
	MarkerEdgeColor MarkerColorRC
	// MarkerFillStyle is markers.fillstyle.
	MarkerFillStyle MarkerFillStyle
	// Antialiased is lines.antialiased.
	Antialiased bool
}

// PatchRC mirrors Matplotlib's patch.* defaults. LineWidth is expressed in
// points. ForceEdgeColor makes an omitted patch edge use EdgeColor; otherwise
// filled patches default to no visible edge.
type PatchRC struct {
	LineWidth float64
	FaceColor render.Color
	// FaceColorRaw preserves dynamic cycle references such as C0. FaceColor is
	// the resolved value for the current PropCycle/ColorCycle.
	FaceColorRaw   string
	EdgeColor      render.Color
	ForceEdgeColor bool
	Antialiased    bool
}

// DefaultPatchFaceColor resolves patch.facecolor against the current property
// cycle. This keeps the default C0 dynamic when callers replace ColorCycle
// directly rather than through an mplstyle parse.
func (r *RC) DefaultPatchFaceColor() render.Color {
	if r == nil {
		return render.Color{}
	}
	if r.Patch.FaceColorRaw != "" {
		if resolved, err := color.ToRGBA(
			r.Patch.FaceColorRaw,
			color.WithColorCycle(r.Palette()),
			color.WithBareHex(),
		); err == nil {
			return resolved
		}
	}
	return r.Patch.FaceColor
}

// MarkerColorMode identifies the special marker color values accepted by
// lines.markerfacecolor and lines.markeredgecolor.
type MarkerColorMode uint8

const (
	MarkerColorAuto MarkerColorMode = iota
	MarkerColorExplicit
	MarkerColorNone
)

// MarkerColorRC is a typed rcParam marker color. Raw preserves cycle colors
// such as C1 for resolution against the final axes.prop_cycle.
type MarkerColorRC struct {
	Mode  MarkerColorMode
	Color render.Color
	Raw   string
}

// MarkerFillStyle is the typed markers.fillstyle value.
type MarkerFillStyle string

const (
	MarkerFillFull   MarkerFillStyle = "full"
	MarkerFillLeft   MarkerFillStyle = "left"
	MarkerFillRight  MarkerFillStyle = "right"
	MarkerFillBottom MarkerFillStyle = "bottom"
	MarkerFillTop    MarkerFillStyle = "top"
	MarkerFillNone   MarkerFillStyle = "none"
)

// FontRC mirrors Matplotlib's font.* defaults. Family is the ordered
// font.family request; the five named lists expand generic family entries
// without discarding later fallbacks.
type FontRC struct {
	Family    []string
	Style     render.FontStyle
	Variant   string
	Weight    int
	Stretch   string
	Serif     []string
	SansSerif []string
	Cursive   []string
	Fantasy   []string
	Monospace []string
}

// TickAxisRC mirrors the axis-wide xtick.* or ytick.* rcParams. Primary means
// bottom for x and left for y; Secondary means top for x and right for y.
type TickAxisRC struct {
	Direction      string
	Alignment      string
	Primary        bool
	Secondary      bool
	LabelPrimary   bool
	LabelSecondary bool
	Major          TickLevelRC
	Minor          TickLevelRC
}

// TickLevelRC mirrors the major/minor tick geometry and per-side visibility.
// NDivs is used only by Minor; zero selects Matplotlib's automatic subdivision.
type TickLevelRC struct {
	Size      float64
	Width     float64
	Pad       float64
	Primary   bool
	Secondary bool
	Visible   bool
	NDivs     int
}

// LegendRC mirrors legend layout rcParams. Dimension values are expressed in
// legend-font-size units, as in Matplotlib. TitleFontSize is zero when
// legend.title_fontsize is None, which makes legend titles use FontSize.
type LegendRC struct {
	Location      string
	FancyBox      bool
	Shadow        bool
	NumPoints     int
	ScatterPoints int
	MarkerScale   float64
	TitleFontSize float64
	BorderPad     float64
	LabelSpacing  float64
	HandleLength  float64
	HandleHeight  float64
	HandleTextPad float64
	BorderAxesPad float64
	ColumnSpacing float64
}

// ScatterRC mirrors Matplotlib's scatter.* rcParams.
type ScatterRC struct {
	// Marker is the default scatter marker (scatter.marker). Consumed by
	// Axes.Scatter.
	Marker string
	// EdgeColors is the default marker edge color (scatter.edgecolors):
	// "face" (match the face color), "none", or a color. Consumed by
	// Axes.Scatter.
	EdgeColors string
}

// ErrorbarRC mirrors Matplotlib's errorbar.* rcParams.
type ErrorbarRC struct {
	// CapSize is the default error bar cap length in points
	// (errorbar.capsize). Consumed by Axes.ErrorBar.
	CapSize float64
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
	// (image.origin): "upper" or "lower". Consumed by the imshow
	// front-ends (core/matrix_helpers.go).
	Origin string
	// Aspect is the default axes aspect for images (image.aspect):
	// "equal", "auto", or a numeric ratio. Consumed by the imshow
	// front-ends (core/matrix_helpers.go).
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

// ContourRC mirrors Matplotlib's contour.* rcParams.
type ContourRC struct {
	// Algorithm selects the structured contour generator: "mpl2005",
	// "mpl2014", "serial", or "threaded".
	Algorithm string
	// CornerMask retains the valid triangular portion of a structured grid
	// cell containing exactly one masked/non-finite corner.
	CornerMask bool
	// LineWidth is the contour line width in points when LineWidthSet is true.
	// When unset, contour lines inherit RC.LineWidth (lines.linewidth).
	LineWidth    float64
	LineWidthSet bool
	// NegativeLineStyle is the linestyle for negative levels in monochrome
	// line contours when no explicit per-call linestyle is supplied.
	NegativeLineStyle string
}

// BoxplotRC mirrors the wired subset of Matplotlib's boxplot.* rcParams.
type BoxplotRC struct {
	// Notch draws notched boxes (boxplot.notch).
	Notch bool
	// Vertical orients boxes vertically (boxplot.vertical). Stored only:
	// BoxPlot2D is always vertical (see unhonoredRCParams).
	Vertical bool
	// Whiskers sets the whisker length convention (boxplot.whiskers). Stored only:
	// the whisker length is not yet configurable (see unhonoredRCParams).
	Whiskers float64
	// PatchArtist draws boxes as filled patches (boxplot.patchartist).
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
	Default string
	// Fallback is the fallback font set (mathtext.fallback); empty means None.
	Fallback string
	// BF is the bold font pattern (mathtext.bf).
	BF string
	// BFit is the bold-italic font pattern (mathtext.bfit).
	BFit string
	// Cal is the calligraphic font pattern (mathtext.cal).
	Cal string
	// It is the italic font pattern (mathtext.it).
	It string
	// RM is the roman font pattern (mathtext.rm).
	RM string
	// SF is the sans-serif font pattern (mathtext.sf).
	SF string
	// TT is the typewriter font pattern (mathtext.tt).
	TT string
}

// DateRC mirrors Matplotlib's date.* rcParams. Auto* formatters are stored as
// strftime strings; non-default values override AutoDateFormatter's built-in
// layout for their resolution bucket (dates/date_tick.go, strftime-interpreted).
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
	// Epoch is the date epoch (date.epoch), resolved lazily by the first
	// date conversion (dates.Epoch mirrors matplotlib get_epoch).
	Epoch string
	// Converter selects the date converter (date.converter): "auto" or
	// "concise". Consumed by the date axis default formatter (core/units.go).
	Converter string
	// IntervalMultiples snaps ticks to interval multiples
	// (date.interval_multiples). Consumed by dates.DateLocator.
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

// HistBinsAuto is the RC.HistBins sentinel for matplotlib's hist.bins: auto
// (numpy's 'auto' bin estimator).
const HistBinsAuto = -1

// Default contains the library defaults. Copy and apply options to customize.
var Default = RC{
	DPI:                   100,
	FontKey:               "DejaVu Sans",
	FontSize:              10,
	LineWidth:             1.5,
	TextColor:             [4]float64{0, 0, 0, 1},
	LineColor:             [4]float64{0, 0, 0, 1},
	Background:            [4]float64{1, 1, 1, 1},
	TickCountX:            5,
	TickCountY:            5,
	FigureWidth:           6.4,
	FigureHeight:          4.8,
	UseTeX:                false,
	HistBins:              10,
	PathSimplify:          true,
	PathSimplifyThreshold: 1.0 / 9.0,
	AggPathChunkSize:      0,
	AxesBackground:        render.Color{R: 1, G: 1, B: 1, A: 1},
	AxesEdgeColor:         render.Color{R: 0, G: 0, B: 0, A: 1},
	AxesTitleColor:        render.Color{R: 0, G: 0, B: 0, A: 1},
	AxesLabelColor:        render.Color{R: 0, G: 0, B: 0, A: 1},
	AxisLineWidth:         0.8,
	XTickColor:            render.Color{R: 0, G: 0, B: 0, A: 1},
	YTickColor:            render.Color{R: 0, G: 0, B: 0, A: 1},
	TitleFontSize:         12,
	AxisLabelFontSize:     10,
	XTickLabelFontSize:    10,
	YTickLabelFontSize:    10,
	GridColor:             render.Color{R: 0xb0 / 255.0, G: 0xb0 / 255.0, B: 0xb0 / 255.0, A: 1},
	MinorGridColor:        render.Color{R: 0xb0 / 255.0, G: 0xb0 / 255.0, B: 0xb0 / 255.0, A: 1},
	GridLineWidth:         0.8,
	MinorGridLineWidth:    0.8,
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
	Axes: AxesRC{
		AxisBelow:     "line",
		XMargin:       0.05,
		YMargin:       0.05,
		AutolimitMode: "data",
		UnicodeMinus:  true,
		Formatter: FormatterRC{
			Limits:          [2]int{-5, 6},
			MinExponent:     0,
			OffsetThreshold: 4,
			UseLocale:       false,
			UseMathText:     false,
			UseOffset:       true,
		},
		Spines:        SpineRC{Top: true, Bottom: true, Left: true, Right: true},
		TitleLocation: "center",
		TitlePad:      6,
		TitleWeight:   400,
		LabelPad:      4,
		LabelWeight:   400,
	},
	Lines: LinesRC{
		LineStyle:       "-",
		Marker:          "None",
		MarkerSize:      6,
		MarkerEdgeWidth: 1,
		DashedPattern:   []float64{3.7, 1.6},
		DashDotPattern:  []float64{6.4, 1.6, 1, 1.6},
		DottedPattern:   []float64{1, 1.65},
		ScaleDashes:     true,
		DashCap:         render.CapButt,
		DashJoin:        render.JoinRound,
		SolidCap:        render.CapSquare,
		SolidJoin:       render.JoinRound,
		MarkerFaceColor: MarkerColorRC{Mode: MarkerColorAuto},
		MarkerEdgeColor: MarkerColorRC{Mode: MarkerColorAuto},
		MarkerFillStyle: MarkerFillFull,
		Antialiased:     true,
	},
	Patch: PatchRC{
		LineWidth:      1,
		FaceColor:      render.Color{R: 0x1f / 255.0, G: 0x77 / 255.0, B: 0xb4 / 255.0, A: 1},
		FaceColorRaw:   "C0",
		EdgeColor:      render.Color{R: 0, G: 0, B: 0, A: 1},
		ForceEdgeColor: false,
		Antialiased:    true,
	},
	Font: FontRC{
		Family:  []string{"sans-serif"},
		Style:   render.FontStyleNormal,
		Variant: "normal",
		Weight:  400,
		Stretch: "normal",
		Serif: []string{
			"DejaVu Serif", "Bitstream Vera Serif", "Computer Modern Roman",
			"New Century Schoolbook", "Century Schoolbook L", "Utopia",
			"ITC Bookman", "Bookman", "Nimbus Roman No9 L", "Times New Roman",
			"Times", "Palatino", "Charter", "serif",
		},
		SansSerif: []string{
			"DejaVu Sans", "Bitstream Vera Sans", "Computer Modern Sans Serif",
			"Lucida Grande", "Verdana", "Geneva", "Lucid", "Arial", "Helvetica",
			"Avant Garde", "sans-serif",
		},
		Cursive: []string{
			"Apple Chancery", "Textile", "Zapf Chancery", "Sand", "Script MT",
			"Felipa", "Comic Neue", "Comic Sans MS", "cursive",
		},
		Fantasy: []string{
			"Chicago", "Charcoal", "Impact", "Western", "xkcd script", "fantasy",
		},
		Monospace: []string{
			"DejaVu Sans Mono", "Bitstream Vera Sans Mono",
			"Computer Modern Typewriter", "Andale Mono", "Nimbus Mono L",
			"Courier New", "Courier", "Fixed", "Terminal", "monospace",
		},
	},
	XTick: TickAxisRC{
		Direction:    "out",
		Alignment:    "center",
		Primary:      true,
		LabelPrimary: true,
		Major:        TickLevelRC{Size: 3.5, Width: 0.8, Pad: 3.5, Primary: true, Secondary: true, Visible: true},
		Minor:        TickLevelRC{Size: 2.0, Width: 0.6, Pad: 3.4, Primary: true, Secondary: true},
	},
	YTick: TickAxisRC{
		Direction:    "out",
		Alignment:    "center_baseline",
		Primary:      true,
		LabelPrimary: true,
		Major:        TickLevelRC{Size: 3.5, Width: 0.8, Pad: 3.5, Primary: true, Secondary: true, Visible: true},
		Minor:        TickLevelRC{Size: 2.0, Width: 0.6, Pad: 3.4, Primary: true, Secondary: true},
	},
	Legend: LegendRC{
		Location:      "best",
		FancyBox:      true,
		Shadow:        false,
		NumPoints:     1,
		ScatterPoints: 1,
		MarkerScale:   1,
		TitleFontSize: 0,
		BorderPad:     0.4,
		LabelSpacing:  0.5,
		HandleLength:  2,
		HandleHeight:  0.7,
		HandleTextPad: 0.8,
		BorderAxesPad: 0.5,
		ColumnSpacing: 2,
	},
	Scatter: ScatterRC{
		Marker:     "o",
		EdgeColors: "face",
	},
	Errorbar: ErrorbarRC{
		CapSize: 0,
	},
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
	Contour: ContourRC{
		Algorithm:         "mpl2014",
		CornerMask:        true,
		NegativeLineStyle: "dashed",
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
	Figure: FigureRC{
		EdgeColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		FrameOn:     true,
		TitleSize:   12,
		TitleWeight: 400,
		LabelSize:   12,
		LabelWeight: 400,
		Subplot: FigureSubplotRC{
			Left: 0.125, Right: 0.9, Bottom: 0.11, Top: 0.88,
			WSpace: 0.2, HSpace: 0.2,
		},
		Constrained: FigureConstrainedLayoutRC{
			HPad: 0.04167, WPad: 0.04167, HSpace: 0.02, WSpace: 0.02,
		},
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
	rc.PropCycle = rc.PropCycle.Clone()
	rc.GridDashes = cloneDashes(rc.GridDashes)
	rc.MinorGridDashes = cloneDashes(rc.MinorGridDashes)
	rc.Lines.DashedPattern = cloneDashes(rc.Lines.DashedPattern)
	rc.Lines.DashDotPattern = cloneDashes(rc.Lines.DashDotPattern)
	rc.Lines.DottedPattern = cloneDashes(rc.Lines.DottedPattern)
	rc.Animation.FFmpegArgs = cloneStrings(rc.Animation.FFmpegArgs)
	rc.Animation.ConvertArgs = cloneStrings(rc.Animation.ConvertArgs)
	rc.Font.Family = cloneStrings(rc.Font.Family)
	rc.Font.Serif = cloneStrings(rc.Font.Serif)
	rc.Font.SansSerif = cloneStrings(rc.Font.SansSerif)
	rc.Font.Cursive = cloneStrings(rc.Font.Cursive)
	rc.Font.Fantasy = cloneStrings(rc.Font.Fantasy)
	rc.Font.Monospace = cloneStrings(rc.Font.Monospace)
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
		props := render.ParseFontProperties(key)
		if len(props.Families) > 0 {
			rc.Font.Family = cloneStrings(props.Families)
		}
		rc.Font.Style = props.Style
		rc.Font.Weight = props.Weight
		if props.Stretch != "" {
			rc.Font.Stretch = props.Stretch
		}
		if props.Variant != "" {
			rc.Font.Variant = props.Variant
		}
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

// WithColorCycle sets the automatic series color palette. It clears any
// multi-property PropCycle so styling reverts to color-only cycling.
func WithColorCycle(palette color.Palette) Option {
	return func(rc *RC) {
		rc.ColorCycle = clonePalette(palette)
		rc.PropCycle = nil
	}
}

// WithPropCycle sets the full axes.prop_cycle. The cycle's color column (if any)
// is also mirrored into ColorCycle so existing color-only consumers and
// RC.Palette stay consistent.
func WithPropCycle(c *cycler.Cycler) Option {
	return func(rc *RC) {
		rc.PropCycle = c.Clone()
		if palette := propCycleColors(c); len(palette) > 0 {
			rc.ColorCycle = palette
		}
	}
}

// propCycleColors extracts the color column of a prop cycle as a palette,
// returning nil when the cycle is absent or carries no color key.
func propCycleColors(c *cycler.Cycler) color.Palette {
	values := c.ByKey("color")
	if len(values) == 0 {
		return nil
	}
	palette := make(color.Palette, 0, len(values))
	for _, v := range values {
		if col, ok := v.(render.Color); ok {
			palette = append(palette, col)
		}
	}
	return palette
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

// FigureTitleSize returns the configured super-title size.
func (rc *RC) FigureTitleSize() float64 {
	if rc == nil {
		return 12
	}
	if rc.Figure.TitleSize > 0 {
		return rc.Figure.TitleSize
	}
	return rc.TitleSize()
}

// FigureLabelSize returns the configured super-label size.
func (rc *RC) FigureLabelSize() float64 {
	if rc == nil {
		return 12
	}
	if rc.Figure.LabelSize > 0 {
		return rc.Figure.LabelSize
	}
	return rc.FigureTitleSize()
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

// Palette returns a copy of the configured automatic color cycle. When a
// multi-property PropCycle is set, its color column takes precedence so the
// palette length matches the prop cycle (e.g. when "*" expands the color
// column), keeping the color index aligned with the prop-cycle rows.
func (rc RC) Palette() color.Palette {
	if palette := propCycleColors(rc.PropCycle); len(palette) > 0 {
		return palette
	}
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
