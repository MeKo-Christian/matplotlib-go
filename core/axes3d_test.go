package core

import (
	"sort"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestAddAxes3DConfiguresProjection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{-1, 1}, []float64{-2, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[0], -0.05, 1e-12) || !approx(maxs[0], 1.05, 1e-12) {
		t.Fatalf("x projection limits = (%v, %v), want 5%% Matplotlib x margin", mins[0], maxs[0])
	}
	if !approx(mins[1], -1.1, 1e-12) || !approx(maxs[1], 1.1, 1e-12) {
		t.Fatalf("y projection limits = (%v, %v), want 5%% Matplotlib y margin", mins[1], maxs[1])
	}
	if !approx(mins[2], -2, 1e-12) || !approx(maxs[2], 2, 1e-12) {
		t.Fatalf("line z projection limits = (%v, %v), want Matplotlib line z margin 0", mins[2], maxs[2])
	}
}

func TestAxes3DScatterProjectionLimitsUseMatplotlibZMargin(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Scatter3D([]float64{0, 1}, []float64{-1, 1}, []float64{-2, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[2], -2.2, 1e-12) || !approx(maxs[2], 2.2, 1e-12) {
		t.Fatalf("scatter z projection limits = (%v, %v), want Matplotlib scatter z margin 5%%", mins[2], maxs[2])
	}
}

func TestAxes3DProjectPointMatchesMatplotlibScatterFixtureLimits(t *testing.T) {
	fig := NewFigure(720, 560)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(22.5767530642105, 32.43415741224721)
	ax.SetYLim(0.09104692778428092, 104.58776816008752)
	ax.SetZLim(-51.007126259038195, -24.886181592223462)

	got := ax.ProjectPoint(29.444073116538434, 91.3534531817269, -37.52903158779476)
	if !approx(got.X, 0.039937290348120026, 1e-12) ||
		!approx(got.Y, 0.013747177404714107, 1e-12) {
		t.Fatalf("scatter fixture projection = %+v, want Matplotlib projection {0.039937290348120026 0.013747177404714107}", got)
	}
}

func TestAxes3DProjectedDataMapsToMatplotlibDisplayCoordinates(t *testing.T) {
	fig := NewFigure(720, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
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

func TestAxes3DScatterDefaultColorUsesShapeCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	palette := fig.RC.Palette()

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	scatter := ax.Scatter3D([]float64{0.5}, []float64{0.2}, []float64{0.1})
	nextLine := ax.Plot3D([]float64{0, 1}, []float64{1, 0}, []float64{0, 1})

	if got, want := line.Col, palette[0]; got != want {
		t.Fatalf("first 3D line color = %+v, want %+v", got, want)
	}
	if got, want := scatter.Color, palette[0]; got != want {
		t.Fatalf("3D scatter color = %+v, want independent shape cycle first color %+v", got, want)
	}
	if got, want := nextLine.Col, palette[1]; got != want {
		t.Fatalf("second 3D line color = %+v, want line cycle second color %+v", got, want)
	}
}

func TestAxes3DScatterUsesMatplotlibDefaultSize(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	scatter := ax.Scatter3D([]float64{0.5}, []float64{0.2}, []float64{0.1})
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := scatter.Size, 20.0; got != want {
		t.Fatalf("3D scatter default size = %v, want Matplotlib Axes3D.scatter s=%v", got, want)
	}

	explicit := 42.0
	scatter = ax.Scatter3D(
		[]float64{0.5},
		[]float64{0.2},
		[]float64{0.1},
		ScatterOptions{Size: &explicit},
	)
	if scatter == nil {
		t.Fatal("Scatter3D with explicit size returned nil")
	}
	if got := scatter.Size; got != explicit {
		t.Fatalf("3D scatter explicit size = %v, want %v", got, explicit)
	}
}

func TestAxes3DScatterDepthShadesAndSortsMarkersLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	scatter := ax.Scatter3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
	)
	if scatter == nil {
		t.Fatal("Scatter3D returned nil")
	}
	if got, want := len(scatter.Colors), 2; got != want {
		t.Fatalf("3D scatter per-marker colors = %d, want %d depth-shaded colors", got, want)
	}
	if !approx(scatter.Colors[0].A, 0.3, 1e-12) || !approx(scatter.Colors[1].A, 1.0, 1e-12) {
		t.Fatalf("3D scatter depth-shaded alphas = %.12g, %.12g; want Matplotlib z-sorted alpha range 0.3..1.0", scatter.Colors[0].A, scatter.Colors[1].A)
	}
	if got, want := len(scatter.EdgeColors), 2; got != want {
		t.Fatalf("3D scatter per-marker edge colors = %d, want %d depth-shaded edge colors", got, want)
	}
	if !approx(scatter.EdgeColors[0].A, 0.3, 1e-12) || !approx(scatter.EdgeColors[1].A, 1.0, 1e-12) {
		t.Fatalf("3D scatter depth-shaded edge alphas = %.12g, %.12g; want Matplotlib z-sorted alpha range 0.3..1.0", scatter.EdgeColors[0].A, scatter.EdgeColors[1].A)
	}
}

func TestAxes3DPlot3DUsesProjectedCoordinates(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
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

func TestAxes3DReprojectsExistingArtistsWhenDataLimitsExpand(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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
	ax, err := fig.AddAxes3D(unitRect())
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

func TestAxes3DSetZLabelRenders3DLabel(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetXLabel("X")
	ax.SetYLabel("Y")
	ax.SetZLabel("Z")
	if got := ax.GetZLabel(); got != "Z" {
		t.Fatalf("GetZLabel = %q, want %q", got, "Z")
	}

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}
	ax.draw3DAxisLabels(r, r, ctx, frameMins, frameMaxs)

	if !containsString(r.texts, "Z") {
		t.Fatalf("expected z-axis label text in draw calls, got %v", r.texts)
	}
}

func dataArtistsForOrderCheck(artists []Artist) []Artist {
	filtered := make([]Artist, 0, len(artists))
	for _, art := range artists {
		if art == nil || art.Z() <= -1000 {
			continue
		}
		filtered = append(filtered, art)
	}
	return filtered
}

func axes3DArtistsEqual(a, b []Artist) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func artistOrderLabel(artists []Artist) string {
	out := ""
	for i, art := range artists {
		if i > 0 {
			out += ">"
		}
		out += artName(art)
	}
	return out
}

func artName(art Artist) string {
	switch art.(type) {
	case *Line2D:
		return "line"
	case *LineCollection:
		return "linecol"
	case *Scatter2D:
		return "scatter"
	case *PolyCollection:
		return "poly"
	default:
		return "artist"
	}
}

func TestAxes3DProjectionLimitsUseMatplotlibXYAutoscaleMargins(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 2})
	mins, maxs := ax.projectionLimits()
	if !approx(mins[0], -0.05, 1e-12) || !approx(maxs[0], 1.05, 1e-12) ||
		!approx(mins[1], -0.05, 1e-12) || !approx(maxs[1], 1.05, 1e-12) ||
		!approx(mins[2], 0, 1e-12) || !approx(maxs[2], 2, 1e-12) {
		t.Fatalf("projection limits = %v..%v, want Matplotlib line autoscale margins [-0.05 -0.05 0]..[1.05 1.05 2]", mins, maxs)
	}
}

func TestAxes3DFrameLimitsAddMatplotlibViewMargin(t *testing.T) {
	mins, maxs := axes3DFrameLimits(vec3{-0.05, -0.05, -0.1}, vec3{1.05, 1.05, 2.1})
	if !approx(mins[0], -0.07291666666666667, 1e-12) ||
		!approx(maxs[0], 1.0729166666666667, 1e-12) ||
		!approx(mins[2], -0.14583333333333334, 1e-12) ||
		!approx(maxs[2], 2.1458333333333335, 1e-12) {
		t.Fatalf("frame limits = %v..%v, want Matplotlib axis3d _get_coord_info view margin", mins, maxs)
	}
}

