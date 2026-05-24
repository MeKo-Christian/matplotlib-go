package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestLegendCollectEntries(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "points"})
	ax.Bar([]float64{1}, []float64{2}, BarOptions{Label: "bars"})
	ax.Plot([]float64{0, 1}, []float64{1, 0})

	legend := ax.AddLegend()
	entries := legend.collectEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 legend entries, got %d", len(entries))
	}

	if entries[0].Label != "line" || entries[0].kind != legendEntryLine {
		t.Fatalf("unexpected first legend entry: %+v", entries[0])
	}
	if entries[1].Label != "points" || entries[1].kind != legendEntryMarker {
		t.Fatalf("unexpected second legend entry: %+v", entries[1])
	}
	if entries[2].Label != "bars" || entries[2].kind != legendEntryPatch {
		t.Fatalf("unexpected third legend entry: %+v", entries[2])
	}
}

func TestLegendCollectsLineMarkers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	marker := MarkerDiamond
	face := render.Color{R: 1, A: 0.7}
	edge := render.Color{B: 1, A: 0.5}
	edgeWidth := 2.0
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{
		Label:           "line markers",
		Marker:          &marker,
		MarkerFaceColor: &face,
		MarkerEdgeColor: &edge,
		MarkerEdgeWidth: &edgeWidth,
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	if entry.kind != legendEntryLine || !entry.lineMarkerSet {
		t.Fatalf("legend entry should be combined line marker, got %+v", entry)
	}
	if entry.marker != marker || entry.markerFill != face || entry.markerEdge != edge || entry.markerEdgeWidth != pointsToPixels(style.Default, edgeWidth) {
		t.Fatalf("legend marker metadata = %+v", entry)
	}
}

func TestLegendDrawRendersLabelsAndSamples(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "samples"})
	ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") || !containsString(r.texts, "samples") {
		t.Fatalf("unexpected legend labels: %v", r.texts)
	}
	if r.pathCount < 4 {
		t.Fatalf("expected legend to draw box and sample paths, got %d paths", r.pathCount)
	}
}

func TestLegendDefaultsMatchMatplotlibSpacing(t *testing.T) {
	fig := NewFigure(800, 600)
	legend := fig.AddLegend()
	fontPx := pointsToPixels(fig.RC, fig.RC.LegendSize())

	if !floatApprox(legend.Padding, 0.4*fontPx, 1e-9) {
		t.Fatalf("legend padding = %v, want %v", legend.Padding, 0.4*fontPx)
	}
	if !floatApprox(legend.Inset, 0.5*fontPx, 1e-9) {
		t.Fatalf("legend inset = %v, want %v", legend.Inset, 0.5*fontPx)
	}
	if !floatApprox(legend.SampleWidth, 2.0*fontPx, 1e-9) {
		t.Fatalf("legend sample width = %v, want %v", legend.SampleWidth, 2.0*fontPx)
	}
	if !floatApprox(legend.SampleTextGap, 0.8*fontPx, 1e-9) {
		t.Fatalf("legend sample-text gap = %v, want %v", legend.SampleTextGap, 0.8*fontPx)
	}
	if legend.CornerRadius <= 0 {
		t.Fatalf("legend corner radius = %v, want rounded Matplotlib fancybox", legend.CornerRadius)
	}
}

type legendRecordingRenderer struct {
	render.NullRenderer
	pathCount int
	texts     []string
}

func (r *legendRecordingRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *legendRecordingRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *legendRecordingRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
