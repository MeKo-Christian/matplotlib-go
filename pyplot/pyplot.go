package pyplot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	// DefaultFigureWidth matches Matplotlib's default 6.4in figure width at
	// the repository's default 100 DPI.
	DefaultFigureWidth = 640
	// DefaultFigureHeight matches Matplotlib's default 4.8in figure height at
	// the repository's default 100 DPI.
	DefaultFigureHeight = 480
)

// ShowHandler renders or presents a figure when Show or Pause is called.
type ShowHandler func(*core.Figure) error

// ManagerFactory creates a figure manager for pyplot Show/Pause lifecycle calls.
type ManagerFactory func(*core.Figure) (canvas.FigureManager, error)

type registryState struct {
	mu             sync.Mutex
	current        *core.Figure
	figures        []*core.Figure
	currentAxes    map[*core.Figure]*core.Axes
	subplotAxes    map[*core.Figure]map[string]*core.Axes
	managers       map[*core.Figure]canvas.FigureManager
	managerFactory ManagerFactory
}

var registry = registryState{
	currentAxes:    make(map[*core.Figure]*core.Axes),
	subplotAxes:    make(map[*core.Figure]map[string]*core.Axes),
	managers:       make(map[*core.Figure]canvas.FigureManager),
	managerFactory: defaultManagerFactory,
}

// Figure creates a new current figure using Matplotlib-like default dimensions.
func Figure(opts ...style.Option) *core.Figure {
	width, height := style.CurrentDefaults().DefaultFigureSizePx()
	return FigureSized(width, height, opts...)
}

// FigureSized creates a new current figure with explicit pixel dimensions.
func FigureSized(width, height int, opts ...style.Option) *core.Figure {
	if width <= 0 {
		width, _ = style.CurrentDefaults().DefaultFigureSizePx()
	}
	if height <= 0 {
		_, height = style.CurrentDefaults().DefaultFigureSizePx()
	}

	fig := core.NewFigure(width, height, opts...)
	registry.mu.Lock()
	registry.figures = append(registry.figures, fig)
	registry.current = fig
	registry.mu.Unlock()
	return fig
}

// GCF returns the current figure, creating one if necessary.
func GCF() *core.Figure {
	registry.mu.Lock()
	fig := registry.current
	registry.mu.Unlock()
	if fig != nil {
		return fig
	}
	return Figure()
}

// GCA returns the current axes for the current figure, creating a default
// 1x1 subplot if necessary.
func GCA() *core.Axes {
	fig := GCF()

	registry.mu.Lock()
	ax := registry.currentAxes[fig]
	registry.mu.Unlock()
	if ax != nil {
		return ax
	}
	return ensureDefaultAxes(fig)
}

// AddAxes3D appends an Axes3D to the current figure and marks it current.
func AddAxes3D(r geom.Rect, opts ...style.Option) *core.Axes3D {
	fig := GCF()
	ax, err := fig.AddAxes3D(r, opts...)
	if err != nil {
		return nil
	}
	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = ax.Axes
	registry.mu.Unlock()
	return ax
}

// AddAxesDivider creates an internal layout helper for structured axes tiling.
func AddAxesDivider(r geom.Rect, rows, cols int, opts ...core.AxesDividerOption) *core.AxesDivider {
	return GCF().NewAxesDivider(r, rows, cols, opts...)
}

// NewImageGrid creates an image-grid composed via an axes divider.
func NewImageGrid(rows, cols int, r geom.Rect, opts ...core.AxesDividerOption) *core.ImageGrid {
	fig := GCF()
	grid := fig.NewImageGrid(rows, cols, r, opts...)
	if grid == nil || len(grid.Axes) == 0 || len(grid.Axes[0]) == 0 {
		return grid
	}

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = grid.Axes[0][0]
	registry.mu.Unlock()
	return grid
}

// NewRGBAxes creates three synchronized axes for channel-wise RGB workflows.
func NewRGBAxes(r geom.Rect, opts ...core.AxesDividerOption) *core.RGBAxes {
	fig := GCF()
	axes := fig.NewRGBAxes(r, opts...)
	if axes == nil {
		return nil
	}

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = axes.Red
	registry.mu.Unlock()
	return axes
}

