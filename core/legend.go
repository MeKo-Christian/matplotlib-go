package core

import (
	"math"
	"reflect"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

// LegendLocation controls where the legend box is anchored inside the axes.
type LegendLocation uint8

const (
	LegendUpperRight LegendLocation = iota
	LegendUpperLeft
	LegendLowerRight
	LegendLowerLeft
	LegendRight
	LegendCenterLeft
	LegendCenterRight
	LegendLowerCenter
	LegendUpperCenter
	LegendCenter
	LegendBest
)

type legendEntryKind uint8

const (
	legendEntryLine legendEntryKind = iota
	legendEntryMarker
	legendEntryPatch
	legendEntryErrorBar
)

// LegendSample selects the sample style for an explicit legend entry.
type LegendSample uint8

const (
	// LegendSampleLine draws a line sample for an explicit legend entry.
	LegendSampleLine LegendSample = iota
	// LegendSampleMarker draws marker samples for an explicit legend entry.
	LegendSampleMarker
	// LegendSamplePatch draws a filled patch sample for an explicit legend entry.
	LegendSamplePatch
)

// LegendEntryOptions configures an explicit legend entry, useful for proxy-like
// entries that are not collected from an artist.
type LegendEntryOptions struct {
	Sample LegendSample

	Color     render.Color
	LineWidth float64
	Dashes    []float64

	Marker          MarkerType
	MarkerPath      geom.Path
	MarkerFaceColor render.Color
	MarkerEdgeColor render.Color
	MarkerEdgeWidth float64

	FaceColor  render.Color
	EdgeColor  render.Color
	EdgeWidth  float64
	Hatch      string
	HatchColor render.Color
	HatchWidth float64
}

type legendEntry struct {
	Label string

	kind legendEntryKind

	lineColor     render.Color
	lineWidth     float64
	lineJoin      render.LineJoin
	lineCap       render.LineCap
	dashes        []float64
	lineMarkerSet bool

	marker          MarkerType
	markerStyle     MarkerStyle
	markerPath      geom.Path
	markerAltPath   geom.Path
	markerEdgePath  geom.Path
	markerHasAlt    bool
	markerLineOnly  bool
	markerFill      render.Color
	markerAltFill   render.Color
	markerEdge      render.Color
	markerEdgeWidth float64
	markerSize      float64
	markerLineJoin  render.LineJoin
	markerLineCap   render.LineCap
	markerSnap      render.SnapMode

	patchFill       render.Color
	patchEdge       render.Color
	patchEdgeWidth  float64
	patchHatch      string
	patchHatchColor render.Color
	patchHatchWidth float64

	errorbarX       bool
	errorbarY       bool
	errorbarCapSize float64
}

type legendEntryProvider interface {
	legendEntry() (legendEntry, bool)
}

type legendHandlerOverride struct {
	artist Artist
	entry  legendEntry
}

// Legend renders a styled legend box inside an axes.
// If no explicit internal entries are present, labeled artists on the owning axes are collected automatically.
type Legend struct {
	Axes   *Axes
	Figure *Figure

	entries  []legendEntry
	handlers []legendHandlerOverride

	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	SampleWidth     float64
	SampleTextGap   float64
	ColumnSpacing   float64
	NumColumns      int
	MarkerScale     float64
	ScatterPoints   int
	Title           string
	TitleFontSize   float64
	FrameOn         bool
	CornerRadius    float64
	BackgroundColor render.Color
	BorderColor     render.Color
	TextColor       render.Color
	BorderWidth     float64
	FontSize        float64
	z               float64
}

// NewLegend creates a legend bound to the provided axes.
func NewLegend(ax *Axes) *Legend {
	rc := style.CurrentDefaults()
	if ax != nil {
		rc = ax.resolvedRC()
	}
	fontSize := rc.LegendSize()
	fontPx := pointsToPixels(rc, fontSize)
	return &Legend{
		Axes:            ax,
		Location:        LegendBest,
		Locator:         nil,
		Padding:         0.4 * fontPx,
		Inset:           0.5 * fontPx,
		RowGap:          0.5 * fontPx,
		SampleWidth:     2.0 * fontPx,
		SampleTextGap:   0.8 * fontPx,
		ColumnSpacing:   2.0 * fontPx,
		NumColumns:      1,
		MarkerScale:     1,
		ScatterPoints:   1,
		FrameOn:         true,
		CornerRadius:    0.2 * fontPx,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
	}
}

// NewFigureLegend creates a legend bound to the provided figure.
func NewFigureLegend(fig *Figure) *Legend {
	rc := style.CurrentDefaults()
	if fig != nil {
		rc = fig.RC
	}
	fontSize := rc.LegendSize()
	fontPx := pointsToPixels(rc, fontSize)
	return &Legend{
		Figure:          fig,
		Location:        LegendUpperRight,
		Locator:         nil,
		Padding:         0.4 * fontPx,
		Inset:           0.5 * fontPx,
		RowGap:          0.5 * fontPx,
		SampleWidth:     2.0 * fontPx,
		SampleTextGap:   0.8 * fontPx,
		ColumnSpacing:   2.0 * fontPx,
		NumColumns:      1,
		MarkerScale:     1,
		ScatterPoints:   1,
		FrameOn:         true,
		CornerRadius:    0.2 * fontPx,
		BackgroundColor: rc.LegendBackground,
		BorderColor:     rc.LegendBorderColor,
		TextColor:       rc.LegendTextColor,
		BorderWidth:     pointsToPixels(rc, 1),
		FontSize:        fontSize,
		z:               1_000,
	}
}

// AddLegend appends a legend to the axes.
func (a *Axes) AddLegend() *Legend {
	legend := NewLegend(a)
	a.Add(legend)
	return legend
}

// AddLegend appends a figure-level legend that collects labeled artists from all axes.
func (f *Figure) AddLegend() *Legend {
	legend := NewFigureLegend(f)
	f.Add(legend)
	return legend
}

// AddEntry appends an explicit proxy-like entry to the legend.
func (l *Legend) AddEntry(label string, opts LegendEntryOptions) *Legend {
	if l == nil || !legendLabelVisible(label) {
		return l
	}
	l.entries = append(l.entries, legendEntryFromOptions(label, opts))
	return l
}

// SetHandler overrides the legend sample for a collected artist using the same
// typed sample options as explicit proxy entries.
func (l *Legend) SetHandler(art Artist, opts LegendEntryOptions) *Legend {
	if l == nil || art == nil {
		return l
	}
	entry := legendEntryFromOptions("", opts)
	for i := range l.handlers {
		if sameLegendArtist(l.handlers[i].artist, art) {
			l.handlers[i].entry = entry
			return l
		}
	}
	l.handlers = append(l.handlers, legendHandlerOverride{artist: art, entry: entry})
	return l
}

// ClearHandler removes a custom legend sample override for a collected artist.
func (l *Legend) ClearHandler(art Artist) *Legend {
	if l == nil || art == nil {
		return l
	}
	for i := range l.handlers {
		if sameLegendArtist(l.handlers[i].artist, art) {
			l.handlers = append(l.handlers[:i], l.handlers[i+1:]...)
			return l
		}
	}
	return l
}

// Draw renders the legend box and entries.
func (l *Legend) Draw(r render.Renderer, ctx *DrawContext) {
	l.draw(r, ctx)
}

// DrawOverlay renders axes legends after the axes clip has been removed,
// matching Matplotlib's unclipped legend artists.
func (l *Legend) DrawOverlay(r render.Renderer, ctx *DrawContext) {
	l.draw(r, ctx)
}

func (l *Legend) draw(r render.Renderer, ctx *DrawContext) {
	textRen, ok := r.(render.TextDrawer)
	if !ok {
		return
	}

	entries := l.entries
	if len(entries) == 0 {
		entries = l.collectEntries()
	}
	if len(entries) == 0 {
		return
	}

	fontSize := l.FontSize
	if fontSize <= 0 {
		fontSize = ctx.RC.LegendSize()
	}
	if fontSize < 8 {
		fontSize = 8
	}

	rowHeights := make([]float64, len(entries))
	labelLayouts := make([]singleLineTextLayout, len(entries))
	for i, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		labelLayouts[i] = layout
		rowHeights[i] = legendRowHeight(layout, fontSize, ctx)
	}

	layout := l.layoutEntries(labelLayouts, rowHeights, ctx, fontSize)
	titleLayout, titleHeight, titleFontSize := l.titleLayout(r, ctx, fontSize)
	contentWidth := layout.Width
	if titleLayout.Width > contentWidth {
		contentWidth = titleLayout.Width
	}
	contentHeight := layout.Height
	if l.Title != "" {
		contentHeight += titleHeight + l.RowGap
	}
	boxWidth := l.Padding*2 + contentWidth
	boxHeight := l.Padding*2 + contentHeight
	box := l.legendBoxRect(ctx, boxWidth, boxHeight)

	if l.FrameOn {
		boxPath := pixelRectPath(box)
		if l.CornerRadius > 0 {
			boxPath = roundedRectPath(box, l.CornerRadius)
		}
		r.Path(boxPath, &render.Paint{
			Fill:      l.BackgroundColor,
			Stroke:    l.BorderColor,
			LineWidth: l.BorderWidth,
			LineJoin:  render.JoinMiter,
			LineCap:   render.CapButt,
		})
	}

	y := box.Max.Y - l.Padding
	if l.Title != "" {
		drawDisplayText(
			textRen,
			l.Title,
			alignedSingleLineOrigin(
				geom.Pt{X: box.Min.X + l.Padding + contentWidth/2, Y: y - titleHeight/2},
				titleLayout,
				TextAlignCenter,
				textLayoutVAlignCenter,
			),
			titleFontSize,
			l.TextColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)
		y -= titleHeight + l.RowGap
	}

	x := box.Min.X + l.Padding
	for _, col := range layout.Columns {
		columnY := y
		for i := col.Start; i < col.End; i++ {
			entry := entries[i]
			rowHeight := rowHeights[i]
			centerY := columnY - rowHeight/2
			labelLayout := labelLayouts[i]

			l.drawSampleWithFontPixels(r, entry, geom.Rect{
				Min: geom.Pt{X: x, Y: centerY - rowHeight/2},
				Max: geom.Pt{X: x + l.SampleWidth, Y: centerY + rowHeight/2},
			}, pointsToPixels(ctx.RC, fontSize))

			drawDisplayText(
				textRen,
				entry.Label,
				alignedSingleLineOrigin(
					geom.Pt{X: x + l.SampleWidth + l.SampleTextGap, Y: centerY},
					labelLayout,
					TextAlignLeft,
					textLayoutVAlignCenter,
				),
				fontSize,
				l.TextColor,
				ctx.RC.FontKey,
				ctx.RC.UseTeX,
			)

			columnY -= rowHeight + l.RowGap
		}
		x += col.Width + layout.ColumnSpacing
	}
}

