package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/transform"
)

const (
	patchCircleSegments = 48
	patchArcKappa       = 0.5522847498307936
)

// BoxStyle controls the shape used by FancyBboxPatch.
type BoxStyle uint8

const (
	BoxStyleSquare BoxStyle = iota
	BoxStyleRound
	BoxStyleCircle
	BoxStyleEllipse
	BoxStyleRArrow
	BoxStyleLArrow
	BoxStyleDArrow
	BoxStyleRound4
	BoxStyleSawtooth
	BoxStyleRoundtooth
)

// Patch stores the common face/edge styling shared by patch-like artists.
//
// In Go this acts as the reusable embedded base for concrete patch artists
// rather than a directly-instantiable drawable type.
type Patch struct {
	ArtistRasterization
	FaceColor render.Color
	EdgeColor render.Color
	EdgeWidth float64
	Alpha     float64
	Dashes    []float64
	Label     string

	Hatch        string
	HatchColor   render.Color
	HatchWidth   float64
	HatchSpacing float64
	PathEffects  []render.PathEffect

	LineJoin render.LineJoin
	LineCap  render.LineCap
	z        float64
}

// Rectangle draws an axis-aligned or rotated rectangle.
type Rectangle struct {
	Patch
	XY     geom.Pt
	Width  float64
	Height float64
	Angle  float64
	Coords CoordinateSpec
}

// Circle draws a circle centered at Center with Radius in the chosen coords.
type Circle struct {
	Patch
	Center geom.Pt
	Radius float64
	Coords CoordinateSpec
}

// Ellipse draws a rotated ellipse centered at Center.
type Ellipse struct {
	Patch
	Center geom.Pt
	Width  float64
	Height float64
	Angle  float64
	Coords CoordinateSpec
}

// Polygon draws a closed polygon by default.
type Polygon struct {
	Patch
	XY     []geom.Pt
	Open   bool
	Coords CoordinateSpec
}

// PathPatch draws an arbitrary path in data, axes, or figure coordinates.
type PathPatch struct {
	Patch
	Path   geom.Path
	Coords CoordinateSpec
}

// FancyArrow draws a filled arrow polygon between XY and XY+{DX,DY}.
type FancyArrow struct {
	Patch
	XY                 geom.Pt
	DX                 float64
	DY                 float64
	Width              float64
	HeadWidth          float64
	HeadLength         float64
	LengthIncludesHead bool
	Coords             CoordinateSpec
}

// Arrow is a convenience alias for FancyArrow with the same behavior.
type Arrow = FancyArrow

// FancyBboxPatch draws a styled box with square or rounded corners.
type FancyBboxPatch struct {
	Patch
	XY             geom.Pt
	Width          float64
	Height         float64
	Pad            float64
	BoxStyle       BoxStyle
	RoundingSize   float64
	ToothSize      float64
	ArrowHeadWidth float64
	ArrowHeadAngle float64
	MutationSize   float64
	MutationAspect float64
	Coords         CoordinateSpec
}

// AddPatch mirrors Matplotlib's patch-oriented API on top of the generic Add.
func (a *Axes) AddPatch(art Artist) {
	if a != nil && art != nil {
		a.Add(art)
	}
}

// Z returns the patch z-order for sorting.
func (p *Patch) Z() float64 {
	if p == nil {
		return 0
	}
	return zOrDefault(p.z, defaultPatchZ)
}

func (p *Patch) legendEntry() (legendEntry, bool) {
	if p == nil || p.Label == "" {
		return legendEntry{}, false
	}
	return legendEntryFromPatchStyle(
		p.Label,
		p.resolvedFaceColor(),
		p.resolvedEdgeColor(),
		p.EdgeWidth,
		p.Hatch,
		p.resolvedHatchColor(),
		p.resolvedHatchWidth(),
	), true
}

func (p *Patch) resolvedFaceColor() render.Color {
	if p == nil {
		return render.Color{}
	}
	return patchAlphaColor(p.FaceColor, p.Alpha)
}

func (p *Patch) resolvedEdgeColor() render.Color {
	if p == nil {
		return render.Color{}
	}
	return patchAlphaColor(p.EdgeColor, p.Alpha)
}