func TestAxes3DCollectionsUseComputedDepthZOrder(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	line := ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	low := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 0}, {0, 0}},
	)
	high := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{1, 1}, {1, 1}},
	)
	if line == nil || low == nil || high == nil {
		t.Fatalf("expected line and surface artists, got line=%v low=%v high=%v", line, low, high)
	}
	if !(low.Z() > line.Z() && high.Z() > line.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want both above 3D line %.6g like Matplotlib computed_zorder", low.Z(), high.Z(), line.Z())
	}
	if !(high.Z() > low.Z()) {
		t.Fatalf("3D surface zorders = low %.6g high %.6g, want higher projected plane drawn after lower plane", low.Z(), high.Z())
	}
}

func TestAxes3DWireframeGeneratesLineCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 4; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}

func TestAxes3DWireframeTreatsZRowsAsYAndColumnsAsX(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	x := []float64{10, 20, 30}
	y := []float64{1, 2}
	z := [][]float64{
		{0, 0, 0},
		{0, 0, 0},
	}
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := collection.Segments[0][0], (Pt{X: 10, Y: 1}); got != want {
		t.Fatalf("first wireframe point = %+v, want %+v", got, want)
	}
	if got, want := collection.Segments[0][1], (Pt{X: 20, Y: 1}); got != want {
		t.Fatalf("first wireframe row segment end = %+v, want %+v", got, want)
	}
}

func TestAxes3DWireframeSupportsRowColumnStridesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	x, y, z := testGrid3D(5, 5)
	rstride := 2
	cstride := 0
	collection := ax.Wireframe(x, y, z, PlotOptions{RStride: &rstride, CStride: &cstride})
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 3; got != want {
		t.Fatalf("wireframe stride line count = %d, want sampled rows 0,2,4 only (%d)", got, want)
	}
	if got, want := len(collection.Segments[0]), 5; got != want {
		t.Fatalf("wireframe row polyline length = %d, want full row length %d", got, want)
	}
	if got, want := collection.Segments[1][0], (Pt{X: 0, Y: 2}); got != want {
		t.Fatalf("second sampled row starts at %+v, want row 2 start %+v", got, want)
	}
}

func TestAxes3DWireframeSupportsRowColumnCountsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(9, 10)
	rcount := 3
	ccount := 4
	collection := ax.Wireframe(x, y, z, PlotOptions{RCount: &rcount, CCount: &ccount})
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := len(collection.Segments), 8; got != want {
		t.Fatalf("wireframe count line count = %d, want 4 sampled rows + 4 sampled columns", got)
	}
}

func TestAxes3DWireframeDefaultLineWidthMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(2, 2)
	collection := ax.Wireframe(x, y, z)
	if collection == nil {
		t.Fatal("Wireframe returned nil")
	}
	if got, want := collection.LineWidth, 2.0; got != want {
		t.Fatalf("wireframe default line width = %v, want Matplotlib default converted to Go pixels %v", got, want)
	}
}

func TestAxes3DFrameSegmentsUseMatplotlibActiveGridPlanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	segments := ax.frameSegments(vec3{0, 0, 0}, vec3{1, 1, 1})
	want := []Pt{
		ax.ProjectPoint(0.2, 0, 0),
		ax.ProjectPoint(0.2, 1, 0),
		ax.ProjectPoint(0.2, 1, 1),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing Matplotlib-style x gridline through active panes; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameSegmentsDoNotAddSeparateCubeBox(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	segments := ax.frameSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	gridSegments := ax.frameGridSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	if got, want := len(segments), len(gridSegments); got != want {
		t.Fatalf("frame segment count = %d, want grid-only Matplotlib frame count %d", got, want)
	}
}

func TestAxes3DFrameGridSegmentsUseAxisTickLocations(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-3.45, -3.45, -0.88}
	maxs := vec3{3.45, 3.45, 0.78}
	segments := ax.frameSegments(mins, maxs)
	highs := ax.activePaneHighs(mins, maxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = maxs[i]
			maxmin[i] = mins[i]
		} else {
			minmax[i] = mins[i]
			maxmin[i] = maxs[i]
		}
	}

	p0 := minmax
	p1 := minmax
	p2 := minmax
	p0[0], p1[0], p2[0] = -3, -3, -3
	p0[1] = maxmin[1]
	p2[2] = maxmin[2]
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p2[0], p2[1], p2[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !contains3DSegment(segments, want, 1e-12) {
		t.Fatalf("missing gridline at Matplotlib AutoLocator tick -3; want %+v in %+v", want, segments)
	}
}

func TestAxes3DFrameAxisTicksMatchMatplotlibDensity(t *testing.T) {
	ticks := frameAxisTicks(-0.1, 2.1)
	if !containsFloat64Approx(ticks, 0.25, 1e-12) {
		t.Fatalf("3D frame ticks = %v, want Matplotlib-like 0.25 z tick", ticks)
	}
}

func TestAxes3DAxisLineSegmentsUseMatplotlibCameraFacingEdges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	segments := ax.axisLineSegmentsProjected(frameMins, frameMaxs, mins, maxs)
	if got, want := len(segments), 3; got != want {
		t.Fatalf("axis line count = %d, want %d", got, want)
	}

	highs := ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)
	minmax := vec3{}
	maxmin := vec3{}
	for i := range 3 {
		if highs[i] {
			minmax[i] = frameMaxs[i]
			maxmin[i] = frameMins[i]
		} else {
			minmax[i] = frameMins[i]
			maxmin[i] = frameMaxs[i]
		}
	}
	x0 := minmax
	x0[1] = maxmin[1]
	x1 := x0
	x1[0] = maxmin[0]
	wantX := []Pt{
		project3DPointWithLimits(x0[0], x0[1], x0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(x1[0], x1[1], x1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], wantX, 1e-12) {
		t.Fatalf("x axis line = %+v, want Matplotlib camera-facing edge %+v", segments[0], wantX)
	}
}

