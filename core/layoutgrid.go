package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// layoutGrid is a direct, kiwisolver-free port of Matplotlib's
// matplotlib._layoutgrid.LayoutGrid used by constrained_layout.
//
// A grid is an nRows by nCols set of boxes. Each column owns left/right margins
// (and a "cb" colorbar/pad variant), each row owns top/bottom margins (and a
// "cb" variant). The "inner" width of each column (and height of each row),
// i.e. the span minus its margins, is constrained to be equal across the grid,
// scaled by the width/height ratios. The columns abut left-to-right and the
// rows abut top-to-bottom, with the first/last column/row pinned to the parent
// rectangle.
//
// Matplotlib expresses these relationships as kiwisolver constraints and lets
// the solver find lefts/rights/bottoms/tops. For the common rectangular grid
// (one constraint per gridspec) the system is fully determined and can be
// solved with a compact direct linear pass, which is what this type does.
//
// All coordinates are figure pixels; rows are numbered from the top, matching
// Matplotlib's convention (row 0 is the topmost row).
type layoutGrid struct {
	nRows int
	nCols int

	widthRatios  []float64
	heightRatios []float64

	// Parent rectangle in figure pixels (the area this grid is laid out in).
	parent geom.Rect

	// Per-column margins (length nCols). The non-cb margins hold Axes
	// decorations (tick labels, axis labels); the cb margins hold padding
	// and colorbars.
	marginLeft    []float64
	marginRight   []float64
	marginLeftCB  []float64
	marginRightCB []float64

	// Per-row margins (length nRows). Rows are top-down.
	marginTop      []float64
	marginBottom   []float64
	marginTopCB    []float64
	marginBottomCB []float64

	// Solved cell edges (figure pixels). Populated by solve().
	lefts   []float64
	rights  []float64
	tops    []float64
	bottoms []float64
}

func newLayoutGrid(nRows, nCols int, parent geom.Rect, widthRatios, heightRatios []float64) *layoutGrid {
	if nRows <= 0 {
		nRows = 1
	}
	if nCols <= 0 {
		nCols = 1
	}
	return &layoutGrid{
		nRows:          nRows,
		nCols:          nCols,
		widthRatios:    normalizedRatios(widthRatios, nCols),
		heightRatios:   normalizedRatios(heightRatios, nRows),
		parent:         parent,
		marginLeft:     make([]float64, nCols),
		marginRight:    make([]float64, nCols),
		marginLeftCB:   make([]float64, nCols),
		marginRightCB:  make([]float64, nCols),
		marginTop:      make([]float64, nRows),
		marginBottom:   make([]float64, nRows),
		marginTopCB:    make([]float64, nRows),
		marginBottomCB: make([]float64, nRows),
		lefts:          make([]float64, nCols),
		rights:         make([]float64, nCols),
		tops:           make([]float64, nRows),
		bottoms:        make([]float64, nRows),
	}
}

// editMarginMin grows the given margin for one cell if size exceeds the current
// value. Mirrors LayoutGrid.edit_margin_min. valid todo values:
// "left", "right", "leftcb", "rightcb" (column-indexed),
// "top", "bottom", "topcb", "bottomcb" (row-indexed).
func (g *layoutGrid) editMarginMin(todo string, size float64, cell int) {
	if g == nil || size <= 0 {
		return
	}
	target := g.marginSlice(todo)
	if target == nil || cell < 0 || cell >= len(target) {
		return
	}
	if size > target[cell] {
		target[cell] = size
	}
}

func (g *layoutGrid) marginSlice(todo string) []float64 {
	switch todo {
	case "left":
		return g.marginLeft
	case "right":
		return g.marginRight
	case "leftcb":
		return g.marginLeftCB
	case "rightcb":
		return g.marginRightCB
	case "top":
		return g.marginTop
	case "bottom":
		return g.marginBottom
	case "topcb":
		return g.marginTopCB
	case "bottomcb":
		return g.marginBottomCB
	}
	return nil
}

// solve computes lefts/rights/bottoms/tops by distributing the parent rect
// minus all margins across the columns/rows in proportion to the ratios.
//
// This is the direct equivalent of the kiwisolver pass: the inner widths of
// the columns are constrained equal (scaled by width_ratios) and the columns
// abut, so once the total margin overhead is known the inner widths are fully
// determined.
func (g *layoutGrid) solve() bool {
	if g == nil {
		return false
	}
	if !g.solveAxis(true) {
		return false
	}
	if !g.solveAxis(false) {
		return false
	}
	return true
}

