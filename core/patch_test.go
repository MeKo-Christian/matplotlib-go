package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestRectangleDrawAndBounds(t *testing.T) {
	rect := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			EdgeWidth: 2,
		},
		XY:     geom.Pt{X: 1, Y: 2},
		Width:  3,
		Height: 4,
	}

	r := &recordingRenderer{}
	rect.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.Fill.A <= 0 {
		t.Fatalf("expected rectangle fill paint, got %+v", r.pathCalls[0].paint)
	}
	if r.pathCalls[0].paint.Stroke.A <= 0 {
		t.Fatalf("expected rectangle stroke paint, got %+v", r.pathCalls[0].paint)
	}

	bounds := rect.Bounds(nil)
	want := geom.Rect{
		Min: geom.Pt{X: 1, Y: 2},
		Max: geom.Pt{X: 4, Y: 6},
	}
	if bounds != want {
		t.Fatalf("bounds = %+v, want %+v", bounds, want)
	}
}

func TestPatchDrawUsesMatplotlibSnapAuto(t *testing.T) {
	rect := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
			EdgeColor: render.Color{A: 1},
			EdgeWidth: 1,
		},
		XY:     geom.Pt{X: 1, Y: 2},
		Width:  3,
		Height: 4,
	}

	r := &recordingRenderer{}
	rect.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one path call, got %d", len(r.pathCalls))
	}
	if got := r.pathCalls[0].paint.Snap; got != render.SnapAuto {
		t.Fatalf("patch snap mode = %v, want Matplotlib SnapAuto", got)
	}
}

func TestPatchDashesCanUseMatplotlibLineWidthScaling(t *testing.T) {
	rect := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{A: 0},
			EdgeColor: render.Color{A: 1},
			EdgeWidth: 2.5,
			Dashes:    []float64{4, 2},
			DashUnits: DashUnitsMatplotlib,
		},
		Width:  1,
		Height: 1,
	}

	r := &recordingRenderer{}
	rect.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one stroked path call, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].paint.Dashes
	want := []float64{10, 5}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("patch dashes = %v, want Matplotlib linewidth-scaled %v", got, want)
	}
}

func TestPatchDashesDefaultToRendererUnits(t *testing.T) {
	rect := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{A: 0},
			EdgeColor: render.Color{A: 1},
			EdgeWidth: 2.5,
			Dashes:    []float64{4, 2},
		},
		Width:  1,
		Height: 1,
	}

	r := &recordingRenderer{}
	rect.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one stroked path call, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].paint.Dashes
	want := []float64{4, 2}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("patch dashes = %v, want raw renderer-unit dashes %v", got, want)
	}
}

func TestFancyBboxPatchRoundUsesQuadraticCornersAndHatch(t *testing.T) {
	box := &FancyBboxPatch{
		Patch: Patch{
			FaceColor:    render.Color{R: 0.3, G: 0.6, B: 0.9, A: 0.7},
			EdgeColor:    render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1},
			EdgeWidth:    1.5,
			Hatch:        "/",
			HatchColor:   render.Color{R: 0.2, G: 0.2, B: 0.2, A: 1},
			HatchSpacing: 6,
		},
		XY:           geom.Pt{X: 1, Y: 1},
		Width:        2.5,
		Height:       1.5,
		Pad:          0.2,
		BoxStyle:     BoxStyleRound,
		RoundingSize: 0.25,
	}

	r := &recordingRenderer{}
	box.Draw(r, createTestDrawContext())

	if len(r.pathCalls) < 2 {
		t.Fatalf("expected fill path plus hatch strokes, got %d calls", len(r.pathCalls))
	}
	if len(r.pathCalls[0].path.C) == 0 {
		t.Fatal("expected rounded path commands")
	}
	hasQuad := false
	for _, cmd := range r.pathCalls[0].path.C {
		if cmd == geom.QuadTo {
			hasQuad = true
			break
		}
	}
	if !hasQuad {
		t.Fatalf("expected rounded fancy box path to use quadratic curves, got %v", r.pathCalls[0].path.C)
	}
}

func TestPatchDefaultHatchSpacingMatchesMatplotlibDensity(t *testing.T) {
	patch := &Patch{}
	want := 100.0 / 6.0
	if got := patch.resolvedHatchSpacing(); !approx(got, want, 1e-12) {
		t.Fatalf("default hatch spacing = %v, want Matplotlib density spacing %v", got, want)
	}
}

func TestFancyBboxPatchRoundUsesMutationScaleAndAspect(t *testing.T) {
	box := &FancyBboxPatch{
		Width:          2,
		Height:         1,
		Pad:            0.3,
		BoxStyle:       BoxStyleRound,
		MutationSize:   2,
		MutationAspect: 2,
	}

	path := box.localPath()
	assertApproxPathBounds(t, path, geom.Rect{
		Min: geom.Pt{X: -0.6, Y: -1.2},
		Max: geom.Pt{X: 2.6, Y: 2.2},
	})
	if got := countPathCmd(path, geom.QuadTo); got != 4 {
		t.Fatalf("round quadratic corners = %d, want 4; commands=%v", got, path.C)
	}
}

func TestFancyBboxPatchAdditionalBoxStyles(t *testing.T) {
	tests := []struct {
		name      string
		style     BoxStyle
		width     float64
		height    float64
		pad       float64
		want      geom.Rect
		wantCmd   geom.Cmd
		minCount  int
		closePath bool
	}{
		{
			name:      "circle",
			style:     BoxStyleCircle,
			width:     2,
			height:    1,
			pad:       0.5,
			want:      geom.Rect{Min: geom.Pt{X: -1, Y: -1.5}, Max: geom.Pt{X: 3, Y: 2.5}},
			wantCmd:   geom.CubicTo,
			minCount:  4,
			closePath: true,
		},
		{
			name:      "ellipse",
			style:     BoxStyleEllipse,
			width:     2,
			height:    1,
			pad:       0.5,
			want:      geom.Rect{Min: geom.Pt{X: 1 - 4/math.Sqrt2, Y: 0.5 - 3/math.Sqrt2}, Max: geom.Pt{X: 1 + 4/math.Sqrt2, Y: 0.5 + 3/math.Sqrt2}},
			wantCmd:   geom.CubicTo,
			minCount:  4,
			closePath: true,
		},
		{
			name:      "round4",
			style:     BoxStyleRound4,
			width:     2,
			height:    1,
			pad:       0.2,
			want:      geom.Rect{Min: geom.Pt{X: -0.4, Y: -0.4}, Max: geom.Pt{X: 2.4, Y: 1.4}},
			wantCmd:   geom.CubicTo,
			minCount:  4,
			closePath: true,
		},
		{
			name:      "rarrow",
			style:     BoxStyleRArrow,
			width:     4,
			height:    2,
			want:      geom.Rect{Min: geom.Pt{X: 0, Y: -0.5}, Max: geom.Pt{X: 5, Y: 2.5}},
			wantCmd:   geom.LineTo,
			minCount:  6,
			closePath: true,
		},
		{
			name:      "larrow",
			style:     BoxStyleLArrow,
			width:     4,
			height:    2,
			want:      geom.Rect{Min: geom.Pt{X: -1, Y: -0.5}, Max: geom.Pt{X: 4, Y: 2.5}},
			wantCmd:   geom.LineTo,
			minCount:  6,
			closePath: true,
		},
		{
			name:      "darrow",
			style:     BoxStyleDArrow,
			width:     4,
			height:    2,
			want:      geom.Rect{Min: geom.Pt{X: -1, Y: -0.5}, Max: geom.Pt{X: 5.5, Y: 2.5}},
			wantCmd:   geom.LineTo,
			minCount:  8,
			closePath: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			box := &FancyBboxPatch{
				Width:        tt.width,
				Height:       tt.height,
				Pad:          tt.pad,
				BoxStyle:     tt.style,
				MutationSize: 2,
			}
			if tt.pad == 0 {
				box.MutationSize = 1
			}
			path := box.localPath()
			assertApproxPathBounds(t, path, tt.want)
			if got := countPathCmd(path, tt.wantCmd); got < tt.minCount {
				t.Fatalf("%s %v count = %d, want at least %d; commands=%v", tt.name, tt.wantCmd, got, tt.minCount, path.C)
			}
			if tt.closePath && (len(path.C) == 0 || path.C[len(path.C)-1] != geom.ClosePath) {
				t.Fatalf("%s should close path, commands=%v", tt.name, path.C)
			}
		})
	}
}

