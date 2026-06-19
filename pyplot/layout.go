package pyplot

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

// Axes appends an axes rectangle to the current figure and marks it current.
func Axes(r geom.Rect, opts ...style.Option) *core.Axes {
	fig := GCF()
	ax := fig.AddAxes(r, opts...)
	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, "")
	registry.mu.Unlock()
	return ax
}

// AddAxes3D appends an Axes3D to the current figure and marks it current.
func AddAxes3D(r geom.Rect, opts ...style.Option) *core.Axes3D {
	fig := GCF()
	ax, err := fig.AddAxes3D(r, opts...)
	if err != nil {
		return nil
	}
	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = ax.Axes
	registry.mu.Unlock()
	return ax
}

// AddAxesDivider creates an internal layout helper for structured axes tiling.
func AddAxesDivider(r geom.Rect, rows, cols int, opts ...core.AxesDividerOption) *core.AxesDivider {
	return GCF().NewAxesDivider(r, rows, cols, opts...)
}

// NewImageGrid creates an image-grid composed via an axes divider.
func NewImageGrid(rows, cols int, r geom.Rect, opts ...core.AxesDividerOption) *core.ImageGrid {
	fig := GCF()
	grid := fig.NewImageGrid(rows, cols, r, opts...)
	if grid == nil || len(grid.Axes) == 0 || len(grid.Axes[0]) == 0 {
		return grid
	}

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = grid.Axes[0][0]
	registry.mu.Unlock()
	return grid
}

// NewRGBAxes creates three synchronized axes for channel-wise RGB workflows.
func NewRGBAxes(r geom.Rect, opts ...core.AxesDividerOption) *core.RGBAxes {
	fig := GCF()
	axes := fig.NewRGBAxes(r, opts...)
	if axes == nil {
		return nil
	}

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = axes.Red
	registry.mu.Unlock()
	return axes
}

// ParasiteAxes creates a multi-view overlay axes over the current axes viewport.
func ParasiteAxes(opts ...core.ParasiteAxesOption) *core.ParasiteAxes {
	ax := GCA()
	parasite := ax.ParasiteAxes(opts...)
	if parasite == nil {
		return nil
	}
	fig := GCF()

	registry.mu.Lock()
	registry.current = fig
	registry.currentAxes[fig] = parasite.Axes
	registry.mu.Unlock()
	return parasite
}

// FloatingXAxis creates an auxiliary x-axis at the requested y data coordinate.
func FloatingXAxis(y float64) *core.AxisArtist {
	return GCA().FloatingXAxis(y)
}

// FloatingYAxis creates an auxiliary y-axis at the requested x data coordinate.
func FloatingYAxis(x float64) *core.AxisArtist {
	return GCA().FloatingYAxis(x)
}

// GCA3D returns the current 3D axes wrapper when the current axes uses a 3D
// projection, or nil otherwise.
func GCA3D() *core.Axes3D {
	ax := GCA()
	if ax == nil {
		return nil
	}
	if name := ax.ProjectionName(); name != "3d" && name != "axes3d" {
		return nil
	}
	return core.NewAxes3D(ax)
}

// Subplot returns the requested subplot axes in the current figure.
func Subplot(nRows, nCols, index int) *core.Axes {
	fig := GCF()
	key := subplotKey(nRows, nCols, index)

	registry.mu.Lock()
	if ax := registry.subplotAxes[fig][key]; ax != nil {
		registry.current = fig
		registry.currentAxes[fig] = ax
		registry.mu.Unlock()
		return ax
	}
	registry.mu.Unlock()

	ax := fig.AddSubplot(nRows, nCols, index)
	if ax == nil {
		return nil
	}

	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, key)
	registry.mu.Unlock()
	return ax
}

// Subplots creates a new figure and subplot grid, then marks the first axes as
// current.
func Subplots(nRows, nCols int, opts ...core.SubplotOption) (*core.Figure, [][]*core.Axes) {
	fig := Figure()
	grid := fig.Subplots(nRows, nCols, opts...)
	if len(grid) == 0 || len(grid[0]) == 0 {
		return fig, grid
	}

	registry.mu.Lock()
	for row := range grid {
		for col, ax := range grid[row] {
			registry.rememberAxesLocked(fig, ax, subplotKey(nRows, nCols, row*nCols+col+1))
		}
	}
	registry.current = fig
	registry.currentAxes[fig] = grid[0][0]
	registry.mu.Unlock()

	return fig, grid
}

// SubplotsAdjust applies persistent subplot layout adjustments to the current figure.
func SubplotsAdjust(cfg core.SubplotAdjust) {
	GCF().SubplotsAdjust(cfg)
}

// Subplot2Grid creates a spanning subplot inside a logical grid on the current figure.
func Subplot2Grid(shape, loc [2]int, rowSpan, colSpan int, opts ...core.SubplotAxesOption) *core.Axes {
	fig := GCF()
	ax := fig.Subplot2Grid(shape, loc, rowSpan, colSpan, opts...)
	registry.mu.Lock()
	registry.rememberAxesLocked(fig, ax, "")
	registry.mu.Unlock()
	return ax
}

// SubplotMosaic creates named subplot areas on the current figure.
func SubplotMosaic(layout [][]string, opts ...core.GridSpecOption) (map[string]*core.Axes, error) {
	fig := GCF()
	axes, err := fig.SubplotMosaic(layout, opts...)
	if err != nil {
		return nil, err
	}
	registry.mu.Lock()
	var firstAxes *core.Axes
	for row := range layout {
		for _, label := range layout[row] {
			if label == "" || label == "." {
				continue
			}
			ax := axes[label]
			if ax == nil {
				continue
			}
			registry.rememberAxesLocked(fig, ax, "")
			if firstAxes == nil {
				firstAxes = ax
			}
		}
	}
	if firstAxes != nil {
		registry.current = fig
		registry.currentAxes[fig] = firstAxes
	}
	registry.mu.Unlock()
	return axes, nil
}

func subplotKey(nRows, nCols, index int) string {
	return fmt.Sprintf("%d:%d:%d", nRows, nCols, index)
}
