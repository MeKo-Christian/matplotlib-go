package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/ticker"
	"github.com/cwbudde/matplotlib-go/transform"
)

func TestSnapDisplayYMatchesMatplotlibDeviceSpacePathSnapper(t *testing.T) {
	if got := snapDisplayY(268.5, 900); got != 267.5 {
		t.Fatalf("snapDisplayY(268.5, 900) = %.17g, want device-space snapped 267.5", got)
	}
	if got := snapDisplayY(67.49999999999996, 900); got != 66.5 {
		t.Fatalf("snapDisplayY(67.49999999999996, 900) = %.17g, want device-space snapped 66.5", got)
	}
	if got := snapDisplayY(373.50000000000006, 720); got != 373.5 {
		t.Fatalf("snapDisplayY(373.50000000000006, 720) = %.17g, want device-space snapped 373.5", got)
	}
}

func TestSpinePixelEndpointsRightBoundaryUsesMatplotlibPathSnapper(t *testing.T) {
	px := geom.Rect{
		Min: geom.Pt{X: 51.06977777777778, Y: 47.444777777777794},
		Max: geom.Pt{X: 846.4725805555556, Y: 673.4996666666666},
	}

	p1, p2 := spinePixelEndpoints(AxisRight, px)
	if p1.X != 846.5 || p2.X != 846.5 {
		t.Fatalf("right spine x = %v..%v, want Matplotlib pixel center 846.5", p1.X, p2.X)
	}
}

func TestSpinePixelEndpointsRightBoundaryRoundsPastHalfPixel(t *testing.T) {
	px := geom.Rect{
		Min: geom.Pt{X: 76.8, Y: 57.6},
		Max: geom.Pt{X: 588.8, Y: 316.8},
	}

	p1, p2 := spinePixelEndpoints(AxisRight, px)
	if p1.X != 589.5 || p2.X != 589.5 {
		t.Fatalf("right spine x = %v..%v, want Matplotlib pixel center 589.5", p1.X, p2.X)
	}
}

func TestSpinePixelEndpointsHorizontalBoundariesUseDeviceSpaceSnap(t *testing.T) {
	ctx := &DrawContext{FigureRect: geom.Rect{
		Min: geom.Pt{},
		Max: geom.Pt{X: 1320, Y: 900},
	}}
	px := geom.Rect{
		Min: geom.Pt{X: 499.4, Y: 67.50000000000001},
		Max: geom.Pt{X: 847.0, Y: 268.5},
	}

	b1, b2 := spinePixelEndpoints(AxisBottom, px, ctx)
	if b1.Y != 66.5 || b2.Y != 66.5 {
		t.Fatalf("bottom spine y = %v..%v, want Matplotlib device-space center 66.5", b1.Y, b2.Y)
	}

	t1, t2 := spinePixelEndpoints(AxisTop, px, ctx)
	if t1.Y != 267.5 || t2.Y != 267.5 {
		t.Fatalf("top spine y = %v..%v, want Matplotlib device-space center 267.5", t1.Y, t2.Y)
	}
}

func TestDrawFrameUsesDeviceSpaceSnapForFallbackTopSpine(t *testing.T) {
	ctx := &DrawContext{
		Clip: geom.Rect{
			Min: geom.Pt{X: 499.4, Y: 67.50000000000001},
			Max: geom.Pt{X: 847.0, Y: 268.5},
		},
		FigureRect: geom.Rect{
			Min: geom.Pt{},
			Max: geom.Pt{X: 1320, Y: 900},
		},
	}
	axis := NewXAxis()
	r := &recordingRenderer{}

	DrawFrame(r, ctx, axis, true, false)

	if len(r.pathCalls) != 1 {
		t.Fatalf("frame path calls = %d, want 1", len(r.pathCalls))
	}
	got := r.pathCalls[0].path.V
	if len(got) != 2 || got[0].Y != 267.5 || got[1].Y != 267.5 {
		t.Fatalf("fallback top spine vertices = %+v, want y=267.5", got)
	}
}

func TestDrawFrameVisibilityIsIndependentFromReferenceSpine(t *testing.T) {
	ctx := &DrawContext{
		Clip: geom.Rect{Max: geom.Pt{X: 100, Y: 80}},
	}
	axis := NewXAxis()
	axis.ShowSpine = false
	r := &recordingRenderer{}

	DrawFrame(r, ctx, axis, true, true)

	if len(r.pathCalls) != 2 {
		t.Fatalf("frame path calls = %d, want independent top and right spines", len(r.pathCalls))
	}
}

