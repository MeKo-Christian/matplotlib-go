// Package demonstrates line join and cap styles with matplotlib-go.
package main

import (
	"log"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func main() {
	fig := core.NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	ax.SetXLim(0, 12)
	ax.SetYLim(0, 8)

	// Top row: same L-shaped path with different join styles.
	createJoinDemo(ax)

	// Bottom row: same horizontal path with different cap styles.
	createCapDemo(ax)

	r, _, createErr := backends.NewRendererFromEnv(backends.Config{
		Width:      800,
		Height:     600,
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        100,
	}, backends.TextCapabilities)
	if createErr != nil {
		log.Fatal(createErr)
	}

	err := core.SavePNG(fig, r, "examples/lines/styles.png")
	if err != nil {
		log.Fatalf("Failed to save PNG: %v", err)
	}

	log.Println("Saved line styles demo to examples/lines/styles.png")
}

func createJoinDemo(ax *core.Axes) {
	basePath := []geom.Pt{
		{X: 1, Y: 6}, {X: 3, Y: 6}, {X: 3, Y: 4},
	}

	for _, spec := range []struct {
		offset float64
		color  render.Color
		join   render.LineJoin
	}{
		{0, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, render.JoinMiter},
		{3, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1}, render.JoinRound},
		{6, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1}, render.JoinBevel},
	} {
		path := make([]geom.Pt, len(basePath))
		for i, pt := range basePath {
			path[i] = geom.Pt{X: pt.X + spec.offset, Y: pt.Y}
		}
		ax.Add(styledLine{
			XY:        path,
			LineWidth: 12,
			Color:     spec.color,
			LineCap:   render.CapButt,
			LineJoin:  spec.join,
		})
	}
}

func createCapDemo(ax *core.Axes) {
	for _, spec := range []struct {
		x0, x1 float64
		color  render.Color
		cap    render.LineCap
	}{
		{1, 3, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, render.CapButt},
		{4, 6, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1}, render.CapRound},
		{7, 9, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1}, render.CapSquare},
	} {
		ax.Add(styledLine{
			XY: []geom.Pt{
				{X: spec.x0, Y: 2},
				{X: spec.x1, Y: 2},
			},
			LineWidth: 12,
			Color:     spec.color,
			LineCap:   spec.cap,
			LineJoin:  render.JoinRound,
		})
	}
}

type styledLine struct {
	XY        []geom.Pt
	LineWidth float64
	Color     render.Color
	LineCap   render.LineCap
	LineJoin  render.LineJoin
}

func (l styledLine) Draw(r render.Renderer, ctx *core.DrawContext) {
	if len(l.XY) == 0 {
		return
	}

	path := geom.Path{}
	for i, pt := range l.XY {
		if i == 0 {
			path.C = append(path.C, geom.MoveTo)
		} else {
			path.C = append(path.C, geom.LineTo)
		}
		path.V = append(path.V, (&ctx.DataToPixel).Apply(pt))
	}

	r.Path(path, &render.Paint{
		LineWidth:  l.LineWidth,
		LineJoin:   l.LineJoin,
		LineCap:    l.LineCap,
		MiterLimit: 10,
		Stroke:     l.Color,
	})
}

func (l styledLine) Z() float64 { return 0 }

func (l styledLine) Bounds(*core.DrawContext) geom.Rect {
	if len(l.XY) == 0 {
		return geom.Rect{}
	}
	bounds := geom.Rect{Min: l.XY[0], Max: l.XY[0]}
	for _, pt := range l.XY[1:] {
		if pt.X < bounds.Min.X {
			bounds.Min.X = pt.X
		}
		if pt.Y < bounds.Min.Y {
			bounds.Min.Y = pt.Y
		}
		if pt.X > bounds.Max.X {
			bounds.Max.X = pt.X
		}
		if pt.Y > bounds.Max.Y {
			bounds.Max.Y = pt.Y
		}
	}
	return bounds
}
