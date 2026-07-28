// Package plot2d_gallery builds the broad 2D showcase used by the project
// README.
package plot2d_gallery

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1320
	Height = 840
	DPI    = 100
)

var (
	blue   = render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	orange = render.Color{R: 1.00, G: 0.50, B: 0.05, A: 1}
	green  = render.Color{R: 0.17, G: 0.63, B: 0.17, A: 1}
	red    = render.Color{R: 0.84, G: 0.15, B: 0.16, A: 1}
	black  = render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1}
)

// Plot builds a compact gallery of representative 2D plot families.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	panels := []struct {
		title string
		draw  func(*core.Axes)
	}{
		{title: "lines + markers", draw: drawLines},
		{title: "scatter", draw: drawScatter},
		{title: "bars", draw: drawBars},
		{title: "histogram", draw: drawHistogram},
		{title: "fill_between", draw: drawFillBetween},
		{title: "heatmap", draw: drawHeatmap},
		{title: "contours", draw: drawContours},
		{title: "box plots", draw: drawBoxPlots},
		{title: "vector field", draw: drawVectorField},
		{title: "pie", draw: drawPie},
	}

	for i, panel := range panels {
		row := i / 5
		col := i % 5
		left := 0.045 + float64(col)*0.193
		bottom := 0.54
		if row == 1 {
			bottom = 0.08
		}
		ax := fig.AddAxes(geom.Rect{
			Min: geom.Pt{X: left, Y: bottom},
			Max: geom.Pt{X: left + 0.16, Y: bottom + 0.36},
		})
		ax.SetTitle(panel.title)
		panel.draw(ax)
	}

	fig.Text(0.035, 0.975, "2D Gallery", core.TextOptions{
		Coords:   core.Coords(core.CoordFigure),
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		FontSize: 13,
	})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.Image()
}

func drawLines(ax *core.Axes) {
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.25, 1.25)
	addGrid(ax)

	const count = 33
	x := make([]float64, count)
	sine := make([]float64, count)
	cosine := make([]float64, count)
	for i := range count {
		x[i] = 2 * math.Pi * float64(i) / float64(count-1)
		sine[i] = math.Sin(x[i])
		cosine[i] = math.Cos(x[i])
	}
	_, _ = ax.Plot(x, sine, core.PlotOptions{
		Color:      optional.Of(blue),
		LineWidth:  optional.Of(1.7),
		Marker:     optional.Of(core.MarkerCircle),
		MarkerSize: optional.Of(4.0),
		MarkEvery:  4,
	})
	_, _ = ax.Plot(x, cosine, core.PlotOptions{
		Color:     optional.Of(orange),
		LineWidth: optional.Of(1.7),
		LineStyle: core.LineStyleDashed,
	})
}

func drawScatter(ax *core.Axes) {
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 9)
	addGrid(ax)
	_, _ = ax.Scatter(
		[]float64{1.0, 2.0, 3.2, 4.0, 5.2, 6.0, 7.1, 8.0, 9.0},
		[]float64{2.0, 6.5, 4.0, 7.2, 2.8, 5.2, 7.8, 3.5, 6.0},
		core.ScatterOptions{
			ScalarValues: []float64{0.1, 0.8, 0.4, 1.0, 0.2, 0.6, 0.9, 0.3, 0.7},
			Colormap:     "viridis",
			Sizes:        []float64{28, 45, 70, 95, 125, 155, 190, 225, 265},
			EdgeColor:    optional.Of(black),
			EdgeWidth:    optional.Of(0.8),
		},
	)
}

func drawBars(ax *core.Axes) {
	ax.SetXLim(0.4, 5.6)
	ax.SetYLim(0, 8)
	addGrid(ax)
	_, _ = ax.Bar(
		[]float64{1, 2, 3, 4, 5},
		[]float64{3.5, 6.4, 4.8, 7.1, 5.6},
		core.BarOptions{
			Colors:    []render.Color{blue, orange, green, red, {R: 0.58, G: 0.40, B: 0.74, A: 1}},
			Width:     optional.Of(0.72),
			EdgeColor: optional.Of(black),
			EdgeWidth: optional.Of(0.7),
		},
	)
}

func drawHistogram(ax *core.Axes) {
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 40)
	addGrid(ax)
	data := make([]float64, 120)
	for i := range data {
		data[i] = 5 + 1.45*math.Sin(float64(i)*1.73) + 0.8*math.Sin(float64(i)*0.31)
	}
	_, _ = ax.Hist(data, core.HistOptions{
		BinEdges:  []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Color:     optional.Of(render.Color{R: blue.R, G: blue.G, B: blue.B, A: 0.82}),
		EdgeColor: optional.Of(black),
		EdgeWidth: optional.Of(0.8),
	})
}

