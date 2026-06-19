package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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
