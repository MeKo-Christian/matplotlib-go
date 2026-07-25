package plot3d

import (
	"sort"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAddAxes3DConfiguresProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	if got, want := ax.ProjectionName(), "3d"; got != want {
		t.Fatalf("projection name = %q, want %q", got, want)
	}
	xMin, xMax := ax.effectiveXScale().Domain()
	yMin, yMax := ax.effectiveYScale().Domain()
	if !approx(xMin, default3DViewMin, 1e-12) || !approx(xMax, default3DViewMax, 1e-12) ||
		!approx(yMin, default3DViewMin, 1e-12) || !approx(yMax, default3DViewMax, 1e-12) {
		t.Fatalf("3D view domain = x(%v,%v) y(%v,%v), want (%v,%v)", xMin, xMax, yMin, yMax, default3DViewMin, default3DViewMax)
	}
	layout := ax.adjustedLayout(fig)
	if !approx(layout.W(), layout.H(), 1e-12) {
		t.Fatalf("3D axes layout = %+v, want square active box", layout)
	}

	elev, azim, distance := ax.View()
	if !approx(elev, default3DElevationDeg, 1e-12) ||
		!approx(azim, default3DAzimuthDeg, 1e-12) ||
		distance != default3DDistance {
		t.Fatalf("View = (%v, %v, %v), want (%v, %v, %v)", elev, azim, distance, default3DElevationDeg, default3DAzimuthDeg, default3DDistance)
	}
}

func TestAxes3DProjectionPointDefaults(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	got := ax.ProjectPoint(1, 2, 3)
	if !approx(got.X, 1, 1e-12) || !approx(got.Y, 2, 1e-12) {
		t.Fatalf("ProjectPoint(1,2,3) = %+v, want {1 2}", got)
	}
}

func TestAxes3DProjectPointMatchesMatplotlibDefaultProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	got := ax.ProjectPoint(1, 1, 1)
	if !approx(got.X, 0.0783182204915425, 1e-12) ||
		!approx(got.Y, 0.04773147362601089, 1e-12) {
		t.Fatalf("ProjectPoint(1,1,1) = %+v, want Matplotlib default projection {0.0783182204915425 0.04773147362601089}", got)
	}
}

func TestAxes3DProjectPointMatchesMatplotlibBasicDataLimits(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe(x, y, z)
	ax.Surface(x, y, z)
	ax.Contour(x, y, z)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")

	got := ax.ProjectPoint(1, 1, 2)
	if !approx(got.X, 0.0711768607286225, 1e-12) ||
		!approx(got.Y, 0.043379132331248196, 1e-12) {
		t.Fatalf("ProjectPoint(1,1,2) with mplot3d_basic limits = %+v, want Matplotlib projection {0.0711768607286225 0.043379132331248196}", got)
	}
}

func TestAxes3DProjectionLimitsUseMatplotlibZMargin(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{-1, 1}, []float64{-2, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[0], -0.07291666666666667, 1e-12) || !approx(maxs[0], 1.0729166666666667, 1e-12) {
		t.Fatalf("x projection limits = (%v, %v), want Matplotlib data margin plus 3D view margin", mins[0], maxs[0])
	}
	if !approx(mins[1], -1.1458333333333335, 1e-12) || !approx(maxs[1], 1.1458333333333335, 1e-12) {
		t.Fatalf("y projection limits = (%v, %v), want Matplotlib data margin plus 3D view margin", mins[1], maxs[1])
	}
	if !approx(mins[2], -2.0833333333333335, 1e-12) || !approx(maxs[2], 2.0833333333333335, 1e-12) {
		t.Fatalf("line z projection limits = (%v, %v), want Matplotlib 3D view margin", mins[2], maxs[2])
	}
}

