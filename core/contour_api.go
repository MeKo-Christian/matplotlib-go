package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

// ContourOptions configures contour, contourf, and tricontour rendering.
type ContourOptions struct {
	X              []float64
	Y              []float64
	XEdges         []float64
	YEdges         []float64
	Levels         []float64
	LevelCount     int
	Colormap       optional.Value[string]
	Norm           ScalarNormalizer
	Colors         []render.Color
	Color          optional.Value[render.Color]
	LineWidth      optional.Value[float64]
	Alpha          optional.Value[float64]
	LabelLines     bool
	LabelFormatter ticker.Formatter
	LabelFontSize  optional.Value[float64]
	LabelColor     optional.Value[render.Color]
	Label          string
	// Algorithm selects the structured contour generator semantics:
	// "mpl2005", "mpl2014", "serial", or "threaded". Empty uses
	// rcParams["contour.algorithm"]. It does not apply to triangular contours.
	Algorithm string
	// CornerMask controls whether a structured cell with exactly one
	// masked/non-finite corner retains its valid triangular portion. Unset uses
	// rcParams["contour.corner_mask"], except mpl2005 defaults it to false.
	// It does not apply to triangular contours.
	CornerMask optional.Value[bool]

	// LineStyles assigns per-level stroke styles for contour lines
	// ("solid", "dashed", "--", "dashdot", "-.", "dotted", ":"). The list is
	// cycled (ceil-padded) and truncated to the level count; a single entry
	// applies to every level. Has no effect on contourf.
	LineStyles []string
	// NegativeLineStyles overrides the style applied to negative levels of a
	// monochrome contour when LineStyles is unset (default "dashed", matching
	// rcParams["contour.negative_linestyle"]).
	NegativeLineStyles optional.Value[string]
	// Extend extends the filled (contourf) color mapping past the level range:
	// "neither" (default), "min", "max", or "both". The extra bands use the
	// colormap's under/over colors.
	Extend ColorbarExtend
	// Hatches assigns hatch patterns to contourf bands, cycled per band
	// (Matplotlib's contourf hatches). Empty disables hatching.
	Hatches []string
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
	Formatter       ticker.Formatter
	FontSize        optional.Value[float64]
	Color           optional.Value[render.Color]
	Colors          []render.Color
	Inline          optional.Value[bool]
	InlineSpacing   float64
	ManualPositions []geom.Pt

	// FormatString applies %-style formatting to each level (e.g. "%1.2f"),
	// matching a Matplotlib clabel fmt string.
	FormatString string
	// FormatDict maps a level to its label text, matching a Matplotlib clabel
	// fmt dict; missing levels fall back to "%1.3f".
	FormatDict map[float64]string
	// RightSideUp keeps inline labels upright by constraining their rotation to
	// [-90°, 90°]. Defaults to true when nil.
	RightSideUp optional.Value[bool]
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
	Algorithm      string
	CornerMask     bool
	Lines          *LineCollection
	Fills          *PolyCollection
	LabelFormatter ticker.Formatter
	LabelFontSize  float64
	LabelColor     render.Color
	labels         []contourLabel
	lineLevels     []float64
	labelLevels    []float64
	labelInlineGap float64
	inlineLabels   bool
	rightSideUp    bool
	z              float64
}

// Clabel delegates contour labeling to the provided contour set, matching
// Matplotlib's Axes.clabel call shape.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Clabel(cs *ContourSet, opt ClabelOptions) []ContourLabel {
	if cs == nil {
		return nil
	}
	return cs.Clabel(opt)
}

