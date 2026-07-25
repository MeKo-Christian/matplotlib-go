// Package animation_gallery demonstrates deterministic animation setup.
package animation_gallery

import (
	"image"
	"math"
	"time"

	"github.com/cwbudde/matplotlib-go/animation"
	"github.com/cwbudde/matplotlib-go/backends"
	"github.com/cwbudde/matplotlib-go/backends/agg"
	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/canvas"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1000
	Height = 560
	DPI    = 100
)

var (
	funcColor   = render.Color{R: 0.12, G: 0.34, B: 0.68, A: 1}
	artistA     = render.Color{R: 0.84, G: 0.35, B: 0.18, A: 1}
	artistB     = render.Color{R: 0.20, G: 0.58, B: 0.34, A: 1}
	artistC     = render.Color{R: 0.55, G: 0.32, B: 0.72, A: 1}
	previewGray = render.Color{R: 0.28, G: 0.28, B: 0.28, A: 0.45}
)

// Demo carries a figure, canvas, and configured animation for examples/tests.
type Demo struct {
	Figure    *core.Figure
	Canvas    canvas.FigureCanvas
	Animation *animation.Animation
	Artists   []core.Artist
}

// Plot builds a static preview of the animation examples.
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	funcAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.16}, Max: geom.Pt{X: 0.48, Y: 0.90}})
	funcAx.SetTitle("FuncAnimation frames")
	funcAx.SetXLabel("phase")
	funcAx.SetYLabel("signal")
	funcAx.SetXLim(0, 2*math.Pi)
	funcAx.SetYLim(-1.35, 1.35)
	funcAx.AddYGrid()
	for _, frame := range []int{0, 4, 8} {
		color := previewGray
		width := 1.2
		if frame == 8 {
			color = funcColor
			width = 2.3
		}
		funcAx.Add(&core.Line2D{XY: sineFrame(frame), W: width, Col: color})
	}

	artistAx := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.58, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	artistAx.SetTitle("ArtistAnimation frames")
	artistAx.SetXLabel("frame")
	artistAx.SetYLabel("value")
	artistAx.SetXLim(-0.5, 3.5)
	artistAx.SetYLim(0, 3.5)
	artistAx.AddYGrid()
	for i, line := range artistPreviewLines() {
		line.Label = []string{"frame 0", "frame 1", "frame 2"}[i]
		artistAx.Add(line)
	}
	artistAx.AddLegend()

	return fig
}

// NewFuncAnimationDemo returns a deterministic line-update animation.
func NewFuncAnimationDemo() (*Demo, error) {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	ax.SetTitle("FuncAnimation")
	ax.SetXLim(0, 2*math.Pi)
	ax.SetYLim(-1.35, 1.35)
	ax.AddYGrid()

	line := &core.Line2D{XY: sineFrame(0), W: 2.2, Col: funcColor}
	ax.Add(line)
	cnv, err := newAnimationCanvas(fig)
	if err != nil {
		return nil, err
	}
	anim, err := animation.NewFuncAnimation(animation.Config{
		Canvas:      cnv,
		Frames:      12,
		Interval:    50 * time.Millisecond,
		Repeat:      true,
		RepeatDelay: 200 * time.Millisecond,
		Blit:        true,
	}, func(frame int) ([]core.Artist, error) {
		line.XY = sineFrame(frame)
		return []core.Artist{line}, nil
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Demo{Figure: fig, Canvas: cnv, Animation: anim, Artists: []core.Artist{line}}, nil
}

// NewArtistAnimationDemo returns a deterministic artist-list animation.
func NewArtistAnimationDemo() (*Demo, error) {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.10, Y: 0.16}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	ax.SetTitle("ArtistAnimation")
	ax.SetXLim(-0.5, 3.5)
	ax.SetYLim(0, 3.5)
	ax.AddYGrid()
	lines := artistPreviewLines()
	artists := make([]core.Artist, 0, len(lines))
	for _, line := range lines {
		ax.Add(line)
		artists = append(artists, line)
	}
	cnv, err := newAnimationCanvas(fig)
	if err != nil {
		return nil, err
	}
	anim, err := animation.NewArtistAnimation(animation.Config{
		Canvas:   cnv,
		Interval: 80 * time.Millisecond,
		Repeat:   true,
		Blit:     true,
	}, [][]core.Artist{
		{lines[0]},
		{lines[1]},
		{lines[2]},
	})
	if err != nil {
		return nil, err
	}
	return &Demo{Figure: fig, Canvas: cnv, Animation: anim, Artists: artists}, nil
}

// Render is the AGG-rendered static preview image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func newAnimationCanvas(fig *core.Figure) (canvas.FigureCanvas, error) {
	cnv, _, err := backends.NewCanvas("agg", backends.Config{
		Width:      int(fig.SizePx.X),
		Height:     int(fig.SizePx.Y),
		Background: render.Color{R: 1, G: 1, B: 1, A: 1},
		DPI:        DPI,
	}, fig, nil)
	return cnv, err
}

func sineFrame(frame int) []geom.Pt {
	points := make([]geom.Pt, 96)
	phase := float64(frame) * 0.35
	for i := range points {
		x := 2 * math.Pi * float64(i) / float64(len(points)-1)
		points[i] = geom.Pt{X: x, Y: math.Sin(x+phase) * (0.75 + 0.02*float64(frame))}
	}
	return points
}

func artistPreviewLines() []*core.Line2D {
	width := 2.1
	return []*core.Line2D{
		{XY: []geom.Pt{{X: 0, Y: 0.7}, {X: 1, Y: 1.4}, {X: 2, Y: 0.9}, {X: 3, Y: 1.8}}, W: width, Col: artistA},
		{XY: []geom.Pt{{X: 0, Y: 1.3}, {X: 1, Y: 2.2}, {X: 2, Y: 1.6}, {X: 3, Y: 2.7}}, W: width, Col: artistB},
		{XY: []geom.Pt{{X: 0, Y: 2.0}, {X: 1, Y: 2.8}, {X: 2, Y: 2.2}, {X: 3, Y: 3.1}}, W: width, Col: artistC},
	}
}