func drawFillBetween(ax *core.Axes) {
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.25, 1.25)
	addGrid(ax)
	const count = 60
	x := make([]float64, count)
	y1 := make([]float64, count)
	y2 := make([]float64, count)
	for i := range count {
		x[i] = 2 * math.Pi * float64(i) / float64(count-1)
		y1[i] = math.Sin(x[i])
		y2[i] = 0.35 * math.Cos(2*x[i])
	}
	_, _ = ax.FillBetween(x, y1, y2, core.FillOptions{
		Color:     optional.Of(render.Color{R: blue.R, G: blue.G, B: blue.B, A: 0.42}),
		EdgeColor: optional.Of(blue),
		EdgeWidth: optional.Of(1.0),
	})
	_, _ = ax.Plot(x, y1, core.PlotOptions{Color: optional.Of(blue), LineWidth: optional.Of(1.3)})
	_, _ = ax.Plot(x, y2, core.PlotOptions{Color: optional.Of(red), LineWidth: optional.Of(1.3)})
}

func drawHeatmap(ax *core.Axes) {
	data := make([][]float64, 8)
	for y := range data {
		data[y] = make([]float64, 8)
		for x := range data[y] {
			dx := float64(x) - 3.5
			dy := float64(y) - 3.5
			data[y][x] = math.Sin(math.Hypot(dx, dy)) + 0.15*float64(x)
		}
	}
	ax.SetXLim(0, 8)
	ax.SetYLim(0, 8)
	ax.Image(data, core.ImageOptions{
		Colormap: optional.Of("viridis"),
		XMin:     optional.Of(0.0),
		XMax:     optional.Of(8.0),
		YMin:     optional.Of(0.0),
		YMax:     optional.Of(8.0),
		Origin:   core.ImageOriginLower,
	})
}

func drawContours(ax *core.Axes) {
	const count = 21
	data := make([][]float64, count)
	for y := range count {
		data[y] = make([]float64, count)
		for x := range count {
			dx := (float64(x) - 10) / 3.2
			dy := (float64(y) - 10) / 4.0
			data[y][x] = math.Exp(-(dx*dx + dy*dy)) +
				0.55*math.Exp(-1.4*((dx-1.7)*(dx-1.7)+(dy+0.7)*(dy+0.7)))
		}
	}
	ax.SetXLim(0, count-1)
	ax.SetYLim(0, count-1)
	levels := []float64{0.05, 0.15, 0.3, 0.45, 0.6, 0.75, 0.9, 1.05}
	ax.Contourf(data, core.ContourOptions{Levels: levels, Colormap: optional.Of("viridis")})
	ax.Contour(data, core.ContourOptions{
		Levels:    []float64{0.15, 0.35, 0.55, 0.75, 0.95},
		Color:     optional.Of(black),
		LineWidth: optional.Of(0.7),
	})
}

func drawBoxPlots(ax *core.Axes) {
	ax.SetXLim(0.4, 3.6)
	ax.SetYLim(0, 10)
	addGrid(ax)
	patchArtist := true
	showFliers := true
	ax.BoxPlots(
		[][]float64{
			{1.0, 1.4, 1.8, 2.2, 2.5, 2.8, 3.1, 4.2},
			{2.4, 3.0, 3.4, 3.8, 4.2, 4.8, 5.3, 7.8},
			{3.0, 3.6, 4.1, 4.7, 5.1, 5.8, 6.4, 8.9},
		},
		core.BoxPlotsOptions{
			Positions:   []float64{1, 2, 3},
			Width:       optional.Of(0.55),
			Colors:      []render.Color{blue, orange, green},
			PatchArtist: optional.Of(patchArtist),
			EdgeColor:   optional.Of(black),
			MedianColor: optional.Of(black),
			FlierColor:  optional.Of(red),
			Alpha:       optional.Of(0.72),
			ShowFliers:  optional.Of(showFliers),
			ManageTicks: optional.Of(false),
		},
	)
}

func drawVectorField(ax *core.Axes) {
	ax.SetXLim(-2.5, 2.5)
	ax.SetYLim(-2.5, 2.5)
	addGrid(ax)
	var x, y, u, v []float64
	for row := -2; row <= 2; row++ {
		for col := -2; col <= 2; col++ {
			px := float64(col)
			py := float64(row)
			x = append(x, px)
			y = append(y, py)
			u = append(u, -py)
			v = append(v, px)
		}
	}
	ax.Quiver(x, y, u, v, core.QuiverOptions{
		Color:      optional.Of(blue),
		Pivot:      "middle",
		Angles:     "xy",
		Scale:      optional.Of(9.0),
		ScaleUnits: "width",
		Units:      "dots",
		Width:      optional.Of(2.0),
	})
}

func drawPie(ax *core.Axes) {
	ax.Pie([]float64{34, 27, 22, 17}, core.PieOptions{
		Labels:        []string{"Lines", "Images", "Stats", "Other"},
		AutoPct:       "%.0f%%",
		StartAngle:    90,
		LabelDistance: 1.08,
		Explode:       []float64{0.03, 0, 0, 0},
		EdgeColor:     optional.Of(render.Color{R: 1, G: 1, B: 1, A: 1}),
		LineWidth:     1.0,
		Colors:        []render.Color{blue, orange, green, red},
	})
}

func addGrid(ax *core.Axes) {
	for _, grid := range []*core.Grid{ax.AddXGrid(), ax.AddYGrid()} {
		grid.Color = render.Color{R: 0.82, G: 0.82, B: 0.82, A: 1}
		grid.LineWidth = 0.5
	}
	ax.SetAxisBelow(true)
}
