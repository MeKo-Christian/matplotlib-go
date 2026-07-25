package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// StackBaseline selects how Axes.StackPlot computes the bottom of the first
// layer, mirroring Matplotlib's stackplot “baseline“ parameter.
type StackBaseline int

const (
	// StackBaselineZero is a constant-zero baseline: a plain stacked area plot.
	StackBaselineZero StackBaseline = iota
	// StackBaselineSym is symmetric around zero (sometimes called "ThemeRiver").
	StackBaselineSym
	// StackBaselineWiggle minimizes the sum of the squared layer slopes.
	StackBaselineWiggle
	// StackBaselineWeightedWiggle is the streamgraph layout, weighting the wiggle
	// by the size of each layer (http://leebyron.com/streamgraph/).
	StackBaselineWeightedWiggle
)

// StackPlotOptions configures Axes.StackPlot.
type StackPlotOptions struct {
	Colors       []render.Color
	Alpha        *float64
	EdgeColor    *render.Color
	EdgeWidth    *float64
	BaselineMode StackBaseline // baseline computation; defaults to StackBaselineZero
	Baseline     []float64     // explicit per-point offset, used only in StackBaselineZero mode
	Labels       []string
}

// ECDFOptions configures Axes.ECDF.
type ECDFOptions struct {
	Color         *render.Color
	LineWidth     *float64
	Dashes        []float64
	Complementary bool
	Compress      bool
	Label         string
	Alpha         *float64
}

// MultiHistOptions configures Axes.HistMulti.
type MultiHistOptions struct {
	Bins       int
	BinEdges   []float64
	BinStrat   BinStrategy
	Norm       HistNorm
	Cumulative bool
	HistType   HistType
	Stacked    bool
	Colors     []render.Color
	EdgeColor  *render.Color
	EdgeWidth  *float64
	Alpha      *float64
	Labels     []string
}

// StackPlot draws cumulative filled layers over a shared x coordinate.
func (a *Axes) StackPlot(x []float64, ys [][]float64, opts ...StackPlotOptions) []*Fill2D {
	n := len(x)
	if n == 0 || len(ys) == 0 {
		return nil
	}
	for _, y := range ys {
		if len(y) < n {
			n = len(y)
		}
	}
	if n < 2 {
		return nil
	}

	opt := optarg.One("stackplot", opts)

	xs := append([]float64(nil), x[:n]...)
	m := len(ys)

	// cum[i][j] is the running cumulative sum over layers, i.e. numpy's
	// np.cumsum(y, axis=0). It is the unshifted top of layer i at point j.
	cum := make([][]float64, m)
	for i := range ys {
		cum[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			cum[i][j] = ys[i][j]
			if i > 0 {
				cum[i][j] += cum[i-1][j]
			}
		}
	}

	firstLine := stackFirstLine(ys, cum, n, m, opt)

	// level[i][j] is the shifted top of layer i (cum[i][j] + firstLine[j]).
	fills := make([]*Fill2D, 0, m)
	lower := append([]float64(nil), firstLine...)
	for i := 0; i < m; i++ {
		upper := make([]float64, n)
		for j := 0; j < n; j++ {
			upper[j] = cum[i][j] + firstLine[j]
		}

		// Pass an explicit color when provided; otherwise leave it nil so
		// FillBetweenPlot advances the property cycle exactly once per layer,
		// matching Matplotlib's stackplot (one cycle color per series).
		var colorPtr *render.Color
		if i < len(opt.Colors) {
			c := opt.Colors[i]
			colorPtr = &c
		}
		label := ""
		if i < len(opt.Labels) {
			label = opt.Labels[i]
		}
		// xs, lower, and upper are all built at length n >= 2 with no Where
		// mask, so the fill cannot be rejected here.
		fill, _ := a.FillBetweenPlot(xs, lower, upper, FillOptions{
			Color:     colorPtr,
			EdgeColor: opt.EdgeColor,
			EdgeWidth: opt.EdgeWidth,
			Alpha:     opt.Alpha,
			Label:     label,
		})
		if fill != nil {
			fills = append(fills, fill)
		}
		lower = upper
	}
	return fills
}

