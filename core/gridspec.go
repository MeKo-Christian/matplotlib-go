package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

// GridSpecOption configures a GridSpec layout.
type GridSpecOption func(*GridSpecOptions)

// GridSpecOptions controls the layout envelope and cell ratios for a grid.
type GridSpecOptions struct {
	Left, Right, Bottom, Top float64
	WSpace                   float64
	HSpace                   float64
	WidthRatios              []float64
	HeightRatios             []float64
}

// GridSpec represents a grid layout over a figure-normalized rectangle.
type GridSpec struct {
	figure  *Figure
	base    geom.Rect
	parent  *SubplotSpec
	nRows   int
	nCols   int
	options GridSpecOptions
}

const (
	matplotlibSubplotLeft   = 0.125
	matplotlibSubplotRight  = 0.9
	matplotlibSubplotBottom = 0.11
	matplotlibSubplotTop    = 0.88
)

// SubplotSpec describes a span within a GridSpec.
type SubplotSpec struct {
	figure   *Figure
	grid     *GridSpec
	rowStart int
	rowEnd   int
	colStart int
	colEnd   int
}

// SubFigure represents a figure-normalized composition region inside a Figure.
type SubFigure struct {
	figure       *Figure
	RectFraction geom.Rect
}

type subplotAxesOptions struct {
	shareX     *Axes
	shareY     *Axes
	style      []style.Option
	projection string
}

// SubplotAxesOption configures axes creation from a subplot spec.
type SubplotAxesOption func(*subplotAxesOptions)

func defaultGridSpecOptions() GridSpecOptions {
	return GridSpecOptions{
		Left:   0,
		Right:  1,
		Bottom: 0,
		Top:    1,
		WSpace: 0.05,
		HSpace: 0.06,
	}
}

// WithGridSpecPadding sets the grid envelope relative to its parent bounds.
func WithGridSpecPadding(left, right, bottom, top float64) GridSpecOption {
	return func(cfg *GridSpecOptions) {
		cfg.Left = left
		cfg.Right = right
		cfg.Bottom = bottom
		cfg.Top = top
	}
}

// WithGridSpecSpacing sets inter-cell spacing relative to the parent bounds.
func WithGridSpecSpacing(wspace, hspace float64) GridSpecOption {
	return func(cfg *GridSpecOptions) {
		cfg.WSpace = wspace
		cfg.HSpace = hspace
	}
}

// WithGridSpecWidthRatios sets relative column widths.
func WithGridSpecWidthRatios(ratios ...float64) GridSpecOption {
	return func(cfg *GridSpecOptions) {
		cfg.WidthRatios = append([]float64(nil), ratios...)
	}
}

// WithGridSpecHeightRatios sets relative row heights.
func WithGridSpecHeightRatios(ratios ...float64) GridSpecOption {
	return func(cfg *GridSpecOptions) {
		cfg.HeightRatios = append([]float64(nil), ratios...)
	}
}

// WithSharedX reuses the x scale and x-axis state from the selected peer.
func WithSharedX(peer *Axes) SubplotAxesOption {
	return func(cfg *subplotAxesOptions) {
		cfg.shareX = peer
	}
}

// WithSharedY reuses the y scale and y-axis state from the selected peer.
func WithSharedY(peer *Axes) SubplotAxesOption {
	return func(cfg *subplotAxesOptions) {
		cfg.shareY = peer
	}
}

// WithSharedAxes reuses both x and y state from the selected peer.
func WithSharedAxes(peer *Axes) SubplotAxesOption {
	return func(cfg *subplotAxesOptions) {
		cfg.shareX = peer
		cfg.shareY = peer
	}
}

// WithSubplotStyle forwards style options when creating axes from a subplot spec.
func WithSubplotStyle(opts ...style.Option) SubplotAxesOption {
	return func(cfg *subplotAxesOptions) {
		cfg.style = append(cfg.style, opts...)
	}
}

