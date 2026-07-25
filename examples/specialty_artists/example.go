package specialty_artists

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 980
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(980, 720)

	eventAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.57}, Max: geom.Pt{X: 0.34, Y: 0.94}})
	eventAx.SetTitle("Eventplot")
	eventAx.SetXLim(0, 10)
	eventAx.SetYLim(0.4, 3.6)
	eventGrid := eventAx.AddXGrid()
	eventGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	eventGrid.LineWidth = 0.5
	eventAx.Eventplot([][]float64{
		{0.8, 1.4, 3.1, 4.6, 7.3},
		{1.2, 2.9, 4.0, 6.4, 8.6},
		{0.5, 2.2, 5.4, 6.8, 9.1},
	}, core.EventPlotOptions{
		LineOffsets: []float64{1, 2, 3},
		LineLengths: []float64{0.6, 0.7, 0.8},
		Colors: []render.Color{
			{R: 0.18, G: 0.44, B: 0.74, A: 1},
			{R: 0.84, G: 0.38, B: 0.16, A: 1},
			{R: 0.20, G: 0.63, B: 0.42, A: 1},
		},
		LineWidth: 1.5,
	})

	hexAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.57}, Max: geom.Pt{X: 0.66, Y: 0.94}})
	hexAx.SetTitle("Hexbin")
	hexAx.SetXLim(0, 1)
	hexAx.SetYLim(0, 1)
	hexExtent := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}
	hexAx.Hexbin(
		[]float64{0.08, 0.15, 0.21, 0.25, 0.34, 0.41, 0.48, 0.56, 0.63, 0.66, 0.74, 0.82, 0.88},
		[]float64{0.14, 0.19, 0.24, 0.31, 0.46, 0.52, 0.61, 0.44, 0.73, 0.81, 0.68, 0.86, 0.58},
		core.HexbinOptions{
			GridSizeX: 7,
			Extent:    &hexExtent,
			C:         []float64{1, 2, 1.5, 2.3, 2.8, 3.1, 3.6, 2.1, 4.5, 4.9, 3.8, 5.2, 4.1},
			Reduce:    "mean",
			MinCount:  1,
		},
	)

	pieAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.73, Y: 0.57}, Max: geom.Pt{X: 0.98, Y: 0.94}})
	pieAx.SetTitle("Pie")
	wedgeEdge := render.Color{R: 1, G: 1, B: 1, A: 1}
	pie := pieAx.Pie([]float64{28, 22, 18, 32}, core.PieOptions{
		Labels:        []string{"Core", "I/O", "Render", "Docs"},
		AutoPct:       "%.0f%%",
		StartAngle:    90,
		LabelDistance: 1.08,
		Explode:       []float64{0, 0.04, 0, 0.02},
		EdgeColor:     &wedgeEdge,
		LineWidth:     1.0,
		Colors: []render.Color{
			{R: 0.12, G: 0.47, B: 0.71, A: 1},
			{R: 1.00, G: 0.50, B: 0.05, A: 1},
			{R: 0.17, G: 0.63, B: 0.17, A: 1},
			{R: 0.84, G: 0.15, B: 0.16, A: 1},
		},
	})
	if pie != nil {
		for _, txt := range pie.AutoText {
			txt.Color = render.Color{R: 0, G: 0, B: 0, A: 1}
		}
	}

	violinAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.08}, Max: geom.Pt{X: 0.34, Y: 0.45}})
	violinAx.SetTitle("Violin")
	violinAx.SetXLim(0.4, 3.6)
	violinAx.SetYLim(0.8, 5.2)
	violinGrid := violinAx.AddYGrid()
	violinGrid.Color = render.Color{R: 0.8, G: 0.8, B: 0.8, A: 1}
	violinGrid.LineWidth = 0.5
	violinBlue := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	violinEdge := render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1}
	violinShowMeans := true
	violinShowMedians := false
	violinShowExtrema := true
	violinPoints := 100
	violinWidths := []float64{0.5, 0.5, 0.5}
	violinAx.Violinplot([][]float64{
		{1.2, 1.5, 1.7, 2.1, 2.4, 2.6, 2.9, 3.0, 3.2},
		{1.8, 2.0, 2.2, 2.5, 2.7, 3.0, 3.4, 3.8, 4.0},
		{2.4, 2.5, 2.7, 2.9, 3.1, 3.4, 3.7, 4.1, 4.6},
	}, core.ViolinOptions{
		Colors:      []render.Color{violinBlue, violinBlue, violinBlue},
		EdgeColor:   &violinEdge,
		LineColor:   &violinBlue,
		Alpha:       0.45,
		Points:      violinPoints,
		Widths:      violinWidths,
		ShowMeans:   &violinShowMeans,
		ShowMedians: &violinShowMedians,
		ShowExtrema: &violinShowExtrema,
	})

	tableAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.08}, Max: geom.Pt{X: 0.66, Y: 0.45}})
	tableAx.SetTitle("Table")
	tableAx.ShowFrame = false
	if tableAx.XAxis != nil {
		tableAx.XAxis.ShowSpine = false
		tableAx.XAxis.ShowTicks = false
		tableAx.XAxis.ShowLabels = false
	}
	if tableAx.YAxis != nil {
		tableAx.YAxis.ShowSpine = false
		tableAx.YAxis.ShowTicks = false
		tableAx.YAxis.ShowLabels = false
	}
	tableAx.Table(core.TableOptions{
		ColLabels: []string{"Metric", "Q1", "Q2"},
		RowLabels: []string{"A", "B"},
		CellText: [][]string{
			{"Latency", "18ms", "14ms"},
			{"Throughput", "220/s", "265/s"},
		},
		BBox: geom.Rect{
			Min: geom.Pt{X: 0.04, Y: 0.18},
			Max: geom.Pt{X: 0.96, Y: 0.82},
		},
		FontSize: 10,
		CellLoc:  "center",
	})

	sankeyAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.73, Y: 0.08}, Max: geom.Pt{X: 0.98, Y: 0.45}})
	sankeyAx.SetTitle("Sankey")
	sankeyAx.ShowFrame = false
	if sankeyAx.XAxis != nil {
		sankeyAx.XAxis.ShowSpine = false
		sankeyAx.XAxis.ShowTicks = false
		sankeyAx.XAxis.ShowLabels = false
	}
	if sankeyAx.YAxis != nil {
		sankeyAx.YAxis.ShowSpine = false
		sankeyAx.YAxis.ShowTicks = false
		sankeyAx.YAxis.ShowLabels = false
	}
	sankeyAx.SetXLim(0, 1)
	sankeyAx.SetYLim(0, 1)
	sankeyAx.AddPatch(&core.Rectangle{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.12, G: 0.47, B: 0.71, A: 0.75},
			EdgeColor: render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
			EdgeWidth: 1,
		},
		XY:     geom.Pt{X: 0.18, Y: 0.47},
		Width:  0.18,
		Height: 0.06,
		Coords: core.Coords(core.CoordAxes),
	})
	sankeyFlows := []struct {
		label string
		flow  float64
		color render.Color
	}{
		{label: "Waste", flow: -2, color: render.Color{R: 0.84, G: 0.15, B: 0.16, A: 0.75}},
		{label: "CPU", flow: 3, color: render.Color{R: 0.17, G: 0.63, B: 0.17, A: 0.75}},
		{label: "Cache", flow: 1.5, color: render.Color{R: 1.00, G: 0.50, B: 0.05, A: 0.75}},
	}
	for idx, flow := range sankeyFlows {
		width := math.Abs(flow.flow) * 0.018
		y := 0.40 + float64(idx)*0.095
		x0 := 0.36
		x1 := 0.66
		path := geom.Path{}
		path.MoveTo(geom.Pt{X: x0, Y: 0.50 - width/2})
		path.LineTo(geom.Pt{X: x0 + 0.10, Y: y - width/2})
		path.LineTo(geom.Pt{X: x1, Y: y - width/2})
		path.LineTo(geom.Pt{X: x1, Y: y + width/2})
		path.LineTo(geom.Pt{X: x0 + 0.10, Y: y + width/2})
		path.LineTo(geom.Pt{X: x0, Y: 0.50 + width/2})
		path.Close()
		sankeyAx.AddPatch(&core.PathPatch{
			Patch: core.Patch{
				FaceColor: flow.color,
				EdgeColor: render.Color{R: 0.10, G: 0.10, B: 0.10, A: 1},
				EdgeWidth: 1,
			},
			Path:   path,
			Coords: core.Coords(core.CoordAxes),
		})
		sankeyAx.Text(0.70, y, flow.label, core.TextOptions{
			Coords:   core.Coords(core.CoordAxes),
			VAlign:   core.TextVAlignMiddle,
			FontSize: 10,
		})
	}
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}
