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
	ax.ShowFrame = false
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	styles := []struct {
		style        core.BoxStyle
		pad          float64
		roundingSize float64
		toothSize    float64
	}{
		{style: core.BoxStyleSquare, pad: 0.10},
		{style: core.BoxStyleRound, pad: 0.10, roundingSize: 0.16},
		{style: core.BoxStyleRound4, pad: 0.10, roundingSize: 0.16},
		{style: core.BoxStyleSawtooth, pad: 0.10, toothSize: 0.12},
		{style: core.BoxStyleRoundtooth, pad: 0.10, toothSize: 0.12},
		{style: core.BoxStyleCircle, pad: 0.10},
		{style: core.BoxStyleEllipse, pad: 0.10},
		{style: core.BoxStyleLArrow, pad: 0.10},
		{style: core.BoxStyleRArrow, pad: 0.10},
		{style: core.BoxStyleDArrow, pad: 0.10},
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
			XY:           geom.Pt{X: 0.65 + float64(col)*2.25, Y: 6.55 - float64(row)*1.15},
			Width:        1.35,
			Height:       0.62,
			Pad:          style.pad,
			BoxStyle:     style.style,
			RoundingSize: style.roundingSize,
			ToothSize:    style.toothSize,
			MutationSize: 1.0,
		})
	}

	for i, hatch := range []string{"/", "//", "o", "oo", ".", "..", "*", "**"} {
		ax.AddPatch(&core.Rectangle{
			Patch: core.Patch{
				FaceColor: render.Color{R: 0.92, G: 0.91, B: 0.84, A: 1},
				EdgeColor: render.Color{R: 0.18, G: 0.22, B: 0.25, A: 1},
				EdgeWidth: 0.85,
				Hatch:     hatch,
			},
			XY:     geom.Pt{X: 0.75 + float64(i)*1.38, Y: 3.58},
			Width:  0.9,
			Height: 0.78,
			Coords: core.Coords(core.CoordData),
		})
	}

	arrowStyleA, _ := core.ArrowStyleFromString("->,head_length=0.35,head_width=0.22")
	arrowStyleB, _ := core.ArrowStyleFromString("|-|")
	arrowStyleC, _ := core.ArrowStyleFromString("wedge,tail_width=0.26,shrink_factor=0.35")
	connectionStyleA, _ := core.ConnectionStyleFromString("arc,armA=0.9,armB=0.65,rad=0.18")
	connectionStyleB, _ := core.ConnectionStyleFromString("bar,fraction=0.25,angle=0")
	connectionStyleC, _ := core.ConnectionStyleFromString("arc3,rad=0.22")

	arrows := []struct {
		a, b            geom.Pt
		arrowStyle      core.ArrowStyle
		connectionStyle core.ConnectionStyle
		face            render.Color
		edge            render.Color
	}{
		{
			a:               geom.Pt{X: 0.9, Y: 2.25},
			b:               geom.Pt{X: 3.1, Y: 2.6},
			arrowStyle:      arrowStyleA,
			connectionStyle: connectionStyleA,
			face:            render.Color{R: 0.28, G: 0.48, B: 0.82, A: 1},
			edge:            render.Color{R: 0.12, G: 0.24, B: 0.47, A: 1},
		},
		{
			a:               geom.Pt{X: 4.0, Y: 2.62},
			b:               geom.Pt{X: 6.25, Y: 1.88},
			arrowStyle:      arrowStyleB,
			connectionStyle: connectionStyleB,
			face:            render.Color{A: 0},
			edge:            render.Color{R: 0.66, G: 0.28, B: 0.23, A: 1},
		},
		{
			a:               geom.Pt{X: 7.05, Y: 2.08},
			b:               geom.Pt{X: 10.8, Y: 2.58},
			arrowStyle:      arrowStyleC,
			connectionStyle: connectionStyleC,
			face:            render.Color{R: 0.40, G: 0.66, B: 0.35, A: 0.82},
			edge:            render.Color{R: 0.20, G: 0.38, B: 0.18, A: 1},
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
			ArrowStyle:      item.arrowStyle,
			ConnectionStyle: item.connectionStyle,
			MutationScale:   15,
			Coords:          core.Coords(core.CoordData),
		})
	}
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}
