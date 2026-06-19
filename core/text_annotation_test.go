package core

import (
	"image"
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestAnnotationFontKeyOverridesRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "note",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		FontKey:  "Annotation Font",
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware annotation draw, got %+v", r.fontTextCalls)
	}
	if got := r.fontTextCalls[0].fontKey; got != "Annotation Font" {
		t.Fatalf("annotation fontKey = %q, want annotation override", got)
	}
}

func TestAnnotationFontPropertiesOverrideRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "note",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Sans Mono"},
			Style:    render.FontStyleOblique,
			Weight:   600,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware annotation draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Style != render.FontStyleOblique || props.Weight != 600 || len(props.Families) != 1 || props.Families[0] != "DejaVu Sans Mono" {
		t.Fatalf("annotation font properties = %+v, want DejaVu Sans Mono oblique 600", props)
	}
}

func TestAnnotationCanDisableMathParsing(t *testing.T) {
	ctx := createTestDrawContext()
	parseMath := false
	annotation := &Annotation{
		Point:     geom.Pt{X: 1, Y: 1},
		Content:   `note $\beta$`,
		OffsetX:   10,
		OffsetY:   -8,
		FontSize:  12,
		ParseMath: &parseMath,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one plain annotation draw, got %+v", r.fontTextCalls)
	}
	if got, want := r.fontTextCalls[0].text, `note $\beta$`; got != want {
		t.Fatalf("parse_math disabled annotation = %q, want %q", got, want)
	}
}

func TestAnnotationAlphaAppliesToTextAndArrow(t *testing.T) {
	ctx := createTestDrawContext()
	arrowStyle, _ := ArrowStyleFromString("-|>")
	connectionStyle, _ := ConnectionStyleFromString("arc3")
	annotation := &Annotation{
		Point:           geom.Pt{X: 1, Y: 1},
		Content:         "note",
		OffsetX:         10,
		OffsetY:         -8,
		FontSize:        12,
		Color:           render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.6},
		ArrowColor:      render.Color{R: 0.8, G: 0.1, B: 0.1, A: 0.8},
		ArrowWidth:      1.25,
		ArrowHeadSize:   8,
		ArrowStyle:      arrowStyle,
		ConnectionStyle: connectionStyle,
	}
	annotation.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.textColors) != 1 {
		t.Fatalf("expected one annotation text color, got %+v", r.textColors)
	}
	if !approx(r.textColors[0].A, 0.3, 1e-12) {
		t.Fatalf("annotation text alpha = %v, want local alpha multiplied by artist alpha", r.textColors[0].A)
	}
	if !hasPaintAlpha(r.pathPaints, 0.4) {
		t.Fatalf("annotation arrow paints should include artist-multiplied alpha 0.4, got %+v", r.pathPaints)
	}
}

func TestAnnotationClipSkipsOutsideAnnotatedPoint(t *testing.T) {
	ctx := createTestDrawContext()
	clip := true
	annotation := &Annotation{
		Point:          geom.Pt{X: 100, Y: 100},
		Content:        "outside",
		OffsetX:        10,
		OffsetY:        -8,
		FontSize:       12,
		AnnotationClip: &clip,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("clipped annotation should not draw text, got font=%+v text=%+v", r.fontTextCalls, r.texts)
	}
}

func TestAnnotationClipFalseDrawsOutsideAnnotatedPoint(t *testing.T) {
	ctx := createTestDrawContext()
	clip := false
	annotation := &Annotation{
		Point:          geom.Pt{X: 100, Y: 100},
		Content:        "outside",
		OffsetX:        10,
		OffsetY:        -8,
		FontSize:       12,
		AnnotationClip: &clip,
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("annotation_clip=false should draw text, got %+v", r.fontTextCalls)
	}
}

