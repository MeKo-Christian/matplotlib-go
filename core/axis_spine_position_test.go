package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAxisSpinePositionAxesMatchesMatplotlibAxesTransform(t *testing.T) {
	ctx := spinePositionTestContext(72)
	tests := []struct {
		name string
		side AxisSide
		want float64
	}{
		{name: "bottom", side: AxisBottom, want: snapDisplayY(90, 300)},
		{name: "top", side: AxisTop, want: snapDisplayY(90, 300)},
		{name: "left", side: AxisLeft, want: snapDisplayX(150)},
		{name: "right", side: AxisRight, want: snapDisplayX(150)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			axis := axisForSide(test.side)
			axis.SetSpinePositionAxes(0.25)

			p1, p2 := axisSpinePixelEndpoints(axis, ctx, ctx.Clip)

			assertPerpendicularSpineCoordinate(t, test.side, p1, p2, test.want)
		})
	}
}

func TestAxisSpinePositionOutwardMatchesMatplotlibPointOffsets(t *testing.T) {
	ctx := spinePositionTestContext(72)
	tests := []struct {
		name string
		side AxisSide
		want float64
	}{
		{name: "bottom", side: AxisBottom, want: snapDisplayY(28, 300)},
		{name: "top", side: AxisTop, want: snapDisplayY(252, 300)},
		{name: "left", side: AxisLeft, want: snapDisplayX(38)},
		{name: "right", side: AxisRight, want: snapDisplayX(462)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			axis := axisForSide(test.side)
			axis.SetSpinePositionOutward(12)

			p1, p2 := axisSpinePixelEndpoints(axis, ctx, ctx.Clip)

			assertPerpendicularSpineCoordinate(t, test.side, p1, p2, test.want)
		})
	}
}

func TestAxisSpinePositionOutwardScalesWithDPIAndAcceptsNegativeValues(t *testing.T) {
	ctx := spinePositionTestContext(144)
	axis := axisForSide(AxisBottom)
	axis.SetSpinePositionOutward(-12)

	p1, p2 := axisSpinePixelEndpoints(axis, ctx, ctx.Clip)

	// -12 points means 24 display pixels inward at 144 DPI.
	assertPerpendicularSpineCoordinate(t, AxisBottom, p1, p2, snapDisplayY(64, 300))
}

func TestAxisTicksUseSameAxesFractionPositionAsSpine(t *testing.T) {
	ctx := spinePositionTestContext(72)
	axis := axisForSide(AxisBottom)
	axis.SetSpinePositionAxes(0.5)

	spineValue := getSpinePosition(axis, ctx)
	tick := axisTickDisplayPoint(axis, ctx, 5, true, spineValue)
	if tick.Y != 140 {
		t.Fatalf("tick base y = %v, want axes-fraction spine y 140", tick.Y)
	}
}

func TestAxisLabelAnchorFollowsRelocatedSpineWithoutTicks(t *testing.T) {
	ctx := spinePositionTestContext(72)
	axis := axisForSide(AxisBottom)
	axis.ShowTicks = false
	axis.ShowLabels = false
	axis.SetSpinePositionAxes(0.5)
	ax := &Axes{XAxis: axis}

	anchor, _ := xLabelAnchorPoint(
		ax,
		&render.NullRenderer{},
		ctx,
		ctx.Clip,
		AxisBottom,
		figureTextAlignment{},
	)

	wantY := 140 - axisLabelPadPx(ctx)
	if anchor.Y != wantY {
		t.Fatalf("x-label anchor y = %v, want relocated-spine anchor %v", anchor.Y, wantY)
	}
}

func TestAxisSpinePositionSettersAndReset(t *testing.T) {
	axis := NewXAxis()

	axis.SetSpinePositionAxes(1.25)
	if axis.SpinePositionMode != AxisSpinePositionAxes || axis.SpinePosition != 1.25 {
		t.Fatalf("axes position = mode %v value %v", axis.SpinePositionMode, axis.SpinePosition)
	}

	axis.SetSpinePositionOutward(-8)
	if axis.SpinePositionMode != AxisSpinePositionOutward || axis.SpinePosition != -8 {
		t.Fatalf("outward position = mode %v value %v", axis.SpinePositionMode, axis.SpinePosition)
	}

	axis.ResetSpinePosition()
	if axis.SpinePositionMode != AxisSpinePositionBoundary || axis.SpinePosition != 0 {
		t.Fatalf("reset position = mode %v value %v", axis.SpinePositionMode, axis.SpinePosition)
	}
}

func spinePositionTestContext(dpi float64) *DrawContext {
	clip := geom.Rect{
		Min: geom.Pt{X: 50, Y: 40},
		Max: geom.Pt{X: 450, Y: 240},
	}
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewDisplayRectTransform(clip),
		},
		RC:         style.RC{DPI: dpi},
		Clip:       clip,
		FigureRect: geom.Rect{Max: geom.Pt{X: 500, Y: 300}},
	}
}

func axisForSide(side AxisSide) *Axis {
	if side == AxisBottom || side == AxisTop {
		axis := NewXAxis()
		axis.Side = side
		return axis
	}
	axis := NewYAxis()
	axis.Side = side
	return axis
}

func assertPerpendicularSpineCoordinate(
	t *testing.T,
	side AxisSide,
	p1, p2 geom.Pt,
	want float64,
) {
	t.Helper()
	switch side {
	case AxisBottom, AxisTop:
		if p1.Y != want || p2.Y != want {
			t.Fatalf("spine y = %v..%v, want %v", p1.Y, p2.Y, want)
		}
	case AxisLeft, AxisRight:
		if p1.X != want || p2.X != want {
			t.Fatalf("spine x = %v..%v, want %v", p1.X, p2.X, want)
		}
	}
}
