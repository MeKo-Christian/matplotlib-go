package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// MarkerType defines the shape of markers in scatter plots.
type MarkerType uint8

const (
	MarkerCircle MarkerType = iota
	MarkerSquare
	MarkerTriangle
	MarkerDiamond
	MarkerPlus
	MarkerCross
	MarkerPixel
	MarkerPoint
	MarkerTriangleDown
	MarkerTriangleLeft
	MarkerTriangleRight
	MarkerTriDown
	MarkerTriUp
	MarkerTriLeft
	MarkerTriRight
	MarkerOctagon
	MarkerPentagon
	MarkerStar
	MarkerHexagon1
	MarkerHexagon2
	MarkerFilledX
	MarkerFilledPlus
	MarkerThinDiamond
	MarkerVLine
	MarkerHLine
	MarkerTickLeft
	MarkerTickRight
	MarkerTickUp
	MarkerTickDown
	MarkerCaretLeft
	MarkerCaretRight
	MarkerCaretUp
	MarkerCaretDown
	MarkerCaretLeftBase
	MarkerCaretRightBase
	MarkerCaretUpBase
	MarkerCaretDownBase
	MarkerNone
)

const MarkerTriangleUp = MarkerTriangle

// MarkerFillStyle controls marker filling. Full and none use the normal marker
// path; half-fill styles split filled marker paths for alternate-side drawing.
type MarkerFillStyle uint8

const (
	MarkerFillFull MarkerFillStyle = iota
	MarkerFillLeft
	MarkerFillRight
	MarkerFillBottom
	MarkerFillTop
	MarkerFillNone
)

// MarkerTupleStyle mirrors Matplotlib's tuple marker style selector.
type MarkerTupleStyle uint8

const (
	MarkerTuplePolygon MarkerTupleStyle = iota
	MarkerTupleStar
	MarkerTupleAsterisk
)

// MarkerStyle describes a marker independently from Scatter2D's legacy
// MarkerType field. It supports built-in markers, mathtext/text markers, and
// Matplotlib's (numsides, style, angle) tuple markers.
type MarkerStyle struct {
	Type      MarkerType
	FillStyle MarkerFillStyle
	MathText  string
	Tuple     *MarkerTuple
	Path      geom.Path
}

// MarkerTuple describes a Matplotlib tuple marker.
type MarkerTuple struct {
	NumSides int
	Style    MarkerTupleStyle
	AngleDeg float64
}

// NewMarkerStyle returns a full-fill built-in marker style.
func NewMarkerStyle(marker MarkerType) MarkerStyle {
	return MarkerStyle{Type: marker, FillStyle: MarkerFillFull}
}

// NewTupleMarkerStyle returns a tuple marker style equivalent to
// (numsides, style, angle) in Matplotlib.
func NewTupleMarkerStyle(numsides int, style MarkerTupleStyle, angleDeg float64) MarkerStyle {
	return MarkerStyle{
		FillStyle: MarkerFillFull,
		Tuple:     &MarkerTuple{NumSides: numsides, Style: style, AngleDeg: angleDeg},
	}
}

// NewMathTextMarkerStyle returns a marker style rendered from a text or
// MathText expression. Dollar-delimited expressions use the MathText layout
// pipeline when renderer metrics are available.
func NewMathTextMarkerStyle(text string) MarkerStyle {
	return MarkerStyle{FillStyle: MarkerFillFull, MathText: text}
}

// MarkerTypeFromString resolves Matplotlib's string marker aliases.
func MarkerTypeFromString(marker string) (MarkerType, bool) {
	switch marker {
	case "o":
		return MarkerCircle, true
	case "s":
		return MarkerSquare, true
	case "^":
		return MarkerTriangleUp, true
	case "v":
		return MarkerTriangleDown, true
	case "<":
		return MarkerTriangleLeft, true
	case ">":
		return MarkerTriangleRight, true
	case "D":
		return MarkerDiamond, true
	case "d":
		return MarkerThinDiamond, true
	case "+":
		return MarkerPlus, true
	case "x":
		return MarkerCross, true
	case ",":
		return MarkerPixel, true
	case ".":
		return MarkerPoint, true
	case "1":
		return MarkerTriDown, true
	case "2":
		return MarkerTriUp, true
	case "3":
		return MarkerTriLeft, true
	case "4":
		return MarkerTriRight, true
	case "8":
		return MarkerOctagon, true
	case "p":
		return MarkerPentagon, true
	case "*":
		return MarkerStar, true
	case "h":
		return MarkerHexagon1, true
	case "H":
		return MarkerHexagon2, true
	case "X":
		return MarkerFilledX, true
	case "P":
		return MarkerFilledPlus, true
	case "|":
		return MarkerVLine, true
	case "_":
		return MarkerHLine, true
	case "", " ", "none", "None":
		return MarkerNone, true
	default:
		return MarkerCircle, false
	}
}

