package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Pie delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Pie(values []float64, opt core.PieOptions) *core.PieContainer {
	return GCA().Pie(values, opt)
}

// PieLabel delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PieLabel(container *core.PieContainer, labels []string, opt core.PieLabelOptions) []*core.Text {
	return GCA().PieLabel(container, labels, opt)
}

// Violinplot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Violinplot(data [][]float64, opt core.ViolinOptions) *core.ViolinContainer {
	return GCA().Violinplot(data, opt)
}

// Violin delegates precomputed violin statistics to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Violin(stats []core.ViolinStat, opt core.ViolinStatsOptions) *core.ViolinContainer {
	return GCA().Violin(stats, opt)
}

// Table delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Table(opt core.TableOptions) *core.Table {
	return GCA().Table(opt)
}

// Sankey returns a builder bound to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Sankey(opt core.SankeyOptions) *core.Sankey {
	return core.NewSankey(GCA(), opt)
}
