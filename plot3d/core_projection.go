package plot3d

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func init() {
	if err := core.RegisterProjection("3d", func() core.Projection { return newAxes3DProjection() }); err != nil {
		panic(err)
	}
	if err := core.RegisterProjection("axes3d", func() core.Projection { return newAxes3DProjection() }); err != nil {
		panic(err)
	}
}

// AddAxes appends a 3D axes to fig.
func AddAxes(fig *core.Figure, rect geom.Rect, opts ...style.Option) (*Axes3D, error) {
	ax, err := fig.AddAxesProjection(rect, "3d", opts...)
	if err != nil {
		return nil, err
	}
	return NewAxes(ax), nil
}

type axes3DProjection struct {
	name string
}

func newAxes3DProjection() *axes3DProjection {
	return &axes3DProjection{name: "3d"}
}

func (p *axes3DProjection) Name() string {
	if p == nil || p.name == "" {
		return "3d"
	}
	return p.name
}

func (*axes3DProjection) OwnsSpines() bool { return true }

func (p *axes3DProjection) CloneProjection() core.Projection {
	if p == nil {
		return newAxes3DProjection()
	}
	clone := *p
	return &clone
}

func (p *axes3DProjection) ConfigureAxes(ax *core.Axes) {
	if ax == nil {
		return
	}
	ax.XScale = transform.NewLinear(default3DViewMin, default3DViewMax)
	ax.YScale = transform.NewLinear(default3DViewMin, default3DViewMax)
	_ = ax.SetBoxAspect(1)
	ax.XAxis = core.NewXAxis()
	ax.YAxis = core.NewYAxis()
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.XAxisTop = nil
	ax.YAxisRight = nil
	ax.ShowFrame = false
}

func (p *axes3DProjection) DataToAxes(*core.Axes) transform.T {
	return nil
}
