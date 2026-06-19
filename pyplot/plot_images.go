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

// GetCMap returns a registered colormap by name.
func GetCMap(name string) matcolor.Colormap {
	return matcolor.GetColormap(name)
}

// Image delegates to the current axes.
func Image(data [][]float64, opts ...core.ImageOptions) *core.Image2D {
	return GCA().Image(data, opts...)
}

// ImShow delegates to the current axes.
func ImShow(data [][]float64, opts ...core.ImShowOptions) *core.Image2D {
	return GCA().ImShow(data, opts...)
}

// MatShow delegates to the current axes.
func MatShow(data [][]float64, opts ...core.MatShowOptions) *core.Image2D {
	return GCA().MatShow(data, opts...)
}

// Spy delegates to the current axes.
func Spy(data [][]float64, opts ...core.SpyOptions) *core.SpyResult {
	return GCA().Spy(data, opts...)
}

// PColor delegates to the current axes.
func PColor(data [][]float64, opts ...core.MeshOptions) *core.QuadMesh {
	return GCA().PColor(data, opts...)
}

// PColorFast delegates to the current axes.
func PColorFast(data [][]float64, opts ...core.MeshOptions) *core.QuadMesh {
	return GCA().PColorFast(data, opts...)
}

// PColorMesh delegates to the current axes.
func PColorMesh(data [][]float64, opts ...core.MeshOptions) *core.QuadMesh {
	return GCA().PColorMesh(data, opts...)
}

// Hist2D delegates to the current axes.
func Hist2D(x, y []float64, opts ...core.Hist2DOptions) *core.Hist2DResult {
	return GCA().Hist2D(x, y, opts...)
}

// Specgram delegates to the current axes.
func Specgram(samples []float64, opts ...core.SpecgramOptions) *core.SpecgramResult {
	return GCA().Specgram(samples, opts...)
}

// PSD delegates to the current axes.
func PSD(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PSD(samples, opts...)
}

// MagnitudeSpectrum delegates to the current axes.
func MagnitudeSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().MagnitudeSpectrum(samples, opts...)
}

// AngleSpectrum delegates to the current axes.
func AngleSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().AngleSpectrum(samples, opts...)
}

// PhaseSpectrum delegates to the current axes.
func PhaseSpectrum(samples []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().PhaseSpectrum(samples, opts...)
}

// CSD delegates to the current axes.
func CSD(x, y []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().CSD(x, y, opts...)
}

// Cohere delegates to the current axes.
func Cohere(x, y []float64, opts ...core.SignalSpectrumOptions) *core.SpectrumResult {
	return GCA().Cohere(x, y, opts...)
}

// XCorr delegates to the current axes.
func XCorr(x, y []float64, opts ...core.CorrelationOptions) *core.CorrelationResult {
	return GCA().XCorr(x, y, opts...)
}

// ACorr delegates to the current axes.
func ACorr(x []float64, opts ...core.CorrelationOptions) *core.CorrelationResult {
	return GCA().ACorr(x, opts...)
}

// AnnotatedHeatmap delegates to the current axes.
func AnnotatedHeatmap(data [][]float64, opts ...core.AnnotatedHeatmapOptions) *core.AnnotatedHeatmapResult {
	return GCA().AnnotatedHeatmap(data, opts...)
}

// Eventplot delegates to the current axes.
func Eventplot(positions [][]float64, opts ...core.EventPlotOptions) *core.EventCollection {
	return GCA().Eventplot(positions, opts...)
}

// Hexbin delegates to the current axes.
func Hexbin(x, y []float64, opts ...core.HexbinOptions) *core.HexbinCollection {
	return GCA().Hexbin(x, y, opts...)
}
