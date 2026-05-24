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
	XY           geom.Pt
	Width        float64
	Height       float64
	Pad          float64
	BoxStyle     BoxStyle
	RoundingSize float64
	Coords       CoordinateSpec
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
	path := buildDisplayPath(ctx, r.Coords, rectanglePath(r.Width, r.Height), patchAffine(r.XY, r.Angle))
	r.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the rectangle's data-space bounding box when applicable.
func (r *Rectangle) Bounds(*DrawContext) geom.Rect {
	if r == nil || r.Width == 0 || r.Height == 0 || !isDataCoords(r.Coords) {
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
	path := buildDisplayPath(ctx, c.Coords, local, translateAffine(c.Center))
	c.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the circle's data-space bounding box when applicable.
func (c *Circle) Bounds(*DrawContext) geom.Rect {
	if c == nil || c.Radius <= 0 || !isDataCoords(c.Coords) {
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
	path := buildDisplayPath(ctx, e.Coords, local, patchAffine(e.Center, e.Angle))
	e.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the ellipse's data-space bounding box when applicable.
func (e *Ellipse) Bounds(*DrawContext) geom.Rect {
	if e == nil || e.Width == 0 || e.Height == 0 || !isDataCoords(e.Coords) {
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
	path := buildDisplayPath(ctx, p.Coords, polygonPath(p.XY, !p.Open), geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the polygon's data-space bounding box when applicable.
func (p *Polygon) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.XY) == 0 || !isDataCoords(p.Coords) {
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
	path := buildDisplayPath(ctx, p.Coords, p.Path, geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the path's data-space bounding box when applicable.
func (p *PathPatch) Bounds(*DrawContext) geom.Rect {
	if p == nil || len(p.Path.C) == 0 || !isDataCoords(p.Coords) {
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
	path := buildDisplayPath(ctx, a.Coords, a.localPath(), geom.Identity())
	a.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the arrow's data-space bounding box when applicable.
func (a *FancyArrow) Bounds(*DrawContext) geom.Rect {
	if a == nil || !isDataCoords(a.Coords) {
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
	path := buildDisplayPath(ctx, b.Coords, b.localPath(), translateAffine(b.XY))
	b.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns the fancy bbox's data-space bounding box when applicable.
func (b *FancyBboxPatch) Bounds(*DrawContext) geom.Rect {
	if b == nil || !isDataCoords(b.Coords) {
		return geom.Rect{}
	}
	path := applyAffinePath(b.localPath(), translateAffine(b.XY))
	bounds, _ := pathBounds(path)
	return bounds
}

func (b *FancyBboxPatch) localPath() geom.Path {
	rect := normalizeRect(geom.Rect{
		Min: geom.Pt{X: -b.Pad, Y: -b.Pad},
		Max: geom.Pt{X: b.Width + b.Pad, Y: b.Height + b.Pad},
	})
	if b.BoxStyle == BoxStyleRound {
		radius := b.RoundingSize
		if radius <= 0 {
			radius = math.Min(rect.W(), rect.H()) * 0.2
		}
		return roundedRectPath(rect, radius)
	}
	return patchRectPath(rect)
}

func buildDisplayPath(ctx *DrawContext, coords CoordinateSpec, local geom.Path, localToCoords geom.Affine) geom.Path {
	path := applyAffinePath(local, localToCoords)
	if ctx == nil {
		return path
	}
	tr := ctx.TransformFor(coords)
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
		entry.lineMarkerSet = true
		entry.marker = l.Marker
		entry.markerPath = l.markerPrototypePath(nil, nil)
		entry.markerFill = l.resolvedMarkerFaceColor()
		entry.markerEdge = l.resolvedMarkerEdgeColor()
		entry.markerEdgeWidth = l.resolvedMarkerEdgeWidth()
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
	return legendEntryFromLine(e.Label, e.Color, e.LineWidth, nil), true
}