// ParasiteAxes creates a multi-view overlay axes over the current axes viewport.
func ParasiteAxes(opts ...core.ParasiteAxesOption) *core.ParasiteAxes {
	ax := GCA()
	parasite := ax.ParasiteAxes(opts...)
	if parasite == nil {
		return nil
	}
	fig := GCF()

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = parasite.Axes
	registry.mu.Unlock()
	return parasite
}

// FloatingXAxis creates an auxiliary x-axis at the requested y data coordinate.
func FloatingXAxis(y float64) *core.AxisArtist {
	return GCA().FloatingXAxis(y)
}

// FloatingYAxis creates an auxiliary y-axis at the requested x data coordinate.
func FloatingYAxis(x float64) *core.AxisArtist {
	return GCA().FloatingYAxis(x)
}

// GCA3D returns the current 3D axes wrapper when the current axes uses a 3D
// projection, or nil otherwise.
func GCA3D() *core.Axes3D {
	ax := GCA()
	if ax == nil {
		return nil
	}
	if name := ax.ProjectionName(); name != "3d" && name != "axes3d" {
		return nil
	}
	return core.NewAxes3D(ax)
}

// Subplot returns the requested subplot axes in the current figure.
func Subplot(nRows, nCols, index int) *core.Axes {
	fig := GCF()
	key := subplotKey(nRows, nCols, index)

	registry.mu.Lock()
	if ax := registry.subplotAxes[fig][key]; ax != nil {
		registry.current = fig
		registry.currentAxes[fig] = ax
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()

	ax := fig.AddSubplot(nRows, nCols, index)
	if ax == nil {
		return nil
	}

	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, key)
	registry.mu.Unlock()
	return ax
}

// Subplots creates a new figure and subplot grid, then marks the first axes as
// current.
func Subplots(nRows, nCols int, opts ...core.SubplotOption) (*core.Figure, [][]*core.Axes) {
	fig := Figure()
	grid := fig.Subplots(nRows, nCols, opts...)
	if len(grid) == 0 || len(grid[0]) == 0 {
		return fig, grid
	}

	registry.mu.Lock()
	for row := range grid {
		for col, ax := range grid[row] {
			registry.rememberAxesLocked(fig, ax, subplotKey(nRows, nCols, row*nCols+col+1))
		}
	}
	registry.current = fig
	registry.currentAxes[fig] = grid[0][0]
	registry.mu.Unlock()

	return fig, grid
}

// Plot delegates to the current axes.
func Plot(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().Plot(x, y, opts...)
}

// PlotDate delegates to the current axes.
func PlotDate(x []time.Time, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().PlotDate(x, y, opts...)
}

// SemilogX delegates to the current axes.
func SemilogX(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().SemilogX(x, y, opts...)
}

// SemilogY delegates to the current axes.
func SemilogY(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().SemilogY(x, y, opts...)
}

// LogLog delegates to the current axes.
func LogLog(x, y []float64, opts ...core.PlotOptions) *core.Line2D {
	return GCA().LogLog(x, y, opts...)
}

// Scatter delegates to the current axes.
func Scatter(x, y []float64, opts ...core.ScatterOptions) *core.Scatter2D {
	return GCA().Scatter(x, y, opts...)
}

// Plot3D delegates to the current 3D axes.
func Plot3D(x, y, z []float64, opts ...core.PlotOptions) *core.Line2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Plot3D(x, y, z, opts...)
}

// Scatter3D delegates to the current 3D axes.
func Scatter3D(x, y, z []float64, opts ...core.ScatterOptions) *core.Scatter2D {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Scatter3D(x, y, z, opts...)
}

// Wireframe delegates to the current 3D axes.
func Wireframe(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Wireframe(x, y, z, opts...)
}

// Surface delegates to the current 3D axes.
func Surface(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Surface(x, y, z, opts...)
}

// Voxel delegates to the current 3D axes.
func Voxel(x, y, z, dx, dy, dz []float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Voxel(x, y, z, dx, dy, dz, opts...)
}

// Trisurf delegates to the current 3D axes.
func Trisurf(tri core.Triangulation, z []float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Trisurf(tri, z, opts...)
}

// Contour3D delegates to the current 3D axes.
func Contour3D(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.LineCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contour(x, y, z, opts...)
}

// Contourf3D delegates to the current 3D axes.
func Contourf3D(x, y []float64, z [][]float64, opts ...core.PlotOptions) *core.PolyCollection {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Contourf(x, y, z, opts...)
}

