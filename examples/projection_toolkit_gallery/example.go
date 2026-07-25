// Package projection_toolkit_gallery demonstrates projection and toolkit
// helpers in one catalog-backed showcase.
package projection_toolkit_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	Width  = 1320
	Height = 900
	DPI    = 100
)

var (
	blue   = render.Color{R: 0.14, G: 0.34, B: 0.70, A: 1}
	orange = render.Color{R: 0.74, G: 0.28, B: 0.18, A: 1}
	green  = render.Color{R: 0.05, G: 0.48, B: 0.28, A: 1}
	red    = render.Color{R: 0.78, G: 0.13, B: 0.16, A: 1}
	grid   = render.Color{R: 0.80, G: 0.82, B: 0.86, A: 1}
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	addPolarPanel(fig, panel(0, 0))
	addGeoPanel(fig, panel(0, 1), "mollweide", "Mollweide", -math.Pi, math.Pi)
	addGeoPanel(fig, panel(0, 2), "aitoff", "Aitoff", -math.Pi, math.Pi)
	addGeoPanel(fig, panel(1, 0), "hammer", "Hammer", -math.Pi, math.Pi)
	addGeoPanel(fig, panel(1, 1), "lambert", "Lambert", -math.Pi/2, math.Pi/2)
	addRadarPanel(fig, panel(1, 2))
	addSkewTPanel(fig, panel(2, 0))
	addAxisArtistPanel(fig, panel(2, 1))
	addAxesGridPanel(fig, panel(2, 2))

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func panel(row, col int) geom.Rect {
	const (
		left   = 0.055
		right  = 0.965
		bottom = 0.075
		top    = 0.935
		hgap   = 0.060
		vgap   = 0.095
	)
	w := (right - left - 2*hgap) / 3
	h := (top - bottom - 2*vgap) / 3
	x0 := left + float64(col)*(w+hgap)
	y1 := top - float64(row)*(h+vgap)
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y1 - h},
		Max: geom.Pt{X: x0 + w, Y: y1},
	}
}

func addPolarPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddPolarAxes(rect)
	ax.SetTitle("Polar")
	ax.SetThetaZeroLocation("N")
	if err := ax.SetThetaDirection("-1"); err != nil {
		panic(err)
	}
	ax.SetYLim(0, 1.15)
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0.25, 0.5, 0.75, 1.0}}
	ax.YAxis.Formatter = ticker.PercentFormatter{XMax: 1, Decimals: 0, DecimalsSet: true}
	styleGrid(ax.AddGrid(core.AxisBottom), grid, 0.8)
	styleGrid(ax.AddGrid(core.AxisLeft), grid, 0.8)

	const n = 241
	theta := make([]float64, n)
	r := make([]float64, n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / float64(n-1)
		theta[i] = t
		r[i] = 0.62 + 0.28*math.Sin(3*t+0.35)
	}
	width := 2.0
	fill := render.Color{R: 0.18, G: 0.50, B: 0.82, A: 0.22}
	ax.Fill(theta, r, core.FillOptions{Color: &fill})
	_, _ = ax.Plot(theta, r, core.PlotOptions{Color: &blue, LineWidth: &width})
}

func addGeoPanel(fig *core.Figure, rect geom.Rect, projection, title string, lonMin, lonMax float64) {
	ax, err := fig.AddAxesProjection(rect, projection)
	if err != nil {
		panic(err)
	}
	ax.SetTitle(title)
	ax.SetXLabel("lon")
	ax.SetYLabel("lat")
	if projection == "lambert" {
		ax.XAxis.Locator = ticker.FixedLocator{TicksList: common.LambertLongitudeTicks()}
	}
	styleGrid(ax.AddGrid(core.AxisBottom), grid, 0.7)
	styleGrid(ax.AddGrid(core.AxisLeft), grid, 0.7)

	const n = 241
	lon := make([]float64, n)
	lat := make([]float64, n)
	for i := range n {
		t := float64(i) / float64(n-1)
		lon[i] = lonMin + (lonMax-lonMin)*t
		lat[i] = 0.35 * math.Sin(3*lon[i])
	}
	width := 1.8
	_, _ = ax.Plot(lon, lat, core.PlotOptions{Color: &blue, LineWidth: &width})
}

func addRadarPanel(fig *core.Figure, rect geom.Rect) {
	labels := []string{"Speed", "Power", "Range", "Handling", "Comfort"}
	ax, err := fig.AddRadarAxes(rect, labels)
	if err != nil {
		panic(err)
	}
	ax.SetTitle("Radar")
	ax.SetThetaZeroLocation("N")
	ax.YScale = transform.NewLinear(0, 1)
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0.25, 0.5, 0.75, 1.0}}
	ax.YAxis.MinorLocator = nil
	ax.YAxis.Formatter = ticker.PercentFormatter{XMax: 1, Decimals: 0, DecimalsSet: true}
	styleGrid(ax.AddGrid(core.AxisBottom), render.Color{R: 0.78, G: 0.80, B: 0.84, A: 1}, 0.75)
	styleGrid(ax.AddGrid(core.AxisLeft), render.Color{R: 0.80, G: 0.83, B: 0.88, A: 1}, 0.75)

	angles := core.RadarAngles(len(labels))
	values := []float64{0.72, 0.88, 0.64, 0.79, 0.58}
	closedAngles := append(append([]float64(nil), angles...), angles[0])
	closedValues := append(append([]float64(nil), values...), values[0])
	fill := render.Color{R: 0.18, G: 0.50, B: 0.82, A: 0.22}
	width := 2.0
	ax.Fill(closedAngles, closedValues, core.FillOptions{Color: &fill})
	_, _ = ax.Plot(closedAngles, closedValues, core.PlotOptions{Color: &blue, LineWidth: &width})
}

