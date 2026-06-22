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