// Z returns the legend z-order.
func (l *Legend) Z() float64 {
	return l.z
}

// Bounds returns an empty rect because legends do not contribute to data bounds.
func (l *Legend) Bounds(*DrawContext) geom.Rect {
	return geom.Rect{}
}

// SetLocator overrides the anchored-box placement strategy for this legend.
func (l *Legend) SetLocator(locator AnchoredBoxLocator) {
	if l == nil {
		return
	}
	l.Locator = locator
}

func (l *Legend) boxRect(r render.Renderer, ctx *DrawContext) (geom.Rect, bool) {
	if l == nil || r == nil || ctx == nil {
		return geom.Rect{}, false
	}

	entries := l.entries
	if len(entries) == 0 {
		entries = l.collectEntries()
	}
	if len(entries) == 0 {
		return geom.Rect{}, false
	}

	fontSize := l.FontSize
	if fontSize <= 0 {
		fontSize = ctx.RC.LegendSize()
	}
	if fontSize < 8 {
		fontSize = 8
	}

	rowHeights := make([]float64, len(entries))
	labelLayouts := make([]singleLineTextLayout, len(entries))
	for i, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		labelLayouts[i] = layout
		rowHeights[i] = legendRowHeight(layout, fontSize, ctx)
	}

	layout := l.layoutEntries(labelLayouts, rowHeights, ctx, fontSize)
	titleLayout, titleHeight, _ := l.titleLayout(r, ctx, fontSize)
	contentWidth := layout.Width
	if titleLayout.Width > contentWidth {
		contentWidth = titleLayout.Width
	}
	contentHeight := layout.Height
	if l.Title != "" {
		contentHeight += titleHeight + l.RowGap
	}
	boxWidth := l.Padding*2 + contentWidth
	boxHeight := l.Padding*2 + contentHeight
	return l.legendBoxRect(ctx, boxWidth, boxHeight), true
}

