package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestTextBBoxFractionalPaddingScalesByFontSize(t *testing.T) {
	ctx := createTestDrawContext()
	got := resolvedTextBBoxOptions(TextBBoxOptions{
		FaceColor: render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		Padding:   0.3,
	}, ctx, 12)
	want := pointsToPixels(ctx.RC, 0.3*12)
	if !approx(got.Padding, want, 1e-12) {
		t.Fatalf("fractional bbox padding = %v, want %v", got.Padding, want)
	}

	pixel := resolvedTextBBoxOptions(TextBBoxOptions{
		FaceColor: render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		Padding:   3,
	}, ctx, 12)
	if !approx(pixel.Padding, 3, 1e-12) {
		t.Fatalf("pixel bbox padding = %v, want 3", pixel.Padding)
	}
}

func TestRotatedTextBBoxRotatesWithText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
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

	if len(r.pathCalls) == 0 {
		t.Fatal("expected rotated text bbox path")
	}
	path := r.pathCalls[0].path
	if len(path.V) < 4 {
		t.Fatalf("bbox path vertices = %+v, want rotated rectangle vertices", path.V)
	}
	if approx(path.V[0].Y, path.V[1].Y, 1e-9) || approx(path.V[1].X, path.V[2].X, 1e-9) {
		t.Fatalf("rotated text bbox remained axis-aligned: %+v", path.V[:4])
	}
}

func TestRotatedTextBBoxUsesDisplayRotationSign(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
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

	if len(r.pathCalls) == 0 || len(r.pathCalls[0].path.V) < 2 {
		t.Fatalf("expected rotated bbox path, got %+v", r.pathCalls)
	}
	edge := geom.Pt{
		X: r.pathCalls[0].path.V[1].X - r.pathCalls[0].path.V[0].X,
		Y: r.pathCalls[0].path.V[1].Y - r.pathCalls[0].path.V[0].Y,
	}
	if edge.X <= 0 || edge.Y <= 0 {
		t.Fatalf("positive text rotation should tilt bbox upward in y-up display coordinates, edge=%+v path=%+v", edge, r.pathCalls[0].path.V[:4])
	}
}

func TestRotatedTextBBoxUsesDefaultRotationDrawOrigin(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "tilt",
		FontSize: 10,
		HAlign:   TextAlignCenter,
		VAlign:   TextVAlignMiddle,
		Angle:    -28,
		ClipOn:   true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   2,
		},
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.pathCalls) == 0 {
		t.Fatal("expected rotated text bbox path")
	}
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	hAlign, vAlign := textRotationLayoutAlignments(text.HAlign, text.VAlign, text.Angle, text.RotationMode)
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	angle := text.Angle * math.Pi / 180
	drawOrigin := tickLabelDrawOriginFromP(anchor, layout, hAlign, vAlign, angle, false)
	want, ok := matplotlibRotatedTextBBoxPathForTest(anchor, drawOrigin, layout, text.BBox, ctx, text.FontSize, text.Angle)
	if !ok {
		t.Fatal("matplotlibRotatedTextBBoxPathForTest returned !ok")
	}

	got := r.pathCalls[0].path
	if len(got.V) < 4 || len(want.V) < 4 {
		t.Fatalf("bbox paths too short: got=%+v want=%+v", got.V, want.V)
	}
	for i := 0; i < 4; i++ {
		if !approx(got.V[i].X, want.V[i].X, 1e-9) || !approx(got.V[i].Y, want.V[i].Y, 1e-9) {
			t.Fatalf("rotated bbox vertex %d = %+v, want %+v; got path=%+v want path=%+v", i, got.V[i], want.V[i], got.V[:4], want.V[:4])
		}
	}
}

