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

// TickLabelFormatOptions configures ScalarFormatter behavior on the current axes.
//
// Axis accepts "", "both", "x", or "y". Style accepts "", "sci",
// "scientific", or "plain". SciLimits and UseMathText only apply when non-nil.
type TickLabelFormatOptions struct {
	Axis        string
	Style       string
	SciLimits   *[2]int
	UseMathText *bool
}

type registryState struct {
	mu             sync.Mutex
	current        *core.Figure
	figures        []*core.Figure
	currentAxes    map[*core.Figure]*core.Axes
	subplotAxes    map[*core.Figure]map[string]*core.Axes
	managers       map[*core.Figure]canvas.FigureManager
	managerFactory ManagerFactory
	interactive    bool
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

// Axes appends an axes rectangle to the current figure and marks it current.
func Axes(r geom.Rect, opts ...style.Option) *core.Axes {
	fig := GCF()
	ax := fig.AddAxes(r, opts...)
	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, "")
	registry.mu.Unlock()
	return ax
}

// SCA sets the provided registered axes as current.
func SCA(ax *core.Axes) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	fig := registry.figureForAxesLocked(ax)
	if fig == nil {
		return errors.New("pyplot: axes is not registered")
	}
	registry.current = fig
	registry.currentAxes[fig] = ax
	return nil
}

// DelAxes removes the provided axes from its registered figure.
func DelAxes(ax *core.Axes) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	fig := registry.figureForAxesLocked(ax)
	if fig == nil {
		return errors.New("pyplot: axes is not registered")
	}

	filtered := fig.Children[:0]
	for _, child := range fig.Children {
		if child != ax {
			filtered = append(filtered, child)
		}
	}
	fig.Children = filtered
	for key, subplot := range registry.subplotAxes[fig] {
		if subplot == ax {
			delete(registry.subplotAxes[fig], key)
		}
	}
	if registry.currentAxes[fig] == ax {
		delete(registry.currentAxes, fig)
		if len(fig.Children) > 0 {
			registry.currentAxes[fig] = fig.Children[len(fig.Children)-1]
		}
	}
	registry.current = fig
	return nil
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

// SubplotsAdjust applies persistent subplot layout adjustments to the current figure.
func SubplotsAdjust(cfg core.SubplotAdjust) {
	GCF().SubplotsAdjust(cfg)
}

// Subplot2Grid creates a spanning subplot inside a logical grid on the current figure.
func Subplot2Grid(shape, loc [2]int, rowSpan, colSpan int, opts ...core.SubplotAxesOption) *core.Axes {
	fig := GCF()
	ax := fig.Subplot2Grid(shape, loc, rowSpan, colSpan, opts...)
	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, "")
	registry.mu.Unlock()
	return ax
}

// SubplotMosaic creates named subplot areas on the current figure.
func SubplotMosaic(layout [][]string, opts ...core.GridSpecOption) (map[string]*core.Axes, error) {
	fig := GCF()
	axes, err := fig.SubplotMosaic(layout, opts...)
	if err != nil {
		return nil, err
	}
	registry.mu.Lock()
	var firstAxes *core.Axes
	for row := range layout {
		for _, label := range layout[row] {
			if label == "" || label == "." {
				continue
			}
			ax := axes[label]
			if ax == nil {
				continue
			}
			registry.rememberAxesLocked(fig, ax, "")
			if firstAxes == nil {
				firstAxes = ax
			}
		}
	}
	if firstAxes != nil {
		registry.current = fig
		registry.currentAxes[fig] = firstAxes
	}
	registry.mu.Unlock()
	return axes, nil
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

// Step delegates to the current axes.
func Step(x, y []float64, opts ...core.StepOptions) *core.Line2D {
	return GCA().Step(x, y, opts...)
}

// Stairs delegates to the current axes.
func Stairs(values, edges []float64, opts ...core.StairsOptions) *core.Stairs2D {
	return GCA().Stairs(values, edges, opts...)
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

// AutoScale updates current axes limits from artist bounds.
func AutoScale(margin float64) {
	GCA().AutoScale(margin)
}

// XScale sets the current axes x-axis scale.
func XScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetXScale(name, opts...)
}

// YScale sets the current axes y-axis scale.
func YScale(name string, opts ...transform.ScaleOption) error {
	return GCA().SetYScale(name, opts...)
}

// Axis applies a supported Matplotlib-style axis mode to the current axes.
func Axis(mode string) error {
	ax := GCA()
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "off":
		show := false
		ax.ShowFrame = false
		return ax.TickParams(core.TickParams{
			Axis:       "both",
			ShowTicks:  &show,
			ShowLabels: &show,
		})
	case "on":
		show := true
		ax.ShowFrame = true
		return ax.TickParams(core.TickParams{
			Axis:       "both",
			ShowTicks:  &show,
			ShowLabels: &show,
		})
	case "equal":
		return ax.SetAspect("equal")
	case "", "auto":
		return ax.SetAspect("auto")
	default:
		return fmt.Errorf("pyplot: unsupported axis mode %q", mode)
	}
}

