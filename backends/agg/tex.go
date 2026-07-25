package agg

import (
	"image"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

// LastTeXError returns the most recent TeX pipeline error recorded by MeasureTeX
// or DrawTeX. A nil value means the last TeX operation succeeded.
func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

// MeasureTeX measures a TeX string by rendering it through the external
// latex+dvipng cache and using the resulting tight PNG dimensions.
func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return result.Metrics, true
}

// DrawTeX draws a TeX string through the external latex+dvipng cache.
func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	// Display space is y-up: the baseline sits at origin.Y, ascent extends up
	// (+Y) and descent down (-Y). The rect's bottom edge (Min.Y) is therefore
	// origin.Y - Descent; drawTeXImage extends up by the image height to the top.
	r.drawTeXImage(result, geom.Pt{X: origin.X, Y: origin.Y - result.Metrics.Descent}, textColor)
	return true
}

// DrawTeXRotated draws a TeX string rotated around the Matplotlib-style text
// rotation anchor.
func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}

	metrics := result.Metrics
	bounds := render.TextBounds{X: 0, Y: -metrics.Ascent, W: metrics.W, H: metrics.H}
	origin := rotatedTextOrigin(anchor, metrics, bounds, true)
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	r.drawTeXImageRotated(result, topLeft, anchor, angle, textColor)
	return true
}

func (r *Renderer) renderTeX(text string, size float64, fontKey string) (tex.RenderResult, bool) {
	if r == nil || text == "" || size <= 0 {
		return tex.RenderResult{}, false
	}
	if r.texManager == nil {
		r.texManager = tex.NewManager(tex.ManagerConfig{})
	}
	result, err := r.texManager.Render(text, size, r.resolution, fontKey)
	if err != nil {
		r.texErr = err
		return tex.RenderResult{}, false
	}
	r.texErr = nil
	return result, true
}

func (r *Renderer) drawTeXImage(result tex.RenderResult, topLeft geom.Pt, textColor render.Color) {
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return
	}
	r.DrawImage(render.NewImageData(img), geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	})
}

func (r *Renderer) drawTeXImageRotated(result tex.RenderResult, topLeft, anchor geom.Pt, angle float64, textColor render.Color) {
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return
	}
	cos := math.Cos(-angle)
	sin := math.Sin(-angle)
	affine := translateAffineAgg(anchor).
		Mul(geom.Affine{A: cos, B: sin, C: -sin, D: cos}).
		Mul(translateAffineAgg(geom.Pt{X: -anchor.X, Y: -anchor.Y})).
		Mul(translateAffineAgg(topLeft))
	r.ImageTransformed(render.NewImageData(img), geom.Rect{}, affine)
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

func translateAffineAgg(p geom.Pt) geom.Affine {
	return geom.Affine{A: 1, D: 1, E: p.X, F: p.Y}
}
