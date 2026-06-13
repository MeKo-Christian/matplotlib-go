package core

import (
	"image"
	"math"
	"strings"
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

	// Display space is y-up: the upper-left corner is at high Y (bounds Max.Y),
	// so the offset box lands near the top at y 85..95.
	if rect != (geom.Rect{
		Min: geom.Pt{X: 23, Y: 85},
		Max: geom.Pt{X: 43, Y: 95},
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

	// Matplotlib transforms bbox_to_anchor through BboxTransformTo(parent.bbox),
	// so the y fraction grows upward in display coordinates. The legend's
	// upper-right corner is inset from the transformed anchor.
	if rect != (geom.Rect{
		Min: geom.Pt{X: 978, Y: 553},
		Max: geom.Pt{X: 1082, Y: 641},
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

func TestAnchoredTextBoxDrawsMultilineContentTopDown(t *testing.T) {
	box := newAnchoredTextBox("first\nsecond", styleRCForAnchoredTextTest(), AnchoredTextOptions{
		Location: LegendUpperLeft,
		Padding:  4,
		Inset:    6,
		FontSize: 10,
	})
	ctx := createTestDrawContext()
	r := &textRecordingRenderer{}

	box.Draw(r, ctx)

	if len(r.texts) != 2 {
		t.Fatalf("anchored multiline text draws = %d, want 2: %v", len(r.texts), r.texts)
	}
	if r.texts[0] != "first" || r.texts[1] != "second" {
		t.Fatalf("anchored multiline draw order = %v, want source order", r.texts)
	}
	if !(r.origins[1].Y < r.origins[0].Y) {
		t.Fatalf("anchored multiline should draw top-down in y-up display space, got origins %v", r.origins)
	}
}

func TestAnchoredTextBoxMultilineHeightUsesMeasuredTextAreaMetrics(t *testing.T) {
	box := newAnchoredTextBox("anchored\ntext", styleRCForAnchoredTextTest(), AnchoredTextOptions{
		Location:   LegendUpperLeft,
		Padding:    pointsToPixels(style.Default, 9*0.35),
		Inset:      pointsToPixels(style.Default, 9*0.6),
		RowGap:     2,
		BoxPadding: 0,
		FontSize:   9,
	})
	ctx := createTestDrawContext()
	ctx.RC.DPI = 100
	r := &anchoredTextMetricRenderer{}
	lines := strings.Split(box.Content, "\n")
	layouts := make([]singleLineTextLayout, len(lines))
	for i, line := range lines {
		layouts[i] = measureSingleLineTextLayout(r, line, box.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	}

	got := box.layout(r, ctx, layouts, box.FontSize).patchBox

	// Matplotlib TextArea("anchored\ntext", size=9) is 58.75 x 28 px at
	// 100 DPI; AnchoredText pad=0.35 expands it by 4.375 px on each side.
	wantH := 28.0 + 2*pointsToPixels(ctx.RC, 9*0.35)
	if !floatApprox(got.H(), wantH, 1e-9) {
		t.Fatalf("anchored multiline frame height = %v, want Matplotlib TextArea height %v", got.H(), wantH)
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

func TestAnchoredSizeBarDrawsDataScaledBarAndLabel(t *testing.T) {
	barColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	frameFill := render.Color{R: 1, G: 1, B: 1, A: 1}
	frameEdge := render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}
	bar := (&Axes{}).AddAnchoredSizeBar(2, "2 units", AnchoredSizeBarOptions{
		Location:        LegendLowerRight,
		Padding:         4,
		Inset:           6,
		Sep:             3,
		Color:           barColor,
		BackgroundColor: frameFill,
		BorderColor:     frameEdge,
		BorderWidth:     1,
	})

	ctx := createTestDrawContext()
	r := &textRecordingRenderer{}
	bar.Draw(r, ctx)

	if !containsTextString(r.texts, "2 units") {
		t.Fatalf("anchored size bar label not drawn, got %v", r.texts)
	}
	if !hasPathPaint(r.pathPaints, frameFill, frameEdge, 1) {
		t.Fatalf("anchored size bar frame paint not found in %+v", r.pathPaints)
	}
	var gotBar geom.Path
	for _, call := range r.pathCalls {
		if len(call.path.C) == 2 && call.path.C[0] == geom.MoveTo && call.path.C[1] == geom.LineTo && call.paint.Stroke == barColor {
			gotBar = call.path
			break
		}
	}
	if len(gotBar.V) != 2 {
		t.Fatalf("anchored size bar line path not found in %+v", r.pathCalls)
	}
	if !floatApprox(gotBar.V[1].X-gotBar.V[0].X, 20, 1e-9) {
		t.Fatalf("size bar display width = %v, want 20", gotBar.V[1].X-gotBar.V[0].X)
	}
	if r.origins[0].Y >= gotBar.V[0].Y {
		t.Fatalf("default label should be below the bar, got label origin %+v bar %+v", r.origins[0], gotBar.V)
	}
}

func TestAnchoredDrawingAreaDrawsLocalPath(t *testing.T) {
	stroke := render.Color{R: 0.8, G: 0.1, B: 0.2, A: 1}
	frameFill := render.Color{R: 1, G: 1, B: 1, A: 1}
	frameEdge := render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1}
	area := (&Axes{}).AddAnchoredDrawingArea(40, 20, AnchoredDrawingAreaOptions{
		Location:        LegendUpperLeft,
		Padding:         0,
		Inset:           5,
		FrameOn:         boolPtr(true),
		BackgroundColor: frameFill,
		BorderColor:     frameEdge,
		BorderWidth:     1,
	})
	area.AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 40, Y: 20}},
	}, render.Paint{Stroke: stroke, LineWidth: 2})

	ctx := createTestDrawContext()
	scale := pointsToPixels(ctx.RC, 1)
	r := &recordingRenderer{}
	area.Draw(r, ctx)

	if !recordedPaintExists(r.pathCalls, frameFill, frameEdge, 1) {
		t.Fatalf("anchored drawing area frame paint not found in %+v", r.pathCalls)
	}
	var child geom.Path
	for _, call := range r.pathCalls {
		if call.paint.Stroke == stroke {
			child = call.path
			break
		}
	}
	if len(child.V) != 2 {
		t.Fatalf("anchored drawing area child path not found in %+v", r.pathCalls)
	}
	// Display space is y-up: the UpperLeft area sits near the top edge (495), with
	// its local lower-left at the bottom of the box and upper-right at the top.
	if got, want := child.V[0], (geom.Pt{X: 5, Y: 495 - 20*scale}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("local lower-left point mapped to %+v, want %+v", got, want)
	}
	if got, want := child.V[1], (geom.Pt{X: 5 + 40*scale, Y: 495}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("local upper-right point mapped to %+v, want %+v", got, want)
	}
}

func TestAnchoredDrawingAreaScalesLocalCoordinatesByDPI(t *testing.T) {
	stroke := render.Color{R: 0.8, G: 0.1, B: 0.2, A: 1}
	area := (&Axes{}).AddAnchoredDrawingArea(40, 20, AnchoredDrawingAreaOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    0,
		FrameOn:  boolPtr(false),
	})
	area.AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 40, Y: 20}},
	}, render.Paint{Stroke: stroke, LineWidth: 2})

	ctx := createTestDrawContext()
	ctx.RC = style.Apply(ctx.RC, style.WithDPI(144))
	r := &recordingRenderer{}
	area.Draw(r, ctx)

	child := recordedStrokePath(r.pathCalls, stroke)
	if len(child.V) != 2 {
		t.Fatalf("anchored drawing area child path not found in %+v", r.pathCalls)
	}
	// Display space is y-up: the UpperLeft area (inset 0) sits at the top edge
	// (500), so local lower-left is at 500-40=460 and upper-right at 500.
	if got, want := child.V[0], (geom.Pt{X: 0, Y: 460}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("local lower-left point mapped to %+v, want %+v", got, want)
	}
	if got, want := child.V[1], (geom.Pt{X: 80, Y: 500}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("local upper-right point mapped to %+v, want %+v", got, want)
	}
}

