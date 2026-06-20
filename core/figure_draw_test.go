package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestDrawFigureDrawsAxesPatchEvenWhenItMatchesFigureBackground(t *testing.T) {
	fig := NewFigure(200, 200)
	first := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	second := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.3, Y: 0.3}, Max: geom.Pt{X: 0.7, Y: 0.7}})
	hideAxesChrome(first)
	hideAxesChrome(second)

	var r recordingRenderer
	DrawFigure(fig, &r)

	if got, want := len(r.pathCalls), 1; got != want {
		t.Fatalf("path calls = %d, want one axes patch for the later overlapping axes", got)
	}
	for i, call := range r.pathCalls {
		if call.paint.Fill != fig.RC.AxesBackground {
			t.Fatalf("path %d fill = %+v, want axes background %+v", i, call.paint.Fill, fig.RC.AxesBackground)
		}
	}
}

func TestTwinAxesPatchIsInvisibleLikeMatplotlib(t *testing.T) {
	fig := NewFigure(200, 200)
	host := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	hideAxesChrome(host)
	host.Add(&Line2D{
		XY:  []geom.Pt{{X: 0, Y: 0}, {X: 1, Y: 1}},
		W:   1,
		Col: render.Color{A: 1},
	})
	twin := host.TwinX()
	if twin == nil {
		t.Fatal("TwinX returned nil")
	}
	hideAxesChrome(twin)

	var r recordingRenderer
	DrawFigure(fig, &r)

	backgrounds := 0
	lines := 0
	for _, call := range r.pathCalls {
		if call.paint.Fill == fig.RC.AxesBackground {
			backgrounds++
		}
		if call.paint.Stroke.A > 0 && call.paint.LineWidth > 0 {
			lines++
		}
	}
	if backgrounds != 0 {
		t.Fatalf("twin axes emitted %d background patches; Matplotlib twinx hides ax2.patch", backgrounds)
	}
	if lines == 0 {
		t.Fatal("host line was not drawn")
	}
}

func hideAxesChrome(ax *Axes) {
	ax.ShowFrame = false
	for _, axis := range []*Axis{ax.XAxis, ax.YAxis, ax.XAxisTop, ax.YAxisRight} {
		if axis == nil {
			continue
		}
		axis.ShowSpine = false
		axis.ShowTicks = false
		axis.ShowLabels = false
	}
}
