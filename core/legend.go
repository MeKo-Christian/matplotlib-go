package core

import (
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
)

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
}

type legendEntryProvider interface {
	legendEntry() (legendEntry, bool)
}

// Legend renders a styled legend box inside an axes.
// If no explicit internal entries are present, labeled artists on the owning axes are collected automatically.
type Legend struct {
	Axes   *Axes
	Figure *Figure

	entries []legendEntry

	Location        LegendLocation
	Locator         AnchoredBoxLocator
	Padding         float64
	Inset           float64
	RowGap          float64
	SampleWidth     float64
	SampleTextGap   float64
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

	maxLabelWidth := 0.0
	rowHeights := make([]float64, len(entries))
	labelLayouts := make([]singleLineTextLayout, len(entries))
	for i, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		labelLayouts[i] = layout
		if layout.Width > maxLabelWidth {
			maxLabelWidth = layout.Width
		}
		rowHeights[i] = legendRowHeight(layout, fontSize, ctx)
	}

	contentHeight := 0.0
	for _, h := range rowHeights {
		contentHeight += h
	}
	if len(rowHeights) > 1 {
		contentHeight += l.RowGap * float64(len(rowHeights)-1)
	}

	boxWidth := l.Padding*2 + l.SampleWidth + l.SampleTextGap + maxLabelWidth
	boxHeight := l.Padding*2 + contentHeight
	box := l.legendBoxRect(ctx, boxWidth, boxHeight)

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

	y := box.Min.Y + l.Padding
	for i, entry := range entries {
		rowHeight := rowHeights[i]
		centerY := y + rowHeight/2
		labelLayout := labelLayouts[i]

		l.drawSample(r, entry, geom.Rect{
			Min: geom.Pt{X: box.Min.X + l.Padding, Y: centerY - rowHeight/2},
			Max: geom.Pt{X: box.Min.X + l.Padding + l.SampleWidth, Y: centerY + rowHeight/2},
		})

		drawDisplayText(
			textRen,
			entry.Label,
			alignedSingleLineOrigin(
				geom.Pt{X: box.Min.X + l.Padding + l.SampleWidth + l.SampleTextGap, Y: centerY},
				labelLayout,
				TextAlignLeft,
				textLayoutVAlignCenter,
			),
			fontSize,
			l.TextColor,
			ctx.RC.FontKey,
			ctx.RC.UseTeX,
		)

		y += rowHeight + l.RowGap
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

	maxLabelWidth := 0.0
	contentHeight := 0.0
	for _, entry := range entries {
		layout := measureSingleLineTextLayout(r, entry.Label, fontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
		if layout.Width > maxLabelWidth {
			maxLabelWidth = layout.Width
		}
		contentHeight += legendRowHeight(layout, fontSize, ctx)
	}
	if len(entries) > 1 {
		contentHeight += l.RowGap * float64(len(entries)-1)
	}

	boxWidth := l.Padding*2 + l.SampleWidth + l.SampleTextGap + maxLabelWidth
	boxHeight := l.Padding*2 + contentHeight
	return l.legendBoxRect(ctx, boxWidth, boxHeight), true
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
		return collectLegendEntries(l.Axes.Artists)
	case l.Figure != nil:
		var entries []legendEntry
		for _, ax := range l.Figure.Children {
			entries = append(entries, collectLegendEntries(ax.Artists)...)
		}
		return entries
	default:
		return nil
	}
}

func collectLegendEntries(artists []Artist) []legendEntry {
	entries := make([]legendEntry, 0, len(artists))
	for _, art := range artists {
		switch art.(type) {
		case *Legend:
			continue
		default:
			provider, ok := art.(legendEntryProvider)
			if !ok {
				continue
			}
			label := ArtistLabel(art)
			if !legendLabelVisible(label) {
				continue
			}
			entry, ok := provider.legendEntry()
			if !ok {
				continue
			}
			entry.Label = label
			entries = append(entries, entry)
		}
	}
	return entries
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

func (l *Legend) drawSample(r render.Renderer, entry legendEntry, sample geom.Rect) {
	center := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}

	switch entry.kind {
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
		l.drawMarkerSample(r, entry, center, 5)
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
			l.drawMarkerSample(r, entry, center, 5)
		}
	}
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