// Text3D delegates to the current 3D axes.
func Text3D(x, y, z float64, text string, opts ...core.TextOptions) *core.Text {
	ax := GCA3D()
	if ax == nil {
		return nil
	}
	return ax.Text3D(x, y, z, text, opts...)
}

// Text delegates to the current axes.
func Text(x, y float64, text string, opts ...core.TextOptions) *core.Text {
	return GCA().Text(x, y, text, opts...)
}

// Annotate delegates to the current axes.
func Annotate(text string, x, y float64, opts ...core.AnnotationOptions) *core.Annotation {
	return GCA().Annotate(text, x, y, opts...)
}

// AxHLine delegates to the current axes.
func AxHLine(y float64, opts ...core.HLineOptions) *core.Segment2D {
	return GCA().AxHLine(y, opts...)
}

// AxVLine delegates to the current axes.
func AxVLine(x float64, opts ...core.VLineOptions) *core.Segment2D {
	return GCA().AxVLine(x, opts...)
}

// AxLine delegates to the current axes.
func AxLine(p1, p2 geom.Pt, opts ...core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLine(p1, p2, opts...)
}

// AxLineSlope delegates to the current axes.
func AxLineSlope(point geom.Pt, slope float64, opts ...core.ReferenceLineOptions) *core.InfiniteLine2D {
	return GCA().AxLineSlope(point, slope, opts...)
}

// AxHSpan delegates to the current axes.
func AxHSpan(yMin, yMax float64, opts ...core.HSpanOptions) *core.Span2D {
	return GCA().AxHSpan(yMin, yMax, opts...)
}

// AxVSpan delegates to the current axes.
func AxVSpan(xMin, xMax float64, opts ...core.VSpanOptions) *core.Span2D {
	return GCA().AxVSpan(xMin, xMax, opts...)
}

// XLim sets the current axes x-axis limits.
func XLim(minVal, maxVal float64) {
	GCA().SetXLim(minVal, maxVal)
}

// YLim sets the current axes y-axis limits.
func YLim(minVal, maxVal float64) {
	GCA().SetYLim(minVal, maxVal)
}

// XScale sets the current axes x-axis scale.
func XScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetXScale(name, opts...)
}

// YScale sets the current axes y-axis scale.
func YScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetYScale(name, opts...)
}

// Bar delegates to the current axes.
func Bar(x, heights []float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().Bar(x, heights, opts...)
}

// BarH delegates to the current axes.
func BarH(y, widths []float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().BarH(y, widths, opts...)
}

// FillBetween delegates to the current axes.
func FillBetween(x, y1, y2 []float64, opts ...core.FillOptions) *core.Fill2D {
	return GCA().FillBetween(x, y1, y2, opts...)
}

// Fill delegates to the current axes.
func Fill(x, y []float64, opts ...core.FillOptions) *core.PolyCollection {
	return GCA().Fill(x, y, opts...)
}

// Hist delegates to the current axes.
func Hist(data []float64, opts ...core.HistOptions) *core.Hist2D {
	return GCA().Hist(data, opts...)
}

// ErrorBar delegates to the current axes.
func ErrorBar(x, y, xErr, yErr []float64, opts ...core.ErrorBarOptions) *core.ErrorBar {
	return GCA().ErrorBar(x, y, xErr, yErr, opts...)
}

// Stem delegates to the current axes.
func Stem(x, y []float64, opts ...core.StemOptions) *core.StemContainer {
	return GCA().Stem(x, y, opts...)
}

// Image delegates to the current axes.
func Image(data [][]float64, opts ...core.ImageOptions) *core.Image2D {
	return GCA().Image(data, opts...)
}

// MatShow delegates to the current axes.
func MatShow(data [][]float64, opts ...core.MatShowOptions) *core.Image2D {
	return GCA().MatShow(data, opts...)
}

// Spy delegates to the current axes.
func Spy(data [][]float64, opts ...core.SpyOptions) *core.SpyResult {
	return GCA().Spy(data, opts...)
}

// PColor delegates to the current axes.
func PColor(data [][]float64, opts ...core.MeshOptions) *core.QuadMesh {
	return GCA().PColor(data, opts...)
}

