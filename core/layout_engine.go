package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// LayoutEngine controls draw-time subplot reflow.
type LayoutEngine uint8

const (
	LayoutEngineNone LayoutEngine = iota
	LayoutEngineTight
	LayoutEngineConstrained
)

// SubplotAdjust mirrors Matplotlib's subplots_adjust surface for managed grids.
type SubplotAdjust struct {
	Left   *float64
	Right  *float64
	Bottom *float64
	Top    *float64
	WSpace *float64
	HSpace *float64
}

type figureMargin struct {
	left   float64
	right  float64
	top    float64
	bottom float64
}

type axesDecorationPadding struct {
	left   float64
	right  float64
	top    float64
	bottom float64
}

const (
	// Matplotlib rc defaults:
	// figure.constrained_layout.{h_pad,w_pad}=0.04167 inch and
	// figure.constrained_layout.{hspace,wspace}=0.02.
	// See third_party/matplotlib/lib/matplotlib/mpl-data/matplotlibrc.
	matplotlibConstrainedLayoutPadPoints = 0.04167 * 72.0
	matplotlibConstrainedLayoutSpace     = 0.02

	// Matplotlib Figure.tight_layout defaults to pad=1.08 font-size units.
	// See third_party/matplotlib/lib/matplotlib/figure.py:tight_layout.
	matplotlibTightLayoutPadFontSize = 1.08
)

// SetLayoutEngine selects the draw-time subplot layout engine.
func (f *Figure) SetLayoutEngine(engine LayoutEngine) {
	if f == nil {
		return
	}
	f.layoutEngine = engine
}

// TightLayout enables a measured layout pass with compact padding.
func (f *Figure) TightLayout() {
	f.SetLayoutEngine(LayoutEngineTight)
}

// ConstrainedLayout enables a measured layout pass with roomier padding.
func (f *Figure) ConstrainedLayout() {
	f.SetLayoutEngine(LayoutEngineConstrained)
}

// LayoutEngine reports the active draw-time layout engine.
func (f *Figure) LayoutEngine() LayoutEngine {
	if f == nil {
		return LayoutEngineNone
	}
	return f.layoutEngine
}

// SubplotsAdjust applies persistent subplot parameter changes to managed grids.
func (f *Figure) SubplotsAdjust(cfg SubplotAdjust) {
	if f == nil {
		return
	}
	for _, grid := range managedRootGrids(f) {
		if cfg.Left != nil {
			grid.options.Left = *cfg.Left
		}
		if cfg.Right != nil {
			grid.options.Right = *cfg.Right
		}
		if cfg.Bottom != nil {
			grid.options.Bottom = *cfg.Bottom
		}
		if cfg.Top != nil {
			grid.options.Top = *cfg.Top
		}
		if cfg.WSpace != nil {
			grid.options.WSpace = *cfg.WSpace
		}
		if cfg.HSpace != nil {
			grid.options.HSpace = *cfg.HSpace
		}
	}
	syncAxesToSubplotSpecs(f, nil)
}

func prepareFigureLayout(fig *Figure, r render.Renderer, vp geom.Rect) {
	if fig == nil {
		return
	}
	syncAxesToSubplotSpecs(fig, nil)
	if fig.layoutEngine == LayoutEngineNone {
		return
	}

	gridAxes := axesByManagedGrid(fig)
	if len(gridAxes) == 0 {
		return
	}

	children := childGrids(gridAxes)
	state := map[*GridSpec]GridSpecOptions{}

	for iter := 0; iter < 2; iter++ {
		syncAxesToSubplotSpecs(fig, state)
		syncColorbarAxes(fig)
		for _, root := range managedRootGrids(fig) {
			resolveMeasuredGridLayout(fig, r, vp, root, gridAxes, children, state, iter)
		}
	}
	syncAxesToSubplotSpecs(fig, state)
	syncColorbarAxes(fig)
}

