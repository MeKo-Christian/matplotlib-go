package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
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
