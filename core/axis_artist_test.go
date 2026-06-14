package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAxisSetTickDirectionControlsTickSegment(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{1}
	axis.Formatter = nil

	ctx := createTestDrawContext()

	var r recordingRenderer
	axis.DrawTicks(&r, ctx)
	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one outward tick path, got %d", len(r.pathCalls))
	}
	snappedTickSize := math.Round(axis.TickSize)
	// Display space is y-up: the bottom spine snaps to 449.5 and outward ticks
	// point away from the plot (downward = smaller Y).
	outward := r.pathCalls[0].path.V
	if !floatApprox(outward[0].Y, 449.5, 1e-9) || !floatApprox(outward[1].Y, 449.5-snappedTickSize, 1e-9) {
		t.Fatalf("outward tick = %+v", outward)
	}

	r.pathCalls = nil
	if err := axis.SetTickDirection("in"); err != nil {
		t.Fatalf("SetTickDirection(in): %v", err)
	}
	axis.DrawTicks(&r, ctx)
	inward := r.pathCalls[0].path.V
	if !floatApprox(inward[0].Y, 449.5, 1e-9) || !floatApprox(inward[1].Y, 449.5+snappedTickSize, 1e-9) {
		t.Fatalf("inward tick = %+v", inward)
	}

	r.pathCalls = nil
	if err := axis.SetTickDirection("inout"); err != nil {
		t.Fatalf("SetTickDirection(inout): %v", err)
	}
	axis.DrawTicks(&r, ctx)
	inout := r.pathCalls[0].path.V
	if !floatApprox(inout[0].Y, 449.5+snappedTickSize/2, 1e-9) || !floatApprox(inout[1].Y, 449.5-snappedTickSize/2, 1e-9) {
		t.Fatalf("inout tick = %+v", inout)
	}
}

func TestAxisSpinePositionDataMovesXAxisSpine(t *testing.T) {
	axis := NewXAxis()
	axis.ShowTicks = false
	axis.ShowLabels = false
	axis.SetSpinePositionData(3)

	ctx := createTestDrawContext()
	var r recordingRenderer
	axis.Draw(&r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one spine path, got %d", len(r.pathCalls))
	}
	spine := r.pathCalls[0].path.V
	wantY := ctx.DataToPixel.Apply(geom.Pt{X: 0, Y: 3}).Y
	wantY = snapDisplayY(wantY, figureSnapHeight(ctx))
	if !floatApprox(spine[0].Y, wantY, 1e-9) || !floatApprox(spine[1].Y, wantY, 1e-9) {
		t.Fatalf("floating x spine = %+v", spine)
	}
}

func TestAxesFloatingAxisArtistRendersThroughDrawFigure(t *testing.T) {
	fig := NewFigure(500, 500)
	ax := fig.AddAxes(unitRect())
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)

	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	floating := ax.FloatingXAxis(3)
	if floating == nil || floating.Axis == nil {
		t.Fatal("FloatingXAxis() returned nil")
	}
	floating.Axis.ShowTicks = false
	floating.Axis.ShowLabels = false

	var r recordingRenderer
	DrawFigure(fig, &r)

	if len(ax.ExtraAxes) != 1 {
		t.Fatalf("len(ExtraAxes) = %d, want 1", len(ax.ExtraAxes))
	}
	// Display space is y-up: data y=3 maps to 169.5 (the flip of the y-down 330.5).
	if !hasHorizontalPathAtY(r.pathCalls, 169.5) {
		t.Fatalf("expected floating axis spine around y=169.5, got %+v", r.pathCalls)
	}
}

func hasHorizontalPathAtY(calls []recordedPathCall, y float64) bool {
	for _, call := range calls {
		if len(call.path.V) != 2 {
			continue
		}
		if floatApprox(call.path.V[0].Y, y, 1e-6) && floatApprox(call.path.V[1].Y, y, 1e-6) {
			return true
		}
	}
	return false
}
