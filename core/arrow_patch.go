package core

import (
	"math"
	"strconv"
	"strings"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

// ArrowStyle describes a Matplotlib-style arrow mutation.
type ArrowStyle struct {
	Name         string
	HeadLength   float64
	HeadWidth    float64
	TailWidth    float64
	WidthA       float64
	WidthB       float64
	LengthA      float64
	LengthB      float64
	AngleA       float64
	AngleB       float64
	ScaleA       *float64
	ScaleB       *float64
	ShrinkFactor float64
}

// ConnectionStyle describes how FancyArrowPatch connects two positions.
type ConnectionStyle struct {
	Name     string
	Rad      float64
	AngleA   float64
	AngleB   float64
	ArmA     float64
	ArmB     float64
	Fraction float64
	Angle    *float64
}

// FancyArrowPatch draws an arrow whose head/tail size is fixed in display pixels.
type FancyArrowPatch struct {
	Patch
	PosA            geom.Pt
	PosB            geom.Pt
	Path            geom.Path
	PatchA          Artist
	PatchB          Artist
	ArrowStyle      ArrowStyle
	ConnectionStyle ConnectionStyle
	ShrinkA         float64
	ShrinkB         float64
	MutationScale   float64
	MutationAspect  float64
	Coords          CoordinateSpec
}

// ConnectionPatch connects two points that may use independent coordinate specs.
type ConnectionPatch struct {
	FancyArrowPatch
	XYA     geom.Pt
	XYB     geom.Pt
	CoordsA CoordinateSpec
	CoordsB CoordinateSpec
}

type arrowPathPart struct {
	path     geom.Path
	fillable bool
}

// ArrowStyleFromString resolves Matplotlib arrow style strings.
func ArrowStyleFromString(style string) (ArrowStyle, bool) {
	name, params := parsePatchStyleSpec(style)
	if name == "" {
		name = "simple"
	}
	normalized := strings.ToLower(name)
	out := ArrowStyle{
		Name:       normalized,
		HeadLength: 0.4,
		HeadWidth:  0.2,
		TailWidth:  0.2,
		WidthA:     1,
		WidthB:     1,
		LengthA:    0.2,
		LengthB:    0.2,
	}
	switch normalized {
	case "-", "->", "<-", "<->", "<|-", "-|>", "<|-|>", "]-", "-[", "]-[", "]->", "<-[":
	case "|-|":
		out.LengthA = 0
		out.LengthB = 0
	case "simple":
		out.HeadLength = 0.5
		out.HeadWidth = 0.5
	case "fancy":
		out.HeadLength = 0.4
		out.HeadWidth = 0.4
		out.TailWidth = 0.4
	case "wedge":
		out.TailWidth = 0.3
		out.ShrinkFactor = 0.5
	default:
		return ArrowStyle{}, false
	}
	applyArrowStyleParams(&out, params)
	return out, true
}

// ConnectionStyleFromString resolves Matplotlib connection style strings.
func ConnectionStyleFromString(style string) (ConnectionStyle, bool) {
	name, params := parsePatchStyleSpec(style)
	if name == "" {
		name = "arc3"
	}
	normalized := strings.ToLower(name)
	out := ConnectionStyle{Name: normalized}
	switch normalized {
	case "arc3":
	case "arc":
	case "angle", "angle3":
		out.AngleA = 90
	case "bar":
		out.Fraction = 0.3
	default:
		return ConnectionStyle{}, false
	}
	applyConnectionStyleParams(&out, params)
	return out, true
}

// Draw renders the fancy arrow in display space.
func (a *FancyArrowPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if a == nil || ren == nil || ctx == nil {
		return
	}
	patch := a.arrowDrawPatch()
	parts := a.displayParts(ctx, a.displayPath(ctx))
	for _, part := range parts {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			patch.drawStyledPath(ren, part.path, geom.Path{})
		} else {
			patch.drawStyledPath(ren, geom.Path{}, part.path)
		}
	}
}

// Bounds returns the data-space bounds when the endpoints are in data coords.
func (a *FancyArrowPatch) Bounds(*DrawContext) geom.Rect {
	if a == nil || !artistUsesDataCoords(a, a.Coords) {
		return geom.Rect{}
	}
	if len(a.Path.C) > 0 {
		bounds, _ := pathBounds(a.Path)
		return bounds
	}
	bounds := geom.Rect{Min: a.PosA, Max: a.PosA}
	return expandRect(bounds, a.PosB)
}

