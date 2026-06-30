package text_annotation_matrix

import (
	"image"
	"image/color"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
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
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.08, Y: 0.12}, Max: geom.Pt{X: 0.94, Y: 0.90}})
	ax.SetTitle("Text + Annotation Matrix")
	ax.SetXLim(0, 6)
	ax.SetYLim(0, 5)

	addTextMatrix(ax)
	addAnnotationMatrix(ax)
	addOffsetBoxMatrix(ax)
	return fig
}

func Render() image.Image {
	return common.RenderFixtureFigure(Plot(), Width, Height)
}

func addTextMatrix(ax *core.Axes) {
	italicBold := render.FontProperties{
		Families: []string{"DejaVu Sans"},
		Style:    render.FontStyleItalic,
		Weight:   700,
	}
	ax.Text(0.55, 4.25, "bold italic\nbbox", core.TextOptions{
		FontSize:       13,
		Color:          render.Color{R: 0.12, G: 0.20, B: 0.42, A: 1},
		HAlign:         core.TextAlignLeft,
		VAlign:         core.TextVAlignTop,
		FontProperties: &italicBold,
		BBox: &core.TextBBoxOptions{
			FaceColor:    render.Color{R: 0.88, G: 0.92, B: 1.00, A: 0.85},
			EdgeColor:    render.Color{R: 0.20, G: 0.32, B: 0.68, A: 1},
			LineWidth:    1,
			Padding:      fontPad(13, 0.3),
			CornerRadius: 5,
		},
	})
	ax.Text(2.25, 4.45, "rotated label", core.TextOptions{
		FontSize: 12,
		Color:    render.Color{R: 0.56, G: 0.24, B: 0.16, A: 1},
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		Angle:    -28,
		BBox: &core.TextBBoxOptions{
			FaceColor: render.Color{R: 1.00, G: 0.90, B: 0.78, A: 0.72},
			EdgeColor: render.Color{R: 0.65, G: 0.35, B: 0.12, A: 1},
			LineWidth: 0.8,
			Padding:   fontPad(12, 0.25),
		},
	})
}

func addAnnotationMatrix(ax *core.Axes) {
	arc, _ := core.ConnectionStyleFromString("arc3,rad=0.25")
	arrow, _ := core.ArrowStyleFromString("-|>,head_length=0.35,head_width=0.20")
	ax.Annotate("bbox arrow", 2.1, 2.15, core.AnnotationOptions{
		OffsetX:         pt(72),
		OffsetY:         pt(-40),
		FontSize:        11,
		Color:           render.Color{R: 0.15, G: 0.18, B: 0.22, A: 1},
		ArrowColor:      render.Color{R: 0.18, G: 0.39, B: 0.70, A: 1},
		ArrowWidth:      1.2,
		ArrowHeadSize:   9,
		ArrowStyle:      arrow,
		ConnectionStyle: arc,
		HAlign:          core.TextAlignCenter,
		VAlign:          core.TextVAlignMiddle,
		BBox: &core.TextBBoxOptions{
			FaceColor: render.Color{R: 0.92, G: 0.97, B: 0.92, A: 0.92},
			EdgeColor: render.Color{R: 0.23, G: 0.48, B: 0.20, A: 1},
			LineWidth: 1,
			Padding:   fontPad(11, 0.25),
		},
	})

	clip := true
	ax.Annotate("clipped", -0.4, 4.7, core.AnnotationOptions{
		OffsetX:         pt(40),
		OffsetY:         pt(-20),
		AnnotationClip:  &clip,
		ArrowColor:      render.Color{R: 0.7, G: 0.1, B: 0.1, A: 1},
		ConnectionStyle: arc,
	})

	boxPos := geom.Pt{X: 4.7, Y: 3.9}
	align := geom.Pt{X: 0.5, Y: 0.5}
	ax.AnnotationBbox("TextArea box", 3.75, 3.05, core.AnnotationBboxOptions{
		BoxPosition:  &boxPos,
		BoxAlignment: &align,
		Padding:      fontPad(10, 0.35),
		FaceColor:    render.Color{R: 0.96, G: 0.90, B: 1.00, A: 0.92},
		EdgeColor:    render.Color{R: 0.45, G: 0.24, B: 0.62, A: 1},
		TextColor:    render.Color{R: 0.23, G: 0.15, B: 0.35, A: 1},
		FontSize:     10,
		Arrow:        true,
		ArrowColor:   render.Color{R: 0.45, G: 0.24, B: 0.62, A: 1},
	})
}

