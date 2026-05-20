package mixedraster

import (
	"math"

	"github.com/cwbudde/matplotlib-go/backends/gobasic"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// Session captures one mixed raster/vector draw group on a transparent
// offscreen surface.
type Session struct {
	renderer *gobasic.Renderer
	rect     geom.Rect
}

// Start creates a transparent raster surface sized to the vector page.
func Start(width, height int, viewport geom.Rect, options render.Rasterization, fallbackDPI uint, clipRect *geom.Rect) (*Session, bool) {
	if width <= 0 || height <= 0 {
		return nil, false
	}
	dpi := fallbackDPI
	if options.DPI > 0 {
		dpi = uint(math.Round(options.DPI))
	}
	if dpi == 0 {
		dpi = 72
	}

	r := gobasic.New(width, height, render.Color{})
	r.SetResolution(dpi)
	if err := r.Begin(viewport); err != nil {
		return nil, false
	}
	if clipRect != nil {
		r.ClipRect(*clipRect)
	}
	return &Session{
		renderer: r,
		rect: geom.Rect{
			Min: geom.Pt{},
			Max: geom.Pt{X: float64(width), Y: float64(height)},
		},
	}, true
}

// Renderer returns the active offscreen renderer.
func (s *Session) Renderer() render.Renderer {
	if s == nil {
		return nil
	}
	return s.renderer
}

// Stop finishes the offscreen draw and returns an image suitable for embedding.
func (s *Session) Stop() (*render.ImageData, geom.Rect, bool) {
	if s == nil || s.renderer == nil {
		return nil, geom.Rect{}, false
	}
	_ = s.renderer.End()
	return render.NewImageData(s.renderer.GetImage()), s.rect, true
}

// ApplyAffine returns a copy of path transformed by affine.
func ApplyAffine(path geom.Path, affine geom.Affine) geom.Path {
	if len(path.C) == 0 {
		return geom.Path{}
	}
	out := geom.Path{
		C: append([]geom.Cmd(nil), path.C...),
		V: make([]geom.Pt, len(path.V)),
	}
	for i, pt := range path.V {
		out.V[i] = affine.Apply(pt)
	}
	return out
}