// solveAxis solves either the horizontal (cols, horizontal=true) or vertical
// (rows, horizontal=false) direction.
func (g *layoutGrid) solveAxis(horizontal bool) bool {
	var (
		n        int
		ratios   []float64
		marginLo []float64 // left / top (per cell, "outer" decoration side)
		marginHi []float64 // right / bottom
		cbLo     []float64 // leftcb / topcb
		cbHi     []float64 // rightcb / bottomcb
		spanLo   float64
		spanHi   float64
	)
	if horizontal {
		n = g.nCols
		ratios = g.widthRatios
		marginLo = g.marginLeft
		marginHi = g.marginRight
		cbLo = g.marginLeftCB
		cbHi = g.marginRightCB
		spanLo = g.parent.Min.X
		spanHi = g.parent.Max.X
	} else {
		// Vertical axis is laid out top-down: row 0 is the topmost row, so the
		// "low" side here is the figure top and the "high" side is the bottom.
		n = g.nRows
		ratios = g.heightRatios
		marginLo = g.marginTop
		marginHi = g.marginBottom
		cbLo = g.marginTopCB
		cbHi = g.marginBottomCB
		spanLo = g.parent.Max.Y
		spanHi = g.parent.Min.Y
	}

	total := spanHi - spanLo // signed: positive for X, negative for Y (top-down)
	if total == 0 {
		return false
	}

	// Total margin overhead removed from the inner span.
	overhead := 0.0
	for i := 0; i < n; i++ {
		overhead += marginLo[i] + marginHi[i] + cbLo[i] + cbHi[i]
	}

	sign := 1.0
	if total < 0 {
		sign = -1.0
	}
	available := math.Abs(total) - overhead
	if available <= 0 {
		return false
	}

	ratioTotal := 0.0
	for _, rr := range ratios {
		ratioTotal += rr
	}
	if ratioTotal <= 0 {
		return false
	}

	// Walk cell by cell, abutting outer edges. Each cell occupies its inner
	// width plus its own four margins.
	cursor := spanLo
	for i := 0; i < n; i++ {
		inner := available * ratios[i] / ratioTotal
		cellOuter := inner + marginLo[i] + marginHi[i] + cbLo[i] + cbHi[i]
		lo := cursor
		hi := cursor + sign*cellOuter
		if horizontal {
			g.lefts[i] = lo
			g.rights[i] = hi
		} else {
			g.tops[i] = lo
			g.bottoms[i] = hi
		}
		cursor = hi
	}
	return true
}

