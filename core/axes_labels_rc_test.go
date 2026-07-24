package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type axesRCTextCall struct {
	text    string
	origin  geom.Pt
	fontKey string
}

type axesRCTextRenderer struct {
	render.NullRenderer
	calls        []axesRCTextCall
	rotatedCalls []axesRCTextCall
}

func (r *axesRCTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size / 2,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func (r *axesRCTextRenderer) DrawText(text string, origin geom.Pt, _ float64, _ render.Color) {
	r.calls = append(r.calls, axesRCTextCall{text: text, origin: origin})
}

func (r *axesRCTextRenderer) DrawTextWithFont(text string, origin geom.Pt, _ float64, _ render.Color, fontKey string) {
	r.calls = append(r.calls, axesRCTextCall{text: text, origin: origin, fontKey: fontKey})
}

func (r *axesRCTextRenderer) DrawTextRotated(text string, anchor geom.Pt, _, _ float64, _ render.Color) {
	r.rotatedCalls = append(r.rotatedCalls, axesRCTextCall{text: text, origin: anchor})
}

func (r *axesRCTextRenderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, _, _ float64, _ render.Color, fontKey string) {
	r.rotatedCalls = append(r.rotatedCalls, axesRCTextCall{text: text, origin: anchor, fontKey: fontKey})
}

func TestAxesTitleAndLabelPlacementConsumesRCDefaults(t *testing.T) {
	fig := NewFigure(240, 160)
	fig.RC.DPI = 72
	fig.RC.Axes.TitleLocation = "right"
	fig.RC.Axes.TitlePad = 9
	fig.RC.Axes.TitleWeight = 700
	fig.RC.Axes.TitleY = 0.8
	fig.RC.Axes.TitleYSet = true
	fig.RC.Axes.LabelPad = 7
	fig.RC.Axes.LabelWeight = 600

	px := geom.Rect{Min: geom.Pt{X: 20, Y: 30}, Max: geom.Pt{X: 220, Y: 130}}
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetTitle("Title")
	ax.SetXLabel("X label")
	ax.SetYLabel("Y label")
	ax.XAxis.ShowTicks, ax.XAxis.ShowLabels = false, false
	ax.YAxis.ShowTicks, ax.YAxis.ShowLabels = false, false
	ctx := newAxesDrawContext(ax, fig, geom.Rect{Max: fig.SizePx}, px)

	titleAnchor := titleAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, figureTextAlignment{})
	if got, want := titleAnchor, (geom.Pt{X: px.Max.X, Y: px.Min.Y + 0.8*px.H() + 9}); got != want {
		t.Fatalf("rc title anchor = %+v, want %+v", got, want)
	}
	xAnchor, _ := xLabelAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, AxisBottom, figureTextAlignment{})
	if got, want := xAnchor.Y, xAxisSpinePixelY(ax.XAxis, ctx, AxisBottom, px)-7; math.Abs(got-want) > 1e-9 {
		t.Fatalf("rc xlabel anchor y = %v, want %v", got, want)
	}
	yAnchor := yLabelAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, AxisLeft, figureTextAlignment{})
	if got, want := yAnchor.X, yAxisSpinePixelX(ax.YAxis, ctx, AxisLeft, px)-7; math.Abs(got-want) > 1e-9 {
		t.Fatalf("rc ylabel anchor x = %v, want %v", got, want)
	}

	r := &axesRCTextRenderer{}
	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})
	if len(r.calls) != 2 || len(r.rotatedCalls) != 1 {
		t.Fatalf("unexpected text calls: plain=%+v rotated=%+v", r.calls, r.rotatedCalls)
	}
	if got := render.ParseFontProperties(r.calls[0].fontKey).Weight; got != 700 {
		t.Fatalf("title font weight = %d, want 700", got)
	}
	if got := render.ParseFontProperties(r.calls[1].fontKey).Weight; got != 600 {
		t.Fatalf("xlabel font weight = %d, want 600", got)
	}
	if got := render.ParseFontProperties(r.rotatedCalls[0].fontKey).Weight; got != 600 {
		t.Fatalf("ylabel font weight = %d, want 600", got)
	}
}