func TestAxes_TickParamsGridStyling(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), YAxis: NewYAxis()}
	xGrid := axes.AddXGrid()
	yGrid := axes.AddYGrid()
	xGrid.Minor = true
	yGrid.Minor = true

	visible := false
	gridColor := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.35
	width := 2.5
	dashes := []float64{4, 2}
	if err := axes.TickParams(TickParams{
		Axis:          "x",
		Which:         "both",
		GridVisible:   &visible,
		GridColor:     &gridColor,
		GridAlpha:     &alpha,
		GridLineWidth: &width,
		GridDashes:    dashes,
	}); err != nil {
		t.Fatalf("TickParams(grid): %v", err)
	}

	if xGrid.Major || xGrid.Minor {
		t.Fatalf("x grid visibility = major %v minor %v, want both false", xGrid.Major, xGrid.Minor)
	}
	if xGrid.Color != gridColor || xGrid.MinorColor.R != gridColor.R || xGrid.MinorColor.A != alpha {
		t.Fatalf("x grid colors = major %+v minor %+v", xGrid.Color, xGrid.MinorColor)
	}
	if xGrid.Alpha != alpha || xGrid.LineWidth != width || xGrid.MinorLineWidth != width {
		t.Fatalf("x grid style = alpha %v width %v minor width %v", xGrid.Alpha, xGrid.LineWidth, xGrid.MinorLineWidth)
	}
	if len(xGrid.Dashes) != 2 || xGrid.Dashes[0] != 4 || len(xGrid.MinorDashes) != 2 || xGrid.MinorDashes[1] != 2 {
		t.Fatalf("x grid dashes = major %v minor %v", xGrid.Dashes, xGrid.MinorDashes)
	}
	if !yGrid.Major || !yGrid.Minor {
		t.Fatalf("y grid should be unchanged, got major %v minor %v", yGrid.Major, yGrid.Minor)
	}
}

func TestAxes_SetAxisLineStyleAppliesToSelectedAxes(t *testing.T) {
	axes := &Axes{XAxis: NewXAxis(), XAxisTop: NewXAxis(), YAxis: NewYAxis()}

	if err := axes.SetAxisLineStyle("x", render.CapRound, render.JoinBevel, 3, 2); err != nil {
		t.Fatalf("SetAxisLineStyle(x): %v", err)
	}

	if axes.XAxis.LineCap != render.CapRound || axes.XAxis.LineJoin != render.JoinBevel {
		t.Fatalf("bottom axis style = cap %v join %v", axes.XAxis.LineCap, axes.XAxis.LineJoin)
	}
	if axes.XAxisTop.LineCap != render.CapRound || axes.XAxisTop.LineJoin != render.JoinBevel {
		t.Fatalf("top axis style = cap %v join %v", axes.XAxisTop.LineCap, axes.XAxisTop.LineJoin)
	}
	if len(axes.XAxis.Dashes) != 2 || axes.XAxis.Dashes[0] != 3 || axes.XAxis.Dashes[1] != 2 {
		t.Fatalf("bottom axis dashes = %v", axes.XAxis.Dashes)
	}
	if axes.YAxis.LineCap != render.CapSquare {
		t.Fatalf("y axis should be unchanged, got cap %v", axes.YAxis.LineCap)
	}
}

func TestAxisSetLineStyleAffectsSpineAndTickPaint(t *testing.T) {
	axis := NewXAxis()
	axis.Locator = staticLocator{2}
	axis.Formatter = nil
	axis.SetLineStyle(render.CapRound, render.JoinBevel, 4, 1)

	ctx := createTestDrawContext()
	r := &recordingRenderer{}

	axis.Draw(r, ctx)
	axis.DrawTicks(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("expected spine and tick path calls, got %d", len(r.pathCalls))
	}
	if r.pathCalls[0].paint.LineCap != render.CapRound || r.pathCalls[0].paint.LineJoin != render.JoinBevel {
		t.Fatalf("spine paint = %+v", r.pathCalls[0].paint)
	}
	if len(r.pathCalls[0].paint.Dashes) != 2 {
		t.Fatalf("spine dashes = %v", r.pathCalls[0].paint.Dashes)
	}
	if r.pathCalls[1].paint.LineCap != render.CapRound || r.pathCalls[1].paint.LineJoin != render.JoinBevel {
		t.Fatalf("tick paint = %+v", r.pathCalls[1].paint)
	}
}