// Draw renders the connection in display space.
func (c *ConnectionPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if c == nil || ren == nil || ctx == nil {
		return
	}
	patch := c.arrowDrawPatch()
	path := c.connectionDisplayPath(ctx)
	parts := c.displayParts(ctx, path)
	for _, part := range parts {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			patch.drawStyledPath(ren, part.path, geom.Path{})
		} else {
			patch.drawStyledPath(ren, geom.Path{}, part.path)
		}
	}
}

// Bounds returns an empty rect because a connection may span multiple spaces.
func (c *ConnectionPatch) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func (a *FancyArrowPatch) arrowDrawPatch() *Patch {
	if a == nil {
		return &Patch{}
	}
	patch := a.Patch
	if patch.LineJoin == render.JoinMiter && patch.LineCap == render.CapButt {
		patch.LineJoin = render.JoinRound
		patch.LineCap = render.CapRound
	}
	return &patch
}

func (a *FancyArrowPatch) displayPath(ctx *DrawContext) geom.Path {
	if len(a.Path.C) > 0 {
		return buildArtistDisplayPath(ctx, a, a.Coords, a.Path, geom.Identity())
	}
	tr := artistTransformFor(ctx, a, a.Coords)
	if tr == nil {
		tr = transform.NewAffine(geom.Identity())
	}
	posA := tr.Apply(a.PosA)
	posB := tr.Apply(a.PosB)
	style := a.ConnectionStyle
	if style.Name == "" {
		style, _ = ConnectionStyleFromString("arc3")
	}
	path := style.connect(posA, posB, 0, 0)
	path = clipConnectionPathToPatch(ctx, path, a.PatchA, true)
	path = clipConnectionPathToPatch(ctx, path, a.PatchB, false)
	return shrinkPathEndpoints(path, arrowShrinkPixels(ctx, a.effectiveShrinkA()), arrowShrinkPixels(ctx, a.effectiveShrinkB()))
}

func (c *ConnectionPatch) connectionDisplayPath(ctx *DrawContext) geom.Path {
	aTrans := ctx.TransformFor(c.CoordsA)
	bTrans := ctx.TransformFor(c.CoordsB)
	if aTrans == nil {
		aTrans = transform.NewAffine(geom.Identity())
	}
	if bTrans == nil {
		bTrans = transform.NewAffine(geom.Identity())
	}
	style := c.ConnectionStyle
	if style.Name == "" {
		style, _ = ConnectionStyleFromString("arc3")
	}
	path := style.connect(aTrans.Apply(c.XYA), bTrans.Apply(c.XYB), 0, 0)
	path = clipConnectionPathToPatch(ctx, path, c.PatchA, true)
	path = clipConnectionPathToPatch(ctx, path, c.PatchB, false)
	return shrinkPathEndpoints(path, arrowShrinkPixels(ctx, c.ShrinkA), arrowShrinkPixels(ctx, c.ShrinkB))
}

func (a *FancyArrowPatch) displayParts(_ *DrawContext, path geom.Path) []arrowPathPart {
	style := a.ArrowStyle
	if style.Name == "" {
		style, _ = ArrowStyleFromString("simple")
	}
	scale := a.MutationScale
	if scale <= 0 {
		scale = 10
	}
	lineWidth := a.EdgeWidth
	if lineWidth <= 0 {
		lineWidth = 1
	}
	aspect := 1.0
	if a != nil && a.MutationAspect > 0 {
		aspect = a.MutationAspect
	}
	if aspect == 1 {
		return style.transmute(path, scale, lineWidth)
	}
	parts := style.transmute(scalePathY(path, 1/aspect), scale, lineWidth)
	for i := range parts {
		parts[i].path = scalePathY(parts[i].path, aspect)
	}
	return parts
}

func (a *FancyArrowPatch) effectiveShrinkA() float64 {
	if a == nil || a.ShrinkA <= 0 {
		return 2
	}
	return a.ShrinkA
}

func (a *FancyArrowPatch) effectiveShrinkB() float64 {
	if a == nil || a.ShrinkB <= 0 {
		return 2
	}
	return a.ShrinkB
}

func arrowShrinkPixels(ctx *DrawContext, points float64) float64 {
	if points <= 0 {
		return 0
	}
	if ctx == nil {
		return points
	}
	return pointsToPixels(ctx.RC, points)
}

func scalePathY(path geom.Path, scale float64) geom.Path {
	if scale == 1 || len(path.V) == 0 {
		return path
	}
	out := path
	out.V = append([]geom.Pt(nil), path.V...)
	for i := range out.V {
		out.V[i].Y *= scale
	}
	return out
}