func (l *Legend) titleLayout(r render.Renderer, ctx *DrawContext, fallbackFontSize float64) (singleLineTextLayout, float64, float64) {
	if l == nil || l.Title == "" {
		return singleLineTextLayout{}, 0, fallbackFontSize
	}
	fontSize := l.TitleFontSize
	if fontSize <= 0 {
		fontSize = fallbackFontSize
	}
	layout := measureSingleLineTextLayout(r, l.Title, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	return layout, legendRowHeight(layout, fontSize, ctx), fontSize
}

type legendLayout struct {
	Columns       []legendColumnLayout
	Width         float64
	Height        float64
	ColumnSpacing float64
}

type legendColumnLayout struct {
	Start int
	End   int
	Width float64
}

func (l *Legend) layoutEntries(labelLayouts []singleLineTextLayout, rowHeights []float64, ctx *DrawContext, fontSize float64) legendLayout {
	count := len(labelLayouts)
	columns := l.effectiveNumColumns(count)
	if columns <= 0 {
		return legendLayout{}
	}
	columnSpacing := l.ColumnSpacing
	if columnSpacing <= 0 && ctx != nil {
		columnSpacing = 2.0 * pointsToPixels(ctx.RC, fontSize)
	}
	layout := legendLayout{
		Columns:       make([]legendColumnLayout, 0, columns),
		ColumnSpacing: columnSpacing,
	}
	start := 0
	for col := 0; col < columns; col++ {
		size := count / columns
		if col < count%columns {
			size++
		}
		if size <= 0 {
			continue
		}
		end := start + size
		maxLabelWidth := 0.0
		columnHeight := 0.0
		for i := start; i < end; i++ {
			labelWidth := legendLabelWidth(labelLayouts[i])
			if labelWidth > maxLabelWidth {
				maxLabelWidth = labelWidth
			}
			columnHeight += rowHeights[i]
		}
		if size > 1 {
			columnHeight += l.RowGap * float64(size-1)
		}
		columnWidth := l.SampleWidth + l.SampleTextGap + maxLabelWidth
		layout.Columns = append(layout.Columns, legendColumnLayout{Start: start, End: end, Width: columnWidth})
		layout.Width += columnWidth
		if columnHeight > layout.Height {
			layout.Height = columnHeight
		}
		start = end
	}
	if len(layout.Columns) > 1 {
		layout.Width += columnSpacing * float64(len(layout.Columns)-1)
	}
	return layout
}

func (l *Legend) effectiveNumColumns(entryCount int) int {
	if entryCount <= 0 {
		return 0
	}
	columns := l.NumColumns
	if columns <= 0 {
		columns = 1
	}
	if columns > entryCount {
		columns = entryCount
	}
	return columns
}

func legendRowHeight(layout singleLineTextLayout, fontSize float64, ctx *DrawContext) float64 {
	fontPx := pointsToPixels(ctx.RC, fontSize)
	rowHeight := layout.RunAscent + layout.RunDescent
	if layout.MathLayout != nil && rowHeight < fontPx*1.28 {
		rowHeight = fontPx * 1.28
	}
	if rowHeight < fontPx {
		rowHeight = fontPx
	}
	if rowHeight <= 0 {
		rowHeight = layout.Height
	}
	return rowHeight
}

func legendLabelWidth(layout singleLineTextLayout) float64 {
	return layout.Width
}

func (l *Legend) collectEntries() []legendEntry {
	if l == nil {
		return nil
	}

	switch {
	case l.Axes != nil:
		return l.collectLegendEntries(l.Axes.Artists)
	case l.Figure != nil:
		var entries []legendEntry
		for _, ax := range l.Figure.Children {
			entries = append(entries, l.collectLegendEntries(ax.Artists)...)
		}
		return entries
	default:
		return nil
	}
}

func collectLegendEntries(artists []Artist) []legendEntry {
	return (*Legend)(nil).collectLegendEntries(artists)
}

func (l *Legend) collectLegendEntries(artists []Artist) []legendEntry {
	entries := make([]legendEntry, 0, len(artists))
	deferredErrorBars := make([]Artist, 0)
	for i := 0; i < len(artists); i++ {
		art := artists[i]
		if entry, ok := l.stemLegendEntryAt(artists, i); ok {
			entries = append(entries, entry)
			i++
			continue
		}
		if _, ok := art.(*ErrorBar); ok {
			deferredErrorBars = append(deferredErrorBars, art)
			continue
		}
		if entry, ok := l.legendEntryForArtist(art); ok {
			entries = append(entries, entry)
		}
	}
	for _, art := range deferredErrorBars {
		if entry, ok := l.legendEntryForArtist(art); ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (l *Legend) legendEntryForArtist(art Artist) (legendEntry, bool) {
	switch art.(type) {
	case *Legend:
		return legendEntry{}, false
	default:
		provider, ok := art.(legendEntryProvider)
		if !ok {
			return legendEntry{}, false
		}
		label := ArtistLabel(art)
		if !legendLabelVisible(label) {
			return legendEntry{}, false
		}
		if entry, ok := l.handlerEntryFor(art, label); ok {
			return entry, true
		}
		entry, ok := provider.legendEntry()
		if !ok {
			return legendEntry{}, false
		}
		entry.Label = label
		return entry, true
	}
}

func (l *Legend) stemLegendEntryAt(artists []Artist, i int) (legendEntry, bool) {
	if i+1 >= len(artists) {
		return legendEntry{}, false
	}
	stems, ok := artists[i].(*LineCollection)
	if !ok || !lineCollectionLooksLikeStems(stems) {
		return legendEntry{}, false
	}
	markers, ok := artists[i+1].(*PathCollection)
	if !ok {
		return legendEntry{}, false
	}
	label := stems.label()
	if label == "" || label != markers.label() || !legendLabelVisible(label) {
		return legendEntry{}, false
	}
	if _, ok := l.handlerEntryFor(stems, label); ok {
		return legendEntry{}, false
	}
	if _, ok := l.handlerEntryFor(markers, label); ok {
		return legendEntry{}, false
	}

	lineEntry, ok := stems.legendEntry()
	if !ok {
		return legendEntry{}, false
	}
	markerEntry, ok := markers.legendEntry()
	if !ok {
		return legendEntry{}, false
	}

	entry := lineEntry
	entry.kind = legendEntryErrorBar
	entry.errorbarY = true
	entry.lineMarkerSet = true
	entry.marker = markerEntry.marker
	entry.markerPath = markerEntry.markerPath
	entry.markerAltPath = markerEntry.markerAltPath
	entry.markerEdgePath = markerEntry.markerEdgePath
	entry.markerHasAlt = markerEntry.markerHasAlt
	entry.markerLineOnly = markerEntry.markerLineOnly
	entry.markerFill = markerEntry.markerFill
	entry.markerAltFill = markerEntry.markerAltFill
	entry.markerEdge = markerEntry.markerEdge
	entry.markerEdgeWidth = markerEntry.markerEdgeWidth
	return entry, true
}

func lineCollectionLooksLikeStems(stems *LineCollection) bool {
	if stems == nil || len(stems.Segments) == 0 {
		return false
	}
	vertical := 0
	horizontal := 0
	for _, segment := range stems.Segments {
		if len(segment) != 2 {
			return false
		}
		dx := math.Abs(segment[1].X - segment[0].X)
		dy := math.Abs(segment[1].Y - segment[0].Y)
		switch {
		case dx <= 1e-12 && dy > 1e-12:
			vertical++
		case dy <= 1e-12 && dx > 1e-12:
			horizontal++
		default:
			return false
		}
	}
	return vertical == len(stems.Segments) || horizontal == len(stems.Segments)
}

func (l *Legend) handlerEntryFor(art Artist, label string) (legendEntry, bool) {
	if l == nil || art == nil {
		return legendEntry{}, false
	}
	for _, handler := range l.handlers {
		if sameLegendArtist(handler.artist, art) {
			entry := handler.entry
			entry.Label = label
			return entry, true
		}
	}
	return legendEntry{}, false
}

func sameLegendArtist(a, b Artist) bool {
	if a == nil || b == nil {
		return false
	}
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	if av.Kind() != reflect.Pointer || bv.Kind() != reflect.Pointer || av.Type() != bv.Type() {
		return false
	}
	return av.Pointer() == bv.Pointer()
}

func (l *Legend) legendBoxRect(ctx *DrawContext, width, height float64) geom.Rect {
	if ctx == nil {
		return geom.Rect{}
	}
	if l.Location == LegendBest && l.Locator == nil && l.Axes != nil {
		return l.bestLegendBoxRect(ctx, width, height)
	}
	return resolveAnchoredBoxRect(l.Locator, ctx.Clip, width, height, l.Location, l.Inset)
}

func (l *Legend) bestLegendBoxRect(ctx *DrawContext, width, height float64) geom.Rect {
	candidates := []LegendLocation{
		LegendUpperRight,
		LegendUpperLeft,
		LegendLowerLeft,
		LegendLowerRight,
		LegendRight,
		LegendCenterLeft,
		LegendCenterRight,
		LegendLowerCenter,
		LegendUpperCenter,
		LegendCenter,
	}
	data := l.legendAvoidanceData(ctx)

	best := anchoredBoxRect(ctx.Clip, width, height, candidates[0], l.Inset)
	bestBadness := legendPlacementBadness(best, data)
	if bestBadness == 0 {
		return best
	}

	for _, location := range candidates[1:] {
		box := anchoredBoxRect(ctx.Clip, width, height, location, l.Inset)
		badness := legendPlacementBadness(box, data)
		if badness == 0 {
			return box
		}
		if badness < bestBadness {
			best = box
			bestBadness = badness
		}
	}
	return best
}

type legendAvoidanceData struct {
	points []geom.Pt
	lines  []geom.Path
	boxes  []geom.Rect
}

func (l *Legend) legendAvoidanceData(ctx *DrawContext) legendAvoidanceData {
	if l == nil || l.Axes == nil || ctx == nil {
		return legendAvoidanceData{}
	}
	data := legendAvoidanceData{}
	appendPoints := func(spec CoordinateSpec, pts []geom.Pt) {
		tr := ctx.TransformFor(spec)
		for _, pt := range pts {
			if tr != nil {
				pt = tr.Apply(pt)
			}
			data.points = append(data.points, pt)
		}
	}
	appendRect := func(spec CoordinateSpec, rect geom.Rect) {
		if rect == (geom.Rect{}) {
			return
		}
		tr := ctx.TransformFor(spec)
		if tr == nil {
			data.boxes = append(data.boxes, rect)
			return
		}
		min := tr.Apply(rect.Min)
		max := tr.Apply(rect.Max)
		data.boxes = append(data.boxes, geom.Rect{
			Min: geom.Pt{X: math.Min(min.X, max.X), Y: math.Min(min.Y, max.Y)},
			Max: geom.Pt{X: math.Max(min.X, max.X), Y: math.Max(min.Y, max.Y)},
		})
	}
	appendLine := func(path geom.Path) {
		if len(path.V) == 0 {
			return
		}
		data.lines = append(data.lines, path)
		data.points = append(data.points, path.V...)
	}

	for _, art := range l.Axes.Artists {
		switch a := art.(type) {
		case *Legend:
			continue
		case *Scatter2D:
			appendPoints(Coords(CoordData), a.XY)
		case *Line2D:
			appendLine(a.displayPath(ctx))
		case *PathCollection:
			appendPoints(a.Coords, a.Offsets)
		case *LineCollection:
			for _, segment := range a.Segments {
				appendLine(displayPolylineForLegend(ctx, a.Coords, segment))
			}
		case *Rectangle:
			appendRect(a.Coords, a.Bounds(ctx))
		case *PathPatch:
			appendLine(buildArtistDisplayPath(ctx, a, a.Coords, a.Path, geom.Identity()))
		case *Image2D:
			appendRect(Coords(CoordData), a.Bounds(ctx))
		case *Text:
			data.points = append(data.points, transformedPoint(ctx, a.Coords, a.Position, a.OffsetX, a.OffsetY))
		case *Annotation:
			target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
			text := transformedPoint(ctx, a.Coords, a.Point, a.OffsetX, a.OffsetY)
			data.points = append(data.points, target, text)
		case *AnnotationBbox:
			target := transformedPoint(ctx, a.XYCoords, a.Point, 0, 0)
			box := target
			if a.BoxPosition != nil {
				box = transformedPoint(ctx, a.BoxCoords, *a.BoxPosition, 0, 0)
			}
			data.points = append(data.points, target, box)
		}
	}
	return data
}

func displayPolylineForLegend(ctx *DrawContext, spec CoordinateSpec, pts []geom.Pt) geom.Path {
	tr := ctx.TransformFor(spec)
	path := geom.Path{}
	inSegment := false
	for _, pt := range pts {
		if !finitePoint(pt) {
			inSegment = false
			continue
		}
		if tr != nil {
			pt = tr.Apply(pt)
		}
		if !finitePoint(pt) {
			inSegment = false
			continue
		}
		if !inSegment {
			path.C = append(path.C, geom.MoveTo)
			inSegment = true
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, pt)
	}
	return path
}

func legendPlacementBadness(box geom.Rect, data legendAvoidanceData) int {
	badness := 0
	for _, pt := range data.points {
		if pointInRect(pt, box) {
			badness++
		}
	}
	for _, other := range data.boxes {
		if rectsOverlap(box, other) {
			badness++
		}
	}
	for _, line := range data.lines {
		if pathIntersectsRect(line, box) {
			badness++
		}
	}
	return badness
}

func pointInRect(pt geom.Pt, rect geom.Rect) bool {
	return pt.X >= rect.Min.X && pt.X <= rect.Max.X && pt.Y >= rect.Min.Y && pt.Y <= rect.Max.Y
}

func rectsOverlap(a, b geom.Rect) bool {
	return a.Min.X < b.Max.X && a.Max.X > b.Min.X && a.Min.Y < b.Max.Y && a.Max.Y > b.Min.Y
}

func pathIntersectsRect(path geom.Path, rect geom.Rect) bool {
	var cur geom.Pt
	haveCur := false
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi >= len(path.V) {
				return false
			}
			cur = path.V[vi]
			haveCur = true
			vi++
		case geom.LineTo:
			if vi >= len(path.V) {
				return false
			}
			next := path.V[vi]
			if haveCur && segmentIntersectsRect(cur, next, rect) {
				return true
			}
			cur = next
			haveCur = true
			vi++
		case geom.QuadTo:
			vi += 2
			haveCur = false
		case geom.CubicTo:
			vi += 3
			haveCur = false
		case geom.ClosePath:
			haveCur = false
		}
	}
	return false
}

func segmentIntersectsRect(a, b geom.Pt, rect geom.Rect) bool {
	corners := []geom.Pt{
		rect.Min,
		{X: rect.Max.X, Y: rect.Min.Y},
		rect.Max,
		{X: rect.Min.X, Y: rect.Max.Y},
	}
	for i := range corners {
		if segmentsIntersect(a, b, corners[i], corners[(i+1)%len(corners)]) {
			return true
		}
	}
	return false
}

func segmentsIntersect(a, b, c, d geom.Pt) bool {
	const eps = 1e-9
	orient := func(p, q, r geom.Pt) float64 {
		return (q.X-p.X)*(r.Y-p.Y) - (q.Y-p.Y)*(r.X-p.X)
	}
	onSegment := func(p, q, r geom.Pt) bool {
		return math.Min(p.X, r.X)-eps <= q.X && q.X <= math.Max(p.X, r.X)+eps &&
			math.Min(p.Y, r.Y)-eps <= q.Y && q.Y <= math.Max(p.Y, r.Y)+eps
	}
	o1 := orient(a, b, c)
	o2 := orient(a, b, d)
	o3 := orient(c, d, a)
	o4 := orient(c, d, b)
	if math.Abs(o1) <= eps && onSegment(a, c, b) {
		return true
	}
	if math.Abs(o2) <= eps && onSegment(a, d, b) {
		return true
	}
	if math.Abs(o3) <= eps && onSegment(c, a, d) {
		return true
	}
	if math.Abs(o4) <= eps && onSegment(c, b, d) {
		return true
	}
	return (o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0)
}

func legendEntryFromOptions(label string, opts LegendEntryOptions) legendEntry {
	switch opts.Sample {
	case LegendSampleMarker:
		return legendEntryFromMarker(
			label,
			opts.Marker,
			opts.MarkerPath,
			defaultVisibleColor(opts.MarkerFaceColor),
			defaultVisibleColor(opts.MarkerEdgeColor),
			opts.MarkerEdgeWidth,
		)
	case LegendSamplePatch:
		return legendEntryFromPatchStyle(
			label,
			opts.FaceColor,
			opts.EdgeColor,
			opts.EdgeWidth,
			opts.Hatch,
			opts.HatchColor,
			opts.HatchWidth,
		)
	default:
		return legendEntryFromLine(label, defaultVisibleColor(opts.Color), opts.LineWidth, opts.Dashes)
	}
}

func defaultVisibleColor(color render.Color) render.Color {
	if color == (render.Color{}) {
		return render.Color{A: 1}
	}
	return color
}

func (l *Legend) drawSample(r render.Renderer, entry legendEntry, sample geom.Rect) {
	fontSize := style.Default.LegendSize()
	if l != nil && l.FontSize > 0 {
		fontSize = l.FontSize
	}
	l.drawSampleWithFontPixels(r, entry, sample, pointsToPixels(style.Default, fontSize))
}

func (l *Legend) drawSampleWithFontPixels(r render.Renderer, entry legendEntry, sample geom.Rect, fontPx float64) {
	center := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}

	switch entry.kind {
	case legendEntryErrorBar:
		l.drawErrorBarSample(r, entry, sample, center, fontPx)
	case legendEntryPatch:
		// Matplotlib's HandlerPatch fills the legend handle box. The
		// handleheight default is 0.7 font-size units, and the default
		// handlelength is 2.0 font-size units, so it occupies 70% of the
		// row height and the full handle width.
		handleHeight := sample.H() * 0.7
		if handleHeight <= 0 {
			handleHeight = sample.H()
		}
		patchRect := geom.Rect{
			Min: geom.Pt{X: sample.Min.X, Y: center.Y - handleHeight/2},
			Max: geom.Pt{X: sample.Max.X, Y: center.Y + handleHeight/2},
		}
		patch := Patch{
			FaceColor:  entry.patchFill,
			EdgeColor:  entry.patchEdge,
			EdgeWidth:  entry.patchEdgeWidth,
			Hatch:      entry.patchHatch,
			HatchColor: entry.patchHatchColor,
			HatchWidth: entry.patchHatchWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		}
		patch.drawStyledPath(r, pixelRectPath(patchRect), geom.Path{})
	case legendEntryMarker:
		for _, pt := range l.markerSampleCenters(sample, center) {
			l.drawMarkerSample(r, entry, pt, l.markerSampleScale(entry, 5))
		}
	default:
		lineWidth := entry.lineWidth
		if lineWidth <= 0 {
			lineWidth = 1.5
		}
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{
				{X: sample.Min.X, Y: center.Y},
				{X: sample.Max.X, Y: center.Y},
			},
		}
		r.Path(path, &render.Paint{
			Stroke:    entry.lineColor,
			LineWidth: lineWidth,
			LineJoin:  entry.lineJoin,
			LineCap:   entry.lineCap,
			Dashes:    entry.dashes,
		})
		if entry.lineMarkerSet {
			l.drawMarkerSample(r, entry, center, l.markerSampleScale(entry, 5))
		}
	}
}

