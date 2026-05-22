package core

import (
	"math"

	matcolor "github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ColorbarOptions configures figure-level colorbar placement.
type ColorbarOptions struct {
	Width    float64
	Padding  float64
	Aspect   float64
	Label    string
	Colormap *string
	VMin     *float64
	VMax     *float64
	Extend   string
}

// Colorbar renders a vertical gradient keyed to a scalar colormap.
type Colorbar struct {
	Mapping     ScalarMapInfo
	Colormap    string
	Extend      string
	Alpha       float64
	BorderColor render.Color
	BorderWidth float64
	z           float64
}

const (
	defaultColorbarFraction = 0.15
	defaultColorbarPadding  = 0.05
	defaultColorbarAspect   = 20.0
)

// AddColorbar creates a dedicated axes to the right of a plot and populates it
// with a colorbar derived from a scalar-mappable artist.
func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable, opts ...ColorbarOptions) *Axes {
	if f == nil || parent == nil || mappable == nil {
		return nil
	}

	cfg := ColorbarOptions{}
	if len(opts) > 0 {
		cfg = opts[0]
	}
	cfg.Aspect = resolvedColorbarAspect(cfg.Aspect)

	mapping := mappable.ScalarMap().Resolved()
	cmap := mapping.Colormap
	if cfg.Colormap != nil && *cfg.Colormap != "" {
		cmap = *cfg.Colormap
		mapping.Colormap = cmap
	}
	vmin := mapping.VMin
	if cfg.VMin != nil {
		vmin = *cfg.VMin
	}
	vmax := mapping.VMax
	if cfg.VMax != nil {
		vmax = *cfg.VMax
	}
	if vmin == vmax {
		vmax = vmin + 1
	}
	mapping.VMin = vmin
	mapping.VMax = vmax
	if _, ok := mapping.Norm.(Normalize); ok && mapping.Norm != nil {
		mapping.Norm = Normalize{VMin: vmin, VMax: vmax}
	}

	base := colorbarBaseRect(parent)
	width := resolvedColorbarWidth(f, base, cfg.Width, cfg.Aspect)
	padding := resolvedColorbarLayoutPadding(f, base, cfg.Padding)
	useResolvedSlot := colorbarUsesResolvedSlot(f, parent)
	parent.RectFraction = colorbarParentRect(base, width, padding, useResolvedSlot)
	slotLeft := colorbarSlotLeft(base, width, useResolvedSlot)
	rect := geom.Rect{
		Min: geom.Pt{
			X: slotLeft,
			Y: parent.RectFraction.Min.Y,
		},
		Max: geom.Pt{
			X: slotLeft + width,
			Y: parent.RectFraction.Max.Y,
		},
	}
	if rect.Min.X >= rect.Max.X {
		return nil
	}

	ax := f.AddAxes(rect)
	ax.colorbarParent = parent
	ax.colorbarWidth = cfg.Width
	ax.colorbarPadding = cfg.Padding
	ax.colorbarAspect = cfg.Aspect
	ax.colorbarBase = base
	ax.ShowFrame = false
	ax.SetXLim(0, 1)
	configureColorbarScale(ax, mapping)

	if ax.XAxis != nil {
		ax.XAxis.ShowSpine = false
		ax.XAxis.ShowTicks = false
		ax.XAxis.ShowLabels = false
	}
	if ax.YAxis != nil {
		ax.YAxis.ShowSpine = false
		ax.YAxis.ShowTicks = false
		ax.YAxis.ShowLabels = false
		ax.YAxis.MinorLocator = nil
	}
	if right := ax.RightAxis(); right != nil {
		right.MinorLocator = nil
	}
	_ = ax.SetYTickLabelPosition("right")
	_ = ax.SetYLabelPosition("right")
	if cfg.Label != "" {
		ax.SetYLabel(cfg.Label)
	}

	ax.Add(&Colorbar{
		Mapping:     mapping,
		Colormap:    cmap,
		Extend:      normalizeColorbarExtend(cfg.Extend),
		Alpha:       1,
		BorderColor: f.RC.AxesEdgeColor,
		BorderWidth: f.RC.AxisLineWidth,
		z:           -10,
	})

	return ax
}

