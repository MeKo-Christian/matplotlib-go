package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestRelativeAnchoredBoxLocatorCentersBox(t *testing.T) {
	locator := RelativeAnchoredBoxLocator{
		X:      0.5,
		Y:      0.5,
		HAlign: BoxAlignCenter,
		VAlign: BoxAlignMiddle,
	}

	rect := locator.Rect(
		geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 100}},
		20,
		10,
	)

	if rect != (geom.Rect{
		Min: geom.Pt{X: 50, Y: 55},
		Max: geom.Pt{X: 70, Y: 65},
	}) {
		t.Fatalf("relative locator rect = %+v", rect)
	}
}

func TestAnchoredOffsetLocatorAppliesCornerAndPixelOffset(t *testing.T) {
	locator := NewAnchoredOffsetLocator(LegendUpperLeft, 8, 5, 3)
	rect := locator.Rect(
		geom.Rect{Min: geom.Pt{X: 10, Y: 20}, Max: geom.Pt{X: 110, Y: 100}},
		20,
		10,
	)

	if rect != (geom.Rect{
		Min: geom.Pt{X: 23, Y: 31},
		Max: geom.Pt{X: 43, Y: 41},
	}) {
		t.Fatalf("offset locator rect = %+v", rect)
	}
}

func TestBBoxToAnchorLocatorUsesMatplotlibFigureFractions(t *testing.T) {
	locator := BBoxToAnchorLocator{
		X:        0.99,
		Y:        0.90,
		Location: LegendUpperRight,
	}

	rect := locator.RectWithInset(
		geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 1100, Y: 720}},
		104,
		88,
		7,
	)

	// anchor: X=1100*0.99=1089, Y=720-720*0.90=72
	// LegendUpperRight with borderaxespad inset: minX=1089-104-7=978, minY=72+7=79
	if rect != (geom.Rect{
		Min: geom.Pt{X: 978, Y: 79},
		Max: geom.Pt{X: 1082, Y: 167},
	}) {
		t.Fatalf("bbox_to_anchor locator rect = %+v", rect)
	}
}

func TestAnchoredTextBoxBoxRectUsesLocator(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	box := ax.AddAnchoredText("Centered", AnchoredTextOptions{
		Location: LegendUpperLeft,
		Locator: RelativeAnchoredBoxLocator{
			X:      0.5,
			Y:      0.5,
			HAlign: BoxAlignCenter,
			VAlign: BoxAlignMiddle,
		},
	})

	var r figureLayoutRecordingRenderer
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	rect, ok := box.boxRect(&r, ctx)
	if !ok {
		t.Fatal("boxRect() returned !ok")
	}

	if !floatApprox(rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2, 1e-9) {
		t.Fatalf("box center x = %v, want %v", rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2)
	}
	if !floatApprox(rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2, 1e-9) {
		t.Fatalf("box center y = %v, want %v", rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2)
	}
}

func TestAnchoredTextOptionsMergeWithDefaults(t *testing.T) {
	box := newAnchoredTextBox("note", styleRCForAnchoredTextTest(), AnchoredTextOptions{
		Location:        LegendLowerRight,
		Padding:         4,
		Inset:           6,
		CornerRadius:    3,
		BackgroundColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		FontSize:        10,
	})

	ctx := &DrawContext{RC: styleRCForAnchoredTextTest()}
	if got, want := box.resolvedRowGap(10, ctx), pointsToPixels(ctx.RC, 2); !floatApprox(got, want, 1e-9) {
		t.Fatalf("resolved row gap = %v, want %v", got, want)
	}
	if box.BorderWidth != 1 {
		t.Fatalf("border width = %v, want default 1", box.BorderWidth)
	}
	if box.TextColor == (render.Color{}) || box.BorderColor == (render.Color{}) {
		t.Fatalf("expected text and border colors to inherit defaults: %+v", box)
	}
}

func styleRCForAnchoredTextTest() style.RC {
	rc := style.Default
	rc.LegendTextColor = render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}
	rc.LegendBorderColor = render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1}
	return rc
}

func TestLegendBoxRectUsesLocator(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{Label: "signal"})

	legend := ax.AddLegend()
	legend.SetLocator(RelativeAnchoredBoxLocator{
		X:      0.5,
		Y:      0.5,
		HAlign: BoxAlignCenter,
		VAlign: BoxAlignMiddle,
	})

	var r legendRecordingRenderer
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	rect, ok := legend.boxRect(&r, ctx)
	if !ok {
		t.Fatal("boxRect() returned !ok")
	}

	if !floatApprox(rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2, 1e-9) {
		t.Fatalf("legend center x = %v, want %v", rect.Min.X+rect.W()/2, ctx.Clip.Min.X+ctx.Clip.W()/2)
	}
	if !floatApprox(rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2, 1e-9) {
		t.Fatalf("legend center y = %v, want %v", rect.Min.Y+rect.H()/2, ctx.Clip.Min.Y+ctx.Clip.H()/2)
	}
}
