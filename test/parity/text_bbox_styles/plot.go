package text_bbox_styles

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 640
	Height = 360
)

// faceColor and edgeColor are shared by every box so the parity comparison
// isolates the boxstyle geometry rather than color handling.
var (
	faceColor = render.Color{R: 0.85, G: 0.92, B: 1.0, A: 1}
	edgeColor = render.Color{R: 0.20, G: 0.30, B: 0.60, A: 1}
)

// boxes exercises each FancyBboxPatch style reachable through a Text bbox, using
// the Matplotlib boxstyle spec strings the bridge now understands. The layout is
// a 2×5 grid; labels are the style names so the text content stays trivial.
var boxes = []struct {
	x, y  float64
	label string
	style string
}{
	{0.27, 0.86, "square", "square,pad=0.4"},
	{0.73, 0.86, "circle", "circle,pad=0.4"},
	{0.27, 0.68, "round", "round,pad=0.4"},
	{0.73, 0.68, "round4", "round4,pad=0.4"},
	{0.27, 0.50, "ellipse", "ellipse,pad=0.4"},
	{0.73, 0.50, "sawtooth", "sawtooth,pad=0.4"},
	{0.27, 0.32, "roundtooth", "roundtooth,pad=0.4"},
	{0.73, 0.32, "rarrow", "rarrow,pad=0.4"},
	{0.27, 0.14, "larrow", "larrow,pad=0.4"},
	{0.73, 0.14, "darrow", "darrow,pad=0.4"},
}

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.05},
		Max: geom.Pt{X: 0.95, Y: 0.95},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	hideAxes(ax)

	for _, b := range boxes {
		ax.Text(b.x, b.y, b.label, core.TextOptions{
			Coords:   core.Coords(core.CoordAxes),
			HAlign:   core.TextAlignCenter,
			VAlign:   core.TextVAlignMiddle,
			FontSize: 16,
			BBox: &core.TextBBoxOptions{
				Style:     b.style,
				FaceColor: faceColor,
				EdgeColor: edgeColor,
			},
		})
	}
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func hideAxes(ax *core.Axes) {
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false
}
