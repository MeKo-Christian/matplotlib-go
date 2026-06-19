package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/render"
)

// StackPlotOptions configures Axes.StackPlot.
type StackPlotOptions struct {
	Colors    []render.Color
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Baseline  []float64
	Labels    []string
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

	var opt StackPlotOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xs := append([]float64(nil), x[:n]...)
	baseline := make([]float64, n)
	copy(baseline, opt.Baseline)

	fills := make([]*Fill2D, 0, len(ys))
	for i, y := range ys {
		lower := append([]float64(nil), baseline...)
		upper := make([]float64, n)
		for j := 0; j < n; j++ {
			upper[j] = lower[j] + y[j]
			baseline[j] = upper[j]
		}

		color := a.NextColor()
		if i < len(opt.Colors) {
			color = opt.Colors[i]
		}
		label := ""
		if i < len(opt.Labels) {
			label = opt.Labels[i]
		}
		fill := a.FillBetweenPlot(xs, upper, lower, FillOptions{
			Color:     &color,
			EdgeColor: opt.EdgeColor,
			EdgeWidth: opt.EdgeWidth,
			Alpha:     opt.Alpha,
			Label:     label,
		})
		if fill != nil {
			fills = append(fills, fill)
		}
	}
	return fills
}

// ECDF draws an empirical cumulative distribution function from raw samples.
func (a *Axes) ECDF(data []float64, opts ...ECDFOptions) *Line2D {
	samples := finiteSorted(data)
	if len(samples) == 0 {
		return nil
	}

	var opt ECDFOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
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

	var opt MultiHistOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

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

		hist := a.Hist(cleanSeries, histOpt)
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