func TestAxes3DScatterProjectionLimitsUseMatplotlibZMargin(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Scatter3D([]float64{0, 1}, []float64{-1, 1}, []float64{-2, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[2], -2.291666666666667, 1e-12) || !approx(maxs[2], 2.291666666666667, 1e-12) {
		t.Fatalf("scatter z projection limits = (%v, %v), want Matplotlib scatter z margin plus 3D view margin", mins[2], maxs[2])
	}
}

func TestAxes3DProjectPointMatchesMatplotlibScatterFixtureLimits(t *testing.T) {
	fig := NewFigure(720, 560)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(22.5767530642105, 32.43415741224721)
	ax.SetYLim(0.09104692778428092, 104.58776816008752)
	ax.SetZLim(-51.007126259038195, -24.886181592223462)

	got := ax.ProjectPoint(29.444073116538434, 91.3534531817269, -37.52903158779476)
	if !approx(got.X, 0.04156475342576397, 1e-12) ||
		!approx(got.Y, 0.014307381250617935, 1e-12) {
		t.Fatalf("scatter fixture projection = %+v, want Matplotlib projection {0.04156475342576397 0.014307381250617935}", got)
	}
}

func TestAxes3DText3DDoesNotExpandDataLimitsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	text := ax.Text3D(9, 8, 7, "label")
	if text == nil {
		t.Fatal("Text3D returned nil")
	}
	mins, maxs := ax.projectionLimits()
	wantTextMins := vec3{-0.020833333333333332, -0.020833333333333332, -0.020833333333333332}
	wantTextMaxs := vec3{1.0208333333333333, 1.0208333333333333, 1.0208333333333333}
	if !approx(mins[0], wantTextMins[0], 1e-12) || !approx(mins[1], wantTextMins[1], 1e-12) || !approx(mins[2], wantTextMins[2], 1e-12) ||
		!approx(maxs[0], wantTextMaxs[0], 1e-12) || !approx(maxs[1], wantTextMaxs[1], 1e-12) || !approx(maxs[2], wantTextMaxs[2], 1e-12) {
		t.Fatalf("text-only projection limits = %v..%v, want Matplotlib default limits plus 3D view margin %v..%v", mins, maxs, wantTextMins, wantTextMaxs)
	}
	if ax.hasData {
		t.Fatal("Text3D marked the axes as having data; Matplotlib text does not update 3D data limits")
	}

	ax.Scatter3D([]float64{1, 9}, []float64{1, 9}, []float64{1, 9})
	mins, maxs = ax.projectionLimits()
	wantMins := vec3{0.41666666666666663, 0.41666666666666663, 0.41666666666666663}
	wantMaxs := vec3{9.583333333333334, 9.583333333333334, 9.583333333333334}
	if !approx(mins[0], wantMins[0], 1e-12) || !approx(mins[1], wantMins[1], 1e-12) || !approx(mins[2], wantMins[2], 1e-12) ||
		!approx(maxs[0], wantMaxs[0], 1e-12) || !approx(maxs[1], wantMaxs[1], 1e-12) || !approx(maxs[2], wantMaxs[2], 1e-12) {
		t.Fatalf("projection limits after scatter = %v..%v, want text ignored and scatter autoscaled to %v..%v", mins, maxs, wantMins, wantMaxs)
	}
	wantPos := ax.ProjectPoint(9, 8, 7)
	if !pointsEqual([]geom.Pt{text.Position}, []geom.Pt{wantPos}, 1e-12) {
		t.Fatalf("Text3D position after later scatter = %+v, want reprojected %+v", text.Position, wantPos)
	}
}

func TestAxes3DFrameHonorsExplicitLimitsWithoutData(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 10)
	ax.SetYLim(0, 10)
	ax.SetZLim(0, 10)

	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}
	(&axes3DFrame{axes: ax}).Draw(r, ctx)

	if !containsString(r.texts, "10") {
		t.Fatalf("3D frame tick labels = %v, want explicit 0..10 limits honored without data", r.texts)
	}
}

func TestAxes3DExplicitLimitsPreserveMatplotlibInvertedAxis(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetXLim(1, 0)
	mins, maxs := ax.projectionLimits()
	if mins[0] != 1 || maxs[0] != 0 {
		t.Fatalf("inverted x projection limits = (%v, %v), want caller order (1, 0)", mins[0], maxs[0])
	}
	got := ax.ProjectPoint(0.25, 0.5, 0.5)
	want := project3DPointWithLimits(0.25, 0.5, 0.5, ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs, ax.projectionState())
	if !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("inverted x ProjectPoint = %+v, want projection with caller-order limits %+v", got, want)
	}
}

