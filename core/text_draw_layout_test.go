package core

import (
	"math"
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/optional"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestDrawDisplayTextUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayText(r, "plain", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	if r.fontTextCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware draw fontKey = %q, want DejaVu Sans", r.fontTextCalls[0].fontKey)
	}
	if len(r.texts) != 0 {
		t.Fatalf("legacy DrawText should not be used when font-aware draw is available, got %+v", r.texts)
	}
}

func TestDrawDisplayTextRotatedUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayTextRotated(r, "rotated", geom.Pt{X: 10, Y: 20}, 12, math.Pi/8, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one font-aware rotated text draw, got %+v", r.fontRotatedCalls)
	}
	if r.fontRotatedCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware rotated draw fontKey = %q, want DejaVu Sans", r.fontRotatedCalls[0].fontKey)
	}
	if len(r.texts) != 0 {
		t.Fatalf("legacy DrawTextRotated should not be used when font-aware draw is available, got %+v", r.texts)
	}
}

func TestDrawDisplayTextVerticalUsesExplicitFontDrawer(t *testing.T) {
	r := &fontAwareTextRecordingRenderer{}

	drawDisplayTextVertical(r, "vertical", geom.Pt{X: 10, Y: 20}, 12, render.Color{A: 1}, "DejaVu Sans")

	if len(r.fontVerticalCalls) != 1 {
		t.Fatalf("expected one font-aware vertical text draw, got %+v", r.fontVerticalCalls)
	}
	if r.fontVerticalCalls[0].fontKey != "DejaVu Sans" {
		t.Fatalf("font-aware vertical draw fontKey = %q, want DejaVu Sans", r.fontVerticalCalls[0].fontKey)
	}
	if len(r.verticalCalls) != 0 {
		t.Fatalf("legacy DrawTextVertical should not be used when font-aware draw is available, got %+v", r.verticalCalls)
	}
}

func TestAxesTextDefaultsToUnclippedLikeMatplotlib(t *testing.T) {
	fig := NewFigure(200, 120)
	ax := fig.AddAxes(geom.Rect{Max: geom.Pt{X: 1, Y: 1}})

	text := ax.Text(0.5, 0.5, "label", TextOptions{})
	if text.ClipOn {
		t.Fatal("Axes.Text default ClipOn = true, want false")
	}

	clipOn := true
	clipped := ax.Text(0.5, 0.5, "clipped", TextOptions{ClipOn: optional.Of(clipOn)})
	if !clipped.ClipOn {
		t.Fatal("Axes.Text explicit ClipOn=true was not preserved")
	}
}

func TestTextArtistUsesTeXRendererWhenRCUseTeX(t *testing.T) {
	fig := NewFigure(320, 240)
	fig.RC.UseTeX = true
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
	ax.Text(0.5, 0.5, `signal $\alpha$`, TextOptions{FontSize: 20})

	r := &texRecordingRenderer{}
	DrawFigure(fig, r)

	if len(r.texDraws) != 1 {
		t.Fatalf("expected text artist to draw through TeX renderer, got %+v", r.texDraws)
	}
	if len(r.texts) != 0 {
		t.Fatalf("TeX-enabled text artist should not fall back to DrawText, got %+v", r.texts)
	}
}

func TestTextArtistCanDisableMathParsing(t *testing.T) {
	ctx := createTestDrawContext()
	parseMath := false
	text := &Text{
		Position:  geom.Pt{X: 1, Y: 1},
		Content:   `signal $\alpha$`,
		FontSize:  12,
		ParseMath: &parseMath,
		ClipOn:    true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one plain text draw, got %+v", r.fontTextCalls)
	}
	if got, want := r.fontTextCalls[0].text, `signal $\alpha$`; got != want {
		t.Fatalf("parse_math disabled text = %q, want %q", got, want)
	}
}

func TestTextArtistAlphaAppliesToDrawnText(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		Color:    render.Color{R: 0.2, G: 0.4, B: 0.6, A: 0.8},
		ClipOn:   true,
	}
	text.SetAlpha(0.5)
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.textColors) != 1 {
		t.Fatalf("expected one drawn text color, got %+v", r.textColors)
	}
	if !approx(r.textColors[0].A, 0.4, 1e-12) {
		t.Fatalf("text alpha = %v, want local alpha multiplied by artist alpha", r.textColors[0].A)
	}
}