// WithProjection selects a named projection when creating subplot axes.
func WithProjection(name string) SubplotAxesOption {
	return func(cfg *subplotAxesOptions) {
		cfg.projection = name
	}
}

// GridSpec creates a grid over the whole figure.
func (f *Figure) GridSpec(nRows, nCols int, opts ...GridSpecOption) *GridSpec {
	cfg := defaultSubplotOptions(f)
	seeded := subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)
	seeded = append(seeded, opts...)
	return newGridSpec(f, geom.Rect{Max: geom.Pt{X: 1, Y: 1}}, nil, nRows, nCols, seeded...)
}

// AddSubFigure creates a composition region inside the figure.
func (f *Figure) AddSubFigure(r geom.Rect) *SubFigure {
	if f == nil {
		return nil
	}
	return &SubFigure{
		figure:       f,
		RectFraction: r,
	}
}

// AddSubplot creates a subplot using Matplotlib-style row/column/index addressing.
func (f *Figure) AddSubplot(nRows, nCols, index int, opts ...SubplotAxesOption) *Axes {
	cfg := defaultSubplotOptions(f)
	gs := f.GridSpec(
		nRows,
		nCols,
		subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)...,
	)
	if gs == nil || index <= 0 || index > nRows*nCols {
		return nil
	}
	index--
	return gs.Cell(index/nCols, index%nCols).AddAxes(opts...)
}

// AddSubplotCode creates a subplot from a three-digit code such as 221.
func (f *Figure) AddSubplotCode(code int, opts ...SubplotAxesOption) *Axes {
	nRows := code / 100
	nCols := (code / 10) % 10
	index := code % 10
	if nRows <= 0 || nCols <= 0 || index <= 0 {
		return nil
	}
	return f.AddSubplot(nRows, nCols, index, opts...)
}

// AddSubplotSpec creates axes for an existing subplot specification.
func (f *Figure) AddSubplotSpec(spec SubplotSpec, opts ...SubplotAxesOption) *Axes {
	if f == nil || spec.figure != f {
		return nil
	}
	return spec.AddAxes(opts...)
}

// Subplot2Grid creates a subplot inside a logical grid using row/column spans.
func (f *Figure) Subplot2Grid(shape [2]int, loc [2]int, rowSpan, colSpan int, opts ...SubplotAxesOption) *Axes {
	cfg := defaultSubplotOptions(f)
	gs := f.GridSpec(
		shape[0],
		shape[1],
		subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)...,
	)
	if gs == nil {
		return nil
	}
	if rowSpan <= 0 {
		rowSpan = 1
	}
	if colSpan <= 0 {
		colSpan = 1
	}
	return gs.Span(loc[0], loc[1], rowSpan, colSpan).AddAxes(opts...)
}

// SubplotMosaic creates named subplot areas from a string matrix. Empty cells may
// be encoded as "" or ".".
func (f *Figure) SubplotMosaic(layout [][]string, opts ...GridSpecOption) (map[string]*Axes, error) {
	if f == nil {
		return nil, nil
	}
	if len(layout) == 0 || len(layout[0]) == 0 {
		return nil, fmt.Errorf("subplot mosaic layout must not be empty")
	}
	nCols := len(layout[0])
	for row := range layout {
		if len(layout[row]) != nCols {
			return nil, fmt.Errorf("subplot mosaic rows must have consistent widths")
		}
	}

	gs := f.GridSpec(len(layout), nCols, opts...)
	if gs == nil {
		return nil, fmt.Errorf("invalid subplot mosaic grid")
	}

	type bounds struct {
		minRow int
		maxRow int
		minCol int
		maxCol int
	}

	regions := map[string]bounds{}
	for row := range layout {
		for col, label := range layout[row] {
			if label == "" || label == "." {
				continue
			}
			b, ok := regions[label]
			if !ok {
				regions[label] = bounds{minRow: row, maxRow: row, minCol: col, maxCol: col}
				continue
			}
			if row < b.minRow {
				b.minRow = row
			}
			if row > b.maxRow {
				b.maxRow = row
			}
			if col < b.minCol {
				b.minCol = col
			}
			if col > b.maxCol {
				b.maxCol = col
			}
			regions[label] = b
		}
	}

	result := make(map[string]*Axes, len(regions))
	for label, region := range regions {
		for row := region.minRow; row <= region.maxRow; row++ {
			for col := region.minCol; col <= region.maxCol; col++ {
				if layout[row][col] != label {
					return nil, fmt.Errorf("subplot mosaic label %q must occupy a rectangular region", label)
				}
			}
		}
		result[label] = gs.Span(
			region.minRow,
			region.minCol,
			region.maxRow-region.minRow+1,
			region.maxCol-region.minCol+1,
		).AddAxes()
	}
	return result, nil
}

