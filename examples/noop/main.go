package main

import (
    "matplotlib-go/core"
    "matplotlib-go/internal/geom"
    "matplotlib-go/render"
)

func main() {
    fig := core.NewFigure(800, 600)
    ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
    ax.Add(core.ArtistFunc(func(r render.Renderer, ctx *core.DrawContext) {
        // no-op, just traversal
    }))
    var r render.NullRenderer
    core.DrawFigure(fig, &r)
}