// PColorMesh delegates to the current axes.
func PColorMesh(data [][]float64, opts ...core.MeshOptions) *core.QuadMesh {
	return GCA().PColorMesh(data, opts...)
}

// Hist2D delegates to the current axes.
func Hist2D(x, y []float64, opts ...core.Hist2DOptions) *core.Hist2DResult {
	return GCA().Hist2D(x, y, opts...)
}

// Specgram delegates to the current axes.
func Specgram(samples []float64, opts ...core.SpecgramOptions) *core.SpecgramResult {
	return GCA().Specgram(samples, opts...)
}

// PSD delegates to the current axes.
func PSD(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PSD(samples, opts...)
}

// MagnitudeSpectrum delegates to the current axes.
func MagnitudeSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().MagnitudeSpectrum(samples, opts...)
}

// AngleSpectrum delegates to the current axes.
func AngleSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().AngleSpectrum(samples, opts...)
}

// PhaseSpectrum delegates to the current axes.
func PhaseSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PhaseSpectrum(samples, opts...)
}

// CSD delegates to the current axes.
func CSD(x, y []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().CSD(x, y, opts...)
}

// Cohere delegates to the current axes.
func Cohere(x, y []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().Cohere(x, y, opts...)
}

// XCorr delegates to the current axes.
func XCorr(x, y []float64, opts ...core.CorrelationOptions) *core.CorrelationResult {
	return GCA().XCorr(x, y, opts...)
}

// ACorr delegates to the current axes.
func ACorr(x []float64, opts ...core.CorrelationOptions) *core.CorrelationResult {
	return GCA().ACorr(x, opts...)
}

// AnnotatedHeatmap delegates to the current axes.
func AnnotatedHeatmap(data [][]float64, opts ...core.AnnotatedHeatmapOptions) *core.AnnotatedHeatmapResult {
	return GCA().AnnotatedHeatmap(data, opts...)
}

// Eventplot delegates to the current axes.
func Eventplot(positions [][]float64, opts ...core.EventPlotOptions) *core.EventCollection {
	return GCA().Eventplot(positions, opts...)
}

// Hexbin delegates to the current axes.
func Hexbin(x, y []float64, opts ...core.HexbinOptions) *core.HexbinCollection {
	return GCA().Hexbin(x, y, opts...)
}

// Contour delegates to the current axes.
func Contour(data [][]float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().Contour(data, opts...)
}

// Contourf delegates to the current axes.
func Contourf(data [][]float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().Contourf(data, opts...)
}

// TriPlot delegates to the current axes.
func TriPlot(tri core.Triangulation, opts ...core.TriPlotOptions) *core.LineCollection {
	return GCA().TriPlot(tri, opts...)
}

// TriColor delegates to the current axes.
func TriColor(tri core.Triangulation, values []float64, opts ...core.TriColorOptions) *core.PolyCollection {
	return GCA().TriColor(tri, values, opts...)
}

// TriContour delegates to the current axes.
func TriContour(tri core.Triangulation, values []float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().TriContour(tri, values, opts...)
}

// TriContourf delegates to the current axes.
func TriContourf(tri core.Triangulation, values []float64, opts ...core.ContourOptions) *core.ContourSet {
	return GCA().TriContourf(tri, values, opts...)
}

// Quiver delegates to the current axes.
func Quiver(x, y, u, v []float64, opts ...core.QuiverOptions) *core.Quiver {
	return GCA().Quiver(x, y, u, v, opts...)
}

// QuiverGrid delegates to the current axes.
func QuiverGrid(x, y []float64, u, v [][]float64, opts ...core.QuiverOptions) *core.Quiver {
	return GCA().QuiverGrid(x, y, u, v, opts...)
}

// QuiverKey delegates to the current axes.
func QuiverKey(q *core.Quiver, x, y, u float64, label string, opts ...core.QuiverKeyOptions) *core.QuiverKey {
	return GCA().QuiverKey(q, x, y, u, label, opts...)
}

// Barbs delegates to the current axes.
func Barbs(x, y, u, v []float64, opts ...core.BarbsOptions) *core.Barbs {
	return GCA().Barbs(x, y, u, v, opts...)
}

// BarbsGrid delegates to the current axes.
func BarbsGrid(x, y []float64, u, v [][]float64, opts ...core.BarbsOptions) *core.Barbs {
	return GCA().BarbsGrid(x, y, u, v, opts...)
}