// Scatter2D renders points with configurable markers.
type Scatter2D struct {
	ArtistRasterization
	XY             []geom.Pt      // data space points
	Sizes          []float64      // marker areas in points^2, if nil uses Size
	Colors         []render.Color // marker colors, if nil uses Color
	EdgeColors     []render.Color // edge colors for marker outlines, if nil uses EdgeColor
	ScalarValues   []float64      // scalar values mapped to marker face colors
	MarkerPath     geom.Path      // optional custom marker path in normalized collection space
	Size           float64        // default marker area in points^2
	Color          render.Color   // default marker color
	EdgeColor      render.Color   // default edge color for marker outlines
	EdgeWidth      float64        // edge width in pixels (0 means no edge)
	Colormap       string
	Norm           ScalarNormalizer
	VMin           float64
	VMax           float64
	EdgeColorsFace bool
	scalarCLimSet  bool
	Alpha          float64 // alpha transparency (0-1), applied to both fill and edge
	PathEffects    []render.PathEffect
	Marker         MarkerType // marker shape
	MarkerStyle    MarkerStyle
	Label          string  // series label for legend
	z              float64 // z-order
}

// ScalarMap exposes scatter scalar mapping for colorbars.
func (s *Scatter2D) ScalarMap() ScalarMapInfo {
	if s == nil {
		return ScalarMapInfo{}
	}
	return ScalarMapInfo{
		Colormap: s.Colormap,
		VMin:     s.VMin,
		VMax:     s.VMax,
		Norm:     s.Norm,
	}
}

// GetArray returns a copy of the scatter scalar array, matching Matplotlib's
// scalar-mappable PathCollection surface.
func (s *Scatter2D) GetArray() []float64 {
	if s == nil || len(s.ScalarValues) == 0 {
		return nil
	}
	return append([]float64(nil), s.ScalarValues...)
}

var stemMarkerScale = math.Sqrt(math.Pi)

// Draw renders scatter points by creating filled paths for each marker.
func (s *Scatter2D) Draw(r render.Renderer, ctx *DrawContext) {
	if s == nil || len(s.XY) == 0 {
		return
	}
	if s.drawHalfFilledMarkers(r, ctx) {
		return
	}
	s.toPathCollection(r, ctx).Draw(r, ctx)
}

func (s *Scatter2D) drawHalfFilledMarkers(r render.Renderer, ctx *DrawContext) bool {
	if s == nil || r == nil || ctx == nil {
		return false
	}
	style := s.resolvedMarkerStyle()
	if markerLineOnly(style) {
		return false
	}
	basePath := s.markerPrototypePathForContext(r, ctx)
	primaryPath, altPath, ok := splitMarkerFillPaths(style, basePath)
	if !ok {
		return false
	}
	transparent := render.Color{}

	primary := s.toPathCollection(r, ctx)
	primary.Path = primaryPath
	primary.EdgeColor = transparent
	primary.EdgeColors = nil
	primary.EdgeWidth = 0
	primary.EdgeWidths = nil
	primary.Draw(r, ctx)

	alt := s.toPathCollection(r, ctx)
	alt.Path = altPath
	alt.FaceColor = transparent
	alt.FaceColors = nil
	alt.EdgeColor = transparent
	alt.EdgeColors = nil
	alt.EdgeWidth = 0
	alt.EdgeWidths = nil
	alt.Draw(r, ctx)

	edge := s.toPathCollection(r, ctx)
	edge.Path = basePath
	edge.FaceColor = transparent
	edge.FaceColors = nil
	edge.Draw(r, ctx)
	return true
}

// createMarkerPath creates a filled path for the given marker type at the specified position and size.
func (s *Scatter2D) createMarkerPath(center geom.Pt, radius float64) geom.Path {
	if radius <= 0 {
		return geom.Path{}
	}
	return scaleAndTranslatePath(s.markerPrototypePath(), radius*stemMarkerScale, center)
}

func (s *Scatter2D) markerPrototypePath() geom.Path {
	return s.markerPrototypePathForContext(nil, nil)
}