func (s ConnectionStyle) connect(posA, posB geom.Pt, shrinkA, shrinkB float64) geom.Path {
	switch strings.ToLower(s.Name) {
	case "angle3":
		return shrinkPathEndpoints(connectionAngle3Path(posA, posB, s.AngleA, s.AngleB), shrinkA, shrinkB)
	case "angle":
		return shrinkPathEndpoints(connectionAnglePath(posA, posB, s.AngleA, s.AngleB, s.Rad), shrinkA, shrinkB)
	case "arc":
		return shrinkPathEndpoints(connectionArcPath(posA, posB, s.AngleA, s.AngleB, s.ArmA, s.ArmB, s.Rad), shrinkA, shrinkB)
	case "bar":
		return shrinkPathEndpoints(connectionBarPath(posA, posB, s.ArmA, s.ArmB, s.Fraction, s.Angle), shrinkA, shrinkB)
	default:
		return shrinkPathEndpoints(connectionArc3Path(posA, posB, s.Rad), shrinkA, shrinkB)
	}
}

func connectionArc3Path(posA, posB geom.Pt, rad float64) geom.Path {
	mid := geom.Pt{X: (posA.X + posB.X) / 2, Y: (posA.Y + posB.Y) / 2}
	dx, dy := posB.X-posA.X, posB.Y-posA.Y
	ctrl := geom.Pt{X: mid.X + rad*dy, Y: mid.Y - rad*dx}
	path := geom.Path{}
	path.MoveTo(posA)
	path.QuadTo(ctrl, posB)
	return path
}

func connectionAngle3Path(posA, posB geom.Pt, angleA, angleB float64) geom.Path {
	dirA := angleUnit(angleA)
	dirB := angleUnit(angleB)
	ctrl, ok := lineIntersection(posA, dirA, posB, dirB)
	if !ok {
		ctrl = geom.Pt{X: (posA.X + posB.X) / 2, Y: (posA.Y + posB.Y) / 2}
	}
	path := geom.Path{}
	path.MoveTo(posA)
	path.QuadTo(ctrl, posB)
	return path
}

func connectionAnglePath(posA, posB geom.Pt, angleA, angleB, rad float64) geom.Path {
	dirA := angleUnit(angleA)
	dirB := angleUnit(angleB)
	corner, ok := lineIntersection(posA, dirA, posB, dirB)
	if !ok {
		return connectionArc3Path(posA, posB, 0)
	}
	path := geom.Path{}
	path.MoveTo(posA)
	if rad <= 0 {
		path.LineTo(corner)
		path.LineTo(posB)
		return path
	}
	p1 := pointToward(corner, posA, rad)
	p2 := pointToward(corner, posB, rad)
	path.LineTo(p1)
	path.QuadTo(corner, p2)
	path.LineTo(posB)
	return path
}

func connectionArcPath(posA, posB geom.Pt, angleA, angleB, armA, armB, rad float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(posA)
	rounded := []geom.Pt{}
	if armA != 0 {
		dir := angleUnit(angleA)
		rounded = append(rounded,
			geom.Pt{X: posA.X + (armA-rad)*dir.X, Y: posA.Y + (armA-rad)*dir.Y},
			geom.Pt{X: posA.X + armA*dir.X, Y: posA.Y + armA*dir.Y},
		)
	}
	if armB != 0 {
		dir := angleUnit(angleB)
		elbow := geom.Pt{X: posB.X + armB*dir.X, Y: posB.Y + armB*dir.Y}
		if len(rounded) > 0 {
			prev := rounded[len(rounded)-1]
			if d := distance(prev, elbow); d > 0 {
				rounded = append(rounded, pointToward(prev, elbow, rad))
				appendArcRoundedPoints(&path, rounded)
				rounded = []geom.Pt{pointToward(prev, elbow, d-rad), elbow}
			}
		} else {
			prev := posA
			if d := distance(prev, elbow); d > 0 {
				rounded = append(rounded, pointToward(prev, elbow, d-rad), elbow)
			}
		}
	}
	if len(rounded) > 0 {
		prev := rounded[len(rounded)-1]
		if d := distance(prev, posB); d > 0 {
			rounded = append(rounded, pointToward(prev, posB, rad))
			appendArcRoundedPoints(&path, rounded)
		}
	}
	path.LineTo(posB)
	return path
}

