package main

import (
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Add(core.ArtistFunc(func(r render.Renderer, ctx *core.DrawContext) {
		// Exercise artist traversal without drawing; this mirrors the Python note-only example.
	}))
	var r render.NullRenderer
	core.DrawFigure(fig, &r)
}