// Streamplot delegates to the current axes.
func Streamplot(x, y []float64, u, v [][]float64, opts ...core.StreamplotOptions) *core.StreamplotSet {
	return GCA().Streamplot(x, y, u, v, opts...)
}

// Pie delegates to the current axes.
func Pie(values []float64, opts ...core.PieOptions) *core.PieContainer {
	return GCA().Pie(values, opts...)
}

// PieLabel delegates to the current axes.
func PieLabel(container *core.PieContainer, labels []string, opts ...core.PieLabelOptions) []*core.Text {
	return GCA().PieLabel(container, labels, opts...)
}

// Violinplot delegates to the current axes.
func Violinplot(data [][]float64, opts ...core.ViolinOptions) *core.ViolinContainer {
	return GCA().Violinplot(data, opts...)
}

// Table delegates to the current axes.
func Table(opts ...core.TableOptions) *core.Table {
	return GCA().Table(opts...)
}

// Sankey returns a builder bound to the current axes.
func Sankey(opts ...core.SankeyOptions) *core.Sankey {
	return core.NewSankey(GCA(), opts...)
}

// Title sets the current axes title.
func Title(label string) {
	GCA().SetTitle(label)
}

// XLabel sets the current axes x-axis label.
func XLabel(label string) {
	GCA().SetXLabel(label)
}

// YLabel sets the current axes y-axis label.
func YLabel(label string) {
	GCA().SetYLabel(label)
}

// Legend adds a legend to the current axes.
func Legend() *core.Legend {
	return GCA().AddLegend()
}

// Colorbar adds a figure-level colorbar for the current axes.
func Colorbar(mappable core.ScalarMappable, opts ...core.ColorbarOptions) *core.Axes {
	ax := GCA()
	fig := GCF()
	if cb := fig.AddColorbar(ax, mappable, opts...); cb != nil {
		return cb
	}
	makeColorbarRoom(ax, opts...)
	return fig.AddColorbar(ax, mappable, opts...)
}

// RCParams returns the active rcParam-style defaults.
func RCParams() style.Params {
	return style.CurrentParams()
}

// RC applies rcParam-style overrides to the active defaults. When group is non-empty,
// keys in params are prefixed with "group." unless already fully qualified.
func RC(group string, params style.Params) error {
	_, err := style.UpdateParams(qualifyRCParams(group, params))
	return err
}

// RCContext applies temporary rcParam overrides and returns a restore function.
func RCContext(params style.Params) (func(), error) {
	restore, _, err := style.PushContext(params)
	if err != nil {
		return nil, err
	}
	return restore, nil
}

// RCDefaults restores the active defaults to the library baseline.
func RCDefaults() {
	style.ResetDefaults()
}

// LoadRCFile loads a Matplotlib-style rc file into the active defaults.
func LoadRCFile(path string) error {
	_, err := style.LoadRCFile(path)
	return err
}

// LoadDefaultRCFile loads the first rc file found in the default search path.
func LoadDefaultRCFile() (string, error) {
	path, _, err := style.LoadDefaultRCFile()
	return path, err
}

// Savefig renders the current figure to a file selected by extension.
func Savefig(path string, opts ...render.SaveOption) error {
	return saveFigure(GCF(), path, opts...)
}

// SetManagerFactory overrides how pyplot creates managers for Show and Pause.
// Passing nil restores the default backend-driven manager selection.
func SetManagerFactory(factory ManagerFactory) {
	if factory == nil {
		factory = defaultManagerFactory
	}

	registry.mu.Lock()
	existing := registry.managers
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = factory
	registry.mu.Unlock()

	for _, manager := range existing {
		if manager != nil {
			_ = manager.Close()
		}
	}
}

// SetShowHandler overrides how Show and Pause present figures. Passing nil
// restores the default manager-backed behavior.
func SetShowHandler(handler ShowHandler) {
	if handler == nil {
		SetManagerFactory(nil)
		return
	}
	SetManagerFactory(func(fig *core.Figure) (canvas.FigureManager, error) {
		return newShowHandlerManager(fig, handler), nil
	})
}

// Show renders all registered figures through the configured show handler.
func Show() error {
	registry.mu.Lock()
	figures := append([]*core.Figure(nil), registry.figures...)
	registry.mu.Unlock()

	for _, fig := range figures {
		if fig == nil {
			continue
		}
		manager, err := ensureManager(fig)
		if err != nil {
			return err
		}
		if err := manager.Show(); err != nil {
			return err
		}
	}
	return nil
}