func TestAxes3DTickSegmentsUseMatplotlibInwardOutwardFactors(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{-0.05, -0.05, -0.1}
	maxs := vec3{1.05, 1.05, 2.1}
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	segments := ax.axisTickSegmentsProjected(frameMins, frameMaxs, mins, maxs, mins, maxs)
	if len(segments) == 0 {
		t.Fatal("axis tick segments are empty")
	}

	pair := ax.axisLineEdgePointPairs(frameMins, frameMaxs, mins, maxs)[0]
	tick := frameAxisTicks(mins[0], maxs[0])[0]
	tickDir := 1
	tickDelta := (maxs[tickDir] - mins[tickDir]) / 12
	if !ax.activePaneHighsProjected(frameMins, frameMaxs, mins, maxs)[tickDir] {
		tickDelta = -tickDelta
	}
	p0 := pair[0]
	p1 := pair[0]
	p0[0] = tick
	p1[0] = tick
	p0[tickDir] = pair[0][tickDir] + 0.1*tickDelta
	p1[tickDir] = pair[0][tickDir] - 0.2*tickDelta
	want := []Pt{
		project3DPointWithLimits(p0[0], p0[1], p0[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
		project3DPointWithLimits(p1[0], p1[1], p1[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, mins, maxs),
	}
	if !pointsEqual(segments[0], want, 1e-12) {
		t.Fatalf("first x tick segment = %+v, want Matplotlib inward/outward segment %+v", segments[0], want)
	}
}

func TestAxes3DPanePolygonsUseMatplotlibActivePanes(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	mins := vec3{0, 0, 0}
	maxs := vec3{1, 1, 1}
	panes := ax.activePanePolygons(mins, maxs)
	if got, want := len(panes), 3; got != want {
		t.Fatalf("pane count = %d, want %d", got, want)
	}
	highs := ax.activePaneHighs(mins, maxs)
	expectedPlanes := [6][4][3]int{
		{{0, 0, 0}, {0, 1, 0}, {0, 1, 1}, {0, 0, 1}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {1, 1, 0}, {1, 1, 1}, {0, 1, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	for axis := range 3 {
		planeIndex := 2 * axis
		if highs[axis] {
			planeIndex++
		}
		want := projectPlaneCorners(ax, expectedPlanes[planeIndex], mins, maxs)
		if !pointsEqual(panes[axis], want, 1e-12) {
			t.Fatalf("pane %d = %+v, want active Matplotlib pane %+v", axis, panes[axis], want)
		}
	}
}

func TestAxes3DPaneFaceColorsMatchMatplotlibDefaults(t *testing.T) {
	got := axes3DPaneFaceColors()
	want := []render.Color{
		{R: 0.95, G: 0.95, B: 0.95, A: 0.5},
		{R: 0.90, G: 0.90, B: 0.90, A: 0.5},
		{R: 0.925, G: 0.925, B: 0.925, A: 0.5},
	}
	if len(got) != len(want) {
		t.Fatalf("pane face colors = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pane face color %d = %+v, want Matplotlib default %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DSurfaceCreatesProjectedPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 1; got != want {
		t.Fatalf("surface polygon count = %d, want %d", got, want)
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("surface face color count = %d, want %d", got, want)
	}
}

func TestAxes3DSurfaceUsesMatplotlibDefaultSampleCounts(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := make([]float64, 90)
	for i := range x {
		x[i] = float64(i)
	}
	y := make([]float64, 70)
	for i := range y {
		y[i] = float64(i)
	}
	z := make([][]float64, len(y))
	for row := range z {
		z[row] = make([]float64, len(x))
		for col := range z[row] {
			z[row][col] = float64(row + col)
		}
	}

	collection := ax.Surface(x, y, z)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 35*45; got != want {
		t.Fatalf("surface polygon count = %d, want Matplotlib default rcount/ccount sampled count %d", got, want)
	}
}

func TestAxes3DSurfaceSupportsRowColumnStridesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(5, 5)
	rstride := 2
	cstride := 2
	collection := ax.Surface(x, y, z, PlotOptions{RStride: &rstride, CStride: &cstride})
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 4; got != want {
		t.Fatalf("surface stride polygon count = %d, want 2x2 sampled patches", got)
	}
}

func TestAxes3DSurfaceSupportsRowColumnCountsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x, y, z := testGrid3D(9, 10)
	rcount := 3
	ccount := 4
	collection := ax.Surface(x, y, z, PlotOptions{RCount: &rcount, CCount: &ccount})
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(collection.Polygons), 9; got != want {
		t.Fatalf("surface count polygon count = %d, want %d sampled patches for 9x10 grid with rcount=3, ccount=4", got, want)
	}
}

func TestAxes3DPlotSurfaceGridHonorsSurfaceOptions(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	collection := ax.PlotSurfaceGrid(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {2, 3}},
		PlotOptions{Antialiased: &antialiased},
	)
	if collection == nil {
		t.Fatal("PlotSurfaceGrid returned nil")
	}
	if got, want := collection.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("plot surface grid antialias = %v, want %v", got, want)
	}
}

func TestAxes3DSurfaceDefaultHasNoEdgeColorsLikeMatplotlibCmapSurface(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)
	if collection == nil {
		t.Fatal("Surface returned nil")
	}
	if got := collection.EdgeColor.A; got != 0 {
		t.Fatalf("surface default edge alpha = %v, want 0 like Matplotlib cmap plot_surface edgecolors", got)
	}
	if got, want := collection.EdgeWidth, 1.0; got != want {
		t.Fatalf("surface default linewidth = %v, want %v like Matplotlib plot_surface", got, want)
	}
}

func TestAxes3DSurfaceExposesScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "plasma"
	vmin := 0.0
	vmax := 10.0
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Colormap: &cmap, VMin: &vmin, VMax: &vmax},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	mapping := surface.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("surface scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
}

func TestAxes3DSurfaceUsesFaceColorsForEdgesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	face := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{FaceColors: []render.Color{face}},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(surface.EdgeColors), len(surface.FaceColors); got != want {
		t.Fatalf("surface edge colors len = %d, want face-colored edges for %d faces", got, want)
	}
	if got := surface.EdgeColors[0]; got != surface.FaceColors[0] {
		t.Fatalf("surface edge color = %+v, want same sampled face color %+v", got, surface.FaceColors[0])
	}
}

func TestAxes3DSurfaceHonorsAntialiasSetting(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	surface := ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Antialiased: &antialiased},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := surface.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("surface antialias = %v, want %v", got, want)
	}
}

func TestAxes3DSurfaceAxLimClipDropsFacesOutsideExplicit3DLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetXLim(0, 1.5)

	surface := ax.Surface(
		[]float64{0, 1, 2},
		[]float64{0, 1},
		[][]float64{{0, 1, 2}, {0, 1, 2}},
		PlotOptions{AxLimClip: true},
	)
	if surface == nil {
		t.Fatal("Surface returned nil")
	}
	if got, want := len(surface.Polygons), 1; got != want {
		t.Fatalf("surface clipped polygon count = %d, want one fully in-limit face", got)
	}
}

func TestAxes3DTrisurfExposesConfiguredNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	cmap := "inferno"
	surface := ax.Trisurf(tri, []float64{1, 10, 100}, PlotOptions{
		Colormap: &cmap,
		Norm:     LogNorm{VMin: 1, VMax: 100},
	})
	if surface == nil {
		t.Fatal("Trisurf returned nil")
	}
	mapping := surface.ScalarMap()
	if mapping.Colormap != cmap || mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("trisurf scalar map = %+v, want inferno/log norm", mapping)
	}
}

func TestAxes3DTrisurfHonorsAntialiasSetting(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	antialiased := false
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	surface := ax.Trisurf(tri, []float64{1, 10, 100}, PlotOptions{Antialiased: &antialiased})
	if surface == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := surface.Antialias, render.AntialiasOff; got != want {
		t.Fatalf("trisurf antialias = %v, want %v", got, want)
	}
}