func TestAnnotationClipDefaultMatchesMatplotlibDataOnlyPolicy(t *testing.T) {
	ctx := createTestDrawContext()
	dataAnnotation := &Annotation{
		Point:    geom.Pt{X: 100, Y: 100},
		Content:  "outside data",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		Coords:   Coords(CoordData),
	}
	axesAnnotation := &Annotation{
		Point:    geom.Pt{X: 1.5, Y: 1.5},
		Content:  "outside axes",
		OffsetX:  10,
		OffsetY:  -8,
		FontSize: 12,
		Coords:   Coords(CoordAxes),
	}
	r := &fontAwareTextRecordingRenderer{}

	dataAnnotation.DrawOverlay(r, ctx)
	axesAnnotation.DrawOverlay(r, ctx)

	if containsFontTextCall(r.fontTextCalls, "outside data") || containsTextString(r.texts, "outside data") {
		t.Fatalf("annotation_clip default should clip outside data-coordinate annotations, got font=%+v text=%+v", r.fontTextCalls, r.texts)
	}
	if !containsFontTextCall(r.fontTextCalls, "outside axes") && !containsTextString(r.texts, "outside axes") {
		t.Fatalf("annotation_clip default should draw outside non-data annotations, got font=%+v text=%+v", r.fontTextCalls, r.texts)
	}
}

func TestAnnotationDrawOverlayRendersArrowAndText(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Annotate("peak", 0.5, 0.5)

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 1 || r.texts[0] != "peak" {
		t.Fatalf("unexpected annotation texts: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected annotation arrow line and head, got %d paths", r.pathCount)
	}
}

func TestAnnotationDrawsTextBBox(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	face := render.Color{R: 1, G: 0.95, B: 0.8, A: 1}
	edge := render.Color{R: 0.2, G: 0.1, B: 0.05, A: 1}
	ax.Annotate("boxed", 0.5, 0.5, AnnotationOptions{
		OffsetX: 10,
		OffsetY: -8,
		BBox: &TextBBoxOptions{
			FaceColor:    face,
			EdgeColor:    edge,
			LineWidth:    2,
			Padding:      3,
			CornerRadius: 4,
		},
	})

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if !containsTextString(r.texts, "boxed") {
		t.Fatalf("expected annotation text to draw, got %v", r.texts)
	}
	if !hasPathPaint(r.pathPaints, face, edge, 2) {
		t.Fatalf("annotation bbox paint not found in %+v", r.pathPaints)
	}
}

func TestAnnotationDefaultAlignmentMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	ann := ax.Annotate("label", 0.5, 0.5, AnnotationOptions{
		OffsetX: -10,
		OffsetY: -8,
	})

	if ann.HAlign != TextAlignLeft || ann.VAlign != TextVAlignBaseline {
		t.Fatalf("default annotation alignment = (%v, %v), want left/baseline", ann.HAlign, ann.VAlign)
	}
}

func TestAnnotationArrowHeadSizeUsesPointMutationScale(t *testing.T) {
	ctx := createTestDrawContext()
	arrow, _ := ArrowStyleFromString("-|>,head_length=0.35,head_width=0.20")
	annotation := &Annotation{
		ArrowWidth:      1.2,
		ArrowHeadSize:   9,
		ArrowStyle:      arrow,
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}

	annotation.drawArrow(r, ctx, geom.Pt{X: 100, Y: 100}, geom.Pt{X: 200, Y: 100})

	if len(r.pathCalls) < 2 {
		t.Fatalf("expected line and arrow head paths, got %+v", r.pathCalls)
	}
	headBounds, ok := pathBounds(r.pathCalls[len(r.pathCalls)-1].path)
	if !ok {
		t.Fatalf("arrow head path has no bounds: %+v", r.pathCalls[len(r.pathCalls)-1].path)
	}
	scale := pointsToPixels(ctx.RC, 9)
	wantWidth := 0.35 * scale
	wantHeight := 2 * 0.20 * scale
	if !approx(headBounds.Max.X-headBounds.Min.X, wantWidth, 1e-9) ||
		!approx(headBounds.Max.Y-headBounds.Min.Y, wantHeight, 1e-9) {
		t.Fatalf("arrow head bounds = %+v, want width %.12g height %.12g from point mutation scale", headBounds, wantWidth, wantHeight)
	}
}

func TestAnnotationArrowDefaultShrinkUsesPointUnits(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}
	start := geom.Pt{X: 100, Y: 100}
	target := geom.Pt{X: 200, Y: 100}

	annotation.drawArrow(r, ctx, start, target)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one plain connection path, got %+v", r.pathCalls)
	}
	path := r.pathCalls[0].path
	if len(path.V) != 3 {
		t.Fatalf("connection path vertices = %+v, want quadratic path", path.V)
	}
	shrink := pointsToPixels(ctx.RC, 2)
	if !approx(path.V[0].X, start.X+shrink, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("shrunk start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + shrink, Y: start.Y})
	}
	if !approx(path.V[2].X, target.X-shrink, 1e-9) || !approx(path.V[2].Y, target.Y, 1e-9) {
		t.Fatalf("shrunk end = %+v, want %+v", path.V[2], geom.Pt{X: target.X - shrink, Y: target.Y})
	}
}