func TestTextArtistWrapWidthUsesMultilineLayoutAndBBox(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:  geom.Pt{X: 1, Y: 1},
		Content:   "alpha beta gamma",
		FontSize:  10,
		WrapWidth: 52,
		ClipOn:    true,
		BBox: &TextBBoxOptions{
			FaceColor: render.Color{R: 1, G: 1, B: 1, A: 1},
			EdgeColor: render.Color{A: 1},
			Padding:   1,
		},
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "alpha beta" || r.texts[1] != "gamma" {
		t.Fatalf("wrapped text lines = %v, want [alpha beta] [gamma]", r.texts)
	}
	if len(r.pathCalls) == 0 {
		t.Fatal("expected wrapped text bbox path")
	}
	bounds, ok := pathBounds(r.pathCalls[0].path)
	if !ok {
		t.Fatalf("missing bbox path bounds: %+v", r.pathCalls[0].path)
	}
	if bounds.W() > text.WrapWidth {
		t.Fatalf("wrapped bbox width = %v, want <= wrap width %v", bounds.W(), text.WrapWidth)
	}
}

func TestTextArtistWrapUsesFigureBoxWidth(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.FigureRect = geom.Rect{Max: geom.Pt{X: 100, Y: 100}}
	text := &Text{
		Position: geom.Pt{X: 0.4, Y: 0.5},
		Coords:   Coords(CoordFigure),
		Content:  "alpha beta gamma",
		FontSize: 10,
		HAlign:   TextAlignLeft,
		Wrap:     true,
		ClipOn:   true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.texts) != 2 || r.texts[0] != "alpha beta" || r.texts[1] != "gamma" {
		t.Fatalf("auto-wrapped text lines = %v, want [alpha beta] [gamma]", r.texts)
	}
}

func TestWrappedTextLinesMatchMatplotlibSpaceAndCeilSemantics(t *testing.T) {
	r := &textRecordingRenderer{}
	got := wrappedTextLines(r, "alpha  beta gamma", 10, "", true, false, 51.4)
	want := []string{"alpha ", "beta gamma"}
	if len(got) != len(want) {
		t.Fatalf("wrapped lines = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("wrapped line %d = %q, want %q (all lines %q)", i, got[i], want[i], got)
		}
	}
}

func TestTextBBoxUsesLineBoxWhenInkBoundsAreShort(t *testing.T) {
	ctx := createTestDrawContext()
	r := &mathInkBoundsRenderer{}
	layout := measureSingleLineTextLayout(r, "bbox", 10, "", false, ctx.RC.UseTeX)
	origin := geom.Pt{X: 100, Y: 100}

	got, ok := textBBoxRect(origin, layout, &TextBBoxOptions{
		FaceColor: render.Color{A: 1},
		EdgeColor: render.Color{A: 1},
		Padding:   1,
	}, ctx, 10)

	if !ok {
		t.Fatal("textBBoxRect returned !ok")
	}
	want := geom.Rect{
		Min: geom.Pt{X: 99, Y: 97},
		Max: geom.Pt{X: 121, Y: 109},
	}
	if !approxRect(got, want, 1e-9) {
		t.Fatalf("text bbox = %+v, want line box %+v", got, want)
	}
}