// GridSpec creates a grid inside the subfigure.
func (sf *SubFigure) GridSpec(nRows, nCols int, opts ...GridSpecOption) *GridSpec {
	if sf == nil {
		return nil
	}
	return newGridSpec(sf.figure, sf.RectFraction, nil, nRows, nCols, opts...)
}

// AddAxes creates axes positioned relative to the subfigure rectangle.
func (sf *SubFigure) AddAxes(r geom.Rect, opts ...style.Option) *Axes {
	if sf == nil || sf.figure == nil {
		return nil
	}
	return sf.figure.AddAxes(composeRect(sf.RectFraction, r), opts...)
}

// AddSubFigure creates a nested subfigure inside the current one.
func (sf *SubFigure) AddSubFigure(r geom.Rect) *SubFigure {
	if sf == nil || sf.figure == nil {
		return nil
	}
	return &SubFigure{
		figure:       sf.figure,
		RectFraction: composeRect(sf.RectFraction, r),
	}
}

// AddSubplot creates a subplot inside the subfigure using row/column/index addressing.
func (sf *SubFigure) AddSubplot(nRows, nCols, index int, opts ...SubplotAxesOption) *Axes {
	cfg := defaultSubplotOptions(sf.figure)
	gs := sf.GridSpec(
		nRows,
		nCols,
		subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)...,
	)
	if gs == nil || index <= 0 || index > nRows*nCols {
		return nil
	}
	index--
	return gs.Cell(index/nCols, index%nCols).AddAxes(opts...)
}

// AddSubplotSpec creates axes for an existing subplot specification inside the same figure.
func (sf *SubFigure) AddSubplotSpec(spec SubplotSpec, opts ...SubplotAxesOption) *Axes {
	if sf == nil || sf.figure == nil || spec.figure != sf.figure {
		return nil
	}
	return spec.AddAxes(opts...)
}

// Subplot2Grid creates a subplot inside a logical subfigure grid using spans.
func (sf *SubFigure) Subplot2Grid(shape [2]int, loc [2]int, rowSpan, colSpan int, opts ...SubplotAxesOption) *Axes {
	cfg := defaultSubplotOptions(sf.figure)
	gs := sf.GridSpec(
		shape[0],
		shape[1],
		subplotGridSpecOptions(cfg.Left, cfg.Right, cfg.Bottom, cfg.Top, cfg.WSpace, cfg.HSpace)...,
	)
	if gs == nil {
		return nil
	}
	if rowSpan <= 0 {
		rowSpan = 1
	}
	if colSpan <= 0 {
		colSpan = 1
	}
	return gs.Span(loc[0], loc[1], rowSpan, colSpan).AddAxes(opts...)
}

// GridSpec creates a nested grid inside the subplot span.
func (spec SubplotSpec) GridSpec(nRows, nCols int, opts ...GridSpecOption) *GridSpec {
	return newGridSpec(spec.figure, geom.Rect{}, &spec, nRows, nCols, opts...)
}