func TestAnnotationArrowStartsFromTextBoxRelposBeforePatchClip(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:           geom.Pt{X: 0.2, Y: 0.8},
		Content:         "box",
		OffsetX:         120,
		OffsetY:         80,
		FontSize:        10,
		Coords:          Coords(CoordFigure),
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
		HAlign:          TextAlignCenter,
		VAlign:          TextVAlignMiddle,
		BBox: &TextBBoxOptions{
			Padding:   4,
			LineWidth: 1,
		},
	}
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected annotation arrow path")
	}
	target := transformedPoint(ctx, annotation.Coords, annotation.Point, 0, 0)
	anchor := transformedPoint(ctx, annotation.Coords, annotation.Point, annotation.OffsetX, annotation.OffsetY)
	layout := measureSingleLineTextLayout(r, annotation.Content, annotation.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	origin := alignedSingleLineOrigin(anchor, layout, annotation.HAlign, layoutVerticalAlign(annotation.VAlign, false))
	box, ok := textBBoxRect(origin, layout, annotation.BBox, ctx, annotation.FontSize)
	if !ok {
		t.Fatal("expected annotation text bbox")
	}
	raw := annotation.ConnectionStyle.connect(rectCenter(box), target, 0, 0)
	boundary, ok := connectionPatchBoundaryPoint(raw, pixelRectPath(box).Interpolated(8).V, true)
	if !ok {
		t.Fatalf("expected center-to-target path to leave bbox: box=%+v path=%+v", box, raw)
	}
	raw.V[0] = boundary
	want := shrinkPathEndpoints(raw, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	got := r.pathCalls[0].path
	if len(got.V) != len(want.V) || distance(got.V[0], want.V[0]) > 1e-9 {
		t.Fatalf("annotation arrow start = %+v, want clipped relpos start %+v (box=%+v target=%+v)", got.V, want.V, box, target)
	}
}

func TestAnnotationAngleUsesRotatedTextDrawer(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
		OffsetX:  10,
		FontSize: 10,
		Angle:    35,
		Coords:   Coords(CoordData),
	}
	r := &fontAwareTextRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated annotation text draw, got %+v", r.fontRotatedCalls)
	}
	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("rotated annotation should not use unrotated text draws, font=%+v legacy=%+v", r.fontTextCalls, r.texts)
	}
}