func (p *Patch) resolvedHatchColor() render.Color {
	if p == nil {
		return render.Color{}
	}

	color := p.HatchColor
	if color.A <= 0 {
		color = p.EdgeColor
	}
	if color.A <= 0 {
		color = p.FaceColor
	}
	if color.A <= 0 {
		color = render.Color{R: 0, G: 0, B: 0, A: 1}
	}
	return patchAlphaColor(color, p.Alpha)
}

func (p *Patch) resolvedHatchWidth() float64 {
	if p == nil || p.HatchWidth <= 0 {
		return 100.0 / 72.0
	}
	return p.HatchWidth
}

func (p *Patch) resolvedHatchSpacing() float64 {
	if p == nil || p.HatchSpacing <= 0 {
		return 32
	}
	return p.HatchSpacing
}

func patchAlphaColor(color render.Color, alpha float64) render.Color {
	if alpha > 0 && alpha <= 1 {
		color.A *= alpha
	}
	if color.A < 0 {
		color.A = 0
	}
	if color.A > 1 {
		color.A = 1
	}
	return color
}

func (p *Patch) strokePaint(color render.Color) render.Paint {
	return render.Paint{
		Stroke:      color,
		LineWidth:   p.EdgeWidth,
		LineJoin:    p.LineJoin,
		LineCap:     p.LineCap,
		Dashes:      append([]float64(nil), p.Dashes...),
		PathEffects: cloneRenderPathEffects(p.PathEffects),
	}
}

func (p *Patch) drawStyledPath(r render.Renderer, fillPath, strokePath geom.Path) {
	if p == nil || r == nil {
		return
	}
	if len(fillPath.C) == 0 && len(strokePath.C) == 0 {
		return
	}

	faceColor := p.resolvedFaceColor()
	edgeColor := p.resolvedEdgeColor()
	hasEdge := p.EdgeWidth > 0 && edgeColor.A > 0
	nativeHatch := false
	if hatcher, ok := r.(render.NativeHatcher); ok {
		nativeHatch = hatcher.SupportsNativeHatch()
	}

	if len(fillPath.C) > 0 {
		paint := render.Paint{Fill: faceColor}
		paint.PathEffects = cloneRenderPathEffects(p.PathEffects)
		if nativeHatch && p.Hatch != "" {
			paint.Hatch = p.Hatch
			paint.HatchColor = p.resolvedHatchColor()
			paint.HatchLineWidth = p.resolvedHatchWidth()
			paint.HatchSpacing = p.resolvedHatchSpacing()
		}
		combinedStroke := len(strokePath.C) == 0 && hasEdge
		if combinedStroke {
			if p.Hatch == "" {
				paint = p.strokePaint(edgeColor)
				paint.Fill = faceColor
			} else if nativeHatch {
				paint.Stroke = edgeColor
				paint.LineWidth = p.EdgeWidth
				paint.LineJoin = p.LineJoin
				paint.LineCap = p.LineCap
				paint.Dashes = append([]float64(nil), p.Dashes...)
			}
		}
		if faceColor.A > 0 || combinedStroke || (nativeHatch && p.Hatch != "") {
			r.Path(fillPath, &paint)
		}
	}

	if len(fillPath.C) > 0 && p.Hatch != "" {
		if !nativeHatch {
			p.drawHatch(r, fillPath)
		}
		if !nativeHatch && len(strokePath.C) == 0 && hasEdge {
			paint := p.strokePaint(edgeColor)
			r.Path(fillPath, &paint)
		}
	}

	if len(strokePath.C) > 0 && hasEdge {
		paint := p.strokePaint(edgeColor)
		r.Path(strokePath, &paint)
	}
}

func (p *Patch) drawHatch(r render.Renderer, clipPath geom.Path) {
	render.DrawHatchFallback(r, clipPath, render.Paint{
		Hatch:          p.Hatch,
		HatchColor:     p.resolvedHatchColor(),
		HatchLineWidth: p.resolvedHatchWidth(),
		HatchSpacing:   p.resolvedHatchSpacing(),
	})
}

// Draw renders the rectangle path using the embedded patch styling.
func (r *Rectangle) Draw(ren render.Renderer, ctx *DrawContext) {
	if r == nil || ctx == nil || ren == nil || r.Width == 0 || r.Height == 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, r, r.Coords, rectanglePath(r.Width, r.Height), patchAffine(r.XY, r.Angle))
	r.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the rectangle's data-space bounding box when applicable.
