package patch_style_matrix

import (
	"image"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/internal/parityutil"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 720
	Height = 420
	DPI    = 100
)

func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.03, Y: 0.06}, Max: geom.Pt{X: 0.97, Y: 0.96}})
	ax.SetXLim(0, 12)
	ax.SetYLim(0, 8)
	hidePatchMatrixAxes(ax)

	addBoxStyleMatrix(ax)
	addHatchDensityMatrix(ax)
	addConnectionStyleMatrix(ax)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func hidePatchMatrixAxes(ax *core.Axes) {
	ax.ShowFrame = false
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
}

func addBoxStyleMatrix(ax *core.Axes) {
	styles := []core.BoxStyle{
		core.BoxStyleSquare,
		core.BoxStyleRound,
		core.BoxStyleRound4,
		core.BoxStyleSawtooth,
		core.BoxStyleRoundtooth,
		core.BoxStyleCircle,
		core.BoxStyleEllipse,
		core.BoxStyleLArrow,
		core.BoxStyleRArrow,
		core.BoxStyleDArrow,
	}
	colors := []render.Color{
		{R: 0.34, G: 0.66, B: 0.82, A: 0.82},
		{R: 0.90, G: 0.52, B: 0.28, A: 0.82},
		{R: 0.47, G: 0.72, B: 0.43, A: 0.82},
		{R: 0.78, G: 0.48, B: 0.73, A: 0.82},
		{R: 0.86, G: 0.70, B: 0.34, A: 0.82},
	}
	for i, style := range styles {
		col := i % 5
		row := i / 5
		ax.AddPatch(&core.FancyBboxPatch{
			Patch: core.Patch{
				FaceColor: colors[col],
				EdgeColor: render.Color{R: 0.13, G: 0.15, B: 0.18, A: 1},
				EdgeWidth: 1.0,
			},
			XY:             geom.Pt{X: 0.65 + float64(col)*2.25, Y: 6.55 - float64(row)*1.15},
			Width:          1.35,
			Height:         0.62,
			Pad:            0.10,
			BoxStyle:       style,
			RoundingSize:   0.16,
			ToothSize:      0.12,
			ArrowHeadWidth: 0.55,
			MutationSize:   1.0,
		})
	}
}

func addHatchDensityMatrix(ax *core.Axes) {
	hatches := []string{"/", "//", "o", "oo", ".", "..", "*", "**"}
	for i, hatch := range hatches {
		ax.AddPatch(&core.Rectangle{
			Patch: core.Patch{
				FaceColor:  render.Color{R: 0.92, G: 0.91, B: 0.84, A: 1},
				EdgeColor:  render.Color{R: 0.18, G: 0.22, B: 0.25, A: 1},
				EdgeWidth:  0.85,
				Hatch:      hatch,
				HatchColor: render.Color{R: 0.13, G: 0.20, B: 0.24, A: 1},
				HatchWidth: 0.75,
			},
			XY:     geom.Pt{X: 0.75 + float64(i)*1.38, Y: 3.58},
			Width:  0.9,
			Height: 0.78,
			Coords: core.Coords(core.CoordData),
		})
	}
}

func addConnectionStyleMatrix(ax *core.Axes) {
	arrowStyle, _ := core.ArrowStyleFromString("->,head_length=0.35,head_width=0.22")
	barStyle, _ := core.ArrowStyleFromString("|-|")
	wedgeStyle, _ := core.ArrowStyleFromString("wedge,tail_width=0.26,shrink_factor=0.35")
	arcStyle, _ := core.ConnectionStyleFromString("arc,armA=0.9,armB=0.65,rad=0.18")
	barConnAngle := 0.0
	barConn := core.ConnectionStyle{Name: "bar", Fraction: 0.25, Angle: &barConnAngle}
	arc3Style, _ := core.ConnectionStyleFromString("arc3,rad=0.22")

	arrows := []struct {
		a, b geom.Pt
		as   core.ArrowStyle
		cs   core.ConnectionStyle
		face render.Color
		edge render.Color
	}{
		{
			a:    geom.Pt{X: 0.9, Y: 2.25},
			b:    geom.Pt{X: 3.1, Y: 2.6},
			as:   arrowStyle,
			cs:   arcStyle,
			face: render.Color{R: 0.28, G: 0.48, B: 0.82, A: 1},
			edge: render.Color{R: 0.12, G: 0.24, B: 0.47, A: 1},
		},
		{
			a:    geom.Pt{X: 4.0, Y: 2.62},
			b:    geom.Pt{X: 6.25, Y: 1.88},
			as:   barStyle,
			cs:   barConn,
			face: render.Color{A: 0},
			edge: render.Color{R: 0.66, G: 0.28, B: 0.23, A: 1},
		},
		{
			a:    geom.Pt{X: 7.05, Y: 2.08},
			b:    geom.Pt{X: 10.8, Y: 2.58},
			as:   wedgeStyle,
			cs:   arc3Style,
			face: render.Color{R: 0.40, G: 0.66, B: 0.35, A: 0.82},
			edge: render.Color{R: 0.20, G: 0.38, B: 0.18, A: 1},
		},
	}
	for _, item := range arrows {
		ax.AddPatch(&core.FancyArrowPatch{
			Patch: core.Patch{
				FaceColor: item.face,
				EdgeColor: item.edge,
				EdgeWidth: 1.25,
			},
			PosA:            item.a,
			PosB:            item.b,
			ArrowStyle:      item.as,
			ConnectionStyle: item.cs,
			MutationScale:   15,
			Coords:          core.Coords(core.CoordData),
		})
	}
}