func TestAnnotationMultilineSplitsTextDraws(t *testing.T) {
	ctx := createTestDrawContext()
	annotation := &Annotation{
		Point:    geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		OffsetX:  10,
		FontSize: 10,
		Coords:   Coords(CoordData),
	}
	annotation.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	annotation.DrawOverlay(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "top" || r.texts[1] != "bottom" {
		t.Fatalf("annotation multiline draws = %v, want [top bottom]", r.texts)
	}
	for i, col := range r.textColors {
		if !approx(col.A, 0.5, 1e-12) {
			t.Fatalf("annotation multiline text alpha[%d] = %g, want 0.5", i, col.A)
		}
	}
}

func TestMultilineBaselineAlignsLastLineLikeMatplotlib(t *testing.T) {
	ctx := createTestDrawContext()
	var r textRecordingRenderer
	anchor := geom.Pt{X: 100, Y: 200}

	block, ok := measureMultilineTextBlock(&r, ctx, anchor, 10, "", false, false,
		[]string{"top", "bottom"}, 0, TextAlignLeft, textLayoutVAlignBaseline)

	if !ok {
		t.Fatal("measureMultilineTextBlock returned !ok")
	}
	if len(block.BaselineYs) != 2 {
		t.Fatalf("baseline count = %d, want 2", len(block.BaselineYs))
	}
	if !approx(block.BaselineYs[1], anchor.Y, 1e-9) {
		t.Fatalf("last multiline baseline = %v, want anchor y %v", block.BaselineYs[1], anchor.Y)
	}
	if block.BaselineYs[0] <= anchor.Y {
		t.Fatalf("first multiline baseline = %v, want above anchor y %v", block.BaselineYs[0], anchor.Y)
	}
}

func TestAnnotateRespectsConfiguredCoordinateSpaces(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Annotate("data", 0.25, 0.75, AnnotationOptions{
		Coords:  Coords(CoordData),
		OffsetX: 10,
		OffsetY: -15,
	})
	ax.Annotate("axes", 0.5, 0.5, AnnotationOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: -12,
		OffsetY: 6,
	})
	ax.Annotate("figure", 0.2, 0.3, AnnotationOptions{
		Coords:  Coords(CoordFigure),
		OffsetX: 7,
		OffsetY: 4,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	dataToPx := ctx.TransformFor(Coords(CoordData))
	axesToPx := ctx.TransformFor(Coords(CoordAxes))
	figureToPx := ctx.TransformFor(Coords(CoordFigure))
	if dataToPx == nil || axesToPx == nil || figureToPx == nil {
		t.Fatalf("unexpected nil transform from coordinate-space transform helpers")
	}

	dataTarget := dataToPx.Apply(geom.Pt{X: 0.25, Y: 0.75})
	dataAnchor := transform.NewOffset(dataToPx, geom.Pt{X: 10, Y: -15}).Apply(geom.Pt{X: 0.25, Y: 0.75})
	axesTarget := axesToPx.Apply(geom.Pt{X: 0.5, Y: 0.5})
	axesAnchor := transform.NewOffset(axesToPx, geom.Pt{X: -12, Y: 6}).Apply(geom.Pt{X: 0.5, Y: 0.5})
	figureTarget := figureToPx.Apply(geom.Pt{X: 0.2, Y: 0.3})
	figureAnchor := transform.NewOffset(figureToPx, geom.Pt{X: 7, Y: 4}).Apply(geom.Pt{X: 0.2, Y: 0.3})

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	var connections []geom.Path
	for _, path := range r.pathCalls {
		if len(path.path.C) == 2 && path.path.C[0] == geom.MoveTo && (path.path.C[1] == geom.LineTo || path.path.C[1] == geom.QuadTo) && len(path.path.V) >= 2 {
			connections = append(connections, path.path)
		}
	}
	if len(connections) != 3 {
		t.Fatalf("expected 3 annotation connection paths, got %d", len(connections))
	}

	expectConnection := func(got geom.Path, anchor, target geom.Pt) {
		if len(got.V) < 2 {
			t.Fatalf("annotation connection path vertices = %d, want at least 2", len(got.V))
		}
		end := got.V[len(got.V)-1]
		if math.Hypot(end.X-target.X, end.Y-target.Y) > 12 {
			t.Fatalf("annotation connection end = %+v, want near target %+v", end, target)
		}
		if math.Hypot(got.V[0].X-anchor.X, got.V[0].Y-anchor.Y) > 40 {
			t.Fatalf("annotation connection start = %+v, expected near label anchor %+v", got.V[0], anchor)
		}
		if !containsPathPointNearForTextTest(r.pathCalls, target, 12) {
			t.Fatalf("annotation arrow head should land near target %+v, got paths %+v", target, r.pathCalls)
		}
	}

	expectConnection(connections[0], dataAnchor, dataTarget)
	expectConnection(connections[1], axesAnchor, axesTarget)
	expectConnection(connections[2], figureAnchor, figureTarget)
}

func TestAnnotateSupportsSeparateTextCoordinateSpace(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false
	textPos := geom.Pt{X: 0.82, Y: 0.18}

	ax.Annotate("mixed", 0.25, 0.75, AnnotationOptions{
		Coords:       Coords(CoordData),
		TextPosition: &textPos,
		TextCoords:   Coords(CoordAxes),
		OffsetX:      6,
		OffsetY:      -4,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	target := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 0.25, Y: 0.75})
	anchor := transform.NewOffset(ctx.TransformFor(Coords(CoordAxes)), geom.Pt{X: 6, Y: -4}).Apply(textPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if len(r.origins) != 1 {
		t.Fatalf("annotation text origins = %+v, want one text draw", r.origins)
	}
	if math.Hypot(r.origins[0].X-anchor.X, r.origins[0].Y-anchor.Y) > 40 {
		t.Fatalf("annotation text origin = %+v, want near axes-coordinate anchor %+v", r.origins[0], anchor)
	}
	if !containsPathPointNearForTextTest(r.pathCalls, target, 12) {
		t.Fatalf("annotation arrow should land near data-coordinate target %+v, got paths %+v", target, r.pathCalls)
	}
}

func TestAnnotationBboxDrawsTextFrameAndArrow(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	boxPos := geom.Pt{X: 0.7, Y: 0.25}
	align := geom.Pt{X: 0.5, Y: 0.5}
	frameOn := true
	frameFill := render.Color{R: 1, G: 1, B: 1, A: 1}
	frameEdge := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	textColor := render.Color{R: 0.9, G: 0.1, B: 0.2, A: 1}
	arrowColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}

	if got := ax.AnnotationBbox("box", 0.25, 0.75, AnnotationBboxOptions{
		XYCoords:      Coords(CoordData),
		BoxCoords:     Coords(CoordAxes),
		BoxPosition:   &boxPos,
		BoxAlignment:  &align,
		FrameOn:       &frameOn,
		Padding:       4,
		FaceColor:     frameFill,
		EdgeColor:     frameEdge,
		LineWidth:     1.5,
		TextColor:     textColor,
		Arrow:         true,
		ArrowColor:    arrowColor,
		ArrowWidth:    1.25,
		ArrowHeadSize: 8,
	}); got == nil {
		t.Fatal("AnnotationBbox returned nil")
	}

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	target := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 0.25, Y: 0.75})
	boxAnchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if len(r.texts) != 1 || r.texts[0] != "box" {
		t.Fatalf("unexpected annotation-box texts: %v", r.texts)
	}
	layout := measureSingleLineTextLayout(r, "box", resolvedFontSize(0, ctx), ctx.RC.FontKey, ctx.RC.UseTeX)
	wantOrigin := alignedSingleLineOrigin(boxAnchor, layout, TextAlignCenter, textLayoutVAlignCenter)
	if !approx(r.origins[0].X, wantOrigin.X, 1e-9) || !approx(r.origins[0].Y, wantOrigin.Y, 1e-9) {
		t.Fatalf("annotation-box text origin = %+v, want %+v", r.origins[0], wantOrigin)
	}
	if !hasPathPaint(r.pathPaints, frameFill, frameEdge, 1.5) {
		t.Fatalf("annotation-box frame paint not found in %+v", r.pathPaints)
	}

	foundArrowToTarget := false
	for _, call := range r.pathCalls {
		if pathHasPointNearForTextTest(call.path, target, 12) {
			foundArrowToTarget = true
			break
		}
	}
	if !foundArrowToTarget {
		t.Fatalf("annotation-box arrow should land near annotated point %+v, got paths %+v", target, r.pathCalls)
	}
}

