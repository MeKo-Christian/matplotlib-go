package core

import (
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Shadow draws a darkened, offset copy of another patch.
type Shadow struct {
	Patch
	Source Artist
	Offset geom.Pt
	Shade  float64
}

// RegularPolygon draws a regular polygon centered at Center.
type RegularPolygon struct {
	Patch
	Center      geom.Pt
	NumVertices int
	Radius      float64
	Orientation float64
	Coords      CoordinateSpec
}

// CirclePolygon draws a polygonal approximation of a circle.
type CirclePolygon struct {
	Patch
	Center     geom.Pt
	Radius     float64
	Resolution int
	Coords     CoordinateSpec
}

// Arc draws an unfilled elliptical arc.
type Arc struct {
	Patch
	Center   geom.Pt
	Width    float64
	Height   float64
	Angle    float64
	Theta1   float64
	Theta2   float64
	EdgeOnly bool
	Coords   CoordinateSpec
}

// Annulus draws an elliptical ring.
type Annulus struct {
	Patch
	Center  geom.Pt
	RadiusA float64
	RadiusB float64
	Width   float64
	Angle   float64
	Coords  CoordinateSpec
}

// StepPatch draws a stepwise-constant function as a patch path.
type StepPatch struct {
	Patch
	Values      []float64
	Edges       []float64
	Baseline    *float64
	Orientation string
	Coords      CoordinateSpec
}

// Draw renders the shadow path using darkened source patch styling.
func (s *Shadow) Draw(ren render.Renderer, ctx *DrawContext) {
	if s == nil || ren == nil || ctx == nil || s.Source == nil {
		return
	}
	path, ok := sourcePatchDisplayPath(s.Source, ctx)
	if !ok {
		return
	}
	if s.Offset != (geom.Pt{}) {
		path = applyAffinePath(path, translateAffine(s.Offset))
	}
	shadow := s.Patch
	src := sourcePatchStyle(s.Source)
	if shadow.FaceColor == (render.Color{}) && src != nil {
		shadow.FaceColor = shadowColor(src.resolvedFaceColor(), s.Shade)
	}
	if shadow.EdgeColor == (render.Color{}) {
		shadow.EdgeColor = shadow.FaceColor
	}
	if shadow.Alpha <= 0 {
		shadow.Alpha = 1
	}
	if shadow.FaceColor.A <= 0 && shadow.FaceColor != (render.Color{}) {
		shadow.FaceColor.A = 0.5
	}
	if shadow.EdgeColor.A <= 0 && shadow.EdgeColor != (render.Color{}) {
		shadow.EdgeColor.A = shadow.FaceColor.A
	}
	if shadow.EdgeWidth <= 0 && src != nil {
		shadow.EdgeWidth = src.EdgeWidth
	}
	shadow.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns an empty rect because shadows should not affect autoscaling.
func (s *Shadow) Bounds(*DrawContext) geom.Rect { return geom.Rect{} }

// Draw renders the regular polygon path.
func (p *RegularPolygon) Draw(ren render.Renderer, ctx *DrawContext) {
	if p == nil || ren == nil || ctx == nil || p.NumVertices < 3 || p.Radius <= 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity())
	p.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns data-space bounds when applicable.
func (p *RegularPolygon) Bounds(*DrawContext) geom.Rect {
	if p == nil || p.NumVertices < 3 || p.Radius <= 0 || !artistUsesDataCoords(p, p.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(p.localPath())
	return bounds
}

func (p *RegularPolygon) localPath() geom.Path {
	return patchRegularPolygonPath(p.Center, p.NumVertices, p.Radius, p.Orientation)
}

// Draw renders the circle polygon path.
func (c *CirclePolygon) Draw(ren render.Renderer, ctx *DrawContext) {
	if c == nil || ren == nil || ctx == nil || c.Radius <= 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, c, c.Coords, c.localPath(), geom.Identity())
	c.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns data-space bounds when applicable.
func (c *CirclePolygon) Bounds(*DrawContext) geom.Rect {
	if c == nil || c.Radius <= 0 || !artistUsesDataCoords(c, c.Coords) {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: c.Center.X - c.Radius, Y: c.Center.Y - c.Radius},
		Max: geom.Pt{X: c.Center.X + c.Radius, Y: c.Center.Y + c.Radius},
	}
}

func (c *CirclePolygon) localPath() geom.Path {
	resolution := c.Resolution
	if resolution <= 0 {
		resolution = 20
	}
	return patchRegularPolygonPath(c.Center, resolution, c.Radius, 0)
}

// Draw renders the arc as a stroked path.
func (a *Arc) Draw(ren render.Renderer, ctx *DrawContext) {
	if a == nil || ren == nil || ctx == nil || a.Width == 0 || a.Height == 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, a, a.Coords, a.localPath(), geom.Identity())
	a.drawStyledPath(ren, geom.Path{}, path)
}

// Bounds returns data-space bounds when applicable.
func (a *Arc) Bounds(*DrawContext) geom.Rect {
	if a == nil || a.Width == 0 || a.Height == 0 || !artistUsesDataCoords(a, a.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(a.localPath())
	return bounds
}

func (a *Arc) localPath() geom.Path {
	path := matplotlibArcPath(geom.Pt{}, 1, a.Theta1, a.Theta2)
	affine := translateAffine(a.Center).
		Mul(rotationAffine(a.Angle)).
		Mul(geom.Affine{A: a.Width / 2, D: a.Height / 2})
	return applyAffinePath(path, affine)
}

// Draw renders the annulus path.
func (a *Annulus) Draw(ren render.Renderer, ctx *DrawContext) {
	if a == nil || ren == nil || ctx == nil || a.RadiusA <= 0 || a.Width <= 0 {
		return
	}
	path := buildArtistDisplayPath(ctx, a, a.Coords, a.localPath(), geom.Identity())
	a.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns data-space bounds when applicable.
func (a *Annulus) Bounds(*DrawContext) geom.Rect {
	if a == nil || a.RadiusA <= 0 || !artistUsesDataCoords(a, a.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(a.localPath())
	return bounds
}

func (a *Annulus) localPath() geom.Path {
	rx := a.RadiusA
	ry := a.RadiusB
	if ry <= 0 {
		ry = rx
	}
	innerX := rx - a.Width
	innerY := ry - a.Width
	if innerX < 0 {
		innerX = 0
	}
	if innerY < 0 {
		innerY = 0
	}
	outer := ellipsePath(rx*2, ry*2)
	inner := ellipsePath(innerX*2, innerY*2)
	affine := translateAffine(a.Center).Mul(rotationAffine(a.Angle))
	out := geom.Path{}
	appendPath := func(path geom.Path, reverse bool) {
		points := append([]geom.Pt(nil), path.V...)
		if reverse {
			for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
				points[i], points[j] = points[j], points[i]
			}
		}
		for i, pt := range points {
			pt = affine.Apply(pt)
			if i == 0 {
				out.MoveTo(pt)
			} else {
				out.LineTo(pt)
			}
		}
		out.Close()
	}
	appendPath(outer, false)
	if innerX > 0 && innerY > 0 {
		appendPath(inner, true)
	}
	return out
}

// Draw renders the step patch path.
func (s *StepPatch) Draw(ren render.Renderer, ctx *DrawContext) {
	if s == nil || ren == nil || ctx == nil || len(s.Values) == 0 || len(s.Edges) != len(s.Values)+1 {
		return
	}
	path := buildArtistDisplayPath(ctx, s, s.Coords, s.localPath(), geom.Identity())
	s.drawStyledPath(ren, path, geom.Path{})
}

// Bounds returns data-space bounds when applicable.
func (s *StepPatch) Bounds(*DrawContext) geom.Rect {
	if s == nil || len(s.Values) == 0 || len(s.Edges) != len(s.Values)+1 || !artistUsesDataCoords(s, s.Coords) {
		return geom.Rect{}
	}
	bounds, _ := pathBounds(s.localPath())
	return bounds
}

func (s *StepPatch) localPath() geom.Path {
	if len(s.Values) == 0 || len(s.Edges) != len(s.Values)+1 {
		return geom.Path{}
	}
	vertical := !strings.EqualFold(s.Orientation, "horizontal")
	xy := func(edge, value float64) geom.Pt {
		if vertical {
			return geom.Pt{X: edge, Y: value}
		}
		return geom.Pt{X: value, Y: edge}
	}
	path := geom.Path{}
	if s.Baseline != nil {
		path.MoveTo(xy(s.Edges[0], *s.Baseline))
		path.LineTo(xy(s.Edges[0], s.Values[0]))
	} else {
		path.MoveTo(xy(s.Edges[0], s.Values[0]))
	}
	for i, value := range s.Values {
		path.LineTo(xy(s.Edges[i+1], value))
		if i+1 < len(s.Values) {
			path.LineTo(xy(s.Edges[i+1], s.Values[i+1]))
		}
	}
	if s.Baseline != nil {
		path.LineTo(xy(s.Edges[len(s.Edges)-1], *s.Baseline))
		path.Close()
	}
	return path
}

func sourcePatchDisplayPath(source Artist, ctx *DrawContext) (geom.Path, bool) {
	switch p := source.(type) {
	case *Rectangle:
		return buildArtistDisplayPath(ctx, p, p.Coords, rectanglePath(p.Width, p.Height), patchAffine(p.XY, p.Angle)), true
	case *Circle:
		return buildArtistDisplayPath(ctx, p, p.Coords, ellipsePath(p.Radius*2, p.Radius*2), translateAffine(p.Center)), true
	case *Ellipse:
		return buildArtistDisplayPath(ctx, p, p.Coords, ellipsePath(p.Width, p.Height), patchAffine(p.Center, p.Angle)), true
	case *Polygon:
		return buildArtistDisplayPath(ctx, p, p.Coords, polygonPath(p.XY, !p.Open), geom.Identity()), true
	case *PathPatch:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.Path, geom.Identity()), true
	case *FancyBboxPatch:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), translateAffine(p.XY)), true
	case *RegularPolygon:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity()), true
	case *CirclePolygon:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity()), true
	case *Annulus:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity()), true
	case *StepPatch:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity()), true
	case *Wedge:
		return buildArtistDisplayPath(ctx, p, p.Coords, p.localPath(), geom.Identity()), true
	default:
		return geom.Path{}, false
	}
}