func (s *Scatter2D) markerPrototypePathForContext(r render.Renderer, ctx *DrawContext) geom.Path {
	if len(s.MarkerPath.C) > 0 {
		return normalizeCustomMarkerPath(s.MarkerPath)
	}

	style := s.resolvedMarkerStyle()
	if len(style.Path.C) > 0 {
		return normalizeCustomMarkerPath(style.Path)
	}
	if style.Tuple != nil {
		return markerTuplePath(*style.Tuple)
	}
	if style.MathText != "" {
		fontKey := "DejaVu Sans"
		if ctx != nil && ctx.RC.FontKey != "" {
			fontKey = ctx.RC.FontKey
		}
		if path, ok := mathTextMarkerPath(r, style.MathText, fontKey); ok {
			return path
		}
	}

	switch style.Type {
	case MarkerCircle:
		return markerCirclePath(1)
	case MarkerPoint:
		return markerCirclePath(0.5)
	case MarkerPixel:
		return markerRectanglePath(-0.5, -0.5, 0.5, 0.5)
	case MarkerSquare:
		return markerRectanglePath(-0.5, -0.5, 0.5, 0.5)
	case MarkerTriangle:
		return markerTrianglePath(0)
	case MarkerTriangleDown:
		return markerTrianglePath(180)
	case MarkerTriangleLeft:
		return markerTrianglePath(90)
	case MarkerTriangleRight:
		return markerTrianglePath(270)
	case MarkerDiamond:
		return s.createDiamondPath(geom.Pt{}, 1)
	case MarkerThinDiamond:
		return applyAffinePath(s.createDiamondPath(geom.Pt{}, 1), geom.Affine{A: 0.6, D: 1})
	case MarkerPlus:
		return s.createPlusPath(geom.Pt{}, 1)
	case MarkerCross:
		return s.createCrossPath(geom.Pt{}, 1)
	case MarkerTriDown:
		return markerTriPath(0)
	case MarkerTriUp:
		return markerTriPath(180)
	case MarkerTriLeft:
		return markerTriPath(270)
	case MarkerTriRight:
		return markerTriPath(90)
	case MarkerOctagon:
		return markerRegularPolygonPath(8, 90+22.5)
	case MarkerPentagon:
		return markerRegularPolygonPath(5, 90)
	case MarkerStar:
		return markerStarPath(5, 0.381966, 90, true)
	case MarkerHexagon1:
		return markerRegularPolygonPath(6, 90)
	case MarkerHexagon2:
		return markerRegularPolygonPath(6, 120)
	case MarkerFilledX:
		return markerFilledXPath()
	case MarkerFilledPlus:
		return markerFilledPlusPath()
	case MarkerVLine:
		return markerLinePath(0, -0.5, 0, 0.5)
	case MarkerHLine:
		return markerLinePath(-0.5, 0, 0.5, 0)
	case MarkerTickLeft:
		return markerLinePath(0, 0, -1, 0)
	case MarkerTickRight:
		return markerLinePath(0, 0, 1, 0)
	case MarkerTickUp:
		return markerLinePath(0, 0, 0, -1)
	case MarkerTickDown:
		return markerLinePath(0, 0, 0, 1)
	case MarkerCaretLeft:
		return markerCaretPath(270, false)
	case MarkerCaretRight:
		return markerCaretPath(90, false)
	case MarkerCaretUp:
		return markerCaretPath(180, false)
	case MarkerCaretDown:
		return markerCaretPath(0, false)
	case MarkerCaretLeftBase:
		return markerCaretPath(270, true)
	case MarkerCaretRightBase:
		return markerCaretPath(90, true)
	case MarkerCaretUpBase:
		return markerCaretPath(180, true)
	case MarkerCaretDownBase:
		return markerCaretPath(0, true)
	case MarkerNone:
		return geom.Path{}
	default:
		return markerCirclePath(1)
	}
}

func (s *Scatter2D) resolvedMarkerStyle() MarkerStyle {
	if s.MarkerStyle.Tuple != nil || s.MarkerStyle.MathText != "" || len(s.MarkerStyle.Path.C) > 0 || s.MarkerStyle.Type != 0 || s.MarkerStyle.FillStyle != 0 {
		style := s.MarkerStyle
		if style.FillStyle == 0 {
			style.FillStyle = MarkerFillFull
		}
		return style
	}
	return MarkerStyle{Type: s.Marker, FillStyle: MarkerFillFull}
}

func (s *Scatter2D) toPathCollection(r render.Renderer, ctx *DrawContext) *PathCollection {
	alpha := s.Alpha
	alpha = s.EffectiveAlpha(alpha)

	style := s.resolvedMarkerStyle()
	lineOnly := markerLineOnly(style)
	lineWidth := s.EdgeWidth
	if lineOnly && lineWidth <= 0 && ctx != nil && ctx.RC.LineWidth > 0 {
		lineWidth = ctx.RC.LineWidth
	}

	size := scatterAreaScale(s.Size, ctx)
	var sizes []float64
	if len(s.Sizes) > 0 {
		sizes = make([]float64, len(s.XY))
		for i := range s.XY {
			itemSize := s.Size
			if i < len(s.Sizes) {
				itemSize = s.Sizes[i]
			}
			sizes[i] = scatterAreaScale(itemSize, ctx)
		}
	}

	faceColor := s.Color
	edgeColor := s.EdgeColor
	if lineOnly && edgeColor.A <= 0 {
		edgeColor = faceColor
	}
	if style.FillStyle == MarkerFillNone {
		if edgeColor.A <= 0 {
			edgeColor = faceColor
		}
		faceColor.A = 0
	}

	pc := &PathCollection{
		Collection: Collection{
			ArtistRasterization: s.ArtistRasterization,
			Label:               s.Label,
			Alpha:               alpha,
			z:                   s.z,
			Colormap:            s.Colormap,
			Norm:                s.Norm,
			VMin:                s.VMin,
			VMax:                s.VMax,
			EdgeColorsFace:      s.EdgeColorsFace,
			PathEffects:         cloneRenderPathEffects(s.PathEffects),
			scalarCLimSet:       s.scalarCLimSet,
		},
		Path:          s.markerPrototypePathForContext(r, ctx),
		Offsets:       append([]geom.Pt(nil), s.XY...),
		Size:          size,
		Sizes:         sizes,
		PathInDisplay: true,
		FaceColors:    append([]render.Color(nil), s.Colors...),
		FaceColor:     faceColor,
		EdgeColors:    append([]render.Color(nil), s.EdgeColors...),
		EdgeColor:     edgeColor,
		EdgeWidth:     lineWidth,
		LineJoin:      s.markerLineJoin(),
		LineJoinSet:   true,
		LineCap:       s.markerLineCap(),
		LineCapSet:    true,
		LineOnly:      lineOnly,
	}
	if s.EdgeColorsFace && len(pc.FaceColors) > 0 {
		pc.EdgeColors = cloneRenderColors(pc.FaceColors)
	}
	if len(s.ScalarValues) > 0 {
		_ = pc.SetArray(s.ScalarValues)
		if len(s.Colors) > 0 {
			pc.FaceColors = append([]render.Color(nil), s.Colors...)
			if pc.EdgeColorsFace {
				pc.EdgeColors = cloneRenderColors(pc.FaceColors)
			}
		}
		if len(s.EdgeColors) > 0 {
			pc.EdgeColors = append([]render.Color(nil), s.EdgeColors...)
		}
	}
	return pc
}

