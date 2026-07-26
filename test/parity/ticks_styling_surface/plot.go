package ticks_styling_surface

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

// Plot builds a tick styling and side-visibility parity fixture.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.12, Y: 0.16}, Max: geom.Pt{X: 0.90, Y: 0.78}})
	ax.SetTitle("Tick Styling Surface")
	ax.SetXLabel("top labels")
	ax.SetYLabel("right labels")
	if err := ax.SetXLabelPosition("top"); err != nil {
		panic(err)
	}
	if err := ax.SetYLabelPosition("right"); err != nil {
		panic(err)
	}
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 12)

	top := ax.TopAxis()
	right := ax.RightAxis()
	majorX := []float64{0, 2, 4, 6}
	majorY := []float64{0, 4, 8, 12}
	ax.XAxis.Locator = ticker.FixedLocator{TicksList: majorX}
	top.Locator = ticker.FixedLocator{TicksList: majorX}
	ax.YAxis.Locator = ticker.FixedLocator{TicksList: majorY}
	right.Locator = ticker.FixedLocator{TicksList: majorY}
	ax.XAxis.MinorLocator = ticker.MultipleLocator{Base: 1}
	top.MinorLocator = ticker.MultipleLocator{Base: 1}
	ax.YAxis.MinorLocator = ticker.MultipleLocator{Base: 2}
	right.MinorLocator = ticker.MultipleLocator{Base: 2}

	xGrid := ax.AddXGrid()
	yGrid := ax.AddYGrid()
	xGrid.Minor = true
	yGrid.Minor = true

	show := true
	hide := false
	majorColor := render.Color{R: 0.18, G: 0.42, B: 0.55, A: 1}
	minorColor := render.Color{R: 0.45, G: 0.45, B: 0.45, A: 1}
	majorLen := 8.0
	minorLen := 4.0
	majorWidth := 1.4
	minorWidth := 0.8
	rotation := 35.0
	gridColor := render.Color{R: 0.50, G: 0.60, B: 0.70, A: 1}
	minorGridColor := render.Color{R: 0.70, G: 0.74, B: 0.78, A: 1}
	gridAlpha := 0.65
	minorGridAlpha := 0.45
	gridWidth := 0.8
	minorGridWidth := 0.5
	if err := ax.TickParams(core.TickParams{
		Axis:          "both",
		Which:         "major",
		Color:         &majorColor,
		Length:        &majorLen,
		Width:         &majorWidth,
		LabelRotation: &rotation,
	}); err != nil {
		panic(err)
	}
	if err := ax.TickParams(core.TickParams{
		Axis:   "both",
		Which:  "minor",
		Color:  &minorColor,
		Length: &minorLen,
		Width:  &minorWidth,
	}); err != nil {
		panic(err)
	}
	if err := ax.TickParams(core.TickParams{
		Axis:        "x",
		Top:         &show,
		Bottom:      &hide,
		LabelTop:    &show,
		LabelBottom: &hide,
	}); err != nil {
		panic(err)
	}
	if err := ax.TickParams(core.TickParams{
		Axis:       "y",
		Right:      &show,
		Left:       &hide,
		LabelRight: &show,
		LabelLeft:  &hide,
	}); err != nil {
		panic(err)
	}
	if err := ax.TickParams(core.TickParams{
		Axis:          "both",
		Which:         "major",
		GridVisible:   &show,
		GridColor:     &gridColor,
		GridAlpha:     &gridAlpha,
		GridLineWidth: &gridWidth,
		GridDashes:    []float64{4, 2},
	}); err != nil {
		panic(err)
	}
	if err := ax.TickParams(core.TickParams{
		Axis:          "both",
		Which:         "minor",
		GridVisible:   &show,
		GridColor:     &minorGridColor,
		GridAlpha:     &minorGridAlpha,
		GridLineWidth: &minorGridWidth,
		GridDashes:    []float64{1, 2},
	}); err != nil {
		panic(err)
	}

	color := render.Color{R: 0.12, G: 0.47, B: 0.71, A: 1}
	width := 2.0
	_, _ = ax.Plot(
		[]float64{0, 1.5, 3, 4.5, 6},
		[]float64{1, 4, 5.5, 9, 11},
		core.PlotOptions{Color: optional.Of(color), LineWidth: optional.Of(width)},
	)
	return fig
}

// Render is the AGG-rendered fixture image.
func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