func (l *Legend) drawErrorBarSample(r render.Renderer, entry legendEntry, sample geom.Rect, center geom.Pt, fontPx float64) {
	lineWidth := entry.lineWidth
	if lineWidth <= 0 {
		lineWidth = 1.5
	}
	paint := render.Paint{
		Stroke:    entry.lineColor,
		LineWidth: lineWidth,
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Dashes:    entry.dashes,
	}
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: sample.Min.X, Y: center.Y}, {X: sample.Max.X, Y: center.Y}},
	}, &paint)

	capHalf := entry.errorbarCapSize / 2
	if capHalf <= 0 {
		capHalf = 3
	} else {
		capHalf = entry.errorbarCapSize
	}
	if entry.errorbarY {
		errSize := fontPx * 0.5
		if errSize <= 0 {
			errSize = sample.H() * 0.5
		}
		top := geom.Pt{X: center.X, Y: center.Y - errSize}
		bottom := geom.Pt{X: center.X, Y: center.Y + errSize}
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{top, bottom},
		}, &paint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: top.X - capHalf, Y: top.Y}, {X: top.X + capHalf, Y: top.Y}},
		}, &paint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: bottom.X - capHalf, Y: bottom.Y}, {X: bottom.X + capHalf, Y: bottom.Y}},
		}, &paint)
	}
	if entry.errorbarX {
		left := geom.Pt{X: sample.Min.X + sample.W()*0.25, Y: center.Y}
		right := geom.Pt{X: sample.Max.X - sample.W()*0.25, Y: center.Y}
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{left, right},
		}, &paint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: left.X, Y: left.Y - capHalf}, {X: left.X, Y: left.Y + capHalf}},
		}, &paint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: right.X, Y: right.Y - capHalf}, {X: right.X, Y: right.Y + capHalf}},
		}, &paint)
	}
	if entry.lineMarkerSet {
		l.drawMarkerSample(r, entry, center, l.markerSampleScale(entry, 5))
	}
}