func (s *Scatter2D) markerLineJoin() render.LineJoin {
	style := s.resolvedMarkerStyle()
	if style.Tuple != nil {
		if style.Tuple.Style == MarkerTupleStar || style.Tuple.Style == MarkerTupleAsterisk {
			return render.JoinBevel
		}
		return render.JoinMiter
	}
	switch style.Type {
	case MarkerSquare, MarkerTriangle, MarkerTriangleDown, MarkerTriangleLeft, MarkerTriangleRight,
		MarkerDiamond, MarkerThinDiamond, MarkerOctagon, MarkerPentagon, MarkerHexagon1, MarkerHexagon2,
		MarkerFilledPlus, MarkerFilledX,
		MarkerCaretLeft, MarkerCaretRight, MarkerCaretUp, MarkerCaretDown,
		MarkerCaretLeftBase, MarkerCaretRightBase, MarkerCaretUpBase, MarkerCaretDownBase:
		return render.JoinMiter
	case MarkerStar:
		return render.JoinBevel
	default:
		return render.JoinRound
	}
}

func (s *Scatter2D) markerLineCap() render.LineCap {
	return render.CapButt
}

func markerSnapMode(style MarkerStyle, markerSizePx float64) render.SnapMode {
	threshold, ok := markerSnapThreshold(style)
	if !ok || markerSizePx < threshold {
		return 0
	}
	return render.SnapAuto
}

func markerSnapThreshold(style MarkerStyle) (float64, bool) {
	if style.Tuple != nil || style.MathText != "" || len(style.Path.C) > 0 {
		return 0, false
	}
	switch style.Type {
	case MarkerCircle, MarkerPoint:
		return math.Inf(1), true
	case MarkerSquare:
		return 2, true
	case MarkerTriangle, MarkerTriangleDown, MarkerTriangleLeft, MarkerTriangleRight,
		MarkerDiamond, MarkerThinDiamond, MarkerPentagon, MarkerStar, MarkerOctagon,
		MarkerFilledPlus, MarkerFilledX:
		return 5, true
	case MarkerPlus, MarkerVLine, MarkerHLine, MarkerTickLeft, MarkerTickRight, MarkerTickUp, MarkerTickDown:
		return 1, true
	case MarkerCross, MarkerCaretLeft, MarkerCaretRight, MarkerCaretUp, MarkerCaretDown,
		MarkerCaretLeftBase, MarkerCaretRightBase, MarkerCaretUpBase, MarkerCaretDownBase:
		return 3, true
	default:
		return 0, false
	}
}

func markerLineOnly(style MarkerStyle) bool {
	if style.Tuple != nil {
		return style.Tuple.Style == MarkerTupleAsterisk
	}
	switch style.Type {
	case MarkerPlus, MarkerCross, MarkerTriDown, MarkerTriUp, MarkerTriLeft, MarkerTriRight,
		MarkerVLine, MarkerHLine, MarkerTickLeft, MarkerTickRight, MarkerTickUp, MarkerTickDown:
		return true
	default:
		return false
	}
}

func splitMarkerFillPaths(style MarkerStyle, full geom.Path) (geom.Path, geom.Path, bool) {
	if !markerHalfFilled(style.FillStyle) || len(full.C) == 0 || markerLineOnly(style) {
		return geom.Path{}, geom.Path{}, false
	}
	if style.Type == MarkerCircle || style.Type == MarkerPoint {
		scale := 1.0
		if style.Type == MarkerPoint {
			scale = 0.5
		}
		return splitCircleMarkerPath(style.FillStyle, scale), splitCircleMarkerPath(oppositeMarkerFill(style.FillStyle), scale), true
	}
	points, ok := closedPathPolygon(full)
	if !ok {
		return geom.Path{}, geom.Path{}, false
	}
	primary := clipMarkerPolygon(points, style.FillStyle)
	alt := clipMarkerPolygon(points, oppositeMarkerFill(style.FillStyle))
	if len(primary) < 3 || len(alt) < 3 {
		return geom.Path{}, geom.Path{}, false
	}
	return polygonPath(primary, true), polygonPath(alt, true), true
}