func TestFancyBboxPatchStyleMatrixArrowBoxesMatchMatplotlibDisplay(t *testing.T) {
	ctx := styleMatrixTestContext()

	rarrow := &FancyBboxPatch{
		XY:           geom.Pt{X: 7.4, Y: 5.4},
		Width:        1.35,
		Height:       0.62,
		Pad:          0.10,
		BoxStyle:     BoxStyleRArrow,
		MutationSize: 1,
		Coords:       Coords(CoordData),
	}
	assertPathVerticesApprox(t, rarrow.displayPath(ctx), []geom.Pt{
		{X: 505.149, Y: 275.625},
		{X: 433.320, Y: 275.625},
		{X: 433.320, Y: 314.370},
		{X: 505.149, Y: 314.370},
		{X: 505.149, Y: 324.056},
		{X: 539.835, Y: 294.998},
		{X: 505.149, Y: 265.939},
	}, 0.02)

	darrow := &FancyBboxPatch{
		XY:           geom.Pt{X: 9.65, Y: 5.4},
		Width:        1.35,
		Height:       0.62,
		Pad:          0.10,
		BoxStyle:     BoxStyleDArrow,
		MutationSize: 1,
		Coords:       Coords(CoordData),
	}
	assertPathVerticesApprox(t, darrow.displayPath(ctx), []geom.Pt{
		{X: 575.811, Y: 275.625},
		{X: 636.360, Y: 275.625},
		{X: 636.360, Y: 265.939},
		{X: 671.046, Y: 294.998},
		{X: 636.360, Y: 324.056},
		{X: 636.360, Y: 314.370},
		{X: 575.811, Y: 314.370},
		{X: 575.811, Y: 324.056},
		{X: 541.125, Y: 294.998},
		{X: 575.811, Y: 265.939},
	}, 0.02)
}

func TestFancyBboxPatchToothStyles(t *testing.T) {
	saw := (&FancyBboxPatch{
		Width:    4,
		Height:   2,
		Pad:      0.4,
		BoxStyle: BoxStyleSawtooth,
	}).localPath()
	round := (&FancyBboxPatch{
		Width:    4,
		Height:   2,
		Pad:      0.4,
		BoxStyle: BoxStyleRoundtooth,
	}).localPath()

	if got := countPathCmd(saw, geom.LineTo); got < 20 {
		t.Fatalf("sawtooth line count = %d, want detailed tooth outline; commands=%v", got, saw.C)
	}
	if got := countPathCmd(round, geom.QuadTo); got < 10 {
		t.Fatalf("roundtooth quadratic count = %d, want rounded tooth outline; commands=%v", got, round.C)
	}
	if len(saw.C) == 0 || saw.C[len(saw.C)-1] != geom.ClosePath {
		t.Fatalf("sawtooth should close path, commands=%v", saw.C)
	}
	if len(round.C) == 0 || round.C[len(round.C)-1] != geom.ClosePath {
		t.Fatalf("roundtooth should close path, commands=%v", round.C)
	}
}

func TestFancyBboxPatchLArrowReflectsAroundOriginalBoxWithPadding(t *testing.T) {
	box := &FancyBboxPatch{
		Width:        4,
		Height:       2,
		Pad:          0.5,
		BoxStyle:     BoxStyleLArrow,
		MutationSize: 2,
	}

	assertApproxPathBounds(t, box.localPath(), geom.Rect{
		Min: geom.Pt{X: -2.2857142857142856, Y: -2},
		Max: geom.Pt{X: 5, Y: 4},
	})
}

func countPathCmd(path geom.Path, want geom.Cmd) int {
	count := 0
	for _, cmd := range path.C {
		if cmd == want {
			count++
		}
	}
	return count
}

func assertApproxPathBounds(t *testing.T, path geom.Path, want geom.Rect) {
	t.Helper()
	got, ok := pathBounds(path)
	if !ok {
		t.Fatal("path has no bounds")
	}
	if !approxPt(got.Min, want.Min, 1e-9) || !approxPt(got.Max, want.Max, 1e-9) {
		t.Fatalf("bounds = %+v, want %+v", got, want)
	}
}

func assertPathVerticesApprox(t *testing.T, got geom.Path, want []geom.Pt, tol float64) {
	t.Helper()
	if len(got.V) != len(want) {
		t.Fatalf("path vertices = %d, want %d: %+v", len(got.V), len(want), got.V)
	}
	for i := range want {
		if !approxPt(got.V[i], want[i], tol) {
			t.Fatalf("vertex %d = %+v, want %+v (all vertices %+v)", i, got.V[i], want[i], got.V)
		}
	}
}

func styleMatrixTestContext() *DrawContext {
	fig := NewFigure(720, 420)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.03, Y: 0.06}, Max: geom.Pt{X: 0.97, Y: 0.96}})
	ax.SetXLim(0, 12)
	ax.SetYLim(0, 8)
	return AxesDrawContext(ax, fig)
}

func approxPt(a, b geom.Pt, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

func containsPointForPatchTest(path geom.Path, want geom.Pt) bool {
	for _, pt := range path.V {
		if approx(pt.X, want.X, 1e-9) && approx(pt.Y, want.Y, 1e-9) {
			return true
		}
	}
	return false
}

func maxPathYForPatchTest(path geom.Path) float64 {
	maxY := math.Inf(-1)
	for _, pt := range path.V {
		maxY = math.Max(maxY, pt.Y)
	}
	return maxY
}

func TestPatchAutoScaleIgnoresNonDataCoords(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&Rectangle{
		Patch:  Patch{FaceColor: render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}},
		XY:     geom.Pt{X: 2, Y: 3},
		Width:  2,
		Height: 3,
	})
	ax.AddPatch(&Circle{
		Patch:  Patch{FaceColor: render.Color{R: 0.2, G: 0.4, B: 0.8, A: 1}},
		Center: geom.Pt{X: 0.5, Y: 0.5},
		Radius: 0.35,
		Coords: Coords(CoordAxes),
	})

	ax.AutoScale(0)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if xMin != 2 || xMax != 4 {
		t.Fatalf("x domain = [%v, %v], want [2, 4]", xMin, xMax)
	}
	if yMin != 3 || yMax != 6 {
		t.Fatalf("y domain = [%v, %v], want [3, 6]", yMin, yMax)
	}
}

