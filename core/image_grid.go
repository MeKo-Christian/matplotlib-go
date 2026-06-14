package core

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

type AxesDivider struct {
	figure      *Figure
	rect        geom.Rect
	rows        int
	cols        int
	hSpace      float64
	vSpace      float64
	widthScale  []float64
	heightScale []float64
	cellAspect  float64
}

type AxesDividerOption func(*AxesDivider)

// WithAxesDividerHorizontalSpace configures gutter space between columns.
func WithAxesDividerHorizontalSpace(space float64) AxesDividerOption {
	return func(d *AxesDivider) {
		if space < 0 {
			space = 0
		}
		d.hSpace = space
	}
}

// WithAxesDividerVerticalSpace configures gutter space between rows.
func WithAxesDividerVerticalSpace(space float64) AxesDividerOption {
	return func(d *AxesDivider) {
		if space < 0 {
			space = 0
		}
		d.vSpace = space
	}
}

// WithAxesDividerWidthScales normalizes these relative widths across grid columns.
func WithAxesDividerWidthScales(scales ...float64) AxesDividerOption {
	return func(d *AxesDivider) {
		d.widthScale = append([]float64(nil), scales...)
	}
}

// WithAxesDividerHeightScales normalizes these relative heights across grid rows.
func WithAxesDividerHeightScales(scales ...float64) AxesDividerOption {
	return func(d *AxesDivider) {
		d.heightScale = append([]float64(nil), scales...)
	}
}

// WithAxesDividerCellAspect constrains grid cells to the requested
// height/width ratio in physical pixels, centering the packed grid.
func WithAxesDividerCellAspect(aspect float64) AxesDividerOption {
	return func(d *AxesDivider) {
		if aspect <= 0 {
			d.cellAspect = 0
			return
		}
		d.cellAspect = aspect
	}
}

// NewAxesDivider creates a light-weight layout helper for structured axes tiling.
func (f *Figure) NewAxesDivider(rect geom.Rect, rows, cols int, opts ...AxesDividerOption) *AxesDivider {
	if f == nil || rows <= 0 || cols <= 0 || rect.W() <= 0 || rect.H() <= 0 {
		return nil
	}
	divider := &AxesDivider{
		figure: f,
		rect:   rect,
		rows:   rows,
		cols:   cols,
		hSpace: 0.02,
		vSpace: 0.02,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(divider)
		}
	}
	return divider
}

func (d *AxesDivider) axisRect(row, col int) (geom.Rect, bool) {
	if d == nil || d.figure == nil || row < 0 || row >= d.rows || col < 0 || col >= d.cols {
		return geom.Rect{}, false
	}
	widths := normalizedRatios(d.widthScale, d.cols)
	heights := normalizedRatios(d.heightScale, d.rows)

	layoutRect := d.aspectAdjustedRect()
	usableW := layoutRect.W() - d.hSpace*float64(d.cols-1)
	usableH := layoutRect.H() - d.vSpace*float64(d.rows-1)
	if usableW < 0 || usableH < 0 {
		return geom.Rect{}, false
	}

	xCur := layoutRect.Min.X + accumulate(widths[:col], col)*usableW + d.hSpace*float64(col)
	maxY := layoutRect.Max.Y - accumulate(heights[:row], row)*usableH - d.vSpace*float64(row)
	w := widths[col] * usableW
	h := heights[row] * usableH
	return geom.Rect{
		Min: geom.Pt{X: xCur, Y: maxY - h},
		Max: geom.Pt{X: xCur + w, Y: maxY},
	}, true
}

// AddAxes adds one axes in the chosen grid cell.
func (d *AxesDivider) AddAxes(row, col int, opts ...style.Option) *Axes {
	rect, ok := d.axisRect(row, col)
	if !ok {
		return nil
	}
	return d.figure.AddAxes(rect, opts...)
}

// AddAxesProjection adds one axes with an explicit projection in the chosen cell.
func (d *AxesDivider) AddAxesProjection(row, col int, projection string, opts ...style.Option) (*Axes, error) {
	rect, ok := d.axisRect(row, col)
	if !ok {
		return nil, fmt.Errorf("axes cell (%d, %d) out of grid bounds", row, col)
	}
	return d.figure.AddAxesProjection(rect, projection, opts...)
}

