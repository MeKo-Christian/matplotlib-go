package svg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/imageinterp"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) DrawImage(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.DrawImage(img, dst)
		return
	}
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	// dst arrives in y-up display space. Flip to device space so row 0 lands at
	// the device-top edge and the image stays upright, mirroring the AGG
	// backend (devRect placement, no row reversal).
	flipped := geom.Rect{
		Min: geom.Pt{X: dst.Min.X, Y: r.flipY(dst.Max.Y)},
		Max: geom.Pt{X: dst.Max.X, Y: r.flipY(dst.Min.Y)},
	}
	r.renderImageNode(rgba, flipped, "", imageRenderingHint(img, flipped.W(), flipped.H()))
}

func (r *Renderer) ImageTransformed(img render.Image, dst geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, dst, transform)
		} else {
			rr.DrawImage(img, dst)
		}
		return
	}
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	// Compose the device y-flip so the placement maps image space -> y-down
	// device space, mirroring the AGG backend (deviceFlipAffine().Mul(affine)).
	// The transform scales dst, so measure the on-canvas size through it: the raw
	// rect would understate how far the raster is stretched and push every
	// transformed image towards a filtered draw.
	drawnW, drawnH := transformedExtent(dst, transform)

	r.renderImageNode(rgba, dst, matrixTransform(r.deviceFlip().Mul(transform)), imageRenderingHint(img, drawnW, drawnH))
}

// transformedExtent returns the axis-aligned width and height that dst covers
// once transform is applied.
func transformedExtent(dst geom.Rect, transform geom.Affine) (float64, float64) {
	corners := [4]geom.Pt{
		transform.Apply(geom.Pt{X: dst.Min.X, Y: dst.Min.Y}),
		transform.Apply(geom.Pt{X: dst.Max.X, Y: dst.Min.Y}),
		transform.Apply(geom.Pt{X: dst.Min.X, Y: dst.Max.Y}),
		transform.Apply(geom.Pt{X: dst.Max.X, Y: dst.Max.Y}),
	}

	minX, maxX := float64(corners[0].X), float64(corners[0].X)
	minY, maxY := float64(corners[0].Y), float64(corners[0].Y)

	for _, corner := range corners[1:] {
		minX = math.Min(minX, float64(corner.X))
		maxX = math.Max(maxX, float64(corner.X))
		minY = math.Min(minY, float64(corner.Y))
		maxY = math.Max(maxY, float64(corner.Y))
	}

	return maxX - minX, maxY - minY
}

// imageRenderingHint returns the value for the SVG image-rendering property, or
// "" to leave the viewer's default in place.
//
// Unlike Matplotlib's SVG backend, which resamples the raster to output
// resolution before embedding it, this backend embeds the source pixels and
// lets the viewer scale them. That is far smaller on disk, but it also hands
// the resampling decision to the viewer, which always smooths - so a 24x7
// heatmap blown up to ~1200x880 arrived as a blur instead of discrete cells.
// Carrying the resolved interpolation across as a rendering hint restores it.
func imageRenderingHint(img render.Image, dstW, dstH float64) string {
	if img == nil {
		return ""
	}

	srcW, srcH := img.Size()

	resolved := imageinterp.Resolve(
		img.Interpolation(),
		float64(srcW), float64(srcH),
		math.Abs(dstW), math.Abs(dstH),
	)
	if !imageinterp.IsNearest(resolved) {
		return ""
	}

	// SVG 1.1 only defines optimizeSpeed; pixelated and crisp-edges come from
	// CSS Images 3. Emit all three in cascade order so old and new viewers each
	// pick up the last one they understand.
	return "image-rendering:optimizeSpeed;image-rendering:crisp-edges;image-rendering:pixelated"
}

func (r *Renderer) renderImageNode(rgba *image.RGBA, dst geom.Rect, transform, style string) {
	x := dst.Min.X
	y := dst.Min.Y
	w := dst.W()
	h := dst.H()
	if w < 0 {
		x += w
		w = -w
	}
	if h < 0 {
		y += h
		h = -h
	}
	if w <= 0 || h <= 0 {
		return
	}

	encoded, err := encodeImage(rgba)
	if err != nil {
		return
	}

	uri := "data:image/png;base64," + encoded

	var b strings.Builder
	b.WriteString(`<image x="`)
	b.WriteString(formatFloat(x))
	b.WriteString(`" y="`)
	b.WriteString(formatFloat(y))
	b.WriteString(`" width="`)
	b.WriteString(formatFloat(w))
	b.WriteString(`" height="`)
	b.WriteString(formatFloat(h))
	b.WriteString(`" preserveAspectRatio="none"`)
	b.WriteString(` href="`)
	b.WriteString(uri)
	b.WriteString(`" xlink:href="`)
	b.WriteString(uri)
	b.WriteString(`"`)
	if transform != "" {
		writeAttr(&b, "transform", transform)
	}

	if style != "" {
		writeAttr(&b, "style", style)
	}

	b.WriteString(` />`)

	r.nodes = append(r.nodes, r.newNode(b.String()))
}

func encodeImage(img *image.RGBA) (string, error) {
	if img == nil {
		return "", errors.New("svg: image is nil")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func asRGBAImage(img render.Image) *image.RGBA {
	rgbaImage, ok := img.(interface {
		RGBA() *image.RGBA
	})
	if !ok {
		return nil
	}

	return rgbaImage.RGBA()
}

func colorizeTeXImage(src *image.RGBA, c render.Color) *image.RGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	r := uint8(clamp01(c.R)*255 + 0.5)
	g := uint8(clamp01(c.G)*255 + 0.5)
	b := uint8(clamp01(c.B)*255 + 0.5)
	alphaScale := clamp01(c.A)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a16 := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			a := uint8(float64(a16>>8)*alphaScale + 0.5)
			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}
