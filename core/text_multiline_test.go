package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRotatedMultilineTextUsesMatplotlibLayoutOffsets(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "rotation\nmode",
		FontSize: 10,
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		Angle:    -32,
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 2 {
		t.Fatalf("rotated multiline calls = %+v, want two lines", r.fontRotatedCalls)
	}
	lines := []string{"rotation", "mode"}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	angle := text.Angle * math.Pi / 180
	wantOffsets := matplotlibMultilineOffsetsForTest(r, lines, text.FontSize, text.FontKey, text.Linespacing, text.HAlign, text.VAlign, angle)
	for i, call := range r.fontRotatedCalls {
		layout := measureSingleLineTextLayoutParseMath(r, call.text, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
		gotOrigin := rotatedTextBaselineOriginForTest(call.anchor, layout, angle)
		got := geom.Pt{X: gotOrigin.X - anchor.X, Y: gotOrigin.Y - anchor.Y}
		if !approx(got.X, wantOffsets[i].X, 1e-9) || !approx(got.Y, wantOffsets[i].Y, 1e-9) {
			t.Fatalf("line %q baseline offset = %+v, want Matplotlib %+v; all calls=%+v", call.text, got, wantOffsets[i], r.fontRotatedCalls)
		}
	}
}

func TestMultilineTextSplitsDrawsAndUsesBlockBBox(t *testing.T) {
	fig := NewFigure(100, 100)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})
	ax.XAxis.ShowSpine = false
	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowSpine = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false
	ax.ShowFrame = false

	ax.Text(0.1, 0.9, "top\nbottom", TextOptions{
		Coords: Coords(CoordAxes),
		VAlign: TextVAlignTop,
		BBox: optional.Of(TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		}),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected two text draws, got %d: %v", len(r.texts), r.texts)
	}
	if r.texts[0] != "top" || r.texts[1] != "bottom" {
		t.Fatalf("unexpected multiline draw order: %v", r.texts)
	}
	if r.pathCount != 1 {
		t.Fatalf("expected one multiline bbox path, got %d", r.pathCount)
	}
	// Display space is y-up: the second line sits below the first at a smaller Y.
	if !(r.origins[1].Y < r.origins[0].Y) {
		t.Fatalf("expected second line below first, got origins %v", r.origins)
	}
}

func TestMultilineTextPathEffectsUseGlyphPaths(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		ClipOn:   true,
		PathEffects: []render.PathEffect{
			render.StrokePathEffect(render.Color{R: 1, A: 1}, 2, geom.Pt{}),
			render.NormalPathEffect(),
		},
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 0 {
		t.Fatalf("path-effect multiline text should draw glyph paths, got text draws %v", r.texts)
	}
	if len(r.pathPaints) < 2 {
		t.Fatalf("expected one effect-painted glyph path per line, got %d path paints", len(r.pathPaints))
	}
	for i, paint := range r.pathPaints {
		if len(paint.PathEffects) != 2 {
			t.Fatalf("path paint %d effects = %+v, want stroke + normal effects", i, paint.PathEffects)
		}
	}
}

func TestRotatedMultilineTextBBoxRotatesWithText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		Angle:    45,
		ClipOn:   true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 2 {
		t.Fatalf("expected two rotated multiline text draws, got %+v", r.fontRotatedCalls)
	}
	if len(r.pathCalls) == 0 {
		t.Fatal("expected rotated multiline text bbox path")
	}
	path := r.pathCalls[0].path
	if len(path.V) < 4 {
		t.Fatalf("bbox path vertices = %+v, want rotated rectangle vertices", path.V)
	}
	if approx(path.V[0].Y, path.V[1].Y, 1e-9) || approx(path.V[1].X, path.V[2].X, 1e-9) {
		t.Fatalf("rotated multiline text bbox remained axis-aligned: %+v", path.V[:4])
	}
}

func TestTextMultiAlignmentControlsLineAlignmentWithinBlock(t *testing.T) {
	ctx := createTestDrawContext()
	multiAlign := TextAlignLeft
	text := &Text{
		Position:       geom.Pt{X: 1, Y: 1},
		Content:        "narrow\nmuch wider",
		FontSize:       10,
		HAlign:         TextAlignRight,
		VAlign:         TextVAlignTop,
		MultiAlignment: &multiAlign,
		ClipOn:         true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("expected two multiline origins, got %d: %+v", len(r.origins), r.origins)
	}
	if !approx(r.origins[0].X, r.origins[1].X, 1e-12) {
		t.Fatalf("left multialignment should keep line origins equal inside right-aligned block, got %+v", r.origins)
	}
	blockRight := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY).X
	if !approx(r.origins[1].X+float64(len("much wider"))*text.FontSize*0.5, blockRight, 1e-12) {
		t.Fatalf("right-aligned multiline block no longer ends at anchor: origins=%+v anchorX=%v", r.origins, blockRight)
	}
}

func TestMultilineTextLinespacingControlsBaselineAdvance(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:    geom.Pt{X: 1, Y: 1},
		Content:     "first\nsecond",
		FontSize:    10,
		Linespacing: 2,
		ClipOn:      true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := text.FontSize*0.2 + text.FontSize*0.8*text.Linespacing
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("multiline baseline advance = %v, want %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextNormalLinespacingMatchesMatplotlibBaselineAdvance(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "first\nsecond",
		FontSize: 10,
		ClipOn:   true,
	}
	r := &fontMetricTextRecordingRenderer{
		fontHeights: render.FontHeightMetrics{Ascent: 12, Descent: 4, LineGap: 3},
	}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := r.fontHeights.Descent + r.fontHeights.Ascent*1.2
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("normal multiline baseline advance = %v, want Matplotlib descent + ascent*linespacing %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextNumericLinespacingUsesFontHeight(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:    geom.Pt{X: 1, Y: 1},
		Content:     "first\nsecond",
		FontSize:    10,
		Linespacing: 1.5,
		ClipOn:      true,
	}
	r := &fontMetricTextRecordingRenderer{
		fontHeights: render.FontHeightMetrics{Ascent: 12, Descent: 4, LineGap: 3},
	}

	text.Draw(r, ctx)

	if len(r.origins) != 2 {
		t.Fatalf("multiline draw origins = %d, want 2", len(r.origins))
	}
	wantAdvance := r.fontHeights.Descent + text.Linespacing*r.fontHeights.Ascent
	gotAdvance := r.origins[0].Y - r.origins[1].Y // y-up: next line is below at smaller Y
	if !approx(gotAdvance, wantAdvance, 1e-9) {
		t.Fatalf("numeric multiline baseline advance = %v, want Matplotlib descent + linespacing*ascent %v", gotAdvance, wantAdvance)
	}
}

func TestMultilineTextAngleUsesRotatedTextDrawer(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "top\nbottom",
		FontSize: 10,
		Angle:    30,
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 2 {
		t.Fatalf("expected two rotated multiline text draws, got %+v", r.fontRotatedCalls)
	}
	if len(r.fontTextCalls) != 0 || len(r.texts) != 0 {
		t.Fatalf("multiline angle should not use unrotated text draws, font=%+v legacy=%+v", r.fontTextCalls, r.texts)
	}
}
