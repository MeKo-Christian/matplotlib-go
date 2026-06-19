package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Quiver delegates to the current axes.
func Quiver(x, y, u, v []float64, opts ...core.QuiverOptions) *core.Quiver {
	return GCA().Quiver(x, y, u, v, opts...)
}

// QuiverGrid delegates to the current axes.
func QuiverGrid(x, y []float64, u, v [][]float64, opts ...core.QuiverOptions) *core.Quiver {
	return GCA().QuiverGrid(x, y, u, v, opts...)
}

// QuiverKey delegates to the current axes.
func QuiverKey(q *core.Quiver, x, y, u float64, label string, opts ...core.QuiverKeyOptions) *core.QuiverKey {
	return GCA().QuiverKey(q, x, y, u, label, opts...)
}

// Barbs delegates to the current axes.
func Barbs(x, y, u, v []float64, opts ...core.BarbsOptions) *core.Barbs {
	return GCA().Barbs(x, y, u, v, opts...)
}

// BarbsGrid delegates to the current axes.
func BarbsGrid(x, y []float64, u, v [][]float64, opts ...core.BarbsOptions) *core.Barbs {
	return GCA().BarbsGrid(x, y, u, v, opts...)
}

// Streamplot delegates to the current axes.
func Streamplot(x, y []float64, u, v [][]float64, opts ...core.StreamplotOptions) *core.StreamplotSet {
	return GCA().Streamplot(x, y, u, v, opts...)
}
