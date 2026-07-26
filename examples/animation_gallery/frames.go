package animation_gallery

import (
	"image"
	"math"
	"time"

	"github.com/cwbudde/matplotlib-go/animation"
	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/color"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

// Deterministic-frame fixtures.
//
// Each animation kind here exposes a closed-form per-frame data generator so a
// single frame can be rendered as a static image and compared against the
// matplotlib reference for the same frame. The matplotlib reference modules in
// test/parity/<id>/plot.py reproduce the identical closed-form data, so the
// frame fixtures verify frame-N parity rather than playback timing.

const (
	// GoldenFrame is the representative frame captured by the deterministic
	// frame fixtures. The matplotlib reference plots draw the same frame.
	GoldenFrame = 8

	// FrameWidth/FrameHeight are the single-panel fixture dimensions and match
	// the matplotlib reference make_fig() default (640x360 at 100 DPI).
	FrameWidth  = 640
	FrameHeight = 360

	// SubplotWidth is wider to fit the two-panel composition fixture.
	SubplotWidth = 800

	lineSamples = 200
	scatterN    = 24
	imRows      = 36
	imCols      = 56
)

var (
	lineColorA  = color.Tab10[0]
	lineColorB  = color.Tab10[1]
	scatterEdge = render.Color{R: 0.12, G: 0.12, B: 0.14, A: 1}
)

// lineFrameXY returns the primary traveling-wave samples for a frame.
func lineFrameXY(frame int) []geom.Pt {
	pts := make([]geom.Pt, lineSamples)
	phase := float64(frame) * 0.30
	for i := range pts {
		x := 2 * math.Pi * float64(i) / float64(lineSamples-1)
		pts[i] = geom.Pt{X: x, Y: math.Sin(x + phase)}
	}
	return pts
}

// lineFrameXYB returns the secondary (cosine) traveling-wave samples.
func lineFrameXYB(frame int) []geom.Pt {
	pts := make([]geom.Pt, lineSamples)
	phase := float64(frame) * 0.30
	for i := range pts {
		x := 2 * math.Pi * float64(i) / float64(lineSamples-1)
		pts[i] = geom.Pt{X: x, Y: 0.6 * math.Cos(x+phase)}
	}
	return pts
}

// scatterFrameData returns orbiting-particle positions, scalar values, and
// marker areas (points^2) for a frame.
func scatterFrameData(frame int) (xy []geom.Pt, scalars, sizes []float64) {
	phase := float64(frame) * 0.20
	xy = make([]geom.Pt, scatterN)
	scalars = make([]float64, scatterN)
	sizes = make([]float64, scatterN)
	for i := 0; i < scatterN; i++ {
		base := 2 * math.Pi * float64(i) / float64(scatterN)
		r := 0.30 + 0.65*float64((i*7)%scatterN)/float64(scatterN)
		theta := base + phase
		xy[i] = geom.Pt{X: r * math.Cos(theta), Y: r * math.Sin(theta)}
		scalars[i] = r
		sizes[i] = 40 + 200*r
	}
	return xy, scalars, sizes
}

// imshowFrameZ returns the ripple heatmap matrix for a frame.
func imshowFrameZ(frame int) [][]float64 {
	cx := float64(imCols-1) / 2
	cy := float64(imRows-1) / 2
	t := float64(frame) * 0.40
	z := make([][]float64, imRows)
	for j := 0; j < imRows; j++ {
		row := make([]float64, imCols)
		for i := 0; i < imCols; i++ {
			d := math.Hypot(float64(i)-cx, float64(j)-cy)
			row[i] = math.Sin(d*0.5 - t)
		}
		z[j] = row
	}
	return z
}

// configureLineAxes applies the shared line-fixture axes scaffold.
func configureLineAxes(ax *core.Axes) {
	ax.SetTitle("Animated Line")
	ax.SetXLabel("phase")
	ax.SetYLabel("signal")
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.2, 1.2)
	ax.AddYGrid()
}