// Rect returns the figure-normalized rectangle occupied by this subplot span.
func (spec SubplotSpec) Rect() geom.Rect {
	return spec.rectWithOptions(nil)
}

func (spec SubplotSpec) rawRect() geom.Rect {
	if spec.grid == nil {
		return geom.Rect{}
	}
	return spec.grid.rawRectForSpan(spec.rowStart, spec.rowEnd, spec.colStart, spec.colEnd)
}

func (spec SubplotSpec) rectWithOptions(state map[*GridSpec]GridSpecOptions) geom.Rect {
	if spec.grid == nil {
		return geom.Rect{}
	}
	return spec.grid.rectForSpan(spec.rowStart, spec.rowEnd, spec.colStart, spec.colEnd, state)
}

// AddAxes creates axes covering this subplot span.
func (spec SubplotSpec) AddAxes(opts ...SubplotAxesOption) *Axes {
	if spec.figure == nil {
		return nil
	}

	cfg := subplotAxesOptions{}
	for _, opt := range opts {
		opt(&cfg)
	}

	rect := spec.Rect()
	var ax *Axes
	if cfg.projection != "" {
		var err error
		ax, err = spec.figure.AddAxesProjection(rect, cfg.projection, cfg.style...)
		if err != nil {
			return nil
		}
	} else {
		ax = spec.figure.AddAxes(rect, cfg.style...)
	}
	ax.RectFraction = rect
	ax.subplotSpec = &SubplotSpec{
		figure:   spec.figure,
		grid:     spec.grid,
		rowStart: spec.rowStart,
		rowEnd:   spec.rowEnd,
		colStart: spec.colStart,
		colEnd:   spec.colEnd,
	}
	if cfg.shareX != nil {
		root := cfg.shareX.xScaleRoot()
		ax.shareX = root
		ax.XAxis = root.XAxis
	}
	if cfg.shareY != nil {
		root := cfg.shareY.yScaleRoot()
		ax.shareY = root
		ax.YAxis = root.YAxis
	}
	return ax
}

// IsFirstRow reports whether the span touches the top row of its grid.
func (spec SubplotSpec) IsFirstRow() bool {
	return spec.grid != nil && spec.rowStart == 0
}

// IsLastRow reports whether the span touches the bottom row of its grid.
func (spec SubplotSpec) IsLastRow() bool {
	return spec.grid != nil && spec.rowEnd == spec.grid.nRows
}

// IsFirstCol reports whether the span touches the leftmost column of its grid.
func (spec SubplotSpec) IsFirstCol() bool {
	return spec.grid != nil && spec.colStart == 0
}

// IsLastCol reports whether the span touches the rightmost column of its grid.
func (spec SubplotSpec) IsLastCol() bool {
	return spec.grid != nil && spec.colEnd == spec.grid.nCols
}

// LabelOuter hides tick labels (and the offset text and axis label) on the
// inner edges of a gridded subplot, leaving them only on the bottom row and the
// leftmost column — mirroring Matplotlib's Axes.label_outer for the default
// bottom/left label positions. It is a no-op for axes not created from a grid.
func (a *Axes) LabelOuter() {
	if a == nil || a.subplotSpec == nil {
		return
	}
	ss := a.subplotSpec
	if !ss.IsLastRow() {
		a.hideXTickLabels = true
		a.hideXLabel = true
	}
	if !ss.IsFirstRow() {
		a.hideTopTickLabels = true
	}
	if !ss.IsFirstCol() {
		a.hideYTickLabels = true
		a.hideYLabel = true
	}
	if !ss.IsLastCol() {
		a.hideRightTickLabels = true
	}
}

