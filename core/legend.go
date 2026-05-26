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
	dashes        []float64
	lineMarkerSet bool

	marker          MarkerType
	markerPath      geom.Path
	markerAltPath   geom.Path
	markerEdgePath  geom.Path
	markerHasAlt    bool
	markerLineOnly  bool
	markerFill      render.Color
	markerAltFill   render.Color
	markerEdge      render.Color
	markerEdgeWidth float64

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
		BorderWidth:     1,
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
		BorderWidth:     1,
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

			l.drawSample(r, entry, geom.Rect{
				Min: geom.Pt{X: x, Y: centerY - rowHeight/2},
				Max: geom.Pt{X: x + l.SampleWidth, Y: centerY + rowHeight/2},
			})

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
			if labelLayouts[i].Width > maxLabelWidth {
				maxLabelWidth = labelLayouts[i].Width
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
	if rowHeight < fontPx {
		rowHeight = fontPx
	}
	if rowHeight <= 0 {
		rowHeight = layout.Height
	}
	return rowHeight
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
	candidates := []LegendLocation{LegendUpperRight, LegendUpperLeft, LegendLowerLeft, LegendLowerRight}
	points := l.legendAvoidancePoints(ctx)

	best := anchoredBoxRect(ctx.Clip, width, height, candidates[0], l.Inset)
	bestBadness := legendPlacementBadness(best, points)
	if bestBadness == 0 {
		return best
	}

	for _, location := range candidates[1:] {
		box := anchoredBoxRect(ctx.Clip, width, height, location, l.Inset)
		badness := legendPlacementBadness(box, points)
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

func (l *Legend) legendAvoidancePoints(ctx *DrawContext) []geom.Pt {
	if l == nil || l.Axes == nil || ctx == nil {
		return nil
	}
	points := []geom.Pt{}
	appendPoints := func(spec CoordinateSpec, pts []geom.Pt) {
		tr := ctx.TransformFor(spec)
		for _, pt := range pts {
			if tr != nil {
				pt = tr.Apply(pt)
			}
			points = append(points, pt)
		}
	}
	appendRectCorners := func(spec CoordinateSpec, rect geom.Rect) {
		if rect == (geom.Rect{}) {
			return
		}
		appendPoints(spec, []geom.Pt{
			rect.Min,
			{X: rect.Max.X, Y: rect.Min.Y},
			rect.Max,
			{X: rect.Min.X, Y: rect.Max.Y},
		})
	}

	for _, art := range l.Axes.Artists {
		switch a := art.(type) {
		case *Legend:
			continue
		case *Scatter2D:
			appendPoints(Coords(CoordData), a.XY)
		case *Line2D:
			appendPoints(Coords(CoordData), a.pathPoints())
		case *PathCollection:
			appendPoints(a.Coords, a.Offsets)
		case *LineCollection:
			for _, segment := range a.Segments {
				appendPoints(a.Coords, segment)
			}
		case *Image2D:
			appendRectCorners(Coords(CoordData), a.Bounds(ctx))
		case *Annotation:
			target := transformedPoint(ctx, a.Coords, a.Point, 0, 0)
			text := transformedPoint(ctx, a.Coords, a.Point, a.OffsetX, a.OffsetY)
			points = append(points, target, text)
		case *AnnotationBbox:
			target := transformedPoint(ctx, a.XYCoords, a.Point, 0, 0)
			box := target
			if a.BoxPosition != nil {
				box = transformedPoint(ctx, a.BoxCoords, *a.BoxPosition, 0, 0)
			}
			points = append(points, target, box)
		}
	}
	return points
}

func legendPlacementBadness(box geom.Rect, points []geom.Pt) int {
	badness := 0
	for _, pt := range points {
		if pointInRect(pt, box) {
			badness++
		}
	}
	return badness
}

func pointInRect(pt geom.Pt, rect geom.Rect) bool {
	return pt.X >= rect.Min.X && pt.X <= rect.Max.X && pt.Y >= rect.Min.Y && pt.Y <= rect.Max.Y
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
	center := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}

	switch entry.kind {
	case legendEntryErrorBar:
		l.drawErrorBarSample(r, entry, sample, center)
	case legendEntryPatch:
		patchRect := geom.Rect{
			Min: geom.Pt{X: sample.Min.X + 2, Y: center.Y - 5},
			Max: geom.Pt{X: sample.Max.X - 2, Y: center.Y + 5},
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
			l.drawMarkerSample(r, entry, pt, l.markerSampleRadius(5))
		}
	default:
		lineWidth := entry.lineWidth
		if lineWidth <= 0 {
			lineWidth = 1.5
		}
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{
				{X: sample.Min.X + 1, Y: center.Y},
				{X: sample.Max.X - 1, Y: center.Y},
			},
		}
		r.Path(path, &render.Paint{
			Stroke:    entry.lineColor,
			LineWidth: lineWidth,
			LineJoin:  render.JoinRound,
			LineCap:   render.CapRound,
			Dashes:    entry.dashes,
		})
		if entry.lineMarkerSet {
			l.drawMarkerSample(r, entry, center, l.markerSampleRadius(5))
		}
	}
}