func (r *Rectangle) Bounds(*DrawContext) geom.Rect {
	if r == nil || r.Width == 0 || r.Height == 0 || !artistUsesDataCoords(r, r.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(rectanglePath(r.Width, r.Height), patchAffine(r.XY, r.Angle))
	bounds, _ := pathBounds(path)
	return bounds
}

// Draw renders the circle path using the embedded patch styling.
func (c *Circle) Draw(ren render.Renderer, ctx *DrawContext) {
	if c == nil || ctx == nil || ren == nil || c.Radius <= 0 {
		return
	}
	local := ellipsePath(c.Radius*2, c.Radius*2)
	path := buildArtistDisplayPath(ctx, c, c.Coords, local, translateAffine(c.Center))
	c.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the circle's data-space bounding box when applicable.
func (c *Circle) Bounds(*DrawContext) geom.Rect {
	if c == nil || c.Radius <= 0 || !artistUsesDataCoords(c, c.Coords) {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: c.Center.X - c.Radius, Y: c.Center.Y - c.Radius},
		Max: geom.Pt{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Radius},
	}
}

// Draw renders the ellipse path using the embedded patch styling.
func (e *Ellipse) Draw(ren render.Renderer, ctx *DrawContext) {
	if e == nil || ctx == nil || ren == nil || e.Width == 0 || e.Height == 0 {
		return
	}
	local := ellipsePath(e.Width, e.Height)
	path := buildArtistDisplayPath(ctx, e, e.Coords, local, patchAffine(e.Center, e.Angle))
	e.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the ellipse's data-space bounding box when applicable.
func (e *Ellipse) Bounds(*DrawContext) geom.Rect {
	if e == nil || e.Width == 0 || e.Height == 0 || !artistUsesDataCoords(e, e.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(ellipsePath(e.Width, e.Height), patchAffine(e.Center, e.Angle))
	bounds, _ := pathBounds(path)
	return bounds
}

// Draw renders the polygon path using the embedded patch styling.
func (p *Polygon) Draw(ren render.Renderer, ctx *DrawContext) {
	if p == nil || ctx == nil || ren == nil || len(p.XY) < 2 {
		return
	}
	path := buildArtistDisplayPath(ctx, p, p.Coords, polygonPath(p.XY, !p.Open), geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the polygon's data-space bounding box when applicable.
func (p *Polygon) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.XY) == 0 || !artistUsesDataCoords(p, p.Coords) {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: p.XY[0], Max: p.XY[0]}
	for _, pt := range p.XY[1:] {
		bounds = expandRect(bounds, pt)
	}
	return bounds
}

// Draw renders the path patch using the embedded patch styling.
func (p *PathPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if p == nil || ctx == nil || ren == nil || len(p.Path.C) == 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, p, p.Coords, p.Path, geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the path's data-space bounding box when applicable.
func (p *PathPatch) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.Path.C) == 0 || !artistUsesDataCoords(p, p.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(p.Path)
	return bounds
}

// Draw renders the fancy arrow using the embedded patch styling.
func (a *FancyArrow) Draw(ren render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || ren == nil {
		return
	}
	path := buildArtistDisplayPath(ctx, a, a.Coords, a.localPath(), geom.Identity())
	a.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the arrow's data-space bounding box when applicable.
func (a *FancyArrow) Bounds(*DrawContext) geom.Rect {
	if a == nil || !artistUsesDataCoords(a, a.Coords) {
		return geom.Rect{}
	}
	path := a.localPath()
	bounds, _ := pathBounds(path)
	return bounds
}

func (a *FancyArrow) localPath() geom.Path {
	length := math.Hypot(a.DX, a.DY)
	if length <= 0 {
		return geom.Path{}
	}

	shaftWidth := a.Width
	if shaftWidth <= 0 {
		shaftWidth = length * 0.05
	}

	headWidth := a.HeadWidth
	if headWidth <= 0 {
		headWidth = shaftWidth * 3
	}

	headLength := a.HeadLength
	if headLength <= 0 {
		headLength = math.Max(shaftWidth*2.5, length*0.25)
	}
	if a.LengthIncludesHead && headLength > length {
		headLength = length
	}
	shaftLength := length
	tipLength := length + headLength
	if a.LengthIncludesHead {
		shaftLength = math.Max(0, length-headLength)
		tipLength = length
	}

	local := []geom.Pt{
		{X: 0, Y: -shaftWidth / 2},
		{X: shaftLength, Y: -shaftWidth / 2},
		{X: shaftLength, Y: -headWidth / 2},
		{X: tipLength, Y: 0},
		{X: shaftLength, Y: headWidth / 2},
		{X: shaftLength, Y: shaftWidth / 2},
		{X: 0, Y: shaftWidth / 2},
	}

	angle := math.Atan2(a.DY, a.DX) * 180 / math.Pi
	return applyAffinePath(polygonPath(local, true), patchAffine(a.XY, angle))
}

// Draw renders the fancy bbox path using the embedded patch styling.
func (b *FancyBboxPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if b == nil || ctx == nil || ren == nil {
		return
	}
	path := buildArtistDisplayPath(ctx, b, b.Coords, b.localPath(), translateAffine(b.XY))
	b.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the fancy bbox's data-space bounding box when applicable.
func (b *FancyBboxPatch) Bounds(*DrawContext) geom.Rect {
	if b == nil || !artistUsesDataCoords(b, b.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(b.localPath(), translateAffine(b.XY))
	bounds, _ := pathBounds(path)
	return bounds
}

func (b *FancyBboxPatch) localPath() geom.Path {
	aspect := b.mutationAspect()
	path := b.boxStylePath(0, 0, b.Width, b.Height/aspect, b.mutationSize())
	if aspect != 1 {
		path = applyAffinePath(path, geom.Affine{A: 1, D: aspect})
	}
	return path
}

func (b *FancyBboxPatch) mutationSize() float64 {
	if b == nil || b.MutationSize <= 0 {
		return 1
	}
	return b.MutationSize
}

func (b *FancyBboxPatch) mutationAspect() float64 {
	if b == nil || b.MutationAspect <= 0 {
		return 1
	}
	return b.MutationAspect
}

func (b *FancyBboxPatch) boxStylePath(x0, y0, width, height, mutationSize float64) geom.Path {
	pad := mutationSize * b.Pad
	switch b.BoxStyle {
	case BoxStyleRound:
		return boxStyleRoundPath(x0, y0, width, height, pad, b.RoundingSize, mutationSize)
	case BoxStyleCircle:
		return boxStyleCirclePath(x0, y0, width, height, pad)
	case BoxStyleEllipse:
		return boxStyleEllipsePath(x0, y0, width, height, pad)
	case BoxStyleRArrow:
		return boxStyleArrowPath(x0, y0, width, height, pad, b.arrowHeadWidth(), b.arrowHeadAngle(), false, false)
	case BoxStyleLArrow:
		return boxStyleArrowPath(x0, y0, width, height, pad, b.arrowHeadWidth(), b.arrowHeadAngle(), true, false)
	case BoxStyleDArrow:
		return boxStyleArrowPath(x0, y0, width, height, pad, b.arrowHeadWidth(), b.arrowHeadAngle(), false, true)
	case BoxStyleRound4:
		return boxStyleRound4Path(x0, y0, width, height, pad, b.RoundingSize, mutationSize)
	case BoxStyleSawtooth:
		return boxStyleSawtoothPath(x0, y0, width, height, pad, b.toothSize(mutationSize), false)
	case BoxStyleRoundtooth:
		return boxStyleSawtoothPath(x0, y0, width, height, pad, b.toothSize(mutationSize), true)
	default:
		return boxStyleSquarePath(x0, y0, width, height, pad)
	}
}

func (b *FancyBboxPatch) toothSize(mutationSize float64) float64 {
	if b != nil && b.ToothSize > 0 {
		return b.ToothSize * mutationSize
	}
	if b != nil && b.Pad > 0 {
		return b.Pad * 0.5 * mutationSize
	}
	return 0
}

func (b *FancyBboxPatch) arrowHeadWidth() float64 {
	if b != nil && b.ArrowHeadWidth > 0 {
		return b.ArrowHeadWidth
	}
	return 1.5
}

func (b *FancyBboxPatch) arrowHeadAngle() float64 {
	if b != nil && math.Mod(b.ArrowHeadAngle, 360) != 0 {
		return b.ArrowHeadAngle
	}
	return 90
}

func boxStyleSquarePath(x0, y0, width, height, pad float64) geom.Path {
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: x0 - pad, Y: y0 - pad},
		Max: geom.Pt{X: x0 + width + pad, Y: y0 + height + pad},
	})
}

func boxStyleRoundPath(x0, y0, width, height, pad, roundingSize, mutationSize float64) geom.Path {
	dr := pad
	if roundingSize > 0 {
		dr = mutationSize * roundingSize
	}
	width += 2 * pad
	height += 2 * pad
	x0 -= pad
	y0 -= pad
	x1, y1 := x0+width, y0+height

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x0 + dr, Y: y0})
	path.LineTo(geom.Pt{X: x1 - dr, Y: y0})
	path.QuadTo(geom.Pt{X: x1, Y: y0}, geom.Pt{X: x1, Y: y0 + dr})
	path.LineTo(geom.Pt{X: x1, Y: y1 - dr})
	path.QuadTo(geom.Pt{X: x1, Y: y1}, geom.Pt{X: x1 - dr, Y: y1})
	path.LineTo(geom.Pt{X: x0 + dr, Y: y1})
	path.QuadTo(geom.Pt{X: x0, Y: y1}, geom.Pt{X: x0, Y: y1 - dr})
	path.LineTo(geom.Pt{X: x0, Y: y0 + dr})
	path.QuadTo(geom.Pt{X: x0, Y: y0}, geom.Pt{X: x0 + dr, Y: y0})
	path.Close()
	return path
}

func boxStyleCirclePath(x0, y0, width, height, pad float64) geom.Path {
	width += 2 * pad
	height += 2 * pad
	x0 -= pad
	y0 -= pad
	radius := math.Max(width, height) / 2
	return ellipseBezierPath(geom.Pt{X: x0 + width/2, Y: y0 + height/2}, radius, radius)
}

func boxStyleEllipsePath(x0, y0, width, height, pad float64) geom.Path {
	width += 2 * pad
	height += 2 * pad
	x0 -= pad
	y0 -= pad
	return ellipseBezierPath(
		geom.Pt{X: x0 + width/2, Y: y0 + height/2},
		width/math.Sqrt2,
		height/math.Sqrt2,
	)
}

func boxStyleRound4Path(x0, y0, width, height, pad, roundingSize, mutationSize float64) geom.Path {
	dr := pad / 2
	if roundingSize > 0 {
		dr = mutationSize * roundingSize
	}
	width = width + 2*pad - 2*dr
	height = height + 2*pad - 2*dr
	x0 = x0 - pad + dr
	y0 = y0 - pad + dr
	x1, y1 := x0+width, y0+height

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x0, Y: y0})
	path.CubicTo(
		geom.Pt{X: x0 + dr, Y: y0 - dr},
		geom.Pt{X: x1 - dr, Y: y0 - dr},
		geom.Pt{X: x1, Y: y0},
	)
	path.CubicTo(
		geom.Pt{X: x1 + dr, Y: y0 + dr},
		geom.Pt{X: x1 + dr, Y: y1 - dr},
		geom.Pt{X: x1, Y: y1},
	)
	path.CubicTo(
		geom.Pt{X: x1 - dr, Y: y1 + dr},
		geom.Pt{X: x0 + dr, Y: y1 + dr},
		geom.Pt{X: x0, Y: y1},
	)
	path.CubicTo(
		geom.Pt{X: x0 - dr, Y: y1 - dr},
		geom.Pt{X: x0 - dr, Y: y0 + dr},
		geom.Pt{X: x0, Y: y0},
	)
	path.Close()
	return path
}

