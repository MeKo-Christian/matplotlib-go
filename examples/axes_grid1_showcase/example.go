// Package axes_grid1_showcase demonstrates an axes_grid1-style layout: a
// 2x2 ImageGrid plus three small RGB channel views and an anchored note.
package axes_grid1_showcase

import (
	"image"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
)

const (
	Width  = 1100
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)

	grid := fig.NewImageGrid(
		2,
		2,
		geom.Rect{
			Min: geom.Pt{X: 0.06, Y: 0.12},
			Max: geom.Pt{X: 0.60, Y: 0.88},
		},
		core.WithAxesDividerHorizontalSpace(0.18/11.0),
		core.WithAxesDividerVerticalSpace(0.20/7.2),
	)
	if grid == nil {
		return fig
	}

	for row := range 2 {
		for col := range 2 {
			ax := grid.At(row, col)
			ax.SetTitle("Tile " + string(rune('1'+row)) + "," + string(rune('1'+col)))
			ax.ImShow(surface(24, 24, float64(row*2+col)))
			labelPad := fontPad(9, 0.25, fig.RC.DPI)
			ax.Text(0.98, 0.02, "image grid", core.TextOptions{
				Coords:   core.Coords(core.CoordAxes),
				HAlign:   core.TextAlignRight,
				VAlign:   core.TextVAlignBottom,
				FontSize: 9,
				BBox: &core.TextBBoxOptions{
					FaceColor:    render.Color{R: 1, G: 1, B: 1, A: 1},
					EdgeColor:    render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
					Padding:      labelPad,
					CornerRadius: labelPad,
				},
			})
		}
	}

	channels := []struct {
		ax    *core.Axes
		title string
		cmap  string
	}{
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.66, Y: 0.34},
				Max: geom.Pt{X: 0.75, Y: 0.56},
			}),
			title: "Red",
			cmap:  "red channel",
		},
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.775, Y: 0.34},
				Max: geom.Pt{X: 0.865, Y: 0.56},
			}),
			title: "Green",
			cmap:  "green channel",
		},
		{
			ax: fig.AddAxes(geom.Rect{
				Min: geom.Pt{X: 0.89, Y: 0.34},
				Max: geom.Pt{X: 0.98, Y: 0.56},
			}),
			title: "Blue",
			cmap:  "blue channel",
		},
	}
	for idx, channel := range channels {
		channel.ax.SetTitle(channel.title)
		channel.ax.XAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 10, 20}}
		channel.ax.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 10, 20}}
		cmap := channel.cmap
		channel.ax.ImShow(channelSurface(28, 28, idx), core.ImShowOptions{
			Colormap: &cmap,
		})
	}

	notePad := 0.35 * 11 * fig.RC.DPI / 72
	noteInset := 0.5 * 11 * fig.RC.DPI / 72
	fig.AddAnchoredText("axes_grid1-style layout\nImageGrid + RGB channel views", core.AnchoredTextOptions{
		Location:        core.LegendUpperRight,
		Padding:         notePad,
		Inset:           noteInset,
		BoxPadding:      notePad,
		CornerRadius:    notePad,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		BorderColor:     render.Color{R: 0.75, G: 0.75, B: 0.75, A: 1},
		FontSize:        11,
	})

	return fig
}

// Render is the AGG-rendered showcase image.
func Render() image.Image {
	fig := Plot()
	r, err := agg.New(Width, Height, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		panic(err)
	}
	r.SetResolution(DPI)
	core.DrawFigure(fig, r)
	return r.Image()
}

func fontPad(fontSize, fraction, dpi float64) float64 {
	return fraction * fontSize * dpi / 72
}

func surface(rows, cols int, phase float64) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			values[y][x] = 0.5 + 0.25*math.Sin((xx+phase)*2*math.Pi) + 0.25*math.Cos((yy+phase*0.3)*3*math.Pi)
		}
	}
	return values
}

func channelSurface(rows, cols, channel int) [][]float64 {
	values := make([][]float64, rows)
	for y := range rows {
		values[y] = make([]float64, cols)
		yy := float64(y) / float64(rows-1)
		for x := range cols {
			xx := float64(x) / float64(cols-1)
			switch channel {
			case 1:
				values[y][x] = 0.5 + 0.32*math.Sin(yy*4*math.Pi) + 0.18*math.Cos(xx*2*math.Pi)
			case 2:
				dx := xx - 0.5
				dy := yy - 0.5
				values[y][x] = 0.58 + 0.36*math.Sin((xx+yy)*3*math.Pi) - 0.18*math.Hypot(dx, dy)
			default:
				dx := xx - 0.35
				dy := yy - 0.42
				values[y][x] = 0.35 + 0.65*math.Exp(-7*(dx*dx+dy*dy))
			}
		}
	}
	return values
}