func TestAxes3DStemProjectsBaselineStemsAndMarkers(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	bottom := -1.0
	container := ax.Stem3D(
		[]float64{0, 1},
		[]float64{2, 3},
		[]float64{4, 5},
		Stem3DOptions{Bottom: &bottom},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	if got, want := len(container.StemLines.Segments), 2; got != want {
		t.Fatalf("stem segment count = %d, want %d", got, want)
	}
	wantStem := []Pt{
		ax.ProjectPoint(0, 2, bottom),
		ax.ProjectPoint(0, 2, 4),
	}
	if !pointsEqual(container.StemLines.Segments[0], wantStem, 1e-12) {
		t.Fatalf("first stem = %+v, want projected z-oriented stem %+v", container.StemLines.Segments[0], wantStem)
	}
	wantBaseline := []Pt{
		ax.ProjectPoint(0, 2, bottom),
		ax.ProjectPoint(1, 3, bottom),
	}
	if !pointsEqual(container.Baseline.XY, wantBaseline, 1e-12) {
		t.Fatalf("stem baseline = %+v, want projected baseline %+v", container.Baseline.XY, wantBaseline)
	}
	if got, want := len(container.MarkerCollection.Offsets), 2; got != want {
		t.Fatalf("stem marker count = %d, want %d", got, want)
	}
	palette := style.Default.Palette()
	if got, want := container.StemLines.Color, palette[0]; got != want {
		t.Fatalf("stem line color = %+v, want Matplotlib linefmt C0 %+v", got, want)
	}
	if got, want := container.StemLines.LineCap, render.CapButt; got != want {
		t.Fatalf("stem line cap = %v, want Matplotlib LineCollection default cap %v", got, want)
	}
	if got, want := container.MarkerCollection.FaceColor, palette[0]; got != want {
		t.Fatalf("stem marker color = %+v, want Matplotlib markerfmt C0 %+v", got, want)
	}
	if got, want := container.MarkerCollection.Size, pointsToPixels(ax.resolvedRC(), 6); !approx(got, want, 1e-12) {
		t.Fatalf("stem marker size = %v, want Matplotlib 6 point Line2D marker diameter %v", got, want)
	}
	if got, want := container.Baseline.Col, palette[3]; got != want {
		t.Fatalf("stem baseline color = %+v, want Matplotlib basefmt C3 %+v", got, want)
	}
}

func TestAxes3DStemSupportsMatplotlibOrientationJuggling(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	bottom := -2.0
	container := ax.Stem3D(
		[]float64{1},
		[]float64{2},
		[]float64{3},
		Stem3DOptions{Bottom: &bottom, Orientation: "x"},
	)
	if container == nil {
		t.Fatal("Stem3D returned nil")
	}
	want := []Pt{
		ax.ProjectPoint(bottom, 2, 3),
		ax.ProjectPoint(1, 2, 3),
	}
	if !pointsEqual(container.StemLines.Segments[0], want, 1e-12) {
		t.Fatalf("x-oriented stem = %+v, want Matplotlib orientation='x' stem %+v", container.StemLines.Segments[0], want)
	}
}

func TestAxes3DFillBetweenCreatesProjectedQuadBands(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	fill := ax.FillBetween3D(
		[]float64{0, 1, 2},
		[]float64{0, 0, 0},
		[]float64{1, 1, 1},
		[]float64{0, 1, 2},
		[]float64{1, 1, 1},
		[]float64{0, 0, 0},
		FillBetween3DOptions{Mode: FillBetween3DModeQuad},
	)
	if fill == nil {
		t.Fatal("FillBetween3D returned nil")
	}
	if got, want := len(fill.Polygons), 2; got != want {
		t.Fatalf("FillBetween3D polygon count = %d, want one quad per adjacent pair (%d)", got, want)
	}
	wantFirst := []Pt{
		ax.ProjectPoint(0, 0, 1),
		ax.ProjectPoint(1, 0, 1),
		ax.ProjectPoint(1, 1, 0),
		ax.ProjectPoint(0, 1, 0),
	}
	if !pointsEqual(fill.Polygons[0], wantFirst, 1e-12) {
		t.Fatalf("first fill polygon = %+v, want projected quad %+v", fill.Polygons[0], wantFirst)
	}
}

func TestAxes3DBarProjects2DBarsIntoSelectedZDirection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	width := 0.4
	zs := []float64{2}
	bars := ax.Bar(
		[]float64{1},
		[]float64{3},
		Bar3DPlaneOptions{Width: &width, Zs: zs, ZDir: "y"},
	)
	if bars == nil {
		t.Fatal("Axes3D.Bar returned nil")
	}
	if got, want := len(bars.Polygons), 1; got != want {
		t.Fatalf("projected bar polygon count = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.8, 2, 0),
		ax.ProjectPoint(0.8, 2, 3),
		ax.ProjectPoint(1.2, 2, 3),
		ax.ProjectPoint(1.2, 2, 0),
	}
	if !pointsEqual(bars.Polygons[0], want, 1e-12) {
		t.Fatalf("projected y-dir bar = %+v, want Matplotlib juggle_axes projection %+v", bars.Polygons[0], want)
	}
}

func TestAxes3DQuiverUsesMatplotlibTailPivotGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	length := 2.0
	q := ax.Quiver(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{1},
		[]float64{0},
		[]float64{0},
		Quiver3DOptions{Length: &length, Pivot: "tail"},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	if got, want := len(q.Segments), 3; got != want {
		t.Fatalf("quiver segment count = %d, want shaft plus two arrowhead segments (%d)", got, want)
	}
	wantShaft := []Pt{
		ax.ProjectPoint(2, 0, 0),
		ax.ProjectPoint(0, 0, 0),
	}
	if !pointsEqual(q.Segments[0], wantShaft, 1e-12) {
		t.Fatalf("quiver shaft = %+v, want Matplotlib tail-pivot shaft %+v", q.Segments[0], wantShaft)
	}
}

func TestAxes3DQuiverNormalizesVectorsAndSupportsMiddlePivot(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	length := 4.0
	q := ax.Quiver(
		[]float64{0},
		[]float64{0},
		[]float64{0},
		[]float64{2},
		[]float64{0},
		[]float64{0},
		Quiver3DOptions{Length: &length, Normalize: true, Pivot: "middle"},
	)
	if q == nil {
		t.Fatal("Quiver returned nil")
	}
	wantShaft := []Pt{
		ax.ProjectPoint(2, 0, 0),
		ax.ProjectPoint(-2, 0, 0),
	}
	if !pointsEqual(q.Segments[0], wantShaft, 1e-12) {
		t.Fatalf("normalized middle-pivot shaft = %+v, want %+v", q.Segments[0], wantShaft)
	}
}

func TestAxes3DErrorBarProjectsXYZRangesAndCaps(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	capSize := 0.2
	errs := ax.ErrorBar3D(
		[]float64{1},
		[]float64{2},
		[]float64{3},
		[]float64{0.5},
		[]float64{0.25},
		[]float64{1},
		ErrorBar3DOptions{CapSize: &capSize},
	)
	if errs == nil {
		t.Fatal("ErrorBar3D returned nil")
	}
	if got, want := len(errs.Segments), 15; got != want {
		t.Fatalf("3D errorbar segment count = %d, want 3 bars plus 12 cap segments", got)
	}
	wantXRange := []Pt{
		ax.ProjectPoint(0.5, 2, 3),
		ax.ProjectPoint(1.5, 2, 3),
	}
	if !pointsEqual(errs.Segments[0], wantXRange, 1e-12) {
		t.Fatalf("x error range = %+v, want projected x range %+v", errs.Segments[0], wantXRange)
	}
	wantZRange := []Pt{
		ax.ProjectPoint(1, 2, 2),
		ax.ProjectPoint(1, 2, 4),
	}
	if !pointsEqual(errs.Segments[2], wantZRange, 1e-12) {
		t.Fatalf("z error range = %+v, want projected z range %+v", errs.Segments[2], wantZRange)
	}
}

func TestAxes3DErrorBarUsesComputedDepthZOrder(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	low := ax.ErrorBar3D([]float64{0}, []float64{0}, []float64{0}, nil, nil, []float64{0.1})
	high := ax.ErrorBar3D([]float64{0}, []float64{0}, []float64{1}, nil, nil, []float64{0.1})
	if low == nil || high == nil {
		t.Fatalf("expected errorbar collections, got low=%v high=%v", low, high)
	}
	if !(high.Z() > low.Z()) {
		t.Fatalf("3D errorbar zorders = low %.6g high %.6g, want projected depth ordering", low.Z(), high.Z())
	}
}

func projectPlaneCorners(ax *Axes3D, plane [4][3]int, mins, maxs vec3) []Pt {
	points := make([]Pt, len(plane))
	for i, corner := range plane {
		x := mins[0]
		if corner[0] == 1 {
			x = maxs[0]
		}
		y := mins[1]
		if corner[1] == 1 {
			y = maxs[1]
		}
		z := mins[2]
		if corner[2] == 1 {
			z = maxs[2]
		}
		points[i] = ax.ProjectPoint(x, y, z)
	}
	return points
}

func pointsEqual(got, want []Pt, tol float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if !approx(got[i].X, want[i].X, tol) || !approx(got[i].Y, want[i].Y, tol) {
			return false
		}
	}
	return true
}

