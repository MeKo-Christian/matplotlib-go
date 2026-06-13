package common

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

type TestDistanceKM float64

type testDistanceConverter struct{}

var registerTestDistanceUnitsOnce sync.Once

func (testDistanceConverter) Convert(value any) (float64, error) {
	v, ok := value.(TestDistanceKM)
	if !ok {
		return 0, fmt.Errorf("unexpected distance value %T", value)
	}
	return float64(v), nil
}

func (testDistanceConverter) AxisInfo([]float64) core.AxisInfo {
	return core.AxisInfo{
		Formatter: core.FormatStrFormatter{Pattern: "%.0f km"},
	}
}

func RegisterTestDistanceUnits() {
	registerTestDistanceUnitsOnce.Do(func() {
		core.MustRegisterUnitConverter(TestDistanceKM(0), func() core.UnitsConverter {
			return testDistanceConverter{}
		})
	})
}

func ReferenceDateNumber(t time.Time) float64 {
	t = t.UTC()
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func ReferencePointsToPixels(Points float64) float64 {
	return Points * style.Default.DPI / 72.0
}

// renderBasicLine creates the same basic line plot as examples/lines/basic.go

// renderJoinsCaps creates a plot demonstrating different line joins and caps

// renderDashes creates a plot demonstrating dash patterns

// renderScatterBasic creates a basic scatter plot for golden testing

// renderScatterMarkerTypes creates a plot showing all marker types

// renderScatterAdvanced creates an advanced scatter plot with edges, alpha, and variable sizes

// PlotBarBasicScaffold builds the bare-bones bar-chart scaffold figure used by
// the bar_basic_* progression cases. Backend-agnostic.
func PlotBarBasicScaffold(showTicks, showTickLabels, showTitle bool) *core.Figure {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 10)

	ax.XAxis.ShowTicks = showTicks
	ax.YAxis.ShowTicks = showTicks
	ax.XAxis.ShowLabels = showTickLabels
	ax.YAxis.ShowLabels = showTickLabels

	if showTitle {
		ax.SetTitle("Basic Bars")
	}

	return fig
}

// RenderBarBasicScaffold AGG-renders PlotBarBasicScaffold.
func RenderBarBasicScaffold(showTicks, showTickLabels, showTitle bool) image.Image {
	return RenderFixtureFigure(PlotBarBasicScaffold(showTicks, showTickLabels, showTitle), 640, 360)
}

// renderBarBasic creates a basic vertical bar chart for golden testing

// renderBarHorizontal creates a horizontal bar chart for golden testing

// renderBarGrouped creates a grouped bar chart with variable colors and edges

// renderFillBasic creates a basic fill to baseline for golden testing

// renderFillBetween creates a fill between two curves for golden testing

// renderFillStacked creates a stacked area chart for golden testing

// renderMultiSeriesBasic creates a plot with multiple series using different plot types

// NormalData generates a seeded normally-distributed sample using Box-Muller.
func NormalData(seed1, seed2 uint64, n int, mean, stddev float64) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	data := make([]float64, n)
	for i := range data {
		u1 := rng.Float64()
		u2 := rng.Float64()
		data[i] = math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)*stddev + mean
	}
	return data
}

// UniformData generates deterministic samples in [low, high) using Go's PCG.
func UniformData(seed1, seed2 uint64, n int, low, high float64) []float64 {
	rng := rand.New(rand.NewPCG(seed1, seed2))
	data := make([]float64, n)
	span := high - low
	for i := range data {
		data[i] = low + rng.Float64()*span
	}
	return data
}

// renderHistBasic creates a basic count histogram for golden testing.

// renderHistDensity creates a density-normalized histogram for golden testing.

// renderHistStrategies creates a histogram comparing two datasets for golden testing.

// renderBoxPlotBasic creates a multi-series box plot for golden testing.

// renderMultiSeriesColorCycle creates a plot demonstrating automatic color cycling

func ApplyMatplotlibGridSpecStyle(fig *core.Figure) {
	fig.RC = style.Apply(fig.RC, style.WithFont("DejaVu Sans", 10))
	fig.RC.TitleFontSize = 12
	fig.RC.AxisLabelFontSize = 10
	fig.RC.XTickLabelFontSize = 10
	fig.RC.YTickLabelFontSize = 10
}

