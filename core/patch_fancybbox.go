package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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
