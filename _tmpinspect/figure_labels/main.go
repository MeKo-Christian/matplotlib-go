package main

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	example "github.com/cwbudde/matplotlib-go/examples/figure_labels_composition"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

type recordingRenderer struct {
	*agg.Renderer
}

func (r *recordingRenderer) DrawText(text string, origin geom.Pt, size float64, textColor render.Color) {
	if text == "time [s]" || text == "Shared-Axis Figure Labels" || text == "upper-left" || text == "note" {
		fmt.Printf("text %q origin=%+v size=%.3f\n", text, origin, size)
	}
	r.Renderer.DrawText(text, origin, size, textColor)
}

func (r *recordingRenderer) DrawTextWithFont(text string, origin geom.Pt, size float64, textColor render.Color, fontKey string) {
	if text == "time [s]" || text == "Shared-Axis Figure Labels" || text == "upper-left" || text == "note" {
		fmt.Printf("font text %q origin=%+v size=%.3f font=%q\n", text, origin, size, fontKey)
	}
	r.Renderer.DrawTextWithFont(text, origin, size, textColor, fontKey)
}

func (r *recordingRenderer) DrawTextRotated(text string, anchor geom.Pt, size, angle float64, textColor render.Color) {
	if text == "amplitude" {
		fmt.Printf("rotated text %q anchor=%+v size=%.3f angle=%.3f\n", text, anchor, size, angle)
	}
	r.Renderer.DrawTextRotated(text, anchor, size, angle, textColor)
}

func (r *recordingRenderer) DrawTextRotatedWithFont(text string, anchor geom.Pt, size, angle float64, textColor render.Color, fontKey string) {
	if text == "amplitude" {
		fmt.Printf("font rotated text %q anchor=%+v size=%.3f angle=%.3f font=%q\n", text, anchor, size, angle, fontKey)
	}
	r.Renderer.DrawTextRotatedWithFont(text, anchor, size, angle, textColor, fontKey)
}

func main() {
	fig := example.Plot()
	r, err := agg.New(example.Width, example.Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	for _, sample := range []string{"lp", "upper-left", "note"} {
		fmt.Printf("metrics %q = %+v\n", sample, r.MeasureText(sample, 10, "DejaVu Sans"))
		if b, ok := r.MeasureTextBounds(sample, 10, "DejaVu Sans"); ok {
			fmt.Printf("bounds %q = %+v\n", sample, b)
		}
		if h, ok := r.MeasureFontHeights(10, "DejaVu Sans"); ok {
			fmt.Printf("font heights = %+v\n", h)
		}
	}
	core.DrawFigure(fig, &recordingRenderer{Renderer: r})
	for i, ax := range fig.Children {
		fmt.Printf("axes %d frac=%+v px=%+v\n", i+1, ax.RectFraction, ax.DisplayRect())
	}
}
