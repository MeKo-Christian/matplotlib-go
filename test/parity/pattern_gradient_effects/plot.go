package pattern_gradient_effects

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	ax.SetXLim(0, Width)
	ax.SetYLim(Height, 0)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.XAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.ShowFrame = false
	ax.Add(effectPaintArtist{})
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

type effectPaintArtist struct{}

func (effectPaintArtist) Draw(r render.Renderer, _ *core.DrawContext) {
	linear := rectPath(42, 42, 205, 142)
	r.Path(linear, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:  render.LinearGradient,
			Start: geom.Pt{X: 42, Y: 92},
			End:   geom.Pt{X: 205, Y: 92},
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 0.90, G: 0.16, B: 0.18, A: 1}},
				{Offset: 0.52, Color: render.Color{R: 0.96, G: 0.78, B: 0.20, A: 1}},
				{Offset: 1, Color: render.Color{R: 0.10, G: 0.30, B: 0.78, A: 1}},
			},
		},
		Stroke:    render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
		LineWidth: 1.8,
	})

	radial := rectPath(235, 42, 398, 142)
	r.Path(radial, &render.Paint{
		FillGradient: render.GradientFill{
			Kind:   render.RadialGradient,
			Center: geom.Pt{X: 316, Y: 92},
			Radius: 82,
			Stops: []render.GradientStop{
				{Offset: 0, Color: render.Color{R: 0.98, G: 0.98, B: 0.86, A: 1}},
				{Offset: 1, Color: render.Color{R: 0.08, G: 0.50, B: 0.36, A: 1}},
			},
		},
		Stroke:    render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
		LineWidth: 1.8,
	})

	tile := geom.Path{}
	tile.MoveTo(geom.Pt{X: 0, Y: 14})
	tile.LineTo(geom.Pt{X: 14, Y: 0})
	tile.MoveTo(geom.Pt{X: -4, Y: 6})
	tile.LineTo(geom.Pt{X: 6, Y: -4})
	pattern := geom.Path{}
	pattern.MoveTo(geom.Pt{X: 455, Y: 44})
	pattern.LineTo(geom.Pt{X: 594, Y: 54})
	pattern.LineTo(geom.Pt{X: 570, Y: 145})
	pattern.LineTo(geom.Pt{X: 430, Y: 128})
	pattern.Close()
	r.Path(pattern, &render.Paint{
		FillPattern: render.PatternFill{
			ID:         "phase2-diagonal",
			Cell:       geom.Rect{Max: geom.Pt{X: 14, Y: 14}},
			Path:       tile,
			Foreground: render.Color{R: 0.10, G: 0.20, B: 0.72, A: 1},
			Background: render.Color{R: 0.93, G: 0.94, B: 0.98, A: 1},
			LineWidth:  1.2,
		},
		Stroke:    render.Color{R: 0.06, G: 0.08, B: 0.10, A: 1},
		LineWidth: 1.8,
	})

	stroked := rectPath(86, 210, 228, 298)
	r.Path(stroked, &render.Paint{
		Fill:      render.Color{R: 0.12, G: 0.56, B: 0.40, A: 1},
		Stroke:    render.Color{R: 0.02, G: 0.09, B: 0.16, A: 1},
		LineWidth: 2.2,
		LineJoin:  render.JoinRound,
		PathEffects: []render.PathEffect{
			render.StrokePathEffect(render.Color{R: 1, G: 0.92, B: 0.58, A: 0.95}, 9, geom.Pt{X: 4, Y: -4}),
			render.NormalPathEffect(),
		},
	})

	filtered := rectPath(362, 210, 506, 298)
	r.Path(filtered, &render.Paint{
		Fill: render.Color{R: 0.10, G: 0.28, B: 0.74, A: 1},
		PathEffects: []render.PathEffect{
			render.FilterPathEffect(render.Color{R: 0.08, G: 0.12, B: 0.24, A: 0.45}, render.Color{}, 0, "blur", 5, geom.Pt{X: 8, Y: 8}),
			render.NormalPathEffect(),
		},
	})
}

func (effectPaintArtist) Z() float64 { return 1 }

func (effectPaintArtist) Bounds(*core.DrawContext) geom.Rect { return geom.Rect{} }

func rectPath(x0, y0, x1, y1 float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x0, Y: y0})
	path.LineTo(geom.Pt{X: x1, Y: y0})
	path.LineTo(geom.Pt{X: x1, Y: y1})
	path.LineTo(geom.Pt{X: x0, Y: y1})
	path.Close()
	return path
}
