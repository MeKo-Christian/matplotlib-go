package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestLine2D_PathPoints_StepStyles(t *testing.T) {
	base := []geom.Pt{
		{X: 0, Y: 1},
		{X: 2, Y: 3},
		{X: 4, Y: 2},
	}

	tests := []struct {
		name  string
		style LineDrawStyle
		want  []geom.Pt
	}{
		{
			name:  "steps-pre",
			style: LineDrawStyleStepsPre,
			want: []geom.Pt{
				{X: 0, Y: 1},
				{X: 0, Y: 3},
				{X: 2, Y: 3},
				{X: 2, Y: 2},
				{X: 4, Y: 2},
			},
		},
		{
			name:  "steps-mid",
			style: LineDrawStyleStepsMid,
			want: []geom.Pt{
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 1, Y: 3},
				{X: 2, Y: 3},
				{X: 3, Y: 3},
				{X: 3, Y: 2},
				{X: 4, Y: 2},
			},
		},
		{
			name:  "steps-post",
			style: LineDrawStyleStepsPost,
			want: []geom.Pt{
				{X: 0, Y: 1},
				{X: 2, Y: 1},
				{X: 2, Y: 3},
				{X: 4, Y: 3},
				{X: 4, Y: 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			line := &Line2D{XY: base, DrawStyle: tc.style}
			got := line.pathPoints()
			if len(got) != len(tc.want) {
				t.Fatalf("got %d points, want %d", len(got), len(tc.want))
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("point %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestAxesStep_UsesRequestedWhere(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	where := StepWhereMid

	line := ax.Step([]float64{0, 1, 2}, []float64{1, 3, 2}, StepOptions{Where: &where})
	if line == nil {
		t.Fatal("expected step line")
	}
	if line.DrawStyle != LineDrawStyleStepsMid {
		t.Fatalf("draw style = %v, want %v", line.DrawStyle, LineDrawStyleStepsMid)
	}
}

func TestFillBetweenX_Bounds(t *testing.T) {
	fill := FillBetweenX(
		[]float64{0, 2, 4},
		[]float64{1, 3, 2},
		[]float64{-1, 0, 1},
		render.Color{R: 1, G: 0, B: 0, A: 1},
	)

	if fill.Orientation != FillHorizontal {
		t.Fatalf("orientation = %v, want horizontal", fill.Orientation)
	}

	got := fill.Bounds(nil)
	want := geom.Rect{
		Min: geom.Pt{X: -1, Y: 0},
		Max: geom.Pt{X: 3, Y: 4},
	}
	if got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func TestBar2D_Bounds_UsePerBarBaselines(t *testing.T) {
	bar := &Bar2D{
		X:           []float64{1, 2},
		Heights:     []float64{3, -2},
		Baselines:   []float64{1, 4},
		Width:       0.5,
		Orientation: BarVertical,
	}

	got := bar.Bounds(nil)
	want := geom.Rect{
		Min: geom.Pt{X: 0.75, Y: 1},
		Max: geom.Pt{X: 2.25, Y: 4},
	}
	if got != want {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func TestAxesBrokenBarH_UsesRangesAsBaselines(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})

	bar := ax.BrokenBarH([][2]float64{{1, 2.5}, {5, 1.25}}, [2]float64{3, 0.8})
	if bar == nil {
		t.Fatal("expected broken_barh bar artist")
	}
	if bar.Orientation != BarHorizontal {
		t.Fatalf("orientation = %v, want horizontal", bar.Orientation)
	}
	if bar.Width != 0.8 {
		t.Fatalf("width = %v, want 0.8", bar.Width)
	}
	if len(bar.Baselines) != 2 || bar.Baselines[0] != 1 || bar.Baselines[1] != 5 {
		t.Fatalf("baselines = %v", bar.Baselines)
	}
	if len(bar.Heights) != 2 || bar.Heights[0] != 2.5 || bar.Heights[1] != 1.25 {
		t.Fatalf("widths/heights = %v", bar.Heights)
	}
	if len(bar.X) != 2 || bar.X[0] != 3.4 || bar.X[1] != 3.4 {
		t.Fatalf("y centers = %v", bar.X)
	}
}

func TestAxesBarLabel_Placement(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	bar := &Bar2D{
		X:           []float64{1, 2},
		Heights:     []float64{3, -2},
		Baselines:   []float64{1, 4},
		Orientation: BarVertical,
	}
	ax.Add(bar)

	labels := ax.BarLabel(bar, []string{"up", "down"})
	if len(labels) != 2 {
		t.Fatalf("got %d labels, want 2", len(labels))
	}
	if labels[0].Content != "up" || labels[1].Content != "down" {
		t.Fatalf("unexpected labels: %q %q", labels[0].Content, labels[1].Content)
	}
	if labels[0].Position != (geom.Pt{X: 1, Y: 4}) || labels[0].VAlign != TextVAlignBottom || labels[0].OffsetY != 0 {
		t.Fatalf("positive bar label = %+v", *labels[0])
	}
	if labels[1].Position != (geom.Pt{X: 2, Y: 2}) || labels[1].VAlign != TextVAlignTop || labels[1].OffsetY != 0 {
		t.Fatalf("negative bar label = %+v", *labels[1])
	}

	padded := ax.BarLabel(bar, []string{"padded"}, BarLabelOptions{Padding: 4})
	if len(padded) == 0 || padded[0].OffsetY != -4 {
		t.Fatalf("explicit padded label = %+v", padded)
	}

	centered := ax.BarLabel(bar, nil, BarLabelOptions{Position: "center", Format: "%.1f"})
	if len(centered) != 2 {
		t.Fatalf("got %d centered labels, want 2", len(centered))
	}
	if centered[0].Content != "3.0" || centered[0].Position != (geom.Pt{X: 1, Y: 2.5}) || centered[0].VAlign != TextVAlignMiddle {
		t.Fatalf("center label = %+v", *centered[0])
	}
}

func TestAxesAxHLine_UsesBlendedCoordinates(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	line := ax.AxHLine(2)

	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].path.V
	want := []geom.Pt{{X: 50, Y: 370}, {X: 150, Y: 370}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("path = %+v, want %+v", got, want)
	}
}

func TestAxesAxLine_ClipsToCurrentView(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	line := ax.AxLine(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 10, Y: 10})

	r := &recordingRenderer{}
	ctx := createTestDrawContext()
	line.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].path.V
	want := []geom.Pt{{X: 50, Y: 450}, {X: 150, Y: 350}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("path = %+v, want %+v", got, want)
	}
}

func TestStairs2D_DrawFilled(t *testing.T) {
	stairs := &Stairs2D{
		Edges:     []float64{0, 1, 3},
		Values:    []float64{2, 4},
		Baseline:  1,
		Fill:      true,
		Color:     render.Color{R: 1, G: 0, B: 0, A: 1},
		EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth: 2,
	}

	r := &recordingRenderer{}
	stairs.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	call := r.pathCalls[0]
	if call.paint.Fill.A != 1 || call.paint.Stroke.A != 1 {
		t.Fatalf("unexpected paint = %+v", call.paint)
	}
	if len(call.path.C) == 0 || call.path.C[len(call.path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed fill path, got %+v", call.path.C)
	}
}

func TestAxesAxVSpan_DrawsFilledRect(t *testing.T) {
	ax := NewFigure(640, 360).AddAxes(geom.Rect{})
	alpha := 0.5
	span := ax.AxVSpan(2, 4, VSpanOptions{Alpha: &alpha})

	r := &recordingRenderer{}
	span.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.Fill.A == 0 {
		t.Fatalf("expected non-transparent fill paint")
	}
}
