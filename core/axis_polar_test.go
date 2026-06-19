package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/render"
)

func TestPolarAxisUsesConfiguredLineStyle(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ax.XAxis.SetLineStyle(render.CapRound, render.JoinBevel, 5, 2)

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &recordingRenderer{}

	ax.XAxis.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one polar spine path, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapRound || r.pathCalls[0].paint.LineJoin != render.JoinBevel {
		t.Fatalf("polar spine paint = %+v", r.pathCalls[0].paint)
	}
	if len(r.pathCalls[0].paint.Dashes) != 2 || r.pathCalls[0].paint.Dashes[0] != 5 || r.pathCalls[0].paint.Dashes[1] != 2 {
		t.Fatalf("polar spine dashes = %v", r.pathCalls[0].paint.Dashes)
	}
}
