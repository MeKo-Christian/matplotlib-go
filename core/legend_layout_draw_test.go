package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

func TestLegendDrawSupportsMultipleColumns(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "a"})
	ax.Plot([]float64{0, 1}, []float64{1, 2}, PlotOptions{Label: "b"})
	ax.Plot([]float64{0, 1}, []float64{2, 3}, PlotOptions{Label: "c"})
	ax.Plot([]float64{0, 1}, []float64{3, 4}, PlotOptions{Label: "d"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.NumColumns = 2

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	a := r.textOrigin("a")
	b := r.textOrigin("b")
	c := r.textOrigin("c")
	d := r.textOrigin("d")
	if !floatApprox(a.X, b.X, 1e-9) {
		t.Fatalf("first column labels should share x origin, got a=%+v b=%+v", a, b)
	}
	if !floatApprox(c.X, d.X, 1e-9) || c.X <= a.X {
		t.Fatalf("second column labels should share a later x origin, got a=%+v c=%+v d=%+v", a, c, d)
	}
	if !floatApprox(a.Y, c.Y, 1e-9) || !floatApprox(b.Y, d.Y, 1e-9) {
		t.Fatalf("multi-column rows should align, got a=%+v b=%+v c=%+v d=%+v", a, b, c, d)
	}
}

func TestLegendMathLabelWidthUsesMeasuredTextWidth(t *testing.T) {
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{Width: 81},
		MathLayout:     &MathTextLayout{},
	}

	if got, want := legendLabelWidth(layout), layout.Width; got != want {
		t.Fatalf("math legend label width = %v, want measured width %v", got, want)
	}
}

func TestLegendDrawSupportsTitle(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.Title = "Series"

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "Series") || !containsString(r.texts, "signal") {
		t.Fatalf("legend title and label should be drawn, got %v", r.texts)
	}
	title := r.textOrigin("Series")
	label := r.textOrigin("signal")
	// Display space is y-up: the title sits above the first entry at a larger Y.
	if title.Y <= label.Y {
		t.Fatalf("legend title should be above first entry label, got title=%+v label=%+v", title, label)
	}

	withoutTitle := NewLegend(ax)
	withTitle := NewLegend(ax)
	withTitle.Title = "Series"
	boxWithout, okWithout := withoutTitle.boxRect(&r, &DrawContext{RC: fig.RC, Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 400, Y: 300}}})
	boxWith, okWith := withTitle.boxRect(&r, &DrawContext{RC: fig.RC, Clip: geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 400, Y: 300}}})
	if !okWithout || !okWith {
		t.Fatalf("expected legend boxes for titled and untitled legends, got %v %v", okWithout, okWith)
	}
	if boxWith.H() <= boxWithout.H() {
		t.Fatalf("titled legend height = %g, want larger than untitled height %g", boxWith.H(), boxWithout.H())
	}
}

func TestLegendFrameOnFalseSkipsFrameOnly(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()
	legend.Location = LegendUpperLeft
	legend.FrameOn = false

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	if !containsString(r.texts, "signal") {
		t.Fatalf("legend label should still be drawn when frame is disabled, got %v", r.texts)
	}
	if r.hasLegendFramePaint(legend) {
		t.Fatalf("legend frame paint should not be drawn when FrameOn is false")
	}
	if len(r.paths) == 0 {
		t.Fatalf("legend samples should still be drawn when frame is disabled")
	}
}

func TestLegendFrameUsesMatplotlibSnapAuto(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	for i, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			if paint.Snap != render.SnapAuto {
				t.Fatalf("legend frame paint %d snap = %v, want Matplotlib SnapAuto", i, paint.Snap)
			}
			return
		}
	}
	t.Fatal("legend frame paint was not drawn")
}

func TestLegendFrameUsesMatplotlibRoundBoxStyle(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})
	legend := ax.AddLegend()

	var r legendRecordingRenderer
	DrawFigure(fig, &r)

	for i, paint := range r.paints {
		if paint.Fill == legend.BackgroundColor && paint.Stroke == legend.BorderColor && paint.LineWidth == legend.BorderWidth {
			if got := countPathCmd(r.paths[i], geom.QuadTo); got != 4 {
				t.Fatalf("legend frame quadratic corners = %d, want Matplotlib BoxStyle.Round with four CURVE3 corners; commands=%v", got, r.paths[i].C)
			}
			return
		}
	}
	t.Fatal("legend frame path was not drawn")
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
	if !floatApprox(legend.BorderWidth, pointsToPixels(fig.RC, 1), 1e-9) {
		t.Fatalf("legend border width = %v, want Matplotlib 1 point linewidth %v", legend.BorderWidth, pointsToPixels(fig.RC, 1))
	}
	if legend.CornerRadius <= 0 {
		t.Fatalf("legend corner radius = %v, want rounded Matplotlib fancybox", legend.CornerRadius)
	}
}

func TestAxesLegendDrawsOutsideAxesClip(t *testing.T) {
	fig := NewFigure(240, 240)
	ax := fig.AddPolarAxes(unitRect())
	ax.SetYLim(0, 1)
	color := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	ax.Plot([]float64{0, 1}, []float64{0.2, 0.8}, PlotOptions{
		Color: &color,
		Label: "legend label",
	})
	ax.AddLegend()

	r := &legendClipTrackingRenderer{}
	DrawFigure(fig, r)

	if clipped, ok := r.textClipped["legend label"]; !ok {
		t.Fatalf("legend label was not drawn; saw texts %v", r.texts)
	} else if clipped {
		t.Fatal("legend label was drawn while the axes clip was active")
	}
}