func TestRotatedTextBBoxMatchesMatplotlibTextAnnotationMatrixLabel(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.DPI = 100
	layout := singleLineTextLayout{
		TextLineLayout: render.TextLineLayout{
			Width:   105.875,
			Ascent:  14,
			Descent: 4,
			Height:  18,
		},
	}
	anchor := geom.Pt{X: 289.8, Y: 341.964}
	angleDeg := -28.0
	angle := angleDeg * math.Pi / 180
	drawOrigin := tickLabelDrawOriginFromP(anchor, layout, TextAlignCenter, textLayoutVAlignCenter, angle, false)
	path, ok := rotatedTextBBoxPath(anchor, drawOrigin, layout, &TextBBoxOptions{
		Padding: pointsToPixels(ctx.RC, 12*0.25),
	}, ctx, 12, angleDeg)
	if !ok {
		t.Fatal("rotatedTextBBoxPath returned !ok")
	}
	got, ok := pathBounds(path)
	if !ok {
		t.Fatalf("rotated bbox path has no bounds: %+v", path)
	}

	// Matplotlib 3.10.9 text_annotation_matrix rotated label bbox patch extent:
	// (233.1986379227708, 303.5297409941049, 113.20272415445842, 76.86851801179034)
	// in y-up display coordinates.
	want := geom.Rect{
		Min: geom.Pt{X: 233.1986379227708, Y: 303.5297409941049},
		Max: geom.Pt{X: 346.40136207722924, Y: 380.3982590058952},
	}
	if !approxRect(got, want, 1e-9) {
		t.Fatalf("rotated bbox bounds = %+v, want Matplotlib %+v; path=%+v", got, want, path.V)
	}
	wantVertices := []geom.Pt{
		{X: 233.1986379227708, Y: 357.14730572727683},
		{X: 334.0386109238674, Y: 303.5297409941048},
		{X: 346.40136207722924, Y: 326.7806942727233},
		{X: 245.56138907613257, Y: 380.3982590058952},
	}
	if len(path.V) < len(wantVertices) {
		t.Fatalf("rotated bbox path has %d vertices, want at least %d: %+v", len(path.V), len(wantVertices), path.V)
	}
	for i, want := range wantVertices {
		if !approx(path.V[i].X, want.X, 1e-9) || !approx(path.V[i].Y, want.Y, 1e-9) {
			t.Fatalf("rotated bbox vertex %d = %+v, want Matplotlib %+v; path=%+v", i, path.V[i], want, path.V)
		}
	}
}