func contains3DSegment(segments [][]Pt, want []Pt, tol float64) bool {
	for _, segment := range segments {
		if len(segment) != len(want) {
			continue
		}
		matches := true
		for i := range want {
			if !approx(segment[i].X, want[i].X, tol) || !approx(segment[i].Y, want[i].Y, tol) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func containsFloat64Approx(values []float64, want, tol float64) bool {
	for _, got := range values {
		if approx(got, want, tol) {
			return true
		}
	}
	return false
}

func testGrid3D(cols, rows int) ([]float64, []float64, [][]float64) {
	x := make([]float64, cols)
	for i := range x {
		x[i] = float64(i)
	}
	y := make([]float64, rows)
	z := make([][]float64, rows)
	for row := range rows {
		y[row] = float64(row)
		z[row] = make([]float64, cols)
		for col := range cols {
			z[row][col] = float64(row + col)
		}
	}
	return x, y, z
}

func TestAxes3DContourAndContourfCreateCollections(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	contour := ax.Contour(x, y, z)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if contourf := ax.Contourf(x, y, z); contourf == nil {
		t.Fatal("Contourf returned nil")
	}
}

func TestAxes3DContourfProjectsFilledContourBands(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	fill := ax.Contourf(x, y, z)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, cellCount := len(fill.Paths), 1; got <= cellCount {
		t.Fatalf("Contourf compound path count = %d, want filled contour band paths rather than %d grid cell", got, cellCount)
	}
}

func TestAxes3DContourfUsesExplicitZOffset(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	offset := -3.0
	fill := ax.Contourf(x, y, z, PlotOptions{LevelCount: 3, Offset: &offset})
	if fill == nil || len(fill.Paths) == 0 || len(fill.Paths[0].V) == 0 {
		t.Fatalf("Contourf returned no polygons: %+v", fill)
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, nil, 3, true)
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	rawPolygons, _ := contourGridBandPolygons(x, y, z, levels, ContourOptions{}, mapping, 0.45)
	if len(rawPolygons) == 0 || len(rawPolygons[0]) == 0 {
		t.Fatal("expected raw contour band polygons")
	}
	want := ax.ProjectPoint(rawPolygons[0][0].X, rawPolygons[0][0].Y, offset)
	if got := fill.Paths[0].V[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("Contourf first point = %+v, want projection at explicit offset %+v", got, want)
	}
}

func TestAxes3DContourfUsesProjectedCollectionZLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	offset := -3.0
	levelCount := 3
	fill := ax.Contourf(x, y, z, PlotOptions{LevelCount: levelCount, Offset: &offset})
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}

	values := flattenGridValues(z)
	levels := contourLevels(values, nil, levelCount, true)
	mapping := resolveScalarMapValues(values, "viridis", nil, nil)
	mapping.VMin = levels[0]
	mapping.VMax = levels[len(levels)-1]
	rawPolygons, _ := contourGridBandPolygons(x, y, z, levels, ContourOptions{}, mapping, 0.45)
	depth := 0.0
	first := true
	for _, polygon := range rawPolygons {
		for _, pt := range polygon {
			_, zDepth := ax.projectPointDepth(pt.X, pt.Y, offset)
			if first || zDepth < depth {
				depth = zDepth
				first = false
			}
		}
	}
	if first {
		t.Fatal("expected raw contour band polygons")
	}
	want := computed3DCollectionZ(depth)
	if got := fill.Z(); !approx(got, want, 1e-12) {
		t.Fatalf("Contourf zorder = %.12g, want computed projected zorder %.12g like Matplotlib Collection3D", got, want)
	}
}

func TestAxes3DContourfAutoscaleUsesFilledLevelMidpointsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0.1, 0.2}, {0.8, 0.9}},
		PlotOptions{Levels: []float64{0, 1, 2}},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	mins, maxs := ax.projectionLimits()
	if !approx(mins[2], 0.5, 1e-12) || !approx(maxs[2], 1.5, 1e-12) {
		t.Fatalf("Contourf projection z limits = %.12g..%.12g, want Matplotlib autoscale from filled midpoints 0.5..1.5", mins[2], maxs[2])
	}
}

func TestAxes3DContourfUsesStructuredGridBandPolygons(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	offset := -1.0
	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0.5, 1.5}, Offset: &offset},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, want := len(fill.Paths), 1; got != want {
		t.Fatalf("Contourf paths = %d, want one structured quad band path", got)
	}
}

func TestAxes3DContourfGroupsBandsIntoCompoundPathsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1, 2},
		[]float64{0, 1, 2},
		[][]float64{
			{0, 1, 0},
			{1, 2, 1},
			{0, 1, 0},
		},
		PlotOptions{Levels: []float64{0.5, 1.5}},
	)
	if fill == nil {
		t.Fatal("Contourf returned nil")
	}
	if got, want := len(fill.Paths), 1; got != want {
		t.Fatalf("Contourf paths = %d, want one compound path per filled contour band like Matplotlib", got)
	}
	if len(fill.Paths[0].C) == 0 || len(fill.Paths[0].V) <= 4 {
		t.Fatalf("Contourf compound path = %+v, want multiple cell polygons grouped into one path", fill.Paths[0])
	}
}

func TestCompoundContourPathsDissolvesSharedBandEdgesLikeMatplotlib(t *testing.T) {
	color := render.Color{R: 0.2, G: 0.4, B: 0.6, A: 1}
	paths, colors := compoundContourPaths(
		[][]geom.Pt{
			{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
				{X: 0, Y: 1},
			},
			{
				{X: 1, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 1},
				{X: 1, Y: 1},
			},
		},
		[]render.Color{color, color},
	)
	if got, want := len(paths), 1; got != want {
		t.Fatalf("compound paths = %d, want one path for the filled band", got)
	}
	if got, want := len(colors), 1; got != want || colors[0] != color {
		t.Fatalf("compound colors = %+v, want [%+v]", colors, color)
	}
	moveCount := 0
	closeCount := 0
	for _, cmd := range paths[0].C {
		switch cmd {
		case geom.MoveTo:
			moveCount++
		case geom.ClosePath:
			closeCount++
		}
	}
	if moveCount != 1 || closeCount != 1 {
		t.Fatalf("compound path commands = %+v, want one closed region without the shared cell edge", paths[0].C)
	}
}

func TestAxes3DContourProjectsLinesAtContourLevels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	ax.observe3DGrid(x, y, z)

	levelCount := 3
	got := ax.projectedContourSegments(x, y, z, levelCount)
	values := flattenGridValues(z)
	levels := contourLevels(values, nil, levelCount, false)
	rawLines, rawLevels := contourGridPolylines(x, y, z, levels)
	want := make([][]Pt, len(rawLines))
	for i, line := range rawLines {
		want[i] = make([]Pt, len(line))
		for j, pt := range line {
			want[i][j] = ax.ProjectPoint(pt.X, pt.Y, rawLevels[i])
		}
	}
	if len(got) != len(want) {
		t.Fatalf("contour segment count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if !pointsEqual(got[i], want[i], 1e-12) {
			t.Fatalf("contour segment %d = %+v, want x/y contour projected at level z %+v", i, got[i], want[i])
		}
	}
}

func TestAxes3DContourUsesLevelColorsByDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1}
	y := []float64{0, 1}
	z := [][]float64{
		{0, 1},
		{1, 2},
	}
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if len(contour.Colors) != len(contour.Segments) {
		t.Fatalf("contour colors = %d, segments = %d; want per-level colormapped colors by default", len(contour.Colors), len(contour.Segments))
	}
	if len(contour.Colors) > 1 && contour.Colors[0] == contour.Colors[len(contour.Colors)-1] {
		t.Fatalf("first and last contour colors are both %+v, want level-dependent colors", contour.Colors[0])
	}
}

