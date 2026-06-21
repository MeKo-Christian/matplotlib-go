package core

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ViolinOptions configures Axes.Violinplot.
type ViolinOptions struct {
	Positions       []float64
	Widths          []float64
	Colors          []render.Color
	EdgeColor       *render.Color
	EdgeWidth       float64
	Alpha           float64
	Bandwidth       float64
	BandwidthMethod string
	Points          int
	Orientation     string
	Side            string
	Quantiles       [][]float64
	LineColor       *render.Color
	ShowMeans       *bool
	ShowMedians     *bool
	ShowExtrema     *bool
	Label           string
}

// ViolinStat contains precomputed statistics for Axes.Violin.
type ViolinStat struct {
	Coords    []float64
	Vals      []float64
	Mean      float64
	Median    float64
	Min       float64
	Max       float64
	Quantiles []float64
}

// ViolinStatsOptions configures Axes.Violin.
type ViolinStatsOptions struct {
	Positions   []float64
	Widths      []float64
	Colors      []render.Color
	EdgeColor   *render.Color
	EdgeWidth   float64
	Alpha       float64
	Orientation string
	Side        string
	LineColor   *render.Color
	ShowMeans   *bool
	ShowMedians *bool
	ShowExtrema *bool
	Label       string
}

// ViolinContainer groups the collections created by Axes.Violinplot.
type ViolinContainer struct {
	Bodies    *PolyCollection
	Means     *LineCollection
	Medians   *LineCollection
	Quantiles *LineCollection
	Extrema   *LineCollection
}

type specialtyViolinSummary struct {
	min    float64
	max    float64
	mean   float64
	median float64
}

// Violinplot draws one violin body per dataset and optional summary lines.
func (a *Axes) Violinplot(data [][]float64, opts ...ViolinOptions) *ViolinContainer {
	if a == nil || len(data) == 0 {
		return nil
	}
	cfg := ViolinOptions{
		EdgeWidth: 1,
		Alpha:     0.3,
		Points:    100,
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.EdgeWidth <= 0 {
			cfg.EdgeWidth = 1
		}
		if cfg.Alpha <= 0 {
			cfg.Alpha = 0.3
		}
		if cfg.Points < 8 {
			cfg.Points = 100
		}
	}

	stats := make([]ViolinStat, 0, len(data))
	for i, series := range data {
		values := specialtyFiniteValues(series)
		if len(values) == 0 {
			continue
		}
		grid, density := specialtyKDE(values, cfg.Points, cfg.Bandwidth, cfg.BandwidthMethod)
		if len(grid) == 0 || len(density) == 0 {
			continue
		}
		summary := specialtyViolinStats(values)
		stats = append(stats, ViolinStat{
			Coords:    grid,
			Vals:      density,
			Mean:      summary.mean,
			Median:    summary.median,
			Min:       summary.min,
			Max:       summary.max,
			Quantiles: violinQuantiles(cfg.Quantiles, i, values),
		})
	}
	if len(stats) == 0 {
		return nil
	}
	return a.renderViolin(stats, ViolinStatsOptions{
		Positions:   cfg.Positions,
		Widths:      cfg.Widths,
		Colors:      cfg.Colors,
		EdgeColor:   cfg.EdgeColor,
		EdgeWidth:   cfg.EdgeWidth,
		Alpha:       cfg.Alpha,
		Orientation: cfg.Orientation,
		Side:        cfg.Side,
		LineColor:   cfg.LineColor,
		ShowMeans:   cfg.ShowMeans,
		ShowMedians: cfg.ShowMedians,
		ShowExtrema: cfg.ShowExtrema,
		Label:       cfg.Label,
	}, true)
}

// Violin draws one violin body per precomputed statistics entry.
func (a *Axes) Violin(stats []ViolinStat, opts ...ViolinStatsOptions) *ViolinContainer {
	if a == nil || len(stats) == 0 {
		return nil
	}
	var cfg ViolinStatsOptions
	if len(opts) > 0 {
		cfg = opts[0]
	}
	if cfg.EdgeWidth <= 0 {
		cfg.EdgeWidth = 1
	}
	if cfg.Alpha <= 0 {
		cfg.Alpha = 0.3
	}
	return a.renderViolin(stats, cfg, false)
}

