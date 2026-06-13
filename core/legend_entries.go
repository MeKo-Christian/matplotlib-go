package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

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
	return legendEntry{
		Label:     label,
		kind:      legendEntryLine,
		lineColor: color,
		lineWidth: width,
		lineJoin:  render.JoinRound,
		lineCap:   render.CapButt,
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
	fill.A *= alpha
	edge.A *= alpha
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
	fill.A *= alpha
	edge.A *= alpha
	return legendEntryFromPatchStyle(b.Label, fill, edge, b.EdgeWidth, "", render.Color{}, 0), true
}

func (f *Fill2D) legendEntry() (legendEntry, bool) {
	if f == nil || f.Label == "" {
		return legendEntry{}, false
	}
	fill := f.Color
	edge := f.EdgeColor
	if f.Alpha > 0 && f.Alpha <= 1 {
		fill.A *= f.Alpha
		edge.A *= f.Alpha
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
	color.A *= alpha
	entry := legendEntryFromLine(e.Label, color, e.LineWidth, nil)
	entry.kind = legendEntryErrorBar
	entry.errorbarX = errorbarHasX(e)
	entry.errorbarY = errorbarHasY(e)
	entry.errorbarCapSize = e.CapSize
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
		entry.markerEdgeWidth = e.LineWidth
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