func TestPathPatchLegendEntryIncludesHatch(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.AddPatch(&PathPatch{
		Patch: Patch{
			FaceColor:  render.Color{R: 0.7, G: 0.7, B: 0.2, A: 1},
			EdgeColor:  render.Color{R: 0.2, G: 0.2, B: 0.1, A: 1},
			EdgeWidth:  1,
			Hatch:      "x",
			HatchColor: render.Color{R: 0, G: 0, B: 0, A: 1},
			Label:      "region",
		},
		Path: patchRectPath(geom.Rect{
			Min: geom.Pt{X: 1, Y: 1},
			Max: geom.Pt{X: 3, Y: 2},
		}),
	})

	entries := ax.AddLegend().collectEntries()
	if len(entries) != 1 {
		t.Fatalf("expected one legend entry, got %d", len(entries))
	}
	if entries[0].Label != "region" || entries[0].kind != legendEntryPatch {
		t.Fatalf("unexpected legend entry: %+v", entries[0])
	}
	if entries[0].patchHatch != "x" {
		t.Fatalf("expected hatch metadata to be preserved, got %+v", entries[0])
	}
}

func TestFancyArrowBoundsAndClosedPath(t *testing.T) {
	arrow := &FancyArrow{
		Patch: Patch{
			FaceColor: render.Color{R: 0.9, G: 0.3, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0.2, G: 0.1, B: 0.1, A: 1},
			EdgeWidth: 1,
		},
		XY:                 geom.Pt{X: 1, Y: 1},
		DX:                 4,
		DY:                 0,
		Width:              0.4,
		HeadWidth:          1.2,
		HeadLength:         1.1,
		LengthIncludesHead: true,
	}

	bounds := arrow.Bounds(nil)
	if bounds.Min.X != 1 || bounds.Max.X != 5 {
		t.Fatalf("x bounds = [%v, %v], want [1, 5]", bounds.Min.X, bounds.Max.X)
	}
	if bounds.Min.Y >= 1 || bounds.Max.Y <= 1 {
		t.Fatalf("expected arrow bounds to span around y=1, got %+v", bounds)
	}

	r := &recordingRenderer{}
	arrow.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one arrow path call, got %d", len(r.pathCalls))
	}
	path := r.pathCalls[0].path
	if len(path.C) == 0 || path.C[len(path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed arrow polygon, got %v", path.C)
	}
}

func TestFancyArrowLengthIncludesHeadMatchesMatplotlib(t *testing.T) {
	arrow := &FancyArrow{
		XY:         geom.Pt{X: 1, Y: 1},
		DX:         4,
		DY:         0,
		Width:      0.4,
		HeadWidth:  1.2,
		HeadLength: 1.1,
	}
	bounds := arrow.Bounds(nil)
	if bounds.Min.X != 1 || bounds.Max.X != 6.1 {
		t.Fatalf("default x bounds = [%v, %v], want Matplotlib length_includes_head=False bounds [1, 6.1]", bounds.Min.X, bounds.Max.X)
	}

	arrow.LengthIncludesHead = true
	bounds = arrow.Bounds(nil)
	if bounds.Min.X != 1 || bounds.Max.X != 5 {
		t.Fatalf("length-includes-head x bounds = [%v, %v], want [1, 5]", bounds.Min.X, bounds.Max.X)
	}
}

func TestArrowAndConnectionStyleRegistriesParseMatplotlibNames(t *testing.T) {
	for _, name := range []string{"-", "->", "<-", "<->", "-[", "]-[", "|-|", "-|>", "<|-|>", "simple", "fancy", "wedge"} {
		style, ok := ArrowStyleFromString(name)
		if !ok {
			t.Fatalf("ArrowStyleFromString(%q) returned !ok", name)
		}
		if style.Name == "" {
			t.Fatalf("ArrowStyleFromString(%q) returned empty name", name)
		}
	}

	style, ok := ArrowStyleFromString("->,head_length=0.8,head_width=0.4")
	if !ok {
		t.Fatal("ArrowStyleFromString with parameters returned !ok")
	}
	if style.HeadLength != 0.8 || style.HeadWidth != 0.4 {
		t.Fatalf("style parameters not parsed: %+v", style)
	}

	for _, name := range []string{"arc3", "arc", "angle", "angle3", "bar"} {
		style, ok := ConnectionStyleFromString(name)
		if !ok {
			t.Fatalf("ConnectionStyleFromString(%q) returned !ok", name)
		}
		if style.Name == "" {
			t.Fatalf("ConnectionStyleFromString(%q) returned empty name", name)
		}
	}

	conn, ok := ConnectionStyleFromString("arc3,rad=0.35")
	if !ok {
		t.Fatal("ConnectionStyleFromString with parameters returned !ok")
	}
	if conn.Rad != 0.35 {
		t.Fatalf("connection parameters not parsed: %+v", conn)
	}
}

func TestConnectionStyleArc3ZeroRadKeepsQuadraticPath(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc3")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc3) returned !ok")
	}
	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo {
		t.Fatalf("arc3 zero-rad commands = %+v, want MoveTo/QuadTo", path.C)
	}
	want := []geom.Pt{{X: 0, Y: 0}, {X: 50, Y: 0}, {X: 100, Y: 0}}
	if len(path.V) != len(want) {
		t.Fatalf("arc3 zero-rad vertices = %+v, want %+v", path.V, want)
	}
	for i := range want {
		if !approxPt(path.V[i], want[i], 1e-9) {
			t.Fatalf("arc3 zero-rad vertex[%d] = %+v, want %+v", i, path.V[i], want[i])
		}
	}
}

func TestConnectionStyleArc3RadUsesDisplayYUpCoordinates(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc3,rad=0.25")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc3,rad=0.25) returned !ok")
	}

	start := geom.Pt{X: 100, Y: 120}
	end := geom.Pt{X: 200, Y: 170}
	path := style.connect(start, end, 0, 0)

	// Display space is y-up, so connect() uses the verbatim Matplotlib formula
	// (Arc3.connect): cx = x12 + rad*dy, cy = y12 - rad*dx.
	wantCtrl := geom.Pt{
		X: (start.X+end.X)/2 + style.Rad*(end.Y-start.Y),
		Y: (start.Y+end.Y)/2 - style.Rad*(end.X-start.X),
	}
	if len(path.V) != 3 || !approxPt(path.V[1], wantCtrl, 1e-9) {
		t.Fatalf("arc3 control = %+v, want y-up Matplotlib-equivalent control %+v", path.V, wantCtrl)
	}
}

