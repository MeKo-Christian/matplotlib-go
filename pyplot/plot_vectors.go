package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Quiver delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Quiver(x, y, u, v []float64, opt core.QuiverOptions) *core.Quiver {
	return GCA().Quiver(x, y, u, v, opt)
}

// QuiverGrid delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func QuiverGrid(x, y []float64, u, v [][]float64, opt core.QuiverOptions) *core.Quiver {
	return GCA().QuiverGrid(x, y, u, v, opt)
}

// QuiverKey delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func QuiverKey(q *core.Quiver, x, y, u float64, label string, opt core.QuiverKeyOptions) *core.QuiverKey {
	return GCA().QuiverKey(q, x, y, u, label, opt)
}

// Barbs delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Barbs(x, y, u, v []float64, opt core.BarbsOptions) *core.Barbs {
	return GCA().Barbs(x, y, u, v, opt)
}

// BarbsGrid delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func BarbsGrid(x, y []float64, u, v [][]float64, opt core.BarbsOptions) *core.Barbs {
	return GCA().BarbsGrid(x, y, u, v, opt)
}

// Streamplot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Streamplot(x, y []float64, u, v [][]float64, opt core.StreamplotOptions) *core.StreamplotSet {
	return GCA().Streamplot(x, y, u, v, opt)
}