func TestAxes3DAxLimClipUsesNumericRangeForInvertedLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetXLim(1, 0)
	scatter := ax.Scatter3D(
		[]float64{0.25, 1.25},
		[]float64{0.5, 0.5},
		[]float64{0.5, 0.5},
		ScatterOptions{AxLimClip: true},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.XY), 1; got != want {
		t.Fatalf("Scatter3D clipped markers with inverted xlim = %d, want %d", got, want)
	}
	want := ax.ProjectPoint(0.25, 0.5, 0.5)
	if got := scatter.XY[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("Scatter3D clipped marker with inverted xlim = %+v, want %+v", got, want)
	}
}

func TestAxes3DProjectedDataMapsToMatplotlibDisplayCoordinates(t *testing.T) {
	fig := NewFigure(720, 560)
	ax, err := AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.16},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	// Display space is y-up: our display Y is the flip of Matplotlib's top-down
	// pixel (H=560), so 560 - 233.38993552 = 326.61006448.
	got := ctx.DataToPixel.Apply(geom.Pt{X: 0.039937290348120026, Y: 0.013747177404714107})
	if !approx(got.X, 452.49035388, 1e-8) ||
		!approx(got.Y, 326.61006448, 1e-8) {
		t.Fatalf("projected display point = %+v, want y-up display point {452.49035388 326.61006448}", got)
	}
}

func TestAxes3DPlot3DUsesProjectedCoordinates(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	line := ax.Plot3D([]float64{0, 1}, []float64{0, 0}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}
	if got, want := len(line.XY), 2; got != want {
		t.Fatalf("projected points = %d, want %d", got, want)
	}
}

func TestAxes3DPlot3DPropagatesColorAndAlphaLikeLine2D(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	alpha := 0.35
	line := ax.Plot3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		PlotOptions{Color: &color, Alpha: &alpha},
	)
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}
	want := color
	want.A = alpha
	if got := line.Col; got != want {
		t.Fatalf("Plot3D color = %+v, want Line2D color with alpha %+v", got, want)
	}
	if _, ok := any(line).(interface{ Array() []float64 }); ok {
		t.Fatal("Plot3D returned scalar-array capable artist, want Line2D-style explicit color artist")
	}
	if _, ok := any(line).(interface{ ScalarMap() ScalarMapInfo }); ok {
		t.Fatal("Plot3D returned scalar-mappable artist, want Line2D-style explicit color artist")
	}
}

func TestAxes3DPlot3DAxLimClipDropsOutsidePoints(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1)

	line := ax.Plot3D(
		[]float64{0, 0.5, 2},
		[]float64{0, 0, 0},
		[]float64{0, 0, 0},
		PlotOptions{AxLimClip: true},
	)
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}
	if got, want := len(line.XY), 2; got != want {
		t.Fatalf("Plot3D clipped points = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0, 0, 0),
		ax.ProjectPoint(0.5, 0, 0),
	}
	if !pointsEqual(line.XY, want, 1e-12) {
		t.Fatalf("Plot3D clipped XY = %+v, want %+v", line.XY, want)
	}
}

func TestAxes3DReprojectsExistingArtistsWhenDataLimitsExpand(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	if line == nil {
		t.Fatal("Plot3D returned nil")
	}

	ax.Wireframe(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{
			{0, 2},
			{0, 2},
		},
	)

	got := line.XY[1]
	if !approx(got.X, 0.06981276096054631, 1e-12) ||
		!approx(got.Y, 0.009353136460382655, 1e-12) {
		t.Fatalf("reprojected line endpoint = %+v, want Matplotlib projection with autoscale margins", got)
	}
}

func TestAxes3DSetViewReprojectsExistingArtistsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	if line == nil || len(line.XY) == 0 {
		t.Fatal("Plot3D returned no line points")
	}
	before := line.XY[0]
	ax.SetView(60, 30)
	want := ax.ProjectPoint(0, 0, 0)
	if got := line.XY[0]; got == before || !pointsEqual([]Pt{got}, []Pt{want}, 1e-12) {
		t.Fatalf("line first point after SetView = %+v, before %+v, want reprojected point %+v", got, before, want)
	}
}