func TestAxes3DContourExposesConfiguredScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "magma"
	vmin := 0.0
	vmax := 10.0
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{
			Colormap:   &cmap,
			VMin:       &vmin,
			VMax:       &vmax,
			LevelCount: 4,
		},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("contour scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
}

func TestAxes3DContourExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Color: &override, LevelCount: 3},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("contour scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
}

func TestAxes3DContourfExposesConfiguredScalarMapForColorbars(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	cmap := "plasma"
	vmin := 0.0
	vmax := 12.0
	contour := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{
			Colormap:   &cmap,
			VMin:       &vmin,
			VMax:       &vmax,
			LevelCount: 3,
		},
	)
	if contour == nil {
		t.Fatal("Contourf returned nil")
	}
	mapping := contour.ScalarMap()
	if mapping.Colormap != cmap || mapping.VMin != vmin || mapping.VMax != vmax {
		t.Fatalf("contourf scalar map = %+v, want cmap=%q range %.1f..%.1f", mapping, cmap, vmin, vmax)
	}
}

func TestAxes3DContourfExplicitColorDisablesScalarMapStateLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	override := render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	contour := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 2}, {4, 6}},
		PlotOptions{Color: &override, LevelCount: 3},
	)
	if contour == nil {
		t.Fatal("Contourf returned nil")
	}
	if contour.Colormap != "" || contour.Norm != nil || contour.VMin != 0 || contour.VMax != 0 {
		t.Fatalf("contourf scalar map state = cmap=%q norm=%T vmin=%g vmax=%g, want no scalar-map metadata", contour.Colormap, contour.Norm, contour.VMin, contour.VMax)
	}
}

func TestAxes3DContourSupportsMatplotlibZDirJuggling(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)

	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0.5}, ZDir: "x"},
	)
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if got, want := len(contour.Segments), 1; got != want {
		t.Fatalf("contour segment count = %d, want %d", got, want)
	}
	want := []Pt{
		ax.ProjectPoint(0.5, 0, 0.5),
		ax.ProjectPoint(0.5, 0.5, 1),
		ax.ProjectPoint(0.5, 1, 1.5),
	}
	if !pointsEqual(contour.Segments[0], want, 1e-12) {
		t.Fatalf("x-directed contour = %+v, want Matplotlib rotate_axes/juggle_axes contour %+v", contour.Segments[0], want)
	}
}

func TestAxes3DContourUsesExplicitOffsetPlane(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	offset := -2.0
	contour := ax.Contour(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{1}, Offset: &offset},
	)
	if contour == nil || len(contour.Segments) == 0 || len(contour.Segments[0]) == 0 {
		t.Fatalf("Contour returned no segments: %+v", contour)
	}

	rawLines, _ := contourGridPolylines([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}, []float64{1})
	if len(rawLines) == 0 || len(rawLines[0]) == 0 {
		t.Fatal("expected raw contour line")
	}
	want := ax.ProjectPoint(rawLines[0][0].X, rawLines[0][0].Y, offset)
	if got := contour.Segments[0][0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("contour offset point = %+v, want explicit offset projection %+v", got, want)
	}
}

func TestAxes3DContourAxLimClipUsesExplicit3DLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetDistance(0)
	ax.SetView(0, 0)
	ax.SetXLim(0, 0.75)
	ax.SetYLim(0, 1)

	x := []float64{0, 0.5, 1}
	y := []float64{0, 0.5, 1}
	z := [][]float64{
		{0, 0.5, 1},
		{0.5, 1, 1.5},
		{1, 1.5, 2},
	}
	contour := ax.Contour(x, y, z, PlotOptions{Levels: []float64{1}, AxLimClip: true})
	if contour == nil {
		t.Fatal("Contour returned nil")
	}
	if got, want := len(contour.Segments), 1; got != want {
		t.Fatalf("contour clipped segment count = %d, want %d", got, want)
	}

	rawLines, _ := contourGridPolylines(x, y, z, []float64{1})
	if len(rawLines) != 1 {
		t.Fatalf("raw contour lines = %d, want 1", len(rawLines))
	}
	wantRuns := ax.clip3DPolylineRuns(contourPolyline3D(rawLines[0], 1, "z"))
	if len(wantRuns) != 1 {
		t.Fatalf("clipped contour runs = %d, want 1", len(wantRuns))
	}
	want := make([]Pt, len(wantRuns[0]))
	for i, point := range wantRuns[0] {
		want[i] = ax.ProjectPoint(point[0], point[1], point[2])
	}
	if !pointsEqual(contour.Segments[0], want, 1e-12) {
		t.Fatalf("contour clipped segment = %+v, want explicit-view-limit run %+v", contour.Segments[0], want)
	}
}

func TestAxes3DContourZOrderUsesContourGeometry(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{-1, 0, 1}
	y := []float64{-1, 0, 1}
	z := [][]float64{
		{0.2, 0.6, 0.2},
		{0.6, 1.0, 0.6},
		{0.2, 0.6, 0.2},
	}
	surface := ax.Surface(x, y, z)
	contour := ax.Contour(x, y, z, PlotOptions{LevelCount: 4})
	if surface == nil || contour == nil {
		t.Fatalf("expected surface and contour collections, got surface=%v contour=%v", surface, contour)
	}
	if !(surface.Z() > contour.Z()) {
		t.Fatalf("surface zorder %.6g, contour zorder %.6g; want surface drawn over 3D contour lines like Matplotlib computed_zorder", surface.Z(), contour.Z())
	}
}

func TestAxes3DContourfUsesFilledLevelMidpointsByDefault(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0, 1, 2}},
	)
	if fill == nil || len(fill.Paths) == 0 || len(fill.Paths[0].V) == 0 {
		t.Fatalf("Contourf returned no paths: %+v", fill)
	}

	rawPolygons := contourGridBandPolygonsForLevel([]float64{0, 1}, []float64{0, 1}, [][]float64{{0, 1}, {1, 2}}, 0, 1)
	if len(rawPolygons) == 0 || len(rawPolygons[0]) == 0 {
		t.Fatal("expected raw first-band contour polygon")
	}
	want := ax.ProjectPoint(rawPolygons[0][0].X, rawPolygons[0][0].Y, 0.5)
	if got := fill.Paths[0].V[0]; !approx(got.X, want.X, 1e-12) || !approx(got.Y, want.Y, 1e-12) {
		t.Fatalf("contourf midpoint point = %+v, want projection at first-band midpoint %+v", got, want)
	}
}

func TestAxes3DContourfAxLimClipDropsOffsetBandsOutsideExplicitZLimits(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetZLim(0, 0.75)

	offset := 1.0
	fill := ax.Contourf(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
		PlotOptions{Levels: []float64{0, 1, 2}, Offset: &offset, AxLimClip: true},
	)
	if fill != nil {
		t.Fatalf("Contourf with offset plane outside z limits returned %+v, want nil", fill)
	}
}

func TestFormat3DTickUsesUnicodeMinusLikeMatplotlib(t *testing.T) {
	ticks := []float64{-0.5, -0.25, 0}
	if got := format3DTick(-0.25, 1, ticks); got != "\u22120.25" {
		t.Fatalf("3D negative tick label = %q, want Unicode minus like Matplotlib", got)
	}
}

func TestAxes3DDrawsYAxisEndpointTickLabels(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	mins, maxs := ax.projectionLimits()
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, mins, maxs, mins, maxs)

	xTicks := frameAxisTicks(mins[0], maxs[0])
	yTicks := frameAxisTicks(mins[1], maxs[1])
	zTicks := frameAxisTicks(mins[2], maxs[2])
	if got, want := len(r.texts), len(xTicks)+len(yTicks)+len(zTicks); got != want {
		t.Fatalf("3D tick label count = %d, want x+y+z endpoint labels included (%d)", got, want)
	}
}