// tickLabelsHiddenForLayout reports whether the given axis's tick labels are
// suppressed (e.g. by LabelOuter) on this Axes, so the layout pass should not
// reserve space for them. Mirrors the axis→hide-flag pairing used at draw time
// in figure_draw.go.
func (a *Axes) tickLabelsHiddenForLayout(axis *Axis) bool {
	if a == nil || axis == nil {
		return false
	}
	switch axis {
	case a.effectiveXAxis():
		return a.hideXTickLabels
	case a.effectiveYAxis():
		return a.hideYTickLabels
	case a.effectiveTopAxis():
		return a.hideTopTickLabels
	case a.effectiveRightAxis():
		return a.hideRightTickLabels
	}
	return false
}

// SubFigure converts the subplot span into a composition region.
func (spec SubplotSpec) SubFigure() *SubFigure {
	if spec.figure == nil {
		return nil
	}
	return &SubFigure{
		figure:       spec.figure,
		RectFraction: spec.rawRect(),
	}
}

// Cell returns the subplot specification for a single grid cell.
func (gs *GridSpec) Cell(row, col int) SubplotSpec {
	return gs.Span(row, col, 1, 1)
}

// Span returns the subplot specification covering the requested row/column span.
func (gs *GridSpec) Span(row, col, rowSpan, colSpan int) SubplotSpec {
	if gs == nil || rowSpan <= 0 || colSpan <= 0 {
		return SubplotSpec{}
	}
	return gs.slice(row, row+rowSpan, col, col+colSpan)
}

// Slice returns the subplot specification spanning the half-open row/column range.
func (gs *GridSpec) Slice(rowStart, rowEnd, colStart, colEnd int) SubplotSpec {
	return gs.slice(rowStart, rowEnd, colStart, colEnd)
}

func newGridSpec(figure *Figure, base geom.Rect, parent *SubplotSpec, nRows, nCols int, opts ...GridSpecOption) *GridSpec {
	cfg := defaultGridSpecOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	if figure == nil || nRows <= 0 || nCols <= 0 {
		return nil
	}
	if cfg.Right <= cfg.Left || cfg.Top <= cfg.Bottom {
		return nil
	}
	if len(cfg.WidthRatios) > 0 && len(cfg.WidthRatios) != nCols {
		return nil
	}
	if len(cfg.HeightRatios) > 0 && len(cfg.HeightRatios) != nRows {
		return nil
	}
	if ratioSum(cfg.WidthRatios) == 0 && len(cfg.WidthRatios) > 0 {
		return nil
	}
	if ratioSum(cfg.HeightRatios) == 0 && len(cfg.HeightRatios) > 0 {
		return nil
	}

	return &GridSpec{
		figure:  figure,
		base:    base,
		parent:  parent,
		nRows:   nRows,
		nCols:   nCols,
		options: cfg,
	}
}

func (gs *GridSpec) slice(rowStart, rowEnd, colStart, colEnd int) SubplotSpec {
	if gs == nil || rowStart < 0 || colStart < 0 || rowEnd <= rowStart || colEnd <= colStart {
		return SubplotSpec{}
	}
	if rowEnd > gs.nRows || colEnd > gs.nCols {
		return SubplotSpec{}
	}

	return SubplotSpec{
		figure:   gs.figure,
		grid:     gs,
		rowStart: rowStart,
		rowEnd:   rowEnd,
		colStart: colStart,
		colEnd:   colEnd,
	}
}