func boxStyleArrowPath(x0, y0, width, height, pad, headWidth, headAngle float64, left, double bool) geom.Path {
	origX0, origWidth := x0, width
	dx, dy := width+2*pad, height+2*pad
	x0 -= pad
	y0 -= pad
	x1, y1 := x0+dx, y0+dy

	headDY := headWidth * dy
	midY := (y0 + y1) / 2
	shaftY0 := midY - headDY/2
	shaftY1 := midY + headDY/2
	cot := 1 / math.Tan((headAngle/2)*math.Pi/180)

	var points []geom.Pt
	if double {
		if cot > 0 {
			tipX0 := x0 - cot*math.Min(dy, headDY)/2
			shaftX0 := tipX0 + cot*headDY/2
			tipX1 := x1 + cot*math.Min(dy, headDY)/2
			shaftX1 := tipX1 - cot*headDY/2
			points = []geom.Pt{
				{X: shaftX0, Y: y1}, {X: shaftX0, Y: shaftY1},
				{X: tipX0, Y: midY},
				{X: shaftX0, Y: shaftY0}, {X: shaftX0, Y: y0},
				{X: shaftX1, Y: y0}, {X: shaftX1, Y: shaftY0},
				{X: tipX1, Y: midY},
				{X: shaftX1, Y: shaftY1}, {X: shaftX1, Y: y1},
			}
		} else {
			shift := math.Min(-cot*math.Max(headDY-dy, 0)/2, dx/2)
			midY0 := math.Min(shaftY0, y0) - shift/cot
			midY1 := math.Max(shaftY1, y1) + shift/cot
			points = []geom.Pt{
				{X: x0, Y: shaftY0}, {X: x0 + shift, Y: midY0},
				{X: x1 - shift, Y: midY0}, {X: x1, Y: shaftY0},
				{X: x1, Y: shaftY1}, {X: x1 - shift, Y: midY1},
				{X: x0 + shift, Y: midY1}, {X: x0, Y: shaftY1},
			}
		}
		return polygonPath(points, true)
	}

	if cot > 0 {
		tipX := x1 + cot*math.Min(dy, headDY)/2
		shaftX := tipX - cot*headDY/2
		points = []geom.Pt{
			{X: x0, Y: y0}, {X: shaftX, Y: y0}, {X: shaftX, Y: shaftY0},
			{X: tipX, Y: midY},
			{X: shaftX, Y: shaftY1}, {X: shaftX, Y: y1}, {X: x0, Y: y1},
		}
	} else {
		shift := math.Min(-cot*math.Max(headDY-dy, 0)/2, dx)
		midY0 := math.Min(shaftY0, y0) - shift/cot
		midY1 := math.Max(shaftY1, y1) + shift/cot
		points = []geom.Pt{
			{X: x0, Y: y0}, {X: x1 - shift, Y: midY0}, {X: x1, Y: shaftY0},
			{X: x1, Y: shaftY1}, {X: x1 - shift, Y: midY1}, {X: x0, Y: y1},
		}
	}
	if left {
		for i := range points {
			points[i].X = 2*origX0 + origWidth - points[i].X
		}
	}
	return polygonPath(points, true)
}