func resolveMeasuredGridLayout(fig *Figure, r render.Renderer, vp geom.Rect, grid *GridSpec, gridAxes map[*GridSpec][]*Axes, children map[*GridSpec][]*GridSpec, state map[*GridSpec]GridSpecOptions, layoutPass int) {
	if grid == nil {
		return
	}

	syncAxesToSubplotSpecs(fig, state)
	alignment := computeFigureTextAlignment(fig, r, vp)
	state[grid] = measuredGridOptions(fig, r, vp, grid, gridAxes[grid], state, alignment, layoutPass)
	syncAxesToSubplotSpecs(fig, state)

	for _, child := range children[grid] {
		resolveMeasuredGridLayout(fig, r, vp, child, gridAxes, children, state, layoutPass)
	}
}

func measuredGridOptions(fig *Figure, r render.Renderer, vp geom.Rect, grid *GridSpec, axes []*Axes, state map[*GridSpec]GridSpecOptions, alignment figureTextAlignment, layoutPass int) GridSpecOptions {
	if grid == nil {
		return GridSpecOptions{}
	}
	opts := grid.options

	parentRect := grid.parentRectForState(state)
	parentPx := fractionRectToPixels(parentRect, vp)
	if parentPx.W() <= 0 || parentPx.H() <= 0 {
		return opts
	}

	leftMargins := make([]float64, grid.nCols)
	rightMargins := make([]float64, grid.nCols)
	topMargins := make([]float64, grid.nRows)
	bottomMargins := make([]float64, grid.nRows)

	for _, ax := range axes {
		if ax == nil || ax.subplotSpec == nil {
			continue
		}
		padding := measureAxesDecorationPadding(ax, fig, r, vp, alignment)
		leftMargins[ax.subplotSpec.colStart] = math.Max(leftMargins[ax.subplotSpec.colStart], padding.left)
		rightMargins[ax.subplotSpec.colEnd-1] = math.Max(rightMargins[ax.subplotSpec.colEnd-1], padding.right)
		topMargins[ax.subplotSpec.rowStart] = math.Max(topMargins[ax.subplotSpec.rowStart], padding.top)
		bottomMargins[ax.subplotSpec.rowEnd-1] = math.Max(bottomMargins[ax.subplotSpec.rowEnd-1], padding.bottom)
	}

	outerPadX := layoutPadPx(fig, fig.layoutEngine)
	outerPadY := outerPadX
	if fig.layoutEngine == LayoutEngineConstrained {
		outerPadY = constrainedLayoutPadPx(fig)
	}
	innerPadX := outerPadX
	innerPadY := outerPadY
	if fig.layoutEngine == LayoutEngineConstrained {
		innerPadX = math.Max(constrainedLayoutDefaultSpacePx(parentPx.W(), grid.nCols), 2*outerPadX)
		innerPadY = math.Max(constrainedLayoutDefaultSpacePx(parentPx.H(), grid.nRows), 2*outerPadY)
	}
	global := figureLayoutMarginsPx(fig, r, vp, fig.layoutEngine, layoutPass)
	if !gridCoversWholeFigure(grid) {
		global = figureMargin{}
	}

	leftPx := leftMargins[0] + outerPadX + global.left
	rightPx := rightMargins[len(rightMargins)-1] + outerPadX + global.right
	topPx := topMargins[0] + outerPadY + global.top
	bottomPx := bottomMargins[len(bottomMargins)-1] + outerPadY + global.bottom

	opts.Left = clamp01(leftPx / parentPx.W())
	opts.Right = clamp01(1 - rightPx/parentPx.W())
	opts.Bottom = clamp01(bottomPx / parentPx.H())
	opts.Top = clamp01(1 - topPx/parentPx.H())

	if opts.Right <= opts.Left || opts.Top <= opts.Bottom {
		return grid.options
	}

	innerW := parentPx.W() * (opts.Right - opts.Left)
	innerH := parentPx.H() * (opts.Top - opts.Bottom)
	if innerW <= 0 || innerH <= 0 {
		return grid.options
	}

	maxGapX := 0.0
	for col := 0; col < len(leftMargins)-1; col++ {
		maxGapX = math.Max(maxGapX, rightMargins[col]+leftMargins[col+1]+innerPadX)
	}
	maxGapY := 0.0
	for row := 0; row < len(topMargins)-1; row++ {
		maxGapY = math.Max(maxGapY, bottomMargins[row]+topMargins[row+1]+innerPadY)
	}

	if grid.nCols > 1 {
		gap := capLayoutGap(maxGapX, innerW, grid.nCols)
		opts.WSpace = gap / innerW
	} else {
		opts.WSpace = 0
	}
	if grid.nRows > 1 {
		gap := capLayoutGap(maxGapY, innerH, grid.nRows)
		opts.HSpace = gap / innerH
	} else {
		opts.HSpace = 0
	}

	return opts
}

