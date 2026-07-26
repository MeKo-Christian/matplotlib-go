package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

type Figure struct {
	SizePx    geom.Pt
	RC        style.RC
	Children  []*Axes
	Artists   []Artist
	zsorted   bool
	supTitle  string
	supXLabel string
	supYLabel string

	layoutEngine LayoutEngine
}

func NewFigure(w, h int, opts ...style.Option) *Figure {
	rc := style.Apply(style.CurrentDefaults(), opts...)
	engine := LayoutEngineNone
	if rc.Figure.AutoLayout {
		engine = LayoutEngineTight
	} else if rc.Figure.Constrained.Use {
		engine = LayoutEngineConstrained
	}
	return &Figure{
		SizePx:       geom.Pt{X: float64(w), Y: float64(h)},
		RC:           rc,
		Children:     nil,
		Artists:      nil,
		layoutEngine: engine,
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

func (f *Figure) addAxesWithProjection(r geom.Rect, proj Projection, opts ...style.Option) *Axes {
	var rc *style.RC
	if len(opts) > 0 {
		v := style.Apply(f.RC, opts...)
		rc = &v
	}
	ax := &Axes{
		RectFraction: r,
		RC:           rc,
		projection:   cloneProjection(proj),
		figure:       f,
	}
	ax.resetToDefaults()
	f.Children = append(f.Children, ax)
	return ax
}

// DelAxes removes ax from the figure, mirroring Matplotlib's
// Figure.delaxes/_remove_axes. The axes is dropped from the figure's child
// list and from any parent axes' inset list, its back-reference is detached,
// and share links pointing at it from the remaining axes are broken.
//
// Divergence from Matplotlib: rather than transferring locators/formatters to a
// surviving shared sibling, the broken share links are simply cleared.
func (f *Figure) DelAxes(ax *Axes) {
	if f == nil || ax == nil {
		return
	}
	f.Children = removeAxes(f.Children, ax)
	for _, child := range f.Children {
		if child == nil {
			continue
		}
		if child.shareX == ax {
			child.shareX = nil
		}
		if child.shareY == ax {
			child.shareY = nil
		}
		child.childAxes = removeAxes(child.childAxes, ax)
	}
	ax.figure = nil
	f.zsorted = false
}

// removeAxes returns axes with the first occurrence of target removed,
// preserving order.
func removeAxes(axes []*Axes, target *Axes) []*Axes {
	for i, candidate := range axes {
		if candidate == target {
			return append(axes[:i:i], axes[i+1:]...)
		}
	}
	return axes
}

// Clear clears the figure: every child axes is cleared and removed, all
// figure-level artists are dropped, and the figure super-labels are reset.
// Mirrors Matplotlib's Figure.clear.
func (f *Figure) Clear() {
	if f == nil {
		return
	}
	for _, ax := range append([]*Axes(nil), f.Children...) {
		ax.Clear()
		f.DelAxes(ax)
	}
	f.Artists = nil
	f.zsorted = false
	f.supTitle = ""
	f.supXLabel = ""
	f.supYLabel = ""
}

// Clf is a synonym for Clear, matching Matplotlib's Figure.clf.
func (f *Figure) Clf() { f.Clear() }

func (f *Figure) Add(art Artist) {
	if f == nil {
		return
	}
	f.Artists = append(f.Artists, art)
	f.zsorted = false
}
