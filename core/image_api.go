package core

import (
	"image"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/style"
)

// imageInterpolationDefault maps the image.interpolation rcParam to the renderer
// convention, where "auto" means the renderer chooses (empty string).
func imageInterpolationDefault(interpolation string) string {
	if strings.EqualFold(strings.TrimSpace(interpolation), "auto") {
		return ""
	}
	return interpolation
}

// imageOriginFromRC maps the image.origin rcParam ("upper"/"lower") to the
// ImageOrigin enum. Unknown values fall back to upper, the matplotlib default.
func imageOriginFromRC(rc *style.RC) ImageOrigin {
	if strings.EqualFold(strings.TrimSpace(rc.Image.Origin), "lower") {
		return ImageOriginLower
	}
	return ImageOriginUpper
}

// imshowAspectDefault returns the image.aspect rcParam ("equal", "auto", or a
// numeric ratio), falling back to matplotlib's "equal" default when unset.
func imshowAspectDefault(rc *style.RC) string {
	if aspect := strings.TrimSpace(rc.Image.Aspect); aspect != "" {
		return aspect
	}
	return "equal"
}

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
	// rgba holds pre-colored pixels for native RGB/RGBA imshow. When non-nil
	// the colormap+norm path is bypassed and Data is nil. The origin flip (for
	// ImageOriginLower) is applied at rasterization time, not baked in here.
	rgba *image.RGBA
	z    float64
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
	if i == nil || i.rgba != nil {
		// True-color (RGB/RGBA) images are not scalar-mapped, so they expose no
		// colormap/norm for colorbars (mirrors matplotlib: RGB images are not
		// ScalarMappable-colored).
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

	rc := a.resolvedRC()
	cmap := rc.Image.Cmap
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

	g := resolveImageGeometry(opt, cols, rows)
	if opt.Interpolation == nil {
		g.interp = imageInterpolationDefault(rc.Image.Interpolation)
	}
	image := &Image2D{
		Data:          data,
		Colormap:      mapping.Colormap,
		Norm:          mapping.Norm,
		VMin:          mapping.VMin,
		VMax:          mapping.VMax,
		Alpha:         g.alpha,
		XMin:          g.xMin,
		XMax:          g.xMax,
		YMin:          g.yMin,
		YMax:          g.yMax,
		Origin:        opt.Origin,
		AngleDeg:      g.angle,
		RotateAt:      g.anchor,
		RotateX:       g.rotateX,
		RotateY:       g.rotateY,
		Label:         opt.Label,
		Interpolation: g.interp,
		z:             imageDefaultZ,
	}
	a.Add(image)
	return image
}

// imageGeometry collects the extent/alpha/rotation/interpolation fields shared
// by the scalar Image and the true-color imageRGBA constructors.
type imageGeometry struct {
	xMin, xMax, yMin, yMax float64
	alpha, angle           float64
	anchor                 ImageAnchor
	rotateX, rotateY       float64
	interp                 string
}

// resolveImageGeometry derives the placement-related Image2D fields from
// ImageOptions, defaulting the extent to the pixel grid [0,cols]×[0,rows].
func resolveImageGeometry(opt ImageOptions, cols, rows int) imageGeometry {
	g := imageGeometry{
		xMin:   0,
		xMax:   float64(cols),
		yMin:   0,
		yMax:   float64(rows),
		alpha:  1,
		anchor: opt.RotationAnchor,
	}
	if opt.XMin != nil && !math.IsNaN(*opt.XMin) && !math.IsInf(*opt.XMin, 0) {
		g.xMin = *opt.XMin
	}
	if opt.XMax != nil && !math.IsNaN(*opt.XMax) && !math.IsInf(*opt.XMax, 0) {
		g.xMax = *opt.XMax
	}
	if opt.YMin != nil && !math.IsNaN(*opt.YMin) && !math.IsInf(*opt.YMin, 0) {
		g.yMin = *opt.YMin
	}
	if opt.YMax != nil && !math.IsNaN(*opt.YMax) && !math.IsInf(*opt.YMax, 0) {
		g.yMax = *opt.YMax
	}
	if opt.Alpha != nil {
		g.alpha = clampOneToOne(*opt.Alpha)
	}
	if opt.Angle != nil {
		g.angle = *opt.Angle
	}
	if g.anchor == ImageAnchorCustom {
		if opt.RotationAnchorX != nil {
			g.rotateX = *opt.RotationAnchorX
		}
		if opt.RotationAnchorY != nil {
			g.rotateY = *opt.RotationAnchorY
		}
	}
	if opt.Interpolation != nil {
		g.interp = *opt.Interpolation
	}
	return g
}

// imageRGBA constructs an Image2D backed by pre-colored RGBA pixels, bypassing
// the colormap+norm path. It mirrors Image's option handling minus the scalar
// mapping. The extent defaults to the pixel grid unless overridden via opt.
func (a *Axes) imageRGBA(rgba *image.RGBA, opt ImageOptions) *Image2D {
	if a == nil || rgba == nil {
		return nil
	}
	b := rgba.Bounds()
	cols := b.Dx()
	rows := b.Dy()
	if cols == 0 || rows == 0 {
		return nil
	}

	g := resolveImageGeometry(opt, cols, rows)
	if opt.Interpolation == nil {
		g.interp = imageInterpolationDefault(a.resolvedRC().Image.Interpolation)
	}
	img := &Image2D{
		rgba:          rgba,
		Alpha:         g.alpha,
		XMin:          g.xMin,
		XMax:          g.xMax,
		YMin:          g.yMin,
		YMax:          g.yMax,
		Origin:        opt.Origin,
		AngleDeg:      g.angle,
		RotateAt:      g.anchor,
		RotateX:       g.rotateX,
		RotateY:       g.rotateY,
		Label:         opt.Label,
		Interpolation: g.interp,
		z:             imageDefaultZ,
	}
	a.Add(img)
	return img
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