// configureScatterAxes applies the shared scatter-fixture axes scaffold.
func configureScatterAxes(ax *core.Axes) {
	ax.SetTitle("Animated Scatter")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetXLim(-1.1, 1.1)
	ax.SetYLim(-1.1, 1.1)
}

// addScatterFrame attaches the frame-N scatter collection to ax.
func addScatterFrame(ax *core.Axes, frame int) *core.Scatter2D {
	xy, scalars, sizes := scatterFrameData(frame)
	x := make([]float64, len(xy))
	y := make([]float64, len(xy))
	for i, p := range xy {
		x[i] = p.X
		y[i] = p.Y
	}
	cmap := "viridis"
	vmin, vmax := 0.30, 0.95
	edgeWidth := 1.2
	scatter, err := ax.Scatter(x, y, core.ScatterOptions{
		ScalarValues: scalars,
		Colormap:     cmap,
		VMin:         optional.Of(vmin),
		VMax:         optional.Of(vmax),
		Sizes:        sizes,
		EdgeColor:    optional.Of(scatterEdge),
		EdgeWidth:    optional.Of(edgeWidth),
	})
	if err != nil {
		return nil
	}
	return scatter
}

// addImshowFrame attaches the frame-N heatmap to ax.
func addImshowFrame(ax *core.Axes, frame int) *core.Image2D {
	cmap := "viridis"
	extent := [4]float64{0, float64(imCols), 0, float64(imRows)}
	img := ax.ImShow(imshowFrameZ(frame), core.ImShowOptions{
		Colormap: optional.Of(cmap),
		VMin:     optional.Of(-1.0),
		VMax:     optional.Of(1.0),
		Origin:   optional.Of(core.ImageOriginLower),
		Extent:   optional.Of(extent),
		Aspect:   optional.Of(core.AspectAuto),
	})
	ax.SetXLim(0, float64(imCols))
	ax.SetYLim(0, float64(imRows))
	return img
}

// configureImshowAxes applies the shared imshow-fixture title/labels.
func configureImshowAxes(ax *core.Axes) {
	ax.SetTitle("Animated Heatmap")
	ax.SetXLabel("column")
	ax.SetYLabel("row")
}

// buildLineFigure renders the line animation at the given frame.
func buildLineFigure(frame int) *core.Figure {
	fig := core.NewFigure(FrameWidth, FrameHeight)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureLineAxes(ax)
	ax.Add(&core.Line2D{XY: lineFrameXY(frame), W: 2.0, Col: lineColorA})
	ax.Add(&core.Line2D{XY: lineFrameXYB(frame), W: 2.0, Col: lineColorB})
	return fig
}

// buildScatterFigure renders the scatter animation at the given frame.
func buildScatterFigure(frame int) *core.Figure {
	fig := core.NewFigure(FrameWidth, FrameHeight)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureScatterAxes(ax)
	addScatterFrame(ax, frame)
	return fig
}

// buildImshowFigure renders the heatmap animation at the given frame.
func buildImshowFigure(frame int) *core.Figure {
	fig := core.NewFigure(FrameWidth, FrameHeight)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureImshowAxes(ax)
	addImshowFrame(ax, frame)
	return fig
}

// buildSubplotsFigure renders the two-panel (line + heatmap) composition.
func buildSubplotsFigure(frame int) *core.Figure {
	fig := core.NewFigure(SubplotWidth, FrameHeight)

	lineAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.15}, Max: geom.Pt{X: 0.45, Y: 0.88}})
	configureLineAxes(lineAx)
	lineAx.Add(&core.Line2D{XY: lineFrameXY(frame), W: 2.0, Col: lineColorA})
	lineAx.Add(&core.Line2D{XY: lineFrameXYB(frame), W: 2.0, Col: lineColorB})

	heatAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureImshowAxes(heatAx)
	addImshowFrame(heatAx, frame)

	return fig
}

// RenderLineFrame renders the animated-line fixture at GoldenFrame.
func RenderLineFrame() image.Image { return renderFrameFigure(buildLineFigure(GoldenFrame)) }

