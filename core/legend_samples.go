package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func (l *Legend) drawSample(r render.Renderer, entry legendEntry, sample geom.Rect) {
	fontSize := style.Default.LegendSize()
	if l != nil && l.FontSize > 0 {
		fontSize = l.FontSize
	}
	l.drawSampleWithFontPixels(r, &style.Default, &entry, sample, pointsToPixels(style.Default, fontSize), false)
}

func (l *Legend) drawSampleWithFontPixels(r render.Renderer, rc *style.RC, entry *legendEntry, sample geom.Rect, fontPx float64, sampleIsHandleBox bool) {
	center := geom.Pt{
		X: sample.Min.X + sample.W()/2,
		Y: sample.Min.Y + sample.H()/2,
	}

	switch entry.kind {
	case legendEntryErrorBar:
		l.drawErrorBarSample(r, rc, entry, sample, center, fontPx)
	case legendEntryPatch:
		// Matplotlib's HandlerPatch fills the legend handle box. The
		// handleheight is already reflected in the supplied handle box.
		patchRect := sample
		if !sampleIsHandleBox {
			handleHeight := sample.H() * 0.7
			patchRect.Min.Y = center.Y - handleHeight/2
			patchRect.Max.Y = center.Y + handleHeight/2
		}
		patch := Patch{
			FaceColor:  optional.Of(entry.patchFill),
			EdgeColor:  optional.Of(entry.patchEdge),
			EdgeWidth:  optional.Of(entry.patchEdgeWidth),
			Hatch:      entry.patchHatch,
			HatchColor: entry.patchHatchColor,
			HatchWidth: entry.patchHatchWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		}
		patch.drawStyledPath(r, rc, pixelRectPath(patchRect), geom.Path{})
	case legendEntryMarker:
		centers := l.markerSampleCentersForHandle(sample, center, sampleIsHandleBox)
		scales := l.collectionSampleScales(entry, len(centers))
		for i, pt := range centers {
			l.drawMarkerSample(r, rc, entry, pt, scales[i], legendMarkerCollection)
		}
	default:
		lineWidth := entry.lineWidth
		if lineWidth <= 0 {
			lineWidth = 1.5
		}
		lineMarkerCenters := l.lineMarkerSampleCenters(sample, center, fontPx)
		lineStart := geom.Pt{X: sample.Min.X, Y: center.Y}
		lineEnd := geom.Pt{X: sample.Max.X, Y: center.Y}
		if len(lineMarkerCenters) > 1 {
			lineStart.X = lineMarkerCenters[0].X
			lineEnd.X = lineMarkerCenters[len(lineMarkerCenters)-1].X
		}
		path := geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{
				lineStart,
				lineEnd,
			},
		}
		r.Path(path, &render.Paint{
			Stroke:    entry.lineColor,
			LineWidth: pointsToPixels(*rc, lineWidth),
			LineJoin:  entry.lineJoin,
			LineCap:   entry.lineCap,
			Dashes:    entry.dashes,
			Snap:      render.SnapAuto,
		})
		if entry.lineMarkerSet {
			for _, pt := range lineMarkerCenters {
				l.drawMarkerSample(r, rc, entry, pt, l.markerSampleScale(*entry, 5), legendMarkerStamp)
			}
		}
	}
}

func (l *Legend) drawErrorBarSample(r render.Renderer, rc *style.RC, entry *legendEntry, sample geom.Rect, center geom.Pt, fontPx float64) {
	lineWidth := entry.lineWidth
	if lineWidth <= 0 {
		lineWidth = 1.5
	}
	paint := render.Paint{
		Stroke:    entry.lineColor,
		LineWidth: pointsToPixels(*rc, lineWidth),
		LineJoin:  render.JoinMiter,
		LineCap:   render.CapButt,
		Dashes:    entry.dashes,
		Snap:      render.SnapAuto,
	}
	capPaint := paint
	if entry.errorbarCapWidth > 0 {
		capPaint.LineWidth = pointsToPixels(*rc, entry.errorbarCapWidth)
	}
	r.Path(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: sample.Min.X, Y: center.Y}, {X: sample.Max.X, Y: center.Y}},
	}, &paint)

	capHalf := entry.errorbarCapSize / 2
	if capHalf <= 0 {
		capHalf = 3
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
		}, &capPaint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: bottom.X - capHalf, Y: bottom.Y}, {X: bottom.X + capHalf, Y: bottom.Y}},
		}, &capPaint)
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
		}, &capPaint)
		r.Path(geom.Path{
			C: []geom.Cmd{geom.MoveTo, geom.LineTo},
			V: []geom.Pt{{X: right.X, Y: right.Y - capHalf}, {X: right.X, Y: right.Y + capHalf}},
		}, &capPaint)
	}
	if entry.lineMarkerSet {
		for _, pt := range l.lineMarkerSampleCenters(sample, center, fontPx) {
			l.drawMarkerSample(r, rc, entry, pt, l.markerSampleScale(*entry, 5), legendMarkerStamp)
		}
	}
}

