package core

import (
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
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
		Alpha:     0.65,
		Points:    100,
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.EdgeWidth <= 0 {
			cfg.EdgeWidth = 1
		}
		if cfg.Alpha <= 0 {
			cfg.Alpha = 0.65
		}
		if cfg.Points < 8 {
			cfg.Points = 100
		}
	}

	polygons := make([][]geom.Pt, 0, len(data))
	faceColors := make([]render.Color, 0, len(data))
	medianSegments := make([][]geom.Pt, 0, len(data))
	meanSegments := make([][]geom.Pt, 0, len(data))
	extremaSegments := make([][]geom.Pt, 0, len(data)*3)
	quantileSegments := make([][]geom.Pt, 0, len(data))
	orientation := normalizeViolinOrientation(cfg.Orientation)
	side := normalizeViolinSide(cfg.Side)

	for i, series := range data {
		values := specialtyFiniteValues(series)
		if len(values) == 0 {
			continue
		}

		width := math.Abs(floatAt(cfg.Widths, i, 0.5))
		if width == 0 {
			width = 0.5
		}
		position := floatAt(cfg.Positions, i, float64(i+1))
		stats := specialtyViolinStats(values)
		grid, density := specialtyKDE(values, cfg.Points, cfg.Bandwidth, cfg.BandwidthMethod)
		if len(grid) == 0 || len(density) == 0 {
			continue
		}

		maxDensity := 0.0
		for _, d := range density {
			if d > maxDensity {
				maxDensity = d
			}
		}
		if maxDensity == 0 {
			maxDensity = 1
		}

		lowSide := make([]geom.Pt, 0, len(grid))
		highSide := make([]geom.Pt, 0, len(grid))
		for j := range grid {
			halfWidth := density[j] / maxDensity * width * 0.5
			lowOffset := -halfWidth
			highOffset := halfWidth
			switch side {
			case "high":
				lowOffset = 0
			case "low":
				highOffset = 0
			}
			lowSide = append(lowSide, violinPoint(position+lowOffset, grid[j], orientation))
			highSide = append(highSide, violinPoint(position+highOffset, grid[j], orientation))
		}
		polygon := make([]geom.Pt, 0, len(lowSide)+len(highSide))
		polygon = append(polygon, lowSide...)
		for j := len(highSide) - 1; j >= 0; j-- {
			polygon = append(polygon, highSide[j])
		}
		polygons = append(polygons, polygon)

		color := colorAt(a.NextColor(), cfg.Colors, i)
		color.A *= clampOneToOne(cfg.Alpha)
		faceColors = append(faceColors, color)

		if specialtyBool(cfg.ShowMeans, false) {
			meanSegments = append(meanSegments, violinPerpSegment(position, width, stats.mean, orientation, side))
		}
		if specialtyBool(cfg.ShowMedians, true) {
			medianSegments = append(medianSegments, violinPerpSegment(position, width*1.2, stats.median, orientation, side))
		}
		for _, q := range violinQuantiles(cfg.Quantiles, i, values) {
			quantileSegments = append(quantileSegments, violinPerpSegment(position, width, q, orientation, side))
		}
		if specialtyBool(cfg.ShowExtrema, true) {
			extremaSegments = append(extremaSegments,
				violinParallelSegment(position, stats.min, stats.max, orientation),
				violinPerpSegment(position, width, stats.min, orientation, side),
				violinPerpSegment(position, width, stats.max, orientation, side),
			)
		}
	}

	if len(polygons) == 0 {
		return nil
	}

	edgeColor := render.Color{R: 0.12, G: 0.12, B: 0.12, A: 0.9}
	if cfg.EdgeColor != nil {
		edgeColor = *cfg.EdgeColor
	}
	lineColor := edgeColor
	if cfg.LineColor != nil {
		lineColor = *cfg.LineColor
	}

	container := &ViolinContainer{
		Bodies: &PolyCollection{
			PatchCollection: PatchCollection{
				Collection: Collection{Label: cfg.Label, Alpha: 1, z: 2},
				FaceColors: faceColors,
				EdgeColor:  edgeColor,
				EdgeWidth:  cfg.EdgeWidth,
			},
			Polygons: polygons,
		},
	}
	a.AddCollection(container.Bodies)

	if len(meanSegments) > 0 {
		container.Means = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.3},
			Segments:   meanSegments,
			Color:      lineColor,
			LineWidth:  math.Max(cfg.EdgeWidth, 1.25),
			LineCap:    render.CapRound,
		}
		a.AddCollection(container.Means)
	}
	if len(medianSegments) > 0 {
		container.Medians = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.35},
			Segments:   medianSegments,
			Color:      render.Color{R: 1, G: 1, B: 1, A: 0.95},
			LineWidth:  math.Max(cfg.EdgeWidth, 1.5),
			LineCap:    render.CapRound,
		}
		a.AddCollection(container.Medians)
	}
	if len(quantileSegments) > 0 {
		container.Quantiles = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.32},
			Segments:   quantileSegments,
			Color:      lineColor,
			LineWidth:  math.Max(cfg.EdgeWidth, 1.25),
			LineCap:    render.CapRound,
		}
		a.AddCollection(container.Quantiles)
	}
	if len(extremaSegments) > 0 {
		container.Extrema = &LineCollection{
			Collection: Collection{Alpha: 1, z: 2.1},
			Segments:   extremaSegments,
			Color:      lineColor,
			LineWidth:  math.Max(cfg.EdgeWidth*0.9, 1),
			LineCap:    render.CapRound,
		}
		a.AddCollection(container.Extrema)
	}
	return container
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
		std := math.Sqrt(variance / float64(len(values)))
		factor := 1.06
		if len(methods) > 0 {
			switch strings.ToLower(strings.TrimSpace(methods[0])) {
			case "scott":
				factor = 1
			case "silverman", "":
				factor = 1.06
			default:
				if parsed, err := strconv.ParseFloat(methods[0], 64); err == nil && parsed > 0 {
					factor = parsed
				}
			}
		}
		bandwidth = factor * std * math.Pow(float64(len(values)), -0.2)
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