// Pause renders open figures and then blocks for the requested interval.
func Pause(interval time.Duration) error {
	if err := Show(); err != nil {
		return err
	}
	if interval > 0 {
		time.Sleep(interval)
	}
	return nil
}

func ensureDefaultAxes(fig *core.Figure) *core.Axes {
	registry.mu.Lock()
	if ax := registry.currentAxes[fig]; ax != nil {
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()
	return Subplot(1, 1, 1)
}

func subplotKey(nRows, nCols, index int) string {
	return fmt.Sprintf("%d:%d:%d", nRows, nCols, index)
}

func qualifyRCParams(group string, params style.Params) style.Params {
	if len(params) == 0 {
		return nil
	}

	qualified := make(style.Params, len(params))
	group = strings.ToLower(strings.TrimSpace(group))
	for key, value := range params {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if group != "" && !strings.Contains(normalizedKey, ".") {
			normalizedKey = group + "." + normalizedKey
		}
		qualified[normalizedKey] = value
	}
	return qualified
}

func (r *registryState) rememberAxesLocked(fig *core.Figure, ax *core.Axes, key string) {
	if fig == nil || ax == nil {
		return
	}
	if r.subplotAxes[fig] == nil {
		r.subplotAxes[fig] = make(map[string]*core.Axes)
	}
	if key != "" {
		r.subplotAxes[fig][key] = ax
	}
	r.current = fig
	r.currentAxes[fig] = ax
}

func defaultManagerFactory(fig *core.Figure) (canvas.FigureManager, error) {
	manager, _, err := backends.NewManagerFromEnv(rendererConfig(fig), fig, backends.TextCapabilities)
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func saveFigure(fig *core.Figure, path string, opts ...render.SaveOption) error {
	if fig == nil {
		return errors.New("pyplot: nil figure")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fmt.Errorf("pyplot: savefig path %q has no extension", path)
	}

	choice := strings.TrimSpace(os.Getenv("MATPLOTLIB_BACKEND"))
	cfg := rendererConfig(fig)

	backend, err := backends.SelectBackendForExtension(choice, ext, nil)
	if err != nil {
		// If the user pinned a backend that does not handle this extension,
		// retry with auto so the registry picks any backend that can.
		if choice != "" {
			backend, err = backends.SelectBackendForExtension("", ext, nil)
		}
		if err != nil {
			return fmt.Errorf("pyplot: %w", err)
		}
	}

	renderer, err := backends.Create(backend, cfg)
	if err != nil {
		return fmt.Errorf("pyplot: create %s renderer: %w", backend, err)
	}

	saveOptions := render.ResolveSaveOptions(opts...)
	if err := saveOptions.ValidateForExtension(ext); err != nil {
		return fmt.Errorf("pyplot: %w", err)
	}
	if setter, ok := renderer.(render.SVGOptionSetter); ok {
		setter.SetSVGOptions(saveOptions.SVG)
	}
	if setter, ok := renderer.(render.PDFOptionSetter); ok {
		setter.SetPDFOptions(saveOptions.PDF)
	}
	if setter, ok := renderer.(render.PSOptionSetter); ok {
		setter.SetPSOptions(saveOptions.PS)
	}
	if setter, ok := renderer.(render.PGFOptionSetter); ok {
		setter.SetPGFOptions(saveOptions.PGF)
	}
	core.DrawFigure(fig, renderer)

	if err := backends.DefaultRegistry.SaveViaExtension(backend, renderer, path, opts...); err != nil {
		return fmt.Errorf("pyplot: %w", err)
	}
	return nil
}

func rendererConfig(fig *core.Figure) backends.Config {
	defaults := style.CurrentDefaults()
	background := defaults.FigureBackground()
	dpi := defaults.DPI
	width, height := defaults.DefaultFigureSizePx()

	if fig != nil {
		if fig.SizePx.X > 0 {
			width = int(fig.SizePx.X + 0.5)
		}
		if fig.SizePx.Y > 0 {
			height = int(fig.SizePx.Y + 0.5)
		}
		background = fig.RC.FigureBackground()
		dpi = fig.RC.DPI
	}

	return backends.Config{
		Width:      width,
		Height:     height,
		Background: background,
		DPI:        dpi,
	}
}

func makeColorbarRoom(ax *core.Axes, opts ...core.ColorbarOptions) {
	if ax == nil {
		return
	}

	width := 0.035
	padding := 0.02
	if len(opts) > 0 {
		if opts[0].Width > 0 {
			width = opts[0].Width
		}
		if opts[0].Padding > 0 {
			padding = opts[0].Padding
		}
	}

	maxRight := 1 - padding - width
	if ax.RectFraction.Max.X <= maxRight {
		return
	}
	if maxRight <= ax.RectFraction.Min.X {
		return
	}
	ax.RectFraction.Max.X = maxRight
}

func resetForTests() {
	registry.mu.Lock()
	registry.current = nil
	registry.figures = nil
	registry.currentAxes = make(map[*core.Figure]*core.Axes)
	registry.subplotAxes = make(map[*core.Figure]map[string]*core.Axes)
	registry.managers = make(map[*core.Figure]canvas.FigureManager)
	registry.managerFactory = defaultManagerFactory
	registry.mu.Unlock()
	style.ResetDefaults()
}

func ensureManager(fig *core.Figure) (canvas.FigureManager, error) {
	registry.mu.Lock()
	if manager := registry.managers[fig]; manager != nil {
		registry.mu.Unlock()
		return manager, nil
	}
	factory := registry.managerFactory
	registry.mu.Unlock()

	manager, err := factory(fig)
	if err != nil {
		return nil, err
	}

	registry.mu.Lock()
	if existing := registry.managers[fig]; existing != nil {
		registry.mu.Unlock()
		_ = manager.Close()
		return existing, nil
	}
	registry.managers[fig] = manager
	registry.mu.Unlock()
	return manager, nil
}

type showHandlerManager struct {
	canvas *showHandlerCanvas
	tools  *canvas.ToolManager
}

type showHandlerCanvas struct {
	figure     *core.Figure
	handler    ShowHandler
	dispatcher canvas.Dispatcher
	closed     bool
}

func newShowHandlerManager(fig *core.Figure, handler ShowHandler) canvas.FigureManager {
	c := &showHandlerCanvas{figure: fig, handler: handler}
	manager := &showHandlerManager{
		canvas: c,
		tools:  canvas.NewToolManager(),
	}
	manager.tools.Register(canvas.ToolFunc{
		Name: "redraw",
		Run: func(canvas.ToolArgs) error {
			return c.Draw()
		},
	})
	return manager
}

func (m *showHandlerManager) Canvas() canvas.FigureCanvas { return m.canvas }

func (m *showHandlerManager) Show() error { return m.canvas.Draw() }

func (m *showHandlerManager) Close() error { return m.canvas.Close() }

func (m *showHandlerManager) SetTitle(string) {}

func (m *showHandlerManager) ToolManager() *canvas.ToolManager { return m.tools }

func (c *showHandlerCanvas) Figure() *core.Figure { return c.figure }

func (c *showHandlerCanvas) Draw() error {
	if c.closed {
		return nil
	}
	if c.handler == nil {
		return nil
	}
	if err := c.handler(c.figure); err != nil {
		return err
	}
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventDraw,
		Figure: c.figure,
		Width:  int(c.figure.SizePx.X + 0.5),
		Height: int(c.figure.SizePx.Y + 0.5),
	})
}

func (c *showHandlerCanvas) Resize(width, height int) error {
	if c.closed {
		return nil
	}
	if c.figure != nil {
		c.figure.SizePx.X = float64(width)
		c.figure.SizePx.Y = float64(height)
	}
	if err := c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventResize,
		Figure: c.figure,
		Width:  width,
		Height: height,
	}); err != nil {
		return err
	}
	return c.Draw()
}

func (c *showHandlerCanvas) Connect(eventType canvas.EventType, handler canvas.Handler) canvas.ConnectionID {
	return c.dispatcher.Connect(eventType, handler)
}

func (c *showHandlerCanvas) Disconnect(id canvas.ConnectionID) {
	c.dispatcher.Disconnect(id)
}

func (c *showHandlerCanvas) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.dispatcher.Emit(canvas.Event{
		Type:   canvas.EventClose,
		Figure: c.figure,
	})
}