func TestConnectionStyleArcUsesMatplotlibDefaultAngles(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc,armA=10,armB=5")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc) returned !ok")
	}
	if style.AngleA != 0 || style.AngleB != 0 {
		t.Fatalf("arc defaults = angleA %v angleB %v, want 0/0", style.AngleA, style.AngleB)
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.V) < 3 {
		t.Fatalf("arc path vertices = %+v, want start arm and end arm", path.V)
	}
	if !approx(path.V[1].X, 10, 1e-9) || !approx(path.V[1].Y, 0, 1e-9) {
		t.Fatalf("arc start arm = %+v, want horizontal +x arm at {10,0}", path.V[1])
	}
}

func TestConnectionStyleArcRadiusRoundsArmEndpointLikeMatplotlib(t *testing.T) {
	style, ok := ConnectionStyleFromString("arc,armA=20,rad=5")
	if !ok {
		t.Fatal("ConnectionStyleFromString(arc) returned !ok")
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 0}, 0, 0)
	if len(path.V) < 5 {
		t.Fatalf("arc rounded path vertices = %+v, want start, rounded arm, and endpoint", path.V)
	}
	want := []geom.Pt{
		{X: 0, Y: 0},
		{X: 15, Y: 0},
		{X: 20, Y: 0},
		{X: 25, Y: 0},
		{X: 100, Y: 0},
	}
	for i := range want {
		if !approxPt(path.V[i], want[i], 1e-9) {
			t.Fatalf("arc rounded vertex %d = %+v, want %+v (all vertices %+v)", i, path.V[i], want[i], path.V)
		}
	}
	if path.C[2] != geom.QuadTo {
		t.Fatalf("arc rounded corner command = %v, want QuadTo in commands %+v", path.C[2], path.C)
	}
}

func TestConnectionStyleBarAngleProjectsEndpointLikeMatplotlib(t *testing.T) {
	style, ok := ConnectionStyleFromString("bar,angle=0,fraction=0.3")
	if !ok {
		t.Fatal("ConnectionStyleFromString(bar) returned !ok")
	}

	path := style.connect(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 100, Y: 100}, 0, 0)
	if len(path.V) != 4 {
		t.Fatalf("bar path vertices = %+v, want 4 vertices", path.V)
	}
	wantY := -0.3 * math.Hypot(100, 100)
	if !approx(path.V[1].X, 0, 1e-9) || !approx(path.V[1].Y, wantY, 1e-9) {
		t.Fatalf("bar first arm = %+v, want {0,%v}", path.V[1], wantY)
	}
	if !approx(path.V[2].X, 100, 1e-9) || !approx(path.V[2].Y, wantY, 1e-9) {
		t.Fatalf("bar projected second arm = %+v, want {100,%v}", path.V[2], wantY)
	}
	if path.V[3] != (geom.Pt{X: 100, Y: 100}) {
		t.Fatalf("bar final endpoint = %+v, want original endpoint", path.V[3])
	}
}

func TestFancyArrowPatchDrawsConnectionAndArrowHead(t *testing.T) {
	arrowStyle, ok := ArrowStyleFromString("-|>")
	if !ok {
		t.Fatal("missing -|> arrow style")
	}
	connectionStyle, ok := ConnectionStyleFromString("arc3,rad=0.2")
	if !ok {
		t.Fatal("missing arc3 connection style")
	}
	patch := &FancyArrowPatch{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.2, B: 0.1, A: 1},
			EdgeColor: render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1},
			EdgeWidth: 2,
		},
		PosA:            geom.Pt{X: 1, Y: 1},
		PosB:            geom.Pt{X: 4, Y: 3},
		ArrowStyle:      arrowStyle,
		ConnectionStyle: connectionStyle,
		MutationScale:   12,
		Coords:          Coords(CoordData),
	}

	r := &recordingRenderer{}
	patch.Draw(r, createTestDrawContext())

	if len(r.pathCalls) < 2 {
		t.Fatalf("expected connection path and arrow head, got %d paths", len(r.pathCalls))
	}
	if r.pathCalls[0].path.C[0] != geom.MoveTo || r.pathCalls[0].path.C[len(r.pathCalls[0].path.C)-1] != geom.QuadTo {
		t.Fatalf("expected curved connection path, got commands %v", r.pathCalls[0].path.C)
	}
	if r.pathCalls[len(r.pathCalls)-1].path.C[len(r.pathCalls[len(r.pathCalls)-1].path.C)-1] != geom.ClosePath {
		t.Fatalf("expected closed arrow-head path, got %v", r.pathCalls[len(r.pathCalls)-1].path.C)
	}
}

func TestFancyArrowPatchDefaultCapAndJoinMatchMatplotlib(t *testing.T) {
	arrowStyle, ok := ArrowStyleFromString("->")
	if !ok {
		t.Fatal("missing -> arrow style")
	}
	patch := &FancyArrowPatch{
		Patch: Patch{
			EdgeColor: render.Color{A: 1},
			EdgeWidth: 1,
		},
		PosA:       geom.Pt{X: 1, Y: 1},
		PosB:       geom.Pt{X: 4, Y: 3},
		ArrowStyle: arrowStyle,
		Coords:     Coords(CoordData),
	}
	r := &recordingRenderer{}

	patch.Draw(r, createTestDrawContext())

	if len(r.pathCalls) == 0 {
		t.Fatal("expected fancy arrow paths")
	}
	for _, call := range r.pathCalls {
		if call.paint.LineJoin != render.JoinRound || call.paint.LineCap != render.CapRound {
			t.Fatalf("fancy arrow paint cap/join = %v/%v, want round/round", call.paint.LineCap, call.paint.LineJoin)
		}
	}
}

func TestFancyArrowPatchMutationAspectScalesArrowMutation(t *testing.T) {
	arrowStyle, ok := ArrowStyleFromString("-|>")
	if !ok {
		t.Fatal("missing -|> arrow style")
	}
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 0, Y: 100})

	base := (&FancyArrowPatch{
		ArrowStyle:    arrowStyle,
		MutationScale: 10,
	}).displayParts(nil, path)
	stretched := (&FancyArrowPatch{
		ArrowStyle:     arrowStyle,
		MutationScale:  10,
		MutationAspect: 2,
	}).displayParts(nil, path)

	if len(base) < 2 || len(stretched) < 2 {
		t.Fatalf("expected line and arrow head parts, got base=%+v stretched=%+v", base, stretched)
	}
	baseHead, ok := pathBounds(base[len(base)-1].path)
	if !ok {
		t.Fatalf("missing base arrow head bounds: %+v", base[len(base)-1].path)
	}
	stretchedHead, ok := pathBounds(stretched[len(stretched)-1].path)
	if !ok {
		t.Fatalf("missing stretched arrow head bounds: %+v", stretched[len(stretched)-1].path)
	}
	if stretchedHead.H() <= baseHead.H()*1.5 {
		t.Fatalf("mutation aspect did not stretch arrow head height: base=%+v stretched=%+v", baseHead, stretchedHead)
	}
	headLength := arrowStyle.HeadLength * 10
	headWidth := arrowStyle.HeadWidth * 10
	headDist := math.Hypot(headLength, headWidth)
	padProjected := 0.5 * 1 / (headWidth / headDist)
	wantTipY := 100 - padProjected*2
	if !approx(stretchedHead.Max.Y, wantTipY, 1e-9) {
		t.Fatalf("mutation aspect arrow tip y = %v, want linewidth-projected tip %v", stretchedHead.Max.Y, wantTipY)
	}
}

