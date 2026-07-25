package main

import (
	"fmt"
	"math"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
)

func main() {
	fig := core.NewFigure(1200, 800)

	// Build the same 2x2 shared-axis layout as pyplot.subplots(...,
	// sharex=True, sharey=True) in the Python counterpart.
	grid := fig.Subplots(
		2,
		2,
		core.WithSubplotPadding(0.08, 0.96, 0.08, 0.92),
		core.WithSubplotSpacing(0.08, 0.08),
		core.WithSubplotShareX(),
		core.WithSubplotShareY(),
	)

	for row, rowAxes := range grid {
		for col, ax := range rowAxes {
			ax.SetTitle(fmt.Sprintf("Panel %d-%d", row+1, col+1))
			ax.AddXGrid()
			ax.AddYGrid()

			// Each panel shares x samples but varies frequency by column and
			// damping by row, matching the Python loop structure.
			const n = 128
			x := make([]float64, n)
			y := make([]float64, n)
			for i := 0; i < n; i++ {
				v := float64(i) / float64(n-1)
				xv := 10 * v
				x[i] = xv
				y[i] = math.Sin(float64(col+1)*xv) * math.Exp(-0.05*v*float64(row+1))
			}

			_, _ = ax.Plot(x, y)
			ax.SetXLabel("x")
			ax.SetYLabel("y")
		}
	}

	// Shared axes are controlled by a single axis; set limits on one axes.
	grid[0][0].SetXLim(0, 10)
	grid[0][0].SetYLim(-1.2, 1.2)

	if err := fig.Save("subplots_basic.png"); err != nil {
		fmt.Printf("error saving PNG: %v\n", err)
		return
	}

	fmt.Println("saved subplots_basic.png")
}
