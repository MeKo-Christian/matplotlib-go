package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
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
