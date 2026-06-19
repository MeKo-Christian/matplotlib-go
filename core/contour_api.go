package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
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