func measureAxesDecorationPadding(ax *Axes, fig *Figure, r render.Renderer, vp geom.Rect, alignment figureTextAlignment) axesDecorationPadding {
	px := ax.adjustedLayout(fig)
	ctx := newAxesDrawContext(ax, fig, vp, px)
	union := px

	for _, axis := range []*Axis{ax.effectiveXAxis(), ax.effectiveYAxis(), ax.effectiveTopAxis(), ax.effectiveRightAxis()} {
		if bounds, ok := axisTickLabelBounds(axis, r, ctx); ok {
			union = unionRect(union, bounds)
		}
	}

	if bounds, ok := titleBounds(ax, r, ctx, px, alignment); ok {
		union = unionRect(union, bounds)
	}
	if bounds, ok := xLabelBounds(ax, r, ctx, px, alignment); ok {
		union = unionRect(union, bounds)
	}
	if bounds, ok := yLabelBounds(ax, r, ctx, px, alignment); ok {
		union = unionRect(union, bounds)
	}

	// Display space is y-up: the visual top edge is px.Max.Y and the bottom edge
	// is px.Min.Y. Decorations above the axes (title, top ticks) extend past
	// px.Max.Y; decorations below (x-label, bottom ticks) extend past px.Min.Y.
	return axesDecorationPadding{
		left:   math.Max(0, px.Min.X-union.Min.X),
		right:  math.Max(0, union.Max.X-px.Max.X),
		top:    math.Max(0, union.Max.Y-px.Max.Y),
		bottom: math.Max(0, px.Min.Y-union.Min.Y),
	}
}

