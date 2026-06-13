package core

import (
	"math"
	"sort"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// ContourOptions configures contour, contourf, and tricontour rendering.
type ContourOptions struct {
	X              []float64
	Y              []float64
	XEdges         []float64
	YEdges         []float64
	Levels         []float64
	LevelCount     int
	Colormap       *string
	Norm           ScalarNormalizer
	Colors         []render.Color
	Color          *render.Color
	LineWidth      *float64
	Alpha          *float64
	LabelLines     bool
	LabelFormatter Formatter
	LabelFontSize  *float64
	LabelColor     *render.Color
	Label          string
}

type contourLabel struct {
	Text     string
	Position geom.Pt
	Angle    float64
	Color    render.Color
	Level    float64
}

// ClabelOptions configures contour label placement for Axes.Clabel and
// ContourSet.Clabel.
type ClabelOptions struct {
	Levels          []float64
	Formatter       Formatter
	FontSize        *float64
	Color           *render.Color
	Colors          []render.Color
	Inline          *bool
	InlineSpacing   float64
	ManualPositions []geom.Pt
}

// ContourLabel stores the public metadata for a placed contour label.
type ContourLabel struct {
	Text     string
	Level    float64
	Position geom.Pt
	Angle    float64
	Color    render.Color
}

// ContourSet stores the artists created by contour/contourf calls.
type ContourSet struct {
	ArtistRasterization
	Levels         []float64
	Lines          *LineCollection
	Fills          *PolyCollection
	LabelFormatter Formatter
	LabelFontSize  float64
	LabelColor     render.Color
	labels         []contourLabel
	lineLevels     []float64
	labelLevels    []float64
	labelInlineGap float64
	inlineLabels   bool
	z              float64
}

// Clabel delegates contour labeling to the provided contour set, matching
// Matplotlib's Axes.clabel call shape.
func (a *Axes) Clabel(cs *ContourSet, opts ...ClabelOptions) []ContourLabel {
	if cs == nil {
		return nil
	}
	return cs.Clabel(opts...)
}

// Clabel adds labels to a contour set and returns metadata for the labels that
// were placed.
func (c *ContourSet) Clabel(opts ...ClabelOptions) []ContourLabel {
	if c == nil || c.Lines == nil || len(c.Lines.Segments) == 0 {
		return nil
	}
	var opt ClabelOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	indices, ok := c.clabelLineIndices(opt.Levels)
	if !ok || len(indices) == 0 {
		return nil
	}

	c.LabelFormatter = contourFormatter(firstFormatter(opt.Formatter, c.LabelFormatter))
	if opt.FontSize != nil && *opt.FontSize > 0 {
		c.LabelFontSize = *opt.FontSize
	} else if c.LabelFontSize <= 0 {
		c.LabelFontSize = 10
	}
	if opt.Color != nil {
		c.LabelColor = *opt.Color
	} else if len(opt.Colors) > 0 {
		c.LabelColor = opt.Colors[0]
	}
	c.inlineLabels = specialtyBool(opt.Inline, true)
	c.labelLevels = uniqueLevelsForIndices(c.lineLevels, indices)
	c.labelInlineGap = opt.InlineSpacing

	labels := c.clabelPlaceAutomatic(indices, opt)
	if len(opt.ManualPositions) > 0 {
		labels = c.clabelPlaceManual(indices, opt)
	}
	c.labels = labels
	return publicContourLabels(labels)
}

// Contour draws isolines over a rectilinear scalar grid.
func (a *Axes) Contour(data [][]float64, opts ...ContourOptions) *ContourSet {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opts)
	if !ok {
		return nil
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	levels := contourLevels(values, opt.Levels, opt.LevelCount, false)
	if len(levels) == 0 {
		return nil
	}

	polylines, polylineLevels := contourGridPolylines(xCoords, yCoords, data, levels)
	if len(polylines) == 0 {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	lineWidth := 1.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}
	colorFallback := a.NextColor()
	cmapName := ""
	if opt.Colormap != nil {
		cmapName = *opt.Colormap
	}
	mapping, err := ResolveScalarMapValues(values, ScalarMapConfig{
		Colormap: cmapName,
		Norm:     opt.Norm,
	})
	if err != nil {
		return nil
	}

	set := &ContourSet{
		Levels:         append([]float64(nil), levels...),
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  valueOrDefaultFloat(opt.LabelFontSize, 10),
		z:              defaultLineZ,
	}
	if opt.LabelColor != nil {
		set.LabelColor = *opt.LabelColor
	}
	colors := make([]render.Color, len(polylines))
	for i, level := range polylineLevels {
		colors[i] = contourLineColor(level, levels, opt, mapping, alpha, colorFallback)
	}
	set.Lines = &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  opt.Label,
			Alpha:  1,
		},
		Segments:  polylines,
		Colors:    colors,
		LineWidth: lineWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapButt,
	}
	set.lineLevels = append([]float64(nil), polylineLevels...)
	if opt.LabelLines {
		set.inlineLabels = true
		set.labels = contourLabels(polylines, polylineLevels, colors, set.LabelFormatter)
	}
	a.Add(set)
	return set
}

// Contourf draws filled contour bands over a rectilinear scalar grid.
func (a *Axes) Contourf(data [][]float64, opts ...ContourOptions) *ContourSet {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opts)
	if !ok {
		return nil
	}
	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	levels := contourLevels(values, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	cmapName := ""
	if opt.Colormap != nil {
		cmapName = *opt.Colormap
	}
	mapping, err := ResolveScalarMapValues(values, ScalarMapConfig{
		Colormap: cmapName,
		Norm:     opt.Norm,
	})
	if err != nil {
		return nil
	}
	if opt.Norm == nil {
		mapping.Norm = Normalize{VMin: levels[0], VMax: levels[len(levels)-1]}
		mapping.VMin = levels[0]
		mapping.VMax = levels[len(levels)-1]
	}

	polygons, faceColors := contourGridBandPolygons(xCoords, yCoords, data, levels, opt, mapping, alpha)
	if len(polygons) == 0 {
		return nil
	}
	cmap := ""
	vmin := 0.0
	vmax := 0.0
	if opt.Color == nil && len(opt.Colors) == 0 {
		cmap = mapping.Colormap
		vmin = mapping.VMin
		vmax = mapping.VMax
	}
	norm := mapping.Norm
	if cmap == "" {
		norm = nil
	}
	set := &ContourSet{
		Levels:         append([]float64(nil), levels...),
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  valueOrDefaultFloat(opt.LabelFontSize, 10),
	}
	if opt.LabelColor != nil {
		set.LabelColor = *opt.LabelColor
	}
	set.Fills = &PolyCollection{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:    Coords(CoordData),
				Label:     opt.Label,
				Alpha:     1,
				Colormap:  cmap,
				Norm:      norm,
				VMin:      vmin,
				VMax:      vmax,
				Antialias: render.AntialiasOff,
			},
			FaceColors: faceColors,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		Polygons: polygons,
	}
	a.Add(set)
	return set
}

// TriContour draws isolines over an explicit triangulation.
func (a *Axes) TriContour(tri Triangulation, values []float64, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, false, opts...)
}

// TriContourf draws filled contour bands over an explicit triangulation.
func (a *Axes) TriContourf(tri Triangulation, values []float64, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, true, opts...)
}

// Draw renders the contour set's filled bands and/or line collection.
func (c *ContourSet) Draw(r render.Renderer, ctx *DrawContext) {
	if c == nil {
		return
	}
	if c.Fills != nil {
		c.Fills.Draw(r, ctx)
	}
	if c.Lines != nil {
		if c.inlineLabels {
			c.drawInlineLabeledLines(r, ctx)
			return
		}
		c.Lines.Draw(r, ctx)
	}
}