func (l *Legend) markerSampleScale(entry legendEntry, base float64) float64 {
	scale := base
	if entry.markerSize > 0 {
		scale = entry.markerSize
	}
	markerScale := 1.0
	if l != nil && (l.MarkerScale > 0 || l.defaultsSet) {
		markerScale = l.MarkerScale
	}
	return scale * markerScale
}

// collectionSampleScales sizes each of a collection handle's sample points,
// porting HandlerRegularPolyCollection.get_sizes: fewer than four points take
// the mean, largest and smallest area in that order, more spread linearly over
// the area range. The port works in linear path scale rather than area, so
// every area is carried as its square root; markerscale multiplies the scale,
// which is Matplotlib's markerscale**2 on the area.
func (l *Legend) collectionSampleScales(entry *legendEntry, points int) []float64 {
	if points <= 0 {
		return nil
	}
	scales := make([]float64, points)
	minScale, maxScale := entry.markerScaleMin, entry.markerScaleMax
	if minScale <= 0 || maxScale <= 0 {
		fallback := l.markerSampleScale(*entry, 5)
		for i := range scales {
			scales[i] = fallback
		}
		return scales
	}
	markerScale := 1.0
	if l != nil && (l.MarkerScale > 0 || l.defaultsSet) {
		markerScale = l.MarkerScale
	}
	minArea, maxArea := minScale*minScale, maxScale*maxScale
	for i := range scales {
		var area float64
		switch {
		case points < 4:
			area = [...]float64{0.5 * (maxArea + minArea), maxArea, minArea}[i]
		default:
			area = minArea + (maxArea-minArea)*float64(i)/float64(points-1)
		}
		scales[i] = math.Sqrt(area) * markerScale
	}
	return scales
}

func (l *Legend) markerSampleCenters(sample geom.Rect, center geom.Pt) []geom.Pt {
	return l.markerSampleCentersForHandle(sample, center, false)
}

func (l *Legend) markerSampleCentersForHandle(sample geom.Rect, center geom.Pt, sampleIsHandleBox bool) []geom.Pt {
	points := 1
	if l != nil && (l.ScatterPoints > 0 || l.defaultsSet) {
		points = l.ScatterPoints
	}
	if points <= 0 {
		return nil
	}
	// Matplotlib's HandlerPathCollection uses Legend._scatteryoffsets
	// [3/8, 4/8, 2.5/8], tiled and truncated to scatterpoints, as fractions of
	// the handle box height measured from its bottom edge
	// (HandlerNpointsYoffsets.get_ydata). A single sample point takes 3/8, not
	// the box centre.
	offsets := [...]float64{3.0 / 8.0, 4.0 / 8.0, 2.5 / 8.0}
	handleHeight := sample.H()
	if !sampleIsHandleBox {
		handleHeight *= 0.7
	}
	sampleY := func(i int) float64 {
		return center.Y - handleHeight/2 + offsets[i%len(offsets)]*handleHeight
	}

	// HandlerNpoints.get_xdata centres a lone sample in the handle box and
	// otherwise spreads scatterpoints over linspace(pad, width-pad), with
	// pad = 0.3*fontsize = 0.15*width at the default handlelength of 2.
	if points == 1 {
		return []geom.Pt{{X: center.X, Y: sampleY(0)}}
	}
	centers := make([]geom.Pt, points)
	pad := sample.W() * 0.15
	if pad < 0 {
		pad = 0
	}
	step := (sample.W() - 2*pad) / float64(points-1)
	for i := 0; i < points; i++ {
		centers[i] = geom.Pt{
			X: sample.Min.X + pad + step*float64(i),
			Y: sampleY(i),
		}
	}
	return centers
}