func (gs *GridSpec) rectForSpan(rowStart, rowEnd, colStart, colEnd int, state map[*GridSpec]GridSpecOptions) geom.Rect {
	if gs == nil {
		return geom.Rect{}
	}

	opts := gs.optionsForState(state)
	parent := gs.parentRectForState(state)
	inner := composeRect(parent, geom.Rect{
		Min: geom.Pt{X: opts.Left, Y: opts.Bottom},
		Max: geom.Pt{X: opts.Right, Y: opts.Top},
	})

	baseW := inner.W()
	baseH := inner.H()
	spacingW := opts.WSpace * baseW
	spacingH := opts.HSpace * baseH
	availableW := baseW - spacingW*float64(gs.nCols-1)
	availableH := baseH - spacingH*float64(gs.nRows-1)
	if availableW <= 0 || availableH <= 0 {
		return geom.Rect{}
	}

	widths := distributeRatios(opts.WidthRatios, gs.nCols, availableW)
	heights := distributeRatios(opts.HeightRatios, gs.nRows, availableH)

	minX := inner.Min.X + accumulate(widths, colStart) + spacingW*float64(colStart)
	maxX := minX + accumulate(widths[colStart:colEnd], len(widths[colStart:colEnd])) + spacingW*float64(colEnd-colStart-1)

	offsetFromTop := accumulate(heights, rowStart) + spacingH*float64(rowStart)
	maxY := inner.Max.Y - offsetFromTop
	minY := maxY - accumulate(heights[rowStart:rowEnd], len(heights[rowStart:rowEnd])) - spacingH*float64(rowEnd-rowStart-1)

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func (gs *GridSpec) rawRectForSpan(rowStart, rowEnd, colStart, colEnd int) geom.Rect {
	if gs == nil {
		return geom.Rect{}
	}

	parent := gs.rawParentRect()
	baseW := parent.W()
	baseH := parent.H()
	if baseW <= 0 || baseH <= 0 {
		return geom.Rect{}
	}

	widths := distributeRatios(gs.options.WidthRatios, gs.nCols, baseW)
	heights := distributeRatios(gs.options.HeightRatios, gs.nRows, baseH)

	minX := parent.Min.X + accumulate(widths, colStart)
	maxX := minX + accumulate(widths[colStart:colEnd], len(widths[colStart:colEnd]))

	offsetFromTop := accumulate(heights, rowStart)
	maxY := parent.Max.Y - offsetFromTop
	minY := maxY - accumulate(heights[rowStart:rowEnd], len(heights[rowStart:rowEnd]))

	return geom.Rect{
		Min: geom.Pt{X: minX, Y: minY},
		Max: geom.Pt{X: maxX, Y: maxY},
	}
}

func (gs *GridSpec) rawParentRect() geom.Rect {
	if gs == nil {
		return geom.Rect{}
	}
	if gs.parent != nil {
		return gs.parent.rawRect()
	}
	return gs.base
}

func (gs *GridSpec) parentRectForState(state map[*GridSpec]GridSpecOptions) geom.Rect {
	if gs == nil {
		return geom.Rect{}
	}
	if gs.parent != nil {
		return gs.parent.rectWithOptions(state)
	}
	return gs.base
}

func (gs *GridSpec) optionsForState(state map[*GridSpec]GridSpecOptions) GridSpecOptions {
	if state != nil {
		if opts, ok := state[gs]; ok {
			return opts
		}
	}
	return gs.options
}

func composeRect(parent, child geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{
			X: parent.Min.X + child.Min.X*parent.W(),
			Y: parent.Min.Y + child.Min.Y*parent.H(),
		},
		Max: geom.Pt{
			X: parent.Min.X + child.Max.X*parent.W(),
			Y: parent.Min.Y + child.Max.Y*parent.H(),
		},
	}
}

func distributeRatios(ratios []float64, count int, available float64) []float64 {
	values := make([]float64, count)
	if count == 0 {
		return values
	}

	sum := ratioSum(ratios)
	if sum == 0 {
		each := available / float64(count)
		for i := range values {
			values[i] = each
		}
		return values
	}

	for i := range values {
		values[i] = available * ratios[i] / sum
	}
	return values
}

func ratioSum(ratios []float64) float64 {
	sum := 0.0
	for _, v := range ratios {
		if v > 0 {
			sum += v
		}
	}
	return sum
}

func accumulate(values []float64, count int) float64 {
	sum := 0.0
	for i := 0; i < count && i < len(values); i++ {
		sum += values[i]
	}
	return sum
}