func markerHalfFilled(fill MarkerFillStyle) bool {
	switch fill {
	case MarkerFillLeft, MarkerFillRight, MarkerFillBottom, MarkerFillTop:
		return true
	default:
		return false
	}
}

func oppositeMarkerFill(fill MarkerFillStyle) MarkerFillStyle {
	switch fill {
	case MarkerFillLeft:
		return MarkerFillRight
	case MarkerFillRight:
		return MarkerFillLeft
	case MarkerFillBottom:
		return MarkerFillTop
	case MarkerFillTop:
		return MarkerFillBottom
	default:
		return fill
	}
}

func splitCircleMarkerPath(fill MarkerFillStyle, size float64) geom.Path {
	r := 0.5 * size
	if r <= 0 {
		return geom.Path{}
	}
	const magic = 0.2652031
	sqrtHalf := math.Sqrt(0.5)
	magic45 := sqrtHalf * magic

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: -r})
	path.CubicTo(
		geom.Pt{X: magic * r, Y: -r},
		geom.Pt{X: (sqrtHalf - magic45) * r, Y: -(sqrtHalf + magic45) * r},
		geom.Pt{X: sqrtHalf * r, Y: -sqrtHalf * r},
	)
	path.CubicTo(
		geom.Pt{X: (sqrtHalf + magic45) * r, Y: -(sqrtHalf - magic45) * r},
		geom.Pt{X: r, Y: -magic * r},
		geom.Pt{X: r, Y: 0},
	)
	path.CubicTo(
		geom.Pt{X: r, Y: magic * r},
		geom.Pt{X: (sqrtHalf + magic45) * r, Y: (sqrtHalf - magic45) * r},
		geom.Pt{X: sqrtHalf * r, Y: sqrtHalf * r},
	)
	path.CubicTo(
		geom.Pt{X: (sqrtHalf - magic45) * r, Y: (sqrtHalf + magic45) * r},
		geom.Pt{X: magic * r, Y: r},
		geom.Pt{X: 0, Y: r},
	)
	path.Close()

	angle := 0.0
	switch fill {
	case MarkerFillRight:
		angle = 0
	case MarkerFillLeft:
		angle = math.Pi
	case MarkerFillTop:
		angle = math.Pi / 2
	case MarkerFillBottom:
		angle = 3 * math.Pi / 2
	default:
		return geom.Path{}
	}
	if angle == 0 {
		return path
	}
	return applyAffinePath(path, geom.Affine{
		A: math.Cos(angle),
		B: math.Sin(angle),
		C: -math.Sin(angle),
		D: math.Cos(angle),
	})
}

func closedPathPolygon(path geom.Path) ([]geom.Pt, bool) {
	if len(path.C) == 0 || len(path.V) == 0 {
		return nil, false
	}
	points := make([]geom.Pt, 0, len(path.V))
	vi := 0
	for _, cmd := range path.C {
		switch cmd {
		case geom.MoveTo:
			if vi != 0 || vi >= len(path.V) {
				return nil, false
			}
			points = append(points, path.V[vi])
			vi++
		case geom.LineTo:
			if vi >= len(path.V) {
				return nil, false
			}
			points = append(points, path.V[vi])
			vi++
		case geom.ClosePath:
		default:
			return nil, false
		}
	}
	if len(points) > 1 && points[0] == points[len(points)-1] {
		points = points[:len(points)-1]
	}
	return points, len(points) >= 3
}

func clipMarkerPolygon(points []geom.Pt, fill MarkerFillStyle) []geom.Pt {
	if len(points) == 0 {
		return nil
	}
	var bounds geom.Rect
	for i, pt := range points {
		if i == 0 {
			bounds = geom.Rect{Min: pt, Max: pt}
		} else {
			bounds = expandRect(bounds, pt)
		}
	}
	cx := (bounds.Min.X + bounds.Max.X) * 0.5
	cy := (bounds.Min.Y + bounds.Max.Y) * 0.5
	switch fill {
	case MarkerFillLeft:
		return clipPolygonAgainstLine(points, func(p geom.Pt) float64 { return cx - p.X })
	case MarkerFillRight:
		return clipPolygonAgainstLine(points, func(p geom.Pt) float64 { return p.X - cx })
	case MarkerFillTop:
		return clipPolygonAgainstLine(points, func(p geom.Pt) float64 { return p.Y - cy })
	case MarkerFillBottom:
		return clipPolygonAgainstLine(points, func(p geom.Pt) float64 { return cy - p.Y })
	default:
		return append([]geom.Pt(nil), points...)
	}
}