func (d *AxesDivider) aspectAdjustedRect() geom.Rect {
	if d == nil || d.figure == nil || d.cellAspect <= 0 || d.rows <= 0 || d.cols <= 0 {
		if d == nil {
			return geom.Rect{}
		}
		return d.rect
	}

	gapW := d.hSpace * float64(d.cols-1)
	gapH := d.vSpace * float64(d.rows-1)
	usableW := d.rect.W() - gapW
	usableH := d.rect.H() - gapH
	if usableW <= 0 || usableH <= 0 || d.figure.SizePx.X <= 0 || d.figure.SizePx.Y <= 0 {
		return d.rect
	}

	target := d.cellAspect * float64(d.rows) / float64(d.cols)
	usableWPx := usableW * d.figure.SizePx.X
	usableHPx := usableH * d.figure.SizePx.Y
	current := usableHPx / usableWPx
	if current == target {
		return d.rect
	}

	out := d.rect
	if current > target {
		newUsableHPx := usableWPx * target
		newUsableH := newUsableHPx / d.figure.SizePx.Y
		newH := newUsableH + gapH
		pad := (d.rect.H() - newH) / 2
		out.Min.Y += pad
		out.Max.Y -= pad
		return out
	}

	newUsableWPx := usableHPx / target
	newUsableW := newUsableWPx / d.figure.SizePx.X
	newW := newUsableW + gapW
	pad := (d.rect.W() - newW) / 2
	out.Min.X += pad
	out.Max.X -= pad
	return out
}

// RGBAxes holds three synchronized axes intended for channel-wise RGB workflows.
type RGBAxes struct {
	Red   *Axes
	Green *Axes
	Blue  *Axes

	divider *AxesDivider
}

// NewRGBAxes creates three shared-viewport axes across a single row for RGB
// composition-style layouts.
func (f *Figure) NewRGBAxes(rect geom.Rect, dividerOpts ...AxesDividerOption) *RGBAxes {
	divider := f.NewAxesDivider(rect, 1, 3, dividerOpts...)
	if divider == nil {
		return nil
	}

	red := divider.AddAxes(0, 0)
	green := divider.AddAxes(0, 1)
	blue := divider.AddAxes(0, 2)
	if red == nil || green == nil || blue == nil {
		return nil
	}

	green.shareX = red
	green.XAxis = red.XAxis
	green.shareY = red
	green.YAxis = red.YAxis

	blue.shareX = red
	blue.XAxis = red.XAxis
	blue.shareY = red
	blue.YAxis = red.YAxis

	return &RGBAxes{
		Red:     red,
		Green:   green,
		Blue:    blue,
		divider: divider,
	}
}

type ImageGrid struct {
	Axes    [][]*Axes
	divider *AxesDivider
}

// NewImageGrid creates an evenly spaced image-grid of axes over figure fractions.
func (f *Figure) NewImageGrid(rows, cols int, rect geom.Rect, dividerOpts ...AxesDividerOption) *ImageGrid {
	opts := append([]AxesDividerOption{WithAxesDividerCellAspect(1)}, dividerOpts...)
	divider := f.NewAxesDivider(rect, rows, cols, opts...)
	if divider == nil {
		return nil
	}
	axes := make([][]*Axes, rows)
	for row := 0; row < rows; row++ {
		axes[row] = make([]*Axes, cols)
		for col := 0; col < cols; col++ {
			axes[row][col] = divider.AddAxes(row, col)
			if axes[row][col] == nil {
				return nil
			}
			if row < rows-1 && axes[row][col].XAxis != nil {
				axes[row][col].XAxis.ShowLabels = false
			}
			if col > 0 && axes[row][col].YAxis != nil {
				axes[row][col].YAxis.ShowLabels = false
			}
		}
	}
	return &ImageGrid{
		Axes:    axes,
		divider: divider,
	}
}

// At returns the axes in the requested grid cell.
func (g *ImageGrid) At(row, col int) *Axes {
	if g == nil || row < 0 || col < 0 {
		return nil
	}
	if row >= len(g.Axes) {
		return nil
	}
	if col >= len(g.Axes[row]) {
		return nil
	}
	return g.Axes[row][col]
}

func normalizedRatios(raw []float64, count int) []float64 {
	if count <= 0 {
		return nil
	}
	out := make([]float64, count)
	if len(raw) == count {
		sum := 0.0
		for _, value := range raw {
			sum += value
		}
		if sum > 0 {
			for i, value := range raw {
				out[i] = value / sum
			}
			return out
		}
	}
	for i := 0; i < count; i++ {
		out[i] = 1.0 / float64(count)
	}
	return out
}
