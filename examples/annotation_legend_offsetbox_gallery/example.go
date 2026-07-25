// Package annotation_legend_offsetbox_gallery is a user-facing gallery for
// annotations, legends, and anchored offset boxes.
package annotation_legend_offsetbox_gallery

import (
	"image"
	"image/color"
	"math"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

const (
	Width  = 1040
	Height = 720
	DPI    = 100
)

// Plot builds the showcase figure (backend-agnostic).
func Plot() *core.Figure {
	fig := core.NewFigure(Width, Height)
	fig.Text(0.05, 0.95, "Annotations, Legends, and Offset Boxes", core.TextOptions{FontSize: 13})

	top := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.55}, Max: geom.Pt{X: 0.94, Y: 0.88}})
	bottom := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.10}, Max: geom.Pt{X: 0.94, Y: 0.41}})
	addAnnotationLegendPanel(top)
	addOffsetBoxPanel(bottom)
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

func addAnnotationLegendPanel(ax *core.Axes) {
	ax.SetTitle("Annotation and Legend Layout")
	ax.SetXLim(0, 8)
	ax.SetYLim(-1.25, 1.25)
	ax.SetXLabel("x")
	ax.SetYLabel("signal")

	x := make([]float64, 160)
	sin := make([]float64, 160)
	cos := make([]float64, 160)
	for i := range x {
		xv := 8 * float64(i) / float64(len(x)-1)
		x[i] = xv
		sin[i] = math.Sin(xv)
		cos[i] = 0.65 * math.Cos(xv*0.8)
	}
	blue := render.Color{R: 0.12, G: 0.31, B: 0.68, A: 1}
	orange := render.Color{R: 0.86, G: 0.43, B: 0.16, A: 1}
	lineWidth := 2.0
	_, _ = ax.Plot(x, sin, core.PlotOptions{Color: &blue, LineWidth: &lineWidth, Label: "sin(x)"})
	_, _ = ax.Plot(x, cos, core.PlotOptions{Color: &orange, LineWidth: &lineWidth, Label: "0.65 cos(0.8x)"})

	arrow, _ := core.ArrowStyleFromString("-|>,head_length=0.35,head_width=0.20")
	arc, _ := core.ConnectionStyleFromString("arc3,rad=0.25")
	ax.Annotate("curved arrow\nbbox label", math.Pi/2, 1, core.AnnotationOptions{
		OffsetX:         optional.Of(pt(68)),
		OffsetY:         optional.Of(pt(-42)),
		FontSize:        10,
		HAlign:          core.TextAlignCenter,
		VAlign:          core.TextVAlignMiddle,
		ArrowStyle:      arrow,
		ConnectionStyle: arc,
		ArrowColor:      blue,
		ArrowWidth:      optional.Of(1.2),
		BBox:            galleryBox(10, 0.28, render.Color{R: 0.92, G: 0.97, B: 1.00, A: 0.90}, blue),
	})
	ax.AnnotationBbox("offset box", 5.65, -0.25, core.AnnotationBboxOptions{
		BoxPosition: &geom.Pt{X: 6.75, Y: 0.55},
		Padding:     pt(3),
		FaceColor:   render.Color{R: 0.96, G: 0.92, B: 1.00, A: 0.92},
		EdgeColor:   render.Color{R: 0.42, G: 0.25, B: 0.60, A: 1},
		FontSize:    10,
		Arrow:       true,
		ArrowColor:  render.Color{R: 0.42, G: 0.25, B: 0.60, A: 1},
	})

	legend := ax.AddLegend()
	legend.Location = core.LegendUpperRight
	legend.Title = "Handles"
	legend.NumColumns = 2

	proxyLegend := ax.AddLegend()
	proxyLegend.Location = core.LegendLowerRight
	proxyLegend.FrameOn = false
	proxyLegend.AddEntry("proxy patch", core.LegendEntryOptions{
		Sample:    core.LegendSamplePatch,
		FaceColor: render.Color{R: 0.94, G: 0.78, B: 0.38, A: 0.95},
		EdgeColor: render.Color{R: 0.46, G: 0.30, B: 0.08, A: 1},
		EdgeWidth: 1.1,
		Hatch:     "//",
	})
}

