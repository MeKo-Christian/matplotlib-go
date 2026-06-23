package svg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) Image(img render.Image, dst geom.Rect) {
	if rr := r.activeRaster(); rr != nil {
		rr.Image(img, dst)
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
	r.renderImageNode(rgba, flipped, "")
}

func (r *Renderer) ImageTransformed(img render.Image, dst geom.Rect, transform geom.Affine) {
	if rr := r.activeRaster(); rr != nil {
		if tr, ok := rr.(render.ImageTransformer); ok {
			tr.ImageTransformed(img, dst, transform)
		} else {
			rr.Image(img, dst)
		}
		return
	}
	rgba := asRGBAImage(img)
	if rgba == nil {
		return
	}
	// Compose the device y-flip so the placement maps image space -> y-down
	// device space, mirroring the AGG backend (deviceFlipAffine().Mul(affine)).
	r.renderImageNode(rgba, dst, matrixTransform(r.deviceFlip().Mul(transform)))
}

func (r *Renderer) renderImageNode(rgba *image.RGBA, dst geom.Rect, transform string) {
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
