package main

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(1080, 720)

	// Static widgets are laid out with the same normalized rectangles as the Python example.
	plot := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.44},
		Max: geom.Pt{X: 0.94, Y: 0.92},
	})
	plot.SetTitle("widgets family")
	plot.SetXLabel("phase")
	plot.SetYLabel("response")
	plot.SetXLim(0, 2*math.Pi)
	plot.SetYLim(-1.3, 1.3)
	plot.AddYGrid()

	x := make([]float64, 240)
	y1 := make([]float64, len(x))
	y2 := make([]float64, len(x))
	for i := range x {
		x[i] = 2 * math.Pi * float64(i) / float64(len(x)-1)
		y1[i] = math.Sin(x[i])
		y2[i] = 0.6 * math.Cos(x[i]*1.5)
	}

	lineA := render.Color{R: 0.13, G: 0.36, B: 0.72, A: 1}
	lineB := render.Color{R: 0.84, G: 0.34, B: 0.18, A: 1}
	widthA := 2.2
	widthB := 1.8
	plot.Plot(x, y1, core.PlotOptions{Color: &lineA, LineWidth: &widthA, Label: "signal"})
	plot.Plot(x, y2, core.PlotOptions{Color: &lineB, LineWidth: &widthB, Label: "modulation"})
	plot.AddLegend()
	// AnchoredText stands in for Matplotlib text with a white bbox.
	plot.AddAnchoredText("static widget showcase\nMatplotlib-style control strip", core.AnchoredTextOptions{
		Location: core.LegendUpperLeft,
	})

	buttonAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.28},
		Max: geom.Pt{X: 0.22, Y: 0.38},
	})
	pressed := true
	// The widgets are rendered as static controls; they do not install event callbacks here.
	buttonAx.Button("Apply", core.ButtonOptions{Pressed: &pressed})

	sliderAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.26, Y: 0.28},
		Max: geom.Pt{X: 0.62, Y: 0.38},
	})
	sliderAx.Slider("gain", 0, 1, 0.68)

	checkAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.66, Y: 0.18},
		Max: geom.Pt{X: 0.80, Y: 0.38},
	})
	checkAx.CheckButtons([]string{"signal", "mod", "grid"}, []bool{true, true, false})

	radioAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.82, Y: 0.18},
		Max: geom.Pt{X: 0.94, Y: 0.38},
	})
	radioAx.RadioButtons([]string{"blue", "amber", "mono"}, 1)

	textAx := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.08, Y: 0.14},
		Max: geom.Pt{X: 0.62, Y: 0.24},
	})
	active := true
	textAx.TextBox("label", "phase scan", core.TextBoxOptions{Active: &active})

	fig.AddAnchoredText("widgets: Button, Slider, CheckButtons, RadioButtons, TextBox", core.AnchoredTextOptions{
		Location: core.LegendLowerRight,
	})

	r, _, err := backends.NewRendererFromEnv(backends.Config{
		Width:      1080,
		Height:     720,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if err != nil {
		fmt.Printf("error creating renderer: %v\n", err)
		return
	}
	if err := core.SavePNG(fig, r, "widgets_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved widgets_basic.png")
}