func clipPolygonAgainstLine(points []geom.Pt, signedDistance func(geom.Pt) float64) []geom.Pt {
	if len(points) == 0 {
		return nil
	}
	out := make([]geom.Pt, 0, len(points)+2)
	const eps = 1e-12
	prev := points[len(points)-1]
	prevD := signedDistance(prev)
	prevInside := prevD >= -eps
	for _, cur := range points {
		curD := signedDistance(cur)
		curInside := curD >= -eps
		switch {
		case curInside && prevInside:
			out = append(out, cur)
		case curInside && !prevInside:
			out = append(out, markerClipIntersection(prev, cur, prevD, curD), cur)
		case !curInside && prevInside:
			out = append(out, markerClipIntersection(prev, cur, prevD, curD))
		}
		prev = cur
		prevD = curD
		prevInside = curInside
	}
	return dedupeAdjacentPoints(out)
}

func markerClipIntersection(a, b geom.Pt, da, db float64) geom.Pt {
	denom := da - db
	if math.Abs(denom) < 1e-12 {
		return b
	}
	t := da / denom
	return geom.Pt{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
	}
}

func dedupeAdjacentPoints(points []geom.Pt) []geom.Pt {
	if len(points) == 0 {
		return nil
	}
	out := points[:0]
	for _, pt := range points {
		if len(out) == 0 || pointDistance(out[len(out)-1], pt) > 1e-12 {
			out = append(out, pt)
		}
	}
	if len(out) > 1 && pointDistance(out[0], out[len(out)-1]) <= 1e-12 {
		out = out[:len(out)-1]
	}
	return out
}

func scatterAreaScale(area float64, ctx *DrawContext) float64 {
	if area <= 0 || math.IsNaN(area) || math.IsInf(area, 0) {
		return 0
	}
	dpi := 72.0
	if ctx != nil && ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	return math.Sqrt(area) * dpi / 72.0
}

// ScatterAreaFromRadius converts a marker radius in pixels to Matplotlib's
// scatter area unit, points^2, for the given DPI.
func ScatterAreaFromRadius(radiusPx, dpi float64) float64 {
	if radiusPx <= 0 || dpi <= 0 || math.IsNaN(radiusPx) || math.IsNaN(dpi) || math.IsInf(radiusPx, 0) || math.IsInf(dpi, 0) {
		return 0
	}
	radiusPt := radiusPx * 72.0 / dpi
	return math.Pi * radiusPt * radiusPt
}

// createCirclePath creates the same cubic unit-circle marker Matplotlib uses.
func (s *Scatter2D) createCirclePath(center geom.Pt, radius float64) geom.Path {
	const segments = 8
	const control = 0.2652031
	r := radius * 0.5
	delta := 2 * math.Pi / segments

	point := func(theta float64) geom.Pt {
		return geom.Pt{
			X: center.X + r*math.Cos(theta),
			Y: center.Y + r*math.Sin(theta),
		}
	}
	tangent := func(theta float64) geom.Pt {
		return geom.Pt{X: -math.Sin(theta), Y: math.Cos(theta)}
	}

	path := geom.Path{}
	theta0 := -math.Pi / 2
	path.MoveTo(point(theta0))
	for i := 0; i < segments; i++ {
		theta1 := theta0 + delta
		p0 := point(theta0)
		p1 := point(theta1)
		t0 := tangent(theta0)
		t1 := tangent(theta1)
		path.CubicTo(
			geom.Pt{X: p0.X + control*r*t0.X, Y: p0.Y + control*r*t0.Y},
			geom.Pt{X: p1.X - control*r*t1.X, Y: p1.Y - control*r*t1.Y},
			p1,
		)
		theta0 = theta1
	}
	path.Close()
	return path
}

func markerCirclePath(size float64) geom.Path {
	return (&Scatter2D{}).createCirclePath(geom.Pt{}, size)
}

func markerRectanglePath(x0, y0, x1, y1 float64) geom.Path {
	return polygonPath([]geom.Pt{
		{X: x0, Y: y0},
		{X: x1, Y: y0},
		{X: x1, Y: y1},
		{X: x0, Y: y1},
	}, true)
}

func markerLinePath(x0, y0, x1, y1 float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x0, Y: y0})
	path.LineTo(geom.Pt{X: x1, Y: y1})
	return path
}

func markerRegularPolygonPath(numsides int, angleDeg float64) geom.Path {
	if numsides < 3 {
		return geom.Path{}
	}
	points := make([]geom.Pt, numsides)
	angle := angleDeg * math.Pi / 180
	step := 2 * math.Pi / float64(numsides)
	for i := range points {
		theta := angle + float64(i)*step
		points[i] = geom.Pt{X: 0.5 * math.Cos(theta), Y: 0.5 * math.Sin(theta)}
	}
	return polygonPath(points, true)
}

func markerTrianglePath(angleDeg float64) geom.Path {
	return rotatePath(polygonPath([]geom.Pt{
		{X: 0, Y: 0.5},
		{X: -0.5, Y: -0.5},
		{X: 0.5, Y: -0.5},
	}, true), angleDeg)
}