func TestAxes3DFrameTextDrawsBeforeDataCollectionsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(420, 320)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DDrawOrderRecorder{}
	DrawFigure(fig, r)

	textAt, dataAt := -1, -1
	for i, event := range r.events {
		if event == "text" && textAt < 0 {
			textAt = i
		}
		if event == "data" && dataAt < 0 {
			dataAt = i
		}
	}
	if textAt < 0 || dataAt < 0 {
		t.Fatalf("draw events = %v, want both 3D frame text and data collection draw events", r.events)
	}
	if textAt >= dataAt {
		t.Fatalf("draw events = %v, want 3D axis/tick text before data collections like Matplotlib Axes3D.draw", r.events)
	}
}

func TestAxes3DFrameUsesRCLineWidthsLikeMatplotlib(t *testing.T) {
	gridWidth := 2.2
	axisWidth := 1.7
	fig := NewFigure(
		420, 320,
		style.WithGridLineWidths(gridWidth, gridWidth),
		style.WithAxisLineWidth(axisWidth),
	)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DLineWidthRecorder{}
	DrawFigure(fig, r)

	if !containsFloat64(r.widths, gridWidth) {
		t.Fatalf("3D frame stroke widths = %v, want grid linewidth from RC %.3g", r.widths, gridWidth)
	}
	if !containsFloat64(r.widths, axisWidth) {
		t.Fatalf("3D frame stroke widths = %v, want axis linewidth from RC %.3g", r.widths, axisWidth)
	}
}

func TestAxes3DFrameUsesRCColorsLikeMatplotlib(t *testing.T) {
	gridColor := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	axisColor := render.Color{R: 0.6, G: 0.1, B: 0.2, A: 1}
	fig := NewFigure(
		420, 320,
		style.WithGridColors(gridColor, gridColor),
		style.WithAxesEdgeColor(axisColor),
	)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.Surface(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{0, 1}, {1, 2}},
	)

	r := &axes3DLineWidthRecorder{}
	DrawFigure(fig, r)

	if !containsColor(r.colors, gridColor) {
		t.Fatalf("3D frame stroke colors = %+v, want grid color from RC %+v", r.colors, gridColor)
	}
	if !containsColor(r.colors, axisColor) {
		t.Fatalf("3D frame stroke colors = %+v, want axes edge color from RC %+v", r.colors, axisColor)
	}
}

func TestAxes3DTickLabelsUseMatplotlibDataSpaceOffset(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.10},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetTitle("3D Toolkit Scaffold")
	ax.SetXLabel("x")
	ax.SetYLabel("y")
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe([]float64{0, 1}, []float64{0, 1}, z)
	ax.Surface([]float64{0, 1}, []float64{0, 1}, z)

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, frameMins, frameMaxs, mins, maxs)

	if len(r.texts) == 0 {
		t.Fatal("expected 3D tick labels to be drawn")
	}
	xTicks := frameAxisTicks(mins[0], maxs[0])
	label := format3DTick(xTicks[0], 0, xTicks)
	if got := r.texts[0]; got != label {
		t.Fatalf("first tick label = %q, want first x tick %q", got, label)
	}
	fontSize := ctx.RC.TickLabelSize("x")
	expectedAnchor := expectedMatplotlib3DTickLabelAnchor(ax, ctx, 0, xTicks[0], frameMins, frameMaxs, mins, maxs)
	layout := measureSingleLineTextLayout(r, label, fontSize, ctx.RC.FontKey)
	want := alignedSingleLineOrigin(expectedAnchor, layout, TextAlignCenter, textLayoutVAlignTop)
	if !approx(r.positions[0].X, want.X, 1e-9) || !approx(r.positions[0].Y, want.Y, 1e-9) {
		t.Fatalf("first x tick label origin = %+v, want Matplotlib top-aligned data-space offset origin %+v", r.positions[0], want)
	}
}

func TestAxes3DTickLabelAnchorsMatchMatplotlibBasicFixture(t *testing.T) {
	fig := NewFigure(760, 560)
	ax, err := fig.AddAxes3D(geom.Rect{
		Min: geom.Pt{X: 0.12, Y: 0.14},
		Max: geom.Pt{X: 0.88, Y: 0.88},
	})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax.SetView(30, -60)
	ax.Plot3D([]float64{0, 1}, []float64{0, 1}, []float64{0, 1})
	ax.Scatter3D([]float64{0.5, 0.7}, []float64{0.2, 0.9}, []float64{0.1, 0.3})
	z := [][]float64{{0, 1}, {1, 2}}
	ax.Wireframe([]float64{0, 1}, []float64{0, 1}, z)
	ax.Surface([]float64{0, 1}, []float64{0, 1}, z)
	ax.Contour([]float64{0, 1}, []float64{0, 1}, z)
	ax.Bar3D([]float64{0.2}, []float64{0.3}, []float64{0.4}, []float64{0.2}, []float64{0.2}, []float64{0.3})
	ax.Text3D(0.2, 0.8, 0.6, "demo point")

	mins, maxs := ax.projectionLimits()
	frameMins, frameMaxs := axes3DFrameLimits(mins, maxs)
	ctx := newAxesDrawContext(ax.Axes, fig, fig.DisplayRect(), ax.adjustedLayout(fig))
	r := &axes3DTextRecorder{}

	ax.draw3DTickLabels(r, r, ctx, frameMins, frameMaxs, mins, maxs)

	fontSize := ctx.RC.TickLabelSize("x")
	for _, tc := range []struct {
		label      string
		occurrence int
		want       geom.Pt
	}{
		// Display space is y-up: Y is the flip of Matplotlib's top-down pixel (H=560).
		{label: "0.0", occurrence: 1, want: geom.Pt{X: 207.736, Y: 154.162}},
		{label: "0.0", occurrence: 2, want: geom.Pt{X: 463.314, Y: 94.691}},
		{label: "0.00", occurrence: 1, want: geom.Pt{X: 586.431, Y: 245.513}},
	} {
		idx := -1
		count := 0
		for i, label := range r.texts {
			if label == tc.label {
				count++
				if count == tc.occurrence {
					idx = i
					break
				}
			}
		}
		if idx == -1 {
			t.Fatalf("tick label %q was not drawn; got labels %v", tc.label, r.texts)
		}
		layout := measureSingleLineTextLayout(r, tc.label, fontSize, ctx.RC.FontKey)
		got := geom.Pt{
			X: r.positions[idx].X + textHorizontalOriginOffset(layout, TextAlignCenter),
			Y: r.positions[idx].Y - textBaselineOffset(layout, textLayoutVAlignTop),
		}
		if !approx(got.X, tc.want.X, 0.01) || !approx(got.Y, tc.want.Y, 0.01) {
			t.Errorf("%q occurrence %d tick label anchor = %+v, want Matplotlib %+v", tc.label, tc.occurrence, got, tc.want)
		}
	}
}

type axes3DTextRecorder struct {
	render.NullRenderer
	texts     []string
	positions []geom.Pt
}

func (r *axes3DTextRecorder) DrawText(text string, pos geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
	r.positions = append(r.positions, pos)
}

func (r *axes3DTextRecorder) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

type axes3DDrawOrderRecorder struct {
	render.NullRenderer
	events []string
}

func (r *axes3DDrawOrderRecorder) Path(_ geom.Path, paint *render.Paint) {
	if paint == nil || paint.Fill.A <= 0.8 {
		return
	}
	if paint.Fill.R > 0.98 && paint.Fill.G > 0.98 && paint.Fill.B > 0.98 {
		return
	}
	r.events = append(r.events, "data")
}

func (r *axes3DDrawOrderRecorder) DrawText(string, geom.Pt, float64, render.Color) {
	r.events = append(r.events, "text")
}

type axes3DLineWidthRecorder struct {
	render.NullRenderer
	widths []float64
	colors []render.Color
}