func boxStyleSawtoothPath(x0, y0, width, height, pad, toothSize float64, round bool) geom.Path {
	if toothSize <= 0 {
		return boxStyleSquarePath(x0, y0, width, height, pad)
	}
	vertices := sawtoothVertices(x0, y0, width, height, pad, toothSize)
	if len(vertices) == 0 {
		return geom.Path{}
	}
	if !round {
		if len(vertices) > 1 && approxPtCore(vertices[0], vertices[len(vertices)-1], 1e-12) {
			vertices = vertices[:len(vertices)-1]
		}
		return polygonPath(vertices, true)
	}

	if len(vertices) < 3 {
		return polygonPath(vertices, true)
	}
	if !approxPtCore(vertices[0], vertices[len(vertices)-1], 1e-12) {
		vertices = append(vertices, vertices[0])
	}
	vertices = append(vertices, vertices[0])
	path := geom.Path{}
	path.MoveTo(vertices[0])
	for i := 1; i+1 < len(vertices); i += 2 {
		path.QuadTo(vertices[i], vertices[i+1])
	}
	path.Close()
	return path
}

func sawtoothVertices(x0, y0, width, height, pad, toothSize float64) []geom.Pt {
	half := toothSize / 2
	width = width + 2*pad - toothSize
	height = height + 2*pad - toothSize
	if width <= 0 || height <= 0 {
		return nil
	}
	dsx := int(math.Round((width-toothSize)/(toothSize*2))) * 2
	dsy := int(math.Round((height-toothSize)/(toothSize*2))) * 2
	if dsx < 0 {
		dsx = 0
	}
	if dsy < 0 {
		dsy = 0
	}

	x0 = x0 - pad + half
	y0 = y0 - pad + half
	x1, y1 := x0+width, y0+height

	xs := []float64{x0}
	xs = append(xs, linspace(x0+half, x1-half, 2*dsx+1)...)
	for i := 0; i < 2*dsy+2; i++ {
		switch i % 4 {
		case 0, 2:
			xs = append(xs, x1)
		case 1:
			xs = append(xs, x1+half)
		default:
			xs = append(xs, x1-half)
		}
	}
	xs = append(xs, x1)
	xs = append(xs, linspace(x1-half, x0+half, 2*dsx+1)...)
	for i := 0; i < 2*dsy+2; i++ {
		switch i % 4 {
		case 0, 2:
			xs = append(xs, x0)
		case 1:
			xs = append(xs, x0-half)
		default:
			xs = append(xs, x0+half)
		}
	}

	ys := make([]float64, 0, len(xs))
	for i := 0; i < 2*dsx+2; i++ {
		switch i % 4 {
		case 0, 2:
			ys = append(ys, y0)
		case 1:
			ys = append(ys, y0-half)
		default:
			ys = append(ys, y0+half)
		}
	}
	ys = append(ys, y0)
	ys = append(ys, linspace(y0+half, y1-half, 2*dsy+1)...)
	for i := 0; i < 2*dsx+2; i++ {
		switch i % 4 {
		case 0, 2:
			ys = append(ys, y1)
		case 1:
			ys = append(ys, y1+half)
		default:
			ys = append(ys, y1-half)
		}
	}
	ys = append(ys, y1)
	ys = append(ys, linspace(y1-half, y0+half, 2*dsy+1)...)

	n := min(len(xs), len(ys))
	vertices := make([]geom.Pt, 0, n+1)
	for i := 0; i < n; i++ {
		vertices = append(vertices, geom.Pt{X: xs[i], Y: ys[i]})
	}
	if len(vertices) > 0 {
		vertices = append(vertices, vertices[0])
	}
	return vertices
}