// DrawOverlay renders contour labels outside the axes clip.
func (c *ContourSet) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil || len(c.labels) == 0 {
		return
	}

	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	fontSize := resolvedFontSize(c.LabelFontSize, ctx)
	for _, label := range c.labels {
		text := label.Text
		if displayTextIsEmpty(text) {
			continue
		}
		displayPt := ctx.DataToPixel.Apply(label.Position)
		color := label.Color
		if color == (render.Color{}) {
			color = resolvedTextColor(c.LabelColor, ctx)
		}

		if rotated, ok := r.(render.RotatedTextDrawer); ok {
			layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
			anchor := contourRotatedTextAnchor(displayPt, layout, label.Angle)
			drawDisplayTextRotated(rotated, text, anchor, fontSize, label.Angle, color, ctx.RC.FontKey, ctx.RC.UseTeX)
			continue
		}

		layout := measureSingleLineTextLayout(r, text, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		origin := alignedSingleLineOrigin(displayPt, layout, TextAlignCenter, textLayoutVAlignCenter)
		drawDisplayText(textRen, text, origin, fontSize, color, ctx.RC.FontKey, ctx.RC.UseTeX)
	}
}

func (c *ContourSet) drawInlineLabeledLines(r render.Renderer, ctx *DrawContext) {
	if c == nil || c.Lines == nil || ctx == nil {
		return
	}
	if _, ok := r.(render.TextDrawer); !ok {
		c.Lines.Draw(r, ctx)
		return
	}

	setRendererResolution(r, ctx.RC.DPI)
	fontSize := resolvedFontSize(c.LabelFontSize, ctx)
	inlineSpacing := c.labelInlineGap
	if inlineSpacing <= 0 {
		inlineSpacing = 5
	}
	segments, colors, widths, labels := contourInlineLabelSegmentsForLevels(c.Lines, c.lineLevels, c.labelLevels, c.LabelFormatter, fontSize, inlineSpacing, r, ctx)
	c.labels = labels
	if len(segments) == 0 {
		return
	}
	lines := *c.Lines
	lines.Segments = segments
	lines.Colors = colors
	lines.Color = render.Color{}
	lines.LineWidths = widths
	lines.LineWidth = c.Lines.LineWidth
	lines.Draw(r, ctx)
}

// Bounds returns the union of the contour set's line and fill geometry.
func (c *ContourSet) Bounds(ctx *DrawContext) geom.Rect {
	if c == nil {
		return geom.Rect{}
	}
	if c.Lines == nil {
		if c.Fills == nil {
			return geom.Rect{}
		}
		return c.Fills.Bounds(ctx)
	}
	if c.Fills == nil {
		return c.Lines.Bounds(ctx)
	}
	return unionRect(c.Lines.Bounds(ctx), c.Fills.Bounds(ctx))
}

// Z returns the contour set's draw order.
func (c *ContourSet) Z() float64 {
	if c == nil {
		return 0
	}
	if c.z != 0 {
		return c.z
	}
	if c.Lines != nil {
		return defaultLineZ
	}
	return defaultPatchZ
}

// ScalarMap exposes the contour fill's scalar mapping for colorbars.
func (c *ContourSet) ScalarMap() ScalarMapInfo {
	if c == nil || c.Fills == nil {
		return ScalarMapInfo{}
	}
	return c.Fills.ScalarMap()
}

func (c *ContourSet) legendEntry() (legendEntry, bool) {
	if c == nil {
		return legendEntry{}, false
	}
	if c.Fills != nil {
		return c.Fills.legendEntry()
	}
	if c.Lines != nil {
		return c.Lines.legendEntry()
	}
	return legendEntry{}, false
}

func (a *Axes) buildContourSet(tri Triangulation, values []float64, filled bool, opts ...ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	levels := contourLevels(values, opt.Levels, opt.LevelCount, filled)
	if (!filled && len(levels) == 0) || (filled && len(levels) < 2) {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	lineWidth := 1.0
	if opt.LineWidth != nil {
		lineWidth = *opt.LineWidth
	}

	colorFallback := a.NextColor()
	cmapName := ""
	if opt.Colormap != nil {
		cmapName = *opt.Colormap
	}
	mapping, err := ResolveScalarMapValues(values, ScalarMapConfig{
		Colormap: cmapName,
		Norm:     opt.Norm,
	})
	if err != nil {
		return nil
	}
	if filled {
		if opt.Norm == nil {
			mapping.Norm = Normalize{VMin: levels[0], VMax: levels[len(levels)-1]}
			mapping.VMin = levels[0]
			mapping.VMax = levels[len(levels)-1]
		}
	}

	set := &ContourSet{
		Levels:         append([]float64(nil), levels...),
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  valueOrDefaultFloat(opt.LabelFontSize, 10),
		z:              0,
	}
	if opt.LabelColor != nil {
		set.LabelColor = *opt.LabelColor
	}

	if filled {
		polygons, faceColors := contourBandPolygons(tri, values, levels, opt, mapping, alpha)
		if len(polygons) > 0 {
			cmap := ""
			vmin := 0.0
			vmax := 0.0
			if opt.Color == nil && len(opt.Colors) == 0 {
				cmap = mapping.Colormap
				vmin = mapping.VMin
				vmax = mapping.VMax
			}
			norm := mapping.Norm
			if cmap == "" {
				norm = nil
			}
			set.Fills = &PolyCollection{
				PatchCollection: PatchCollection{
					Collection: Collection{
						Coords:   Coords(CoordData),
						Label:    opt.Label,
						Alpha:    1,
						Colormap: cmap,
						Norm:     norm,
						VMin:     vmin,
						VMax:     vmax,
						// matplotlib ContourSet: antialiased=None with
						// filled=True resolves to False (contour.py), so
						// band fills render with hard edges.
						Antialias: render.AntialiasOff,
					},
					FaceColors: faceColors,
					LineJoin:   render.JoinMiter,
					LineCap:    render.CapButt,
				},
				Polygons: polygons,
			}
		}
	} else {
		polylines, polylineLevels := contourPolylines(tri, values, levels)
		if len(polylines) > 0 {
			colors := make([]render.Color, len(polylines))
			for i, level := range polylineLevels {
				colors[i] = contourLineColor(level, levels, opt, mapping, alpha, colorFallback)
			}
			set.Lines = &LineCollection{
				Collection: Collection{
					Coords: Coords(CoordData),
					Label:  opt.Label,
					Alpha:  1,
				},
				Segments:  polylines,
				Colors:    colors,
				LineWidth: lineWidth,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapButt,
			}
			set.lineLevels = append([]float64(nil), polylineLevels...)
			if opt.LabelLines {
				set.inlineLabels = true
				set.labels = contourLabels(polylines, polylineLevels, colors, set.LabelFormatter)
			}
		}
	}

	if set.Fills == nil && set.Lines == nil {
		return nil
	}
	a.Add(set)
	return set
}

func contourGridTriangulation(data [][]float64, opts []ContourOptions) (Triangulation, []float64, bool) {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opts)
	if !ok {
		return Triangulation{}, nil, false
	}
	rows := len(data)
	cols := len(data[0])

	xPoints := make([]float64, 0, rows*cols)
	yPoints := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		for xi := 0; xi < cols; xi++ {
			xPoints = append(xPoints, xCoords[xi])
			yPoints = append(yPoints, yCoords[yi])
		}
	}

	triangles := make([][3]int, 0, (rows-1)*(cols-1)*2)
	mask := make([]bool, 0, (rows-1)*(cols-1)*2)
	index := func(row, col int) int { return row*cols + col }
	for yi := 0; yi+1 < rows; yi++ {
		for xi := 0; xi+1 < cols; xi++ {
			p00 := index(yi, xi)
			p10 := index(yi, xi+1)
			p01 := index(yi+1, xi)
			p11 := index(yi+1, xi+1)

			triangles = append(triangles, [3]int{p00, p10, p11})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p10, p11}))
			triangles = append(triangles, [3]int{p00, p11, p01})
			mask = append(mask, !triangleFinite(values, [3]int{p00, p11, p01}))
		}
	}

	return Triangulation{
		X:         xPoints,
		Y:         yPoints,
		Triangles: triangles,
		Mask:      mask,
	}, values, true
}