func TestTextRotationModeAnchorRotatesAroundAlignedTextBox(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tilt",
		FontSize:     10,
		HAlign:       TextAlignLeft,
		VAlign:       TextVAlignTop,
		Angle:        45,
		RotationMode: TextRotationModeAnchor,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	vAlign := layoutVerticalAlign(text.VAlign, false)
	origin := alignedSingleLineOrigin(anchor, layout, text.HAlign, vAlign)
	angle := text.Angle * math.Pi / 180
	// rotation_mode="anchor" ports matplotlib Text._get_layout's anchor branch:
	// the (ha,va) reference of the UNROTATED box is aligned, then rotated.
	p := geom.Pt{
		X: origin.X + textHorizontalOriginOffset(layout, text.HAlign),
		Y: origin.Y - textBaselineOffset(layout, vAlign),
	}
	want := rotatedTextBackendAnchorFromP(p, layout, text.HAlign, vAlign, angle, true)
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("rotation_mode anchor draw anchor = %+v, want %+v", r.fontRotatedCalls[0].anchor, want)
	}
	// Anchor mode must differ from default (rotated-bbox) mode.
	defaultAnchor := tickLabelRotationAnchor(origin, layout, text.HAlign, vAlign, angle)
	if approx(r.fontRotatedCalls[0].anchor.X, defaultAnchor.X, 1e-9) && approx(r.fontRotatedCalls[0].anchor.Y, defaultAnchor.Y, 1e-9) {
		t.Fatalf("rotation_mode anchor unexpectedly matched default rotated-bbox anchor %+v", defaultAnchor)
	}
}

func TestTextRotationModeXTickAdjustsHorizontalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tick",
		FontSize:     10,
		HAlign:       TextAlignCenter,
		VAlign:       TextVAlignBottom,
		Angle:        45,
		RotationMode: TextRotationModeXTick,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	vAlign := layoutVerticalAlign(text.VAlign, false)
	wantOrigin := alignedSingleLineOrigin(anchor, layout, TextAlignLeft, vAlign)
	want := tickLabelRotationAnchor(wantOrigin, layout, TextAlignLeft, vAlign, text.Angle*math.Pi/180)
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("xtick rotation anchor = %+v, want left-aligned anchor %+v", r.fontRotatedCalls[0].anchor, want)
	}
}

func TestTextRotationModeYTickAdjustsVerticalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position:     geom.Pt{X: 1, Y: 1},
		Content:      "tick",
		FontSize:     10,
		HAlign:       TextAlignLeft,
		VAlign:       TextVAlignMiddle,
		Angle:        45,
		RotationMode: TextRotationModeYTick,
		ClipOn:       true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontRotatedCalls) != 1 {
		t.Fatalf("expected one rotated text draw, got %+v", r.fontRotatedCalls)
	}
	anchor := transformedPoint(ctx, text.Coords, text.Position, text.OffsetX, text.OffsetY)
	layout := measureSingleLineTextLayoutParseMath(r, text.Content, text.FontSize, text.FontKey, true, ctx.RC.UseTeX)
	wantVAlign := textLayoutVAlignBaseline
	wantOrigin := alignedSingleLineOrigin(anchor, layout, TextAlignLeft, wantVAlign)
	want := tickLabelRotationAnchor(wantOrigin, layout, TextAlignLeft, wantVAlign, text.Angle*math.Pi/180)
	if !approx(r.fontRotatedCalls[0].anchor.X, want.X, 1e-9) || !approx(r.fontRotatedCalls[0].anchor.Y, want.Y, 1e-9) {
		t.Fatalf("ytick rotation anchor = %+v, want baseline-aligned anchor %+v", r.fontRotatedCalls[0].anchor, want)
	}
}

func TestTextCenterBaselineVerticalAlignment(t *testing.T) {
	ctx := createTestDrawContext()
	text := &Text{
		Position: geom.Pt{X: 1, Y: 2},
		Content:  "center",
		FontSize: 10,
		HAlign:   TextAlignLeft,
		VAlign:   TextVAlignCenterBaseline,
		Coords:   Coords(CoordData),
		ClipOn:   true,
	}
	r := &textRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.origins) != 1 {
		t.Fatalf("text draws = %d, want 1", len(r.origins))
	}
	anchor := ctx.DataToPixel.Apply(text.Position)
	layout := measureSingleLineTextLayout(r, text.Content, text.FontSize, "")
	// Display space is y-up: centering the baseline lowers the origin by half the
	// ascent (smaller Y), so the offset is subtracted.
	want := geom.Pt{X: anchor.X, Y: anchor.Y - layout.Ascent/2}
	if !approx(r.origins[0].X, want.X, 1e-9) || !approx(r.origins[0].Y, want.Y, 1e-9) {
		t.Fatalf("center-baseline origin = %+v, want %+v", r.origins[0], want)
	}
}

func TestTextArtistFontKeyOverridesRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		FontKey:  "Artist Font",
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	if got := r.fontTextCalls[0].fontKey; got != "Artist Font" {
		t.Fatalf("text fontKey = %q, want artist override", got)
	}
}

func TestTextArtistUsesRCFontPropertiesAndFallbackOrder(t *testing.T) {
	theme, _, err := style.ParseMPLStyle("font-rc", `
font.family: serif, Backup Face
font.serif: First Serif, Second Serif
font.style: italic
font.variant: small-caps
font.weight: 700
font.stretch: condensed
`)
	if err != nil {
		t.Fatal(err)
	}
	ctx := createTestDrawContext()
	ctx.RC = theme.RC
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		ClipOn:   true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	wantFamilies := []string{"First Serif", "Second Serif", "Backup Face"}
	if !reflect.DeepEqual(props.Families, wantFamilies) ||
		props.Style != render.FontStyleItalic || props.Variant != "small-caps" ||
		props.Weight != 700 || props.Stretch != "condensed" {
		t.Fatalf("rc font properties = %+v, want families %v with italic small-caps 700 condensed", props, wantFamilies)
	}
}

func TestTextArtistFontPropertiesOverrideRCFontKey(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "plain",
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Serif"},
			Style:    render.FontStyleItalic,
			Weight:   700,
		},
		ClipOn: true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Style != render.FontStyleItalic || props.Weight != 700 || len(props.Families) != 1 || props.Families[0] != "DejaVu Serif" {
		t.Fatalf("text font properties = %+v, want DejaVu Serif italic 700", props)
	}
}

func TestTextArtistFontPropertiesRouteFeatureOptions(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.FontKey = "RC Font"
	text := &Text{
		Position: geom.Pt{X: 1, Y: 1},
		Content:  "fi",
		FontSize: 12,
		FontProperties: &render.FontProperties{
			Families: []string{"DejaVu Sans"},
			Stretch:  "condensed",
			Variant:  "small-caps",
			Language: "de",
			Features: []render.TextFeature{{Tag: "liga", Value: 0}},
		},
		ClipOn: true,
	}
	r := &fontAwareTextRecordingRenderer{}

	text.Draw(r, ctx)

	if len(r.fontTextCalls) != 1 {
		t.Fatalf("expected one font-aware text draw, got %+v", r.fontTextCalls)
	}
	props := render.ParseFontProperties(r.fontTextCalls[0].fontKey)
	if props.Stretch != "condensed" || props.Variant != "small-caps" || props.Language != "de" {
		t.Fatalf("text extended font properties = %+v, want condensed small-caps de", props)
	}
	if len(props.Features) != 2 ||
		props.Features[0] != (render.TextFeature{Tag: "liga", Value: 0}) ||
		props.Features[1] != (render.TextFeature{Tag: "smcp", Value: 1}) {
		t.Fatalf("text font features = %+v, want liga=0 smcp=1", props.Features)
	}
}

func TestAxesTextSupportsAxesAndBlendedCoordinates(t *testing.T) {
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

	ax.Text(0.25, 0.75, "axes", TextOptions{
		Coords:  Coords(CoordAxes),
		OffsetX: 5,
		OffsetY: -7,
	})
	ax.Text(0.25, 0.75, "blend", TextOptions{
		Coords: BlendCoords(CoordFigure, CoordAxes),
	})

	var r textRecordingRenderer
	DrawFigure(fig, &r)

	if len(r.texts) != 2 {
		t.Fatalf("expected 2 text draws, got %d", len(r.texts))
	}

	// Display space is y-up: axes y=0.75 -> 60+0.75*480=420, plus OffsetY=-7 = 413.
	wantAxes := geom.Pt{X: 245, Y: 413}
	if r.origins[0] != wantAxes {
		t.Fatalf("axes coords origin = %+v, want %+v", r.origins[0], wantAxes)
	}

	wantBlend := geom.Pt{X: 200, Y: 420}
	if r.origins[1] != wantBlend {
		t.Fatalf("blended coords origin = %+v, want %+v", r.origins[1], wantBlend)
	}
}