func linspace(start, end float64, count int) []float64 {
	if count <= 0 {
		return nil
	}
	if count == 1 {
		return []float64{start}
	}
	step := (end - start) / float64(count-1)
	out := make([]float64, count)
	for i := range out {
		out[i] = start + float64(i)*step
	}
	return out
}

func ellipseBezierPath(center geom.Pt, rx, ry float64) geom.Path {
	if rx <= 0 || ry <= 0 {
		return geom.Path{}
	}
	kx := rx * patchArcKappa
	ky := ry * patchArcKappa
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: center.X + rx, Y: center.Y})
	path.CubicTo(
		geom.Pt{X: center.X + rx, Y: center.Y + ky},
		geom.Pt{X: center.X + kx, Y: center.Y + ry},
		geom.Pt{X: center.X, Y: center.Y + ry},
	)
	path.CubicTo(
		geom.Pt{X: center.X - kx, Y: center.Y + ry},
		geom.Pt{X: center.X - rx, Y: center.Y + ky},
		geom.Pt{X: center.X - rx, Y: center.Y},
	)
	path.CubicTo(
		geom.Pt{X: center.X - rx, Y: center.Y - ky},
		geom.Pt{X: center.X - kx, Y: center.Y - ry},
		geom.Pt{X: center.X, Y: center.Y - ry},
	)
	path.CubicTo(
		geom.Pt{X: center.X + kx, Y: center.Y - ry},
		geom.Pt{X: center.X + rx, Y: center.Y - ky},
		geom.Pt{X: center.X + rx, Y: center.Y},
	)
	path.Close()
	return path
}