func contourGridCoordsValues(data [][]float64, opts []ContourOptions) ([]float64, []float64, []float64, bool) {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok || rows < 2 || cols < 2 {
		return nil, nil, nil, false
	}

	var opt ContourOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xCoords := resolvedContourCoords(cols, opt.X, opt.XEdges)
	yCoords := resolvedContourCoords(rows, opt.Y, opt.YEdges)
	if len(xCoords) != cols || len(yCoords) != rows {
		return nil, nil, nil, false
	}

	values := make([]float64, 0, rows*cols)
	for yi := 0; yi < rows; yi++ {
		if len(data[yi]) != cols {
			return nil, nil, nil, false
		}
		values = append(values, data[yi]...)
	}
	return xCoords, yCoords, values, true
}

func triangleFinite(values []float64, tri [3]int) bool {
	return isFinite(values[tri[0]]) && isFinite(values[tri[1]]) && isFinite(values[tri[2]])
}

func resolvedContourCoords(size int, coords, edges []float64) []float64 {
	switch {
	case len(coords) == size:
		return append([]float64(nil), coords...)
	case len(edges) == size:
		return append([]float64(nil), edges...)
	case len(edges) == size+1:
		out := make([]float64, size)
		for i := 0; i < size; i++ {
			out[i] = (edges[i] + edges[i+1]) * 0.5
		}
		return out
	default:
		out := make([]float64, size)
		for i := range out {
			out[i] = float64(i)
		}
		return out
	}
}

func contourLevels(values, explicit []float64, levelCount int, filled bool) []float64 {
	if len(explicit) > 0 {
		levels := make([]float64, 0, len(explicit))
		for _, level := range explicit {
			if isFinite(level) {
				levels = append(levels, level)
			}
		}
		sort.Float64s(levels)
		return dedupeFloat64(levels)
	}

	if levelCount <= 0 {
		levelCount = 7
	}
	if filled && levelCount < 2 {
		levelCount = 2
	}

	minValue, maxValue := finiteRange(values)
	if !isFinite(minValue) || !isFinite(maxValue) {
		return nil
	}
	if minValue == maxValue {
		if filled {
			return []float64{minValue, minValue + 1}
		}
		return []float64{minValue}
	}

	levels := contourLocatorLevels(minValue, maxValue, levelCount, filled)
	if len(levels) > 0 {
		return levels
	}

	levels = make([]float64, levelCount)
	step := (maxValue - minValue) / float64(levelCount-1)
	for i := range levels {
		levels[i] = minValue + float64(i)*step
	}
	return levels
}

func contourLocatorLevels(minValue, maxValue float64, levelCount int, filled bool) []float64 {
	// Match matplotlib's _ensure_locator_exists: MaxNLocator(N + 1, min_n_ticks=1).
	// For levels=N the locator is asked for N+1 intervals so the resulting "nice" step
	// is roughly (zmax-zmin)/(N+1) — matching ContourSet._autolev's tick layout.
	levels := (MaxNLocator{
		N:     levelCount + 1,
		Steps: []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10},
	}).Ticks(minValue, maxValue, 0)
	if len(levels) == 0 {
		return nil
	}

	out := levels[:0]
	for _, level := range levels {
		if !isFinite(level) {
			continue
		}
		out = append(out, level)
	}
	return dedupeFloat64(out)
}

