package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/optarg"
	"github.com/cwbudde/matplotlib-go/render"
)

// ColorbarOptions configures figure-level colorbar placement.
type ColorbarOptions struct {
	Width       float64
	Padding     float64
	Aspect      float64
	Shrink      float64
	Anchor      *geom.Pt
	Label       string
	Colormap    *string
	VMin        *float64
	VMax        *float64
	Extend      string
	Location    string
	Orientation string
	Ticks       []float64
	Boundaries  []float64
	Values      []float64
	Spacing     string
	DrawEdges   bool
	ExtendRect  bool
	// ExtendFrac sets the extension length as a fraction of the interior body.
	// nil = matplotlib default (5%); [f] = both sides f; [min, max] = per-side.
	ExtendFrac []float64
	// ExtendFracAuto mirrors matplotlib extendfrac='auto' (extensions sized to
	// the first/last boundary interval). Takes precedence over ExtendFrac.
	ExtendFracAuto bool
	// MinorTicks opts in to colorbar minor ticks (off by default, matching
	// matplotlib's xtick/ytick.minor.visible default of False).
	MinorTicks bool
}

// Colorbar renders a vertical gradient keyed to a scalar colormap.
type Colorbar struct {
	Mapping       ScalarMapInfo
	Mappable      ScalarMappable
	Colormap      string
	Extend        string
	Orientation   string
	Boundaries    []float64
	Values        []float64
	Spacing       string
	DrawEdges     bool
	ExtendRect    bool
	ExtendFracMin float64
	ExtendFracMax float64
	Alpha         float64
	BorderColor   render.Color
	BorderWidth   float64
	z             float64
}

// AddColorbar creates a dedicated axes to the right of a plot and populates it
// with a colorbar derived from a scalar-mappable artist.
func (f *Figure) AddColorbar(parent *Axes, mappable ScalarMappable, opts ...ColorbarOptions) *Axes {
	if f == nil || parent == nil || mappable == nil {
		return nil
	}

	cfg := optarg.One("colorbar", opts)
	cfg.Aspect = resolvedColorbarAspect(cfg.Aspect)
	location := normalizeColorbarLocation(cfg.Location, cfg.Orientation)
	extend := normalizeColorbarExtend(cfg.Extend)

	mapping := mappable.ScalarMap().Resolved()
	cmapOverride := ""
	if cfg.Colormap != nil && *cfg.Colormap != "" {
		cmapOverride = *cfg.Colormap
		mapping.Colormap = cmapOverride
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
	boundaries := colorbarOptionBoundaries(cfg.Values, cfg.Boundaries)
	if len(boundaries) >= 2 {
		if cfg.VMin == nil {
			mapping.VMin = boundaries[0]
		}
		if cfg.VMax == nil {
			mapping.VMax = boundaries[len(boundaries)-1]
		}
	}
	if _, ok := mapping.Norm.(Normalize); ok && mapping.Norm != nil {
		mapping.Norm = Normalize{VMin: mapping.VMin, VMax: mapping.VMax}
	}

	base := colorbarBaseRect(parent)
	thickness := resolvedColorbarThickness(f, base, cfg.Width, cfg.Aspect, location)
	slotThickness := resolvedColorbarSlotThickness(base, cfg.Width, location)
	padding := resolvedColorbarLayoutPadding(f, base, cfg.Padding, location)
	useResolvedSlot := colorbarUsesResolvedSlot(f, parent)
	if useResolvedSlot {
		padding = resolvedColorbarPadding(base, cfg.Padding, location)
		slotThickness = resolvedColorbarSlotThickness(base, cfg.Width, location)
	}
	automin, automax := 0.05, 0.05
	if cfg.ExtendFracAuto {
		automin, automax = colorbarAutoExtendLengths(colorbarInteriorBoundaries(boundaries, extend), cfg.Spacing)
	}
	fracMin, fracMax := colorbarExtendLengths(cfg.ExtendFrac, cfg.ExtendFracAuto, automin, automax)

	parentRect, rect := colorbarPlacementRect(f, base, thickness, slotThickness, padding, location, useResolvedSlot)
	if !useResolvedSlot {
		parent.RectFraction = parentRect
	}
	rect = insetColorbarRectForExtensions(f, rect, extend, location, fracMin, fracMax)
	if rect.Min.X >= rect.Max.X {
		return nil
	}
	if rect.Min.Y >= rect.Max.Y {
		return nil
	}
	rect = applyColorbarShrinkAnchor(rect, cfg.Shrink, cfg.Anchor, location)

	ax := f.AddAxes(rect)
	ax.colorbarParent = parent
	ax.colorbarWidth = cfg.Width
	ax.colorbarPadding = cfg.Padding
	ax.colorbarAspect = cfg.Aspect
	ax.colorbarBase = base
	ax.colorbarExtend = extend
	ax.colorbarLocation = location
	ax.colorbarTicks = cloneFloat64s(cfg.Ticks)
	ax.colorbarBounds = cloneFloat64s(boundaries)
	ax.colorbarMinorTicks = cfg.MinorTicks
	ax.colorbarExtendFracMin = fracMin
	ax.colorbarExtendFracMax = fracMax
	// matplotlib hides every rectangular spine of the colorbar axes and draws
	// the border as a dedicated outline spine instead (colorbar.py:
	// `for spine in self.ax.spines.values(): spine.set_visible(False)`).
	ax.SetFrameOn(false)
	configureColorbarAxes(ax, location, cfg.Label)
	configureColorbarScale(ax, mapping, location, cfg.Ticks, boundaries, extend)

	ax.Add(&Colorbar{
		Mapping:       mapping,
		Mappable:      mappable,
		Colormap:      cmapOverride,
		Extend:        extend,
		Orientation:   colorbarOrientation(location),
		Boundaries:    cloneFloat64s(boundaries),
		Values:        cloneFloat64s(cfg.Values),
		Spacing:       normalizeColorbarSpacing(cfg.Spacing),
		DrawEdges:     cfg.DrawEdges,
		ExtendRect:    cfg.ExtendRect,
		ExtendFracMin: fracMin,
		ExtendFracMax: fracMax,
		Alpha:         1,
		BorderColor:   f.RC.AxesEdgeColor,
		BorderWidth:   f.RC.AxisLineWidth,
		z:             -10,
	})

	return ax
}

// Bounds returns an empty rect so colorbars do not affect autoscaling.
func (c *Colorbar) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the colorbar z-order.
func (c *Colorbar) Z() float64 { return c.z }
