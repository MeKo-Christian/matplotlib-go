package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

// showcaseTriangulation returns the mesh and field used by the
// unstructured_showcase parity case, matching
// test/matplotlib_ref/plots/unstructured_showcase.py.
func showcaseTriangulation() (Triangulation, []float64) {
	x := []float64{0.0, 0.85, 1.75, 2.85, 0.2, 1.1, 2.1, 0.55, 1.55, 2.55}
	y := []float64{0.0, 0.2, 0.05, 0.3, 1.0, 1.15, 1.25, 2.15, 2.3, 2.05}
	triangles := [][3]int{
		{0, 1, 4},
		{1, 5, 4},
		{1, 2, 5},
		{2, 6, 5},
		{2, 3, 6},
		{4, 5, 7},
		{5, 8, 7},
		{5, 6, 8},
		{6, 9, 8},
	}
	values := make([]float64, len(x))
	for i := range x {
		values[i] = math.Sin(x[i]*1.4) + 0.7*math.Cos((y[i]+0.15)*2.1)
	}
	return Triangulation{X: x, Y: y, Triangles: triangles}, values
}

// showcaseLevels are the levels Matplotlib's MaxNLocator picks for
// tricontour(..., levels=6) on the showcase field.
func showcaseLevels() []float64 {
	return []float64{
		-0.6000000000000001,
		-0.30000000000000004,
		0.0,
		0.30000000000000004,
		0.6000000000000001,
		0.9000000000000001,
		1.2000000000000002,
		1.5000000000000004,
	}
}

// TestTriContourPolylinesMatchMatplotlibTraversal pins the TriContourGenerator
// port to vertex lists dumped from matplotlib 3.10.9. Order and direction are
// part of the contract: ContourLabeler places inline labels at a vertex index,
// so a line traced backwards labels a different point.
func TestTriContourPolylinesMatchMatplotlibTraversal(t *testing.T) {
	tri, values := showcaseTriangulation()
	polylines, levels := triContourPolylines(tri, values, showcaseLevels())

	want := []struct {
		level float64
		pts   []geom.Pt
	}{
		{-0.30000000000000004, []geom.Pt{
			{X: 2.822649694221663, Y: 0.29378402141401422},
			{X: 2.0631448456878871, Y: 1.123639470929898},
			{X: 1.8799922915454563, Y: 1.2279992291545456},
			{X: 2.036394742131201, Y: 1.3714282195677074},
			{X: 2.4374459599836293, Y: 2.0781385100040923},
		}},
		{0, []geom.Pt{
			{X: 2.6185851329851424, Y: 0.24740571204207781},
			{X: 2.003571094925626, Y: 0.91938661117357556},
			{X: 1.5243653885008706, Y: 1.192436538850087},
			{X: 1.9335813405821054, Y: 1.5677083497977988},
			{X: 2.2498589448832322, Y: 2.1250352637791918},
		}},
		{0, []geom.Pt{
			{X: 0.28428113870282257, Y: 1.2769237414521313},
			{X: 0.56731767160593238, Y: 1.061219611934322},
			{X: 0.29469145110206851, Y: 0.88345667556668483},
			{X: 0.14588976671033527, Y: 0.72944883355167633},
		}},
		{0.30000000000000004, []geom.Pt{
			{X: 2.4145205717486222, Y: 0.20102740267014141},
			{X: 1.9439973441633656, Y: 0.71513375141725322},
			{X: 1.1687384854562848, Y: 1.1568738485456285},
			{X: 1.83076793903301, Y: 1.7639884800278902},
			{X: 2.0622719297828351, Y: 2.1719320175542913},
		}},
		{0.30000000000000004, []geom.Pt{
			{X: 0.38670744554161246, Y: 1.6134673210652979},
			{X: 1.0137163928177888, Y: 1.1356193988029646},
			{X: 0.40976934130226267, Y: 0.74182234916644596},
			{X: 0.080129960384176885, Y: 0.4006498019208844},
		}},
		{0.60000000000000009, []geom.Pt{
			{X: 2.2104560105121021, Y: 0.154649093298205},
			{X: 1.8844235934011049, Y: 0.51088089166093087},
			{X: 1.27118496827118, Y: 0.86030236138723359},
			{X: 1.0444751934249161, Y: 0.93900573501468088},
			{X: 0.52484723150245671, Y: 0.60018802276620709},
			{X: 0.014370154058018489, Y: 0.071850770290092444},
		}},
		{0.60000000000000009, []geom.Pt{
			{X: 0.4891337523804023, Y: 1.9500109006784645},
			{X: 0.78329379249270104, Y: 1.7258294681950892},
			{X: 1.2430587276343212, Y: 1.5155945261765984},
			{X: 1.7279545374839145, Y: 1.9602686102579814},
			{X: 1.874684914682438, Y: 2.2188287713293904},
		}},
		{0.90000000000000013, []geom.Pt{
			{X: 2.0063914492755819, Y: 0.10827078392626857},
			{X: 1.8248498426388442, Y: 0.30662803190460858},
			{X: 1.4833858672333275, Y: 0.50119314775898394},
			{X: 0.97564663920707595, Y: 0.67745722898688809},
			{X: 0.63992512170265081, Y: 0.45855369636596821},
			{X: 0.2547985304271555, Y: 0.05995259539462483},
		}},
		{0.90000000000000013, []geom.Pt{
			{X: 0.90698909719130638, Y: 2.2035483645786957},
			{X: 1.4203943366832072, Y: 1.9687855270793073},
			{X: 1.6251411359348191, Y: 2.1565487404880725},
			{X: 1.6870978995820407, Y: 2.2657255251044894},
		}},
		{1.2000000000000002, []geom.Pt{
			{X: 1.8023268880390617, Y: 0.061892474554332208},
			{X: 1.7652760918765833, Y: 0.10237517214828629},
			{X: 1.6955867661954751, Y: 0.14208393413073428},
			{X: 0.90681808498923555, Y: 0.41590872295909531},
			{X: 0.75500301190284502, Y: 0.31691936996572928},
			{X: 0.58084669914511067, Y: 0.13666981156355545},
		}},
	}

	if len(polylines) != len(want) {
		t.Fatalf("got %d polylines, want %d", len(polylines), len(want))
	}
	const tol = 1e-12
	for i, expected := range want {
		if levels[i] != expected.level {
			t.Errorf("polyline %d: level = %v, want %v", i, levels[i], expected.level)
		}
		got := polylines[i]
		if len(got) != len(expected.pts) {
			t.Errorf("polyline %d: %d vertices, want %d (%v)", i, len(got), len(expected.pts), got)
			continue
		}
		for j, pt := range expected.pts {
			if math.Abs(got[j].X-pt.X) > tol || math.Abs(got[j].Y-pt.Y) > tol {
				t.Errorf("polyline %d vertex %d = (%.17g, %.17g), want (%.17g, %.17g)",
					i, j, got[j].X, got[j].Y, pt.X, pt.Y)
			}
		}
	}
}