// Grid shows or hides grid lines on the current axes, creating grid artists as needed.
func Grid(visible bool, params ...core.TickParams) ([]*core.Grid, error) {
	ax := GCA()
	tickParams := core.TickParams{
		Axis:  "both",
		Which: "major",
	}
	if len(params) > 0 {
		tickParams = params[0]
		if tickParams.Axis == "" {
			tickParams.Axis = "both"
		}
		if tickParams.Which == "" {
			tickParams.Which = "major"
		}
	}
	tickParams.GridVisible = &visible

	grids, err := ensureGridArtists(ax, tickParams.Axis)
	if err != nil {
		return nil, err
	}
	if err := ax.TickParams(tickParams); err != nil {
		return nil, err
	}
	return grids, nil
}

// TickParams applies tick visibility and styling options to the current axes.
func TickParams(params core.TickParams) error {
	return GCA().TickParams(params)
}

// LocatorParams applies tick locator density options to the current axes.
func LocatorParams(params core.LocatorParams) error {
	return GCA().LocatorParams(params)
}

// MinorTicksOn enables minor ticks on the selected current-axes side.
func MinorTicksOn(axis string) error {
	return GCA().MinorticksOn(axis)
}

// MinorTicksOff disables minor ticks on the selected current-axes side.
func MinorTicksOff(axis string) error {
	return GCA().MinorticksOff(axis)
}

// XTicks sets fixed x-axis tick locations and optional labels on the current axes.
func XTicks(ticks []float64, labels ...[]string) error {
	return setFixedTicks(GCA().XAxis, "x", ticks, labels...)
}

// YTicks sets fixed y-axis tick locations and optional labels on the current axes.
func YTicks(ticks []float64, labels ...[]string) error {
	return setFixedTicks(GCA().YAxis, "y", ticks, labels...)
}

// TickLabelFormat applies ScalarFormatter options to current axes tick labels.
func TickLabelFormat(opts TickLabelFormatOptions) error {
	axes, err := tickLabelFormatAxes(GCA(), opts.Axis)
	if err != nil {
		return err
	}
	formatters := make([]core.ScalarFormatter, len(axes))
	for i, target := range axes {
		formatter, ok := target.axis.Formatter.(core.ScalarFormatter)
		if !ok {
			return fmt.Errorf("pyplot: %s-axis formatter is %T, want core.ScalarFormatter", target.name, target.axis.Formatter)
		}
		formatter, err = applyTickLabelFormat(formatter, opts)
		if err != nil {
			return err
		}
		formatters[i] = formatter
	}
	for i, target := range axes {
		target.axis.Formatter = formatters[i]
	}
	return nil
}

// Bar delegates to the current axes.
func Bar(x, heights []float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().Bar(x, heights, opts...)
}

// BarH delegates to the current axes.
func BarH(y, widths []float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().BarH(y, widths, opts...)
}

// BrokenBarH delegates to the current axes.
func BrokenBarH(xRanges [][2]float64, yRange [2]float64, opts ...core.BarOptions) *core.Bar2D {
	return GCA().BrokenBarH(xRanges, yRange, opts...)
}

