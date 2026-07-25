package core

import (
	"fmt"
	"math"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/render"
)

// ScalarMappable describes artists that map scalar values through a colormap
// and expose that mapping to helpers such as colorbars.
type ScalarMappable interface {
	ScalarMap() ScalarMapInfo
}

// ScalarMapInfo stores the colormap and value range used by a scalar-mappable
// artist.
type ScalarMapInfo struct {
	Colormap string
	VMin     float64
	VMax     float64
	Norm     ScalarNormalizer

	resolvedColormap     matcolor.Colormap
	resolvedColormapName string
}

func scalarMapConfigured(m ScalarMapInfo) bool {
	return m.Colormap != "" || m.Norm != nil
}

// ScalarMapConfig describes user-facing scalar-map inputs.
type ScalarMapConfig struct {
	Colormap string
	Norm     ScalarNormalizer
	VMin     *float64
	VMax     *float64
}

// Resolved returns a copy with sane defaults for downstream consumers.
func (m ScalarMapInfo) Resolved() ScalarMapInfo {
	if m.Colormap == "" {
		m.Colormap = "viridis"
	}
	if m.resolvedColormapName != m.Colormap || m.resolvedColormap.Name() == "" {
		m.resolvedColormap = matcolor.LookupColormap(m.Colormap)
		m.resolvedColormapName = m.Colormap
	}
	if m.Norm != nil {
		if vmin, vmax := m.Norm.Range(); isFinite(vmin) && isFinite(vmax) {
			m.VMin = vmin
			m.VMax = vmax
		}
		return m
	}
	if !isFinite(m.VMin) || !isFinite(m.VMax) {
		m.VMin = 0
		m.VMax = 1
	}
	if m.VMin == m.VMax {
		m.VMax = m.VMin + 1
	}
	m.Norm = Normalize{VMin: m.VMin, VMax: m.VMax}
	return m
}

// Normalize maps a scalar into the colormap domain.
func (m ScalarMapInfo) Normalize(v float64) float64 {
	m = m.Resolved()
	return clamp01(m.normalizeRaw(v))
}

// Color maps a scalar into a display color using the configured colormap.
func (m ScalarMapInfo) Color(v, alpha float64) render.Color {
	m = m.Resolved()
	color := m.resolvedColormap.AtValue(m.normalizeRaw(v))
	return color.WithAlphaMultiplier(clampOneToOne(alpha))
}

func (m ScalarMapInfo) normalizeRaw(v float64) float64 {
	m = m.Resolved()
	if m.Norm != nil {
		return m.Norm.Map(v)
	}
	span := m.VMax - m.VMin
	if span == 0 {
		return 0
	}
	return (v - m.VMin) / span
}

// ResolveScalarMapValues resolves colormap and norm configuration for scalar values.
func ResolveScalarMapValues(values []float64, cfg ScalarMapConfig) (ScalarMapInfo, error) {
	minValue, maxValue := finiteRange(values)
	return resolveScalarMapRange(minValue, maxValue, cfg)
}

// ResolveScalarMapGrid resolves colormap and norm configuration for a scalar grid.
func ResolveScalarMapGrid(data [][]float64, cfg ScalarMapConfig) (ScalarMapInfo, error) {
	minValue, maxValue := dataRange(data)
	return resolveScalarMapRange(minValue, maxValue, cfg)
}

func resolveScalarMapRange(minValue, maxValue float64, cfg ScalarMapConfig) (ScalarMapInfo, error) {
	if cfg.Norm != nil && (cfg.VMin != nil || cfg.VMax != nil) {
		return ScalarMapInfo{}, fmt.Errorf("cannot pass vmin/vmax with an explicit norm")
	}

	norm := cfg.Norm
	if norm == nil {
		vmin := math.NaN()
		vmax := math.NaN()
		if cfg.VMin != nil && isFinite(*cfg.VMin) {
			vmin = *cfg.VMin
		}
		if cfg.VMax != nil && isFinite(*cfg.VMax) {
			vmax = *cfg.VMax
		}
		norm = Normalize{VMin: vmin, VMax: vmax}
	}
	norm = norm.Autoscale([]float64{minValue, maxValue})
	if err := norm.Validate(); err != nil {
		return ScalarMapInfo{}, err
	}
	vmin, vmax := norm.Range()
	return ScalarMapInfo{
		Colormap: resolvedColormapName(cfg.Colormap),
		VMin:     vmin,
		VMax:     vmax,
		Norm:     norm,
	}.Resolved(), nil
}

func resolvedColormapName(name string) string {
	if name == "" {
		return "viridis"
	}
	return name
}

func finiteRange(values []float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	if math.IsInf(minValue, 1) || math.IsInf(maxValue, -1) {
		return 0, 1
	}
	if minValue == maxValue {
		return minValue, minValue + 1
	}
	return minValue, maxValue
}

func finiteMatrixSize(data [][]float64) (rows int, cols int, ok bool) {
	rows = len(data)
	if rows == 0 {
		return 0, 0, false
	}
	cols = len(data[0])
	if cols == 0 {
		return 0, 0, false
	}
	for _, row := range data[1:] {
		if len(row) != cols {
			return 0, 0, false
		}
	}
	return rows, cols, true
}
