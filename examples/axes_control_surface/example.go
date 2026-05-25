package axes_control_surface

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

const (
	Width  = 760
	Height = 360
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(760, 360)

	left := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.07, Y: 0.14},
		Max: geom.Pt{X: 0.47, Y: 0.78},
	})
	left.SetTitle("Moved Axes + Aspect")
	left.SetXLabel("Top X")
	left.SetYLabel("Right Y")
	left.SetXLim(-1, 5)
	left.SetYLim(-1, 5)
	if err := left.SetXLabelPosition("top"); err != nil {
		panic(err)
	}
	if err := left.SetYLabelPosition("right"); err != nil {
		panic(err)
	}
	top := left.TopAxis()
	top.ShowLabels = true
	top.ShowTicks = true
	rightAxis := left.RightAxis()
	rightAxis.ShowLabels = true
	rightAxis.ShowTicks = true
	left.XAxis.ShowLabels = false
	left.XAxis.ShowTicks = false
	left.YAxis.ShowLabels = false
	left.YAxis.ShowTicks = false
	left.SetAxisEqual()
	if err := left.SetBoxAspect(1); err != nil {
		panic(err)
	}
	if err := left.MinorticksOn("both"); err != nil {
		panic(err)
	}
	if err := left.LocatorParams(core.LocatorParams{
		Axis:       "both",
		MajorCount: 6,
		MinorCount: 6,
	}); err != nil {
		panic(err)
	}
	left.XAxis.MinorLocator = core.MultipleLocator{Base: 0.2}
	left.XAxisTop.MinorLocator = core.MultipleLocator{Base: 0.2}
	left.YAxis.MinorLocator = core.MultipleLocator{Base: 0.2}
	left.YAxisRight.MinorLocator = core.MultipleLocator{Base: 0.2}
	majorLen := 7.0
	minorLen := 4.0
	majorWidth := 1.2
	minorWidth := 0.9
	tickColor := render.Color{R: 0.18, G: 0.42, B: 0.55, A: 1}
	if err := left.TickParams(core.TickParams{
		Axis:   "both",
		Which:  "major",
		Color:  &tickColor,
		Length: &majorLen,
		Width:  &majorWidth,
	}); err != nil {
		panic(err)
	}
	if err := left.TickParams(core.TickParams{
		Axis:   "both",
		Which:  "minor",
		Color:  &tickColor,
		Length: &minorLen,
		Width:  &minorWidth,
	}); err != nil {
		panic(err)
	}
	left.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: -0.5, Y: -0.2},
			{X: 0.8, Y: 1.0},
			{X: 2.2, Y: 2.1},
			{X: 4.2, Y: 4.4},
		},
		W:   2.0,
		Col: render.Color{R: 0.10, G: 0.32, B: 0.76, A: 1},
	})
	left.Add(&core.Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1.5, Y: 1.8},
			{X: 3.5, Y: 3.2},
			{X: 4.5, Y: 4.6},
		},
		Size:      core.ScatterAreaFromRadius(8.0, style.Default.DPI),
		Color:     render.Color{R: 0.92, G: 0.48, B: 0.20, A: 0.92},
		EdgeColor: render.Color{R: 0.52, G: 0.22, B: 0.08, A: 1},
		EdgeWidth: 1.0,
		Marker:    core.MarkerCircle,
		Alpha:     1,
	})

	right := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.58, Y: 0.14},
		Max: geom.Pt{X: 0.95, Y: 0.78},
	})
	right.SetTitle("Twin + Secondary")
	right.SetXLim(0, 10)
	right.SetYLim(0, 20)
	right.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 2},
			{X: 2, Y: 6},
			{X: 4, Y: 9},
			{X: 6, Y: 13},
			{X: 8, Y: 16},
			{X: 10, Y: 19},
		},
		W:   2.0,
		Col: render.Color{R: 0.12, G: 0.45, B: 0.72, A: 1},
	})

	twin := right.TwinX()
	twin.SetYLim(0, 100)
	twinLineColor := render.Color{R: 0.80, G: 0.22, B: 0.22, A: 1}
	if axis := twin.RightAxis(); axis != nil {
		axis.Color = twinLineColor
		axis.MinorLocator = nil
	}
	twin.Add(&core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 10},
			{X: 2, Y: 22},
			{X: 4, Y: 38},
			{X: 6, Y: 58},
			{X: 8, Y: 81},
			{X: 10, Y: 96},
		},
		W:   1.8,
		Col: twinLineColor,
	})

	sec, err := right.SecondaryXAxis(
		core.AxisTop,
		func(x float64) float64 { return x * 10 },
		func(x float64) (float64, bool) { return x / 10, true },
	)
	if err != nil {
		panic(err)
	}
	if axis := sec.TopAxis(); axis != nil {
		axis.Color = render.Color{R: 0.16, G: 0.42, B: 0.30, A: 1}
		axis.MinorLocator = nil
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
	return r.GetImage()
}