func appendArcRoundedPoints(path *geom.Path, rounded []geom.Pt) {
	if path == nil || len(rounded) == 0 {
		return
	}
	path.LineTo(rounded[0])
	if len(rounded) >= 3 {
		path.QuadTo(rounded[1], rounded[2])
	}
}

func connectionBarPath(posA, posB geom.Pt, armA, armB, fraction float64, angle *float64) geom.Path {
	dx, dy := posB.X-posA.X, posB.Y-posA.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return geom.Path{}
	}
	ux, uy := dx/length, dy/length
	projectedB := posB
	if angle != nil {
		theta := *angle * math.Pi / 180
		targetTheta := math.Atan2(dy, dx)
		dtheta := targetTheta - theta
		offAxis := length * math.Sin(dtheta)
		onAxis := length * math.Cos(dtheta)
		projectedB = geom.Pt{
			X: posA.X + onAxis*math.Cos(theta),
			Y: posA.Y + onAxis*math.Sin(theta),
		}
		armB -= offAxis

		dx, dy = projectedB.X-posA.X, projectedB.Y-posA.Y
		projectedLength := math.Hypot(dx, dy)
		if projectedLength == 0 {
			return geom.Path{}
		}
		ux, uy = dx/projectedLength, dy/projectedLength
	}
	perp := geom.Pt{X: uy, Y: -ux}
	arm := math.Max(armA, armB) + fraction*length
	c1 := geom.Pt{X: posA.X + perp.X*arm, Y: posA.Y + perp.Y*arm}
	c2 := geom.Pt{X: projectedB.X + perp.X*arm, Y: projectedB.Y + perp.Y*arm}
	path := geom.Path{}
	path.MoveTo(posA)
	path.LineTo(c1)
	path.LineTo(c2)
	path.LineTo(posB)
	return path
}

func (s ArrowStyle) transmute(path geom.Path, mutationSize, lineWidth float64) []arrowPathPart {
	if len(path.C) == 0 || len(path.V) < 2 {
		return nil
	}
	name := strings.ToLower(s.Name)
	switch name {
	case "wedge":
		return []arrowPathPart{{path: wedgeArrowPathForConnection(path, s.TailWidth*mutationSize, s.ShrinkFactor), fillable: true}}
	case "simple":
		return []arrowPathPart{{path: filledArrowPathForConnection(path, s.TailWidth*mutationSize, s.HeadWidth*mutationSize, s.HeadLength*mutationSize), fillable: true}}
	case "fancy":
		return []arrowPathPart{{path: filledArrowPathForConnection(path, s.TailWidth*mutationSize, s.HeadWidth*mutationSize, s.HeadLength*mutationSize), fillable: true}}
	}

	beginHead := strings.HasPrefix(name, "<")
	endHead := strings.HasSuffix(name, ">")
	fillBegin := strings.HasPrefix(name, "<|")
	fillEnd := strings.HasSuffix(name, "|>")
	beginBracket := strings.HasPrefix(name, "]") || strings.HasPrefix(name, "|")
	endBracket := strings.HasSuffix(name, "[") || strings.HasSuffix(name, "|")
	headLength := curveArrowHeadLength(s, mutationSize)
	linePath := shortenedCurveLinePath(path, beginHead, endHead, headLength)
	parts := []arrowPathPart{{path: linePath, fillable: false}}

	if beginHead {
		head := arrowHeadPath(pathSecond(path), pathStart(path), headLength, s.HeadWidth*mutationSize, fillBegin, lineWidth)
		parts = append(parts, arrowPathPart{path: head, fillable: fillBegin})
	} else if beginBracket {
		scale := bracketScale(s.ScaleA, mutationSize)
		parts = append(parts, arrowPathPart{path: bracketPath(pathStart(path), pathSecond(path), s.WidthA*scale, s.LengthA*scale, s.AngleA), fillable: false})
	}
	if endHead {
		head := arrowHeadPath(pathPenultimate(path), pathEnd(path), headLength, s.HeadWidth*mutationSize, fillEnd, lineWidth)
		parts = append(parts, arrowPathPart{path: head, fillable: fillEnd})
	} else if endBracket {
		scale := bracketScale(s.ScaleB, mutationSize)
		parts = append(parts, arrowPathPart{path: bracketPath(pathEnd(path), pathPenultimate(path), s.WidthB*scale, s.LengthB*scale, s.AngleB), fillable: false})
	}
	return parts
}

func bracketScale(scale *float64, mutationSize float64) float64 {
	if scale != nil {
		return *scale
	}
	return mutationSize
}