func addSkewTPanel(fig *core.Figure, rect geom.Rect) {
	ax, err := fig.AddSkewXAxes(rect)
	if err != nil {
		panic(err)
	}
	ax.SetTitle("Skew-T")
	ax.SetXLabel("temp")
	ax.SetYLabel("pressure")
	if err := ax.SetYScale("log"); err != nil {
		panic(err)
	}
	ax.SetXLim(-70, 35)
	ax.SetYLim(1050, 180)
	ax.XAxis.Locator = ticker.MultipleLocator{Base: 20}
	ax.XAxis.MinorLocator = ticker.MultipleLocator{Base: 10}
	if ax.XAxisTop != nil {
		ax.XAxisTop.Locator = ax.XAxis.Locator
		ax.XAxisTop.MinorLocator = ax.XAxis.MinorLocator
	}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{200, 300, 500, 700, 850, 1000}}
	ax.YAxis.MinorFormatter = ticker.NullFormatter{}
	styleGrid(ax.AddGrid(core.AxisBottom), grid, 0.75)
	styleGrid(ax.AddGrid(core.AxisLeft), grid, 0.75)

	pressure := []float64{1000, 925, 850, 700, 600, 500, 400, 300, 250, 200}
	temperature := []float64{24, 20, 15, 5, -4, -14, -28, -43, -51, -58}
	dewpoint := []float64{18, 14, 8, -4, -14, -25, -38, -50, -57, -64}
	width := 2.0
	_, _ = ax.Plot(temperature, pressure, core.PlotOptions{Color: &red, LineWidth: &width, Label: "temp"})
	_, _ = ax.Plot(dewpoint, pressure, core.PlotOptions{Color: &green, LineWidth: &width, Label: "dew"})
	ax.AddLegend()
}

func addAxisArtistPanel(fig *core.Figure, rect geom.Rect) {
	ax := fig.AddAxes(rect)
	ax.SetTitle("AxisArtist / Twin")
	ax.SetXLabel("phase")
	ax.SetYLabel("signal")
	ax.SetXLim(-3.5, 3.5)
	ax.SetYLim(-1.3, 1.3)
	styleGrid(ax.AddYGrid(), grid, 0.75)

	const n = 180
	x := make([]float64, n)
	sine := make([]float64, n)
	cosScaled := make([]float64, n)
	for i := range n {
		x[i] = -3.5 + 7*float64(i)/float64(n-1)
		sine[i] = math.Sin(x[i])
		cosScaled[i] = 55 + 35*math.Cos(x[i]*0.8)
	}
	width := 2.0
	_, _ = ax.Plot(x, sine, core.PlotOptions{Color: &blue, LineWidth: &width, Label: "sin"})
	referenceWidth := 1.2
	reference := render.Color{R: 0.26, G: 0.26, B: 0.30, A: 1}
	ax.AxHLine(0, core.HLineOptions{Color: &reference, LineWidth: &referenceWidth, Dashes: []float64{5 * 36.0 / DPI, 3 * 36.0 / DPI}})
	ax.AxVLine(0, core.VLineOptions{Color: &reference, LineWidth: &referenceWidth, Dashes: []float64{5 * 36.0 / DPI, 3 * 36.0 / DPI}})

	overlay := ax.TwinX()
	if overlay != nil {
		overlay.SetYLim(0, 100)
		_, _ = overlay.Plot(x, cosScaled, core.PlotOptions{Color: &orange, LineWidth: &width, Label: "scaled cos"})
		if right := overlay.RightAxis(); right != nil {
			right.Color = orange
		}
	}
	ax.Text(0.03, 0.97, "floating reference\nparasite scale", core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		VAlign:   core.TextVAlignTop,
		FontSize: 8,
		BBox: &core.TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
			Padding:   0.25,
		},
	})
}

func addAxesGridPanel(fig *core.Figure, rect geom.Rect) {
	outer := fig.AddAxes(rect)
	outer.SetTitle("axes_grid1")
	outer.XAxis.Locator = ticker.FixedLocator{}
	outer.YAxis.Locator = ticker.FixedLocator{}
	outer.SetFrameOn(false)

	gridRect := geom.Rect{
		Min: geom.Pt{X: rect.Min.X + 0.01, Y: rect.Min.Y + 0.02},
		Max: geom.Pt{X: rect.Max.X - 0.01, Y: rect.Max.Y - 0.05},
	}
	imageGrid := fig.NewImageGrid(2, 2, gridRect, core.WithAxesDividerHorizontalSpace(0.012), core.WithAxesDividerVerticalSpace(0.018))
	if imageGrid == nil {
		return
	}
	for row := range 2 {
		for col := range 2 {
			ax := imageGrid.At(row, col)
			ax.SetTitle("Tile")
			ax.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 12, 23}}
			ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 12, 23}}
			ax.ImShow(surface(24, 24, float64(row*2+col)), core.ImShowOptions{})
		}
	}
}

func styleGrid(grid *core.Grid, color render.Color, width float64) {
	if grid == nil {
		return
	}
	grid.Color = color
	grid.LineWidth = width
}

func surface(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.5 + 0.25*math.Sin((xx+phase)*2*math.Pi) + 0.25*math.Cos((yy+phase*0.3)*3*math.Pi)
		}
	}
	return values
}
