package pyplot

import (
	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

// ImRead decodes an image file into renderer-facing image data.
func ImRead(path string) (*render.ImageData, error) {
	return core.ImRead(path)
}

// ImSave writes image data to disk through the core image IO helper.
func ImSave(path string, img render.Image) error {
	return core.ImSave(path, img)
}

// CMap returns a registered colormap by name.
func CMap(name string) matcolor.Colormap {
	return matcolor.LookupColormap(name)
}

// Image delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Image(data [][]float64, opt core.ImageOptions) *core.Image2D {
	return GCA().Image(data, opt)
}

// ImShow delegates to the current axes.
//
//nolint:gocritic // ImShowOptions is forwarded unchanged to the axes method.
func ImShow(data [][]float64, opt core.ImShowOptions) *core.Image2D {
	return GCA().ImShow(data, opt)
}

// MatShow delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func MatShow(data [][]float64, opt core.MatShowOptions) *core.Image2D {
	return GCA().MatShow(data, opt)
}

// Spy delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Spy(data [][]float64, opt core.SpyOptions) *core.SpyResult {
	return GCA().Spy(data, opt)
}

// PColor delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PColor(data [][]float64, opt core.MeshOptions) *core.QuadMesh {
	return GCA().PColor(data, opt)
}

// PColorFast delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PColorFast(data [][]float64, opt core.MeshOptions) *core.QuadMesh {
	return GCA().PColorFast(data, opt)
}

// PColorMesh delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PColorMesh(data [][]float64, opt core.MeshOptions) *core.QuadMesh {
	return GCA().PColorMesh(data, opt)
}

// Hist2D delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Hist2D(x, y []float64, opt core.Hist2DOptions) *core.Hist2DResult {
	return GCA().Hist2D(x, y, opt)
}

// Specgram delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Specgram(samples []float64, opt core.SpecgramOptions) *core.SpecgramResult {
	return GCA().Specgram(samples, opt)
}

// PSD delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PSD(samples []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PSD(samples, opt)
}

// MagnitudeSpectrum delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func MagnitudeSpectrum(samples []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().MagnitudeSpectrum(samples, opt)
}

// AngleSpectrum delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AngleSpectrum(samples []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().AngleSpectrum(samples, opt)
}

// PhaseSpectrum delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func PhaseSpectrum(samples []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PhaseSpectrum(samples, opt)
}

// CSD delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func CSD(x, y []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().CSD(x, y, opt)
}

// Cohere delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Cohere(x, y []float64, opt core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().Cohere(x, y, opt)
}

// XCorr delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func XCorr(x, y []float64, opt core.CorrelationOptions) *core.CorrelationResult {
	return GCA().XCorr(x, y, opt)
}

// ACorr delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func ACorr(x []float64, opt core.CorrelationOptions) *core.CorrelationResult {
	return GCA().ACorr(x, opt)
}

// AnnotatedHeatmap delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func AnnotatedHeatmap(data [][]float64, opt core.AnnotatedHeatmapOptions) *core.AnnotatedHeatmapResult {
	return GCA().AnnotatedHeatmap(data, opt)
}

// Eventplot delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Eventplot(positions [][]float64, opt core.EventPlotOptions) *core.EventCollection {
	return GCA().Eventplot(positions, opt)
}

// Hexbin delegates to the current axes.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func Hexbin(x, y []float64, opt core.HexbinOptions) *core.HexbinCollection {
	return GCA().Hexbin(x, y, opt)
}