func clipConnectionPathToPatch(ctx *DrawContext, path geom.Path, patch Artist, start bool) geom.Path {
	if patch == nil || len(path.V) < 2 {
		return path
	}
	patchPath, ok := sourcePatchDisplayPath(patch, ctx)
	if !ok {
		return path
	}
	polygon := patchPath.Interpolated(8).V
	if len(polygon) < 3 {
		return path
	}
	var endpoint geom.Pt
	if start {
		endpoint = pathStart(path)
	} else {
		endpoint = pathEnd(path)
	}
	if !pointInPolygon(endpoint, polygon) {
		return path
	}
	boundary, ok := connectionPatchBoundaryPoint(path, polygon, start)
	if !ok {
		return path
	}
	out := path
	out.V = append([]geom.Pt(nil), path.V...)
	if start {
		out.V[0] = boundary
	} else {
		out.V[len(out.V)-1] = boundary
	}
	return out
}

func connectionPatchBoundaryPoint(path geom.Path, polygon []geom.Pt, start bool) (geom.Pt, bool) {
	pts := path.Interpolated(64).V
	if len(pts) < 2 {
		return geom.Pt{}, false
	}
	if !start {
		for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
			pts[i], pts[j] = pts[j], pts[i]
		}
	}
	inside := pts[0]
	if !pointInPolygon(inside, polygon) {
		return geom.Pt{}, false
	}
	for _, outside := range pts[1:] {
		if pointInPolygon(outside, polygon) {
			inside = outside
			continue
		}
		return refinePatchBoundaryPoint(inside, outside, polygon), true
	}
	return geom.Pt{}, false
}

func refinePatchBoundaryPoint(inside, outside geom.Pt, polygon []geom.Pt) geom.Pt {
	lo := inside
	hi := outside
	for i := 0; i < 32; i++ {
		mid := geom.Pt{X: (lo.X + hi.X) / 2, Y: (lo.Y + hi.Y) / 2}
		if pointInPolygon(mid, polygon) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

func curveArrowHeadLength(style ArrowStyle, mutationSize float64) float64 {
	if style.HeadLength > 0 {
		return style.HeadLength * mutationSize
	}
	return 4
}

func shortenedCurveLinePath(path geom.Path, beginHead, endHead bool, headLength float64) geom.Path {
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: append([]geom.Pt(nil), path.V...),
	}
	if len(out.V) < 2 || headLength <= 0 {
		return out
	}
	if beginHead {
		out.V[0] = pointToward(pathStart(path), pathSecond(path), headLength)
	}
	if endHead {
		last := len(out.V) - 1
		out.V[last] = pointToward(pathEnd(path), pathPenultimate(path), headLength)
	}
	return out
}

func filledArrowPath(start, tip geom.Pt, tailWidth, headWidth, headLength float64) geom.Path {
	dx, dy := tip.X-start.X, tip.Y-start.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return geom.Path{}
	}
	if tailWidth <= 0 {
		tailWidth = math.Max(1, length*0.04)
	}
	if headWidth <= 0 {
		headWidth = tailWidth * 2.5
	}
	if headLength <= 0 || headLength > length {
		headLength = math.Min(length, math.Max(headWidth, length*0.25))
	}
	ux, uy := dx/length, dy/length
	px, py := -uy, ux
	base := geom.Pt{X: tip.X - ux*headLength, Y: tip.Y - uy*headLength}
	points := []geom.Pt{
		{X: start.X + px*tailWidth/2, Y: start.Y + py*tailWidth/2},
		{X: base.X + px*tailWidth/2, Y: base.Y + py*tailWidth/2},
		{X: base.X + px*headWidth/2, Y: base.Y + py*headWidth/2},
		tip,
		{X: base.X - px*headWidth/2, Y: base.Y - py*headWidth/2},
		{X: base.X - px*tailWidth/2, Y: base.Y - py*tailWidth/2},
		{X: start.X - px*tailWidth/2, Y: start.Y - py*tailWidth/2},
	}
	return polygonPath(points, true)
}

