// Package text_layout_gallery is a user-facing gallery for text layout.
package text_layout_gallery

import (
	"image"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 900
	Height = 560
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	ax.SetTitle("Text Layout Gallery")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 4)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	addAlignmentSamples(ax)
	addRotatedSamples(ax)
	addMultilineSamples(ax)
	addWrappedSample(ax)
	fig.Text(0.05, 0.94, "Alignment, rotation, multiline, wrapping, and bbox text", core.TextOptions{FontSize: 11})
	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func addAlignmentSamples(ax *core.Axes) {
	crossColor := render.Color{R: 0.70, G: 0.70, B: 0.74, A: 1}
	for _, x := range []float64{0.9, 1.6, 2.3} {
		_, _ = ax.Plot([]float64{x, x}, []float64{2.6, 3.5}, core.PlotOptions{Color: optional.Of(crossColor)})
	}
	for _, y := range []float64{2.75, 3.05, 3.35} {
		_, _ = ax.Plot([]float64{0.5, 2.7}, []float64{y, y}, core.PlotOptions{Color: optional.Of(crossColor)})
	}
	ax.Text(0.9, 3.35, "left/top", core.TextOptions{HAlign: core.TextAlignLeft, VAlign: core.TextVAlignTop, FontSize: 10})
	ax.Text(1.6, 3.05, "center", core.TextOptions{HAlign: core.TextAlignCenter, VAlign: core.TextVAlignMiddle, FontSize: 10})
	ax.Text(2.3, 2.75, "right/bottom", core.TextOptions{HAlign: core.TextAlignRight, VAlign: core.TextVAlignBottom, FontSize: 10})
}

func addRotatedSamples(ax *core.Axes) {
	box := textBox(12, 0.25, render.Color{R: 1.00, G: 0.92, B: 0.78, A: 0.78}, render.Color{R: 0.68, G: 0.38, B: 0.12, A: 1})
	ax.Text(3.45, 3.35, "rotation\nmode", core.TextOptions{
		HAlign: core.TextAlignCenter,
		VAlign: core.TextVAlignMiddle,
		Angle:  -32,
		BBox:   optional.Of(box),
	})
	ax.Text(4.50, 3.18, "anchor", core.TextOptions{
		HAlign:       core.TextAlignCenter,
		VAlign:       core.TextVAlignMiddle,
		Angle:        34,
		RotationMode: core.TextRotationModeAnchor,
		FontSize:     12,
		BBox:         optional.Of(textBox(12, 0.22, render.Color{R: 0.88, G: 0.94, B: 1.00, A: 0.78}, render.Color{R: 0.20, G: 0.38, B: 0.68, A: 1})),
	})
}

func addMultilineSamples(ax *core.Axes) {
	left := core.TextAlignLeft
	right := core.TextAlignRight
	ax.Text(0.75, 1.75, "multi-line\nleft aligned\ntext", core.TextOptions{
		HAlign:         core.TextAlignLeft,
		VAlign:         core.TextVAlignTop,
		MultiAlignment: optional.Of(left),
		Linespacing:    1.3,
		FontSize:       11,
		BBox:           optional.Of(textBox(11, 0.28, render.Color{R: 0.94, G: 0.97, B: 0.92, A: 0.86}, render.Color{R: 0.24, G: 0.50, B: 0.20, A: 1})),
	})
	ax.Text(2.55, 1.65, "right\naligned\nblock", core.TextOptions{
		HAlign:         core.TextAlignRight,
		VAlign:         core.TextVAlignTop,
		MultiAlignment: optional.Of(right),
		Linespacing:    1.15,
		FontSize:       11,
		BBox:           optional.Of(textBox(11, 0.28, render.Color{R: 0.98, G: 0.94, B: 1.00, A: 0.86}, render.Color{R: 0.45, G: 0.26, B: 0.58, A: 1})),
	})
}

func addWrappedSample(ax *core.Axes) {
	ax.Text(3.45, 1.78, "wrapped text uses a fixed display width inside a rounded bbox", core.TextOptions{
		HAlign:   core.TextAlignLeft,
		VAlign:   core.TextVAlignTop,
		Wrap:     true,
		FontSize: 11,
		BBox:     optional.Of(textBox(11, 0.28, render.Color{R: 1.00, G: 0.98, B: 0.88, A: 0.88}, render.Color{R: 0.52, G: 0.45, B: 0.16, A: 1})),
	})
}

func textBox(fontSize, pad float64, face, edge render.Color) core.TextBBoxOptions {
	return core.TextBBoxOptions{
		FaceColor:    face,
		EdgeColor:    edge,
		LineWidth:    0.9,
		Padding:      fontSize * pad * DPI / 72,
		CornerRadius: 5,
	}
}