func TestFancyArrowPatchMutationScaleUsesPointUnits(t *testing.T) {
	arrowStyle, ok := ArrowStyleFromString("->,head_length=0.4,head_width=0.2")
	if !ok {
		t.Fatal("missing -> arrow style")
	}
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 100, Y: 0})
	ctx := createTestDrawContext()
	ctx.RC.DPI = 144

	parts := (&FancyArrowPatch{
		ArrowStyle:    arrowStyle,
		MutationScale: 10,
	}).displayParts(ctx, path)

	if len(parts) != 2 {
		t.Fatalf("curve arrow parts = %d, want line plus open head", len(parts))
	}
	headBounds, ok := pathBounds(parts[1].path)
	if !ok {
		t.Fatalf("missing arrow head bounds: %+v", parts[1].path)
	}
	wantHeight := 2 * arrowStyle.HeadWidth * 10 * ctx.RC.DPI / 72.0
	if !approx(headBounds.H(), wantHeight, 1e-9) {
		t.Fatalf("open arrow head height = %v, want mutation_scale in points -> %v px", headBounds.H(), wantHeight)
	}
}

func TestFancyArrowPatchStyleMatrixCurveArrowMatchesMatplotlib(t *testing.T) {
	arrowStyle, ok := ArrowStyleFromString("->,head_length=0.35,head_width=0.22")
	if !ok {
		t.Fatal("missing -> arrow style")
	}
	connectionStyle, ok := ConnectionStyleFromString("arc,armA=0.9,armB=0.65,rad=0.18")
	if !ok {
		t.Fatal("missing arc connection style")
	}
	patch := &FancyArrowPatch{
		PosA:            geom.Pt{X: 0.9, Y: 2.25},
		PosB:            geom.Pt{X: 3.1, Y: 2.6},
		ArrowStyle:      arrowStyle,
		ConnectionStyle: connectionStyle,
		MutationScale:   15,
		Coords:          Coords(CoordData),
		Patch:           Patch{EdgeWidth: 1.25},
	}
	ctx := styleMatrixTestContext()

	parts := patch.displayParts(ctx, patch.displayPath(ctx))
	if len(parts) != 2 {
		t.Fatalf("curve arrow parts = %d, want line plus head", len(parts))
	}
	assertPathVerticesApprox(t, parts[0].path, []geom.Pt{
		{X: 75.130, Y: 131.762},
		{X: 192.533, Y: 147.441},
	}, 0.02)
	assertPathVerticesApprox(t, parts[1].path, []geom.Pt{
		{X: 184.699, Y: 151.019},
		{X: 192.533, Y: 147.441},
		{X: 185.912, Y: 141.933},
	}, 0.02)
}

func TestFancyArrowPatchStyleMatrixBarAndWedgeMatchMatplotlib(t *testing.T) {
	ctx := styleMatrixTestContext()

	barStyle, ok := ArrowStyleFromString("|-|")
	if !ok {
		t.Fatal("missing |-| arrow style")
	}
	barConnection, ok := ConnectionStyleFromString("bar,fraction=0.25,angle=0")
	if !ok {
		t.Fatal("missing bar connection style")
	}
	bar := &FancyArrowPatch{
		PosA:            geom.Pt{X: 4.0, Y: 2.62},
		PosB:            geom.Pt{X: 6.25, Y: 1.88},
		ArrowStyle:      barStyle,
		ConnectionStyle: barConnection,
		MutationScale:   15,
		Coords:          Coords(CoordData),
		Patch:           Patch{EdgeWidth: 1.25},
	}
	barParts := bar.displayParts(ctx, bar.displayPath(ctx))
	if len(barParts) != 3 {
		t.Fatalf("bar arrow parts = %d, want line plus two bars", len(barParts))
	}
	assertPathVerticesApprox(t, barParts[0].path, []geom.Pt{
		{X: 247.200, Y: 146.215},
		{X: 247.200, Y: 81.123},
		{X: 374.100, Y: 81.123},
		{X: 374.100, Y: 111.254},
	}, 0.02)

	wedgeStyle, ok := ArrowStyleFromString("wedge,tail_width=0.26,shrink_factor=0.35")
	if !ok {
		t.Fatal("missing wedge arrow style")
	}
	wedgeConnection, ok := ConnectionStyleFromString("arc3,rad=0.22")
	if !ok {
		t.Fatal("missing arc3 connection style")
	}
	wedge := &FancyArrowPatch{
		PosA:            geom.Pt{X: 7.05, Y: 2.08},
		PosB:            geom.Pt{X: 10.8, Y: 2.58},
		ArrowStyle:      wedgeStyle,
		ConnectionStyle: wedgeConnection,
		MutationScale:   15,
		Coords:          Coords(CoordData),
		Patch:           Patch{EdgeWidth: 1.25},
	}
	wedgeParts := wedge.displayParts(ctx, wedge.displayPath(ctx))
	if len(wedgeParts) != 1 {
		t.Fatalf("wedge arrow parts = %d, want one filled path", len(wedgeParts))
	}
	assertPathVerticesApprox(t, wedgeParts[0].path, []geom.Pt{
		{X: 421.091, Y: 120.070},
		{X: 530.650, Y: 89.275},
		{X: 628.314, Y: 145.723},
		{X: 628.314, Y: 145.723},
		{X: 529.444, Y: 90.451},
		{X: 422.662, Y: 125.253},
	}, 0.02)
}