func contourPolylines(tri Triangulation, values, levels []float64) ([][]geom.Pt, []float64) {
	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		segments := contourSegmentsForLevel(tri, values, level)
		for _, polyline := range stitchContourSegments(segments) {
			if len(polyline) < 2 {
				continue
			}
			polylines = append(polylines, polyline)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

func contourGridPolylines(x, y []float64, data [][]float64, levels []float64) ([][]geom.Pt, []float64) {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 {
		return nil, nil
	}
	cols := len(data[0])
	if cols < 2 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(data[row]) != cols {
			return nil, nil
		}
	}

	var polylines [][]geom.Pt
	var polylineLevels []float64
	for _, level := range levels {
		var segments [][]geom.Pt
		for row := 0; row+1 < rows; row++ {
			for col := 0; col+1 < cols; col++ {
				cellSegments := contourCellSegmentsForLevel(
					[4]geom.Pt{
						{X: x[col], Y: y[row]},
						{X: x[col+1], Y: y[row]},
						{X: x[col+1], Y: y[row+1]},
						{X: x[col], Y: y[row+1]},
					},
					[4]float64{
						data[row][col],
						data[row][col+1],
						data[row+1][col+1],
						data[row+1][col],
					},
					level,
				)
				segments = append(segments, cellSegments...)
			}
		}
		for _, polyline := range stitchContourSegments(segments) {
			if len(polyline) < 2 {
				continue
			}
			polyline = orientStructuredOpenBoundaryPolyline(polyline, x, y)
			polylines = append(polylines, polyline)
			polylineLevels = append(polylineLevels, level)
		}
	}
	return polylines, polylineLevels
}

type contourBoundarySide uint8

const (
	contourBoundaryNone contourBoundarySide = iota
	contourBoundaryLeft
	contourBoundaryRight
	contourBoundaryBottom
	contourBoundaryTop
)

func orientStructuredOpenBoundaryPolyline(polyline []geom.Pt, x, y []float64) []geom.Pt {
	if len(polyline) < 2 || contourPolylineClosed(polyline) || len(x) == 0 || len(y) == 0 {
		return polyline
	}
	first := structuredBoundarySide(polyline[0], x, y)
	last := structuredBoundarySide(polyline[len(polyline)-1], x, y)
	if first == contourBoundaryNone || last == contourBoundaryNone || first == last {
		return polyline
	}
	if structuredBoundarySideCount(polyline, x, y) != 2 {
		return polyline
	}
	desired, ok := contourpyOpenBoundaryStartSide(first, last)
	if !ok || first == desired {
		return polyline
	}
	if last == desired {
		return reversePoints(polyline)
	}
	return polyline
}

func structuredBoundarySideCount(polyline []geom.Pt, x, y []float64) int {
	sides := map[contourBoundarySide]bool{}
	for _, pt := range polyline {
		side := structuredBoundarySide(pt, x, y)
		if side != contourBoundaryNone {
			sides[side] = true
		}
	}
	return len(sides)
}

func contourpyOpenBoundaryStartSide(a, b contourBoundarySide) (contourBoundarySide, bool) {
	switch {
	case boundarySidePair(a, b, contourBoundaryBottom, contourBoundaryLeft):
		return contourBoundaryBottom, true
	case boundarySidePair(a, b, contourBoundaryBottom, contourBoundaryRight):
		return contourBoundaryRight, true
	case boundarySidePair(a, b, contourBoundaryLeft, contourBoundaryTop):
		return contourBoundaryLeft, true
	case boundarySidePair(a, b, contourBoundaryTop, contourBoundaryRight):
		return contourBoundaryTop, true
	default:
		return contourBoundaryNone, false
	}
}

func boundarySidePair(a, b, c, d contourBoundarySide) bool {
	return (a == c && b == d) || (a == d && b == c)
}

func structuredBoundarySide(pt geom.Pt, x, y []float64) contourBoundarySide {
	minX, maxX := math.Min(x[0], x[len(x)-1]), math.Max(x[0], x[len(x)-1])
	minY, maxY := math.Min(y[0], y[len(y)-1]), math.Max(y[0], y[len(y)-1])
	switch {
	case math.Abs(pt.X-minX) <= 1e-9:
		return contourBoundaryLeft
	case math.Abs(pt.X-maxX) <= 1e-9:
		return contourBoundaryRight
	case math.Abs(pt.Y-minY) <= 1e-9:
		return contourBoundaryBottom
	case math.Abs(pt.Y-maxY) <= 1e-9:
		return contourBoundaryTop
	default:
		return contourBoundaryNone
	}
}

func contourBandPolygons(tri Triangulation, values, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color) {
	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		for triIdx, triangle := range tri.Triangles {
			if tri.masked(triIdx) {
				continue
			}
			polygon := triangleBandPolygon(
				[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
				[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
				low,
				high,
			)
			if len(polygon) < 3 {
				continue
			}
			polygons = append(polygons, polygon)
			colors = append(colors, color)
		}
	}
	return polygons, colors
}

func contourCellSegmentsForLevel(points [4]geom.Pt, values [4]float64, level float64) [][]geom.Pt {
	for _, value := range values {
		if !isFinite(value) {
			return nil
		}
	}

	above := [4]bool{}
	for i, value := range values {
		above[i] = value >= level
	}

	edgePairs := [4][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}
	edgePoints := make([]geom.Pt, 4)
	edgeHit := [4]bool{}
	for edgeIdx, pair := range edgePairs {
		aIdx, bIdx := pair[0], pair[1]
		aValue, bValue := values[aIdx], values[bIdx]
		if aValue == bValue {
			continue
		}
		minValue, maxValue := math.Min(aValue, bValue), math.Max(aValue, bValue)
		if level < minValue || level > maxValue {
			continue
		}
		t := (level - aValue) / (bValue - aValue)
		if t < 0 || t > 1 {
			continue
		}
		edgePoints[edgeIdx] = interpolatePoint(points[aIdx], points[bIdx], t)
		edgeHit[edgeIdx] = true
	}

	if above[0] == above[2] && above[1] == above[3] && above[0] != above[1] {
		order := []int{3, 2, 1, 0}
		if above[0] {
			order = []int{0, 3, 2, 1}
		}
		polyline := make([]geom.Pt, 0, 4)
		for _, edgeIdx := range order {
			if edgeHit[edgeIdx] && !containsPoint(polyline, edgePoints[edgeIdx]) {
				polyline = append(polyline, edgePoints[edgeIdx])
			}
		}
		if len(polyline) >= 2 {
			return [][]geom.Pt{polyline}
		}
	}

	var hits []geom.Pt
	for edgeIdx := range edgeHit {
		if edgeHit[edgeIdx] && !containsPoint(hits, edgePoints[edgeIdx]) {
			hits = append(hits, edgePoints[edgeIdx])
		}
	}
	if len(hits) == 2 {
		return [][]geom.Pt{hits}
	}
	if len(hits) == 4 {
		return [][]geom.Pt{{hits[0], hits[1]}, {hits[2], hits[3]}}
	}
	return nil
}

func contourGridBandPolygons(x, y []float64, data [][]float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64) ([][]geom.Pt, []render.Color) {
	rows := len(data)
	if rows < 2 || len(x) < 2 || len(y) < 2 || len(levels) < 2 {
		return nil, nil
	}
	cols := len(data[0])
	if cols < 2 || len(x) < cols || len(y) < rows {
		return nil, nil
	}
	for row := 1; row < rows; row++ {
		if len(data[row]) != cols {
			return nil, nil
		}
	}

	polygons := [][]geom.Pt{}
	colors := []render.Color{}
	for levelIdx := 0; levelIdx+1 < len(levels); levelIdx++ {
		low := levels[levelIdx]
		high := levels[levelIdx+1]
		color := contourBandColor(low, high, levelIdx, opt, mapping, alpha)
		for row := 0; row+1 < rows; row++ {
			for col := 0; col+1 < cols; col++ {
				cellPolygons := contourCellBandPolygons(
					[4]geom.Pt{
						{X: x[col], Y: y[row]},
						{X: x[col+1], Y: y[row]},
						{X: x[col+1], Y: y[row+1]},
						{X: x[col], Y: y[row+1]},
					},
					[4]float64{
						data[row][col],
						data[row][col+1],
						data[row+1][col+1],
						data[row+1][col],
					},
					low,
					high,
				)
				for _, polygon := range cellPolygons {
					if len(polygon) < 3 {
						continue
					}
					polygons = append(polygons, polygon)
					colors = append(colors, color)
				}
			}
		}
	}
	return polygons, colors
}

func contourCellBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	if polygons := contourSaddleBandPolygons(points, values, low, high); len(polygons) > 0 {
		return polygons
	}
	polygon := contourCellBandPolygon(points, values, low, high)
	if len(polygon) < 3 {
		return nil
	}
	return [][]geom.Pt{polygon}
}

func contourCellBandPolygon(points [4]geom.Pt, values [4]float64, low, high float64) []geom.Pt {
	polygon := []contourVertex{
		{Point: points[0], Value: values[0]},
		{Point: points[1], Value: values[1]},
		{Point: points[2], Value: values[2]},
		{Point: points[3], Value: values[3]},
	}
	polygon = clipContourPolygonMin(polygon, low)
	if len(polygon) < 3 {
		return nil
	}
	polygon = clipContourPolygonMax(polygon, high)
	if len(polygon) < 3 {
		return nil
	}
	out := make([]geom.Pt, len(polygon))
	for i, vertex := range polygon {
		out[i] = vertex.Point
	}
	out = rotateContourPolygonToMatplotlibStart(out)
	if contourPolygonHasConsecutiveDuplicate(out) && !contourPolygonClosed(out) {
		out = append(out, out[0])
	}
	return out
}

func contourSaddleBandPolygons(points [4]geom.Pt, values [4]float64, low, high float64) [][]geom.Pt {
	inBand := [4]bool{}
	for i, value := range values {
		if !isFinite(value) {
			return nil
		}
		inBand[i] = value >= low && value <= high
	}
	if inBand[0] != inBand[2] || inBand[1] != inBand[3] || inBand[0] == inBand[1] {
		return nil
	}

	polygons := [][]geom.Pt{}
	for i, inside := range inBand {
		if !inside {
			continue
		}
		prev := (i + 3) % 4
		next := (i + 1) % 4
		if !contourBandOutsideSameSide(values[prev], values[next], low, high) {
			return nil
		}
		nextPoint, nextOK := contourBandBoundaryIntersection(points[i], values[i], points[next], values[next], low, high)
		prevPoint, prevOK := contourBandBoundaryIntersection(points[i], values[i], points[prev], values[prev], low, high)
		if !nextOK || !prevOK {
			continue
		}
		polygon := rotateContourPolygonToMatplotlibStart([]geom.Pt{points[i], nextPoint, prevPoint})
		if len(polygon) > 0 {
			polygon = append(polygon, polygon[0])
		}
		polygons = append(polygons, polygon)
	}
	return polygons
}

func contourBandOutsideSameSide(a, b, low, high float64) bool {
	return (a < low && b < low) || (a > high && b > high)
}

func contourBandBoundaryIntersection(insidePoint geom.Pt, insideValue float64, outsidePoint geom.Pt, outsideValue float64, low, high float64) (geom.Pt, bool) {
	if !isFinite(insideValue) || !isFinite(outsideValue) || insideValue == outsideValue {
		return geom.Pt{}, false
	}
	threshold := low
	if outsideValue > high {
		threshold = high
	}
	if outsideValue >= low && outsideValue <= high {
		return geom.Pt{}, false
	}
	t := (threshold - insideValue) / (outsideValue - insideValue)
	if t < 0 || t > 1 {
		return geom.Pt{}, false
	}
	return interpolatePoint(insidePoint, outsidePoint, t), true
}

func contourSegmentsForLevel(tri Triangulation, values []float64, level float64) [][]geom.Pt {
	segments := make([][]geom.Pt, 0, len(tri.Triangles))
	for triIdx, triangle := range tri.Triangles {
		if tri.masked(triIdx) {
			continue
		}
		segment, ok := triangleContourSegment(
			[3]geom.Pt{tri.point(triangle[0]), tri.point(triangle[1]), tri.point(triangle[2])},
			[3]float64{values[triangle[0]], values[triangle[1]], values[triangle[2]]},
			level,
		)
		if ok {
			segments = append(segments, segment)
		}
	}
	return segments
}

func triangleContourSegment(points [3]geom.Pt, values [3]float64, level float64) ([]geom.Pt, bool) {
	intersections := []geom.Pt{}
	edges := [][2]int{{0, 1}, {1, 2}, {2, 0}}
	for _, edge := range edges {
		aIdx, bIdx := edge[0], edge[1]
		aValue := values[aIdx]
		bValue := values[bIdx]
		if !isFinite(aValue) || !isFinite(bValue) {
			return nil, false
		}
		if aValue == bValue {
			continue
		}
		minValue := math.Min(aValue, bValue)
		maxValue := math.Max(aValue, bValue)
		if level < minValue || level > maxValue {
			continue
		}
		t := (level - aValue) / (bValue - aValue)
		if t < 0 || t > 1 {
			continue
		}
		point := interpolatePoint(points[aIdx], points[bIdx], t)
		if !containsPoint(intersections, point) {
			intersections = append(intersections, point)
		}
	}
	if len(intersections) != 2 {
		return nil, false
	}
	return intersections, true
}

type contourVertex struct {
	Point geom.Pt
	Value float64
}

func triangleBandPolygon(points [3]geom.Pt, values [3]float64, low, high float64) []geom.Pt {
	polygon := []contourVertex{
		{Point: points[0], Value: values[0]},
		{Point: points[1], Value: values[1]},
		{Point: points[2], Value: values[2]},
	}
	polygon = clipContourPolygonMin(polygon, low)
	if len(polygon) < 3 {
		return nil
	}
	polygon = clipContourPolygonMax(polygon, high)
	if len(polygon) < 3 {
		return nil
	}
	out := make([]geom.Pt, len(polygon))
	for i, vertex := range polygon {
		out[i] = vertex.Point
	}
	return out
}

func rotateContourPolygonToMatplotlibStart(points []geom.Pt) []geom.Pt {
	if len(points) == 0 {
		return points
	}
	start := 0
	for i := 1; i < len(points); i++ {
		if points[i].X > points[start].X ||
			(points[i].X == points[start].X && points[i].Y < points[start].Y) {
			start = i
		}
	}
	if start == 0 {
		return points
	}
	out := make([]geom.Pt, 0, len(points))
	out = append(out, points[start:]...)
	out = append(out, points[:start]...)
	return out
}

func contourPolygonHasConsecutiveDuplicate(points []geom.Pt) bool {
	for i := 1; i < len(points); i++ {
		if points[i] == points[i-1] {
			return true
		}
	}
	return false
}

func contourPolygonClosed(points []geom.Pt) bool {
	return len(points) > 1 && points[0] == points[len(points)-1]
}

func clipContourPolygonMin(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value >= threshold
	}, threshold)
}

