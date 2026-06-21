package specialty_depth

import (
	"image"

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

	errAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.60}, Max: geom.Pt{X: 0.34, Y: 0.94}})
	errAx.SetTitle("ErrorBar limits")
	errAx.SetXLim(0, 5)
	errAx.SetYLim(0, 5)
	errAx.AddYGrid()
	errColor := render.Color{R: 0.12, G: 0.35, B: 0.70, A: 1}
	errWidth := 1.4
	errCap := 8.0
	errMarker := core.MarkerCircle
	errMarkerSize := 4.0
	errAx.ErrorBar(
		[]float64{1, 2, 3, 4},
		[]float64{1.2, 2.5, 3.1, 3.7},
		nil,
		nil,
		core.ErrorBarOptions{
			Color:      &errColor,
			LineWidth:  &errWidth,
			CapSize:    &errCap,
			Marker:     &errMarker,
			MarkerSize: &errMarkerSize,
			XErrLower:  []float64{0.25, 0.35, 0.20, 0.30},
			XErrUpper:  []float64{0.45, 0.25, 0.35, 0.20},
			YErrLower:  []float64{0.35, 0.50, 0.30, 0.60},
			YErrUpper:  []float64{0.55, 0.30, 0.65, 0.40},
			LoLimits:   []bool{false, true, false, false},
			UpLimits:   []bool{false, false, false, true},
			XLoLimits:  []bool{true, false, false, false},
			XUpLimits:  []bool{false, false, true, false},
			NoDataLine: true,
		},
	)

	boxAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.39, Y: 0.60}, Max: geom.Pt{X: 0.66, Y: 0.94}})
	boxAx.SetTitle("BoxPlot depth")
	boxAx.SetXLim(0.4, 2.6)
	boxAx.SetYLim(0, 8)
	boxAx.AddYGrid()
	notch := true
	whiskers := [2]float64{5, 95}
	ci1 := [2]float64{2.45, 3.10}
	ci2 := [2]float64{4.25, 5.00}
	median1 := 2.8
	median2 := 4.6
	flierSize := 4.0
	flierMarker := core.MarkerDiamond
	patchArtist := true
	boxes := boxAx.BoxPlots(
		[][]float64{
			{1.1, 1.8, 2.2, 2.6, 2.9, 3.1, 3.7, 6.8},
			{2.4, 3.1, 3.7, 4.3, 4.8, 5.2, 5.9, 7.2},
		},
		core.BoxPlotsOptions{
			PatchArtist:         &patchArtist,
			Notch:               &notch,
			WhiskerPercentiles:  &whiskers,
			ConfidenceIntervals: [][2]float64{ci1, ci2},
			CustomMedians:       []float64{median1, median2},
			FlierMarker:         &flierMarker,
			FlierSize:           &flierSize,
			Colors: []render.Color{
				{R: 0.45, G: 0.65, B: 0.90, A: 0.78},
				{R: 0.90, G: 0.55, B: 0.28, A: 0.78},
			},
		},
	)
	for _, box := range boxes {
		box.Label = ""
	}

	violinAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.73, Y: 0.60}, Max: geom.Pt{X: 0.97, Y: 0.94}})
	violinAx.SetTitle("Violin side")
	violinAx.SetXLim(0.5, 5.5)
	violinAx.SetYLim(0.5, 2.4)
	violinAx.AddXGrid()
	showMedians := true
	showExtrema := true
	violinEdge := render.Color{R: 0.12, G: 0.12, B: 0.12, A: 0.9}
	violinAx.Violinplot(
		[][]float64{
			{1.0, 1.3, 1.7, 2.2, 2.5, 3.1, 3.7, 4.1, 4.7},
			{1.4, 1.8, 2.2, 2.8, 3.2, 3.5, 4.2, 4.8, 5.1},
		},
		core.ViolinOptions{
			Orientation:     "horizontal",
			Side:            "high",
			Quantiles:       [][]float64{{0.25, 0.75}, {0.25, 0.75}},
			BandwidthMethod: "scott",
			EdgeColor:       &violinEdge,
			ShowMedians:     &showMedians,
			ShowExtrema:     &showExtrema,
			Colors: []render.Color{
				{R: 0.30, G: 0.60, B: 0.78, A: 0.58},
				{R: 0.30, G: 0.60, B: 0.78, A: 0.58},
			},
		},
	)

	pieAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.34, Y: 0.45}})
	pieAx.SetTitle("Pie labels")
	normalize := false
	wedgeEdge := render.Color{R: 1, G: 1, B: 1, A: 1}
	pie := pieAx.Pie([]float64{0.22, 0.18, 0.30}, core.PieOptions{
		Labels:       []string{"Alpha", "Beta", "Gamma"},
		Normalize:    &normalize,
		RotateLabels: true,
		Hatches:      []string{"/", "x", "\\"},
		Shadow:       true,
		StartAngle:   30,
		EdgeColor:    &wedgeEdge,
		LineWidth:    1.0,
		Colors: []render.Color{
			{R: 0.22, G: 0.55, B: 0.75, A: 1},
			{R: 0.90, G: 0.45, B: 0.18, A: 1},
			{R: 0.32, G: 0.64, B: 0.34, A: 1},
		},
	})
	if pie != nil {
		pieAx.PieLabel(pie, []string{"22%", "18%", "30%"}, core.PieLabelOptions{Distance: 0.62, Alignment: "center"})
	}

	hexAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.43, Y: 0.10}, Max: geom.Pt{X: 0.93, Y: 0.45}})
	hexAx.SetTitle("Hexbin log + marginals")
	hexAx.SetXLim(1, 120)
	hexAx.SetYLim(1, 120)
	hexAx.Hexbin(
		[]float64{1.2, 1.8, 2.6, 4.0, 6.5, 9.0, 14, 22, 35, 58, 92},
		[]float64{1.1, 2.2, 3.0, 5.5, 7.0, 12, 20, 28, 48, 80, 105},
		core.HexbinOptions{
			GridSizeX: 6,
			C:         []float64{1, 3, 2, 5, 7, 6, 11, 14, 18, 23, 30},
			Reduce:    "max",
			Bins:      "log",
			XScale:    "log",
			YScale:    "log",
			Marginals: true,
		},
	)
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
	return r.GetImage()
}