func addOffsetBoxMatrix(ax *core.Axes) {
	ax.AddAnchoredText("anchored\ntext", core.AnchoredTextOptions{
		Location:        core.LegendUpperLeft,
		Padding:         fontPad(9, 0.35),
		Inset:           fontPad(9, 0.6),
		RowGap:          2,
		BoxPadding:      0,
		BackgroundColor: render.Color{R: 0.98, G: 0.98, B: 0.94, A: 0.92},
		BorderColor:     render.Color{R: 0.45, G: 0.45, B: 0.24, A: 1},
		BorderWidth:     1,
		TextColor:       render.Color{R: 0.18, G: 0.18, B: 0.10, A: 1},
		FontSize:        9,
	})

	img := smallAnnotationImage()
	imgPos := geom.Pt{X: 4.7, Y: 1.25}
	ax.AnnotationBbox("", 3.75, 1.55, core.AnnotationBboxOptions{
		BoxPosition: &imgPos,
		Image:       img,
		ImageZoom:   2.4,
		Padding:     fontPad(10, 0.25),
		FaceColor:   render.Color{R: 1, G: 1, B: 1, A: 0.95},
		EdgeColor:   render.Color{R: 0.25, G: 0.25, B: 0.25, A: 1},
		Arrow:       true,
		ArrowColor:  render.Color{R: 0.25, G: 0.25, B: 0.25, A: 1},
	})

	area := ax.AddAnchoredDrawingArea(46, 28, core.AnchoredDrawingAreaOptions{
		Location:        core.LegendUpperRight,
		Padding:         fontPad(10, 0.4),
		Inset:           fontPad(10, 0.7),
		BackgroundColor: render.Color{R: 0.98, G: 0.96, B: 0.88, A: 0.92},
		BorderColor:     render.Color{R: 0.45, G: 0.36, B: 0.14, A: 1},
		BorderWidth:     1,
	})
	area.AddPath(localTrianglePath(), render.Paint{
		Fill:      render.Color{R: 0.83, G: 0.44, B: 0.18, A: 0.85},
		Stroke:    render.Color{R: 0.43, G: 0.20, B: 0.08, A: 1},
		LineWidth: 1,
	})

	packer := ax.AddAnchoredPacker(core.PackHorizontal, core.AnchoredPackerOptions{
		Location:        core.LegendLowerLeft,
		Padding:         fontPad(10, 0.4),
		Inset:           fontPad(10, 0.6),
		Sep:             pt(5),
		BackgroundColor: render.Color{R: 0.93, G: 0.97, B: 1.00, A: 0.92},
		BorderColor:     render.Color{R: 0.16, G: 0.37, B: 0.54, A: 1},
		BorderWidth:     1,
		FontSize:        9,
		TextColor:       render.Color{R: 0.10, G: 0.24, B: 0.35, A: 1},
	})
	packer.AddDrawingArea(18, 18).AddPath(localDiamondPath(), render.Paint{
		Fill:      render.Color{R: 0.25, G: 0.62, B: 0.78, A: 0.9},
		Stroke:    render.Color{R: 0.08, G: 0.25, B: 0.35, A: 1},
		LineWidth: 1,
	})
	packer.AddText("HPack")

	fill := true
	ax.AddAnchoredSizeBar(1.25, "1.25 data", core.AnchoredSizeBarOptions{
		Location:        core.LegendLowerRight,
		Padding:         fontPad(9, 0.4),
		Inset:           fontPad(9, 0.6),
		Sep:             pt(4),
		SizeVertical:    0.09,
		FillBar:         &fill,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 0.82},
		BorderColor:     render.Color{R: 0.20, G: 0.20, B: 0.20, A: 1},
		BorderWidth:     1,
		LineWidth:       1,
		Color:           render.Color{R: 0.08, G: 0.08, B: 0.08, A: 1},
		FontSize:        9,
	})
}

func smallAnnotationImage() *render.ImageData {
	img := image.NewRGBA(image.Rect(0, 0, 12, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 12; x++ {
			c := color.RGBA{R: 229, G: 239, B: 255, A: 255}
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
	p.MoveTo(geom.Pt{X: 6, Y: 24})
	p.LineTo(geom.Pt{X: 40, Y: 21})
	p.LineTo(geom.Pt{X: 24, Y: 5})
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
	return common.ReferencePointsToPixels(points)
}

func fontPad(fontSize, fraction float64) float64 {
	return pt(fontSize * fraction)
}
