package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func TestContourLabelsDrawOverlay(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels:     []float64{1.5},
		LabelLines: true,
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	var renderer contourTextRenderer
	DrawFigure(fig, &renderer)
	if len(renderer.texts) == 0 {
		t.Fatal("expected contour labels to be rendered")
	}
}

func TestAxesClabelDelegatesToContourSetAndFiltersLevels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels: []float64{1, 2, 3},
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	fontSize := 12.0
	color := render.Color{R: 0.8, G: 0.1, B: 0.2, A: 1}
	labels := ax.Clabel(contours, ClabelOptions{
		Levels:    []float64{2},
		Formatter: ticker.FuncFormatter(func(float64) string { return "L2" }),
		FontSize:  optional.Of(fontSize),
		Color:     optional.Of(color),
	})

	if len(labels) != 1 {
		t.Fatalf("labels = %d, want one filtered contour label", len(labels))
	}
	if labels[0].Text != "L2" || labels[0].Level != 2 || labels[0].Color != color {
		t.Fatalf("label = %+v, want level 2 text/color", labels[0])
	}
	if !contours.inlineLabels {
		t.Fatal("Axes.Clabel should default to inline labels like Matplotlib")
	}
	if contours.LabelFontSize != fontSize {
		t.Fatalf("label font size = %v, want %v", contours.LabelFontSize, fontSize)
	}

	var renderer contourTextRenderer
	DrawFigure(fig, &renderer)
	if len(renderer.texts) == 0 {
		t.Fatal("expected clabel text to be rendered")
	}
}

func TestAxesClabelManualPositionsPlaceNearestContourLabels(t *testing.T) {
	ax := NewFigure(640, 480).AddAxes(geom.Rect{})
	contours := ax.Contour([][]float64{
		{0, 1, 2},
		{1, 2, 3},
		{2, 3, 4},
	}, ContourOptions{
		Levels: []float64{1, 2, 3},
	})
	if contours == nil {
		t.Fatal("expected contour set")
	}

	inline := false
	labels := ax.Clabel(contours, ClabelOptions{
		Levels:          []float64{2},
		ManualPositions: []geom.Pt{{X: 1, Y: 1}},
		Inline:          optional.Of(inline),
	})

	if len(labels) != 1 {
		t.Fatalf("manual labels = %d, want 1", len(labels))
	}
	if labels[0].Level != 2 {
		t.Fatalf("manual label level = %v, want nearest requested level 2", labels[0].Level)
	}
	if contours.inlineLabels {
		t.Fatal("manual non-inline clabel should not erase contour lines")
	}
	if len(contours.labels) != 1 || contours.labels[0].Position == (geom.Pt{}) {
		t.Fatalf("stored contour labels = %+v", contours.labels)
	}
}

func TestContourInlineLabelAngleUsesMatplotlibDisplayConvention(t *testing.T) {
	screen := []geom.Pt{
		{X: 0, Y: 10},
		{X: 5, Y: 5},
		{X: 10, Y: 0},
	}
	data := []geom.Pt{
		{X: 0, Y: 0},
		{X: 5, Y: 5},
		{X: 10, Y: 10},
	}

	angle, parts := splitContourPolylineForLabel(data, screen, 1, 4, 0, true)
	if len(parts) == 0 {
		t.Fatal("expected split contour parts")
	}
	if !approx(angle, -math.Pi/4, 1e-12) {
		t.Fatalf("angle = %v, want %v", angle, -math.Pi/4)
	}
}

func TestContourInlineLabelErasesAcrossClosedPathBoundary(t *testing.T) {
	screen := []geom.Pt{
		{X: 0, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
		{X: 0, Y: 0},
	}
	data := append([]geom.Pt(nil), screen...)

	angle, parts := splitContourPolylineForLabel(data, screen, 0, 4, 1, true)
	if got, want := len(parts), 1; got != want {
		t.Fatalf("closed contour split parts = %d, want %d: %+v", got, want, parts)
	}
	want := []geom.Pt{
		{X: 3, Y: 0},
		{X: 10, Y: 0},
		{X: 10, Y: 10},
		{X: 0, Y: 10},
		{X: 0, Y: 3},
	}
	if !pointsEqual(parts[0], want, 1e-12) {
		t.Fatalf("closed contour split = %+v, want %+v", parts[0], want)
	}
	if !approx(angle, -math.Pi/4, 1e-12) {
		t.Fatalf("angle = %v, want %v", angle, -math.Pi/4)
	}
}

func TestContourInlineLabelsCoverDenseSparseAndShortContours(t *testing.T) {
	ctx := createTestDrawContext()
	renderer := &contourTextRenderer{}
	fontSize := 10.0

	dense := make([]geom.Pt, 0, 41)
	for i := 0; i <= 40; i++ {
		x := float64(i) / 10
		dense = append(dense, geom.Pt{X: x, Y: 2 + 0.05*math.Sin(x)})
	}
	sparse := []geom.Pt{
		{X: 0, Y: 5},
		{X: 4, Y: 5},
	}
	short := []geom.Pt{
		{X: 0, Y: 8},
		{X: 0.05, Y: 8},
	}
	lines := &LineCollection{
		Segments: [][]geom.Pt{dense, sparse, short},
		Colors: []render.Color{
			{R: 1, A: 1},
			{G: 1, A: 1},
			{B: 1, A: 1},
		},
		LineWidths: []float64{1, 2, 3},
	}

	segments, colors, widths, labels := contourInlineLabelSegments(
		lines,
		[]float64{1, 2, 3},
		ticker.ScalarFormatter{Prec: 0},
		fontSize,
		renderer,
		ctx,
	)
	if got, want := len(labels), 2; got != want {
		t.Fatalf("inline labels = %d, want dense and sparse labels only: %+v", got, labels)
	}
	if labels[0].Text != "1" || labels[1].Text != "2" {
		t.Fatalf("inline label texts = %q, %q; want dense/sparse levels 1 and 2", labels[0].Text, labels[1].Text)
	}
	if len(segments) <= len(lines.Segments) {
		t.Fatalf("segments were not split for inline erasure: got %d, original %d", len(segments), len(lines.Segments))
	}
	if len(colors) != len(segments) || len(widths) != len(segments) {
		t.Fatalf("split line style arrays do not match segments: segments=%d colors=%d widths=%d", len(segments), len(colors), len(widths))
	}
	if !approx(labels[1].Angle, 0, 1e-12) {
		t.Fatalf("sparse horizontal label angle = %v, want 0", labels[1].Angle)
	}
	for _, label := range labels {
		screenPos := ctx.DataToPixel.Apply(label.Position)
		if screenPos.X < 40 || screenPos.X > 460 || screenPos.Y < -60 || screenPos.Y > 460 {
			t.Fatalf("label position %+v maps outside expected display area: %+v", label.Position, screenPos)
		}
	}
}

func TestContourRotatedTextAnchorKeepsCenterFixed(t *testing.T) {
	center := geom.Pt{X: 100, Y: 200}
	angle := math.Pi / 6
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Width:   24,
			Height:  12,
			Ascent:  9,
			Descent: 3,
		},
	}

	anchor := contourRotatedTextAnchor(center, layout, angle)
	want := rotatedTextBackendAnchorFromP(center, layout, TextAlignCenter, textLayoutVAlignCenter, angle, false)
	if !pointsApprox(anchor, want, 1e-12) {
		t.Fatalf("contour rotated text anchor = %+v, want center-aligned backend anchor %+v", anchor, want)
	}
}
