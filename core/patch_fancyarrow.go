package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

// Draw renders the fancy arrow using the embedded patch styling.
func (a *FancyArrow) Draw(ren render.Renderer, ctx *DrawContext) {
	if a == nil || ctx == nil || ren == nil {
		return
	}
	path := buildArtistDisplayPath(ctx, a, a.Coords, a.localPath(), geom.Identity())
	a.drawStyledPath(ren, &ctx.RC, path, geom.Path{})
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
