package pgf

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawImage draws an RGBA raster image as deterministic pure-PGF pixel rectangles.
// This keeps .pgf output self-contained at the cost of large output for dense
// images; callers that need compact publication files should prefer PDF/SVG.
func (r *Renderer) DrawImage(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.DrawImage(img, dst)
		return
	}
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	width, height := img.Size()
	if width <= 0 || height <= 0 {
		return
	}
	// Display space is y-up and the raster's row 0 is its top row, so map row 0
	// to the top edge dst.Max.Y with rows advancing downward (negative D). This
	// matches core's imageTransform convention and the rotated ImageTransformed
	// path, keeping images upright without a global device flip.
	r.writeImagePixels(img, geom.Affine{
		A: dst.W() / float64(width),
		D: -dst.H() / float64(height),
		E: dst.Min.X,
		F: dst.Max.Y,
	})
}

// ImageTransformed draws a raster image through an arbitrary affine transform.
// The affine maps source image pixels into display coordinates.
func (r *Renderer) ImageTransformed(img render.Image, _ geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, geom.Rect{}, transform)
		}
		return
	}
	if !r.began || img == nil {
		return
	}
	r.writeImagePixels(img, transform)
}

func (r *Renderer) writeImagePixels(img render.Image, transform geom.Affine) {
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return
	}
	src := rgbaSource.RGBA()
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return
	}
	alphaMul := imageAlphaMultiplier(img)
	r.content.WriteString("\\pgfscope\n")
	writeTransform(&r.content, normalizedAffine(transform))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := src.RGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
			alpha := clamp01(float64(c.A) / 255.0 * alphaMul)
			if alpha <= 0 {
				continue
			}
			writeFillOpacity(&r.content, alpha)
			writeFillColor(&r.content, r.colorName(render.Color{
				R: float64(c.R) / 255.0,
				G: float64(c.G) / 255.0,
				B: float64(c.B) / 255.0,
				A: 1,
			}))
			fmt.Fprintf(&r.content, "\\pgfpathrectangle{\\pgfpoint{%spt}{%spt}}{\\pgfpoint{1pt}{1pt}}\n",
				shortFloat(float64(x)), shortFloat(float64(y)))
			r.content.WriteString("\\pgfusepath{fill}\n")
		}
	}
	r.content.WriteString("\\endpgfscope\n")
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}