func (l *Legend) drawErrorBarSample(r render.Renderer, entry legendEntry, sample geom.Rect, center geom.Pt) {
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
		V: []geom.Pt{{X: sample.Min.X + 1, Y: center.Y}, {X: sample.Max.X - 1, Y: center.Y}},
	}, &paint)

	capHalf := entry.errorbarCapSize / 2
	if capHalf <= 0 {
		capHalf = 3
	}
	if entry.errorbarY {
		top := geom.Pt{X: center.X, Y: sample.Min.Y + sample.H()*0.2}
		bottom := geom.Pt{X: center.X, Y: sample.Max.Y - sample.H()*0.2}
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
		l.drawMarkerSample(r, entry, center, l.markerSampleRadius(5))
	}
}

func (l *Legend) markerSampleRadius(base float64) float64 {
	scale := 1.0
	if l != nil && l.MarkerScale > 0 {
		scale = l.MarkerScale
	}
	return base * scale
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
	step := sample.W() / float64(points+1)
	for i := 0; i < points; i++ {
		centers[i] = geom.Pt{
			X: sample.Min.X + step*float64(i+1),
			Y: center.Y,
		}
	}
	return centers
}

func (l *Legend) drawMarkerSample(r render.Renderer, entry legendEntry, center geom.Pt, radius float64) {
	markerPath := entry.markerPath
	if len(markerPath.C) == 0 {
		sampleScatter := Scatter2D{Marker: entry.marker}
		markerPath = sampleScatter.createMarkerPath(center, radius)
	} else {
		markerPath = scaleAndTranslatePath(markerPath, radius, center)
	}
	if entry.markerHasAlt {
		r.Path(markerPath, &render.Paint{
			Fill:     entry.markerFill,
			LineJoin: render.JoinRound,
			LineCap:  render.CapRound,
		})
		if len(entry.markerAltPath.C) > 0 {
			r.Path(scaleAndTranslatePath(entry.markerAltPath, radius, center), &render.Paint{
				Fill:     entry.markerAltFill,
				LineJoin: render.JoinRound,
				LineCap:  render.CapRound,
			})
		}
		edgePath := entry.markerEdgePath
		if len(edgePath.C) == 0 {
			edgePath = entry.markerPath
		}
		if len(edgePath.C) > 0 {
			r.Path(scaleAndTranslatePath(edgePath, radius, center), &render.Paint{
				Stroke:    entry.markerEdge,
				LineWidth: entry.markerEdgeWidth,
				LineJoin:  render.JoinRound,
				LineCap:   render.CapRound,
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
	r.Path(markerPath, &render.Paint{
		Fill:      fill,
		Stroke:    edge,
		LineWidth: entry.markerEdgeWidth,
		LineJoin:  render.JoinRound,
		LineCap:   render.CapRound,
	})
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
