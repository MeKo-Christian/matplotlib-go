package svg

import (
	"math"

	"github.com/cwbudde/matplotlib-go/geom"
	tex "github.com/cwbudde/matplotlib-go/internal/tex"
	"github.com/cwbudde/matplotlib-go/render"
)

func (r *Renderer) LastTeXError() error {
	if r == nil {
		return nil
	}
	return r.texErr
}

func (r *Renderer) MeasureTeX(text string, size float64, fontKey string) (render.TextMetrics, bool) {
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok {
		return render.TextMetrics{}, false
	}
	return result.Metrics, true
}

func (r *Renderer) DrawTeX(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.TeXDrawer); ok {
			return texRen.DrawTeX(text, origin, size, textColor, fontKey)
		}
		if textRen, ok := rr.(render.TextDrawer); ok {
			textRen.DrawText(text, origin, size, textColor)
			return true
		}
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - result.Metrics.Ascent}
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, "")
	return true
}

func (r *Renderer) DrawTeXRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) bool {
	if rr := r.activeRaster(); rr != nil {
		if texRen, ok := rr.(render.RotatedTeXDrawer); ok {
			return texRen.DrawTeXRotated(text, anchor, size, angle, textColor, fontKey)
		}
		if textRen, ok := rr.(render.RotatedTextDrawer); ok {
			textRen.DrawTextRotated(text, anchor, size, angle, textColor)
			return true
		}
		return false
	}
	if math.IsNaN(angle) || math.IsInf(angle, 0) {
		return false
	}
	result, ok := r.renderTeX(text, size, fontKey)
	if !ok || result.Image == nil {
		return false
	}
	img := colorizeTeXImage(result.Image, textColor)
	if img == nil {
		return false
	}

	metrics := result.Metrics
	origin := geom.Pt{X: anchor.X - metrics.W/2, Y: anchor.Y - metrics.Descent}
	topLeft := geom.Pt{X: origin.X, Y: origin.Y - metrics.Ascent}
	transform := rotateTransform(-angle*180/math.Pi, anchor.X, anchor.Y)
	r.renderImageNode(img, geom.Rect{
		Min: topLeft,
		Max: geom.Pt{X: topLeft.X + float64(img.Bounds().Dx()), Y: topLeft.Y + float64(img.Bounds().Dy())},
	}, transform)
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