func ConfigureCompositionAxes(ax *core.Axes, title string, x, y []float64, c render.Color) {
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	width := 2.0
	ax.Plot(x, y, core.PlotOptions{
		Color:     &c,
		LineWidth: &width,
		Label:     title,
	})
	ax.AutoScale(0.10)
}

func ConfigureCompositionTicks(ax *core.Axes, xTicks, yTicks []float64, yFormat string) {
	ax.XAxis.Locator = core.FixedLocator{TicksList: xTicks}
	ax.YAxis.Locator = core.FixedLocator{TicksList: yTicks}
	ax.YAxis.Formatter = core.FormatStrFormatter{Pattern: yFormat}
}

func AddReferenceYGrid(ax *core.Axes) {
	ax.SetAxisBelow(true)
	grid := ax.AddGrid(core.AxisLeft)
	grid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	grid.LineWidth = 0.5
}

func AddReferenceXYGrid(ax *core.Axes) {
	ax.SetAxisBelow(true)
	xGrid := ax.AddGrid(core.AxisBottom)
	xGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	xGrid.LineWidth = 0.5
	AddReferenceYGrid(ax)
}

func LambertLongitudeTicks() []float64 {
	ticks := make([]float64, 0, 9)
	for deg := -120; deg <= 120; deg += 30 {
		ticks = append(ticks, float64(deg)*math.Pi/180)
	}
	return ticks
}

// PlainDegreeFormat mirrors the matplotlib reference fixtures' plain degree
// formatter (`lambda x, _: f"{round(x * 180.0 / math.pi):.0f}"`), which
// replaces the geo axes' default ThetaFormatter (degrees with a degree sign).
func PlainDegreeFormat(rad float64) string {
	return fmt.Sprintf("%.0f", math.Round(rad*180/math.Pi))
}

// PlotGeoProjectionAxes builds a geographic-projection figure with a single
// sinusoidal lat/lon trace. Backend-agnostic.
func PlotGeoProjectionAxes(projection, title string, lonMin, lonMax float64) *core.Figure {
	fig := core.NewFigure(720, 420)
	ax, err := fig.AddAxesProjection(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.14},
		Max: geom.Pt{X: 0.92, Y: 0.86},
	}, projection)
	if err != nil {
		panic(err)
	}
	ax.SetTitle(title)
	ax.SetXLabel("longitude")
	ax.SetYLabel("latitude")
	ax.XAxis.Formatter = core.FuncFormatter(PlainDegreeFormat)
	ax.YAxis.Formatter = core.FuncFormatter(PlainDegreeFormat)

	gridColor := render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}
	lonGrid := ax.AddGrid(core.AxisBottom)
	lonGrid.Color = gridColor
	lonGrid.LineWidth = 0.8
	latGrid := ax.AddGrid(core.AxisLeft)
	latGrid.Color = gridColor
	latGrid.LineWidth = 0.8

	const n = 361
	lon := make([]float64, n)
	lat := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		lon[i] = lonMin + (lonMax-lonMin)*t
		lat[i] = 0.35 * math.Sin(3*lon[i])
	}

	lineColor := render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
	lineWidth := 2.0
	ax.Plot(lon, lat, core.PlotOptions{
		Color:     &lineColor,
		LineWidth: &lineWidth,
	})

	return fig
}

// RenderGeoProjectionAxes AGG-renders PlotGeoProjectionAxes.
func RenderGeoProjectionAxes(projection, title string, lonMin, lonMax float64) image.Image {
	return RenderFixtureFigure(PlotGeoProjectionAxes(projection, title, lonMin, lonMax), 720, 420)
}

func GridMin(Values [][]float64) float64 {
	minValue := math.Inf(1)
	for _, row := range Values {
		for _, value := range row {
			if value < minValue {
				minValue = value
			}
		}
	}
	return minValue
}

