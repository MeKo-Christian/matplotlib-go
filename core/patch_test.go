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
			want:      geom.Rect{Min: geom.Pt{X: -1, Y: -0.5}, Max: geom.Pt{X: 5, Y: 2.5}},
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
		Min: geom.Pt{X: -3, Y: -2},
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
	if !approx(stretchedHead.Max.Y, 100, 1e-9) {
		t.Fatalf("mutation aspect moved arrow tip to y=%v, want 100", stretchedHead.Max.Y)
	}
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
	if !approx(line.V[1].X, 96, 1e-9) || !approx(line.V[1].Y, 0, 1e-9) {
		t.Fatalf("curve line end = %+v, want shortened to x=96", line.V[1])
	}
	head := parts[1].path
	if !containsPointForPatchTest(head, geom.Pt{X: 100, Y: 0}) {
		t.Fatalf("arrow head should still reach original tip, got %+v", head.V)
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

func TestFancyArrowPatchDefaultShrinkMatchesMatplotlib(t *testing.T) {
	patch := &FancyArrowPatch{
		PosA:            geom.Pt{X: 0.25, Y: 0.5},
		PosB:            geom.Pt{X: 0.75, Y: 0.5},
		ConnectionStyle: ConnectionStyle{Name: "arc3"},
		Coords:          Coords(CoordAxes),
	}
	ctx := createTestDrawContext()
	path := patch.displayPath(ctx)
	if len(path.V) != 2 {
		t.Fatalf("default arc3 path vertices = %d, want 2: %+v", len(path.V), path.V)
	}

	start := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.PosA)
	end := ctx.TransformFor(Coords(CoordAxes)).Apply(patch.PosB)
	if !approx(path.V[0].X, start.X+2, 1e-9) || !approx(path.V[0].Y, start.Y, 1e-9) {
		t.Fatalf("default-shrunk start = %+v, want %+v", path.V[0], geom.Pt{X: start.X + 2, Y: start.Y})
	}
	if !approx(path.V[1].X, end.X-2, 1e-9) || !approx(path.V[1].Y, end.Y, 1e-9) {
		t.Fatalf("default-shrunk end = %+v, want %+v", path.V[1], geom.Pt{X: end.X - 2, Y: end.Y})
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
	if len(got.V) != 2 {
		t.Fatalf("expected straight arc3 path vertices, got %d: %+v", len(got.V), got.V)
	}
	if !approx(got.V[0].X, wantA.X, 1e-9) || !approx(got.V[0].Y, wantA.Y, 1e-9) {
		t.Fatalf("connection start = %+v, want %+v", got.V[0], wantA)
	}
	if !approx(got.V[1].X, wantB.X, 1e-9) || !approx(got.V[1].Y, wantB.Y, 1e-9) {
		t.Fatalf("connection end = %+v, want %+v", got.V[1], wantB)
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