// solveConstrainedGrid is the constrained_layout driver for a single managed
// gridspec. It is the direct port of do_constrained_layout / make_layout_margins
// for the (common) flat-grid case: it measures each Axes' decorations, suptitle,
// colorbar and figure-legend reservations into per-margin variables, solves the
// grid twice (so decorations settle after repositioning), and returns the
// resulting GridSpecOptions (Left/Right/Top/Bottom + WSpace/HSpace) so the
// existing syncAxesToSubplotSpecs machinery can position the Axes.
//
// Limitations vs. upstream:
//   - match_submerged_margins (equalizing interior margins of column/row-spanning
//     Axes in mosaics) is not implemented; plain grids are unaffected.
//   - Nested GridSpecFromSubplotSpec recursion reuses the existing recursion in
//     prepareFigureLayout; each grid is solved independently against its parent
//     rect rather than via a shared kiwisolver tree.
//   - Figure legends are reserved only for the outer ("outside") locations
//     (right/left/upper/lower), matching Matplotlib's _outside_loc handling.
func solveConstrainedGrid(fig *Figure, r render.Renderer, vp geom.Rect, grid *GridSpec, axes []*Axes, state map[*GridSpec]GridSpecOptions, alignment figureTextAlignment, layoutPass int) GridSpecOptions {
	if grid == nil {
		return GridSpecOptions{}
	}
	opts := grid.options

	parentRect := grid.parentRectForState(state)
	parentPx := fractionRectToPixels(parentRect, vp)
	if parentPx.W() <= 0 || parentPx.H() <= 0 {
		return opts
	}

	g := newLayoutGrid(grid.nRows, grid.nCols, parentPx, grid.options.WidthRatios, grid.options.HeightRatios)

	// Padding around the Axes (the "cb" margins start at the pad). These mirror
	// Matplotlib's w_pad/h_pad applied to every cell edge.
	wPad := layoutPadPx(fig, fig.layoutEngine)
	hPad := wPad
	if fig.layoutEngine == LayoutEngineConstrained {
		hPad = constrainedLayoutPadPx(fig)
	}

	// Inter-cell spacing (wspace/hspace) is split between adjacent cells and
	// lives in the "cb" margins. If the spacing is smaller than the pad, the
	// pad is used instead (Matplotlib get_margin_from_padding).
	wSpaceHalf := constrainedLayoutDefaultSpacePx(fig, parentPx.W(), grid.nCols, true) / 2
	hSpaceHalf := constrainedLayoutDefaultSpacePx(fig, parentPx.H(), grid.nRows, false) / 2

	for col := 0; col < grid.nCols; col++ {
		left := wPad
		right := wPad
		if col > 0 && wSpaceHalf > wPad {
			left = wSpaceHalf
		}
		if col < grid.nCols-1 && wSpaceHalf > wPad {
			right = wSpaceHalf
		}
		g.editMarginMin("leftcb", left, col)
		g.editMarginMin("rightcb", right, col)
	}
	for row := 0; row < grid.nRows; row++ {
		top := hPad
		bottom := hPad
		if row > 0 && hSpaceHalf > hPad {
			top = hSpaceHalf
		}
		if row < grid.nRows-1 && hSpaceHalf > hPad {
			bottom = hSpaceHalf
		}
		g.editMarginMin("topcb", top, row)
		g.editMarginMin("bottomcb", bottom, row)
	}

	// Decoration margins for each Axes in the grid.
	for _, ax := range axes {
		if ax == nil || ax.subplotSpec == nil {
			continue
		}
		padding := measureAxesDecorationPadding(ax, fig, r, vp, alignment)
		ss := ax.subplotSpec
		g.editMarginMin("left", padding.left, ss.colStart)
		g.editMarginMin("right", padding.right, ss.colEnd-1)
		g.editMarginMin("top", padding.top, ss.rowStart)
		g.editMarginMin("bottom", padding.bottom, ss.rowEnd-1)
	}

	// Colorbar reservations (only when this grid is the colorbar parent's grid).
	addColorbarMarginsToGrid(g, fig, r, vp, grid, alignment, layoutPass)

	// Figure-level reservations (suptitle / sup labels / outside legends) only
	// apply to the grid that covers the whole figure. Colorbars are handled
	// per-grid above (addColorbarMarginsToGrid), so we use figureLabelMarginsPx
	// rather than figureLayoutMarginsPx here to avoid reserving colorbar space
	// twice.
	if gridCoversWholeFigure(grid) {
		global := figureLabelMarginsPx(fig, r, vp, fig.layoutEngine)
		legend := figureLegendMarginsPx(fig, r, vp, fig.layoutEngine)
		global = addFigureMargins(global, legend)
		// Apply the figure-level margins to the outermost cells, on top of any
		// per-Axes decoration already accumulated there.
		g.editMarginMin("left", g.marginLeft[0]+global.left, 0)
		g.editMarginMin("right", g.marginRight[grid.nCols-1]+global.right, grid.nCols-1)
		g.editMarginMin("top", g.marginTop[0]+global.top, 0)
		g.editMarginMin("bottom", g.marginBottom[grid.nRows-1]+global.bottom, grid.nRows-1)
	}

	if !g.solve() {
		return grid.options
	}

	// Convert the solved grid back into GridSpecOptions relative to the parent.
	innerL := g.lefts[0] + g.marginLeft[0] + g.marginLeftCB[0]
	innerR := g.rights[grid.nCols-1] - g.marginRight[grid.nCols-1] - g.marginRightCB[grid.nCols-1]
	innerT := g.tops[0] - g.marginTop[0] - g.marginTopCB[0]
	innerB := g.bottoms[grid.nRows-1] + g.marginBottom[grid.nRows-1] + g.marginBottomCB[grid.nRows-1]

	leftPx := innerL - parentPx.Min.X
	rightPx := parentPx.Max.X - innerR
	// parentPx is y-up: top edge is Max.Y.
	topPx := parentPx.Max.Y - innerT
	bottomPx := innerB - parentPx.Min.Y

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

	// The gap between adjacent cells is the right margin of the left cell plus
	// the left margin of the right cell (decorations + the shared cb spacing).
	maxGapX := 0.0
	for col := 0; col < grid.nCols-1; col++ {
		gap := g.marginRight[col] + g.marginRightCB[col] + g.marginLeft[col+1] + g.marginLeftCB[col+1]
		maxGapX = math.Max(maxGapX, gap)
	}
	maxGapY := 0.0
	for row := 0; row < grid.nRows-1; row++ {
		gap := g.marginBottom[row] + g.marginBottomCB[row] + g.marginTop[row+1] + g.marginTopCB[row+1]
		maxGapY = math.Max(maxGapY, gap)
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

// addColorbarMarginsToGrid reserves "cb" margins for colorbars attached to Axes
// in this grid, mirroring the colorbar branch of make_layout_margins. The
// reserved space is the colorbar thickness, plus its pad, plus the colorbar's
// own decorations (its tick labels and label), measured with the same helpers
// the legacy heuristic used.
func addColorbarMarginsToGrid(g *layoutGrid, fig *Figure, r render.Renderer, vp geom.Rect, grid *GridSpec, alignment figureTextAlignment, layoutPass int) {
	if g == nil || fig == nil {
		return
	}
	for _, ax := range fig.Children {
		if ax == nil || ax.colorbarParent == nil {
			continue
		}
		parent := ax.colorbarParent
		if parent.subplotSpec == nil || parent.subplotSpec.grid != grid {
			continue
		}
		location := ax.colorbarLocation
		if location == "" {
			location = "right"
		}
		base := parent.RectFraction
		thickness := resolvedColorbarThickness(fig, base, ax.colorbarWidth, resolvedColorbarAspect(ax.colorbarAspect), location)
		if thickness <= 0 {
			continue
		}
		colorbarPad := resolvedColorbarPadding(base, ax.colorbarPadding, location)
		if layoutPass == 0 {
			colorbarPad = 0
		}
		// The colorbar's own decorations (its tick labels / label) extend past
		// the colorbar band and must also be reserved.
		decoration := measureAxesDecorationPadding(ax, fig, r, vp, alignment)
		ss := parent.subplotSpec
		if colorbarIsHorizontal(location) {
			space := (thickness+colorbarPad)*vp.H() + ax.effectiveRC(fig).AxisLineWidth
			if location == "top" {
				g.editMarginMin("topcb", g.marginTopCB[ss.rowStart]+space+decoration.top, ss.rowStart)
			} else {
				g.editMarginMin("bottomcb", g.marginBottomCB[ss.rowEnd-1]+space+decoration.bottom, ss.rowEnd-1)
			}
		} else {
			space := (thickness + colorbarPad) * vp.W()
			if location == "left" {
				g.editMarginMin("leftcb", g.marginLeftCB[ss.colStart]+space+decoration.left, ss.colStart)
			} else {
				g.editMarginMin("rightcb", g.marginRightCB[ss.colEnd-1]+space+decoration.right, ss.colEnd-1)
			}
		}
	}
}

// figureLegendMarginsPx reserves figure margins for figure-level legends placed
// outside the Axes, mirroring the figure-legend branch of make_layout_margins.
//
// Matplotlib reserves space only for figure legends with an explicit outside
// location. Outside and ordinary anchor location are independent: for example,
// OutsideRight plus LegendUpperRight mirrors "outside right upper".
func figureLegendMarginsPx(fig *Figure, r render.Renderer, vp geom.Rect, engine LayoutEngine) figureMargin {
	margins := figureMargin{}
	if fig == nil || r == nil {
		return margins
	}
	wPad := layoutPadPx(fig, engine)
	hPad := wPad
	if engine == LayoutEngineConstrained {
		hPad = constrainedLayoutPadPx(fig)
	}
	ctx := newFigureDrawContext(fig, vp)
	for _, art := range fig.Artists {
		leg, ok := art.(*Legend)
		if !ok || leg == nil || leg.Figure == nil {
			continue
		}
		box, ok := leg.boxRect(r, ctx)
		if !ok || box.W() <= 0 || box.H() <= 0 {
			continue
		}
		switch leg.Outside {
		case LegendOutsideRight:
			margins.right = math.Max(margins.right, box.W()+2*wPad)
		case LegendOutsideLeft:
			margins.left = math.Max(margins.left, box.W()+2*wPad)
		case LegendOutsideUpper:
			margins.top = math.Max(margins.top, box.H()+2*hPad)
		case LegendOutsideLower:
			margins.bottom = math.Max(margins.bottom, box.H()+2*hPad)
		}
	}
	return margins
}
