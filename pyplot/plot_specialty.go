package pyplot

import "github.com/cwbudde/matplotlib-go/core"

// Pie delegates to the current axes.
func Pie(values []float64, opts ...core.PieOptions) *core.PieContainer {
	return GCA().Pie(values, opts...)
}

// PieLabel delegates to the current axes.
func PieLabel(container *core.PieContainer, labels []string, opts ...core.PieLabelOptions) []*core.Text {
	return GCA().PieLabel(container, labels, opts...)
}

// Violinplot delegates to the current axes.
func Violinplot(data [][]float64, opts ...core.ViolinOptions) *core.ViolinContainer {
	return GCA().Violinplot(data, opts...)
}

// Violin delegates precomputed violin statistics to the current axes.
func Violin(stats []core.ViolinStat, opts ...core.ViolinStatsOptions) *core.ViolinContainer {
	return GCA().Violin(stats, opts...)
}

// Table delegates to the current axes.
func Table(opts ...core.TableOptions) *core.Table {
	return GCA().Table(opts...)
}

// Sankey returns a builder bound to the current axes.
func Sankey(opts ...core.SankeyOptions) *core.Sankey {
	return core.NewSankey(GCA(), opts...)
}