func (a *Axes) renderViolin(stats []ViolinStat, cfg ViolinStatsOptions, defaultShowMedians bool) *ViolinContainer {
	n := len(stats)
	if !validOptionalList(cfg.Positions, n) || !validOptionalScalarList(cfg.Widths, n) {
		return nil
	}
	positions := expandFloatOption(cfg.Positions, n, func(i int) float64 { return float64(i + 1) })
	widths := expandFloatOption(cfg.Widths, n, func(int) float64 { return 0.5 })
	polygons := make([][]geom.Pt, 0, n)
	faceColors := make([]render.Color, 0, n)
	medianSegments := make([][]geom.Pt, 0, n)
	meanSegments := make([][]geom.Pt, 0, n)
	extremaSegments := make([][]geom.Pt, 0, n*3)
	quantileSegments := make([][]geom.Pt, 0, n)
	orientation := normalizeViolinOrientation(cfg.Orientation)
	side := normalizeViolinSide(cfg.Side)
	defaultFaceColor := a.NextColor()
	defaultLineColor := defaultFaceColor

	for i, stat := range stats {
		if !validViolinStat(stat) {
			return nil
		}
		width := math.Abs(widths[i])
		if width == 0 {
			width = 0.5
		}
		position := positions[i]
		maxDensity := 0.0
		for _, d := range stat.Vals {
			if d > maxDensity {
				maxDensity = d
			}
		}

		lowSide := make([]geom.Pt, 0, len(stat.Coords))
		highSide := make([]geom.Pt, 0, len(stat.Coords))
		for j := range stat.Coords {
			halfWidth := stat.Vals[j] / maxDensity * width * 0.5
			lowOffset := -halfWidth
			highOffset := halfWidth
			switch side {
			case "high":
				lowOffset = 0
			case "low":
				highOffset = 0
			}
			lowSide = append(lowSide, violinPoint(position+lowOffset, stat.Coords[j], orientation))
			highSide = append(highSide, violinPoint(position+highOffset, stat.Coords[j], orientation))
		}
		polygon := make([]geom.Pt, 0, len(lowSide)+len(highSide))
		polygon = append(polygon, lowSide...)
		for j := len(highSide) - 1; j >= 0; j-- {
			polygon = append(polygon, highSide[j])
		}
		polygons = append(polygons, polygon)

		color := colorAt(defaultFaceColor, cfg.Colors, i)
		if cfg.Alpha > 0 && cfg.Alpha <= 1 {
			color.A = 1
		}
		faceColors = append(faceColors, color)

		if specialtyBool(cfg.ShowMeans, false) {
			meanSegments = append(meanSegments, violinPerpSegment(position, width, stat.Mean, orientation, side))
		}
		if specialtyBool(cfg.ShowMedians, defaultShowMedians) {
			medianSegments = append(medianSegments, violinPerpSegment(position, width, stat.Median, orientation, side))
		}
		for _, q := range stat.Quantiles {
			quantileSegments = append(quantileSegments, violinPerpSegment(position, width, q, orientation, side))
		}
		if specialtyBool(cfg.ShowExtrema, true) {
			extremaSegments = append(
				extremaSegments,
				violinParallelSegment(position, stat.Min, stat.Max, orientation),
				violinPerpSegment(position, width, stat.Min, orientation, side),
				violinPerpSegment(position, width, stat.Max, orientation, side),
			)
		}
	}

	if len(polygons) == 0 {
		return nil
	}

	edgeColor := render.Color{}
	if cfg.EdgeColor != nil {
		edgeColor = *cfg.EdgeColor
		if cfg.Alpha > 0 && cfg.Alpha <= 1 {
			edgeColor.A = 1
		}
	}
	lineColor := defaultLineColor
	if cfg.LineColor != nil {
		lineColor = *cfg.LineColor
	}

	container := &ViolinContainer{
		Bodies: &PolyCollection{
			PatchCollection: PatchCollection{
				Collection: Collection{Label: cfg.Label, Alpha: clampOneToOne(cfg.Alpha), z: 2},
				FaceColors: faceColors,
				EdgeColor:  edgeColor,
				EdgeWidth:  cfg.EdgeWidth,
			},
			Polygons: polygons,
		},
	}
	a.AddCollection(container.Bodies)

	summaryCap := violinSummaryLineCap(side)
	summaryLineWidth := pointsToPixels(a.figure.RC, 1.5)
	if len(meanSegments) > 0 {
		container.Means = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.3},
			Segments:   meanSegments,
			Color:      lineColor,
			LineWidth:  summaryLineWidth,
			LineCap:    summaryCap,
		}
		a.AddCollection(container.Means)
	}
	if len(medianSegments) > 0 {
		container.Medians = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.35},
			Segments:   medianSegments,
			Color:      lineColor,
			LineWidth:  summaryLineWidth,
			LineCap:    summaryCap,
		}
		a.AddCollection(container.Medians)
	}
	if len(quantileSegments) > 0 {
		container.Quantiles = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.32},
			Segments:   quantileSegments,
			Color:      lineColor,
			LineWidth:  summaryLineWidth,
			LineCap:    summaryCap,
		}
		a.AddCollection(container.Quantiles)
	}
	if len(extremaSegments) > 0 {
		container.Extrema = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.1},
			Segments:   extremaSegments,
			Color:      lineColor,
			LineWidth:  summaryLineWidth,
			LineCap:    summaryCap,
		}
		a.AddCollection(container.Extrema)
	}
	return container
}