func SinusoidalTerrain(xCount, yCount int) ([]float64, []float64, [][]float64) {
	if xCount < 2 {
		xCount = 2
	}
	if yCount < 2 {
		yCount = 2
	}
	x := make([]float64, xCount)
	y := make([]float64, yCount)
	z := make([][]float64, yCount)

	for yi := 0; yi < yCount; yi++ {
		y[yi] = -math.Pi + 2*math.Pi*float64(yi)/float64(yCount-1)
	}
	for xi := 0; xi < xCount; xi++ {
		x[xi] = -math.Pi + 2*math.Pi*float64(xi)/float64(xCount-1)
	}
	for yi := 0; yi < yCount; yi++ {
		row := make([]float64, xCount)
		for xi := 0; xi < xCount; xi++ {
			angleX := x[xi]
			angleY := y[yi]
			row[xi] = 0.5*math.Sin(angleX)*math.Cos(angleY) +
				0.35*math.Sin(2*angleX+0.6)*math.Cos(angleY/2) +
				0.15*math.Cos(3*angleY-angleX)
		}
		z[yi] = row
	}
	return x, y, z
}

func DisableMplot3DTickLabels(ax *core.Axes3D) {
	if ax == nil {
		return
	}
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowLabels = false
	ax.SetShowXTickLabels(false)
	ax.SetShowYTickLabels(false)
	ax.SetShowZTickLabels(false)
}

func Get3DWireframeTestData(delta float64) (x []float64, y []float64, z [][]float64) {
	if delta <= 0 {
		delta = 0.05
	}

	count := int(math.Ceil(6.0 / delta))
	x = make([]float64, count)
	for i := range x {
		x[i] = -3.0 + delta*float64(i)
	}
	y = make([]float64, len(x))
	copy(y, x)

	yCount := len(y)
	xCount := len(x)
	z = make([][]float64, yCount)
	for yi := range y {
		yRow := make([]float64, xCount)
		for xi := range x {
			gridX := x[xi]
			gridY := y[yi]
			z1 := math.Exp(-(gridX*gridX+gridY*gridY)/2) / (2 * math.Pi)
			z2 := (math.Exp(-(((gridX-1)/1.5)*((gridX-1)/1.5)+((gridY-1)/0.5)*((gridY-1)/0.5))/2) / (2 * math.Pi * 0.5 * 1.5))
			yRow[xi] = (z2 - z1) * 500
		}
		z[yi] = yRow
	}
	for i := range x {
		x[i] *= 10
		y[i] *= 10
	}

	return x, y, z
}

func MinInSlice(Values []float64) float64 {
	minValue := math.Inf(1)
	for _, value := range Values {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}

func BoolPtr(v bool) *bool {
	return &v
}

func MinInGrid(Values [][]float64) float64 {
	minValue := math.Inf(1)
	for _, row := range Values {
		for _, value := range row {
			if value < minValue {
				minValue = value
			}
		}
	}
	return minValue
}

func FloatPtr(v float64) *float64 {
	return &v
}

func ColorNormFixtureFigure(title string) (*core.Figure, *core.Axes) {
	fig := core.NewFigure(640, 360)
	gs := fig.GridSpec(
		1,
		1,
		core.WithGridSpecPadding(0.125, 0.9, 0.11, 0.88),
		core.WithGridSpecSpacing(0, 0),
	)
	ax := gs.Cell(0, 0).AddAxes()
	ax.SetPosition(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.88}})
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	return fig, ax
}

func LogNormFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			t := float64(row*cols+col) / float64(rows*cols-1)
			data[row][col] = math.Pow(10, 3*t)
		}
	}
	return data
}

func TwoSlopeFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		y := float64(row) / float64(max(1, rows-1))
		for col := range data[row] {
			x := float64(col) / float64(max(1, cols-1))
			data[row][col] = 6*x - 3 + 1.5*math.Sin((y-0.5)*math.Pi)
		}
	}
	return data
}

func MeshFixtureFigure(title string) (*core.Figure, *core.Axes) {
	fig := core.NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.88}})
	ax.SetTitle(title)
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	return fig, ax
}

func MeshFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for yi := range data {
		data[yi] = make([]float64, cols)
		y := float64(yi) / float64(max(1, rows-1))
		for xi := range data[yi] {
			x := float64(xi) / float64(max(1, cols-1))
			data[yi][xi] = 0.65*math.Sin((x*2.3+0.15)*math.Pi) + 0.28*math.Cos((y*2.1-0.25)*math.Pi)
		}
	}
	return data
}

