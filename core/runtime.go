package core

import "github.com/cwbudde/matplotlib-go/internal/geom"

// DisplayRect reports the figure display rectangle in pixels.
func (f *Figure) DisplayRect() geom.Rect {
	if f == nil {
		return geom.Rect{}
	}
	return geom.Rect{
		Min: geom.Pt{X: 0, Y: 0},
		Max: geom.Pt{X: f.SizePx.X, Y: f.SizePx.Y},
	}
}

// DisplayRect reports the axes display rectangle in pixels.
func (a *Axes) DisplayRect() geom.Rect {
	if a == nil || a.figure == nil {
		return geom.Rect{}
	}
	return a.adjustedLayout(a.figure)
}

// ContainsDisplayPoint reports whether a figure-pixel point lies inside the axes.
func (a *Axes) ContainsDisplayPoint(p geom.Pt) bool {
	if a == nil {
		return false
	}
	rect := a.DisplayRect()
	if !rect.Contains(p) {
		return false
	}
	return projectionContainsDisplayPoint(a.projection, rect, p)
}

// PixelToData resolves a figure-pixel point into this axes' data coordinates.
func (a *Axes) PixelToData(p geom.Pt) (geom.Pt, bool) {
	if a == nil || a.figure == nil {
		return geom.Pt{}, false
	}
	figureRect := a.figure.DisplayRect()
	clip := a.adjustedLayout(a.figure)
	ctx := newAxesDrawContext(a, a.figure, figureRect, clip)
	return ctx.DataToPixel.Invert(p)
}

// AxesDrawContext builds the draw context an axes uses for rendering — exposed
// for the canvas hit-testing and coordinate-inspection helpers.
func AxesDrawContext(a *Axes, fig *Figure) *DrawContext {
	if a == nil || fig == nil {
		return nil
	}
	figureRect := fig.DisplayRect()
	clip := a.adjustedLayout(fig)
	return newAxesDrawContext(a, fig, figureRect, clip)
}

// FigureDrawContext builds the draw context figure-level artists use for
// rendering. Useful when picking or inspecting non-axes children.
func FigureDrawContext(fig *Figure) *DrawContext {
	if fig == nil {
		return nil
	}
	return newFigureDrawContext(fig, fig.DisplayRect())
}