// TestTriContourLabelsLandOnMatplotlibVertices asserts the label anchor each
// contour line yields, which is locate_label's answer for these short lines:
// the vertex at index len/2. These are matplotlib's labelXYs mapped back to
// data space.
func TestTriContourLabelsLandOnMatplotlibVertices(t *testing.T) {
	tri, values := showcaseTriangulation()
	polylines, _ := triContourPolylines(tri, values, showcaseLevels())

	want := []geom.Pt{
		{X: 1.8799922915454563, Y: 1.2279992291545456},
		{X: 1.5243653885008706, Y: 1.192436538850087},
		{X: 0.29469145110206851, Y: 0.88345667556668483},
		{X: 1.1687384854562848, Y: 1.1568738485456285},
		{X: 0.40976934130226267, Y: 0.74182234916644596},
		{X: 1.0444751934249161, Y: 0.93900573501468088},
		{X: 1.2430587276343212, Y: 1.5155945261765984},
		{X: 0.97564663920707595, Y: 0.67745722898688809},
		{X: 1.6251411359348191, Y: 2.1565487404880725},
		{X: 0.90681808498923555, Y: 0.41590872295909531},
	}

	if len(polylines) != len(want) {
		t.Fatalf("got %d polylines, want %d", len(polylines), len(want))
	}
	const tol = 1e-12
	for i, expected := range want {
		got := polylines[i][len(polylines[i])/2]
		if math.Abs(got.X-expected.X) > tol || math.Abs(got.Y-expected.Y) > tol {
			t.Errorf("label anchor %d = (%.17g, %.17g), want (%.17g, %.17g)",
				i, got.X, got.Y, expected.X, expected.Y)
		}
	}
}