func Hist2DWeightedData() (x, y, weights []float64) {
	x = []float64{-1.8, -1.4, -0.8, -0.3, 0.2, 0.7, 1.1, 1.6, 2.1, 2.5, -0.6, 0.4, 1.3, 2.7}
	y = []float64{-1.1, -0.4, -0.8, 0.1, 0.5, 0.9, 1.2, 1.7, 2.0, 2.2, 0.7, -0.2, 0.4, 1.3}
	weights = []float64{0.8, 1.3, 0.7, 1.1, 1.6, 0.9, 1.4, 1.2, 1.8, 0.6, 1.5, 0.9, 1.1, 1.7}
	return x, y, weights
}

func RenderImageFixture(fig *core.Figure, width, height int) image.Image {
	r, err := agg.New(width, height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func WaveImageData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		yy := float64(row) / float64(max(1, rows-1))
		for col := range data[row] {
			xx := float64(col) / float64(max(1, cols-1))
			data[row][col] = 0.5 + 0.25*math.Sin((xx*2.4+0.15)*math.Pi) + 0.22*math.Cos((yy*2.7-0.2)*math.Pi)
		}
	}
	return data
}

func SparseFixtureData(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for row := range data {
		data[row] = make([]float64, cols)
		for col := range data[row] {
			if row == col || row+col == cols-1 || (col+2*row)%5 == 0 || (2*col+row)%9 == 0 {
				data[row][col] = 1
			}
		}
	}
	return data
}

type GouraudFixtureArtist struct {
	Points    []geom.Pt
	Triangles [][3]int
	Values    []float64
}

func (a *GouraudFixtureArtist) Draw(r render.Renderer, ctx *core.DrawContext) {
	drawer, ok := r.(render.GouraudTriangleDrawer)
	if !ok || ctx == nil || len(a.Points) == 0 {
		return
	}
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	mapping := core.ScalarMapInfo{Colormap: "viridis", VMin: 0, VMax: 1}.Resolved()
	batch := render.GouraudTriangleBatch{Antialiased: true}
	for _, idx := range a.Triangles {
		var tri render.GouraudTriangle
		for i, pointIndex := range idx {
			pt := a.Points[pointIndex]
			if tr != nil {
				pt = tr.Apply(pt)
			}
			tri.P[i] = pt
			tri.Color[i] = mapping.Color(a.Values[pointIndex], 1)
		}
		batch.Triangles = append(batch.Triangles, tri)
	}
	drawer.DrawGouraudTriangles(batch)
}

func (a *GouraudFixtureArtist) Z() float64 { return 0 }

func (a *GouraudFixtureArtist) Bounds(*core.DrawContext) geom.Rect {
	if len(a.Points) == 0 {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: a.Points[0], Max: a.Points[0]}
	for _, pt := range a.Points[1:] {
		bounds.Min.X = math.Min(bounds.Min.X, pt.X)
		bounds.Min.Y = math.Min(bounds.Min.Y, pt.Y)
		bounds.Max.X = math.Max(bounds.Max.X, pt.X)
		bounds.Max.Y = math.Max(bounds.Max.Y, pt.Y)
	}
	return bounds
}

type ClipPathBatchFixtureArtist struct{}

