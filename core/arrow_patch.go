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
	Name       string
	HeadLength float64
	HeadWidth  float64
	TailWidth  float64
	WidthA     float64
	WidthB     float64
	LengthA    float64
	LengthB    float64
	AngleA     float64
	AngleB     float64
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
	case "-", "->", "<-", "<->", "<|-", "-|>", "<|-|>", "]-", "-[", "]-[", "|-|", "]->", "<-[":
	case "simple":
		out.HeadLength = 0.5
		out.HeadWidth = 0.5
	case "fancy":
		out.HeadLength = 0.4
		out.HeadWidth = 0.4
		out.TailWidth = 0.4
	case "wedge":
		out.TailWidth = 0.3
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
	out := ConnectionStyle{Name: normalized, Fraction: 0.3, AngleA: 90}
	switch normalized {
	case "arc3", "arc", "angle", "angle3", "bar":
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
	parts := a.displayParts(ctx, a.displayPath(ctx))
	for _, part := range parts {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			a.drawStyledPath(ren, part.path, geom.Path{})
		} else {
			a.drawStyledPath(ren, geom.Path{}, part.path)
		}
	}
}

// Bounds returns the data-space bounds when the endpoints are in data coords.
func (a *FancyArrowPatch) Bounds(*DrawContext) geom.Rect {
	if a == nil || !isDataCoords(a.Coords) {
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
	path := c.connectionDisplayPath(ctx)
	parts := c.displayParts(ctx, path)
	for _, part := range parts {
		if len(part.path.C) == 0 {
			continue
		}
		if part.fillable {
			c.drawStyledPath(ren, part.path, geom.Path{})
		} else {
			c.drawStyledPath(ren, geom.Path{}, part.path)
		}
	}
}

// Bounds returns an empty rect because a connection may span multiple spaces.
func (c *ConnectionPatch) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

func (a *FancyArrowPatch) displayPath(ctx *DrawContext) geom.Path {
	if len(a.Path.C) > 0 {
		return buildDisplayPath(ctx, a.Coords, a.Path, geom.Identity())
	}
	tr := ctx.TransformFor(a.Coords)
	if tr == nil {
		tr = transform.NewAffine(geom.Identity())
	}
	posA := tr.Apply(a.PosA)
	posB := tr.Apply(a.PosB)
	style := a.ConnectionStyle
	if style.Name == "" {
		style, _ = ConnectionStyleFromString("arc3")
	}
	return style.connect(posA, posB, a.ShrinkA, a.ShrinkB)
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
	return style.connect(aTrans.Apply(c.XYA), bTrans.Apply(c.XYB), c.ShrinkA, c.ShrinkB)
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
	return style.transmute(path, scale, lineWidth)
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
	if rad == 0 {
		path := geom.Path{}
		path.MoveTo(posA)
		path.LineTo(posB)
		return path
	}
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
	last := posA
	if armA != 0 {
		dir := angleUnit(angleA)
		last = geom.Pt{X: posA.X + armA*dir.X, Y: posA.Y + armA*dir.Y}
		path.LineTo(last)
	}
	if armB != 0 {
		dir := angleUnit(angleB)
		elbow := geom.Pt{X: posB.X + armB*dir.X, Y: posB.Y + armB*dir.Y}
		if rad > 0 {
			p1 := pointToward(last, elbow, math.Min(rad, distance(last, elbow)/2))
			p2 := pointToward(elbow, last, math.Min(rad, distance(last, elbow)/2))
			if distance(last, p1) > 0 {
				path.LineTo(p1)
			}
			path.QuadTo(elbow, p2)
		} else {
			path.LineTo(elbow)
		}
		last = elbow
	}
	if rad > 0 && distance(last, posB) > rad {
		p := pointToward(posB, last, rad)
		path.LineTo(p)
		path.QuadTo(posB, posB)
	} else {
		path.LineTo(posB)
	}
	return path
}

func connectionBarPath(posA, posB geom.Pt, armA, armB, fraction float64, angle *float64) geom.Path {
	dx, dy := posB.X-posA.X, posB.Y-posA.Y
	length := math.Hypot(dx, dy)
	if length == 0 {
		return geom.Path{}
	}
	ux, uy := dx/length, dy/length
	if angle != nil {
		dir := angleUnit(*angle)
		ux, uy = dir.X, dir.Y
	}
	perp := geom.Pt{X: uy, Y: -ux}
	arm := math.Max(armA, armB) + fraction*length
	c1 := geom.Pt{X: posA.X + perp.X*arm, Y: posA.Y + perp.Y*arm}
	c2 := geom.Pt{X: posB.X + perp.X*arm, Y: posB.Y + perp.Y*arm}
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
	case "simple", "fancy", "wedge":
		return []arrowPathPart{{path: filledArrowPath(pathStart(path), pathEnd(path), s.TailWidth*mutationSize, s.HeadWidth*mutationSize, s.HeadLength*mutationSize), fillable: true}}
	}

	parts := []arrowPathPart{{path: path, fillable: false}}
	beginHead := strings.HasPrefix(name, "<")
	endHead := strings.HasSuffix(name, ">")
	fillBegin := strings.HasPrefix(name, "<|")
	fillEnd := strings.HasSuffix(name, "|>")
	beginBracket := strings.HasPrefix(name, "]") || strings.HasPrefix(name, "|")
	endBracket := strings.HasSuffix(name, "[") || strings.HasSuffix(name, "|")

	if beginHead {
		head := arrowHeadPath(pathSecond(path), pathStart(path), s.HeadLength*mutationSize, s.HeadWidth*mutationSize, fillBegin, lineWidth)
		parts = append(parts, arrowPathPart{path: head, fillable: fillBegin})
	} else if beginBracket {
		parts = append(parts, arrowPathPart{path: bracketPath(pathStart(path), pathSecond(path), s.WidthA*mutationSize, s.LengthA*mutationSize, s.AngleA), fillable: false})
	}
	if endHead {
		head := arrowHeadPath(pathPenultimate(path), pathEnd(path), s.HeadLength*mutationSize, s.HeadWidth*mutationSize, fillEnd, lineWidth)
		parts = append(parts, arrowPathPart{path: head, fillable: fillEnd})
	} else if endBracket {
		parts = append(parts, arrowPathPart{path: bracketPath(pathEnd(path), pathPenultimate(path), s.WidthB*mutationSize, s.LengthB*mutationSize, s.AngleB), fillable: false})
	}
	return parts
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