func sourcePatchStyle(source Artist) *Patch {
	switch p := source.(type) {
	case *Rectangle:
		return &p.Patch
	case *Circle:
		return &p.Patch
	case *Ellipse:
		return &p.Patch
	case *Polygon:
		return &p.Patch
	case *PathPatch:
		return &p.Patch
	case *FancyArrow:
		return &p.Patch
	case *FancyBboxPatch:
		return &p.Patch
	case *RegularPolygon:
		return &p.Patch
	case *CirclePolygon:
		return &p.Patch
	case *Arc:
		return &p.Patch
	case *Annulus:
		return &p.Patch
	case *StepPatch:
		return &p.Patch
	case *Wedge:
		return &p.Patch
	default:
		return nil
	}
}

func shadowColor(c render.Color, shade float64) render.Color {
	if shade <= 0 {
		shade = 0.7
	}
	if shade > 1 {
		shade = 1
	}
	return render.Color{R: c.R * (1 - shade), G: c.G * (1 - shade), B: c.B * (1 - shade), A: 0.5}
}

func patchRegularPolygonPath(center geom.Pt, sides int, radius, orientationRad float64) geom.Path {
	if sides < 3 || radius <= 0 {
		return geom.Path{}
	}
	return geom.UnitRegularPolygon(sides).Transformed(
		unitPolyAffine(center, radius, orientationRad),
	)
}

func rotationAffine(angleDeg float64) geom.Affine {
	rad := angleDeg * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)
	return geom.Affine{A: cosA, B: sinA, C: -sinA, D: cosA}
}