func TestAnnotationBboxDefaultWidthsMatchMatplotlibPoints(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())

	box := ax.AnnotationBbox("box", 0.25, 0.75, AnnotationBboxOptions{Arrow: true})
	if box == nil {
		t.Fatal("AnnotationBbox returned nil")
	}
	want := pointsToPixels(fig.RC, 1)
	if box.LineWidth != want {
		t.Fatalf("annotation-box linewidth = %v, want Matplotlib 1 pt = %v px", box.LineWidth, want)
	}
	if box.ArrowWidth != want {
		t.Fatalf("annotation-box arrow linewidth = %v, want Matplotlib 1 pt = %v px", box.ArrowWidth, want)
	}
}

func TestAnnotationBboxArrowStartsFromBoxRelposBeforePatchClip(t *testing.T) {
	ctx := createTestDrawContext()
	boxPos := geom.Pt{X: 0.44, Y: 0.64}
	box := &AnnotationBbox{
		Point:           geom.Pt{X: 0.2, Y: 0.8},
		Content:         "box",
		XYCoords:        Coords(CoordFigure),
		BoxCoords:       Coords(CoordFigure),
		BoxPosition:     &boxPos,
		BoxAlignment:    geom.Pt{X: 0.5, Y: 0.5},
		FrameOn:         true,
		Padding:         4,
		FontSize:        10,
		Arrow:           true,
		ArrowWidth:      1,
		ArrowHeadSize:   9,
		ArrowStyle:      ArrowStyle{Name: "-"},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
	}
	r := &textRecordingRenderer{}

	box.DrawOverlay(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected annotation-box arrow path")
	}
	target := transformedPoint(ctx, box.XYCoords, box.Point, 0, 0)
	layout := measureSingleLineTextLayout(r, box.Content, box.FontSize, ctx.RC.FontKey, ctx.RC.UseTeX)
	boxAnchor := box.boxAnchor(ctx)
	contentBox := annotationBoxRect(boxAnchor, box.contentSize(layout, ctx), box.BoxAlignment)
	frame := expandAnchoredRect(contentBox, box.resolvedPadding(box.FontSize, ctx))
	raw := box.ConnectionStyle.connect(rectCenter(frame), target, 0, 0)
	boundary, ok := connectionPatchBoundaryPoint(raw, pixelRectPath(frame).Interpolated(8).V, true)
	if !ok {
		t.Fatalf("expected center-to-target path to leave annotation box: box=%+v path=%+v", frame, raw)
	}
	raw.V[0] = boundary
	want := shrinkPathEndpoints(raw, arrowShrinkPixels(ctx, 2), arrowShrinkPixels(ctx, 2))
	got := r.pathCalls[0].path
	if len(got.V) != len(want.V) || distance(got.V[0], want.V[0]) > 1e-9 {
		t.Fatalf("annotation-box arrow start = %+v, want clipped relpos start %+v (box=%+v target=%+v)", got.V, want.V, frame, target)
	}
}

