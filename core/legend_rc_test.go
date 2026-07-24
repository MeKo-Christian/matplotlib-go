package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestLegendConsumesRCLayoutDefaults(t *testing.T) {
	rc := style.Default
	rc.Legend.Location = "lower left"
	rc.Legend.FancyBox = false
	rc.Legend.Shadow = true
	rc.Legend.NumPoints = 3
	rc.Legend.ScatterPoints = 2
	rc.Legend.MarkerScale = 1.75
	rc.Legend.TitleFontSize = 14
	rc.Legend.BorderPad = 0.6
	rc.Legend.LabelSpacing = 0.7
	rc.Legend.HandleLength = 2.5
	rc.Legend.HandleHeight = 0.9
	rc.Legend.HandleTextPad = 1.1
	rc.Legend.BorderAxesPad = 0.8
	rc.Legend.ColumnSpacing = 2.4

	fig := NewFigure(400, 300, style.WithTheme(style.Theme{RC: rc}))
	ax := fig.AddAxes(unitRect())
	legend := ax.AddLegend()
	fontPx := pointsToPixels(fig.RC, fig.RC.LegendSize())

	if legend.Location != LegendLowerLeft || legend.CornerRadius != 0 || !legend.Shadow {
		t.Fatalf("unexpected legend placement/frame defaults: %+v", legend)
	}
	if legend.NumPoints != 3 || legend.ScatterPoints != 2 {
		t.Fatalf("unexpected legend point counts: num=%d scatter=%d", legend.NumPoints, legend.ScatterPoints)
	}
	if !floatApprox(legend.MarkerScale, 1.75, 1e-9) || !floatApprox(legend.TitleFontSize, 14, 1e-9) {
		t.Fatalf("unexpected marker/title sizes: marker=%g title=%g", legend.MarkerScale, legend.TitleFontSize)
	}
	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"borderpad", legend.Padding, 0.6 * fontPx},
		{"labelspacing", legend.RowGap, 0.7 * fontPx},
		{"handlelength", legend.SampleWidth, 2.5 * fontPx},
		{"handletextpad", legend.SampleTextGap, 1.1 * fontPx},
		{"borderaxespad", legend.Inset, 0.8 * fontPx},
		{"columnspacing", legend.ColumnSpacing, 2.4 * fontPx},
	}
	for _, check := range checks {
		if !floatApprox(check.got, check.want, 1e-9) {
			t.Errorf("%s = %g, want %g", check.name, check.got, check.want)
		}
	}
	height, descent := legend.handleMetrics(fontPx)
	if want := fontPx*0.9 - 0.35*fontPx*(0.9-0.7); !floatApprox(height, want, 1e-9) {
		t.Errorf("handle height = %g, want %g", height, want)
	}
	if want := 0.35 * fontPx * (0.9 - 0.7); !floatApprox(descent, want, 1e-9) {
		t.Errorf("handle descent = %g, want %g", descent, want)
	}
}

func TestLegendRCPointCountsAndShadowAffectDrawing(t *testing.T) {
	rc := style.Default
	rc.Legend.Shadow = true
	rc.Legend.NumPoints = 3
	rc.Legend.ScatterPoints = 2
	rc.Legend.MarkerScale = 1.5

	fig := NewFigure(400, 300, style.WithTheme(style.Theme{RC: rc}))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	marker := MarkerCircle
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line", Marker: &marker})
	legend := ax.AddLegend()

	sample := geom.Rect{Min: geom.Pt{X: 10, Y: 10}, Max: geom.Pt{X: 70, Y: 30}}
	center := geom.Pt{X: 40, Y: 20}
	fontPx := pointsToPixels(fig.RC, legend.FontSize)
	if got := len(legend.lineMarkerSampleCenters(sample, center, fontPx)); got != 3 {
		t.Fatalf("line legend marker count = %d, want 3", got)
	}
	if got := len(legend.markerSampleCenters(sample, center)); got != 2 {
		t.Fatalf("scatter legend marker count = %d, want 2", got)
	}

	var renderer legendRecordingRenderer
	DrawFigure(fig, &renderer)
	shadowColor := render.Color{
		R: legend.BackgroundColor.R * 0.3,
		G: legend.BackgroundColor.G * 0.3,
		B: legend.BackgroundColor.B * 0.3,
		A: 0.5,
	}
	if !renderer.hasFillColor(shadowColor) {
		t.Fatalf("legend shadow fill %+v was not drawn; paints=%+v", shadowColor, renderer.paints)
	}
}

func TestLegendExplicitFieldsOverrideRCSeed(t *testing.T) {
	rc := style.Default
	rc.Legend.Location = "lower left"
	rc.Legend.Shadow = true
	rc.Legend.NumPoints = 3
	rc.Legend.BorderPad = 0.9

	fig := NewFigure(400, 300, style.WithTheme(style.Theme{RC: rc}))
	ax := fig.AddAxes(unitRect())
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "line"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperRight
	legend.Shadow = false
	legend.NumPoints = 1
	legend.Padding = 2

	var renderer legendRecordingRenderer
	DrawFigure(fig, &renderer)
	if legend.Location != LegendUpperRight || legend.Shadow || legend.NumPoints != 1 || legend.Padding != 2 {
		t.Fatalf("explicit legend fields were overwritten during draw: %+v", legend)
	}
	shadowColor := render.Color{
		R: legend.BackgroundColor.R * 0.3,
		G: legend.BackgroundColor.G * 0.3,
		B: legend.BackgroundColor.B * 0.3,
		A: 0.5,
	}
	if renderer.hasFillColor(shadowColor) {
		t.Fatal("explicit Shadow=false should suppress the rc-seeded shadow")
	}
}
