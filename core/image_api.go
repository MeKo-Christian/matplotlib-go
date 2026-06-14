package core

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
)

// ImageOrigin selects how image rows map to the Y-axis direction.
type ImageOrigin int

const (
	// ImageOriginUpper maps row 0 to the upper Y extent.
	ImageOriginUpper ImageOrigin = iota
	// ImageOriginLower maps row 0 to the lower Y extent.
	ImageOriginLower
)

// ImageAnchor selects the point around which image rotation is applied.
type ImageAnchor int

const (
	ImageAnchorCenter ImageAnchor = iota
	ImageAnchorTopLeft
	ImageAnchorTopRight
	ImageAnchorBottomLeft
	ImageAnchorBottomRight
	ImageAnchorCustom
)

const imageDefaultZ = -1100

// ImageOptions controls Image2D rendering.
type ImageOptions struct {
	Colormap *string
	Norm     ScalarNormalizer
	VMin     *float64
	VMax     *float64
	Alpha    *float64
	XMin     *float64
	XMax     *float64
	YMin     *float64
	YMax     *float64
	Origin   ImageOrigin
	Angle    *float64
	// RotationAnchor selects where rotation is centered.
	RotationAnchor ImageAnchor
	// AnchorX/Y are used only when RotationAnchor is ImageAnchorCustom
	// and are interpreted in data coordinates.
	RotationAnchorX *float64
	RotationAnchorY *float64
	Label           string
	// Interpolation selects the filter used when resampling the image.
	// An empty string (the default) lets the renderer choose its default
	// (typically nearest-neighbor). Recognized values mirror matplotlib's
	// imshow interpolation names (e.g. "nearest", "bilinear", "bicubic").
	Interpolation *string
}

// Image2D renders scalar matrix data as an image/heatmap.
type Image2D struct {
	ArtistRasterization
	Data     [][]float64
	Colormap string
	Norm     ScalarNormalizer
	VMin     float64
	VMax     float64
	Alpha    float64
	XMin     float64
	XMax     float64
	YMin     float64
	YMax     float64
	Origin   ImageOrigin
	AngleDeg float64
	RotateAt ImageAnchor
	RotateX  float64
	RotateY  float64
	Label    string
	// Interpolation is the resampling filter name (matplotlib imshow style).
	// An empty string means the renderer's default.
	Interpolation string
	z             float64
}

// Bounds returns the image extent in data space.
func (i *Image2D) Bounds(*DrawContext) geom.Rect {
	if i == nil {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: i.minCoord(i.XMin, i.XMax), Y: i.minCoord(i.YMin, i.YMax)},
		Max: geom.Pt{X: i.maxCoord(i.XMin, i.XMax), Y: i.maxCoord(i.YMin, i.YMax)},
	}
}

// Z returns z-order.
func (i *Image2D) Z() float64 {
	return i.z
}

// ScalarMap exposes the image's scalar mapping for helpers such as colorbars.
func (i *Image2D) ScalarMap() ScalarMapInfo {
	if i == nil {
		return ScalarMapInfo{}
	}
	return ScalarMapInfo{
		Colormap: i.Colormap,
		VMin:     i.VMin,
		VMax:     i.VMax,
		Norm:     i.Norm,
	}
}

// Image creates an Image2D artist for matrix-like heatmap rendering.
func (a *Axes) Image(data [][]float64, opts ...ImageOptions) *Image2D {
	if len(data) == 0 {
		return nil
	}

	var opt ImageOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	rows := len(data)
	cols := 0
	for _, row := range data {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if rows == 0 || cols == 0 {
		return nil
	}

	xMin := 0.0
	xMax := float64(cols)
	yMin := 0.0
	yMax := float64(rows)

	if opt.XMin != nil && !math.IsNaN(*opt.XMin) && !math.IsInf(*opt.XMin, 0) {
		xMin = *opt.XMin
	}
	if opt.XMax != nil && !math.IsNaN(*opt.XMax) && !math.IsInf(*opt.XMax, 0) {
		xMax = *opt.XMax
	}
	if opt.YMin != nil && !math.IsNaN(*opt.YMin) && !math.IsInf(*opt.YMin, 0) {
		yMin = *opt.YMin
	}
	if opt.YMax != nil && !math.IsNaN(*opt.YMax) && !math.IsInf(*opt.YMax, 0) {
		yMax = *opt.YMax
	}

	cmap := "viridis"
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	mapping, err := ResolveScalarMapGrid(data, ScalarMapConfig{
		Colormap: cmap,
		Norm:     opt.Norm,
		VMin:     opt.VMin,
		VMax:     opt.VMax,
	})
	if err != nil {
		return nil
	}

	alpha := 1.0
	if opt.Alpha != nil {
		alpha = clampOneToOne(*opt.Alpha)
	}

	angle := 0.0
	if opt.Angle != nil {
		angle = *opt.Angle
	}

	anchor := opt.RotationAnchor
	rotateX := 0.0
	rotateY := 0.0
	if anchor == ImageAnchorCustom {
		if opt.RotationAnchorX != nil {
			rotateX = *opt.RotationAnchorX
		}
		if opt.RotationAnchorY != nil {
			rotateY = *opt.RotationAnchorY
		}
	}

	interp := ""
	if opt.Interpolation != nil {
		interp = *opt.Interpolation
	}

	image := &Image2D{
		Data:          data,
		Colormap:      mapping.Colormap,
		Norm:          mapping.Norm,
		VMin:          mapping.VMin,
		VMax:          mapping.VMax,
		Alpha:         alpha,
		XMin:          xMin,
		XMax:          xMax,
		YMin:          yMin,
		YMax:          yMax,
		Origin:        opt.Origin,
		AngleDeg:      angle,
		RotateAt:      anchor,
		RotateX:       rotateX,
		RotateY:       rotateY,
		Label:         opt.Label,
		Interpolation: interp,
		z:             imageDefaultZ,
	}
	a.Add(image)
	return image
}

func clampOneToOne(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (i *Image2D) minCoord(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (i *Image2D) maxCoord(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