func resolvedColorbarPadding(base geom.Rect, padding float64) float64 {
	if padding > 0 {
		return base.W() * padding
	}
	return base.W() * defaultColorbarPadding
}

func resolvedColorbarLayoutPadding(fig *Figure, base geom.Rect, padding float64) float64 {
	resolved := resolvedColorbarPadding(base, padding)
	if padding > 0 || fig == nil || fig.layoutEngine != LayoutEngineConstrained || fig.SizePx.X <= 0 {
		return resolved
	}
	return resolved + layoutPadPx(fig, LayoutEngineConstrained)/fig.SizePx.X
}

func resolvedColorbarAspect(aspect float64) float64 {
	if aspect > 0 {
		return aspect
	}
	return defaultColorbarAspect
}

func resolvedColorbarWidth(fig *Figure, base geom.Rect, width, aspect float64) float64 {
	if width > 0 {
		return width
	}
	fractionWidth := base.W() * defaultColorbarFraction
	if fig == nil || fig.SizePx.X <= 0 || fig.SizePx.Y <= 0 || aspect <= 0 {
		return fractionWidth
	}
	aspectWidth := base.H() * fig.SizePx.Y / (aspect * fig.SizePx.X)
	if aspectWidth <= 0 {
		return fractionWidth
	}
	return math.Min(fractionWidth, aspectWidth)
}

func colorbarParentRect(base geom.Rect, width, padding float64, useResolvedSlot bool) geom.Rect {
	if padding < 0 {
		return base
	}
	shrunk := base
	right := colorbarSlotLeft(base, width, useResolvedSlot) - padding
	if right <= base.Min.X {
		return shrunk
	}
	shrunk.Max.X = right
	return shrunk
}

func colorbarSlotLeft(base geom.Rect, width float64, useResolvedSlot bool) float64 {
	if !useResolvedSlot {
		return base.Max.X - base.W()*defaultColorbarFraction
	}
	return base.Max.X - width
}

func colorbarUsesResolvedSlot(fig *Figure, parent *Axes) bool {
	return fig != nil && fig.layoutEngine == LayoutEngineConstrained && parent != nil && parent.subplotSpec != nil
}

func colorbarBaseRect(parent *Axes) geom.Rect {
	if parent == nil {
		return geom.Rect{}
	}
	if parent.subplotSpec != nil {
		return parent.subplotSpec.Rect()
	}
	return parent.RectFraction
}

func configureColorbarScale(ax *Axes, mapping ScalarMapInfo) {
	if ax == nil {
		return
	}
	vmin, vmax := mapping.VMin, mapping.VMax
	switch norm := mapping.Norm.(type) {
	case LogNorm:
		base := 10.0
		if vmin > 0 && vmax > 0 {
			ax.SetYLimLog(vmin, vmax, base)
		} else {
			ax.SetYLim(vmin, vmax)
		}
		if right := ax.RightAxis(); right != nil {
			right.Locator = LogLocator{Base: base}
			right.Formatter = LogFormatter{Base: base}
		}
	case BoundaryNorm:
		ax.SetYLim(vmin, vmax)
		if right := ax.RightAxis(); right != nil {
			right.Locator = FixedLocator{TicksList: append([]float64(nil), norm.Boundaries...)}
			right.Formatter = ScalarFormatter{Prec: 6}
		}
	default:
		ax.SetYLim(vmin, vmax)
	}
}

func normalizeColorbarExtend(extend string) string {
	switch extend {
	case "min", "max", "both":
		return extend
	default:
		return "neither"
	}
}

type colorbarExtensionPath struct {
	Path      geom.Path
	OverRange bool
}

