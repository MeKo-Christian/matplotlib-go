package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func contourPolygonsContainPoint(polygons [][]geom.Pt, want geom.Pt) bool {
	for _, polygon := range polygons {
		for _, got := range polygon {
			if sameContourPoint(got, want) {
				return true
			}
		}
	}
	return false
}

type contourTextRenderer struct {
	render.NullRenderer
	texts []string
}

func (r *contourTextRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func (r *contourTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type arraysShowcaseContourTextMetricRenderer struct {
	recordingRenderer
}

func (r *arraysShowcaseContourTextMetricRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	widths := map[string]float64{
		"0.15": 30.75,
		"0.3":  21.875,
		"0.30": 30.625,
		"0.45": 30.625,
		"0.6":  22,
		"0.60": 30.75,
		"0.75": 30.625,
		"0.9":  22,
		"0.90": 30.75,
	}
	width, ok := widths[text]
	if !ok {
		width = float64(len(text)) * size * 0.5
	}
	return render.TextMetrics{
		W:       width,
		H:       size * 1.4,
		Ascent:  size,
		Descent: size * 0.4,
	}
}