func (l *Legend) lineMarkerSampleCenters(sample geom.Rect, center geom.Pt, fontPx float64) []geom.Pt {
	points := 1
	if l != nil && (l.NumPoints > 0 || l.defaultsSet) {
		points = l.NumPoints
	}
	if points <= 0 {
		return nil
	}
	if points == 1 {
		return []geom.Pt{center}
	}
	pad := 0.3 * fontPx
	if pad*2 > sample.W() {
		pad = sample.W() / 2
	}
	centers := make([]geom.Pt, points)
	step := 0.0
	if points > 1 {
		step = (sample.W() - 2*pad) / float64(points-1)
	}
	for i := range centers {
		centers[i] = geom.Pt{X: sample.Min.X + pad + step*float64(i), Y: center.Y}
	}
	return centers
}

// legendMarkerRoute names the Matplotlib draw call a legend handle's markers
// imitate. The two differ in how they place an offset, so the choice is worth
// one pixel: `draw_markers` (`_backend_agg.h`) rasterizes the marker once and
// stamps it at `floor(offset)` after a `translate(0.5, height+0.5)`, while
// `_draw_path_collection_generic` translates the path by the raw offset and
// flips with `translate(0, height)`, keeping the subpixel part.
type legendMarkerRoute uint8

const (
	// legendMarkerStamp mirrors draw_markers, used for Line2D-derived handles
	// (HandlerLine2D, HandlerErrorbar).
	legendMarkerStamp legendMarkerRoute = iota
	// legendMarkerCollection mirrors draw_path_collection, used for the
	// collection handles HandlerPathCollection builds from a scatter.
	legendMarkerCollection
)

func (l *Legend) drawMarkerSample(r render.Renderer, rc *style.RC, entry *legendEntry, center geom.Pt, radius float64, route legendMarkerRoute) {
	markerEdgeWidthPx := pointsToPixels(*rc, entry.markerEdgeWidth)
	snap := entry.markerSnap
	if route == legendMarkerCollection {
		// Collections carry Artist.get_snap() == None, i.e. SNAP_AUTO; the
		// marker-size threshold table has no Matplotlib counterpart.
		snap = render.SnapAuto
	}
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
		drawLegendMarkerPath(r, markerPath, center, markerScale, snap, route, &render.Paint{
			Fill:      entry.markerFill,
			Stroke:    entry.markerEdge,
			LineWidth: markerEdgeWidthPx,
			LineJoin:  lineJoin,
			LineCap:   lineCap,
		})
		if len(entry.markerAltPath.C) > 0 {
			drawLegendMarkerPath(r, entry.markerAltPath, center, radius, snap, route, &render.Paint{
				Fill:      entry.markerAltFill,
				Stroke:    entry.markerEdge,
				LineWidth: markerEdgeWidthPx,
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
	drawLegendMarkerPath(r, markerPath, center, markerScale, snap, route, &render.Paint{
		Fill:      fill,
		Stroke:    edge,
		LineWidth: markerEdgeWidthPx,
		LineJoin:  lineJoin,
		LineCap:   lineCap,
	})
}

func drawLegendMarkerPath(r render.Renderer, markerPath geom.Path, center geom.Pt, scale float64, snap render.SnapMode, route legendMarkerRoute, paint *render.Paint) {
	if len(markerPath.C) == 0 || scale <= 0 {
		return
	}
	paint.Snap = snap
	if drawer, ok := r.(render.MarkerDrawer); ok && route == legendMarkerStamp {
		if drawer.DrawMarkers(render.MarkerBatch{
			Marker: markerPath,
			Items: []render.MarkerItem{{
				Offset:      center,
				Transform:   geom.Affine{A: scale, D: scale},
				Paint:       *paint,
				Antialiased: true,
			}},
		}) {
			return
		}
	}
	path := scaleAndTranslatePath(markerPath, scale, center)
	r.Path(path, paint)
}