func TestAnnotationBboxDrawsImageContent(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	align := geom.Pt{X: 0, Y: 1}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:    Coords(CoordAxes),
		BoxPosition:  &boxPos,
		BoxAlignment: &align,
		Image:        img,
		ImageZoom:    2,
		Padding:      3,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	anchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	if got := len(r.imageDsts); got != 1 {
		t.Fatalf("annotation image draw count = %d, want 1", got)
	}
	scale := pointsToPixels(fig.RC, 1)
	wantDst := geom.Rect{
		Min: anchor,
		Max: geom.Pt{X: anchor.X + 20*scale, Y: anchor.Y + 12*scale},
	}
	if !approxRect(r.imageDsts[0], wantDst, 1e-9) {
		t.Fatalf("annotation image dst = %+v, want %+v", r.imageDsts[0], wantDst)
	}
}

func TestAnnotationBboxImageDefaultsToMatplotlibAntialiasedInterpolation(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:   Coords(CoordAxes),
		BoxPosition: &boxPos,
		Image:       img,
	})

	r := &bboxImageInterpolationRenderer{}
	DrawFigure(fig, r)

	if !r.called {
		t.Fatal("DrawBboxImage was not called")
	}
	if r.interpolation != "antialiased" {
		t.Fatalf("annotation image interpolation = %q, want Matplotlib OffsetImage default antialiased", r.interpolation)
	}
}

func TestAnnotationBboxImagePreservesExplicitInterpolation(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(unitRect())
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	img.SetInterpolation("nearest")
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:   Coords(CoordAxes),
		BoxPosition: &boxPos,
		Image:       img,
	})

	r := &bboxImageInterpolationRenderer{}
	DrawFigure(fig, r)

	if !r.called {
		t.Fatal("DrawBboxImage was not called")
	}
	if r.interpolation != "nearest" {
		t.Fatalf("annotation image interpolation = %q, want explicit interpolation preserved", r.interpolation)
	}
}

func TestAnnotationBboxImageZoomScalesByDPI(t *testing.T) {
	fig := NewFigure(800, 600)
	fig.RC = style.Apply(fig.RC, style.WithDPI(144))
	ax := fig.AddAxes(unitRect())
	img := render.NewImageData(image.NewRGBA(image.Rect(0, 0, 10, 6)))
	boxPos := geom.Pt{X: 0.4, Y: 0.6}
	align := geom.Pt{X: 0, Y: 1}
	ax.AnnotationBbox("", 0.1, 0.2, AnnotationBboxOptions{
		BoxCoords:    Coords(CoordAxes),
		BoxPosition:  &boxPos,
		BoxAlignment: &align,
		Image:        img,
		ImageZoom:    2,
		Padding:      0,
	})

	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	anchor := ctx.TransformFor(Coords(CoordAxes)).Apply(boxPos)

	r := &textRecordingRenderer{}
	DrawFigure(fig, r)

	wantDst := geom.Rect{
		Min: anchor,
		Max: geom.Pt{X: anchor.X + 40, Y: anchor.Y + 24},
	}
	if len(r.imageDsts) != 1 || !approxRect(r.imageDsts[0], wantDst, 1e-9) {
		t.Fatalf("annotation image dst = %+v, want [%+v]", r.imageDsts, wantDst)
	}
}
