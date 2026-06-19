package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// TableOptions configures Axes.Table.
type TableOptions struct {
	CellText        [][]string
	CellColors      [][]render.Color
	RowLabels       []string
	ColLabels       []string
	BBox            geom.Rect
	Coords          CoordinateSpec
	FontSize        float64
	TextColor       render.Color
	HeaderTextColor render.Color
	HeaderFillColor render.Color
	CellFillColor   render.Color
	EdgeColor       render.Color
	LineWidth       float64
	CellLoc         string
	RowLoc          string
	ColLoc          string
}

type tableCell struct {
	Text       string
	Fill       render.Color
	IsHeader   bool
	IsRowLabel bool
	Rect       geom.Rect
	HAlign     TextAlign
}

// Table renders a simple grid of cells with optional row and column headers.
type Table struct {
	Cells           [][]tableCell
	BBox            geom.Rect
	Coords          CoordinateSpec
	FontSize        float64
	TextColor       render.Color
	HeaderTextColor render.Color
	EdgeColor       render.Color
	LineWidth       float64
	ClipOn          bool
	z               float64
}

// Table adds a simple table artist positioned in axes coordinates by default.
func (a *Axes) Table(opts ...TableOptions) *Table {
	if a == nil {
		return nil
	}
	cfg := TableOptions{
		BBox:            geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}},
		Coords:          Coords(CoordAxes),
		FontSize:        a.resolvedRC().FontSize,
		TextColor:       a.resolvedRC().DefaultTextColor(),
		HeaderTextColor: a.resolvedRC().DefaultTextColor(),
		HeaderFillColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		CellFillColor:   render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor:       render.Color{R: 0, G: 0, B: 0, A: 1},
		LineWidth:       a.resolvedRC().DPI / 72.0,
		CellLoc:         "right",
		RowLoc:          "left",
		ColLoc:          "center",
	}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.BBox == (geom.Rect{}) {
			cfg.BBox = geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1, Y: 1}}
		}
		if cfg.Coords == (CoordinateSpec{}) {
			cfg.Coords = Coords(CoordAxes)
		}
		if cfg.FontSize <= 0 {
			cfg.FontSize = a.resolvedRC().FontSize
		}
		if cfg.TextColor == (render.Color{}) {
			cfg.TextColor = a.resolvedRC().DefaultTextColor()
		}
		if cfg.HeaderTextColor == (render.Color{}) {
			cfg.HeaderTextColor = a.resolvedRC().DefaultTextColor()
		}
		if cfg.HeaderFillColor == (render.Color{}) {
			cfg.HeaderFillColor = render.Color{R: 1, G: 1, B: 1, A: 1}
		}
		if cfg.CellFillColor == (render.Color{}) {
			cfg.CellFillColor = render.Color{R: 1, G: 1, B: 1, A: 1}
		}
		if cfg.EdgeColor == (render.Color{}) {
			cfg.EdgeColor = render.Color{R: 0, G: 0, B: 0, A: 1}
		}
		if cfg.LineWidth <= 0 {
			cfg.LineWidth = a.resolvedRC().DPI / 72.0
		}
		if cfg.CellLoc == "" {
			cfg.CellLoc = "right"
		}
		if cfg.RowLoc == "" {
			cfg.RowLoc = "left"
		}
		if cfg.ColLoc == "" {
			cfg.ColLoc = "center"
		}
	}

	rows := len(cfg.CellText)
	cols := 0
	for _, row := range cfg.CellText {
		cols = max(cols, len(row))
	}
	if rows == 0 && len(cfg.RowLabels) == 0 && len(cfg.ColLabels) == 0 {
		return nil
	}
	if cols == 0 {
		cols = 1
	}

	hasRowLabels := len(cfg.RowLabels) > 0
	hasColLabels := len(cfg.ColLabels) > 0
	gridRows := rows
	if hasColLabels {
		gridRows++
	}
	gridCols := cols + boolOffset(hasRowLabels)
	dataCellW := cfg.BBox.W() / float64(cols)
	rowLabelW := 0.0
	if hasRowLabels {
		rowLabelW = tableRowLabelWidth(cfg.RowLabels, dataCellW)
	}
	cellH := cfg.BBox.H() / float64(gridRows)
	cellAlign := tableTextAlign(cfg.CellLoc, TextAlignRight)
	rowAlign := tableTextAlign(cfg.RowLoc, TextAlignLeft)
	colAlign := tableTextAlign(cfg.ColLoc, TextAlignCenter)

	cells := make([][]tableCell, gridRows)
	for r := range gridRows {
		cells[r] = make([]tableCell, gridCols)
		for c := range gridCols {
			cells[r][c] = tableCell{
				Fill:   cfg.CellFillColor,
				Rect:   tableCellRect(cfg.BBox, gridRows, cols, hasRowLabels, hasColLabels, rowLabelW, r, c),
				HAlign: cellAlign,
			}
		}
	}
	if hasColLabels {
		for c := 0; c < cols; c++ {
			col := c + boolOffset(hasRowLabels)
			cells[0][col] = tableCell{
				Text:     stringAt("", cfg.ColLabels, c),
				Fill:     cfg.HeaderFillColor,
				IsHeader: true,
				Rect:     tableCellRect(cfg.BBox, gridRows, cols, hasRowLabels, hasColLabels, rowLabelW, 0, col),
				HAlign:   colAlign,
			}
		}
	}
	if hasRowLabels {
		for r := 0; r < rows; r++ {
			row := r + boolOffset(hasColLabels)
			cells[row][0] = tableCell{
				Text:       stringAt("", cfg.RowLabels, r),
				Fill:       cfg.HeaderFillColor,
				IsHeader:   true,
				IsRowLabel: true,
				Rect: geom.Rect{
					Min: geom.Pt{X: cfg.BBox.Min.X - rowLabelW, Y: cfg.BBox.Max.Y - float64(row+1)*cellH},
					Max: geom.Pt{X: cfg.BBox.Min.X, Y: cfg.BBox.Max.Y - float64(row)*cellH},
				},
				HAlign: rowAlign,
			}
		}
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			fill := cfg.CellFillColor
			if r < len(cfg.CellColors) && c < len(cfg.CellColors[r]) && cfg.CellColors[r][c] != (render.Color{}) {
				fill = cfg.CellColors[r][c]
			}
			row := r + boolOffset(hasColLabels)
			col := c + boolOffset(hasRowLabels)
			cells[row][col] = tableCell{
				Text:   stringAt("", cfg.CellText[r], c),
				Fill:   fill,
				Rect:   tableCellRect(cfg.BBox, gridRows, cols, hasRowLabels, hasColLabels, rowLabelW, row, col),
				HAlign: cellAlign,
			}
		}
	}

	table := &Table{
		Cells:           cells,
		BBox:            cfg.BBox,
		Coords:          cfg.Coords,
		FontSize:        cfg.FontSize,
		TextColor:       cfg.TextColor,
		HeaderTextColor: cfg.HeaderTextColor,
		EdgeColor:       cfg.EdgeColor,
		LineWidth:       cfg.LineWidth,
		z:               4,
	}
	a.Add(table)
	return table
}