func clipContourPolygonMax(polygon []contourVertex, threshold float64) []contourVertex {
	return clipContourPolygon(polygon, func(value float64) bool {
		return value <= threshold
	}, threshold)
}

func clipContourPolygon(polygon []contourVertex, inside func(float64) bool, threshold float64) []contourVertex {
	if len(polygon) == 0 {
		return nil
	}
	out := make([]contourVertex, 0, len(polygon)+2)
	prev := polygon[len(polygon)-1]
	prevInside := inside(prev.Value)
	for _, curr := range polygon {
		currInside := inside(curr.Value)
		if currInside != prevInside && curr.Value != prev.Value {
			t := (threshold - prev.Value) / (curr.Value - prev.Value)
			out = append(out, contourVertex{
				Point: interpolatePoint(prev.Point, curr.Point, t),
				Value: threshold,
			})
		}
		if currInside {
			out = append(out, curr)
		}
		prev = curr
		prevInside = currInside
	}
	return out
}

func contourBandColor(low, high float64, idx int, opt ContourOptions, mapping ScalarMapInfo, alpha float64) render.Color {
	if len(opt.Colors) > 0 {
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	return mapping.Color((low+high)*0.5, alpha)
}

func contourLineColor(level float64, levels []float64, opt ContourOptions, mapping ScalarMapInfo, alpha float64, fallback render.Color) render.Color {
	if opt.Color != nil {
		color := *opt.Color
		color.A *= alpha
		return color
	}
	if len(opt.Colors) > 0 {
		idx := indexOfLevel(levels, level)
		color := opt.Colors[idx%len(opt.Colors)]
		color.A *= alpha
		return color
	}
	if opt.Colormap != nil {
		return mapping.Color(level, alpha)
	}
	fallback.A *= alpha
	return fallback
}

func contourLabels(polylines [][]geom.Pt, levels []float64, colors []render.Color, formatter Formatter) []contourLabel {
	type candidate struct {
		polyline []geom.Pt
		color    render.Color
	}
	best := map[float64]candidate{}
	bestLen := map[float64]float64{}
	for i, polyline := range polylines {
		length := polylineLength(polyline)
		level := levels[i]
		if length <= bestLen[level] {
			continue
		}
		bestLen[level] = length
		best[level] = candidate{polyline: polyline, color: colors[i]}
	}

	labels := make([]contourLabel, 0, len(best))
	for _, level := range dedupeFloat64(levels) {
		candidate, ok := best[level]
		if !ok {
			continue
		}
		position, angle := polylineLabelPlacement(candidate.polyline)
		labels = append(labels, contourLabel{
			Text:     formatter.Format(level),
			Position: position,
			Angle:    normalizeLabelAngle(angle),
			Color:    candidate.color,
			Level:    level,
		})
	}
	return labels
}

func (c *ContourSet) clabelLineIndices(levels []float64) ([]int, bool) {
	if c == nil || len(c.lineLevels) == 0 {
		return nil, false
	}
	if len(levels) == 0 {
		indices := make([]int, len(c.lineLevels))
		for i := range c.lineLevels {
			indices[i] = i
		}
		return indices, true
	}
	for _, level := range levels {
		if !contourLevelAvailable(c.Levels, level) {
			return nil, false
		}
	}
	indices := []int{}
	for i, level := range c.lineLevels {
		if contourLevelAvailable(levels, level) {
			indices = append(indices, i)
		}
	}
	return indices, true
}

func (c *ContourSet) clabelPlaceAutomatic(indices []int, opt ClabelOptions) []contourLabel {
	polylines := make([][]geom.Pt, 0, len(indices))
	levels := make([]float64, 0, len(indices))
	colors := make([]render.Color, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || c.Lines == nil || idx >= len(c.Lines.Segments) || idx >= len(c.lineLevels) {
			continue
		}
		level := c.lineLevels[idx]
		polylines = append(polylines, c.Lines.Segments[idx])
		levels = append(levels, level)
		colors = append(colors, c.clabelColor(idx, level, len(levels)-1, opt))
	}
	return contourLabels(polylines, levels, colors, c.LabelFormatter)
}

func (c *ContourSet) clabelPlaceManual(indices []int, opt ClabelOptions) []contourLabel {
	labels := make([]contourLabel, 0, len(opt.ManualPositions))
	for _, point := range opt.ManualPositions {
		idx, projection, angle, ok := c.nearestContourLabelPoint(indices, point)
		if !ok {
			continue
		}
		level := c.lineLevels[idx]
		labels = append(labels, contourLabel{
			Text:     c.LabelFormatter.Format(level),
			Position: projection,
			Angle:    angle,
			Color:    c.clabelColor(idx, level, len(labels), opt),
			Level:    level,
		})
	}
	return labels
}

func (c *ContourSet) nearestContourLabelPoint(indices []int, point geom.Pt) (int, geom.Pt, float64, bool) {
	bestIdx := -1
	bestPoint := geom.Pt{}
	bestAngle := 0.0
	bestDist := math.Inf(1)
	for _, idx := range indices {
		if c == nil || c.Lines == nil || idx < 0 || idx >= len(c.Lines.Segments) {
			continue
		}
		segment := c.Lines.Segments[idx]
		for i := 1; i < len(segment); i++ {
			projection, dist := projectPointToSegment(point, segment[i-1], segment[i])
			if dist >= bestDist {
				continue
			}
			bestDist = dist
			bestIdx = idx
			bestPoint = projection
			bestAngle = normalizeLabelAngle(math.Atan2(segment[i].Y-segment[i-1].Y, segment[i].X-segment[i-1].X))
		}
	}
	return bestIdx, bestPoint, bestAngle, bestIdx >= 0
}

func (c *ContourSet) clabelColor(segmentIndex int, level float64, labelIndex int, opt ClabelOptions) render.Color {
	if opt.Color != nil {
		return *opt.Color
	}
	if len(opt.Colors) > 0 {
		return opt.Colors[labelIndex%len(opt.Colors)]
	}
	if c != nil && c.Lines != nil {
		return colorAt(c.Lines.Color, c.Lines.Colors, segmentIndex)
	}
	return render.Color{}
}

func publicContourLabels(labels []contourLabel) []ContourLabel {
	out := make([]ContourLabel, len(labels))
	for i, label := range labels {
		out[i] = ContourLabel{
			Text:     label.Text,
			Level:    label.Level,
			Position: label.Position,
			Angle:    label.Angle,
			Color:    label.Color,
		}
	}
	return out
}

func uniqueLevelsForIndices(levels []float64, indices []int) []float64 {
	out := []float64{}
	for _, idx := range indices {
		if idx < 0 || idx >= len(levels) || contourLevelAvailable(out, levels[idx]) {
			continue
		}
		out = append(out, levels[idx])
	}
	return out
}

func contourLevelAvailable(levels []float64, level float64) bool {
	for _, existing := range levels {
		if math.Abs(existing-level) <= 1e-12 {
			return true
		}
	}
	return false
}

func firstFormatter(primary, fallback Formatter) Formatter {
	if primary != nil {
		return primary
	}
	return fallback
}

func projectPointToSegment(point, a, b geom.Pt) (geom.Pt, float64) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	den := dx*dx + dy*dy
	if den <= 0 {
		return a, pointDistanceSquared(point, a)
	}
	t := ((point.X-a.X)*dx + (point.Y-a.Y)*dy) / den
	t = clampFloat(t, 0, 1)
	projection := geom.Pt{X: a.X + t*dx, Y: a.Y + t*dy}
	return projection, pointDistanceSquared(point, projection)
}