func colorbarExtensionPaths(clip geom.Rect, extend string) []colorbarExtensionPath {
	extend = normalizeColorbarExtend(extend)
	if extend == "neither" || clip.W() <= 0 || clip.H() <= 0 {
		return nil
	}
	height := math.Min(clip.H()*0.05, clip.W()*0.5)
	out := make([]colorbarExtensionPath, 0, 2)
	if extend == "min" || extend == "both" {
		out = append(out, colorbarExtensionPath{
			OverRange: false,
			Path: geom.Path{
				V: []geom.Pt{
					{X: clip.Min.X, Y: clip.Max.Y},
					{X: clip.Max.X, Y: clip.Max.Y},
					{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Max.Y + height},
				},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
			},
		})
	}
	if extend == "max" || extend == "both" {
		out = append(out, colorbarExtensionPath{
			OverRange: true,
			Path: geom.Path{
				V: []geom.Pt{
					{X: clip.Min.X, Y: clip.Min.Y},
					{X: (clip.Min.X + clip.Max.X) * 0.5, Y: clip.Min.Y - height},
					{X: clip.Max.X, Y: clip.Min.Y},
				},
				C: []geom.Cmd{geom.MoveTo, geom.LineTo, geom.LineTo, geom.ClosePath},
			},
		})
	}
	return out
}

// Draw renders a vertical gradient across the colorbar axes.
func (c *Colorbar) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil {
		return
	}

	const gradientHeight = 256

	mapping := c.Mapping.Resolved()
	if c.Colormap != "" {
		mapping.Colormap = c.Colormap
	}
	cmap := matcolor.GetColormap(mapping.Colormap)
	alpha := c.Alpha
	if alpha <= 0 {
		alpha = 1
	}

	outlinePath := pixelRectPath(ctx.Clip)
	if snapped := snappedStrokeRectPath(ctx.Clip); len(snapped.C) > 0 {
		outlinePath = snapped
	}

	for _, ext := range colorbarExtensionPaths(ctx.Clip, c.Extend) {
		t := -1.0
		if ext.OverRange {
			t = 2
		}
		col := cmap.AtValue(t)
		col.A *= alpha
		r.Path(ext.Path, &render.Paint{
			Fill:      col,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
			Antialias: render.AntialiasDefault,
		})
	}

	if norm, ok := mapping.Norm.(BoundaryNorm); ok && len(norm.Boundaries) >= 2 {
		vmin, vmax := mapping.VMin, mapping.VMax
		span := vmax - vmin
		if span != 0 {
			for i := 0; i+1 < len(norm.Boundaries); i++ {
				low := norm.Boundaries[i]
				high := norm.Boundaries[i+1]
				y0 := ctx.Clip.Max.Y - ctx.Clip.H()*((high-vmin)/span)
				y1 := ctx.Clip.Max.Y - ctx.Clip.H()*((low-vmin)/span)
				path := snappedFillRectPath(geom.Rect{
					Min: geom.Pt{X: ctx.Clip.Min.X, Y: y0},
					Max: geom.Pt{X: ctx.Clip.Max.X, Y: y1},
				})
				if len(path.C) == 0 {
					continue
				}
				col := mapping.Color((low+high)*0.5, alpha)
				r.Path(path, &render.Paint{
					Fill:      col,
					LineJoin:  render.JoinMiter,
					LineCap:   render.CapButt,
					Antialias: render.AntialiasDefault,
				})
			}
		}
	} else {
		for i := 0; i < gradientHeight; i++ {
			t := (float64(i) + 0.5) / float64(gradientHeight)
			col := cmap.AtValue(t)
			col.A *= alpha

			path := snappedFillRectPath(colorbarCellRect(ctx.Clip, i, gradientHeight))
			if len(path.C) == 0 {
				continue
			}
			r.Path(path, &render.Paint{
				Fill:      col,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Antialias: render.AntialiasDefault,
			})
		}
	}

	r.Path(outlinePath, &render.Paint{
		Stroke:    c.BorderColor,
		LineWidth: c.BorderWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
	})
}

func colorbarCellRect(clip geom.Rect, index, count int) geom.Rect {
	if count <= 0 {
		return geom.Rect{}
	}
	y0 := clip.Max.Y - clip.H()*float64(index+1)/float64(count)
	y1 := clip.Max.Y - clip.H()*float64(index)/float64(count)
	return geom.Rect{
		Min: geom.Pt{X: clip.Min.X, Y: y0},
		Max: geom.Pt{X: clip.Max.X, Y: y1},
	}
}

// Bounds returns an empty rect so colorbars do not affect autoscaling.
func (c *Colorbar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the colorbar z-order.
func (c *Colorbar) Z() float64 { return c.z }