// stackFirstLine computes the bottom edge of the first stacked layer per the
// selected baseline mode, faithfully porting matplotlib's stackplot baselines.
func stackFirstLine(ys, cum [][]float64, n, m int, opt StackPlotOptions) []float64 {
	firstLine := make([]float64, n)

	switch opt.BaselineMode {
	case StackBaselineSym:
		// first_line = -sum(y, 0) * 0.5
		for j := 0; j < n; j++ {
			firstLine[j] = -0.5 * cum[m-1][j]
		}

	case StackBaselineWiggle:
		// first_line = (y * (m - 0.5 - arange(m))).sum(0) / -m
		for j := 0; j < n; j++ {
			var s float64
			for i := 0; i < m; i++ {
				s += ys[i][j] * (float64(m) - 0.5 - float64(i))
			}
			firstLine[j] = -s / float64(m)
		}

	case StackBaselineWeightedWiggle:
		// Streamgraph layout (http://leebyron.com/streamgraph/).
		var center float64
		for j := 0; j < n; j++ {
			total := cum[m-1][j]
			var invTotal float64
			if total > 0 {
				invTotal = 1.0 / total
			}
			var contrib float64
			for i := 0; i < m; i++ {
				// increase = hstack((y[:, 0:1], diff(y)))
				increase := ys[i][j]
				if j > 0 {
					increase = ys[i][j] - ys[i][j-1]
				}
				// below_size = total - stack + 0.5*y
				belowSize := total - cum[i][j] + 0.5*ys[i][j]
				// move_up = below_size * inv_total, forced to 0.5 at j == 0
				moveUp := belowSize * invTotal
				if j == 0 {
					moveUp = 0.5
				}
				contrib += (moveUp - 0.5) * increase
			}
			// center = cumsum(center.sum(0)); first_line = center - 0.5*total
			center += contrib
			firstLine[j] = center - 0.5*total
		}

	default: // StackBaselineZero
		copy(firstLine, opt.Baseline)
	}

	return firstLine
}

// ECDF draws an empirical cumulative distribution function from raw samples.
func (a *Axes) ECDF(data []float64, opts ...ECDFOptions) *Line2D {
	samples := finiteSorted(data)
	if len(samples) == 0 {
		return nil
	}

	opt := optarg.One("ecdf", opts)
	total := len(samples)
	values := samples
	probabilities := make([]float64, 0, total)
	if opt.Compress {
		values = make([]float64, 0, total)
		for i := 0; i < total; i++ {
			if i > 0 && samples[i-1] == samples[i] {
				continue
			}
			values = append(values, samples[i])
			probabilities = append(probabilities, float64(i+1)/float64(total))
		}
	} else {
		probabilities = make([]float64, total)
		for i := range probabilities {
			probabilities[i] = float64(i+1) / float64(total)
		}
	}

	x := make([]float64, 0, len(values)+1)
	y := make([]float64, 0, len(values)+1)
	x = append(x, values[0])
	if opt.Complementary {
		y = append(y, 1)
	} else {
		y = append(y, 0)
	}
	for i, v := range values {
		x = append(x, v)
		p := probabilities[i]
		if opt.Complementary {
			p = 1 - p
		}
		y = append(y, p)
	}

	where := StepWherePost
	return a.Step(x, y, StepOptions{
		Color:     opt.Color,
		LineWidth: opt.LineWidth,
		Dashes:    opt.Dashes,
		Where:     &where,
		Label:     opt.Label,
		Alpha:     opt.Alpha,
	})
}

// HistMulti draws multiple histograms using shared bin edges.
func (a *Axes) HistMulti(data [][]float64, opts ...MultiHistOptions) []*Hist2D {
	if len(data) == 0 {
		return nil
	}

	opt := optarg.One("hist multi", opts)

	edges := append([]float64(nil), opt.BinEdges...)
	if len(edges) < 2 {
		combined := flattenFinite(data)
		if len(combined) == 0 {
			return nil
		}
		edges = computeBinEdges(combined, opt.Bins, opt.BinStrat)
	}
	if len(edges) < 2 {
		return nil
	}

	baseline := make([]float64, len(edges)-1)
	hists := make([]*Hist2D, 0, len(data))
	for i, series := range data {
		cleanSeries := finiteValues(series)
		if len(cleanSeries) == 0 {
			continue
		}

		color := a.NextColor()
		if i < len(opt.Colors) {
			color = opt.Colors[i]
		}
		label := ""
		if i < len(opt.Labels) {
			label = opt.Labels[i]
		}

		histOpt := HistOptions{
			BinEdges:   edges,
			Norm:       opt.Norm,
			Cumulative: opt.Cumulative,
			HistType:   opt.HistType,
			Color:      &color,
			EdgeColor:  opt.EdgeColor,
			EdgeWidth:  opt.EdgeWidth,
			Alpha:      opt.Alpha,
			Label:      label,
		}
		if opt.Stacked {
			histOpt.Baselines = append([]float64(nil), baseline...)
		}

		// cleanSeries is non-empty and histOpt carries no weights, so the
		// histogram cannot be rejected here.
		hist, _ := a.Hist(cleanSeries, histOpt)
		if hist == nil {
			continue
		}
		hists = append(hists, hist)

		if opt.Stacked {
			_, counts := hist.BinCounts()
			for j := range baseline {
				if j < len(counts) {
					baseline[j] += counts[j]
				}
			}
		}
	}

	return hists
}

func finiteSorted(data []float64) []float64 {
	out := finiteValues(data)
	sort.Float64s(out)
	return out
}

func finiteValues(data []float64) []float64 {
	out := make([]float64, 0, len(data))
	for _, v := range data {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			out = append(out, v)
		}
	}
	return out
}

func flattenFinite(series [][]float64) []float64 {
	total := 0
	for _, s := range series {
		total += len(s)
	}
	out := make([]float64, 0, total)
	for _, s := range series {
		for _, v := range s {
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				out = append(out, v)
			}
		}
	}
	return out
}
