package core

import (
	"fmt"
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/transform"
)

const defaultSkewXAngleDeg = 30.0

type skewXProjection struct {
	angleDeg float64
}

func newSkewXProjection() *skewXProjection {
	return &skewXProjection{angleDeg: defaultSkewXAngleDeg}
}

func skewXProjectionForAxes(ax *Axes) (*skewXProjection, bool) {
	if ax == nil {
		return nil, false
	}
	proj, ok := ax.projection.(*skewXProjection)
	return proj, ok && proj != nil
}

func (p *skewXProjection) Name() string {
	return "skewx"
}

func (p *skewXProjection) CloneProjection() Projection {
	if p == nil {
		return newSkewXProjection()
	}
	clone := *p
	return &clone
}

func (p *skewXProjection) ConfigureAxes(ax *Axes) {
	if ax == nil {
		return
	}

	ax.XScale = transform.NewLinear(-40, 50)
	ax.YScale = transform.NewLog(1050, 100, 10)
	ax.XAxis = NewXAxis()
	ax.YAxis = NewYAxis()
	ax.XAxisTop = cloneAxisForSide(ax.XAxis, AxisTop)
	ax.YAxisRight = nil
	ax.ShowFrame = true

	ax.XAxis.Locator = MultipleLocator{Base: 10}
	ax.XAxis.MinorLocator = MultipleLocator{Base: 5}
	ax.XAxis.Formatter = ScalarFormatter{Prec: 0}
	ax.XAxisTop.Locator = ax.XAxis.Locator
	ax.XAxisTop.MinorLocator = ax.XAxis.MinorLocator
	ax.XAxisTop.Formatter = ax.XAxis.Formatter
	ax.XAxisTop.ShowTicks = false
	ax.XAxisTop.ShowLabels = false

	pressureTicks := []float64{100, 200, 300, 500, 700, 850, 1000}
	ax.YAxis.Locator = FixedLocator{TicksList: pressureTicks}
	ax.YAxis.MinorLocator = LogLocator{Base: 10, Minor: true, Subs: []float64{2, 3, 4, 5, 6, 7, 8, 9}}
	ax.YAxis.Formatter = FuncFormatter(func(v float64) string {
		return fmt.Sprintf("%.0f", v)
	})
}

func (p *skewXProjection) DataToAxes(ax *Axes) transform.T {
	if ax == nil {
		return nil
	}
	angle := defaultSkewXAngleDeg
	if p != nil {
		angle = p.angleDeg
	}
	return skewXDataTransform{
		x:      ax.effectiveXScale(),
		y:      ax.effectiveYScale(),
		factor: math.Tan(angle * math.Pi / 180),
	}
}

type skewXDataTransform struct {
	x      transform.Scale
	y      transform.Scale
	factor float64
}

func (t skewXDataTransform) Apply(p geom.Pt) geom.Pt {
	u := p.X
	if t.x != nil {
		u = t.x.Fwd(p.X)
	}
	v := p.Y
	if t.y != nil {
		v = t.y.Fwd(p.Y)
	}
	return geom.Pt{
		X: u + t.factor*v,
		Y: v,
	}
}

func (t skewXDataTransform) Invert(p geom.Pt) (geom.Pt, bool) {
	v := p.Y
	y := v
	if t.y != nil {
		var ok bool
		y, ok = t.y.Inv(v)
		if !ok {
			return geom.Pt{}, false
		}
	}

	u := p.X - t.factor*v
	x := u
	if t.x != nil {
		var ok bool
		x, ok = t.x.Inv(u)
		if !ok {
			return geom.Pt{}, false
		}
	}

	return geom.Pt{X: x, Y: y}, true
}

func skewXGridXTickDomain(ctx *DrawContext) (float64, float64, bool) {
	if ctx == nil || ctx.Axes == nil || ctx.DataToPixel.XScale == nil {
		return 0, 0, false
	}
	proj, ok := skewXProjectionForAxes(ctx.Axes)
	if !ok {
		return 0, 0, false
	}
	angle := defaultSkewXAngleDeg
	if proj != nil {
		angle = proj.angleDeg
	}
	factor := math.Tan(angle * math.Pi / 180)
	topLeft, ok := ctx.DataToPixel.XScale.Inv(-factor)
	if !ok {
		return 0, 0, false
	}
	topRight, ok := ctx.DataToPixel.XScale.Inv(1 - factor)
	if !ok {
		return 0, 0, false
	}
	lowerLeft, lowerRight := ctx.DataToPixel.XScale.Domain()
	minVal := math.Min(math.Min(lowerLeft, lowerRight), math.Min(topLeft, topRight))
	maxVal := math.Max(math.Max(lowerLeft, lowerRight), math.Max(topLeft, topRight))
	return minVal, maxVal, true
}

func skewYAxisDisplayPoint(axis *Axis, ctx *DrawContext, tickValue float64) (geom.Pt, bool) {
	if axis == nil || ctx == nil || ctx.Axes == nil || ctx.DataToPixel.YScale == nil {
		return geom.Pt{}, false
	}
	if axis.Side != AxisLeft && axis.Side != AxisRight {
		return geom.Pt{}, false
	}
	if _, ok := skewXProjectionForAxes(ctx.Axes); !ok {
		return geom.Pt{}, false
	}
	transAxes := ctx.TransAxes()
	if transAxes == nil {
		return geom.Pt{}, false
	}
	y := ctx.DataToPixel.YScale.Fwd(tickValue)
	if math.IsNaN(y) || math.IsInf(y, 0) {
		return geom.Pt{}, false
	}
	x := 0.0
	if axis.Side == AxisRight {
		x = 1
	}
	return transAxes.Apply(geom.Pt{X: x, Y: y}), true
}
