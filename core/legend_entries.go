package core

import (
	"math"
	"reflect"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
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

	errorbarX        bool
	errorbarY        bool
	errorbarCapSize  float64
	errorbarCapWidth float64
}

type legendEntryProvider interface {
	legendEntry() (legendEntry, bool)
}

type legendHandlerOverride struct {
	artist Artist
	entry  legendEntry
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

func legendEntryFromPatchStyle(label string, face, edge render.Color, edgeWidth float64, hatch string, hatchColor render.Color, hatchWidth float64) legendEntry {
	if hatch != "" {
		if hatchColor.A <= 0 {
			hatchColor = edge
		}
		if hatchColor.A <= 0 {
			hatchColor = face
		}
		if hatchColor.A <= 0 {
			hatchColor = render.Color{A: 1}
		}
		if hatchWidth <= 0 {
			hatchWidth = 100.0 / 72.0
		}
	}
	return legendEntry{
		Label:           label,
		kind:            legendEntryPatch,
		patchFill:       face,
		patchEdge:       edge,
		patchEdgeWidth:  edgeWidth,
		patchHatch:      hatch,
		patchHatchColor: hatchColor,
		patchHatchWidth: hatchWidth,
	}
}

func legendEntryFromLine(label string, color render.Color, width float64, dashes []float64) legendEntry {
	lineCap := render.CapSquare
	if len(dashes) > 0 {
		lineCap = render.CapButt
	}
	return legendEntry{
		Label:     label,
		kind:      legendEntryLine,
		lineColor: color,
		lineWidth: width,
		lineJoin:  render.JoinRound,
		lineCap:   lineCap,
		dashes:    append([]float64(nil), dashes...),
	}
}

func legendEntryFromMarker(label string, marker MarkerType, markerPath geom.Path, fill, edge render.Color, edgeWidth float64) legendEntry {
	scatter := Scatter2D{Marker: marker}
	markerSize := 0.0
	return legendEntry{
		Label:           label,
		kind:            legendEntryMarker,
		marker:          marker,
		markerStyle:     scatter.resolvedMarkerStyle(),
		markerPath:      markerPath,
		markerFill:      fill,
		markerEdge:      edge,
		markerEdgeWidth: edgeWidth,
		markerLineJoin:  scatter.markerLineJoin(),
		markerLineCap:   scatter.markerLineCap(),
		markerSnap:      markerSnapMode(scatter.resolvedMarkerStyle(), markerSize),
	}
}

func (l *Line2D) legendEntry() (legendEntry, bool) {
	if l == nil || l.Label == "" {
		return legendEntry{}, false
	}
	entry := legendEntryFromLine(l.Label, l.ApplyArtistAlpha(l.Col), l.W, l.Dashes)
	if l.hasMarkers() {
		style := l.resolvedMarkerStyle()
		spec := l.markerPathSpec(nil, nil)
		entry.lineMarkerSet = true
		entry.marker = l.Marker
		entry.markerStyle = style
		if style.MathText == "" {
			entry.markerPath = spec.Path
		}
		entry.markerAltPath = spec.AltPath
		entry.markerEdgePath = spec.EdgePath
		entry.markerHasAlt = spec.HasAlt
		entry.markerLineOnly = markerLineOnly(style)
		entry.markerFill = l.resolvedMarkerFaceColor()
		entry.markerAltFill = l.resolvedMarkerFaceColorAlt()
		entry.markerEdge = l.resolvedMarkerEdgeColor()
		entry.markerEdgeWidth = l.resolvedMarkerEdgeWidth(nil)
		entry.markerSize = l.resolvedMarkerSize(nil)
		markerScatter := Scatter2D{Marker: l.Marker, MarkerStyle: l.MarkerStyle, MarkerPath: l.MarkerPath}
		entry.markerLineJoin = markerScatter.markerLineJoin()
		entry.markerLineCap = markerScatter.markerLineCap()
		entry.markerSnap = markerSnapMode(l.resolvedMarkerStyle(), entry.markerSize)
	}
	return entry, true
}

func (s *Scatter2D) legendEntry() (legendEntry, bool) {
	if s == nil || s.Label == "" {
		return legendEntry{}, false
	}
	fill := s.Color
	if len(s.Colors) > 0 {
		fill = s.Colors[0]
	}
	edge := s.EdgeColor
	if len(s.EdgeColors) > 0 {
		edge = s.EdgeColors[0]
	}
	alpha := s.Alpha
	if alpha <= 0 {
		alpha = 1
	}
	fill = fill.WithAlphaMultiplier(alpha)
	edge = edge.WithAlphaMultiplier(alpha)
	if markerLineOnly(s.resolvedMarkerStyle()) {
		if edge.A <= 0 {
			edge = fill
		}
		fill.A = 0
	}
	entry := legendEntryFromMarker(s.Label, s.Marker, s.markerPrototypePathForContext(nil, nil), fill, edge, s.EdgeWidth)
	entry.markerStyle = s.resolvedMarkerStyle()
	if entry.markerStyle.MathText != "" {
		entry.markerPath = geom.Path{}
	}
	entry.markerLineOnly = markerLineOnly(s.resolvedMarkerStyle())
	size := s.Size
	if size <= 0 {
		size = 36
	}
	if legendSize, ok := legendCollectionArea(s.Sizes, size); ok {
		size = legendSize
	}
	entry.markerSize = pointsToPixels(style.Default, math.Sqrt(size))
	entry.markerLineJoin = s.markerLineJoin()
	entry.markerLineCap = s.markerLineCap()
	entry.markerSnap = markerSnapMode(s.resolvedMarkerStyle(), entry.markerSize)
	return entry, true
}

func legendCollectionArea(sizes []float64, fallback float64) (float64, bool) {
	if len(sizes) == 0 {
		if fallback > 0 {
			return fallback, true
		}
		return 0, false
	}
	minSize := math.Inf(1)
	maxSize := 0.0
	for _, size := range sizes {
		if size <= 0 || math.IsNaN(size) || math.IsInf(size, 0) {
			continue
		}
		if size < minSize {
			minSize = size
		}
		if size > maxSize {
			maxSize = size
		}
	}
	if maxSize <= 0 || math.IsInf(minSize, 1) {
		if fallback > 0 {
			return fallback, true
		}
		return 0, false
	}
	return 0.5 * (minSize + maxSize), true
}

func (b *Bar2D) legendEntry() (legendEntry, bool) {
	if b == nil || b.Label == "" {
		return legendEntry{}, false
	}
	fill := b.Color
	if len(b.Colors) > 0 {
		fill = b.Colors[0]
	}
	edge := b.EdgeColor
	if len(b.EdgeColors) > 0 {
		edge = b.EdgeColors[0]
	}
	alpha := b.Alpha
	if alpha <= 0 {
		alpha = 1
	}
	fill = fill.WithAlphaMultiplier(alpha)
	edge = edge.WithAlphaMultiplier(alpha)
	return legendEntryFromPatchStyle(b.Label, fill, edge, b.EdgeWidth, "", render.Color{}, 0), true
}

func (f *Fill2D) legendEntry() (legendEntry, bool) {
	if f == nil || f.Label == "" {
		return legendEntry{}, false
	}
	fill := f.Color
	edge := f.EdgeColor
	if f.Alpha > 0 && f.Alpha <= 1 {
		fill = fill.WithAlphaMultiplier(f.Alpha)
		edge = edge.WithAlphaMultiplier(f.Alpha)
	}
	return legendEntryFromPatchStyle(f.Label, fill, edge, f.EdgeWidth, "", render.Color{}, 0), true
}

func (h *Hist2D) legendEntry() (legendEntry, bool) {
	if h == nil || h.Label == "" {
		return legendEntry{}, false
	}
	return legendEntryFromPatchStyle(h.Label, h.Color, h.EdgeColor, h.EdgeWidth, "", render.Color{}, 0), true
}

func (b *BoxPlot2D) legendEntry() (legendEntry, bool) {
	if b == nil || b.Label == "" {
		return legendEntry{}, false
	}
	return legendEntryFromPatchStyle(b.Label, b.Color, b.EdgeColor, b.EdgeWidth, "", render.Color{}, 0), true
}

func (i *Image2D) legendEntry() (legendEntry, bool) {
	if i == nil || i.Label == "" {
		return legendEntry{}, false
	}
	return legendEntryFromPatchStyle(
		i.Label,
		render.Color{R: 0.45, G: 0.45, B: 0.45, A: 1},
		render.Color{R: 0.2, G: 0.2, B: 0.2, A: 0.9},
		1,
		"",
		render.Color{},
		0,
	), true
}

func (e *ErrorBar) legendEntry() (legendEntry, bool) {
	if e == nil || e.Label == "" {
		return legendEntry{}, false
	}
	color := e.Color
	alpha := e.Alpha
	if alpha <= 0 {
		alpha = 1
	}
	if alpha > 1 {
		alpha = 1
	}
	color = color.WithAlphaMultiplier(alpha)
	entry := legendEntryFromLine(e.Label, color, e.LineWidth, nil)
	entry.kind = legendEntryErrorBar
	entry.errorbarX = errorbarHasX(e)
	entry.errorbarY = errorbarHasY(e)
	entry.errorbarCapSize = e.CapSize
	entry.errorbarCapWidth = 1.0
	if e.MarkerSet {
		markerLine := &Line2D{
			Marker:     e.Marker,
			MarkerSet:  true,
			MarkerSize: e.MarkerSize,
		}
		spec := markerLine.markerPathSpec(nil, nil)
		entry.lineMarkerSet = true
		entry.marker = e.Marker
		entry.markerStyle = markerLine.resolvedMarkerStyle()
		entry.markerPath = spec.Path
		entry.markerAltPath = spec.AltPath
		entry.markerEdgePath = spec.EdgePath
		entry.markerHasAlt = spec.HasAlt
		entry.markerLineOnly = markerLineOnly(markerLine.resolvedMarkerStyle())
		entry.markerFill = color
		entry.markerEdge = color
		entry.markerEdgeWidth = markerLine.resolvedMarkerEdgeWidth(nil)
		entry.errorbarCapWidth = entry.markerEdgeWidth
		entry.markerSize = markerLine.resolvedMarkerSize(nil)
		markerScatter := Scatter2D{Marker: e.Marker}
		entry.markerLineJoin = markerScatter.markerLineJoin()
		entry.markerLineCap = markerScatter.markerLineCap()
		entry.markerSnap = markerSnapMode(markerLine.resolvedMarkerStyle(), entry.markerSize)
	}
	return entry, true
}

func errorbarHasX(e *ErrorBar) bool {
	return len(e.XErr) > 0 || len(e.XErrLower) > 0 || len(e.XErrUpper) > 0 || len(e.XLoLimits) > 0 || len(e.XUpLimits) > 0
}

func errorbarHasY(e *ErrorBar) bool {
	return len(e.YErr) > 0 || len(e.YErrLower) > 0 || len(e.YErrUpper) > 0 || len(e.LoLimits) > 0 || len(e.UpLimits) > 0
}
