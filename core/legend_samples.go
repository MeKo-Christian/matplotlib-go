package core

import (
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

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
			l.drawMarkerSample(r, entry, pt, l.markerSampleScale(entry, 5), true)
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
			Snap:      render.SnapAuto,
		})
		if entry.lineMarkerSet {
			l.drawMarkerSample(r, entry, center, l.markerSampleScale(entry, 5), false)
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
		Snap:      render.SnapAuto,
	}
	capPaint := paint
	if entry.errorbarCapWidth > 0 {
		capPaint.LineWidth = entry.errorbarCapWidth
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
		l.drawMarkerSample(r, entry, center, l.markerSampleScale(entry, 5), false)
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

func (l *Legend) drawMarkerSample(r render.Renderer, entry legendEntry, center geom.Pt, radius float64, offsetCenter bool) {
	if offsetCenter {
		center.X += 0.5
		center.Y += 0.5
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
	if drawer, ok := r.(render.MarkerDrawer); ok {
		if drawer.DrawMarkers(render.MarkerBatch{
			Marker: markerPath,
			Items: []render.MarkerItem{{
				Offset:      center,
				Transform:   geom.Affine{A: scale, D: scale},
				Paint:       paint,
				Antialiased: true,
			}},
		}) {
			return
		}
	}
	path := scaleAndTranslatePath(markerPath, scale, center)
	r.Path(path, &paint)
}