func TestAnchoredDrawingAreaCanClipChildren(t *testing.T) {
	area := (&Axes{}).AddAnchoredDrawingArea(40, 20, AnchoredDrawingAreaOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    5,
		Clip:     true,
		FrameOn:  boolPtr(false),
	})
	area.AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: -10, Y: -10}, {X: 50, Y: 30}},
	}, render.Paint{Stroke: render.Color{A: 1}, LineWidth: 1})

	ctx := createTestDrawContext()
	scale := pointsToPixels(ctx.RC, 1)
	r := &clipRecordingRenderer{}
	area.Draw(r, ctx)

	// Display space is y-up: the UpperLeft content box sits near the top (495).
	wantClip := geom.Rect{Min: geom.Pt{X: 5, Y: 495 - 20*scale}, Max: geom.Pt{X: 5 + 40*scale, Y: 495}}
	if len(r.rects) != 1 || !approxRect(r.rects[0], wantClip, 1e-9) {
		t.Fatalf("drawing area clip rects = %+v, want [%+v]", r.rects, wantClip)
	}
	wantEvents := []string{"save", "clipRect", "restore"}
	if len(r.events) != len(wantEvents) {
		t.Fatalf("clip events = %v, want %v", r.events, wantEvents)
	}
	for i := range wantEvents {
		if r.events[i] != wantEvents[i] {
			t.Fatalf("clip events = %v, want %v", r.events, wantEvents)
		}
	}
}