func TestArrowStyleWedgeUsesShrinkFactor(t *testing.T) {
	style, ok := ArrowStyleFromString("wedge,tail_width=0.6,shrink_factor=0.25")
	if !ok {
		t.Fatal("ArrowStyleFromString(wedge) returned !ok")
	}
	if !approx(style.ShrinkFactor, 0.25, 1e-12) {
		t.Fatalf("wedge shrink factor = %v, want 0.25", style.ShrinkFactor)
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 1 || !parts[0].fillable {
		t.Fatalf("wedge parts = %+v, want one fillable path", parts)
	}
	got := parts[0].path
	if len(got.V) != 5 {
		t.Fatalf("wedge path vertices = %d, want tapered closed outline: %+v", len(got.V), got)
	}

	tailWidth := got.V[0].Y - got.V[4].Y
	midWidth := got.V[1].Y - got.V[3].Y
	if !approx(tailWidth, 6, 1e-9) {
		t.Fatalf("wedge tail width = %v, want 6", tailWidth)
	}
	if !approx(midWidth, 1.5, 1e-9) {
		t.Fatalf("wedge middle width = %v, want 1.5 from shrink_factor", midWidth)
	}
	if got.V[2] != (geom.Pt{X: 100, Y: 0}) {
		t.Fatalf("wedge should taper to the arrow endpoint, got tip %+v", got.V[2])
	}
}

func TestArrowStyleWedgeFollowsQuadraticConnection(t *testing.T) {
	style, ok := ArrowStyleFromString("wedge,tail_width=0.6,shrink_factor=0.25")
	if !ok {
		t.Fatal("ArrowStyleFromString(wedge) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.QuadTo(geom.Pt{X: 50, Y: 50}, geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 1 || !parts[0].fillable {
		t.Fatalf("wedge parts = %+v, want one fillable path", parts)
	}
	got := parts[0].path
	if len(got.V) != 6 {
		t.Fatalf("wedge path vertices = %d, want tapered closed outline: %+v", len(got.V), got)
	}
	if got.V[1].Y < 20 || got.V[4].Y < 20 {
		t.Fatalf("quadratic wedge ignored connection control point, got vertices %+v", got.V)
	}
	if got.V[2] != (geom.Pt{X: 100, Y: 0}) {
		t.Fatalf("wedge should taper to the quadratic endpoint, got tip %+v", got.V[2])
	}
}

func TestArrowStyleWedgeUsesQuadraticOutlineForQuadraticConnection(t *testing.T) {
	style, ok := ArrowStyleFromString("wedge,tail_width=0.6,shrink_factor=0.25")
	if !ok {
		t.Fatal("ArrowStyleFromString(wedge) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.QuadTo(geom.Pt{X: 50, Y: 50}, geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 1 || !parts[0].fillable {
		t.Fatalf("wedge parts = %+v, want one fillable path", parts)
	}

	want := []geom.Cmd{
		geom.MoveTo,
		geom.QuadTo,
		geom.LineTo,
		geom.QuadTo,
		geom.ClosePath,
	}
	if got := parts[0].path.C; len(got) != len(want) {
		t.Fatalf("wedge commands = %+v, want %+v", got, want)
	} else {
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("wedge commands = %+v, want %+v", got, want)
			}
		}
	}
}

func TestArrowStyleSimpleFollowsQuadraticConnection(t *testing.T) {
	style, ok := ArrowStyleFromString("simple,tail_width=0.3,head_width=0.8,head_length=0.4")
	if !ok {
		t.Fatal("ArrowStyleFromString(simple) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.QuadTo(geom.Pt{X: 50, Y: 60}, geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 1 || !parts[0].fillable {
		t.Fatalf("simple parts = %+v, want one fillable path", parts)
	}
	got := parts[0].path
	if len(got.V) < 7 {
		t.Fatalf("simple path vertices = %d, want curved outline vertices: %+v", len(got.V), got)
	}
	if maxPathYForPatchTest(got) < 20 {
		t.Fatalf("simple arrow ignored quadratic connection control point, got vertices %+v", got.V)
	}
	if !containsPointForPatchTest(got, geom.Pt{X: 100, Y: 0}) {
		t.Fatalf("simple arrow should keep the quadratic endpoint as tip, got %+v", got.V)
	}
}

func TestArrowStyleFancyFollowsQuadraticConnection(t *testing.T) {
	style, ok := ArrowStyleFromString("fancy,tail_width=0.4,head_width=0.8,head_length=0.4")
	if !ok {
		t.Fatal("ArrowStyleFromString(fancy) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.QuadTo(geom.Pt{X: 50, Y: 60}, geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 1 || !parts[0].fillable {
		t.Fatalf("fancy parts = %+v, want one fillable path", parts)
	}
	got := parts[0].path
	if len(got.V) < 7 {
		t.Fatalf("fancy path vertices = %d, want curved outline vertices: %+v", len(got.V), got)
	}
	if maxPathYForPatchTest(got) < 20 {
		t.Fatalf("fancy arrow ignored quadratic connection control point, got vertices %+v", got.V)
	}
	if !containsPointForPatchTest(got, geom.Pt{X: 100, Y: 0}) {
		t.Fatalf("fancy arrow should keep the quadratic endpoint as tip, got %+v", got.V)
	}
}

func TestArrowStyleCurveShortensLineForArrowHeads(t *testing.T) {
	style, ok := ArrowStyleFromString("->,head_length=0.4,head_width=0.2")
	if !ok {
		t.Fatal("ArrowStyleFromString(->) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 2 {
		t.Fatalf("curve arrow parts = %d, want line plus head", len(parts))
	}
	line := parts[0].path
	if len(line.V) != 2 {
		t.Fatalf("curve line vertices = %+v, want straight shortened line", line.V)
	}
	headLength := style.HeadLength * 10
	headWidth := style.HeadWidth * 10
	headDist := math.Hypot(headLength, headWidth)
	padProjected := 0.5 * 1 / (headWidth / headDist)
	wantTip := geom.Pt{X: 100 - padProjected, Y: 0}
	if !approx(line.V[1].X, wantTip.X, 1e-9) || !approx(line.V[1].Y, wantTip.Y, 1e-9) {
		t.Fatalf("curve line end = %+v, want linewidth-projected end %+v", line.V[1], wantTip)
	}
	head := parts[1].path
	if !containsPointForPatchTest(head, wantTip) {
		t.Fatalf("arrow head should use linewidth-projected tip %+v, got %+v", wantTip, head.V)
	}
}

func TestArrowStyleBarABUsesZeroBracketLength(t *testing.T) {
	style, ok := ArrowStyleFromString("|-|")
	if !ok {
		t.Fatal("ArrowStyleFromString(|-|) returned !ok")
	}
	if style.LengthA != 0 || style.LengthB != 0 {
		t.Fatalf("|-| bracket lengths = %v/%v, want zero-length bars", style.LengthA, style.LengthB)
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 3 {
		t.Fatalf("|-| parts = %d, want line plus two bar brackets", len(parts))
	}
	begin := parts[1].path
	end := parts[2].path
	for _, pt := range begin.V {
		if !approx(pt.X, 0, 1e-9) {
			t.Fatalf("begin bar protrudes from anchor: %+v", begin.V)
		}
	}
	for _, pt := range end.V {
		if !approx(pt.X, 100, 1e-9) {
			t.Fatalf("end bar protrudes from anchor: %+v", end.V)
		}
	}
}

func TestArrowStyleBracketScaleOverridesMutationSize(t *testing.T) {
	style, ok := ArrowStyleFromString("]-,widthA=2,lengthA=1,scaleA=3")
	if !ok {
		t.Fatal("ArrowStyleFromString(]-) returned !ok")
	}

	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.LineTo(geom.Pt{X: 100, Y: 0})
	parts := style.transmute(path, 10, 1)
	if len(parts) != 2 {
		t.Fatalf("]- parts = %d, want line plus bracket", len(parts))
	}
	bounds, ok := pathBounds(parts[1].path)
	if !ok {
		t.Fatal("bracket has no bounds")
	}
	want := geom.Rect{
		Min: geom.Pt{X: -3, Y: -3},
		Max: geom.Pt{X: 0, Y: 3},
	}
	if !approxRect(bounds, want, 1e-9) {
		t.Fatalf("scaled bracket bounds = %+v, want %+v", bounds, want)
	}
}

func TestFancyArrowPatchDefaultShrinkMatchesMatplotlib(t *testing.T) {
	patch := &FancyArrowPatch{
		PosA:            geom.Pt{X: 0.25, Y: 0.5},
		PosB:            geom.Pt{X: 0.75, Y: 0.5},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
		Coords:          Coords(CoordAxes),
	}
	ctx := createTestDrawContext()
	ctx.RC.DPI = 144
	path := patch.displayPath(ctx)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) != 3 {
		t.Fatalf("default arc3 path = commands %+v vertices %+v, want quadratic path", path.C, path.V)
	}

	start := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.PosA)
	end := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.PosB)
	shrink := pointsToPixels(ctx.RC, 2)
	if !approx(path.V[0].X, start.X+shrink, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("default-shrunk start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + shrink, Y: start.Y})
	}
	if !approx(path.V[1].X, (start.X+end.X)/2, 1e-9) || !approx(path.V[1].Y, start.Y, 1e-9) {
		t.Fatalf("default arc3 control = %+v, want midpoint", path.V[1])
	}
	if !approx(path.V[2].X, end.X-shrink, 1e-9) || !approx(path.V[2].Y, end.Y, 1e-9) {
		t.Fatalf("default-shrunk end = %+v, want %+v", path.V[2], geom.Pt{X: end.X - shrink, Y: end.Y})
	}
}

func TestShrinkPathEndpointsSplitsQuadraticSegment(t *testing.T) {
	path := geom.Path{}
	path.MoveTo(geom.Pt{X: 0, Y: 0})
	path.QuadTo(geom.Pt{X: 50, Y: 80}, geom.Pt{X: 100, Y: 0})

	got := shrinkPathEndpoints(path, 10, 0)

	if len(got.V) != 3 {
		t.Fatalf("shrunk quadratic vertices = %+v, want 3 vertices", got.V)
	}
	t0 := quadraticDistanceT(path.V[0], path.V[1], path.V[2], path.V[0], 10, true)
	wantStart := quadraticPoint(path.V[0], path.V[1], path.V[2], t0)
	wantCtrl := lerpPoint(path.V[1], path.V[2], t0)
	if distance(got.V[0], wantStart) > 1e-9 || distance(got.V[1], wantCtrl) > 1e-9 {
		t.Fatalf("shrunk quadratic = %+v, want start %+v control %+v", got.V, wantStart, wantCtrl)
	}
}

func TestConnectionPatchShrinkUsesPointUnits(t *testing.T) {
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
			ShrinkA:         1.5,
			ShrinkB:         3,
		},
		XYA:     geom.Pt{X: 0.25, Y: 0.5},
		XYB:     geom.Pt{X: 0.75, Y: 0.5},
		CoordsA: Coords(CoordAxes),
		CoordsB: Coords(CoordAxes),
	}
	ctx := createTestDrawContext()
	ctx.RC.DPI = 144

	path := patch.connectionDisplayPath(ctx)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) != 3 {
		t.Fatalf("connection path = commands %+v vertices %+v, want quadratic path", path.C, path.V)
	}
	start := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYA)
	end := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYB)
	shrinkA := pointsToPixels(ctx.RC, patch.ShrinkA)
	shrinkB := pointsToPixels(ctx.RC, patch.ShrinkB)
	if !approx(path.V[0].X, start.X+shrinkA, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("connection start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + shrinkA, Y: start.Y})
	}
	if !approx(path.V[2].X, end.X-shrinkB, 1e-9) || !approx(path.V[2].Y, end.Y, 1e-9) {
		t.Fatalf("connection end = %+v, want %+v", path.V[2], geom.Pt{X: end.X - shrinkB, Y: end.Y})
	}
}