func TestTextBBoxDrawsBehindAxesAndFigureText(t *testing.T) {
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

	box := TextBBoxOptions{
		FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
		EdgeColor: render.Color{R: 0.7, G: 0.7, B: 0.7, A: 1},
	}
	ax.Text(0.02, 0.98, "axes note", TextOptions{
		Coords: Coords(CoordAxes),
		HAlign: TextAlignLeft,
		VAlign: TextVAlignTop,
		BBox:   optional.Of(box),
	})
	fig.Text(0.98, 0.02, "figure note", TextOptions{
		HAlign: TextAlignRight,
		VAlign: TextVAlignBottom,
		BBox:   optional.Of(box),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if !containsTextString(r.texts, "axes note") || !containsTextString(r.texts, "figure note") {
		t.Fatalf("missing bbox text draws: %v", r.texts)
	}
	if r.pathCount < 2 {
		t.Fatalf("expected text bbox paths, got %d", r.pathCount)
	}
	if len(r.pathPaints) < 2 || r.pathPaints[0].Fill.A == 0 || r.pathPaints[0].Stroke.A == 0 {
		t.Fatalf("expected visible bbox fill and stroke, got %+v", r.pathPaints)
	}
}

func TestParseBoxStyleSpec(t *testing.T) {
	spec, err := parseBoxStyleSpec("round, pad=0.3, rounding_size=0.2")
	if err != nil {
		t.Fatalf("parse round spec: %v", err)
	}
	if spec.style != BoxStyleRound {
		t.Fatalf("style = %v, want BoxStyleRound", spec.style)
	}
	if !spec.hasPad || !approx(spec.pad, 0.3, 1e-12) {
		t.Fatalf("pad = %v (has=%v), want 0.3", spec.pad, spec.hasPad)
	}
	if !spec.hasRounding || !approx(spec.roundingSize, 0.2, 1e-12) {
		t.Fatalf("rounding_size = %v (has=%v), want 0.2", spec.roundingSize, spec.hasRounding)
	}

	saw, err := parseBoxStyleSpec("sawtooth,tooth_size=0.1")
	if err != nil {
		t.Fatalf("parse sawtooth spec: %v", err)
	}
	if saw.style != BoxStyleSawtooth || !saw.hasTooth || !approx(saw.toothSize, 0.1, 1e-12) {
		t.Fatalf("sawtooth spec = %+v", saw)
	}
	if saw.hasPad {
		t.Fatalf("sawtooth spec should not carry pad: %+v", saw)
	}

	for _, bad := range []string{"nope", "round,bogus=1", "round,pad=abc", "round,pad"} {
		if _, err := parseBoxStyleSpec(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestResolvedTextBBoxOptionsAppliesStyleSpec(t *testing.T) {
	ctx := createTestDrawContext()
	got := resolvedTextBBoxOptions(TextBBoxOptions{Style: "sawtooth,pad=0.3,tooth_size=0.1"}, ctx, 12)
	if got.BoxStyle != BoxStyleSawtooth {
		t.Fatalf("BoxStyle = %v, want BoxStyleSawtooth", got.BoxStyle)
	}
	if want := pointsToPixels(ctx.RC, 0.1*12); !approx(got.ToothSize, want, 1e-9) {
		t.Fatalf("ToothSize = %v, want %v (fraction scaled to px)", got.ToothSize, want)
	}
	if want := pointsToPixels(ctx.RC, 0.3*12); !approx(got.Padding, want, 1e-9) {
		t.Fatalf("Padding = %v, want %v", got.Padding, want)
	}
}

func TestResolvedTextBBoxOptionsFallsBackOnBadStyle(t *testing.T) {
	ctx := createTestDrawContext()
	got := resolvedTextBBoxOptions(TextBBoxOptions{Style: "not-a-style"}, ctx, 12)
	if got.BoxStyle != BoxStyleSquare {
		t.Fatalf("BoxStyle = %v, want BoxStyleSquare fallback", got.BoxStyle)
	}
}

func bboxPathHasCurves(path geom.Path) bool {
	for _, c := range path.C {
		if c == geom.QuadTo || c == geom.CubicTo {
			return true
		}
	}
	return false
}

func TestTextBBoxStyledPathSquareIsPaddedRect(t *testing.T) {
	local := geom.Rect{Max: geom.Pt{X: 20, Y: 10}}
	path, snap := textBBoxStyledPath(local, TextBBoxOptions{Padding: 4})
	if bboxPathHasCurves(path) {
		t.Fatalf("square box should have no curves: %v", path.C)
	}
	if snap != render.SnapAuto {
		t.Fatalf("square snap = %v, want SnapAuto", snap)
	}
	b, ok := pathBounds(path)
	if !ok {
		t.Fatal("square path has no bounds")
	}
	if !approx(b.Min.X, -4, 1e-9) || !approx(b.Min.Y, -4, 1e-9) ||
		!approx(b.Max.X, 24, 1e-9) || !approx(b.Max.Y, 14, 1e-9) {
		t.Fatalf("square bounds = %+v, want padded by 4", b)
	}
}

func TestTextBBoxStyledPathRoundHasCurves(t *testing.T) {
	local := geom.Rect{Max: geom.Pt{X: 20, Y: 10}}
	path, snap := textBBoxStyledPath(local, TextBBoxOptions{Padding: 4, BoxStyle: BoxStyleRound, RoundingSize: 3})
	if !bboxPathHasCurves(path) {
		t.Fatalf("round box should contain curves: %v", path.C)
	}
	if snap != render.SnapOff {
		t.Fatalf("round snap = %v, want SnapOff", snap)
	}
}

func TestTextBBoxStyledPathLegacyCornerRadius(t *testing.T) {
	local := geom.Rect{Max: geom.Pt{X: 20, Y: 10}}
	path, snap := textBBoxStyledPath(local, TextBBoxOptions{Padding: 4, CornerRadius: 3})
	if !bboxPathHasCurves(path) {
		t.Fatalf("CornerRadius box should contain curves: %v", path.C)
	}
	if snap != render.SnapOff {
		t.Fatalf("CornerRadius snap = %v, want SnapOff", snap)
	}
}