func addOffsetBoxPanel(ax *core.Axes) {
	ax.SetTitle("Anchored Offset Boxes")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 4)
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	ax.AddAnchoredText("AnchoredText\nupper left", core.AnchoredTextOptions{
		Location:        core.LegendUpperLeft,
		Padding:         pt(4),
		Inset:           pt(8),
		BackgroundColor: render.Color{R: 0.98, G: 0.98, B: 0.92, A: 0.94},
		BorderColor:     render.Color{R: 0.45, G: 0.42, B: 0.16, A: 1},
		BorderWidth:     1,
		FontSize:        10,
	})

	area := ax.AddAnchoredDrawingArea(58, 34, core.AnchoredDrawingAreaOptions{
		Location:        core.LegendUpperRight,
		Padding:         pt(4),
		Inset:           pt(8),
		BackgroundColor: render.Color{R: 0.98, G: 0.94, B: 0.88, A: 0.94},
		BorderColor:     render.Color{R: 0.55, G: 0.30, B: 0.12, A: 1},
		BorderWidth:     1,
	})
	area.AddPath(localTrianglePath(), render.Paint{
		Fill:      render.Color{R: 0.84, G: 0.44, B: 0.18, A: 0.88},
		Stroke:    render.Color{R: 0.43, G: 0.20, B: 0.08, A: 1},
		LineWidth: 1,
	})

	packer := ax.AddAnchoredPacker(core.PackHorizontal, core.AnchoredPackerOptions{
		Location:        core.LegendLowerLeft,
		Padding:         pt(4),
		Inset:           pt(8),
		Sep:             pt(6),
		BackgroundColor: render.Color{R: 0.92, G: 0.97, B: 1.00, A: 0.94},
		BorderColor:     render.Color{R: 0.16, G: 0.37, B: 0.54, A: 1},
		BorderWidth:     1,
		FontSize:        10,
		TextColor:       render.Color{R: 0.10, G: 0.24, B: 0.35, A: 1},
	})
	packer.AddDrawingArea(18, 18).AddPath(localDiamondPath(), render.Paint{
		Fill:      render.Color{R: 0.25, G: 0.62, B: 0.78, A: 0.9},
		Stroke:    render.Color{R: 0.08, G: 0.25, B: 0.35, A: 1},
		LineWidth: 1,
	})
	packer.AddImage(smallAnnotationImage(), 1.35)
	packer.AddText("HPacker")

	fill := true
	ax.AddAnchoredSizeBar(1.4, "1.4 data", core.AnchoredSizeBarOptions{
		Location:        core.LegendLowerRight,
		Padding:         pt(4),
		Inset:           pt(8),
		Sep:             pt(4),
		SizeVertical:    0.10,
		FillBar:         &fill,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 0.86},
		BorderColor:     render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1},
		BorderWidth:     1,
		LineWidth:       1,
		FontSize:        10,
	})
}

func galleryBox(fontSize, pad float64, face, edge render.Color) *core.TextBBoxOptions {
	return &core.TextBBoxOptions{
		FaceColor:    face,
		EdgeColor:    edge,
		LineWidth:    0.9,
		Padding:      fontSize * pad * DPI / 72,
		CornerRadius: 5,
	}
}

func smallAnnotationImage() *render.ImageData {
	img := image.NewRGBA(image.Rect(0, 0, 12, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 12; x++ {
			c := color.RGBA{R: 231, G: 242, B: 255, A: 255}
			if (x+y)%2 == 0 {
				c = color.RGBA{R: 96, G: 150, B: 209, A: 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return render.NewImageData(img)
}

func localTrianglePath() geom.Path {
	p := geom.Path{}
	p.MoveTo(geom.Pt{X: 7, Y: 28})
	p.LineTo(geom.Pt{X: 50, Y: 25})
	p.LineTo(geom.Pt{X: 30, Y: 6})
	p.Close()
	return p
}

func localDiamondPath() geom.Path {
	p := geom.Path{}
	p.MoveTo(geom.Pt{X: 9, Y: 1})
	p.LineTo(geom.Pt{X: 17, Y: 9})
	p.LineTo(geom.Pt{X: 9, Y: 17})
	p.LineTo(geom.Pt{X: 1, Y: 9})
	p.Close()
	return p
}

func pt(points float64) float64 {
	return points * DPI / 72
}