func markerStarPath(numsides int, innerCircle float64, angleDeg float64, close bool) geom.Path {
	if numsides < 3 {
		return geom.Path{}
	}
	points := make([]geom.Pt, 0, numsides*2)
	angle := angleDeg * math.Pi / 180
	step := math.Pi / float64(numsides)
	for i := 0; i < numsides*2; i++ {
		radius := 0.5
		if i%2 == 1 {
			radius *= innerCircle
		}
		theta := angle + float64(i)*step
		points = append(points, geom.Pt{X: radius * math.Cos(theta), Y: radius * math.Sin(theta)})
	}
	return polygonPath(points, close)
}

func markerTriPath(angleDeg float64) geom.Path {
	path := geom.Path{}
	for _, seg := range [][2]geom.Pt{
		{{X: 0, Y: 0}, {X: 0, Y: -0.5}},
		{{X: 0, Y: 0}, {X: 0.4, Y: 0.25}},
		{{X: 0, Y: 0}, {X: -0.4, Y: 0.25}},
	} {
		path.MoveTo(seg[0])
		path.LineTo(seg[1])
	}
	return rotatePath(path, angleDeg)
}

func markerCaretPath(angleDeg float64, base bool) geom.Path {
	path := geom.Path{}
	if base {
		path.MoveTo(geom.Pt{X: -0.5, Y: 0})
		path.LineTo(geom.Pt{X: 0, Y: -0.75})
		path.LineTo(geom.Pt{X: 0.5, Y: 0})
	} else {
		path.MoveTo(geom.Pt{X: -0.5, Y: 0.75})
		path.LineTo(geom.Pt{X: 0, Y: 0})
		path.LineTo(geom.Pt{X: 0.5, Y: 0.75})
	}
	return rotatePath(path, angleDeg)
}

func markerFilledPlusPath() geom.Path {
	scale := 1.0 / 6.0
	points := []geom.Pt{
		{X: -1 * scale, Y: -3 * scale},
		{X: 1 * scale, Y: -3 * scale},
		{X: 1 * scale, Y: -1 * scale},
		{X: 3 * scale, Y: -1 * scale},
		{X: 3 * scale, Y: 1 * scale},
		{X: 1 * scale, Y: 1 * scale},
		{X: 1 * scale, Y: 3 * scale},
		{X: -1 * scale, Y: 3 * scale},
		{X: -1 * scale, Y: 1 * scale},
		{X: -3 * scale, Y: 1 * scale},
		{X: -3 * scale, Y: -1 * scale},
		{X: -1 * scale, Y: -1 * scale},
	}
	return polygonPath(points, true)
}

func markerFilledXPath() geom.Path {
	scale := 1.0 / 4.0
	points := []geom.Pt{
		{X: -1 * scale, Y: -2 * scale},
		{X: 0, Y: -1 * scale},
		{X: 1 * scale, Y: -2 * scale},
		{X: 2 * scale, Y: -1 * scale},
		{X: 1 * scale, Y: 0},
		{X: 2 * scale, Y: 1 * scale},
		{X: 1 * scale, Y: 2 * scale},
		{X: 0, Y: 1 * scale},
		{X: -1 * scale, Y: 2 * scale},
		{X: -2 * scale, Y: 1 * scale},
		{X: -1 * scale, Y: 0},
		{X: -2 * scale, Y: -1 * scale},
	}
	return polygonPath(points, true)
}

func markerTuplePath(tuple MarkerTuple) geom.Path {
	switch tuple.Style {
	case MarkerTuplePolygon:
		return markerRegularPolygonPath(tuple.NumSides, 90+tuple.AngleDeg)
	case MarkerTupleStar:
		return markerStarPath(tuple.NumSides, 0.5, 90+tuple.AngleDeg, true)
	case MarkerTupleAsterisk:
		path := geom.Path{}
		if tuple.NumSides < 2 {
			return path
		}
		angle := (90 + tuple.AngleDeg) * math.Pi / 180
		step := math.Pi / float64(tuple.NumSides)
		for i := 0; i < tuple.NumSides; i++ {
			theta := angle + float64(i)*step
			p := geom.Pt{X: 0.5 * math.Cos(theta), Y: 0.5 * math.Sin(theta)}
			path.MoveTo(geom.Pt{X: -p.X, Y: -p.Y})
			path.LineTo(p)
		}
		return path
	default:
		return geom.Path{}
	}
}

func mathTextMarkerPath(r render.Renderer, text, fontKey string) (geom.Path, bool) {
	if r != nil {
		if layout, ok := layoutDisplayText(r, text, 1, fontKey); ok {
			paths, ok := mathTextLayoutPaths(r, layout, geom.Pt{}, fontKey)
			if ok {
				if path, ok := combineAndNormalizeMarkerPaths(paths); ok {
					return path, true
				}
			}
		}
	}
	display := normalizeDisplayText(text)
	if display == "" {
		return geom.Path{}, false
	}
	path, ok := render.TextPath(display, geom.Pt{}, 1, fontKey)
	if !ok {
		return geom.Path{}, false
	}
	return normalizeMarkerPath(path)
}