func (l *Legend) markerSampleScale(entry legendEntry, base float64) float64 {
	scale := base
	if entry.markerSize > 0 {
		scale = entry.markerSize
	}
	markerScale := 1.0
	if l != nil && l.MarkerScale > 0 {
		markerScale = l.MarkerScale
	}
	return scale * markerScale
}

func (l *Legend) markerSampleCenters(sample geom.Rect, center geom.Pt) []geom.Pt {
	points := 1
	if l != nil && l.ScatterPoints > 0 {
		points = l.ScatterPoints
	}
	if points <= 1 {
		return []geom.Pt{center}
	}
	centers := make([]geom.Pt, points)
	pad := sample.W() * 0.15
	if pad < 0 {
		pad = 0
	}
	step := 0.0
	if points > 1 {
		step = (sample.W() - 2*pad) / float64(points-1)
	}
	for i := 0; i < points; i++ {
		y := center.Y
		if points > 1 {
			// Matplotlib's HandlerPathCollection uses Legend._scatteryoffsets
			// [3/8, 4/8, 2.5/8] within a default 0.7-fontsize handle box.
			offsets := [...]float64{3.0 / 8.0, 4.0 / 8.0, 2.5 / 8.0}
			handleHeight := sample.H() * 0.7
			y = center.Y - handleHeight/2 + offsets[i%len(offsets)]*handleHeight
		}
		centers[i] = geom.Pt{
			X: sample.Min.X + pad + step*float64(i),
			Y: y,
		}
	}
	return centers
}