func filledArrowPathForConnection(path geom.Path, tailWidth, headWidth, headLength float64) geom.Path {
	start, ctrl, tip, ok := quadraticConnectionPoints(path)
	if !ok {
		return filledArrowPath(pathStart(path), pathEnd(path), tailWidth, headWidth, headLength)
	}
	if tailWidth <= 0 {
		tailWidth = math.Max(1, distance(start, tip)*0.04)
	}
	if headWidth <= 0 {
		headWidth = tailWidth * 2.5
	}
	if headLength <= 0 {
		headLength = headWidth
	}

	totalLength := approximateQuadraticLength(start, ctrl, tip)
	if totalLength <= 0 {
		return geom.Path{}
	}
	headT := quadraticTAtDistanceFromEnd(start, ctrl, tip, math.Min(headLength, totalLength))
	headBase := quadraticPoint(start, ctrl, tip, headT)
	midT := headT / 2
	mid := quadraticPoint(start, ctrl, tip, midT)

	startNormal := normalForVector(quadraticDerivative(start, ctrl, tip, 0))
	midNormal := normalForVector(quadraticDerivative(start, ctrl, tip, midT))
	headNormal := normalForVector(quadraticDerivative(start, ctrl, tip, headT))
	if startNormal == (geom.Pt{}) || midNormal == (geom.Pt{}) || headNormal == (geom.Pt{}) {
		return filledArrowPath(start, tip, tailWidth, headWidth, headLength)
	}

	tailHalf := tailWidth / 2
	midHalf := tailHalf
	startInset := start
	headHalf := headWidth / 2
	points := []geom.Pt{
		{X: startInset.X + startNormal.X*tailHalf, Y: startInset.Y + startNormal.Y*tailHalf},
		{X: mid.X + midNormal.X*midHalf, Y: mid.Y + midNormal.Y*midHalf},
		{X: headBase.X + headNormal.X*tailWidth/2, Y: headBase.Y + headNormal.Y*tailWidth/2},
		{X: headBase.X + headNormal.X*headHalf, Y: headBase.Y + headNormal.Y*headHalf},
		tip,
		{X: headBase.X - headNormal.X*headHalf, Y: headBase.Y - headNormal.Y*headHalf},
		{X: headBase.X - headNormal.X*tailWidth/2, Y: headBase.Y - headNormal.Y*tailWidth/2},
		{X: mid.X - midNormal.X*midHalf, Y: mid.Y - midNormal.Y*midHalf},
		{X: startInset.X - startNormal.X*tailHalf, Y: startInset.Y - startNormal.Y*tailHalf},
	}
	return polygonPath(points, true)
}

func wedgeArrowPath(start, tip geom.Pt, tailWidth, shrinkFactor float64) geom.Path {
	dx, dy := tip.X-start.X, tip.Y-start.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return geom.Path{}
	}
	if tailWidth <= 0 {
		tailWidth = math.Max(1, length*0.04)
	}
	if shrinkFactor <= 0 {
		shrinkFactor = 0.5
	}
	ux, uy := dx/length, dy/length
	px, py := -uy, ux
	mid := geom.Pt{X: (start.X + tip.X) / 2, Y: (start.Y + tip.Y) / 2}
	tailHalf := tailWidth / 2
	midHalf := tailHalf * shrinkFactor
	points := []geom.Pt{
		{X: start.X + px*tailHalf, Y: start.Y + py*tailHalf},
		{X: mid.X + px*midHalf, Y: mid.Y + py*midHalf},
		tip,
		{X: mid.X - px*midHalf, Y: mid.Y - py*midHalf},
		{X: start.X - px*tailHalf, Y: start.Y - py*tailHalf},
	}
	return polygonPath(points, true)
}

func wedgeArrowPathForConnection(path geom.Path, tailWidth, shrinkFactor float64) geom.Path {
	start, ctrl, tip, ok := quadraticConnectionPoints(path)
	if !ok {
		return wedgeArrowPath(pathStart(path), pathEnd(path), tailWidth, shrinkFactor)
	}
	if tailWidth <= 0 {
		tailWidth = math.Max(1, distance(start, tip)*0.04)
	}
	if shrinkFactor <= 0 {
		shrinkFactor = 0.5
	}

	mid := quadraticPoint(start, ctrl, tip, 0.5)
	startNormal := normalForVector(geom.Pt{X: ctrl.X - start.X, Y: ctrl.Y - start.Y})
	midNormal := normalForVector(geom.Pt{X: tip.X - start.X, Y: tip.Y - start.Y})
	if midNormal == (geom.Pt{}) {
		midNormal = normalForVector(geom.Pt{X: tip.X - ctrl.X, Y: tip.Y - ctrl.Y})
	}
	if startNormal == (geom.Pt{}) || midNormal == (geom.Pt{}) {
		return wedgeArrowPath(start, tip, tailWidth, shrinkFactor)
	}

	tailHalf := tailWidth / 2
	midHalf := tailHalf * shrinkFactor
	points := []geom.Pt{
		{X: start.X + startNormal.X*tailHalf, Y: start.Y + startNormal.Y*tailHalf},
		{X: mid.X + midNormal.X*midHalf, Y: mid.Y + midNormal.Y*midHalf},
		tip,
		{X: mid.X - midNormal.X*midHalf, Y: mid.Y - midNormal.Y*midHalf},
		{X: start.X - startNormal.X*tailHalf, Y: start.Y - startNormal.Y*tailHalf},
	}
	return polygonPath(points, true)
}