func normalizeCustomMarkerPath(path geom.Path) geom.Path {
	maxAbs := 0.0
	for _, pt := range path.V {
		maxAbs = math.Max(maxAbs, math.Abs(pt.X))
		maxAbs = math.Max(maxAbs, math.Abs(pt.Y))
	}
	if maxAbs <= 0 || math.IsNaN(maxAbs) || math.IsInf(maxAbs, 0) {
		return geom.Path{}
	}
	scale := 0.5 / maxAbs
	return applyAffinePath(path, geom.Affine{A: scale, D: scale})
}

func combineAndNormalizeMarkerPaths(paths []geom.Path) (geom.Path, bool) {
	var combined geom.Path
	for _, path := range paths {
		combined.C = append(combined.C, path.C...)
		combined.V = append(combined.V, path.V...)
	}
	return normalizeMarkerPath(combined)
}

func normalizeMarkerPath(path geom.Path) (geom.Path, bool) {
	bounds, ok := pathBounds(path)
	if !ok {
		return geom.Path{}, false
	}
	maxDim := math.Max(bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y)
	if maxDim <= 0 || math.IsNaN(maxDim) || math.IsInf(maxDim, 0) {
		return geom.Path{}, false
	}
	affine := geom.Affine{
		A: 1 / maxDim,
		D: 1 / maxDim,
		E: -(bounds.Min.X + 0.5*(bounds.Max.X-bounds.Min.X)) / maxDim,
		F: -(bounds.Min.Y + 0.5*(bounds.Max.Y-bounds.Min.Y)) / maxDim,
	}
	return applyAffinePath(path, affine), true
}

func rotatePath(path geom.Path, angleDeg float64) geom.Path {
	theta := angleDeg * math.Pi / 180
	cos := math.Cos(theta)
	sin := math.Sin(theta)
	return applyAffinePath(path, geom.Affine{A: cos, B: sin, C: -sin, D: cos})
}

// createSquarePath creates a square marker centered at the given point.
func (s *Scatter2D) createSquarePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Square vertices
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X - half, Y: center.Y - half}, // bottom-left
		{X: center.X + half, Y: center.Y - half}, // bottom-right
		{X: center.X + half, Y: center.Y + half}, // top-right
		{X: center.X - half, Y: center.Y + half}, // top-left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createTrianglePath creates an upward-pointing triangle marker.
func (s *Scatter2D) createTrianglePath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Triangle vertices matching matplotlib's '^' marker geometry.
	half := 0.5 * radius
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X - half, Y: center.Y + half},
		{X: center.X + half, Y: center.Y + half},
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createDiamondPath creates a diamond (rotated square) marker.
func (s *Scatter2D) createDiamondPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Diamond vertices matching matplotlib's 'D' marker geometry.
	half := radius / math.Sqrt2
	vertices := []geom.Pt{
		{X: center.X, Y: center.Y - half}, // top
		{X: center.X + half, Y: center.Y}, // right
		{X: center.X, Y: center.Y + half}, // bottom
		{X: center.X - half, Y: center.Y}, // left
	}

	for i, v := range vertices {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}
	path.C = append(path.C, geom.ClosePath)

	return path
}

// createPlusPath creates a plus sign marker.
func (s *Scatter2D) createPlusPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// Plus is a line marker in matplotlib.
	half := 0.5 * radius
	hBar := []geom.Pt{
		{X: center.X - half, Y: center.Y},
		{X: center.X + half, Y: center.Y},
	}

	for i, v := range hBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Vertical bar
	vBar := []geom.Pt{
		{X: center.X, Y: center.Y - half},
		{X: center.X, Y: center.Y + half},
	}

	for i, v := range vBar {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// createCrossPath creates a cross (X) marker.
func (s *Scatter2D) createCrossPath(center geom.Pt, radius float64) geom.Path {
	path := geom.Path{}

	// First diagonal bar (\)
	diag1 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y - 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y + 0.5*radius},
	}

	for i, v := range diag1 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	// Second diagonal bar (/)
	diag2 := []geom.Pt{
		{X: center.X - 0.5*radius, Y: center.Y + 0.5*radius},
		{X: center.X + 0.5*radius, Y: center.Y - 0.5*radius},
	}

	for i, v := range diag2 {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, v)
	}

	return path
}

// Z returns the z-order for sorting.
func (s *Scatter2D) Z() float64 {
	return zOrDefault(s.z, defaultPatchZ)
}

// Bounds returns the data-space bounding box of all marker centers.
func (s *Scatter2D) Bounds(*DrawContext) geom.Rect {
	if len(s.XY) == 0 {
		return geom.Rect{}
	}

	// Initialize bounds with first point
	bounds := geom.Rect{
		Min: s.XY[0],
		Max: s.XY[0],
	}

	// Expand bounds to include all points
	for _, pt := range s.XY[1:] {
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}

	return bounds
}