// BarLabel delegates to the current axes.
func BarLabel(bar *core.Bar2D, labels []string, opts ...core.BarLabelOptions) []*core.Text {
	return GCA().BarLabel(bar, labels, opts...)
}

// FillBetween delegates to the current axes.
func FillBetween(x, y1, y2 []float64, opts ...core.FillOptions) *core.Fill2D {
	return GCA().FillBetween(x, y1, y2, opts...)
}

// FillBetweenX delegates to the current axes.
func FillBetweenX(y, x1, x2 []float64, opts ...core.FillOptions) *core.Fill2D {
	return GCA().FillBetweenX(y, x1, x2, opts...)
}

// Fill delegates to the current axes.
func Fill(x, y []float64, opts ...core.FillOptions) *core.PolyCollection {
	return GCA().Fill(x, y, opts...)
}

// Hist delegates to the current axes.
func Hist(data []float64, opts ...core.HistOptions) *core.Hist2D {
	return GCA().Hist(data, opts...)
}

// BoxPlot delegates to the current axes.
func BoxPlot(data []float64, opts ...core.BoxPlotOptions) *core.BoxPlot2D {
	return GCA().BoxPlot(data, opts...)
}

// StackPlot delegates to the current axes.
func StackPlot(x []float64, ys [][]float64, opts ...core.StackPlotOptions) []*core.Fill2D {
	return GCA().StackPlot(x, ys, opts...)
}

// ECDF delegates to the current axes.
func ECDF(data []float64, opts ...core.ECDFOptions) *core.Line2D {
	return GCA().ECDF(data, opts...)
}

// ErrorBar delegates to the current axes.
func ErrorBar(x, y, xErr, yErr []float64, opts ...core.ErrorBarOptions) *core.ErrorBar {
	return GCA().ErrorBar(x, y, xErr, yErr, opts...)
}

// Arrow adds a FancyArrow artist to the current axes.
func Arrow(x, y, dx, dy float64, opts ...core.Arrow) *core.Arrow {
	arrow := core.Arrow{
		XY:     geom.Pt{X: x, Y: y},
		DX:     dx,
		DY:     dy,
		Coords: core.Coords(core.CoordData),
	}
	if len(opts) > 0 {
		arrow = opts[0]
		arrow.XY = geom.Pt{X: x, Y: y}
		arrow.DX = dx
		arrow.DY = dy
		if arrow.Coords == (core.CoordinateSpec{}) {
			arrow.Coords = core.Coords(core.CoordData)
		}
	}
	GCA().Add(&arrow)
	return &arrow
}

// HLines adds horizontal line segments to the current axes.
func HLines(y, xMin, xMax []float64, opts ...core.LineCollection) *core.LineCollection {
	if len(y) == 0 || len(y) != len(xMin) || len(y) != len(xMax) {
		return nil
	}
	segments := make([][]geom.Pt, len(y))
	for i := range y {
		segments[i] = []geom.Pt{
			{X: xMin[i], Y: y[i]},
			{X: xMax[i], Y: y[i]},
		}
	}
	return addLineCollection(segments, opts...)
}

// VLines adds vertical line segments to the current axes.
func VLines(x, yMin, yMax []float64, opts ...core.LineCollection) *core.LineCollection {
	if len(x) == 0 || len(x) != len(yMin) || len(x) != len(yMax) {
		return nil
	}
	segments := make([][]geom.Pt, len(x))
	for i := range x {
		segments[i] = []geom.Pt{
			{X: x[i], Y: yMin[i]},
			{X: x[i], Y: yMax[i]},
		}
	}
	return addLineCollection(segments, opts...)
}

// Stem delegates to the current axes.
func Stem(x, y []float64, opts ...core.StemOptions) *core.StemContainer {
	return GCA().Stem(x, y, opts...)
}

// ImRead decodes an image file into renderer-facing image data.
func ImRead(path string) (*render.ImageData, error) {
	return core.ImRead(path)
}