func quadraticConnectionPoints(path geom.Path) (geom.Pt, geom.Pt, geom.Pt, bool) {
	if len(path.C) < 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) < 3 {
		return geom.Pt{}, geom.Pt{}, geom.Pt{}, false
	}
	return path.V[0], path.V[1], path.V[2], true
}

func quadraticPoint(start, ctrl, end geom.Pt, t float64) geom.Pt {
	mt := 1 - t
	return geom.Pt{
		X: mt*mt*start.X + 2*mt*t*ctrl.X + t*t*end.X,
		Y: mt*mt*start.Y + 2*mt*t*ctrl.Y + t*t*end.Y,
	}
}

func quadraticDerivative(start, ctrl, end geom.Pt, t float64) geom.Pt {
	return geom.Pt{
		X: 2*(1-t)*(ctrl.X-start.X) + 2*t*(end.X-ctrl.X),
		Y: 2*(1-t)*(ctrl.Y-start.Y) + 2*t*(end.Y-ctrl.Y),
	}
}

func approximateQuadraticLength(start, ctrl, end geom.Pt) float64 {
	const steps = 24
	length := 0.0
	prev := start
	for i := 1; i <= steps; i++ {
		pt := quadraticPoint(start, ctrl, end, float64(i)/steps)
		length += distance(prev, pt)
		prev = pt
	}
	return length
}

func quadraticTAtDistanceFromEnd(start, ctrl, end geom.Pt, target float64) float64 {
	const steps = 48
	if target <= 0 {
		return 1
	}
	accum := 0.0
	prev := end
	for i := steps - 1; i >= 0; i-- {
		t := float64(i) / steps
		pt := quadraticPoint(start, ctrl, end, t)
		seg := distance(prev, pt)
		if accum+seg >= target {
			if seg == 0 {
				return t
			}
			f := (target - accum) / seg
			return (float64(i) + 1 - f) / steps
		}
		accum += seg
		prev = pt
	}
	return 0
}

func normalForVector(v geom.Pt) geom.Pt {
	length := math.Hypot(v.X, v.Y)
	if length == 0 {
		return geom.Pt{}
	}
	return geom.Pt{X: -v.Y / length, Y: v.X / length}
}

func arrowHeadPath(from, tip geom.Pt, headLength, headWidth float64, fill bool, lineWidth float64) geom.Path {
	dx, dy := tip.X-from.X, tip.Y-from.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return geom.Path{}
	}
	if headLength <= 0 {
		headLength = 4
	}
	if headWidth <= 0 {
		headWidth = headLength / 2
	}
	ux, uy := dx/length, dy/length
	base := geom.Pt{X: tip.X - ux*headLength, Y: tip.Y - uy*headLength}
	px, py := -uy*headWidth/2, ux*headWidth/2
	left := geom.Pt{X: base.X + px, Y: base.Y + py}
	right := geom.Pt{X: base.X - px, Y: base.Y - py}
	if fill {
		return polygonPath([]geom.Pt{tip, left, right}, true)
	}
	path := geom.Path{}
	path.MoveTo(left)
	path.LineTo(tip)
	path.LineTo(right)
	return path
}

func bracketPath(anchor, toward geom.Pt, width, length, angleDeg float64) geom.Path {
	dx, dy := anchor.X-toward.X, anchor.Y-toward.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		return geom.Path{}
	}
	ux, uy := dx/dist, dy/dist
	px, py := -uy, ux
	half := width / 2
	out := geom.Pt{X: ux * length, Y: uy * length}
	p1 := geom.Pt{X: anchor.X + px*half + out.X, Y: anchor.Y + py*half + out.Y}
	p2 := geom.Pt{X: anchor.X + px*half, Y: anchor.Y + py*half}
	p3 := geom.Pt{X: anchor.X - px*half, Y: anchor.Y - py*half}
	p4 := geom.Pt{X: anchor.X - px*half + out.X, Y: anchor.Y - py*half + out.Y}
	path := geom.Path{}
	path.MoveTo(p1)
	path.LineTo(p2)
	path.LineTo(p3)
	path.LineTo(p4)
	if angleDeg != 0 {
		path = rotatePathAround(path, anchor, angleDeg)
	}
	return path
}