// RenderScatterFrame renders the animated-scatter fixture at GoldenFrame.
func RenderScatterFrame() image.Image { return renderFrameFigure(buildScatterFigure(GoldenFrame)) }

// RenderImshowFrame renders the animated-heatmap fixture at GoldenFrame.
func RenderImshowFrame() image.Image { return renderFrameFigure(buildImshowFigure(GoldenFrame)) }

// RenderSubplotsFrame renders the two-panel composition fixture at GoldenFrame.
func RenderSubplotsFrame() image.Image { return renderFrameFigure(buildSubplotsFigure(GoldenFrame)) }

// renderFrameFigure rasterizes a figure to an AGG image at its pixel size.
func renderFrameFigure(fig *core.Figure) image.Image {
	r, err := agg.New(int(fig.SizePx.X), int(fig.SizePx.Y), render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

// NewScatterAnimationDemo returns a deterministic orbiting-scatter animation.
func NewScatterAnimationDemo() (*Demo, error) {
	fig := core.NewFigure(FrameWidth, FrameHeight)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureScatterAxes(ax)
	scatter := addScatterFrame(ax, 0)
	cnv, err := newAnimationCanvas(fig)
	if err != nil {
		return nil, err
	}
	anim, err := animation.NewFuncAnimation(animation.Config{
		Canvas:   cnv,
		Frames:   16,
		Interval: 60 * time.Millisecond,
		Repeat:   true,
		Blit:     true,
	}, func(frame int) ([]core.Artist, error) {
		xy, _, _ := scatterFrameData(frame)
		scatter.XY = xy
		return []core.Artist{scatter}, nil
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Demo{Figure: fig, Canvas: cnv, Animation: anim, Artists: []core.Artist{scatter}}, nil
}

// NewImshowAnimationDemo returns a deterministic ripple-heatmap animation.
func NewImshowAnimationDemo() (*Demo, error) {
	fig := core.NewFigure(FrameWidth, FrameHeight)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureImshowAxes(ax)
	img := addImshowFrame(ax, 0)
	cnv, err := newAnimationCanvas(fig)
	if err != nil {
		return nil, err
	}
	anim, err := animation.NewFuncAnimation(animation.Config{
		Canvas:   cnv,
		Frames:   16,
		Interval: 60 * time.Millisecond,
		Repeat:   true,
		Blit:     false,
	}, func(frame int) ([]core.Artist, error) {
		img.Data = imshowFrameZ(frame)
		return []core.Artist{img}, nil
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Demo{Figure: fig, Canvas: cnv, Animation: anim, Artists: []core.Artist{img}}, nil
}

// NewSubplotAnimationDemo returns a deterministic two-panel composition
// animation that updates a line and a heatmap together.
func NewSubplotAnimationDemo() (*Demo, error) {
	fig := core.NewFigure(SubplotWidth, FrameHeight)

	lineAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.07, Y: 0.15}, Max: geom.Pt{X: 0.45, Y: 0.88}})
	configureLineAxes(lineAx)
	line := &core.Line2D{XY: lineFrameXY(0), W: 2.0, Col: lineColorA}
	lineAx.Add(line)

	heatAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.55, Y: 0.15}, Max: geom.Pt{X: 0.93, Y: 0.88}})
	configureImshowAxes(heatAx)
	img := addImshowFrame(heatAx, 0)

	cnv, err := newAnimationCanvas(fig)
	if err != nil {
		return nil, err
	}
	anim, err := animation.NewFuncAnimation(animation.Config{
		Canvas:   cnv,
		Frames:   16,
		Interval: 60 * time.Millisecond,
		Repeat:   true,
		Blit:     false,
	}, func(frame int) ([]core.Artist, error) {
		line.XY = lineFrameXY(frame)
		img.Data = imshowFrameZ(frame)
		return []core.Artist{line, img}, nil
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Demo{Figure: fig, Canvas: cnv, Animation: anim, Artists: []core.Artist{line, img}}, nil
}

func ptr[T any](v T) *T { return &v }