func approxPtCore(a, b geom.Pt, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func buildDisplayPath(ctx *DrawContext, coords CoordinateSpec, local geom.Path, localToCoords geom.Affine) geom.Path {
	return buildArtistDisplayPath(ctx, nil, coords, local, localToCoords)
}

func buildArtistDisplayPath(ctx *DrawContext, art any, fallback CoordinateSpec, local geom.Path, localToCoords geom.Affine) geom.Path {
	path := applyAffinePath(local, localToCoords)
	if ctx == nil {
		return path
	}
	tr := artistTransformFor(ctx, art, fallback)
	if tr == nil {
		return path
	}
	return applyTransformPath(path, tr)
}

func isDataCoords(spec CoordinateSpec) bool {
	return spec.X == CoordData && spec.Y == CoordData
}

func patchAffine(origin geom.Pt, angleDeg float64) geom.Affine {
	rad := angleDeg * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	return geom.Affine{
		A: cosA,
		B: sinA,
		C: -sinA,
		D: cosA,
		E: origin.X,
		F: origin.Y,
	}
}

func translateAffine(offset geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: offset.X, F: offset.Y}
}

func patchRectPath(rect geom.Rect) geom.Path {
	rect = normalizeRect(rect)
	return polygonPath([]geom.Pt{
		rect.Min,
		{X: rect.Max.X, Y: rect.Min.Y},
		rect.Max,
		{X: rect.Min.X, Y: rect.Max.Y},
	}, true)
}