// ImSave writes image data to disk through the core image IO helper.
func ImSave(path string, img render.Image) error {
	return core.ImSave(path, img)
}

// Image delegates to the current axes.
func Image(data [][]float64, opts ...core.ImageOptions) *core.Image2D {
	return GCA().Image(data, opts...)
}

// ImShow delegates to the current axes.
func ImShow(data [][]float64, opts ...core.ImShowOptions) *core.Image2D {
	return GCA().ImShow(data, opts...)
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

// FigText adds text in figure coordinates to the current figure.
func FigText(x, y float64, text string, opts ...core.TextOptions) *core.Text {
	return GCF().Text(x, y, text, opts...)
}

// Suptitle sets the current figure-level title.
func Suptitle(label string) {
	GCF().SetSuptitle(label)
}

// SupXLabel sets the current figure-level x label.
func SupXLabel(label string) {
	GCF().SetSupXLabel(label)
}

// SupYLabel sets the current figure-level y label.
func SupYLabel(label string) {
	GCF().SetSupYLabel(label)
}

// Box toggles the current axes frame.
func Box(on bool) {
	GCA().ShowFrame = on
}

// TwinX creates a current overlay axes sharing the current axes x-scale.
func TwinX() *core.Axes {
	return setCurrentAxes(GCA().TwinX())
}

// TwinY creates a current overlay axes sharing the current axes y-scale.
func TwinY() *core.Axes {
	return setCurrentAxes(GCA().TwinY())
}

// Legend adds a legend to the current axes.
func Legend() *core.Legend {
	return GCA().AddLegend()
}

// FigLegend adds a figure-level legend to the current figure.
func FigLegend() *core.Legend {
	return GCF().AddLegend()
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

// TightLayout enables tight layout on the current figure.
func TightLayout() {
	GCF().TightLayout()
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

// IsInteractive reports whether pyplot interactive mode is enabled.
func IsInteractive() bool {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	return registry.interactive
}

// Ion enables pyplot interactive mode and returns a restore function for the
// previous state.
func Ion() func() {
	return setInteractive(true)
}

// Ioff disables pyplot interactive mode and returns a restore function for the
// previous state.
func Ioff() func() {
	return setInteractive(false)
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

// GetCurrentFigManager returns the cached or newly-created manager for the current figure.
func GetCurrentFigManager() (canvas.FigureManager, error) {
	return ensureManager(GCF())
}

// Draw redraws the current figure through its manager canvas.
func Draw() error {
	manager, err := ensureManager(GCF())
	if err != nil {
		return err
	}
	c := manager.Canvas()
	if c == nil {
		return nil
	}
	if idle, ok := c.(canvas.DrawIdleCanvas); ok {
		return idle.DrawIdle()
	}
	return c.Draw()
}

// DrawIfInteractive redraws the current figure only when interactive mode is enabled.
func DrawIfInteractive() error {
	if !IsInteractive() {
		return nil
	}
	return Draw()
}

// Connect registers an event handler on the current figure canvas.
func Connect(eventType canvas.EventType, handler canvas.Handler) (canvas.ConnectionID, error) {
	manager, err := ensureManager(GCF())
	if err != nil {
		return 0, err
	}
	return canvas.Connect(manager.Canvas(), eventType, handler), nil
}

// Disconnect removes an event handler from the current figure canvas.
func Disconnect(id canvas.ConnectionID) error {
	manager, err := ensureManager(GCF())
	if err != nil {
		return err
	}
	canvas.Disconnect(manager.Canvas(), id)
	return nil
}

// Close removes the given figures from pyplot state and closes their cached
// managers. With no arguments, Close closes the current figure.
func Close(figs ...*core.Figure) error {
	registry.mu.Lock()
	if len(figs) == 0 {
		if registry.current == nil {
			registry.mu.Unlock()
			return nil
		}
		figs = []*core.Figure{registry.current}
	}

	targets := make(map[*core.Figure]struct{}, len(figs))
	managers := make([]canvas.FigureManager, 0, len(figs))
	for _, fig := range figs {
		if fig == nil {
			continue
		}
		if _, seen := targets[fig]; seen {
			continue
		}
		targets[fig] = struct{}{}
		if manager := registry.managers[fig]; manager != nil {
			managers = append(managers, manager)
		}
		delete(registry.managers, fig)
		delete(registry.currentAxes, fig)
		delete(registry.subplotAxes, fig)
	}
	if len(targets) > 0 {
		filtered := registry.figures[:0]
		for _, fig := range registry.figures {
			if _, closeFig := targets[fig]; closeFig {
				continue
			}
			filtered = append(filtered, fig)
		}
		registry.figures = filtered
		if _, closedCurrent := targets[registry.current]; closedCurrent {
			registry.current = nil
			if len(registry.figures) > 0 {
				registry.current = registry.figures[len(registry.figures)-1]
			}
		}
	}
	registry.mu.Unlock()

	var err error
	for _, manager := range managers {
		if manager == nil {
			continue
		}
		err = errors.Join(err, manager.Close())
	}
	return err
}

// CloseAll removes every registered pyplot figure and closes cached managers.
func CloseAll() error {
	registry.mu.Lock()
	figures := append([]*core.Figure(nil), registry.figures...)
	registry.mu.Unlock()
	return Close(figures...)
}

// CLF clears the current figure while keeping it registered and current.
func CLF() {
	fig := GCF()
	if fig == nil {
		return
	}

	fig.Children = nil
	fig.Artists = nil
	fig.SupTitle = ""
	fig.SupXLabel = ""
	fig.SupYLabel = ""

	registry.mu.Lock()
	delete(registry.currentAxes, fig)
	delete(registry.subplotAxes, fig)
	registry.current = fig
	registry.mu.Unlock()
}

// CLA clears the current axes while keeping it current.
func CLA() {
	clearAxes(GCA())
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

func clearAxes(ax *core.Axes) {
	if ax == nil {
		return
	}
	ax.Artists = nil
	ax.WidgetArtists = nil
	ax.Title = ""
	ax.XLabel = ""
	ax.YLabel = ""
	ax.XScale = transform.NewLinear(0, 1)
	ax.YScale = transform.NewLinear(0, 1)
	ax.XAxis = core.NewXAxis()
	ax.YAxis = core.NewYAxis()
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ExtraAxes = nil
	ax.ShowFrame = true
}

func ensureGridArtists(ax *core.Axes, axisSpec string) ([]*core.Grid, error) {
	sides, err := gridAxisSides(axisSpec)
	if err != nil {
		return nil, err
	}

	grids := make([]*core.Grid, 0, len(sides))
	for _, side := range sides {
		if grid := findGridArtist(ax, side); grid != nil {
			grids = append(grids, grid)
			continue
		}
		grids = append(grids, ax.AddGrid(side))
	}
	return grids, nil
}

func findGridArtist(ax *core.Axes, side core.AxisSide) *core.Grid {
	if ax == nil {
		return nil
	}
	for _, artist := range ax.Artists {
		grid, ok := artist.(*core.Grid)
		if ok && grid.Axis == side {
			return grid
		}
	}
	return nil
}

func gridAxisSides(axisSpec string) ([]core.AxisSide, error) {
	switch strings.ToLower(strings.TrimSpace(axisSpec)) {
	case "", "both":
		return []core.AxisSide{core.AxisBottom, core.AxisLeft}, nil
	case "x", "bottom":
		return []core.AxisSide{core.AxisBottom}, nil
	case "top":
		return []core.AxisSide{core.AxisTop}, nil
	case "y", "left":
		return []core.AxisSide{core.AxisLeft}, nil
	case "right":
		return []core.AxisSide{core.AxisRight}, nil
	default:
		return nil, fmt.Errorf("pyplot: unsupported grid axis %q", axisSpec)
	}
}

type tickLabelFormatTarget struct {
	name string
	axis *core.Axis
}

func tickLabelFormatAxes(ax *core.Axes, axisSpec string) ([]tickLabelFormatTarget, error) {
	switch strings.ToLower(strings.TrimSpace(axisSpec)) {
	case "", "both":
		return []tickLabelFormatTarget{
			{name: "x", axis: ax.XAxis},
			{name: "y", axis: ax.YAxis},
		}, nil
	case "x":
		return []tickLabelFormatTarget{{name: "x", axis: ax.XAxis}}, nil
	case "y":
		return []tickLabelFormatTarget{{name: "y", axis: ax.YAxis}}, nil
	default:
		return nil, fmt.Errorf("pyplot: unsupported ticklabel_format axis %q", axisSpec)
	}
}

func applyTickLabelFormat(formatter core.ScalarFormatter, opts TickLabelFormatOptions) (core.ScalarFormatter, error) {
	switch strings.ToLower(strings.TrimSpace(opts.Style)) {
	case "":
	case "sci", "scientific":
		formatter.DisableScientific = false
	case "plain":
		formatter.DisableScientific = true
	default:
		return formatter, fmt.Errorf("pyplot: unsupported ticklabel_format style %q", opts.Style)
	}
	if opts.SciLimits != nil {
		formatter.UsePowerLimits = true
		formatter.PowerLimits = *opts.SciLimits
	}
	if opts.UseMathText != nil {
		formatter.UseMathText = *opts.UseMathText
	}
	return formatter, nil
}

func setCurrentAxes(ax *core.Axes) *core.Axes {
	if ax == nil {
		return nil
	}
	registry.mu.Lock()
	if fig := registry.figureForAxesLocked(ax); fig != nil {
		registry.rememberAxesLocked(fig, ax, "")
	}
	registry.mu.Unlock()
	return ax
}

func addLineCollection(segments [][]geom.Pt, opts ...core.LineCollection) *core.LineCollection {
	ax := GCA()
	collection := core.LineCollection{
		Collection: core.Collection{
			Coords: core.Coords(core.CoordData),
			Alpha:  1,
		},
		Segments:  segments,
		Color:     ax.NextColor(),
		LineWidth: 1,
	}
	if len(opts) > 0 {
		collection = opts[0]
		collection.Segments = segments
		if collection.Coords == (core.CoordinateSpec{}) {
			collection.Coords = core.Coords(core.CoordData)
		}
		if collection.Alpha == 0 {
			collection.Alpha = 1
		}
		if collection.Color.A == 0 && len(collection.Colors) == 0 {
			collection.Color = ax.NextColor()
		}
		if collection.LineWidth == 0 && len(collection.LineWidths) == 0 {
			collection.LineWidth = 1
		}
	}
	ax.Add(&collection)
	return &collection
}

func setFixedTicks(axis *core.Axis, name string, ticks []float64, labels ...[]string) error {
	if axis == nil {
		return fmt.Errorf("pyplot: nil %s axis", name)
	}
	if len(labels) > 1 {
		return fmt.Errorf("pyplot: %sticks accepts at most one label set", name)
	}
	copiedTicks := append([]float64(nil), ticks...)
	axis.Locator = core.FixedLocator{TicksList: copiedTicks}
	if len(labels) == 0 {
		return nil
	}
	if len(labels[0]) != len(ticks) {
		return fmt.Errorf("pyplot: %sticks label count %d does not match tick count %d", name, len(labels[0]), len(ticks))
	}
	axis.Formatter = core.FixedFormatter{Labels: append([]string(nil), labels[0]...)}
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

func (r *registryState) figureForAxesLocked(ax *core.Axes) *core.Figure {
	if ax == nil {
		return nil
	}
	for _, fig := range r.figures {
		for _, child := range fig.Children {
			if child == ax {
				return fig
			}
		}
	}
	return nil
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
	registry.interactive = false
	registry.mu.Unlock()
	style.ResetDefaults()
}

func setInteractive(enabled bool) func() {
	registry.mu.Lock()
	previous := registry.interactive
	registry.interactive = enabled
	registry.mu.Unlock()
	return func() {
		registry.mu.Lock()
		registry.interactive = previous
		registry.mu.Unlock()
	}
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
