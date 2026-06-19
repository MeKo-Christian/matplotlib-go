package core

import (
	"reflect"
	"testing"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
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

func TestLegendDrawKeepsCollectionOrderAfterZSorting(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.Scatter([]float64{0.5}, []float64{0.5}, ScatterOptions{Label: "scatter"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "handler"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.NumColumns = 2

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	entries := legend.collectEntries()
	labels := make([]string, len(entries))
	for i, entry := range entries {
		labels[i] = entry.Label
	}
	want := []string{"line", "scatter", "handler"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("legend collection order after draw = %v, want insertion order %v", labels, want)
	}
}

func TestLegendCollectsErrorBarsAfterPlainArtistsLikeMatplotlibContainers(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	ax.ErrorBar([]float64{0.5}, []float64{0.5}, nil, []float64{0.1}, ErrorBarOptions{Label: "errorbar"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "handler"})

	entries := ax.AddLegend().collectEntries()
	labels := make([]string, len(entries))
	for i, entry := range entries {
		labels[i] = entry.Label
	}
	want := []string{"line", "handler", "errorbar"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("legend collection order = %v, want Matplotlib child/container order %v", labels, want)
	}
}

func TestLegendCollectsStemAsSingleCombinedSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Stem([]float64{0, 1, 2}, []float64{1, 3, 2}, StemOptions{Label: "stem"})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("stem legend entries = %d, want one combined sample: %+v", len(entries), entries)
	}
	entry := entries[0]
	if entry.Label != "stem" || entry.kind != legendEntryErrorBar || !entry.errorbarY || !entry.lineMarkerSet {
		t.Fatalf("stem legend entry = %+v, want combined stem line+marker sample", entry)
	}
}

func TestLegendAddEntryDrawsProxyPatchSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	legend := ax.AddLegend()
	proxyFill := render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}
	proxyEdge := render.Color{R: 0.05, G: 0.1, B: 0.2, A: 1}
	legend.AddEntry("proxy", LegendEntryOptions{
		Sample:    LegendSamplePatch,
		FaceColor: proxyFill,
		EdgeColor: proxyEdge,
		EdgeWidth: 2,
	})

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "proxy") {
		t.Fatalf("explicit proxy legend entry should be drawn, got labels %v", r.texts)
	}
	if !r.hasFillColor(proxyFill) {
		t.Fatalf("explicit proxy legend patch sample should use fill color %+v, got paints %+v", proxyFill, r.paints)
	}
}

func TestLegendSetHandlerOverridesCollectedArtistSample(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	line := ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "custom"})
	legend := ax.AddLegend()
	overrideFill := render.Color{R: 0.7, G: 0.2, B: 0.1, A: 1}
	legend.SetHandler(line, LegendEntryOptions{
		Sample:    LegendSamplePatch,
		FaceColor: overrideFill,
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	})

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "custom") {
		t.Fatalf("legend should still collect the artist label, got labels %v", r.texts)
	}
	if !r.hasFillColor(overrideFill) {
		t.Fatalf("custom legend handler sample should use fill color %+v, got paints %+v", overrideFill, r.paints)
	}
}

func TestLegendCollectionSamplesUseScalarMappedColors(t *testing.T) {
	cmapName := "legend-collection-scalar"
	low := render.Color{R: 1, A: 1}
	high := render.Color{B: 1, A: 1}
	matcolor.RegisterColormap(cmapName, matcolor.NewColormap(cmapName, []matcolor.ColorStop{
		{Pos: 0, Color: low},
		{Pos: 1, Color: high},
	}))

	pathCollection := &PathCollection{
		Collection: Collection{
			Label:        "mapped",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{0, 10},
		},
		Path:      markerCirclePath(1),
		Offsets:   []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}
	pathEntry, ok := pathCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped path collection legend entry not collected")
	}
	if got := pathEntry.markerFill; got != low {
		t.Fatalf("legend marker fill = %+v, want first scalar-mapped face %+v", got, low)
	}

	explicit := render.Color{G: 1, A: 1}
	pathCollection.FaceColor = explicit
	pathEntry, ok = pathCollection.legendEntry()
	if !ok {
		t.Fatal("explicit-colored path collection legend entry not collected")
	}
	if got := pathEntry.markerFill; got != explicit {
		t.Fatalf("legend marker fill with explicit face = %+v, want explicit color %+v", got, explicit)
	}
	pathCollection.FaceColor = render.Color{}

	patchCollection := &PatchCollection{
		Collection: Collection{
			Label:        "mapped patch",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{10},
		},
		Paths:     []geom.Path{polygonPath([]geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}, true)},
		EdgeColor: render.Color{A: 1},
		EdgeWidth: 1,
	}
	patchEntry, ok := patchCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped patch collection legend entry not collected")
	}
	if got := patchEntry.patchFill; got != high {
		t.Fatalf("legend patch fill = %+v, want first scalar-mapped face %+v", got, high)
	}

	lineCollection := &LineCollection{
		Collection: Collection{
			Label:        "mapped line",
			Colormap:     cmapName,
			VMin:         0,
			VMax:         10,
			ScalarValues: []float64{10},
		},
		Segments:  [][]geom.Pt{{{X: 0, Y: 0}, {X: 1, Y: 1}}},
		LineWidth: 1,
	}
	lineEntry, ok := lineCollection.legendEntry()
	if !ok {
		t.Fatal("scalar-mapped line collection legend entry not collected")
	}
	if got := lineEntry.lineColor; got != high {
		t.Fatalf("legend line color = %+v, want first scalar-mapped stroke %+v", got, high)
	}
}