func TestAnchoredDrawingAreaFrameUsesMatplotlibSnapAuto(t *testing.T) {
	area := (&Axes{}).AddAnchoredDrawingArea(10, 10, AnchoredDrawingAreaOptions{
		Location:        LegendUpperLeft,
		Padding:         1.25,
		Inset:           5.4,
		FrameOn:         boolPtr(true),
		BackgroundColor: render.Color{A: 1},
		BorderColor:     render.Color{A: 1},
	})
	ctx := createTestDrawContext()
	r := &recordingRenderer{}

	area.Draw(r, ctx)

	if len(r.pathCalls) == 0 || len(r.pathCalls[0].path.V) < 4 {
		t.Fatalf("expected frame path, got %+v", r.pathCalls)
	}
	call := r.pathCalls[0]
	if call.paint.Snap != render.SnapAuto {
		t.Fatalf("frame snap mode = %v, want Matplotlib SnapAuto", call.paint.Snap)
	}
	foundFractional := false
	for _, pt := range call.path.V[:4] {
		if !floatApprox(pt.X, math.Round(pt.X), 1e-9) || !floatApprox(pt.Y, math.Round(pt.Y), 1e-9) {
			foundFractional = true
			break
		}
	}
	if !foundFractional {
		t.Fatalf("frame path was pre-rounded to integer vertices; want unsnapped patch bounds for renderer SnapAuto: %+v", call.path.V[:4])
	}
}

func TestAnchoredPackerPacksDrawingAreaAndTextHorizontally(t *testing.T) {
	stroke := render.Color{R: 0.9, G: 0.1, B: 0.2, A: 1}
	textColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	packer := (&Axes{}).AddAnchoredPacker(PackHorizontal, AnchoredPackerOptions{
		Location:  LegendUpperLeft,
		Padding:   0,
		Inset:     5,
		Sep:       4,
		FrameOn:   boolPtr(false),
		Align:     PackAlignCenter,
		FontSize:  10,
		TextColor: textColor,
	})
	packer.AddDrawingArea(20, 10).AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 20, Y: 10}},
	}, render.Paint{Stroke: stroke, LineWidth: 1})
	packer.AddText("Go")

	ctx := createTestDrawContext()
	scale := pointsToPixels(ctx.RC, 1)
	r := &textRecordingRenderer{}
	packer.Draw(r, ctx)

	if !containsTextString(r.texts, "Go") {
		t.Fatalf("packed text not drawn, got %v", r.texts)
	}
	var child geom.Path
	for _, call := range r.pathCalls {
		if call.paint.Stroke == stroke {
			child = call.path
			break
		}
	}
	if len(child.V) != 2 {
		t.Fatalf("packed drawing area path not found in %+v", r.pathCalls)
	}
	// Display space is y-up: the UpperLeft packer sits near the top edge (495).
	if got, want := child.V[0], (geom.Pt{X: 5, Y: 495 - 10*scale}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("drawing child lower-left mapped to %+v, want %+v", got, want)
	}
	if got, want := child.V[1], (geom.Pt{X: 5 + 20*scale, Y: 495}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("drawing child upper-right mapped to %+v, want %+v", got, want)
	}
	if len(r.origins) != 1 {
		t.Fatalf("packed text origins = %v, want one", r.origins)
	}
	wantOrigin := geom.Pt{X: 5 + 20*scale + 4, Y: 500 - (5 + 10*scale/2 + 3)}
	if got := r.origins[0]; !floatApprox(got.X, wantOrigin.X, 1e-9) || !floatApprox(got.Y, wantOrigin.Y, 1e-9) {
		t.Fatalf("packed text origin = %+v, want %+v", got, wantOrigin)
	}
}

