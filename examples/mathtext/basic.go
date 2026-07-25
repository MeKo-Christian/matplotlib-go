package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
)

func main() {
	fig := core.NewFigure(900, 540)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.10, Y: 0.12},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})

	// Damped sinusoid keeps the example simple while exercising math text in
	// titles, labels, free text, annotations, and anchored text.
	const n = 200
	x := make([]float64, n)
	y := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * 4 * math.Pi
		x[i] = t
		y[i] = math.Sin(t) * math.Exp(-0.08*t)
	}

	_, _ = ax.Plot(x, y)
	ax.SetTitle(`MathText $\\alpha^2 + \\beta_i$`)
	ax.SetXLabel(`phase $\\theta$`)
	ax.SetYLabel(`amplitude $\\frac{1}{\\sqrt{2}}$`)
	ax.Text(0.98, 0.92, `$x_{\\mathrm{max}}$`, core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignRight,
		VAlign:   core.TextVAlignTop,
		FontSize: 12,
	})
	ax.Annotate(`$\\Delta y \\approx \\frac{1}{2}$`, 3.2, 0.35, core.AnnotationOptions{
		OffsetX:  optional.Of(34.0),
		OffsetY:  optional.Of(-26.0),
		FontSize: 12,
	})
	ax.AddAnchoredText("$\\\\omega_n = 2\\\\pi f_n$", core.AnchoredTextOptions{
		Location: core.LegendUpperLeft,
		FontSize: 11,
	})

	if err := fig.Save("mathtext_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved mathtext_basic.png")
}