func parsePatchStyleSpec(spec string) (string, map[string]float64) {
	parts := strings.Split(spec, ",")
	name := strings.TrimSpace(parts[0])
	params := map[string]float64{}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			continue
		}
		params[strings.ToLower(strings.TrimSpace(key))] = parsed
	}
	return name, params
}

func applyArrowStyleParams(style *ArrowStyle, params map[string]float64) {
	for key, value := range params {
		switch key {
		case "head_length":
			style.HeadLength = value
		case "head_width":
			style.HeadWidth = value
		case "tail_width":
			style.TailWidth = value
		case "widtha":
			style.WidthA = value
		case "widthb":
			style.WidthB = value
		case "lengtha":
			style.LengthA = value
		case "lengthb":
			style.LengthB = value
		case "anglea":
			style.AngleA = value
		case "angleb":
			style.AngleB = value
		case "scalea":
			v := value
			style.ScaleA = &v
		case "scaleb":
			v := value
			style.ScaleB = &v
		case "shrink_factor":
			style.ShrinkFactor = value
		}
	}
}

func applyConnectionStyleParams(style *ConnectionStyle, params map[string]float64) {
	for key, value := range params {
		switch key {
		case "rad":
			style.Rad = value
		case "anglea":
			style.AngleA = value
		case "angleb":
			style.AngleB = value
		case "arma":
			style.ArmA = value
		case "armb":
			style.ArmB = value
		case "fraction":
			style.Fraction = value
		case "angle":
			v := value
			style.Angle = &v
		}
	}
}

func shrinkPathEndpoints(path geom.Path, shrinkA, shrinkB float64) geom.Path {
	if len(path.V) < 2 || (shrinkA <= 0 && shrinkB <= 0) {
		return path
	}
	out := path
	out.V = append([]geom.Pt(nil), path.V...)
	if shrinkA > 0 {
		out.V[0] = pointToward(out.V[0], pathSecond(out), shrinkA)
	}
	if shrinkB > 0 {
		last := len(out.V) - 1
		out.V[last] = pointToward(out.V[last], pathPenultimate(out), shrinkB)
	}
	return out
}

func pathStart(path geom.Path) geom.Pt {
	if len(path.V) == 0 {
		return geom.Pt{}
	}
	return path.V[0]
}

func pathEnd(path geom.Path) geom.Pt {
	if len(path.V) == 0 {
		return geom.Pt{}
	}
	return path.V[len(path.V)-1]
}

func pathSecond(path geom.Path) geom.Pt {
	if len(path.V) < 2 {
		return pathStart(path)
	}
	return path.V[1]
}

func pathPenultimate(path geom.Path) geom.Pt {
	if len(path.V) < 2 {
		return pathEnd(path)
	}
	return path.V[len(path.V)-2]
}

func pointToward(from, to geom.Pt, dist float64) geom.Pt {
	dx, dy := to.X-from.X, to.Y-from.Y
	length := math.Hypot(dx, dy)
	if length == 0 || dist == 0 {
		return from
	}
	if math.Abs(dist) > length {
		dist = math.Copysign(length, dist)
	}
	return geom.Pt{X: from.X + dx/length*dist, Y: from.Y + dy/length*dist}
}

func distance(a, b geom.Pt) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func angleUnit(angleDeg float64) geom.Pt {
	rad := angleDeg * math.Pi / 180
	return geom.Pt{X: math.Cos(rad), Y: math.Sin(rad)}
}

func lineIntersection(p geom.Pt, d geom.Pt, q geom.Pt, e geom.Pt) (geom.Pt, bool) {
	den := d.X*e.Y - d.Y*e.X
	if math.Abs(den) < 1e-12 {
		return geom.Pt{}, false
	}
	t := ((q.X-p.X)*e.Y - (q.Y-p.Y)*e.X) / den
	return geom.Pt{X: p.X + d.X*t, Y: p.Y + d.Y*t}, true
}

func rotatePathAround(path geom.Path, origin geom.Pt, angleDeg float64) geom.Path {
	rad := angleDeg * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	affine := translateAffine(origin).
		Mul(geom.Affine{A: cosA, B: sinA, C: -sinA, D: cosA}).
		Mul(translateAffine(geom.Pt{X: -origin.X, Y: -origin.Y}))
	return applyAffinePath(path, affine)
}