func TestAnchoredPackerOptionsKeepMatplotlibCenterAlignDefault(t *testing.T) {
	packer := newAnchoredPacker(PackHorizontal, styleRCForAnchoredTextTest(), AnchoredPackerOptions{
		Location: LegendLowerLeft,
		Padding:  4,
		Inset:    8,
		Sep:      6,
	})
	if got := packer.Align; got != PackAlignCenter {
		t.Fatalf("default packer align with options = %v, want Matplotlib center", got)
	}

	start := newAnchoredPacker(PackHorizontal, styleRCForAnchoredTextTest(), AnchoredPackerOptions{
		Align: PackAlignStart,
	})
	if got := start.Align; got != PackAlignStart {
		t.Fatalf("explicit start packer align = %v, want PackAlignStart", got)
	}
}

func TestAnchoredPackerPacksChildrenVertically(t *testing.T) {
	topStroke := render.Color{R: 0.1, G: 0.5, B: 0.2, A: 1}
	bottomStroke := render.Color{R: 0.2, G: 0.1, B: 0.7, A: 1}
	packer := (&Axes{}).AddAnchoredPacker(PackVertical, AnchoredPackerOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    5,
		Sep:      3,
		FrameOn:  boolPtr(false),
		Align:    PackAlignEnd,
	})
	packer.AddDrawingArea(20, 10).AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 20, Y: 10}},
	}, render.Paint{Stroke: topStroke, LineWidth: 1})
	packer.AddDrawingArea(10, 6).AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 10, Y: 6}},
	}, render.Paint{Stroke: bottomStroke, LineWidth: 1})

	ctx := createTestDrawContext()
	scale := pointsToPixels(ctx.RC, 1)
	r := &recordingRenderer{}
	packer.Draw(r, ctx)

	top := recordedStrokePath(r.pathCalls, topStroke)
	bottom := recordedStrokePath(r.pathCalls, bottomStroke)
	if len(top.V) != 2 || len(bottom.V) != 2 {
		t.Fatalf("packed vertical child paths not found: top=%+v bottom=%+v calls=%+v", top, bottom, r.pathCalls)
	}
	// Display space is y-up: vertical packing runs top-to-bottom from 495, so the
	// first (top) child sits highest and the second stacks below it.
	if got, want := top.V[0], (geom.Pt{X: 5, Y: 495 - 10*scale}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("top child lower-left mapped to %+v, want %+v", got, want)
	}
	if got, want := bottom.V[0], (geom.Pt{X: 5 + 20*scale - 10*scale, Y: 495 - 10*scale - 3 - 6*scale}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("bottom child lower-left mapped to %+v, want %+v", got, want)
	}
}

func TestAnchoredPackerPacksImageChildren(t *testing.T) {
	stroke := render.Color{R: 0.8, G: 0.2, B: 0.1, A: 1}
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 3)))
	packer := (&Axes{}).AddAnchoredPacker(PackHorizontal, AnchoredPackerOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    5,
		Sep:      2,
		FrameOn:  boolPtr(false),
		Align:    PackAlignStart,
	})
	packer.AddImage(img, 2)
	packer.AddDrawingArea(4, 4).AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 4, Y: 4}},
	}, render.Paint{Stroke: stroke, LineWidth: 1})

	ctx := createTestDrawContext()
	scale := pointsToPixels(ctx.RC, 1)
	r := &textRecordingRenderer{}
	packer.Draw(r, ctx)

	if len(r.imageDsts) != 1 {
		t.Fatalf("packed image destinations = %+v, want one", r.imageDsts)
	}
	// Display space is y-up: Align Start aligns to the top edge (495).
	wantImageDst := geom.Rect{Min: geom.Pt{X: 5, Y: 495 - 6*scale}, Max: geom.Pt{X: 5 + 8*scale, Y: 495}}
	if !approxRect(r.imageDsts[0], wantImageDst, 1e-9) {
		t.Fatalf("packed image dst = %+v, want %+v", r.imageDsts[0], wantImageDst)
	}
	path := recordedStrokePath(r.pathCalls, stroke)
	if len(path.V) != 2 {
		t.Fatalf("packed drawing path not found in %+v", r.pathCalls)
	}
	if got, want := path.V[0], (geom.Pt{X: 5 + 8*scale + 2, Y: 495 - 4*scale}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("drawing child lower-left after image mapped to %+v, want %+v", got, want)
	}
}

