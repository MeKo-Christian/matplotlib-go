package core

import (
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func legendEntryFromPatchStyle(label string, face, edge render.Color, edgeWidth float64, hatch string, hatchColor render.Color, hatchWidth float64) legendEntry {
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
		dashes:    append([]float64(nil), dashes...),
	}
}

func legendEntryFromMarker(label string, marker MarkerType, markerPath geom.Path, fill, edge render.Color, edgeWidth float64) legendEntry {
	return legendEntry{
		Label:           label,
		kind:            legendEntryMarker,
		marker:          marker,
		markerPath:      markerPath,
		markerFill:      fill,
		markerEdge:      edge,
		markerEdgeWidth: edgeWidth,
	}
}

func (l *Line2D) legendEntry() (legendEntry, bool) {
	if l == nil || l.Label == "" {
		return legendEntry{}, false
	}
	entry := legendEntryFromLine(l.Label, l.ApplyArtistAlpha(l.Col), l.W, l.Dashes)
	if l.hasMarkers() {
		spec := l.markerPathSpec(nil, nil)
		entry.lineMarkerSet = true
		entry.marker = l.Marker
		entry.markerPath = spec.Path
		entry.markerAltPath = spec.AltPath
		entry.markerEdgePath = spec.EdgePath
		entry.markerHasAlt = spec.HasAlt
		entry.markerLineOnly = markerLineOnly(l.resolvedMarkerStyle())
		entry.markerFill = l.resolvedMarkerFaceColor()
		entry.markerAltFill = l.resolvedMarkerFaceColorAlt()
		entry.markerEdge = l.resolvedMarkerEdgeColor()
		entry.markerEdgeWidth = l.resolvedMarkerEdgeWidth(nil)
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
	return legendEntryFromMarker(s.Label, s.Marker, s.MarkerPath, fill, edge, s.EdgeWidth), true
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
		entry.lineMarkerSet = true
		entry.marker = e.Marker
		entry.markerFill = color
		entry.markerEdge = color
		entry.markerEdgeWidth = e.LineWidth
	}
	return entry, true
}

func errorbarHasX(e *ErrorBar) bool {
	return len(e.XErr) > 0 || len(e.XErrLower) > 0 || len(e.XErrUpper) > 0 || len(e.XLoLimits) > 0 || len(e.XUpLimits) > 0
}

func errorbarHasY(e *ErrorBar) bool {
	return len(e.YErr) > 0 || len(e.YErrLower) > 0 || len(e.YErrUpper) > 0 || len(e.LoLimits) > 0 || len(e.UpLimits) > 0
}