func rectanglePath(width, height float64) geom.Path {
	return patchRectPath(geom.Rect{
		Min: geom.Pt{X: math.Min(0, width), Y: math.Min(0, height)},
		Max: geom.Pt{X: math.Max(0, width), Y: math.Max(0, height)},
	})
}

func roundedRectPath(rect geom.Rect, radius float64) geom.Path {
	rect = normalizeRect(rect)
	if rect.W() == 0 || rect.H() == 0 {
		return geom.Path{}
	}

	maxRadius := math.Min(rect.W(), rect.H()) / 2
	if radius <= 0 {
		return patchRectPath(rect)
	}
	if radius > maxRadius {
		radius = maxRadius
	}

	left, top := rect.Min.X, rect.Min.Y
	right, bottom := rect.Max.X, rect.Max.Y
	k := radius * patchArcKappa

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: left + radius, Y: top})
	path.LineTo(geom.Pt{X: right - radius, Y: top})
	path.CubicTo(
		geom.Pt{X: right - radius + k, Y: top},
		geom.Pt{X: right, Y: top + radius - k},
		geom.Pt{X: right, Y: top + radius},
	)
	path.LineTo(geom.Pt{X: right, Y: bottom - radius})
	path.CubicTo(
		geom.Pt{X: right, Y: bottom - radius + k},
		geom.Pt{X: right - radius + k, Y: bottom},
		geom.Pt{X: right - radius, Y: bottom},
	)
	path.LineTo(geom.Pt{X: left + radius, Y: bottom})
	path.CubicTo(
		geom.Pt{X: left + radius - k, Y: bottom},
		geom.Pt{X: left, Y: bottom - radius + k},
		geom.Pt{X: left, Y: bottom - radius},
	)
	path.LineTo(geom.Pt{X: left, Y: top + radius})
	path.CubicTo(
		geom.Pt{X: left, Y: top + radius - k},
		geom.Pt{X: left + radius - k, Y: top},
		geom.Pt{X: left + radius, Y: top},
	)
	path.Close()
	return path
}

func ellipsePath(width, height float64) geom.Path {
	rx := math.Abs(width) / 2
	ry := math.Abs(height) / 2
	if rx == 0 || ry == 0 {
		return geom.Path{}
	}

	points := make([]geom.Pt, 0, patchCircleSegments)
	for i := 0; i < patchCircleSegments; i++ {
		angle := 2 * math.Pi * float64(i) / patchCircleSegments
		points = append(points, geom.Pt{
			X: rx * math.Cos(angle),
			Y: ry * math.Sin(angle),
		})
	}
	return polygonPath(points, true)
}

func polygonPath(points []geom.Pt, close bool) geom.Path {
	if len(points) == 0 {
		return geom.Path{}
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	if close {
		path.Close()
	}
	return path
}

func applyAffinePath(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}

func applyTransformPath(path geom.Path, tr transform.T) geom.Path {
	if len(path.C) == 0 || tr == nil {
		return path
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = tr.Apply(pt)
	}
	return out
}

func pathBounds(path geom.Path) (geom.Rect, bool) {
	if len(path.V) == 0 {
		return geom.Rect{}, false
	}
	bounds := geom.Rect{Min: path.V[0], Max: path.V[0]}
	for _, pt := range path.V[1:] {
		bounds = expandRect(bounds, pt)
	}
	return bounds, true
}

func normalizeRect(rect geom.Rect) geom.Rect {
	return geom.Rect{
		Min: geom.Pt{X: math.Min(rect.Min.X, rect.Max.X), Y: math.Min(rect.Min.Y, rect.Max.Y)},
		Max: geom.Pt{X: math.Max(rect.Min.X, rect.Max.X), Y: math.Max(rect.Min.Y, rect.Max.Y)},
	}
}

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