func tableTextAlign(loc string, fallback TextAlign) TextAlign {
	switch strings.ToLower(strings.TrimSpace(loc)) {
	case "left":
		return TextAlignLeft
	case "center", "centre":
		return TextAlignCenter
	case "right":
		return TextAlignRight
	default:
		return fallback
	}
}

func tableRowLabelWidth(labels []string, dataCellW float64) float64 {
	maxLen := 0
	for _, label := range labels {
		maxLen = max(maxLen, len([]rune(label)))
	}
	width := 0.015 + float64(maxLen)*0.025
	if width < 0.04 {
		width = 0.04
	}
	if dataCellW > 0 && width > dataCellW*0.5 {
		width = dataCellW * 0.5
	}
	return width
}

func tableCellRect(bbox geom.Rect, gridRows, dataCols int, hasRowLabels, hasColLabels bool, rowLabelW float64, row, col int) geom.Rect {
	if gridRows <= 0 || dataCols <= 0 {
		return geom.Rect{}
	}
	cellH := bbox.H() / float64(gridRows)
	y0 := bbox.Max.Y - float64(row+1)*cellH
	y1 := bbox.Max.Y - float64(row)*cellH
	if hasRowLabels && col == 0 {
		if hasColLabels && row == 0 {
			return geom.Rect{}
		}
		return geom.Rect{
			Min: geom.Pt{X: bbox.Min.X - rowLabelW, Y: y0},
			Max: geom.Pt{X: bbox.Min.X, Y: y1},
		}
	}

	dataCol := col
	if hasRowLabels {
		dataCol--
	}
	if dataCol < 0 || dataCol >= dataCols {
		return geom.Rect{}
	}
	cellW := bbox.W() / float64(dataCols)
	x0 := bbox.Min.X + float64(dataCol)*cellW
	return geom.Rect{
		Min: geom.Pt{X: x0, Y: y0},
		Max: geom.Pt{X: x0 + cellW, Y: y1},
	}
}

func tableTextAnchor(rect geom.Rect, align TextAlign) geom.Pt {
	x := rect.Min.X + rect.W()/2
	switch align {
	case TextAlignLeft:
		x = rect.Min.X + rect.W()*0.1
	case TextAlignRight:
		x = rect.Max.X - rect.W()*0.1
	}
	return geom.Pt{
		X: x,
		Y: rect.Min.Y + rect.H()/2,
	}
}