func contourInlineLabelSegments(lines *LineCollection, levels []float64, formatter Formatter, fontSize float64, r render.Renderer, ctx *DrawContext) ([][]geom.Pt, []render.Color, []float64, []contourLabel) {
	return contourInlineLabelSegmentsForLevels(lines, levels, nil, formatter, fontSize, 5, r, ctx)
}

func contourInlineLabelSegmentsForLevels(lines *LineCollection, levels, selectedLevels []float64, formatter Formatter, fontSize, inlineSpacing float64, r render.Renderer, ctx *DrawContext) ([][]geom.Pt, []render.Color, []float64, []contourLabel) {
	segments := make([][]geom.Pt, 0, len(lines.Segments))
	colors := make([]render.Color, 0, len(lines.Segments))
	widths := make([]float64, 0, len(lines.Segments))
	labels := []contourLabel{}
	placed := []geom.Pt{}

	for i, segment := range lines.Segments {
		color := colorAt(lines.Color, lines.Colors, i)
		width := widthAt(lines.LineWidth, lines.LineWidths, i)
		appendSegment := func(part []geom.Pt) {
			if len(part) < 2 {
				return
			}
			segments = append(segments, part)
			colors = append(colors, color)
			widths = append(widths, width)
		}

		if len(segment) < 2 || i >= len(levels) {
			appendSegment(segment)
			continue
		}
		if len(selectedLevels) > 0 && !contourLevelAvailable(selectedLevels, levels[i]) {
			appendSegment(segment)
			continue
		}

		text := formatter.Format(levels[i])
		if displayTextIsEmpty(text) {
			appendSegment(segment)
			continue
		}
		labelWidth := contourLabelWidth(text, fontSize, r, ctx)
		screen := contourDisplayPolyline(segment, ctx)
		if !contourPrintLabel(screen, labelWidth) {
			appendSegment(segment)
			continue
		}

		screenPos, labelIdx := contourLocateLabel(screen, labelWidth, placed)
		if labelIdx < 0 || labelIdx >= len(segment) {
			appendSegment(segment)
			continue
		}

		angle, parts := splitContourPolylineForLabel(segment, screen, labelIdx, labelWidth, inlineSpacing)
		if len(parts) == 0 {
			appendSegment(segment)
			continue
		}
		for _, part := range parts {
			appendSegment(part)
		}
		placed = append(placed, screenPos)
		labels = append(labels, contourLabel{
			Text:     text,
			Position: segment[labelIdx],
			Angle:    angle,
			Color:    color,
			Level:    levels[i],
		})
	}

	return segments, colors, widths, labels
}

func contourLabelWidth(text string, fontSize float64, r render.Renderer, ctx *DrawContext) float64 {
	fontKey := ""
	useTeX := false
	if ctx != nil {
		fontKey = ctx.RC.FontKey
		useTeX = ctx.RC.UseTeX
	}
	layout := measureSingleLineTextLayout(r, text, fontSize, fontKey, useTeX)
	if layout.Width > 0 {
		return layout.Width
	}
	return math.Max(fontSize, 1)
}

