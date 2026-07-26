package clip_path_batch

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

func Render() image.Image {
	fig := core.NewFigure(980, 620)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.14}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("RendererAgg clip path batch")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5.4)
	ax.XAxis.Locator = ticker.MultipleLocator{Base: 1}
	ax.YAxis.Locator = ticker.MultipleLocator{Base: 1}
	ax.AddYGrid()

	xEdges := []float64{0, 0.75, 1.5, 2.35, 3.1, 4.0, 4.85, 5.45, 6.0}
	yEdges := []float64{0, 0.7, 1.55, 2.4, 3.2, 4.15, 5.4}
	data := make([][]float64, len(yEdges)-1)
	for yi := range data {
		data[yi] = make([]float64, len(xEdges)-1)
		for xi := range data[yi] {
			cx := (xEdges[xi] + xEdges[xi+1]) * 0.5
			cy := (yEdges[yi] + yEdges[yi+1]) * 0.5
			data[yi][xi] = 0.45 + 0.42*math.Sin(cx*1.15) + 0.33*math.Cos(cy*1.35) + 0.06*float64((xi+yi)%3)
		}
	}

	cmap := "viridis"
	vmin, vmax := -0.35, 1.15
	alpha := 0.84
	edgeColor := render.Color{R: 0.97, G: 0.97, B: 0.97, A: 0.72}
	edgeWidth := 0.55
	antialias := true
	clipPath := clipBatchPath()
	mesh := ax.PColorMesh(data, core.MeshOptions{
		XEdges:    xEdges,
		YEdges:    yEdges,
		Shading:   core.MeshShadingFlat,
		Colormap:  optional.Of(cmap),
		VMin:      optional.Of(vmin),
		VMax:      optional.Of(vmax),
		Alpha:     optional.Of(alpha),
		EdgeColor: optional.Of(edgeColor),
		EdgeWidth: optional.Of(edgeWidth),
		Antialias: optional.Of(antialias),
	})
	mesh.SetClipPathCoords(clipPath, core.Coords(core.CoordData))

	outline := &core.PathPatch{
		Patch: core.Patch{
			EdgeColor: optional.Of(render.Color{R: 0.05, G: 0.08, B: 0.12, A: 1}),
			EdgeWidth: optional.Of(2.0),
			LineJoin:  render.JoinMiter,
		},
		Path:   clipPath,
		Coords: core.Coords(core.CoordData),
	}
	outline.SetFaceColor(render.Color{})
	ax.AddPatch(outline)

	return common.RenderFixtureFigure(fig, 980, 620)
}

func clipBatchPath() geom.Path {
	points := []geom.Pt{
		{X: 0.55, Y: 1.10},
		{X: 2.05, Y: 0.50},
		{X: 3.10, Y: 1.05},
		{X: 5.35, Y: 0.80},
		{X: 4.70, Y: 2.45},
		{X: 5.50, Y: 4.05},
		{X: 3.70, Y: 3.80},
		{X: 2.55, Y: 5.05},
		{X: 1.75, Y: 3.55},
		{X: 0.55, Y: 3.85},
		{X: 1.20, Y: 2.35},
	}
	path := geom.Path{}
	for i, pt := range points {
		if i == 0 {
			path.MoveTo(pt)
		} else {
			path.LineTo(pt)
		}
	}
	path.Close()
	return path
}