func TestGrid_Draw(t *testing.T) {
	// Test grid drawing
	grid := NewGrid(AxisBottom)

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_UsesOwningAxisMajorLocator(t *testing.T) {
	grid := NewGrid(AxisLeft)
	ctx := createTestDrawContext()
	ctx.DataToPixel.YScale = transform.NewLinear(0, 80)
	ctx.Axes = &Axes{
		XAxis: NewXAxis(),
		YAxis: NewYAxis(),
	}
	ctx.Axes.YAxis.Locator = ticker.FixedLocator{TicksList: []float64{0, 20, 40, 60, 80}}

	renderer := &gridRecordingRenderer{}
	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()

	if got, want := len(renderer.paths), 5; got != want {
		t.Fatalf("grid path count = %d, want %d", got, want)
	}
}

func TestGrid_AutoMinorLocatorUsesOwningAxisMajorLocator(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Major = false
	grid.Minor = true
	ctx := createTestDrawContext()
	ctx.DataToPixel.XScale = transform.NewLinear(0, 6)
	ctx.Axes = &Axes{
		XAxis: NewXAxis(),
		YAxis: NewYAxis(),
	}
	ctx.Axes.XAxis.Locator = ticker.MultipleLocator{Base: 1.5}
	ctx.Axes.XAxis.MinorLocator = ticker.AutoMinorLocator{N: 3}

	renderer := &gridRecordingRenderer{}
	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()

	got := make([]float64, 0, len(renderer.paths))
	for _, path := range renderer.paths {
		if len(path.V) == 0 {
			continue
		}
		data, ok := ctx.DataToPixel.Invert(path.V[0])
		if !ok {
			t.Fatalf("invert grid path point %+v", path.V[0])
		}
		got = append(got, math.Round(data.X*2)/2)
	}
	want := []float64{0.5, 1, 2, 2.5, 3.5, 4, 5, 5.5}
	if len(got) != len(want) {
		t.Fatalf("minor grid ticks = %v, want %v", got, want)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("minor grid tick %d = %v, want %v (all %v)", i, got[i], want[i], got)
		}
	}
}

func TestGrid_Disabled(t *testing.T) {
	// Test grid with major disabled
	grid := NewGrid(AxisLeft)
	grid.Major = false

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not draw anything
	grid.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestGrid_MinorDraw(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Minor = true
	grid.MinorDashes = []float64{2, 3}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	// Should not panic with minor grid enabled
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_MinorOnlyDraw(t *testing.T) {
	grid := NewGrid(AxisLeft)
	grid.Major = false
	grid.Minor = true

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestGrid_CustomLocator(t *testing.T) {
	grid := NewGrid(AxisBottom)
	grid.Locator = ticker.LogLocator{Base: 10}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	_ = renderer.Begin(geom.Rect{})
	grid.Draw(renderer, ctx)
	_ = renderer.End()
}

func TestAxes_AddGrid(t *testing.T) {
	// Test adding grids to axes
	axes := &Axes{
		Artists: []Artist{},
		XAxis:   NewXAxis(),
		YAxis:   NewYAxis(),
	}

	initialCount := len(axes.Artists)

	// Add X grid
	xGrid := axes.AddXGrid()
	if len(axes.Artists) != initialCount+1 {
		t.Errorf("AddXGrid should add one artist, got %d artists", len(axes.Artists))
	}
	if xGrid.Axis != AxisBottom {
		t.Errorf("AddXGrid should create grid for AxisBottom, got %v", xGrid.Axis)
	}

	// Add Y grid
	yGrid := axes.AddYGrid()
	if len(axes.Artists) != initialCount+2 {
		t.Errorf("AddYGrid should add second artist, got %d artists", len(axes.Artists))
	}
	if yGrid.Axis != AxisLeft {
		t.Errorf("AddYGrid should create grid for AxisLeft, got %v", yGrid.Axis)
	}
}

func TestDrawFigure_ExplicitTopRightAxesSuppressFrameFallback(t *testing.T) {
	fig := NewFigure(240, 180)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})

	ax.XAxis.ShowTicks = false
	ax.XAxis.ShowLabels = false
	ax.YAxis.ShowTicks = false
	ax.YAxis.ShowLabels = false

	top := ax.TopAxis()
	top.ShowTicks = false
	top.ShowLabels = false

	right := ax.RightAxis()
	right.ShowTicks = false
	right.ShowLabels = false

	r := &recordingRenderer{}
	DrawFigure(fig, r)

	if got := len(r.pathCalls); got != 4 {
		t.Fatalf("expected exactly four spine paths with explicit top/right axes, got %d", got)
	}
}
