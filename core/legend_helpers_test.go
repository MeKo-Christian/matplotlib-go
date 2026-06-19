package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func legendBestPlacementTestContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale: transform.NewLinear(0, 1),
			YScale: transform.NewLinear(0, 1),
			// Display space is y-up: data y grows upward (no device flip here; the
			// backend owns that), so data (0.9,0.95) lands at the upper-right corner.
			AxesToPixel: transform.NewAffine(geom.Affine{A: 500, D: 500, F: 0}),
		},
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 500, Y: 500}},
	}
}

func legendDebugPathBounds(paths []geom.Path) []geom.Rect {
	out := make([]geom.Rect, 0, len(paths))
	for _, path := range paths {
		if bounds, ok := pathBounds(path); ok {
			out = append(out, bounds)
		}
	}
	return out
}

type legendRecordingRenderer struct {
	render.NullRenderer
	pathCount   int
	paths       []geom.Path
	paints      []render.Paint
	texts       []string
	textOrigins map[string]geom.Pt
}

type legendMarkerBatchRecordingRenderer struct {
	legendRecordingRenderer
	markerBatches []render.MarkerBatch
}

func (r *legendMarkerBatchRecordingRenderer) DrawMarkers(batch render.MarkerBatch) bool {
	r.markerBatches = append(r.markerBatches, batch)
	return true
}

type legendClipTrackingRenderer struct {
	legendRecordingRenderer
	clipStack   []bool
	clipActive  bool
	textClipped map[string]bool
}

func (r *legendClipTrackingRenderer) Save() {
	r.clipStack = append(r.clipStack, r.clipActive)
	r.legendRecordingRenderer.Save()
}

func (r *legendClipTrackingRenderer) Restore() {
	if len(r.clipStack) > 0 {
		r.clipActive = r.clipStack[len(r.clipStack)-1]
		r.clipStack = r.clipStack[:len(r.clipStack)-1]
	}
	r.legendRecordingRenderer.Restore()
}

func (r *legendClipTrackingRenderer) ClipRect(rect geom.Rect) {
	r.clipActive = true
	r.legendRecordingRenderer.ClipRect(rect)
}

func (r *legendClipTrackingRenderer) ClipPath(path geom.Path) {
	r.clipActive = true
	r.legendRecordingRenderer.ClipPath(path)
}

func (r *legendClipTrackingRenderer) DrawText(text string, origin geom.Pt, size float64, color render.Color) {
	r.legendRecordingRenderer.DrawText(text, origin, size, color)
	if r.textClipped == nil {
		r.textClipped = map[string]bool{}
	}
	r.textClipped[text] = r.clipActive
}

func (r *legendRecordingRenderer) Path(path geom.Path, paint *render.Paint) {
	r.pathCount++
	r.paths = append(r.paths, path)
	if paint == nil {
		r.paints = append(r.paints, render.Paint{})
		return
	}
	r.paints = append(r.paints, *paint)
}

func (r *legendRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *legendRecordingRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	if r.textOrigins == nil {
		r.textOrigins = map[string]geom.Pt{}
	}
	r.textOrigins[text] = origin
}

func (r *legendRecordingRenderer) textOrigin(text string) geom.Pt {
	if r.textOrigins == nil {
		return geom.Pt{}
	}
	return r.textOrigins[text]
}

func (r *legendRecordingRenderer) hasLegendFramePaint(legend *Legend) bool {
	for _, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			return true
		}
	}
	return false
}

func (r *legendRecordingRenderer) hasFillColor(color render.Color) bool {
	for _, paint := range r.paints {
		if paint.Fill == color {
			return true
		}
	}
	return false
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func pathBoundsForLegendTest(path geom.Path) geom.Rect {
	if len(path.V) == 0 {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: path.V[0], Max: path.V[0]}
	for _, pt := range path.V[1:] {
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	return bounds
}

func pathCenterX(path geom.Path) float64 {
	bounds := pathBoundsForLegendTest(path)
	return (bounds.Min.X + bounds.Max.X) / 2
}

func containsVerticalLegendPath(paths []geom.Path) bool {
	for _, path := range paths {
		if len(path.V) == 2 && floatApprox(path.V[0].X, path.V[1].X, 1e-9) && !floatApprox(path.V[0].Y, path.V[1].Y, 1e-9) {
			return true
		}
	}
	return false
}

func countHorizontalLegendSegments(paths []geom.Path) int {
	count := 0
	for _, path := range paths {
		if len(path.V) == 2 && floatApprox(path.V[0].Y, path.V[1].Y, 1e-9) && !floatApprox(path.V[0].X, path.V[1].X, 1e-9) {
			count++
		}
	}
	return count
}
