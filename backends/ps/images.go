package ps

import (
	"fmt"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// DrawImage draws an RGBA raster image into the destination rectangle using a
// Level-2 colorimage operator. PostScript has no native alpha channel, so
// translucent image pixels are pre-composited over white for this first slice.
func (r *Renderer) DrawImage(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.DrawImage(img, dst)
		return
	}
	if !r.began || img == nil || dst.W() <= 0 || dst.H() <= 0 {
		return
	}
	rgb, width, height, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: dst.W(),
		D: dst.H(),
		E: dst.Min.X,
		F: dst.Min.Y,
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
	width, height := img.Size()
	if width <= 0 || height <= 0 {
		return
	}
	rgb, _, _, ok := encodePSImageRGB(img)
	if !ok {
		return
	}
	r.writeImageWithMatrix(rgb, width, height, geom.Affine{
		A: transform.A * float64(width),
		B: transform.B * float64(width),
		C: transform.C * float64(height),
		D: transform.D * float64(height),
		E: transform.E,
		F: transform.F,
	})
}

func (r *Renderer) writeImageWithMatrix(rgb string, width, height int, matrix geom.Affine) {
	name := r.registerImageProcedure(rgb, width, height)
	if name == "" {
		return
	}
	fmt.Fprintf(
		&r.content, "gsave\n[%s %s %s %s %s %s] concat\n%s\ngrestore\n",
		shortFloat(matrix.A),
		shortFloat(matrix.B),
		shortFloat(matrix.C),
		shortFloat(matrix.D),
		shortFloat(matrix.E),
		shortFloat(matrix.F),
		name,
	)
}

func (r *Renderer) registerImageProcedure(rgb string, width, height int) string {
	if width <= 0 || height <= 0 || rgb == "" {
		return ""
	}
	key := fmt.Sprintf("%dx%d:%s", width, height, rgb)
	if name, ok := r.imageIDs[key]; ok {
		return name
	}
	name := fmt.Sprintf("I%d", len(r.imageIDs)+1)
	r.imageIDs[key] = name
	fmt.Fprintf(&r.content, "/%s {\n/DeviceRGB setcolorspace\n", name)
	fmt.Fprintf(&r.content, "%d %d 8 [%d 0 0 -%d 0 %d]\n", width, height, width, height, height)
	fmt.Fprintf(&r.content, "{<%s>} false 3 colorimage\n} bind def\n", rgb)
	return name
}

func imageAlphaMultiplier(img render.Image) float64 {
	if alphaImage, ok := img.(render.ImageAlpha); ok {
		return clamp01(alphaImage.Alpha())
	}
	return 1
}

func encodePSImageRGB(img render.Image) (string, int, int, bool) {
	rgbaSource, ok := img.(render.RGBAImage)
	if !ok || rgbaSource.RGBA() == nil {
		return "", 0, 0, false
	}
	src := rgbaSource.RGBA()
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return "", 0, 0, false
	}
	alphaMul := imageAlphaMultiplier(img)
	var b strings.Builder
	b.Grow(width * height * 6)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.RGBAAt(x, y)
			r, g, blue := compositeImagePixelOverWhite(c.R, c.G, c.B, c.A, alphaMul)
			fmt.Fprintf(&b, "%02x%02x%02x", r, g, blue)
		}
	}
	return b.String(), width, height, true
}

func compositeImagePixelOverWhite(red, green, blue, alpha uint8, alphaMul float64) (uint8, uint8, uint8) {
	a := clamp01(float64(alpha) / 255.0 * alphaMul)
	return uint8(float64(red)*a + 255*(1-a) + 0.5),
		uint8(float64(green)*a + 255*(1-a) + 0.5),
		uint8(float64(blue)*a + 255*(1-a) + 0.5)
}
