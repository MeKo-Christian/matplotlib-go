package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestDrawContextCoordinateHelpersSupportBlendedNestedProjectionTransforms(t *testing.T) {
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			DataToAxes: transform.ChainSeparable(
				transform.NewScaleTransform(transform.NewLinear(10, 20), transform.NewLinear(-1, 1)),
				transform.NewSeparable(
					transform.OffsetAxis{Delta: 0.1},
					transform.OffsetAxis{Delta: 0.2},
				),
			),
			AxesToPixel: transform.NewDisplayRectTransform(geom.Rect{
				Min: geom.Pt{X: 100, Y: 100},
				Max: geom.Pt{X: 300, Y: 500},
			}),
		},
		FigureRect: geom.Rect{
			Min: geom.Pt{X: 0, Y: 0},
			Max: geom.Pt{X: 800, Y: 600},
		},
		Clip: geom.Rect{
			Min: geom.Pt{X: 100, Y: 100},
			Max: geom.Pt{X: 300, Y: 500},
		},
	}

	// Display space is y-up: NewDisplayRectTransform maps fraction (0,1) to
	// (Min.Y, Max.Y), so display Y grows upward.
	data := ctx.TransData().Apply(geom.Pt{X: 15, Y: 0})
	if data.X != 220 || data.Y != 380 {
		t.Fatalf("transData = %+v, want {220 380}", data)
	}

	axes := ctx.TransAxes().Apply(geom.Pt{X: 0.5, Y: 0.5})
	if axes.X != 200 || axes.Y != 300 {
		t.Fatalf("transAxes = %+v, want {200 300}", axes)
	}

	figure := ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.25})
	if figure.X != 200 || figure.Y != 150 {
		t.Fatalf("transFigure = %+v, want {200 150}", figure)
	}

	blended := ctx.TransformFor(BlendCoords(CoordData, CoordAxes)).Apply(geom.Pt{X: 15, Y: 0.75})
	if blended.X != 220 || blended.Y != 400 {
		t.Fatalf("blended transform = %+v, want {220 400}", blended)
	}
}

func TestTransformForCoordinateSpecsExposeExpectedTransforms(t *testing.T) {
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(-5, 5),
			DataToAxes:  transform.NewScaleTransform(transform.NewLinear(0, 10), transform.NewLinear(-5, 5)),
			AxesToPixel: transform.NewDisplayRectTransform(geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}}),
		},
		Clip:       geom.Rect{Min: geom.Pt{X: 50, Y: 100}, Max: geom.Pt{X: 250, Y: 300}},
		FigureRect: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 400, Y: 500}},
	}

	if got := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 2.5, Y: 0}); got != ctx.TransData().Apply(geom.Pt{X: 2.5, Y: 0}) {
		t.Fatalf("CoordData TransformFor = %+v, want %+v", got, ctx.TransData().Apply(geom.Pt{X: 2.5, Y: 0}))
	}
	if got := ctx.TransformFor(Coords(CoordAxes)).Apply(geom.Pt{X: 0.25, Y: 0.75}); got != ctx.TransAxes().Apply(geom.Pt{X: 0.25, Y: 0.75}) {
		t.Fatalf("CoordAxes TransformFor = %+v, want %+v", got, ctx.TransAxes().Apply(geom.Pt{X: 0.25, Y: 0.75}))
	}
	if got := ctx.TransformFor(Coords(CoordFigure)).Apply(geom.Pt{X: 0.25, Y: 0.75}); got != ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.75}) {
		t.Fatalf("CoordFigure TransformFor = %+v, want %+v", got, ctx.TransFigure().Apply(geom.Pt{X: 0.25, Y: 0.75}))
	}

	// Display space is y-up: CoordFigure y=0.25 maps to 0.25*500 from the bottom.
	blended := ctx.TransformFor(BlendCoords(CoordData, CoordFigure)).Apply(geom.Pt{X: 5, Y: 0.25})
	if blended.X != 150 || blended.Y != 125 {
		t.Fatalf("CoordData/Figure blended transform = %+v, want {150 125}", blended)
	}
}

func TestCoordinateTransformsRejectMixedProjectionStateWithoutSeparability(t *testing.T) {
	fig := NewFigure(400, 400)
	ax := fig.AddPolarAxes(unitRect())
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	if tr, ok := ctx.TransProjection().(transform.Separable); ok {
		t.Fatalf("expected non-separable projection transform, got %+v", tr)
	}
	if got := ctx.TransformFor(BlendCoords(CoordData, CoordAxes)); got != nil {
		t.Fatalf("non-affine projection should reject mixed CoordData/CoordAxes transforms, got %#v", got)
	}

	projectionPoint := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: math.Pi / 2, Y: 1})
	dataPoint := ctx.DataToPixel.Apply(geom.Pt{X: math.Pi / 2, Y: 1})
	if projectionPoint != dataPoint {
		t.Fatalf("CoordData transform = %+v, want %+v", projectionPoint, dataPoint)
	}
}