func (r *axes3DLineWidthRecorder) Path(_ geom.Path, paint *render.Paint) {
	if paint == nil || paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		return
	}
	r.widths = append(r.widths, paint.LineWidth)
	r.colors = append(r.colors, paint.Stroke)
}

func containsFloat64(values []float64, want float64) bool {
	for _, got := range values {
		if approx(got, want, 1e-12) {
			return true
		}
	}
	return false
}

func containsColor(values []render.Color, want render.Color) bool {
	for _, got := range values {
		if approx(got.R, want.R, 1e-12) &&
			approx(got.G, want.G, 1e-12) &&
			approx(got.B, want.B, 1e-12) &&
			approx(got.A, want.A, 1e-12) {
			return true
		}
	}
	return false
}

func expectedMatplotlib3DTickLabelAnchor(ax *Axes3D, ctx *DrawContext, axis int, tick float64, mins, maxs, projMins, projMaxs vec3) geom.Pt {
	pair := ax.axisLineEdgePointPairs(mins, maxs, projMins, projMaxs)[axis]
	pos := pair[0]
	pos[axis] = tick
	tickDirs := [3]int{1, 0, 0}
	pos[tickDirs[axis]] = pair[0][tickDirs[axis]]

	centers, deltas := testAxes3DLabelCentersDeltas(ctx, mins, maxs)
	labelDeltas := vec3{}
	for i := range 3 {
		labelDeltas[i] = (defaultTickPadPt + 8) * deltas[i]
	}
	pos = testMove3DLabelFromCenter(pos, centers, labelDeltas, axis)
	projected := project3DPointWithLimits(pos[0], pos[1], pos[2], ax.elevationDeg, ax.azimuthDeg, ax.distance, projMins, projMaxs)
	return ctx.TransformFor(Coords(CoordData)).Apply(projected)
}

func testAxes3DLabelCentersDeltas(ctx *DrawContext, mins, maxs vec3) (vec3, vec3) {
	centers := vec3{}
	deltas := vec3{}
	dpi := 100.0
	if ctx != nil && ctx.RC.DPI > 0 {
		dpi = ctx.RC.DPI
	}
	deltasPerPoint := 48 / (72 * (ctx.Clip.W() + ctx.Clip.H()) / dpi)
	const scale = 0.08 // matches matplotlib's 1/12 * 24/25 with automargin compensation
	for i := range 3 {
		centers[i] = (mins[i] + maxs[i]) / 2
		deltas[i] = (maxs[i] - mins[i]) * scale * deltasPerPoint
	}
	return centers, deltas
}

func testMove3DLabelFromCenter(pos, centers, deltas vec3, axis int) vec3 {
	for i := range 3 {
		if i == axis {
			continue
		}
		if pos[i] < centers[i] {
			pos[i] -= deltas[i]
		} else {
			pos[i] += deltas[i]
		}
	}
	return pos
}

func TestAxes3DBar3DCreatesSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Bar3D(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Bar3D returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
	if got, want := collection.Alpha, 0.0; got != want {
		t.Fatalf("default Bar3D edge alpha = %v, want Matplotlib default edgecolors-none alpha %v", got, want)
	}
	foundFaces := false
	for _, artist := range ax.Artists {
		polys, ok := artist.(*PolyCollection)
		if ok && len(polys.Polygons) == 6*2 {
			foundFaces = true
			break
		}
	}
	if !foundFaces {
		t.Fatal("Bar3D did not add filled projected cuboid faces")
	}
}

func TestAxes3DTrisurfCreatesSinglePolyCollection(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	polyCount := 0
	lineCount := 0
	for _, artist := range ax.Artists {
		switch art := artist.(type) {
		case *PolyCollection:
			if len(art.Polygons) == 2 {
				polyCount++
			}
		case *LineCollection:
			lineCount++
		}
	}
	if polyCount != 1 || lineCount != 0 {
		t.Fatalf("Trisurf artists = %d matching PolyCollection, %d LineCollection; want one Poly3DCollection-equivalent and no separate edge collection", polyCount, lineCount)
	}
}

func TestAxes3DTrisurfAutoTriangulatesWhenTrianglesAreOmitted(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X: []float64{0, 1, 1, 0},
		Y: []float64{0, 0, 1, 1},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if len(collection.Polygons) != 2 {
		t.Fatalf("auto-triangulated polygon count = %d, want 2", len(collection.Polygons))
	}
}

func TestAxes3DTrisurfSkipsMaskedTriangles(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	tri := Triangulation{
		X:         []float64{0, 1, 1, 0},
		Y:         []float64{0, 0, 1, 1},
		Triangles: [][3]int{{0, 1, 2}, {0, 2, 3}},
		Mask:      []bool{false, true},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2, 3})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := len(collection.Polygons), 1; got != want {
		t.Fatalf("masked trisurf polygon count = %d, want %d visible triangle", got, want)
	}
}

func TestAxes3DTrisurfUsesConfiguredEdgeColor(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	edge := render.Color{R: 0.2, G: 0.3, B: 0.4, A: 1}
	width := 0.25
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	collection := ax.Trisurf(tri, []float64{0, 1, 2}, PlotOptions{
		EdgeColor: &edge,
		EdgeWidth: &width,
	})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got := collection.EdgeColor; got != edge {
		t.Fatalf("trisurf edge color = %+v, want %+v", got, edge)
	}
	if got := collection.EdgeWidth; got != width {
		t.Fatalf("trisurf edge width = %v, want %v", got, width)
	}
}

func TestAxes3DTrisurfShadesFaceColorsLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	base := render.Color{R: 1, G: 0.5, B: 0.05, A: 1}
	tri := Triangulation{
		X:         []float64{0, 1, 0},
		Y:         []float64{0, 0, 1},
		Triangles: [][3]int{{0, 1, 2}},
	}
	collection := ax.Trisurf(tri, []float64{0, 0, 1}, PlotOptions{Color: &base})
	if collection == nil {
		t.Fatal("Trisurf returned nil")
	}
	if got, want := len(collection.FaceColors), 1; got != want {
		t.Fatalf("trisurf face colors = %d, want %d shaded color per face", got, want)
	}
	if collection.FaceColors[0] == base {
		t.Fatalf("trisurf face color = %+v, want Matplotlib-style shaded variant of %+v", collection.FaceColors[0], base)
	}
}

func TestAxes3DVoxelsCullInternalFacesLikeMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	voxels := ax.Voxels([][][]bool{
		{{true}},
		{{true}},
	})
	if got, want := len(voxels), 2; got != want {
		t.Fatalf("voxel collection count = %d, want %d filled voxels", got, want)
	}
	totalFaces := 0
	for _, voxel := range voxels {
		totalFaces += len(voxel.Polygons)
	}
	if got, want := totalFaces, 10; got != want {
		t.Fatalf("visible voxel face count = %d, want %d after internal-face culling", got, want)
	}
}

func TestAxes3DVoxelCallsBarLikeSegments(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	collection := ax.Voxel(
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{0, 1},
		[]float64{1, 1},
		[]float64{1, 1},
		[]float64{1, 1},
	)
	if collection == nil {
		t.Fatal("Voxel returned nil")
	}
	if got, want := len(collection.Segments), 16; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
}

func TestAxes3DText3DProjectsInput(t *testing.T) {
	fig := NewFigure(640, 480)
	ax, err := fig.AddAxes3D(unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	ax.SetDistance(0)
	ax.SetView(0, 0)
	text := ax.Text3D(1, 2, 3, "hello")
	if text == nil || text.Content != "hello" {
		t.Fatalf("Text3D returned unexpected value: %#v", text)
	}
	if !approx(text.Position.X, 1, 1e-12) || !approx(text.Position.Y, 2, 1e-12) {
		t.Fatalf("Text position = %+v, want {1 2}", text.Position)
	}
}