func TestAxesTitleAndLabelExplicitPlacementOverridesRC(t *testing.T) {
	fig := NewFigure(240, 160)
	fig.RC.DPI = 72
	fig.RC.Axes.TitleLocation = "right"
	fig.RC.Axes.TitlePad = 9
	fig.RC.Axes.TitleWeight = 700
	fig.RC.Axes.TitleY = 0.8
	fig.RC.Axes.TitleYSet = true
	fig.RC.Axes.LabelPad = 7
	fig.RC.Axes.LabelWeight = 600

	px := geom.Rect{Min: geom.Pt{X: 20, Y: 30}, Max: geom.Pt{X: 220, Y: 130}}
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetTitle("Title")
	ax.SetXLabel("X label")
	ax.SetYLabel("Y label")
	if err := ax.SetTitleLocation("left"); err != nil {
		t.Fatalf("SetTitleLocation: %v", err)
	}
	ax.SetTitleAutoY()
	ax.SetTitlePad(2)
	ax.SetTitleWeight(800)
	ax.SetXLabelPad(3)
	ax.SetYLabelPad(4)
	ax.SetXLabelWeight(900)
	ax.SetYLabelWeight(500)
	ax.XAxis.ShowTicks, ax.XAxis.ShowLabels = false, false
	ax.YAxis.ShowTicks, ax.YAxis.ShowLabels = false, false
	ctx := newAxesDrawContext(ax, fig, geom.Rect{Max: fig.SizePx}, px)

	titleAnchor := titleAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, figureTextAlignment{})
	if got, want := titleAnchor, (geom.Pt{X: px.Min.X, Y: px.Max.Y + 2}); got != want {
		t.Fatalf("explicit title anchor = %+v, want %+v", got, want)
	}
	xAnchor, _ := xLabelAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, AxisBottom, figureTextAlignment{})
	if got, want := xAnchor.Y, xAxisSpinePixelY(ax.XAxis, ctx, AxisBottom, px)-3; math.Abs(got-want) > 1e-9 {
		t.Fatalf("explicit xlabel anchor y = %v, want %v", got, want)
	}
	yAnchor := yLabelAnchorPoint(ax, &axesRCTextRenderer{}, ctx, px, AxisLeft, figureTextAlignment{})
	if got, want := yAnchor.X, yAxisSpinePixelX(ax.YAxis, ctx, AxisLeft, px)-4; math.Abs(got-want) > 1e-9 {
		t.Fatalf("explicit ylabel anchor x = %v, want %v", got, want)
	}

	r := &axesRCTextRenderer{}
	drawAxesLabels(ax, r, ctx, px, figureTextAlignment{})
	if got := render.ParseFontProperties(r.calls[0].fontKey).Weight; got != 800 {
		t.Fatalf("explicit title font weight = %d, want 800", got)
	}
	if got := render.ParseFontProperties(r.calls[1].fontKey).Weight; got != 900 {
		t.Fatalf("explicit xlabel font weight = %d, want 900", got)
	}
	if got := render.ParseFontProperties(r.rotatedCalls[0].fontKey).Weight; got != 500 {
		t.Fatalf("explicit ylabel font weight = %d, want 500", got)
	}

	ax.Clear()
	if ax.titleLocation != "right" || ax.titlePadPt != 9 || ax.titleWeight != 700 ||
		!ax.titleYSet || ax.titleY != 0.8 ||
		ax.xLabelPadPt != 7 || ax.yLabelPadPt != 7 ||
		ax.xLabelWeight != 600 || ax.yLabelWeight != 600 {
		t.Fatalf("Clear did not restore rc text defaults: %+v", ax)
	}
}
