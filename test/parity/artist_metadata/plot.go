package artist_metadata

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 540
	Height = 360
)

var axesRect = geom.Rect{
	Min: geom.Pt{X: 0.12, Y: 0.16},
	Max: geom.Pt{X: 0.88, Y: 0.86},
}

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(axesRect)
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	background := &core.PathPatch{
		Patch: core.Patch{
			FaceColor: render.Color{R: 0.16, G: 0.68, B: 0.43, A: 0.65},
			EdgeColor: render.Color{R: 0.02, G: 0.25, B: 0.14, A: 1},
			EdgeWidth: 1,
		},
		Path:   rectPath(0.08, 0.12, 0.84, 0.76),
		Coords: core.Coords(core.CoordData),
	}
	background.SetTransformCoords(core.Coords(core.CoordAxes))
	ax.Add(background)

	alphaLine := &core.Line2D{
		XY: []geom.Pt{
			{X: 0.6, Y: 1.4},
			{X: 9.4, Y: 8.6},
		},
		W:   8,
		Col: render.Color{R: 0.10, G: 0.26, B: 0.78, A: 1},
	}
	alphaLine.SetAlpha(0.45)
	ax.Add(alphaLine)

	hidden := &core.Line2D{
		XY: []geom.Pt{
			{X: 0.6, Y: 8.7},
			{X: 9.4, Y: 1.3},
		},
		W:   16,
		Col: render.Color{R: 1.0, G: 0.0, B: 0.75, A: 1},
	}
	hidden.SetVisible(false)
	ax.Add(hidden)

	clipped := &core.Line2D{
		XY: []geom.Pt{
			{X: 0.4, Y: 5.0},
			{X: 9.6, Y: 5.0},
		},
		W:   18,
		Col: render.Color{R: 0.86, G: 0.12, B: 0.10, A: 0.70},
	}
	clipped.SetClipRect(dataClipRect(2.2, 4.45, 7.8, 5.55))
	ax.Add(clipped)

	return fig
}

func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.GetImage()
}

func rectPath(x, y, w, h float64) geom.Path {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: x, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y})
	path.LineTo(geom.Pt{X: x + w, Y: y + h})
	path.LineTo(geom.Pt{X: x, Y: y + h})
	path.Close()
	return path
}

func dataClipRect(x0, y0, x1, y1 float64) geom.Rect {
	min := dataToDisplay(geom.Pt{X: x0, Y: y1})
	max := dataToDisplay(geom.Pt{X: x1, Y: y0})
	return geom.Rect{Min: min, Max: max}
}

func dataToDisplay(pt geom.Pt) geom.Pt {
	minX := float64(Width) * axesRect.Min.X
	maxX := float64(Width) * axesRect.Max.X
	minY := float64(Height) * (1 - axesRect.Max.Y)
	maxY := float64(Height) * (1 - axesRect.Min.Y)
	return geom.Pt{
		X: minX + (pt.X/10)*(maxX-minX),
		Y: maxY - (pt.Y/10)*(maxY-minY),
	}
}
