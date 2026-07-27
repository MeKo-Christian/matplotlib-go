package core

import (
	"fmt"
	"math"

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
		SizePx:       figureSizePx(w, h, rc.DPI),
		RC:           rc,
		Children:     nil,
		Artists:      nil,
		layoutEngine: engine,
	}
}

// figureSizePx converts a pixel figure size into the display extent matplotlib
// would compute for it.
//
// Matplotlib's figure size is authoritative in *inches*; the pixel extent is
// derived as figsize*dpi and carries that product's float64 noise. A figure
// asked for in pixels is `figsize=(px/dpi, px/dpi)`, and the round trip is not
// the identity: 980/100*100 == 980.0000000000001, not 980.
//
// Every display coordinate inherits it — Axes.layout multiplies SizePx by the
// axes fraction — and the difference decides ties. In a 980-wide figure the
// x=0 tick label's centred origin lands on exactly 94.5 with a width of 980 and
// on 94.50000000000001 with matplotlib's, and the AGG text blit rounds
// ties-to-even: 94 against 95, one pixel apart. Casting float64(w) directly
// drew that label a pixel left of the reference.
//
// Only sizes that are not exact multiples of the dpi are affected; 640, 360 and
// 620 all round-trip exactly and are unchanged.
func figureSizePx(w, h int, dpi float64) geom.Pt {
	if dpi <= 0 {
		return geom.Pt{X: float64(w), Y: float64(h)}
	}
	return geom.Pt{X: float64(w) / dpi * dpi, Y: float64(h) / dpi * dpi}
}

// CanvasSize reports the integer pixel canvas the figure should be rasterized
// into.
//
// SizePx is the display extent matplotlib computes (see figureSizePx), so it can
// sit a fraction of an ULP either side of the requested integer — 402 px at
// 100 dpi is 401.99999999999994. Truncating that allocates a canvas one pixel
// short, so round instead. Matplotlib does truncate here (a figsize=(4.02, 2)
// figure really does save a 401-px PNG), but no fixture depends on that and
// reproducing it would silently change committed image dimensions.
func (f *Figure) CanvasSize() (width, height int) {
	if f == nil {
		return 0, 0
	}
	return int(math.Round(f.SizePx.X)), int(math.Round(f.SizePx.Y))
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