func validViolinStat(stat ViolinStat) bool {
	if len(stat.Coords) == 0 || len(stat.Coords) != len(stat.Vals) {
		return false
	}
	if !isFinite(stat.Mean) || !isFinite(stat.Median) || !isFinite(stat.Min) || !isFinite(stat.Max) {
		return false
	}
	hasPositiveDensity := false
	for i := range stat.Coords {
		if !isFinite(stat.Coords[i]) || !isFinite(stat.Vals[i]) || stat.Vals[i] < 0 {
			return false
		}
		hasPositiveDensity = hasPositiveDensity || stat.Vals[i] > 0
	}
	return hasPositiveDensity
}

func specialtyFiniteValues(values []float64) []float64 {
	out := make([]float64, 0, len(values))
	for _, value := range values {
		if isFinite(value) {
			out = append(out, value)
		}
	}
	slices.Sort(out)
	return out
}

func specialtyViolinStats(values []float64) specialtyViolinSummary {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return specialtyViolinSummary{
		min:    values[0],
		max:    values[len(values)-1],
		mean:   sum / float64(len(values)),
		median: percentileSorted(values, 50),
	}
}

func specialtyKDE(values []float64, points int, bandwidth float64, methods ...string) ([]float64, []float64) {
	if len(values) == 0 {
		return nil, nil
	}
	minValue := values[0]
	maxValue := values[len(values)-1]
	if minValue == maxValue {
		minValue -= 0.5
		maxValue += 0.5
	}
	if bandwidth <= 0 {
		mean := 0.0
		for _, value := range values {
			mean += value
		}
		mean /= float64(len(values))
		variance := 0.0
		for _, value := range values {
			diff := value - mean
			variance += diff * diff
		}
		if len(values) > 1 {
			variance /= float64(len(values) - 1)
		} else {
			variance = 0
		}
		std := math.Sqrt(variance)
		n := float64(len(values))
		factor := math.Pow(n, -0.2)
		if len(methods) > 0 {
			switch strings.ToLower(strings.TrimSpace(methods[0])) {
			case "", "scott":
				factor = math.Pow(n, -0.2)
			case "silverman":
				factor = math.Pow(n*0.75, -0.2)
			default:
				if parsed, err := strconv.ParseFloat(methods[0], 64); err == nil && parsed > 0 {
					factor = parsed
				}
			}
		}
		bandwidth = factor * std
		if bandwidth <= 0 || !isFinite(bandwidth) {
			bandwidth = (maxValue - minValue) / 12
		}
		if bandwidth <= 0 {
			bandwidth = 1
		}
	}

	grid := make([]float64, points)
	density := make([]float64, points)
	span := maxValue - minValue
	for i := range points {
		t := float64(i) / float64(max(points-1, 1))
		x := minValue + span*t
		grid[i] = x
		sum := 0.0
		for _, value := range values {
			z := (x - value) / bandwidth
			sum += math.Exp(-0.5 * z * z)
		}
		density[i] = sum / (float64(len(values)) * bandwidth * math.Sqrt(2*math.Pi))
	}
	return grid, density
}

func normalizeViolinOrientation(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "horizontal", "h", "x":
		return "horizontal"
	default:
		return "vertical"
	}
}

func normalizeViolinSide(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low", "left", "bottom":
		return "low"
	case "high", "right", "top":
		return "high"
	default:
		return "both"
	}
}

func violinPoint(posAxis, valueAxis float64, orientation string) geom.Pt {
	if orientation == "horizontal" {
		return geom.Pt{X: valueAxis, Y: posAxis}
	}
	return geom.Pt{X: posAxis, Y: valueAxis}
}

func violinPerpSegment(position, width, value float64, orientation, side string) []geom.Pt {
	low := position - width*0.25
	high := position + width*0.25
	switch side {
	case "low":
		high = position
	case "high":
		low = position
	}
	return []geom.Pt{violinPoint(low, value, orientation), violinPoint(high, value, orientation)}
}

func violinSummaryLineCap(side string) render.LineCap {
	if side == "low" || side == "high" {
		// Matplotlib switches one-sided violins to capstyle='projecting'
		// for the hlines/vlines summary artists in axes/_axes.py.
		return render.CapSquare
	}
	return render.CapButt
}

func violinParallelSegment(position, minValue, maxValue float64, orientation string) []geom.Pt {
	return []geom.Pt{violinPoint(position, minValue, orientation), violinPoint(position, maxValue, orientation)}
}

func violinQuantiles(quantiles [][]float64, i int, values []float64) []float64 {
	if i >= len(quantiles) {
		return nil
	}
	out := make([]float64, 0, len(quantiles[i]))
	for _, q := range quantiles[i] {
		if q < 0 || q > 1 || !isFinite(q) {
			continue
		}
		out = append(out, percentileSorted(values, q*100))
	}
	return out
}
