package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestLegendBestPlacementAvoidsLineAndScatterPoints(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	line := &Line2D{
		XY:    []geom.Pt{{X: 0.86, Y: 0.94}, {X: 0.96, Y: 0.94}},
		Label: "line",
	}
	scatter := &Scatter2D{
		XY:    []geom.Pt{{X: 0.9, Y: 0.95}},
		Label: "scatter",
	}
	ax.Artists = []Artist{line, scatter, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if !approxRect(box, want, 1e-9) {
		t.Fatalf("best legend box = %+v, want upper-left %+v when line/scatter occupy upper-right", box, want)
	}
}

func TestLegendBestPlacementCountsLineIntersections(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	crossingLowerLeft := &Line2D{
		XY:    []geom.Pt{{X: -0.1, Y: 0.06}, {X: 0.3, Y: 0.06}},
		Label: "crossing",
	}
	pointsInOtherCorners := &Scatter2D{
		XY: []geom.Pt{
			{X: 0.9, Y: 0.95},
			{X: 0.1, Y: 0.95},
			{X: 0.9, Y: 0.06},
		},
		Label: "points",
	}
	ax.Artists = []Artist{crossingLowerLeft, pointsInOtherCorners, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendRight, legend.Inset)
	if !approxRect(box, want, 1e-9) {
		t.Fatalf("best legend box = %+v, want center-right %+v when lower-left is crossed by a line segment", box, want)
	}
}

func TestLegendBestPlacementMatchesMathtextInlineLabels(t *testing.T) {
	fig := NewFigure(640, 360)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.92, Y: 0.88},
	})
	ax.SetXLim(0, 5)
	ax.SetYLim(0.165, 0.925)

	const n = 90
	x := make([]float64, n)
	y1 := make([]float64, n)
	y2 := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * 5
		x[i] = t
		y1[i] = 0.55 + 0.35*math.Sin(1.5*t)
		y2[i] = 0.48 + 0.28*math.Cos(1.5*t+0.45)
	}
	ax.Plot(x, y1, PlotOptions{Label: `state $x_i(t)$`})
	ax.Plot(x, y2, PlotOptions{Label: `state $y_i(t)$`})
	ax.Text(0.03, 0.88, `peak $\alpha_i^2$`, TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignLeft,
		VAlign: TextVAlignTop,
	})
	ax.Text(0.97, 0.08, `ratio $\frac{a}{b}$`, TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignRight,
		VAlign: TextVAlignBottom,
	})
	legend := ax.AddLegend()

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	var r legendRecordingRenderer
	box, ok := legend.boxRect(&r, ctx)
	if !ok {
		t.Fatal("legend.boxRect returned !ok")
	}
	want := anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerLeft, legend.Inset)
	if !approxRect(box, want, 1e-9) {
		data := legend.legendAvoidanceData(ctx)
		t.Fatalf("best legend box = %+v, want lower-left %+v; scores UR=%d UL=%d LL=%d LR=%d UC=%d C=%d lines=%+v",
			box, want,
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperRight, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperLeft, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerLeft, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendLowerRight, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendUpperCenter, legend.Inset), data),
			legendPlacementBadness(anchoredBoxRect(ctx.Clip, box.W(), box.H(), LegendCenter, legend.Inset), data),
			legendDebugPathBounds(data.lines))
	}
}

func TestLegendBestPlacementAvoidsImageExtent(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	image := &Image2D{
		XMin: 0.85,
		XMax: 0.95,
		YMin: 0.85,
		YMax: 0.95,
	}
	ax.Artists = []Artist{image, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when image occupies upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsPatchBounds(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	patch := &Rectangle{
		XY:     geom.Pt{X: 0.84, Y: 0.86},
		Width:  0.14,
		Height: 0.12,
		Coords: Coords(CoordData),
	}
	ax.Artists = []Artist{patch, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when patch bbox occupies upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsPatchPaths(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	patch := &PathPatch{
		Path: geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: 0.84, Y: 0.94}, {X: 0.99, Y: 0.94}},
		},
		Coords: Coords(CoordData),
	}
	ax.Artists = []Artist{patch, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when patch path occupies upper-right", box, want)
	}
}

func TestLegendBestPlacementAvoidsAnnotationAnchors(t *testing.T) {
	ax := &Axes{}
	legend := NewLegend(ax)
	annotation := &Annotation{
		Point:   geom.Pt{X: 0.9, Y: 0.95},
		Coords:  Coords(CoordData),
		OffsetX: 8,
		OffsetY: 8,
	}
	boxPosition := geom.Pt{X: 0.9, Y: 0.95}
	annotationBox := &AnnotationBbox{
		Point:       geom.Pt{X: 0.88, Y: 0.92},
		XYCoords:    Coords(CoordData),
		BoxCoords:   Coords(CoordData),
		BoxPosition: &boxPosition,
	}
	ax.Artists = []Artist{annotation, annotationBox, legend}

	ctx := legendBestPlacementTestContext()
	box := legend.bestLegendBoxRect(ctx, 80, 40)
	want := anchoredBoxRect(ctx.Clip, 80, 40, LegendUpperLeft, legend.Inset)
	if box != want {
		t.Fatalf("best legend box = %+v, want upper-left %+v when annotations occupy upper-right", box, want)
	}
}