func (l *Legend) drawMarkerSample(r render.Renderer, entry legendEntry, center geom.Pt, radius float64) {
	center.X += 0.5
	center.Y += 0.5
	lineJoin := entry.markerLineJoin
	if lineJoin == 0 {
		lineJoin = render.JoinRound
	}
	lineCap := entry.markerLineCap
	if lineCap == 0 {
		lineCap = render.CapButt
	}
	markerPath := entry.markerPath
	markerScale := radius
	if len(markerPath.C) == 0 {
		sampleScatter := Scatter2D{Marker: entry.marker, MarkerStyle: entry.markerStyle}
		markerPath = sampleScatter.markerPrototypePathForContext(r, nil)
		if entry.markerStyle.Tuple == nil && entry.markerStyle.MathText == "" && len(entry.markerStyle.Path.C) == 0 && entry.markerStyle.Type == 0 && entry.markerStyle.FillStyle == 0 {
			markerScale = radius * stemMarkerScale
		}
	}
	if entry.markerHasAlt {
		drawLegendMarkerPath(r, markerPath, center, markerScale, entry.markerSnap, render.Paint{
			Fill:      entry.markerFill,
			Stroke:    entry.markerEdge,
			LineWidth: entry.markerEdgeWidth,
			LineJoin:  lineJoin,
			LineCap:   lineCap,
		})
		if len(entry.markerAltPath.C) > 0 {
			drawLegendMarkerPath(r, entry.markerAltPath, center, radius, entry.markerSnap, render.Paint{
				Fill:      entry.markerAltFill,
				Stroke:    entry.markerEdge,
				LineWidth: entry.markerEdgeWidth,
				LineJoin:  lineJoin,
				LineCap:   lineCap,
			})
		}
		return
	}
	fill := entry.markerFill
	edge := entry.markerEdge
	if entry.markerLineOnly {
		if edge.A <= 0 {
			edge = fill
		}
		fill.A = 0
	}
	drawLegendMarkerPath(r, markerPath, center, markerScale, entry.markerSnap, render.Paint{
		Fill:      fill,
		Stroke:    edge,
		LineWidth: entry.markerEdgeWidth,
		LineJoin:  lineJoin,
		LineCap:   lineCap,
	})
}

func drawLegendMarkerPath(r render.Renderer, markerPath geom.Path, center geom.Pt, scale float64, snap render.SnapMode, paint render.Paint) {
	if len(markerPath.C) == 0 || scale <= 0 {
		return
	}
	paint.Snap = snap
	path := scaleAndTranslatePath(markerPath, scale, center)
	r.Path(path, &paint)
}

func pixelRectPath(r geom.Rect) geom.Path {
	path := geom.Path{}
	corners := []geom.Pt{
		r.Min,
		{X: r.Max.X, Y: r.Min.Y},
		r.Max,
		{X: r.Min.X, Y: r.Max.Y},
	}
	for i, corner := range corners {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, corner)
	}
	path.C = append(path.C, geom.ClosePath)
	return path
}

func snappedPixelRectPath(r geom.Rect) geom.Path {
	return pixelRectPath(geom.Rect{
		Min: geom.Pt{X: math.Round(r.Min.X), Y: math.Round(r.Min.Y)},
		Max: geom.Pt{X: math.Round(r.Max.X), Y: math.Round(r.Max.Y)},
	})
}