func titleBounds(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) (geom.Rect, bool) {
	if ax == nil || ax.Title == "" {
		return geom.Rect{}, false
	}
	layout := measureSingleLineTextLayout(r, ax.Title, titleFontSize(ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
	return textInkRect(alignedSingleLineOrigin(titleAnchorPoint(ax, r, ctx, px, alignment), layout, TextAlignCenter, textLayoutVAlignBaseline), layout)
}

func xLabelBounds(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) (geom.Rect, bool) {
	if ax == nil || ax.XLabel == "" {
		return geom.Rect{}, false
	}
	side := ax.effectiveXLabelSide()
	size := axisLabelFontSize(ctx)
	layout := measureSingleLineTextLayout(r, ax.XLabel, size, ctx.RC.FontKey, ctx.RC.UseTeX)
	anchor, vAlign := xLabelAnchorPoint(ax, r, ctx, px, side, alignment)
	lineHeight := math.Max(layout.Height, pointsToPixels(ctx.RC, size))
	return alignedTextLayoutRect(anchor, layout, TextAlignCenter, vAlign, lineHeight)
}

func yLabelBounds(ax *Axes, r render.Renderer, ctx *DrawContext, px geom.Rect, alignment figureTextAlignment) (geom.Rect, bool) {
	if ax == nil || ax.YLabel == "" {
		return geom.Rect{}, false
	}
	side := ax.effectiveYLabelSide()
	size := axisLabelFontSize(ctx)
	layout := measureSingleLineTextLayout(r, ax.YLabel, size, ctx.RC.FontKey, ctx.RC.UseTeX)
	lineHeight := math.Max(layout.Height, pointsToPixels(ctx.RC, size))
	anchor := yLabelAnchorPoint(ax, r, ctx, px, side, alignment)
	centerY := px.Min.Y + px.H()/2
	if side == AxisRight {
		return geom.Rect{
			Min: geom.Pt{X: anchor.X, Y: centerY - layout.Width/2},
			Max: geom.Pt{X: anchor.X + lineHeight, Y: centerY + layout.Width/2},
		}, true
	}
	return geom.Rect{
		Min: geom.Pt{X: anchor.X - lineHeight, Y: centerY - layout.Width/2},
		Max: geom.Pt{X: anchor.X, Y: centerY + layout.Width/2},
	}, true
}

func figureLabelMarginsPx(fig *Figure, r render.Renderer, vp geom.Rect, engine LayoutEngine) figureMargin {
	if fig == nil {
		return figureMargin{}
	}
	pad := layoutPadPx(fig, engine)
	margins := figureMargin{}

	ctx := newFigureDrawContext(fig, vp)
	if fig.SupTitle != "" {
		margins.top += figureLabelTightHeight(r, fig.SupTitle, titleFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX) + 2*pad
	}
	if fig.SupXLabel != "" {
		margins.bottom += figureLabelTightHeight(r, fig.SupXLabel, figureLabelFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX) + 2*pad
	}
	if fig.SupYLabel != "" {
		margins.left += figureLabelTightHeight(r, fig.SupYLabel, figureLabelFontSize(ctx), fig.RC.FontKey, fig.RC.UseTeX) + 2*pad
	}
	return margins
}

func figureLabelTightHeight(r render.Renderer, text string, size float64, fontKey string, useTeX bool) float64 {
	lineHeight := measureSingleLineTextLayout(r, text, size, fontKey, useTeX).Height
	if bounder, ok := r.(render.TextBounder); ok {
		if bounds, ok := bounder.MeasureTextBounds(text, size, fontKey); ok && bounds.H > 0 {
			return math.Max(bounds.H, lineHeight)
		}
	}
	return lineHeight
}

func figureLayoutMarginsPx(fig *Figure, r render.Renderer, vp geom.Rect, engine LayoutEngine, layoutPass int) figureMargin {
	margins := figureLabelMarginsPx(fig, r, vp, engine)
	margins = addFigureMargins(margins, figureColorbarMarginsPx(fig, r, vp, engine, layoutPass))
	return margins
}

func addFigureMargins(a, b figureMargin) figureMargin {
	return figureMargin{
		left:   a.left + b.left,
		right:  a.right + b.right,
		top:    a.top + b.top,
		bottom: a.bottom + b.bottom,
	}
}

func figureColorbarMarginsPx(fig *Figure, r render.Renderer, vp geom.Rect, engine LayoutEngine, layoutPass int) figureMargin {
	if fig == nil {
		return figureMargin{}
	}
	alignment := computeFigureTextAlignment(fig, r, vp)
	margin := figureMargin{}
	for _, ax := range fig.Children {
		if ax == nil || ax.colorbarParent == nil {
			continue
		}
		location := ax.colorbarLocation
		if location == "" {
			location = "right"
		}
		if engine == LayoutEngineConstrained && ax.colorbarParent.subplotSpec != nil {
			base := ax.colorbarParent.RectFraction
			thickness := resolvedColorbarThickness(fig, base, ax.colorbarWidth, resolvedColorbarAspect(ax.colorbarAspect), location)
			if thickness <= 0 {
				continue
			}
			padding := measureAxesDecorationPadding(ax, fig, r, vp, alignment)
			colorbarPad := resolvedColorbarPadding(base, ax.colorbarPadding, location)
			if layoutPass == 0 {
				colorbarPad = 0
			}
			if colorbarIsHorizontal(location) {
				padding.bottom += ax.effectiveRC(fig).AxisLineWidth
				colorbarSpace := (thickness + colorbarPad) * vp.H()
				if location == "top" {
					margin.top = math.Max(margin.top, colorbarSpace+padding.top)
				} else {
					margin.bottom = math.Max(margin.bottom, colorbarSpace+padding.bottom)
				}
			} else {
				colorbarSpace := (thickness + colorbarPad) * vp.W()
				if location == "left" {
					margin.left = math.Max(margin.left, colorbarSpace+padding.left)
				} else {
					margin.right = math.Max(margin.right, colorbarSpace+padding.right)
				}
			}
			continue
		}
		base := colorbarLayoutBase(ax.colorbarParent, ax)
		if resolvedColorbarThickness(fig, base, ax.colorbarWidth, resolvedColorbarAspect(ax.colorbarAspect), location) <= 0 {
			continue
		}
		padding := measureAxesDecorationPadding(ax, fig, r, vp, alignment)
		if colorbarIsHorizontal(location) {
			padding.bottom += ax.effectiveRC(fig).AxisLineWidth
			if location == "top" {
				margin.top = math.Max(margin.top, padding.top)
			} else {
				margin.bottom = math.Max(margin.bottom, padding.bottom)
			}
		} else if location == "left" {
			padding.left += ax.effectiveRC(fig).AxisLineWidth
			margin.left = math.Max(margin.left, padding.left)
		} else {
			padding.right += ax.effectiveRC(fig).AxisLineWidth
			margin.right = math.Max(margin.right, padding.right)
		}
	}
	return margin
}

func layoutPadPx(fig *Figure, engine LayoutEngine) float64 {
	rc := fig.RC
	if rc.DPI <= 0 {
		rc = style.CurrentDefaults()
	}
	switch engine {
	case LayoutEngineConstrained:
		return pointsToPixels(rc, matplotlibConstrainedLayoutPadPoints)
	case LayoutEngineTight:
		return pointsToPixels(rc, matplotlibTightLayoutPadFontSize*rc.FontSize)
	default:
		return 0
	}
}

func constrainedLayoutPadPx(fig *Figure) float64 {
	if fig == nil {
		return 0
	}
	rc := fig.RC
	if rc.DPI <= 0 {
		rc = style.CurrentDefaults()
	}
	return pointsToPixels(rc, matplotlibConstrainedLayoutPadPoints)
}

func constrainedLayoutDefaultSpacePx(parentSpanPx float64, cells int) float64 {
	if parentSpanPx <= 0 || cells <= 0 {
		return 0
	}
	return parentSpanPx * matplotlibConstrainedLayoutSpace / float64(cells)
}

func capLayoutGap(gap, inner float64, count int) float64 {
	if gap <= 0 || inner <= 0 || count <= 1 {
		return 0
	}
	maxGap := inner / float64(count+1)
	if gap > maxGap {
		return maxGap
	}
	return gap
}

func syncAxesToSubplotSpecs(fig *Figure, state map[*GridSpec]GridSpecOptions) {
	if fig == nil {
		return
	}
	for _, ax := range fig.Children {
		if ax != nil && ax.subplotSpec != nil && !ax.positionManual {
			ax.RectFraction = ax.subplotSpec.rectWithOptions(state)
		}
	}
}

func syncColorbarAxes(fig *Figure) {
	syncColorbarAxesMeasured(fig, nil, geom.Rect{})
}

func syncColorbarAxesMeasured(fig *Figure, r render.Renderer, vp geom.Rect) {
	if fig == nil {
		return
	}
	var alignment figureTextAlignment
	if r != nil {
		alignment = computeFigureTextAlignment(fig, r, vp)
	}
	for _, ax := range fig.Children {
		if ax == nil || ax.colorbarParent == nil {
			continue
		}
		syncColorbarMapping(ax)
		parent := ax.colorbarParent
		base := colorbarLayoutBase(parent, ax)
		ax.colorbarBase = base
		location := ax.colorbarLocation
		if location == "" {
			location = "right"
		}
		padding := resolvedColorbarLayoutPadding(fig, base, ax.colorbarPadding, location)
		thickness := resolvedColorbarThickness(fig, base, ax.colorbarWidth, resolvedColorbarAspect(ax.colorbarAspect), location)
		slotThickness := resolvedColorbarSlotThickness(base, ax.colorbarWidth, location)
		useResolvedSlot := colorbarUsesResolvedSlot(fig, parent)
		if useResolvedSlot {
			base = parent.RectFraction
			ax.colorbarBase = base
			padding = resolvedColorbarPadding(base, ax.colorbarPadding, location)
			thickness = resolvedColorbarThickness(fig, base, ax.colorbarWidth, resolvedColorbarAspect(ax.colorbarAspect), location)
			slotThickness = resolvedColorbarSlotThickness(base, ax.colorbarWidth, location)
			slotOffset := math.NaN()
			if r != nil && location == "right" && fig.SizePx.X > 0 {
				slotOffset = measureAxesDecorationPadding(parent, fig, r, vp, alignment).right / fig.SizePx.X
			}
			_, ax.RectFraction = colorbarPlacementRectWithSlotOffset(fig, base, thickness, slotThickness, padding, location, useResolvedSlot, slotOffset)
			ax.RectFraction = insetColorbarRectForExtensions(fig, ax.RectFraction, ax.colorbarExtend, location)
			continue
		}
		parent.RectFraction, ax.RectFraction = colorbarPlacementRect(fig, base, thickness, slotThickness, padding, location, useResolvedSlot)
		ax.RectFraction = insetColorbarRectForExtensions(fig, ax.RectFraction, ax.colorbarExtend, location)
	}
}

func colorbarLayoutBase(parent, colorbar *Axes) geom.Rect {
	if colorbar == nil {
		return geom.Rect{}
	}
	base := colorbar.colorbarBase
	if parent == nil {
		return base
	}
	if parent.positionManual {
		return base
	}
	if parent.subplotSpec == nil {
		return base
	}
	if colorbar.RectFraction.W() > 0 && colorbar.RectFraction.Min.X > parent.RectFraction.Min.X {
		return geom.Rect{
			Min: parent.RectFraction.Min,
			Max: geom.Pt{
				X: colorbar.RectFraction.Max.X,
				Y: parent.RectFraction.Max.Y,
			},
		}
	}
	return parent.RectFraction
}

func axesByManagedGrid(fig *Figure) map[*GridSpec][]*Axes {
	out := map[*GridSpec][]*Axes{}
	if fig == nil {
		return out
	}
	for _, ax := range fig.Children {
		if ax == nil || ax.subplotSpec == nil || ax.subplotSpec.grid == nil {
			continue
		}
		out[ax.subplotSpec.grid] = append(out[ax.subplotSpec.grid], ax)
	}
	return out
}

func childGrids(gridAxes map[*GridSpec][]*Axes) map[*GridSpec][]*GridSpec {
	children := map[*GridSpec][]*GridSpec{}
	for grid := range gridAxes {
		if grid == nil || grid.parent == nil || grid.parent.grid == nil {
			continue
		}
		parent := grid.parent.grid
		children[parent] = append(children[parent], grid)
	}
	return children
}

func managedRootGrids(fig *Figure) []*GridSpec {
	seen := map[*GridSpec]bool{}
	var roots []*GridSpec
	for _, ax := range fig.Children {
		if ax == nil || ax.subplotSpec == nil || ax.subplotSpec.grid == nil {
			continue
		}
		grid := ax.subplotSpec.grid
		for grid != nil && grid.parent != nil && grid.parent.grid != nil {
			grid = grid.parent.grid
		}
		if grid != nil && !seen[grid] {
			seen[grid] = true
			roots = append(roots, grid)
		}
	}
	return roots
}

func gridCoversWholeFigure(grid *GridSpec) bool {
	if grid == nil || grid.parent != nil {
		return false
	}
	return grid.base.Min.X == 0 && grid.base.Min.Y == 0 && grid.base.Max.X == 1 && grid.base.Max.Y == 1
}

func fractionRectToPixels(r geom.Rect, vp geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{
			X: vp.Min.X + r.Min.X*vp.W(),
			Y: vp.Min.Y + r.Min.Y*vp.H(),
		},
		Max: geom.Pt{
			X: vp.Min.X + r.Max.X*vp.W(),
			Y: vp.Min.Y + r.Max.Y*vp.H(),
		},
	}
}

func unionRect(a, b geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{
			X: math.Min(a.Min.X, b.Min.X),
			Y: math.Min(a.Min.Y, b.Min.Y),
		},
		Max: geom.Pt{
			X: math.Max(a.Max.X, b.Max.X),
			Y: math.Max(a.Max.Y, b.Max.Y),
		},
	}
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