// Draw renders the table cells and text labels.
func (t *Table) Draw(r render.Renderer, ctx *DrawContext) {
	if t == nil || !t.ClipOn {
		return
	}
	t.drawTable(r, ctx)
}

// DrawOverlay renders unclipped tables after the axes clip has been removed.
func (t *Table) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if t == nil || t.ClipOn {
		return
	}
	t.drawTable(r, ctx)
}

func (t *Table) drawTable(r render.Renderer, ctx *DrawContext) {
	if t == nil || ctx == nil || r == nil || len(t.Cells) == 0 || len(t.Cells[0]) == 0 {
		return
	}
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}
	cells := t.layoutCells(r, ctx)

	for row := range cells {
		for col := range cells[row] {
			cell := cells[row][col]
			if cell.Rect == (geom.Rect{}) {
				continue
			}
			path := buildDisplayPath(ctx, t.Coords, patchRectPath(cell.Rect), geom.Identity())
			r.Path(path, &render.Paint{
				Fill:      cell.Fill,
				Stroke:    t.EdgeColor,
				LineWidth: t.LineWidth,
				LineJoin:  render.JoinMiter,
				LineCap:   render.CapButt,
				Snap:      render.SnapAuto,
			})

			text := cell.Text
			if displayTextIsEmpty(text) {
				continue
			}
			anchor := transformedPoint(ctx, t.Coords, tableTextAnchor(cell.Rect, cell.HAlign), 0, 0)
			layout := measureSingleLineTextLayout(r, text, t.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			color := t.TextColor
			if cell.IsHeader {
				color = t.HeaderTextColor
			}
			drawDisplayText(textRen, text, alignedTableTextOrigin(anchor, layout, cell.HAlign), t.FontSize, color, ctx.RC.FontKey, ctx.RC.UseTeX)
		}
	}
}

func (t *Table) layoutCells(r render.Renderer, ctx *DrawContext) [][]tableCell {
	cells := make([][]tableCell, len(t.Cells))
	for row := range t.Cells {
		cells[row] = make([]tableCell, len(t.Cells[row]))
		copy(cells[row], t.Cells[row])
	}

	rowLabelW, ok := t.measuredRowLabelWidth(r, ctx)
	if !ok {
		return cells
	}
	scaleX := t.BBox.W()
	if scaleX <= 0 {
		return cells
	}
	finalRowLabelW := rowLabelW * scaleX
	xShift := -rowLabelW * (1 - scaleX)
	dataLeft := t.BBox.Min.X + xShift
	for row := range cells {
		for col := range cells[row] {
			if cells[row][col].Rect == (geom.Rect{}) {
				continue
			}
			cells[row][col].Rect.Min.X += xShift
			cells[row][col].Rect.Max.X += xShift
			if cells[row][col].IsRowLabel {
				cells[row][col].Rect.Min.X = dataLeft - finalRowLabelW
				cells[row][col].Rect.Max.X = dataLeft
			}
		}
	}
	return cells
}

func (t *Table) measuredRowLabelWidth(r render.Renderer, ctx *DrawContext) (float64, bool) {
	if t == nil || r == nil || ctx == nil {
		return 0, false
	}
	coordsToDisplay := ctx.TransformFor(t.Coords)
	if coordsToDisplay == nil {
		return 0, false
	}
	x0 := coordsToDisplay.Apply(geom.Pt{X: t.BBox.Min.X, Y: t.BBox.Min.Y}).X
	x1 := coordsToDisplay.Apply(geom.Pt{X: t.BBox.Min.X + 1, Y: t.BBox.Min.Y}).X
	displayPerCoord := math.Abs(x1 - x0)
	if displayPerCoord <= 0 {
		return 0, false
	}

	maxWidthPx := 0.0
	for row := range t.Cells {
		for col := range t.Cells[row] {
			cell := t.Cells[row][col]
			if !cell.IsRowLabel || displayTextIsEmpty(cell.Text) {
				continue
			}
			layout := measureSingleLineTextLayout(r, cell.Text, t.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			widthPx := layout.Width
			if widthPx > maxWidthPx {
				maxWidthPx = widthPx
			}
		}
	}
	if maxWidthPx <= 0 {
		return 0, false
	}

	const matplotlibCellPad = 0.1
	return maxWidthPx * (1 + 2*matplotlibCellPad) / displayPerCoord, true
}

func alignedTableTextOrigin(anchor geom.Pt, layout singleLineTextLayout, hAlign TextAlign) geom.Pt {
	return alignedSingleLineOrigin(anchor, layout, hAlign, textLayoutVAlignCenter)
}

// Bounds returns an empty rect so table placement does not affect autoscaling.
func (t *Table) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Z returns the table's draw order.
func (t *Table) Z() float64 { return t.z }

func boolOffset(v bool) int {
	if v {
		return 1
	}
	return 0
}