func TestAnchoredPackerImageAndDrawingAreaScaleByDPI(t *testing.T) {
	stroke := render.Color{R: 0.8, G: 0.2, B: 0.1, A: 1}
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 3)))
	packer := (&Axes{}).AddAnchoredPacker(PackHorizontal, AnchoredPackerOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    0,
		Sep:      0,
		FrameOn:  boolPtr(false),
		Align:    PackAlignStart,
	})
	packer.AddImage(img, 2)
	packer.AddDrawingArea(4, 4).AddPath(geom.Path{
		C: []geom.Cmd{geom.MoveTo, geom.LineTo},
		V: []geom.Pt{{X: 0, Y: 0}, {X: 4, Y: 4}},
	}, render.Paint{Stroke: stroke, LineWidth: 1})

	ctx := createTestDrawContext()
	ctx.RC = style.Apply(ctx.RC, style.WithDPI(144))
	r := &textRecordingRenderer{}
	packer.Draw(r, ctx)

	// Display space is y-up: Align Start aligns to the top edge (500, inset 0).
	wantImageDst := geom.Rect{Min: geom.Pt{X: 0, Y: 500 - 12}, Max: geom.Pt{X: 16, Y: 500}}
	if len(r.imageDsts) != 1 || !approxRect(r.imageDsts[0], wantImageDst, 1e-9) {
		t.Fatalf("packed image destinations = %+v, want [%+v]", r.imageDsts, wantImageDst)
	}
	path := recordedStrokePath(r.pathCalls, stroke)
	if len(path.V) != 2 {
		t.Fatalf("packed drawing path not found in %+v", r.pathCalls)
	}
	if got, want := path.V[0], (geom.Pt{X: 16, Y: 500 - 8}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("drawing child lower-left after image mapped to %+v, want %+v", got, want)
	}
	if got, want := path.V[1], (geom.Pt{X: 24, Y: 500}); !pointsApprox(got, want, 1e-9) {
		t.Fatalf("drawing child upper-right after image mapped to %+v, want %+v", got, want)
	}
}

func TestAnchoredPackerImageDefaultsToMatplotlibAntialiasedInterpolation(t *testing.T) {
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 3)))
	packer := (&Axes{}).AddAnchoredPacker(PackHorizontal, AnchoredPackerOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    0,
		FrameOn:  boolPtr(false),
	})
	packer.AddImage(img, 1)

	r := &bboxImageInterpolationRenderer{}
	packer.Draw(r, createTestDrawContext())

	if !r.called {
		t.Fatal("DrawBboxImage was not called")
	}
	if r.interpolation != "antialiased" {
		t.Fatalf("packed image interpolation = %q, want Matplotlib OffsetImage default antialiased", r.interpolation)
	}
}

func TestAnchoredPackerImagePreservesExplicitInterpolation(t *testing.T) {
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 4, 3)))
	img.SetInterpolation("nearest")
	packer := (&Axes{}).AddAnchoredPacker(PackHorizontal, AnchoredPackerOptions{
		Location: LegendUpperLeft,
		Padding:  0,
		Inset:    0,
		FrameOn:  boolPtr(false),
	})
	packer.AddImage(img, 1)

	r := &bboxImageInterpolationRenderer{}
	packer.Draw(r, createTestDrawContext())

	if !r.called {
		t.Fatal("DrawBboxImage was not called")
	}
	if r.interpolation != "nearest" {
		t.Fatalf("packed image interpolation = %q, want explicit interpolation preserved", r.interpolation)
	}
}

func recordedPaintExists(calls []recordedPathCall, fill, stroke render.Color, lineWidth float64) bool {
	for _, call := range calls {
		if call.paint.Fill == fill && call.paint.Stroke == stroke && floatApprox(call.paint.LineWidth, lineWidth, 1e-12) {
			return true
		}
	}
	return false
}

func pointsApprox(got, want geom.Pt, tol float64) bool {
	return floatApprox(got.X, want.X, tol) && floatApprox(got.Y, want.Y, tol)
}

type anchoredTextMetricRenderer struct {
	render.NullRenderer
}

func (r *anchoredTextMetricRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	widths := map[string]float64{
		"anchored": 58.75,
		"text":     24.625,
	}
	width := widths[text]
	if width == 0 {
		width = float64(len(text)) * size * 0.5
	}
	return render.TextMetrics{W: width, H: 13, Ascent: 10, Descent: 3}
}

func recordedStrokePath(calls []recordedPathCall, stroke render.Color) geom.Path {
	for _, call := range calls {
		if call.paint.Stroke == stroke {
			return call.path
		}
	}
	return geom.Path{}
}

type bboxImageInterpolationRenderer struct {
	render.NullRenderer
	called        bool
	interpolation string
}

func (r *bboxImageInterpolationRenderer) DrawBboxImage(img render.Image, _ geom.Rect) bool {
	r.called = true
	r.interpolation = img.Interpolation()
	return true
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