func TestConnectionPatchClipsEndpointsToPatches(t *testing.T) {
	patchA := &Rectangle{
		XY:     geom.Pt{X: 0.5, Y: 4.5},
		Width:  1,
		Height: 1,
		Coords: Coords(CoordData),
	}
	patchB := &Rectangle{
		XY:     geom.Pt{X: 8.5, Y: 4.5},
		Width:  1,
		Height: 1,
		Coords: Coords(CoordData),
	}
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
			PatchA:          patchA,
			PatchB:          patchB,
		},
		XYA:     geom.Pt{X: 1, Y: 5},
		XYB:     geom.Pt{X: 9, Y: 5},
		CoordsA: Coords(CoordData),
		CoordsB: Coords(CoordData),
	}
	ctx := createTestDrawContext()

	path := patch.connectionDisplayPath(ctx)
	if len(path.C) != 2 || path.C[0] != geom.MoveTo || path.C[1] != geom.QuadTo || len(path.V) != 3 {
		t.Fatalf("connection path = commands %+v vertices %+v, want quadratic path", path.C, path.V)
	}
	wantStart := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 1.5, Y: 5})
	wantEnd := ctx.TransformFor(Coords(CoordData)).Apply(geom.Pt{X: 8.5, Y: 5})
	if !approxPt(path.V[0], wantStart, 1e-9) {
		t.Fatalf("patch-clipped start = %+v, want %+v", path.V[0], wantStart)
	}
	if !approxPt(path.V[2], wantEnd, 1e-9) {
		t.Fatalf("patch-clipped end = %+v, want %+v", path.V[2], wantEnd)
	}
}

func TestConnectionPatchResolvesIndependentCoordinateSpaces(t *testing.T) {
	patch := &ConnectionPatch{
		FancyArrowPatch: FancyArrowPatch{
			Patch: Patch{
				EdgeColor: render.Color{R: 0, G: 0, B: 0, A: 1},
				EdgeWidth: 1,
			},
			ArrowStyle:      ArrowStyle{Name: "-", HeadLength: 0.2, HeadWidth: 0.1},
			ConnectionStyle: ConnectionStyle{Name: "arc3"},
		},
		XYA:     geom.Pt{X: 0.25, Y: 0.75},
		XYB:     geom.Pt{X: 0.5, Y: 0.5},
		CoordsA: Coords(CoordData),
		CoordsB: Coords(CoordAxes),
	}

	ctx := createTestDrawContext()
	r := &recordingRenderer{}
	patch.Draw(r, ctx)

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one connection path, got %d", len(r.pathCalls))
	}
	got := r.pathCalls[0].path
	wantA := ctx.TransformFor(Coords(CoordData)).Apply(patch.XYA)
	wantB := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.XYB)
	if len(got.C) != 2 || got.C[0] != geom.MoveTo || got.C[1] != geom.QuadTo || len(got.V) != 3 {
		t.Fatalf("expected straight quadratic arc3 path, got commands %+v vertices %+v", got.C, got.V)
	}
	if !approx(got.V[0].X, wantA.X, 1e-9) || !approx(got.V[0].Y, wantA.Y, 1e-9) {
		t.Fatalf("connection start = %+v, want %+v", got.V[0], wantA)
	}
	if !approx(got.V[2].X, wantB.X, 1e-9) || !approx(got.V[2].Y, wantB.Y, 1e-9) {
		t.Fatalf("connection end = %+v, want %+v", got.V[2], wantB)
	}
}