// Clabel adds labels to a contour set and returns metadata for the labels that
// were placed.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (c *ContourSet) Clabel(opt ClabelOptions) []ContourLabel {
	if c == nil || c.Lines == nil || len(c.Lines.Segments) == 0 {
		return nil
	}
	indices, ok := c.clabelLineIndices(opt.Levels)
	if !ok || len(indices) == 0 {
		return nil
	}

	c.LabelFormatter = newContourLabelFormatter(&opt, c.LabelFormatter)
	c.rightSideUp = specialtyBool(opt.RightSideUp, true)
	if v, ok := opt.FontSize.Get(); ok && v > 0 {
		c.LabelFontSize = opt.FontSize.OrZero()
	} else if c.LabelFontSize <= 0 {
		c.LabelFontSize = 10
	}
	if v, ok := opt.Color.Get(); ok {
		c.LabelColor = v
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
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Contour(data [][]float64, opt ContourOptions) *ContourSet {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opt)
	if !ok {
		return nil
	}

	algorithm, cornerMask, ok := a.resolveStructuredContourOptions(&opt)
	if !ok {
		return nil
	}
	levels := contourLevels(values, opt.Levels, opt.LevelCount, false)
	if len(levels) == 0 {
		return nil
	}

	polylines, polylineLevels := contourGridPolylinesCornerMask(xCoords, yCoords, data, levels, cornerMask)
	if len(polylines) == 0 {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	lineWidth := opt.LineWidth.OrZero()
	colorFallback := a.NextColor()
	cmapName := ""
	if name, ok := opt.Colormap.Get(); ok {
		cmapName = name
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
		Algorithm:      algorithm,
		CornerMask:     cornerMask,
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  opt.LabelFontSize.Or(10),
		rightSideUp:    true,
		z:              defaultLineZ,
	}
	if color, ok := opt.LabelColor.Get(); ok {
		set.LabelColor = color
	}
	colors := make([]render.Color, len(polylines))
	for i, level := range polylineLevels {
		colors[i] = contourLineColor(level, levels, opt, mapping, alpha, colorFallback)
	}
	styles := resolveContourLineStyles(levels, opt, contourMonochrome(opt))
	resolvedRC := a.resolvedRC()
	dashPatterns := contourLineDashPatterns(polylineLevels, levels, styles, lineWidth, &resolvedRC.Lines)
	set.Lines = &LineCollection{
		Collection: Collection{
			Coords: Coords(CoordData),
			Label:  opt.Label,
			Alpha:  1,
		},
		Segments:     polylines,
		Colors:       colors,
		LineWidth:    lineWidth,
		DashPatterns: dashPatterns,
		LineJoin:     render.JoinRound,
		LineCap:      render.CapButt,
	}
	set.lineLevels = append([]float64(nil), polylineLevels...)
	if opt.LabelLines {
		set.inlineLabels = true
		set.labels = contourLabels(polylines, polylineLevels, colors, set.LabelFormatter, set.rightSideUp)
	}
	a.Add(set)
	return set
}

// Contourf draws filled contour bands over a rectilinear scalar grid.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) Contourf(data [][]float64, opt ContourOptions) *ContourSet {
	xCoords, yCoords, values, ok := contourGridCoordsValues(data, opt)
	if !ok {
		return nil
	}
	algorithm, cornerMask, ok := a.resolveStructuredContourOptions(&opt)
	if !ok {
		return nil
	}
	levels := contourLevels(values, opt.Levels, opt.LevelCount, true)
	if len(levels) < 2 {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	cmapName := ""
	if name, ok := opt.Colormap.Get(); ok {
		cmapName = name
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

	geomLevels := contourExtendedLevels(levels, opt.Extend)
	rawPolygons, rawColors, rawHatches, bands := contourGridBandPolygonsBands(xCoords, yCoords, data, geomLevels, opt, mapping, alpha)
	if len(rawPolygons) == 0 {
		return nil
	}
	fillPaths, polygons, faceColors, hatches := contourFillGeometry(rawPolygons, rawColors, rawHatches, bands)
	cmap := ""
	vmin := 0.0
	vmax := 0.0
	if !opt.Color.IsSet() && len(opt.Colors) == 0 {
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
		Algorithm:      algorithm,
		CornerMask:     cornerMask,
		LabelFormatter: contourFormatter(opt.LabelFormatter),
		LabelFontSize:  opt.LabelFontSize.Or(10),
	}
	if color, ok := opt.LabelColor.Get(); ok {
		set.LabelColor = color
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
			Hatches:    hatches,
			Paths:      fillPaths,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		Polygons: polygons,
	}
	hatchRC := a.resolvedRC()
	applyContourHatchStyle(set.Fills, hatches, &hatchRC)
	a.Add(set)
	return set
}

// TriContour draws isolines over an explicit triangulation.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) TriContour(tri Triangulation, values []float64, opt ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, false, opt)
}

// TriContourf draws filled contour bands over an explicit triangulation.
//
//nolint:gocritic // The option value is an immutable snapshot forwarded unchanged.
func (a *Axes) TriContourf(tri Triangulation, values []float64, opt ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}
	return a.buildContourSet(tri, values, true, opt)
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
	segments, colors, widths, dashes, labels := contourInlineLabelSegmentsForLevels(c.Lines, c.lineLevels, c.labelLevels, c.LabelFormatter, fontSize, inlineSpacing, c.rightSideUp, r, ctx)
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
	lines.DashPatterns = dashes
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

//nolint:gocritic // Triangulation and ContourOptions are value snapshots; buildContourSet resolves defaults into its own copy.
func (a *Axes) buildContourSet(tri Triangulation, values []float64, filled bool, opt ContourOptions) *ContourSet {
	if err := tri.Validate(); err != nil || len(values) != len(tri.X) {
		return nil
	}

	a.resolveContourLineDefaults(&opt)

	levels := contourLevels(values, opt.Levels, opt.LevelCount, filled)
	if (!filled && len(levels) == 0) || (filled && len(levels) < 2) {
		return nil
	}

	alpha := meshAlpha(opt.Alpha)
	lineWidth := opt.LineWidth.OrZero()

	colorFallback := a.NextColor()
	cmapName := ""
	if name, ok := opt.Colormap.Get(); ok {
		cmapName = name
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
		LabelFontSize:  opt.LabelFontSize.Or(10),
		rightSideUp:    true,
		z:              0,
	}
	if color, ok := opt.LabelColor.Get(); ok {
		set.LabelColor = color
	}

	if filled {
		geomLevels := contourExtendedLevels(levels, opt.Extend)
		rawPolygons, rawColors, rawHatches, bands := contourBandPolygonsBands(tri, values, geomLevels, opt, mapping, alpha)
		fillPaths, polygons, faceColors, hatches := contourFillGeometry(rawPolygons, rawColors, rawHatches, bands)
		if len(rawPolygons) > 0 {
			cmap := ""
			vmin := 0.0
			vmax := 0.0
			if !opt.Color.IsSet() && len(opt.Colors) == 0 {
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
					Hatches:    hatches,
					Paths:      fillPaths,
					LineJoin:   render.JoinMiter,
					LineCap:    render.CapButt,
				},
				Polygons: polygons,
			}
			hatchRC := a.resolvedRC()
			applyContourHatchStyle(set.Fills, hatches, &hatchRC)
		}
	} else {
		polylines, polylineLevels := contourPolylines(tri, values, levels)
		if len(polylines) > 0 {
			colors := make([]render.Color, len(polylines))
			for i, level := range polylineLevels {
				colors[i] = contourLineColor(level, levels, opt, mapping, alpha, colorFallback)
			}
			styles := resolveContourLineStyles(levels, opt, contourMonochrome(opt))
			resolvedRC := a.resolvedRC()
			dashPatterns := contourLineDashPatterns(polylineLevels, levels, styles, lineWidth, &resolvedRC.Lines)
			set.Lines = &LineCollection{
				Collection: Collection{
					Coords: Coords(CoordData),
					Label:  opt.Label,
					Alpha:  1,
				},
				Segments:     polylines,
				Colors:       colors,
				LineWidth:    lineWidth,
				DashPatterns: dashPatterns,
				LineJoin:     render.JoinRound,
				LineCap:      render.CapButt,
			}
			set.lineLevels = append([]float64(nil), polylineLevels...)
			if opt.LabelLines {
				set.inlineLabels = true
				set.labels = contourLabels(polylines, polylineLevels, colors, set.LabelFormatter, set.rightSideUp)
			}
		}
	}

	if set.Fills == nil && set.Lines == nil {
		return nil
	}
	a.Add(set)
	return set
}