func contourDisplayPolyline(polyline []geom.Pt, ctx *DrawContext) []geom.Pt {
	out := make([]geom.Pt, len(polyline))
	for i, pt := range polyline {
		out[i] = ctx.DataToPixel.Apply(pt)
	}
	return out
}

func contourPrintLabel(line []geom.Pt, labelWidth float64) bool {
	if len(line) == 0 || labelWidth <= 0 {
		return false
	}
	if float64(len(line)) > 10*labelWidth {
		return true
	}
	minX, maxX := line[0].X, line[0].X
	minY, maxY := line[0].Y, line[0].Y
	for _, pt := range line[1:] {
		minX = math.Min(minX, pt.X)
		maxX = math.Max(maxX, pt.X)
		minY = math.Min(minY, pt.Y)
		maxY = math.Max(maxY, pt.Y)
	}
	return maxX-minX > 1.2*labelWidth || maxY-minY > 1.2*labelWidth
}

func contourLocateLabel(line []geom.Pt, labelWidth float64, placed []geom.Pt) (geom.Pt, int) {
	if len(line) == 0 {
		return geom.Pt{}, -1
	}
	ctrSize := len(line)
	nBlocks := 1
	if labelWidth > 1 {
		nBlocks = int(math.Ceil(float64(ctrSize) / labelWidth))
		if nBlocks < 1 {
			nBlocks = 1
		}
	}
	blockSize := ctrSize
	if nBlocks != 1 {
		blockSize = int(labelWidth)
		if blockSize < 1 {
			blockSize = 1
		}
	}

	type candidate struct {
		block    int
		distance float64
	}
	candidates := make([]candidate, nBlocks)
	for block := 0; block < nBlocks; block++ {
		first := contourResizedPoint(line, block, blockSize, 0)
		last := contourResizedPoint(line, block, blockSize, blockSize-1)
		length := math.Hypot(last.X-first.X, last.Y-first.Y)
		distance := math.Inf(1)
		if length > 0 {
			distance = 0
			for j := 0; j < blockSize; j++ {
				pt := contourResizedPoint(line, block, blockSize, j)
				cross := (first.Y-pt.Y)*(last.X-first.X) - (first.X-pt.X)*(last.Y-first.Y)
				distance += math.Abs(cross) / length
			}
		}
		candidates[block] = candidate{block: block, distance: distance}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	halfBlock := blockSize / 2
	chosen := candidates[0].block
	for _, candidate := range candidates {
		idx := (candidate.block*blockSize + halfBlock) % ctrSize
		pt := line[idx]
		if !contourLabelTooClose(pt, labelWidth, placed) {
			chosen = candidate.block
			break
		}
	}
	idx := (chosen*blockSize + halfBlock) % ctrSize
	return line[idx], idx
}

func contourResizedPoint(line []geom.Pt, block, blockSize, offset int) geom.Pt {
	return line[(block*blockSize+offset)%len(line)]
}

func contourLabelTooClose(pt geom.Pt, labelWidth float64, placed []geom.Pt) bool {
	threshold := 1.2 * labelWidth
	threshold *= threshold
	for _, existing := range placed {
		dx := pt.X - existing.X
		dy := pt.Y - existing.Y
		if dx*dx+dy*dy < threshold {
			return true
		}
	}
	return false
}

func splitContourPolylineForLabel(data, screen []geom.Pt, labelIdx int, labelWidth, spacing float64) (float64, [][]geom.Pt) {
	if len(data) < 2 || len(data) != len(screen) || labelIdx < 0 || labelIdx >= len(data) {
		return 0, nil
	}
	cpls := contourCumulativeDisplayLengths(screen)
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return 0, nil
	}
	if contourPolylineClosed(screen) {
		return splitClosedContourPolylineForLabel(data, screen, cpls, labelIdx, labelWidth, spacing)
	}

	center := cpls[labelIdx]
	angleStart := clampFloat(center-labelWidth/2, 0, total)
	angleEnd := clampFloat(center+labelWidth/2, 0, total)
	p0, ok0 := contourInterpolateAtCPL(screen, cpls, angleStart)
	p1, ok1 := contourInterpolateAtCPL(screen, cpls, angleEnd)
	angle := 0.0
	if ok0 && ok1 {
		// Matplotlib contour labels use screen-space angles, but its display
		// coordinate system has positive Y upward. Go renderers use top-left
		// pixels, so flip Y before normalizing the label angle.
		angle = normalizeLabelAngle(math.Atan2(-(p1.Y - p0.Y), p1.X-p0.X))
	}

	gapStart := center - labelWidth/2 - spacing
	gapEnd := center + labelWidth/2 + spacing
	parts := [][]geom.Pt{}
	if gapStart > 0 {
		before := contourPolylineBeforeCPL(data, cpls, gapStart)
		if len(before) >= 2 {
			parts = append(parts, before)
		}
	}
	if gapEnd < total {
		after := contourPolylineAfterCPL(data, cpls, gapEnd)
		if len(after) >= 2 {
			parts = append(parts, after)
		}
	}
	return angle, parts
}

func splitClosedContourPolylineForLabel(data, screen []geom.Pt, cpls []float64, labelIdx int, labelWidth, spacing float64) (float64, [][]geom.Pt) {
	total := cpls[len(cpls)-1]
	gap := labelWidth + 2*spacing
	if total <= 0 || gap >= total {
		return 0, nil
	}

	center := cpls[labelIdx]
	p0, ok0 := contourInterpolateAtClosedCPL(screen, cpls, center-labelWidth/2)
	p1, ok1 := contourInterpolateAtClosedCPL(screen, cpls, center+labelWidth/2)
	angle := 0.0
	if ok0 && ok1 {
		angle = normalizeLabelAngle(math.Atan2(-(p1.Y - p0.Y), p1.X-p0.X))
	}

	gapStart := center - labelWidth/2 - spacing
	gapEnd := center + labelWidth/2 + spacing
	part := contourClosedPolylineComplement(data, cpls, gapEnd, gapStart+total)
	if len(part) < 2 {
		return angle, nil
	}
	return angle, [][]geom.Pt{part}
}

func contourPolylineClosed(points []geom.Pt) bool {
	return len(points) > 2 && sameContourPoint(points[0], points[len(points)-1])
}

func contourClosedPolylineComplement(data []geom.Pt, cpls []float64, start, end float64) []geom.Pt {
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return nil
	}
	for start < 0 {
		start += total
		end += total
	}
	for start >= total {
		start -= total
		end -= total
	}
	if end <= start {
		end += total
	}
	if end-start <= 0 || end-start >= total {
		return nil
	}

	startPt, ok := contourInterpolateAtClosedCPL(data, cpls, start)
	if !ok {
		return nil
	}
	out := []geom.Pt{startPt}

	type closedVertex struct {
		cpl float64
		pt  geom.Pt
	}
	vertices := []closedVertex{}
	base := math.Floor(start/total) - 1
	for copyIdx := 0; copyIdx < 4; copyIdx++ {
		offset := (base + float64(copyIdx)) * total
		for i := 0; i < len(data); i++ {
			vertexCPL := cpls[i] + offset
			if vertexCPL <= start || vertexCPL >= end {
				continue
			}
			vertices = append(vertices, closedVertex{cpl: vertexCPL, pt: data[i]})
		}
	}
	sort.SliceStable(vertices, func(i, j int) bool {
		return vertices[i].cpl < vertices[j].cpl
	})
	for _, vertex := range vertices {
		out = appendContourPoint(out, vertex.pt)
	}

	endPt, ok := contourInterpolateAtClosedCPL(data, cpls, end)
	if !ok {
		return nil
	}
	out = appendContourPoint(out, endPt)
	return out
}