func TestAxes3DSetViewResortsReprojectedArtists(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	// This concrete geometry set was validated to produce a post-view depth-order
	// change in mixed 3D collections before resorting.
	line := ax.Plot3D(
		[]float64{0.39189601719238726, -1.908212056724124, 0.19907234935342766},
		[]float64{0.2639219137021458, -2.7281573716616, -1.0175265380463943},
		[]float64{-0.3983193607443909, -0.3863074357062186, 0.045714226799383334},
	)
	scatter := ax.Scatter3D(
		[]float64{-0.12342427959032432, -0.8461028562161538, -0.3297070194526802},
		[]float64{-0.20409987823742703, -0.3647173958208699, 0.42341190997306155},
		[]float64{0.3650874803042903, -0.11731873114148794, 0.2924595905237891},
	)
	surface := ax.Surface(
		[]float64{-0.3919689221384559, -1.4448104140971734, -1.1327180161636174},
		[]float64{1.1560797832875775, -1.2420708801235487, 0.08422976168325602},
		[][]float64{
			{0.0068234861782554, -1.9896369601280388, -0.20234936763553057},
			{0.8622284518717414, 0.05819260888944283, 0.5210971150122212},
			{0.322145795095663, -0.49739907363322877, -0.34695595800546974},
		},
	)
	if line == nil || scatter == nil || surface == nil {
		t.Fatalf("expected mixed 3D artists, got line=%v scatter=%v surface=%v", line, scatter, surface)
	}

	// Establish baseline sorted draw order.
	DrawFigure(fig, &render.NullRenderer{})

	// This update changes object depths and should reorder artists.
	ax.SetView(0, 0)
	DrawFigure(fig, &render.NullRenderer{})
	got := dataArtistsForOrderCheck(sortedArtistDrawOrder(ax.Artists))
	want := dataArtistsForOrderCheck(ax.Artists)
	want = append([]Artist(nil), want...)
	sort.SliceStable(want, func(i, j int) bool {
		zi, zj := want[i].Z(), want[j].Z()
		if zi == zj {
			return i < j
		}
		return zi < zj
	})
	if !axes3DArtistsEqual(got, want) {
		t.Fatalf("mixed 3D artist draw order not resorted after view change; got=%s, want=%s", artistOrderLabel(got), artistOrderLabel(want))
	}

	_ = line
	_ = scatter
	_ = surface
}

func TestAxes3DSetViewInitAppliesRollAndVerticalAxis(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	before := ax.ProjectPoint(1, 1, 0)
	if err := ax.SetViewInit(30, -60, 17, "x"); err != nil {
		t.Fatalf("SetViewInit: %v", err)
	}
	after := ax.ProjectPoint(1, 1, 0)
	if approx(before.X, after.X, 1e-12) && approx(before.Y, after.Y, 1e-12) {
		t.Fatalf("projected point did not move after SetViewInit with roll/vertical axis change: before=%+v after=%+v", before, after)
	}
	if err := ax.SetViewInit(0, 0, 0, "not-an-axis"); err == nil {
		t.Fatal("SetViewInit with invalid axis: got nil, want error")
	}
}

func TestAxes3DSetBoxAspect3DReprojectsAndValidatesZoom(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	before := ax.ProjectPoint(0.8, 0.2, 0.6)
	if err := ax.SetBoxAspect3D([3]float64{2, 1, 3}, 1.25); err != nil {
		t.Fatalf("SetBoxAspect3D: %v", err)
	}
	after := ax.ProjectPoint(0.8, 0.2, 0.6)
	if approx(before.X, after.X, 1e-12) && approx(before.Y, after.Y, 1e-12) {
		t.Fatalf("projected point did not move after SetBoxAspect3D: before=%+v after=%+v", before, after)
	}

	mins, maxs := ax.projectionLimits()
	expected := project3DPointWithLimits(0.8, 0.2, 0.6, ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs, ax.projectionState())
	if !approx(expected.X, after.X, 1e-12) || !approx(expected.Y, after.Y, 1e-12) {
		t.Fatalf("reprojected point after SetBoxAspect3D = %+v, want %+v", after, expected)
	}
	if err := ax.SetBoxAspect3D([3]float64{1, 1, 1}, 0); err == nil {
		t.Fatal("SetBoxAspect3D with zoom=0: got nil, want error")
	}
}