func TestAdditionalPatchClassesDrawExpectedPaths(t *testing.T) {
	ctx := createTestDrawContext()
	patches := []Artist{
		&RegularPolygon{
			Patch:       Patch{FaceColor: render.Color{A: 1}},
			Center:      geom.Pt{X: 2, Y: 2},
			NumVertices: 5,
			Radius:      1,
		},
		&CirclePolygon{
			Patch:      Patch{FaceColor: render.Color{A: 1}},
			Center:     geom.Pt{X: 2, Y: 2},
			Radius:     1,
			Resolution: 12,
		},
		&Arc{
			Patch:    Patch{EdgeColor: render.Color{A: 1}, EdgeWidth: 1},
			Center:   geom.Pt{X: 2, Y: 2},
			Width:    2,
			Height:   1,
			Theta1:   10,
			Theta2:   220,
			EdgeOnly: true,
		},
		&Annulus{
			Patch:   Patch{FaceColor: render.Color{A: 1}},
			Center:  geom.Pt{X: 2, Y: 2},
			RadiusA: 1.5,
			RadiusB: 1.0,
			Width:   0.3,
			Angle:   15,
		},
		&StepPatch{
			Patch:    Patch{FaceColor: render.Color{A: 1}, EdgeColor: render.Color{A: 1}, EdgeWidth: 1},
			Values:   []float64{1, 2, 1.5},
			Edges:    []float64{0, 1, 2, 3},
			Baseline: float64Ptr(0),
		},
	}

	for _, patch := range patches {
		r := &recordingRenderer{}
		patch.Draw(r, ctx)
		if len(r.pathCalls) == 0 {
			t.Fatalf("%T did not draw a path", patch)
		}
	}
}

func TestShadowOffsetsAndDarkensSourcePatch(t *testing.T) {
	source := &Rectangle{
		Patch: Patch{
			FaceColor: render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1},
			EdgeColor: render.Color{R: 0.2, G: 0.1, B: 0.1, A: 1},
			EdgeWidth: 1,
		},
		XY:     geom.Pt{X: 1, Y: 2},
		Width:  2,
		Height: 1,
	}
	shadow := &Shadow{Patch: Patch{}, Source: source, Offset: geom.Pt{X: 0.5, Y: -0.25}, Shade: 0.7}

	r := &recordingRenderer{}
	shadow.Draw(r, createTestDrawContext())

	if len(r.pathCalls) != 1 {
		t.Fatalf("expected one shadow path, got %d", len(r.pathCalls))
	}
	paint := r.pathCalls[0].paint
	if paint.Fill.A != 0.5 {
		t.Fatalf("shadow alpha = %v, want 0.5", paint.Fill.A)
	}
	if paint.Fill.R >= source.FaceColor.R {
		t.Fatalf("shadow face color was not darkened: got %+v source %+v", paint.Fill, source.FaceColor)
	}
}

// TestNormalForVectorIsYUpCounterClockwisePerpendicular locks the arrow-head
// normal to Matplotlib's y-up convention: the perpendicular is the +90° (CCW)
// rotation of the direction, so a rightward shaft (+x) yields an upward (+y)
// normal. Under the old y-down display space this normal pointed the opposite
// way and arrow heads splayed on the wrong side.
func TestNormalForVectorIsYUpCounterClockwisePerpendicular(t *testing.T) {
	cases := []struct {
		v, want geom.Pt
	}{
		{geom.Pt{X: 1, Y: 0}, geom.Pt{X: 0, Y: 1}},   // +x -> +y (up)
		{geom.Pt{X: 0, Y: 1}, geom.Pt{X: -1, Y: 0}},  // +y -> -x
		{geom.Pt{X: -1, Y: 0}, geom.Pt{X: 0, Y: -1}}, // -x -> -y
		{geom.Pt{X: 3, Y: 4}, geom.Pt{X: -4.0 / 5, Y: 3.0 / 5}},
		{geom.Pt{X: 0, Y: 0}, geom.Pt{X: 0, Y: 0}}, // degenerate
	}
	for _, c := range cases {
		if got := normalForVector(c.v); !approxPt(got, c.want, 1e-9) {
			t.Fatalf("normalForVector(%+v) = %+v, want CCW perpendicular %+v", c.v, got, c.want)
		}
	}
}

// TestArrowHeadPathPlacesBaseCornersPerpendicularInYUp asserts that a rightward
// arrow head keeps its base behind the tip along the shaft and straddles the
// shaft with the left corner above (+y) and the right corner below (-y), the
// y-up orientation Matplotlib draws.
func TestArrowHeadPathPlacesBaseCornersPerpendicularInYUp(t *testing.T) {
	const headLength, headWidth = 4.0, 4.0
	path := arrowHeadPath(geom.Pt{X: 0, Y: 0}, geom.Pt{X: 10, Y: 0}, headLength, headWidth, true, 0)
	if len(path.V) < 3 {
		t.Fatalf("arrow head path vertices = %+v, want at least tip/left/right", path.V)
	}
	tip, left, right := path.V[0], path.V[1], path.V[2]
	if !approxPt(tip, geom.Pt{X: 10, Y: 0}, 1e-9) {
		t.Fatalf("arrow head tip = %+v, want {10,0}", tip)
	}
	if !approxPt(left, geom.Pt{X: 6, Y: 4}, 1e-9) {
		t.Fatalf("arrow head left base = %+v, want {6,4} (behind tip, above shaft)", left)
	}
	if !approxPt(right, geom.Pt{X: 6, Y: -4}, 1e-9) {
		t.Fatalf("arrow head right base = %+v, want {6,-4} (behind tip, below shaft)", right)
	}
}

// TestRotationAffineRotatesCounterClockwiseInYUp pins the rotation used for
// rotated text and its bbox to Matplotlib's CCW-positive convention in y-up
// display space: a positive angle moves a point on a box's right edge toward
// the top edge. Under y-down this rotation ran clockwise and rotated labels
// tilted the wrong way.
func TestRotationAffineRotatesCounterClockwiseInYUp(t *testing.T) {
	rot := rotationAffine(90)
	if got := rot.Apply(geom.Pt{X: 1, Y: 0}); !approxPt(got, geom.Pt{X: 0, Y: 1}, 1e-9) {
		t.Fatalf("rotationAffine(90).Apply({1,0}) = %+v, want {0,1} (CCW: right -> up)", got)
	}

	// Rotate the corners of a unit box about its center by +90°. The corner on
	// the right edge must land on the top edge (CCW), not the bottom.
	box := geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 2, Y: 2}}
	center := rectCenter(box)
	toCenter := translateAffine(geom.Pt{X: -center.X, Y: -center.Y})
	fromCenter := translateAffine(center)
	aboutCenter := fromCenter.Mul(rotationAffine(90)).Mul(toCenter)

	rightEdgeMid := geom.Pt{X: box.Max.X, Y: center.Y}   // {2,1}
	wantTopEdgeMid := geom.Pt{X: center.X, Y: box.Max.Y} // {1,2}
	if got := aboutCenter.Apply(rightEdgeMid); !approxPt(got, wantTopEdgeMid, 1e-9) {
		t.Fatalf("CCW box rotation mapped right edge %+v to %+v, want top edge %+v", rightEdgeMid, got, wantTopEdgeMid)
	}
}