func contourInterpolateAtClosedCPL(points []geom.Pt, cpls []float64, target float64) (geom.Pt, bool) {
	total := cpls[len(cpls)-1]
	if total <= 0 {
		return geom.Pt{}, false
	}
	target = math.Mod(target, total)
	if target < 0 {
		target += total
	}
	return contourInterpolateAtCPL(points, cpls, target)
}

func contourRotatedTextAnchor(center geom.Pt, layout singleLineTextLayout, angle float64) geom.Pt {
	height := layout.Height
	if layout.HaveInkBounds && layout.InkBounds.H > 0 {
		height = layout.InkBounds.H
	}
	if height <= 0 {
		return center
	}

	halfHeight := height * 0.5
	return geom.Pt{
		X: center.X + halfHeight*math.Sin(angle),
		Y: center.Y + halfHeight*math.Cos(angle),
	}
}

func contourCumulativeDisplayLengths(points []geom.Pt) []float64 {
	cpls := make([]float64, len(points))
	for i := 1; i < len(points); i++ {
		cpls[i] = cpls[i-1] + math.Hypot(points[i].X-points[i-1].X, points[i].Y-points[i-1].Y)
	}
	return cpls
}

func contourPolylineBeforeCPL(data []geom.Pt, cpls []float64, target float64) []geom.Pt {
	out := []geom.Pt{}
	for i, pt := range data {
		if cpls[i] <= target {
			out = append(out, pt)
		}
	}
	if pt, ok := contourInterpolateAtCPL(data, cpls, target); ok {
		out = appendContourPoint(out, pt)
	}
	return out
}

func contourPolylineAfterCPL(data []geom.Pt, cpls []float64, target float64) []geom.Pt {
	out := []geom.Pt{}
	if pt, ok := contourInterpolateAtCPL(data, cpls, target); ok {
		out = append(out, pt)
	}
	for i, pt := range data {
		if cpls[i] >= target {
			out = appendContourPoint(out, pt)
		}
	}
	return out
}

func contourInterpolateAtCPL(points []geom.Pt, cpls []float64, target float64) (geom.Pt, bool) {
	if len(points) == 0 || len(points) != len(cpls) {
		return geom.Pt{}, false
	}
	if target <= cpls[0] {
		return points[0], true
	}
	last := len(cpls) - 1
	if target >= cpls[last] {
		return points[last], true
	}
	for i := 1; i < len(cpls); i++ {
		if cpls[i] < target {
			continue
		}
		if cpls[i] == cpls[i-1] {
			return points[i], true
		}
		t := (target - cpls[i-1]) / (cpls[i] - cpls[i-1])
		return interpolatePoint(points[i-1], points[i], t), true
	}
	return points[last], true
}

func appendContourPoint(points []geom.Pt, point geom.Pt) []geom.Pt {
	if len(points) > 0 && sameContourPoint(points[len(points)-1], point) {
		return points
	}
	return append(points, point)
}

func contourFormatter(formatter Formatter) Formatter {
	if formatter != nil {
		return formatter
	}
	return ScalarFormatter{Prec: 3}
}

func dedupeFloat64(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := []float64{values[0]}
	for _, value := range values[1:] {
		if math.Abs(value-out[len(out)-1]) <= 1e-12 {
			continue
		}
		out = append(out, value)
	}
	return out
}

func containsPoint(points []geom.Pt, point geom.Pt) bool {
	for _, existing := range points {
		if sameContourPoint(existing, point) {
			return true
		}
	}
	return false
}

func sameContourPoint(a, b geom.Pt) bool {
	return math.Abs(a.X-b.X) <= 1e-9 && math.Abs(a.Y-b.Y) <= 1e-9
}

func interpolatePoint(a, b geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func stitchContourSegments(segments [][]geom.Pt) [][]geom.Pt {
	remaining := make([][]geom.Pt, 0, len(segments))
	for _, segment := range segments {
		if len(segment) >= 2 {
			remaining = append(remaining, append([]geom.Pt(nil), segment...))
		}
	}
	out := [][]geom.Pt{}
	for len(remaining) > 0 {
		polyline := append([]geom.Pt(nil), remaining[0]...)
		remaining = remaining[1:]
		progress := true
		for progress {
			progress = false
			for i := 0; i < len(remaining); i++ {
				segment := remaining[i]
				switch {
				case sameContourPoint(polyline[len(polyline)-1], segment[0]):
					polyline = append(polyline, segment[1:]...)
				case sameContourPoint(polyline[len(polyline)-1], segment[len(segment)-1]):
					polyline = append(polyline, reversePoints(segment[:len(segment)-1])...)
				case sameContourPoint(polyline[0], segment[len(segment)-1]):
					polyline = append(segment[:len(segment)-1], polyline...)
				case sameContourPoint(polyline[0], segment[0]):
					polyline = append(reversePoints(segment[1:]), polyline...)
				default:
					continue
				}
				remaining = append(remaining[:i], remaining[i+1:]...)
				progress = true
				break
			}
		}
		out = append(out, rotateClosedContourPolylineToMatplotlibStart(polyline))
	}
	return out
}

func rotateClosedContourPolylineToMatplotlibStart(polyline []geom.Pt) []geom.Pt {
	if !contourPolylineClosed(polyline) {
		return polyline
	}
	body := polyline[:len(polyline)-1]
	if len(body) == 0 {
		return polyline
	}
	start := 0
	for i := 1; i < len(body); i++ {
		if body[i].Y < body[start].Y || (body[i].Y == body[start].Y && body[i].X < body[start].X) {
			start = i
		}
	}
	out := make([]geom.Pt, 0, len(polyline))
	out = append(out, body[start:]...)
	out = append(out, body[:start]...)
	out = append(out, out[0])
	return out
}

func reversePoints(points []geom.Pt) []geom.Pt {
	out := append([]geom.Pt(nil), points...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func polylineLength(polyline []geom.Pt) float64 {
	total := 0.0
	for i := 1; i < len(polyline); i++ {
		total += math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
	}
	return total
}

func polylineLabelPlacement(polyline []geom.Pt) (geom.Pt, float64) {
	total := polylineLength(polyline)
	if total == 0 || len(polyline) < 2 {
		if len(polyline) == 0 {
			return geom.Pt{}, 0
		}
		return polyline[0], 0
	}

	target := total * 0.5
	run := 0.0
	for i := 1; i < len(polyline); i++ {
		segLen := math.Hypot(polyline[i].X-polyline[i-1].X, polyline[i].Y-polyline[i-1].Y)
		if run+segLen >= target {
			t := (target - run) / segLen
			point := interpolatePoint(polyline[i-1], polyline[i], t)
			return point, math.Atan2(polyline[i].Y-polyline[i-1].Y, polyline[i].X-polyline[i-1].X)
		}
		run += segLen
	}

	last := polyline[len(polyline)-1]
	prev := polyline[len(polyline)-2]
	return last, math.Atan2(last.Y-prev.Y, last.X-prev.X)
}

func normalizeLabelAngle(angle float64) float64 {
	for angle > math.Pi/2 {
		angle -= math.Pi
	}
	for angle < -math.Pi/2 {
		angle += math.Pi
	}
	return angle
}

func indexOfLevel(levels []float64, level float64) int {
	for i, candidate := range levels {
		if math.Abs(candidate-level) <= 1e-12 {
			return i
		}
	}
	return 0
}

func valueOrDefaultFloat(value *float64, fallback float64) float64 {
	if value == nil {
		return fallback
	}
	return *value
}
