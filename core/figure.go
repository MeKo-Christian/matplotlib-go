package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

type Figure struct {
	SizePx    geom.Pt
	RC        style.RC
	Children  []*Axes
	Artists   []Artist
	zsorted   bool
	SupTitle  string
	SupXLabel string
	SupYLabel string

	layoutEngine LayoutEngine
}

func NewFigure(w, h int, opts ...style.Option) *Figure {
	rc := style.Apply(style.CurrentDefaults(), opts...)
	return &Figure{
		SizePx:       geom.Pt{X: float64(w), Y: float64(h)},
		RC:           rc,
		Children:     nil,
		Artists:      nil,
		SupTitle:     "",
		SupXLabel:    "",
		SupYLabel:    "",
		layoutEngine: LayoutEngineNone,
	}
}

func (f *Figure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
	proj, _ := lookupProjection("rectilinear")
	return f.addAxesWithProjection(r, proj, opts...)
}

func (f *Figure) AddAxesProjection(r geom.Rect, projection string, opts ...style.Option) (*Axes, error) {
	proj, err := lookupProjection(projection)
	if err != nil {
		return nil, err
	}
	return f.addAxesWithProjection(r, proj, opts...), nil
}

func (f *Figure) AddPolarAxes(r geom.Rect, opts ...style.Option) *Axes {
	ax, err := f.AddAxesProjection(r, "polar", opts...)
	if err != nil {
		return nil
	}
	return ax
}

func (f *Figure) AddRadarAxes(r geom.Rect, labels []string, opts ...style.Option) (*Axes, error) {
	if len(labels) > 0 && len(labels) < 3 {
		return nil, fmt.Errorf("radar axes require at least 3 labels")
	}
	ax, err := f.AddAxesProjection(r, "radar", opts...)
	if err != nil {
		return nil, err
	}
	if len(labels) > 0 {
		if err := ax.SetRadarLabels(labels); err != nil {
			return nil, err
		}
	}
	return ax, nil
}

func (f *Figure) AddSkewXAxes(r geom.Rect, opts ...style.Option) (*Axes, error) {
	return f.AddAxesProjection(r, "skewx", opts...)
}

func (f *Figure) AddAxes3D(r geom.Rect, opts ...style.Option) (*Axes3D, error) {
	ax, err := f.AddAxesProjection(r, "3d", opts...)
	if err != nil {
		return nil, err
	}
	return NewAxes3D(ax), nil
}

func (f *Figure) addAxesWithProjection(r geom.Rect, proj Projection, opts ...style.Option) *Axes {
	var rc *style.RC
	effective := f.RC
	if len(opts) > 0 {
		v := style.Apply(f.RC, opts...)
		rc = &v
		effective = v
	}
	ax := &Axes{
		RectFraction:    r,
		RC:              rc,
		XScale:          transform.NewLinear(0, 1),
		YScale:          transform.NewLinear(0, 1),
		projection:      cloneProjection(proj),
		XAxis:           NewXAxis(),
		YAxis:           NewYAxis(),
		ShowFrame:       true,
		ColorCycle:      color.NewColorCycle(effective.Palette()),
		PatchColorCycle: color.NewColorCycle(effective.Palette()),
		aspectMode:      "auto",
		aspectValue:     1,
		xLabelSide:      AxisBottom,
		yLabelSide:      AxisLeft,
		figure:          f,
	}
	if ax.projection == nil {
		ax.projection, _ = lookupProjection("rectilinear")
	}
	ax.projection.ConfigureAxes(ax)
	ax.applyStyleDefaults(effective)
	ax.addDefaultGrids(effective)
	f.Children = append(f.Children, ax)
	return ax
}

func (f *Figure) Add(art Artist) {
	if f == nil {
		return
	}
	f.Artists = append(f.Artists, art)
	f.zsorted = false
}