func TestAxes3DSetProjectionTypeMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	if got := ax.ProjectionType(); got != "persp" {
		t.Fatalf("default projection type = %q, want persp", got)
	}
	if err := ax.SetProjectionType("ortho"); err != nil {
		t.Fatalf("SetProjectionType(ortho): %v", err)
	}
	if got := ax.ProjectionType(); got != "ortho" {
		t.Fatalf("projection type = %q, want ortho", got)
	}
	got := ax.ProjectPoint(1, 1, 1)
	if !approx(got.X, 0.07805859468646568, 1e-12) ||
		!approx(got.Y, 0.04757324323990303, 1e-12) {
		t.Fatalf("orthographic ProjectPoint(1,1,1) = %+v, want Matplotlib set_proj_type('ortho') projection", got)
	}

	if err := ax.SetProjectionType("persp", 2); err != nil {
		t.Fatalf("SetProjectionType(persp, 2): %v", err)
	}
	got = ax.ProjectPoint(1, 1, 1)
	if !approx(got.X, 0.0781881920661384, 1e-12) ||
		!approx(got.Y, 0.047652227081351764, 1e-12) {
		t.Fatalf("focal-length ProjectPoint(1,1,1) = %+v, want Matplotlib set_proj_type('persp', focal_length=2) projection", got)
	}

	if err := ax.SetProjectionType("persp", 0); err == nil {
		t.Fatal("SetProjectionType(persp, 0): got nil, want error")
	}
	if err := ax.SetProjectionType("ortho", 1); err == nil {
		t.Fatal("SetProjectionType(ortho, 1): got nil, want error")
	}
	if err := ax.SetProjectionType("bad"); err == nil {
		t.Fatal("SetProjectionType(bad): got nil, want error")
	}
}

func TestAxes3DSetDefaultsResetsViewAspectAndProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	if err := ax.SetViewInit(5, 15, 25, "x"); err != nil {
		t.Fatalf("SetViewInit: %v", err)
	}
	if err := ax.SetProjectionType("ortho"); err != nil {
		t.Fatalf("SetProjectionType(ortho): %v", err)
	}
	if err := ax.SetBoxAspect3D([3]float64{2, 1, 4}, 1.5); err != nil {
		t.Fatalf("SetBoxAspect3D: %v", err)
	}
	ax.SetDistance(7)

	ax.SetDefaults()
	elev, azim, distance := ax.View()
	if !approx(elev, default3DElevationDeg, 1e-12) ||
		!approx(azim, default3DAzimuthDeg, 1e-12) ||
		!approx(distance, default3DDistance, 1e-12) {
		t.Fatalf("View after SetDefaults = (%v, %v, %v), want Matplotlib defaults", elev, azim, distance)
	}
	if ax.rollDeg != default3DRollDeg || ax.verticalAxis != default3DVerticalAxis {
		t.Fatalf("roll/vertical axis after SetDefaults = (%v, %v), want (%v, %v)", ax.rollDeg, ax.verticalAxis, default3DRollDeg, default3DVerticalAxis)
	}
	if got := ax.ProjectionType(); got != "persp" {
		t.Fatalf("projection type after SetDefaults = %q, want persp", got)
	}
	if ax.focalLength != default3DFocalLength {
		t.Fatalf("focal length after SetDefaults = %v, want %v", ax.focalLength, default3DFocalLength)
	}
	if got, want := ax.boxAspect, default3DBoxAspect(); !approx(got[0], want[0], 1e-12) ||
		!approx(got[1], want[1], 1e-12) || !approx(got[2], want[2], 1e-12) {
		t.Fatalf("box aspect after SetDefaults = %v, want %v", got, want)
	}
}

func TestAxes3DProjectionLimitsUseMatplotlibXYAutoscaleMargins(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[0], -0.07291666666666667, 1e-12) || !approx(maxs[0], 1.0729166666666667, 1e-12) ||
		!approx(mins[1], -0.07291666666666667, 1e-12) || !approx(maxs[1], 1.0729166666666667, 1e-12) ||
		!approx(mins[2], -0.041666666666666664, 1e-12) || !approx(maxs[2], 2.0416666666666665, 1e-12) {
		t.Fatalf("projection limits = %v..%v, want Matplotlib line autoscale margins plus 3D view margin", mins, maxs)
	}
}
