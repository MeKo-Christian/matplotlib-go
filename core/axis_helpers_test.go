package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type staticLocator []float64

func (s staticLocator) Ticks(min, max float64, count int) []float64 {
	return append([]float64(nil), s...)
}

type offsetFormatterForTest struct {
	labels []string
	offset string
}

func (f offsetFormatterForTest) Format(x float64) string {
	return fmt.Sprintf("%.0f", x)
}

func (f offsetFormatterForTest) FormatTick(_ float64, index int, _ []float64) string {
	if index >= 0 && index < len(f.labels) {
		return f.labels[index]
	}
	return ""
}

func (f offsetFormatterForTest) OffsetText([]float64) string {
	return f.offset
}

type gridRecordingRenderer struct {
	render.NullRenderer
	paths []geom.Path
}

func (r *gridRecordingRenderer) Path(p geom.Path, _ *render.Paint) {
	r.paths = append(r.paths, p)
}

type axisLabelRecordingRenderer struct {
	render.NullRenderer
	texts          []string
	origins        []geom.Pt
	rotatedText    []string
	rotatedAnchors []geom.Pt
	textPathCalls  []string
	bounds         map[string]render.TextBounds
	useBounds      bool
	fontHeights    render.FontHeightMetrics
	useFontHeights bool
	pathCount      int
}

func (r *axisLabelRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *axisLabelRecordingRenderer) MeasureTextBounds(text string, _ float64, _ string) (render.TextBounds, bool) {
	if !r.useBounds || r.bounds == nil {
		return render.TextBounds{}, false
	}
	b, ok := r.bounds[text]
	return b, ok
}

func (r *axisLabelRecordingRenderer) MeasureFontHeights(_ float64, _ string) (render.FontHeightMetrics, bool) {
	if !r.useFontHeights {
		return render.FontHeightMetrics{}, false
	}
	return r.fontHeights, true
}

func (r *axisLabelRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *axisLabelRecordingRenderer) TextPath(text string, origin geom.Pt, _ float64, _ string) (geom.Path, bool) {
	r.textPathCalls = append(r.textPathCalls, text)
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: origin.X, Y: origin.Y - 4},
		Max: geom.Pt{X: origin.X + 4, Y: origin.Y},
	}), true
}

func (r *axisLabelRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.origins = append(r.origins, origin)
}

func (r *axisLabelRecordingRenderer) DrawTextRotated(text string, anchor geom.Pt, _ float64, _ float64, _ render.Color) {
	r.rotatedText = append(r.rotatedText, text)
	r.rotatedAnchors = append(r.rotatedAnchors, anchor)
}
