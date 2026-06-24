package mathtext_accents

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
)

const (
	Width  = 640
	Height = 360
)

// labels are MathText accent expressions supported by Matplotlib's mathtext, so
// the rendered output can be compared against the matplotlib reference. The
// best-effort, non-parity extensions (\overbrace, \underbrace, \stackrel, \not)
// are intentionally excluded here.
var labels = []struct {
	x, y float64
	expr string
}{
	{0.20, 0.85, `$\hat{x}$`},
	{0.50, 0.85, `$\bar{x}$`},
	{0.80, 0.85, `$\vec{v}$`},
	{0.20, 0.62, `$\dot{x}$`},
	{0.50, 0.62, `$\ddot{y}$`},
	{0.80, 0.62, `$\tilde{n}$`},
	{0.20, 0.39, `$\widehat{AB}$`},
	{0.50, 0.39, `$\widetilde{xy}$`},
	{0.80, 0.39, `$\overline{x+y}$`},
	{0.20, 0.16, `$\overset{a}{X}$`},
	{0.50, 0.16, `$\underset{b}{Y}$`},
	{0.80, 0.16, `$X_{\substack{i \\ j}}$`},
}

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.05, Y: 0.08},
		Max: geom.Pt{X: 0.95, Y: 0.92},
	})
	ax.SetXLim(0, 1)
	ax.SetYLim(0, 1)
	hideAxes(ax)

	for _, l := range labels {
		ax.Text(l.x, l.y, l.expr, centerText(26))
	}
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func centerText(size float64) core.TextOptions {
	return core.TextOptions{
		Coords:   core.Coords(core.CoordAxes),
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: size,
	}
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