func (a *ClipPathBatchFixtureArtist) Draw(r render.Renderer, ctx *core.DrawContext) {
	if ctx == nil {
		return
	}
	drawer, ok := r.(render.QuadMeshDrawer)
	if !ok {
		return
	}
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	if tr == nil {
		return
	}

	clip := transformFixturePath(clipBatchPath(), tr)
	r.Save()
	r.ClipPath(clip)
	drawer.DrawQuadMesh(clipBatchQuadMesh(ctx))
	r.Restore()

	r.Path(clip, &render.Paint{
		Stroke:    render.Color{R: 0.05, G: 0.08, B: 0.12, A: 1},
		LineWidth: 2.0,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func (a *ClipPathBatchFixtureArtist) Z() float64 { return 0 }

func (a *ClipPathBatchFixtureArtist) Bounds(*core.DrawContext) geom.Rect {
	return geom.Rect{Min: geom.Pt{X: 0.45, Y: 0.45}, Max: geom.Pt{X: 5.55, Y: 5.05}}
}

func clipBatchQuadMesh(ctx *core.DrawContext) render.QuadMeshBatch {
	xEdges := []float64{0, 0.75, 1.5, 2.35, 3.1, 4.0, 4.85, 5.45, 6.0}
	yEdges := []float64{0, 0.7, 1.55, 2.4, 3.2, 4.15, 5.4}
	mapping := core.ScalarMapInfo{Colormap: "viridis", VMin: -0.35, VMax: 1.15}.Resolved()
	tr := ctx.TransformFor(core.Coords(core.CoordData))
	batch := render.QuadMeshBatch{Cells: make([]render.QuadMeshCell, 0, (len(xEdges)-1)*(len(yEdges)-1))}
	for yi := 0; yi+1 < len(yEdges); yi++ {
		for xi := 0; xi+1 < len(xEdges); xi++ {
			cx := (xEdges[xi] + xEdges[xi+1]) * 0.5
			cy := (yEdges[yi] + yEdges[yi+1]) * 0.5
			value := 0.45 + 0.42*math.Sin(cx*1.15) + 0.33*math.Cos(cy*1.35) + 0.06*float64((xi+yi)%3)
			local := [4]geom.Pt{
				{X: xEdges[xi], Y: yEdges[yi]},
				{X: xEdges[xi+1], Y: yEdges[yi]},
				{X: xEdges[xi+1], Y: yEdges[yi+1]},
				{X: xEdges[xi], Y: yEdges[yi+1]},
			}
			var quad [4]geom.Pt
			for i, pt := range local {
				quad[i] = tr.Apply(pt)
			}
			batch.Cells = append(batch.Cells, render.QuadMeshCell{
				Quad:        quad,
				Face:        mapping.Color(value, 0.84),
				Edge:        render.Color{R: 0.97, G: 0.97, B: 0.97, A: 0.72},
				LineWidth:   0.55,
				Antialiased: true,
			})
		}
	}
	return batch
}

func clipBatchPath() geom.Path {
	Points := []geom.Pt{
		{X: 0.55, Y: 1.10},
		{X: 2.05, Y: 0.50},
		{X: 3.10, Y: 1.05},
		{X: 5.35, Y: 0.80},
		{X: 4.70, Y: 2.45},
		{X: 5.50, Y: 4.05},
		{X: 3.70, Y: 3.80},
		{X: 2.55, Y: 5.05},
		{X: 1.75, Y: 3.55},
		{X: 0.55, Y: 3.85},
		{X: 1.20, Y: 2.35},
	}
	path := geom.Path{}
	for i, pt := range Points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func transformFixturePath(path geom.Path, tr interface{ Apply(geom.Pt) geom.Pt }) geom.Path {
	out := geom.Path{C: append([]geom.Cmd(nil), path.C...), V: make([]geom.Pt, len(path.V))}
	for i, pt := range path.V {
		out.V[i] = tr.Apply(pt)
	}
	return out
}

func RenderFixtureFigure(fig *core.Figure, w, h int) image.Image {
	r, err := agg.New(w, h, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func FixtureRectPath(x, y, w, h float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y + h})
	path.LineTo(geom.Pt{X: x, Y: y + h})
	path.Close()
	return path
}

func FixtureTrianglePath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	for i := 0; i < 3; i++ {
		angle := -math.Pi/2 + float64(i)*2*math.Pi/3
		pt := geom.Pt{X: cx + r*math.Cos(angle), Y: cy + r*math.Sin(angle)}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}

func FixtureDiamondPath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: cx, Y: cy + r})
	path.LineTo(geom.Pt{X: cx + r, Y: cy})
	path.LineTo(geom.Pt{X: cx, Y: cy - r})
	path.LineTo(geom.Pt{X: cx - r, Y: cy})
	path.Close()
	return path
}

func FixtureStarPath(cx, cy, r float64) geom.Path {
	path := geom.Path{}
	for i := 0; i < 10; i++ {
		radius := r
		if i%2 == 1 {
			radius = r * 0.45
		}
		angle := -math.Pi/2 + float64(i)*math.Pi/5
		pt := geom.Pt{X: cx + radius*math.Cos(angle), Y: cy + radius*math.Sin(angle)}
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}
