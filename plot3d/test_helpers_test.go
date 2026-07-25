package plot3d

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

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

//nolint:gocritic // A four-corner plane is a small fixed geometry value in these projection tests.
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
		points[i] = ax.project3DPointWithState(x, y, z, mins, maxs)
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
	dashes [][]float64
}

func (r *axes3DLineWidthRecorder) Path(_ geom.Path, paint *render.Paint) {
	if paint == nil || paint.Stroke.A <= 0 || paint.LineWidth <= 0 {
		return
	}
	r.widths = append(r.widths, paint.LineWidth)
	r.colors = append(r.colors, paint.Stroke)
	r.dashes = append(r.dashes, append([]float64(nil), paint.Dashes...))
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

func countString(items []string, want string) int {
	count := 0
	for _, got := range items {
		if got == want {
			count++
		}
	}
	return count
}

func containsDashPattern(values [][]float64, want []float64) bool {
	for _, got := range values {
		if len(got) != len(want) {
			continue
		}
		matches := true
		for i := range got {
			if !approx(got[i], want[i], 1e-12) {
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

func latestBar3DFaceCollection(t *testing.T, ax *Axes3D, faceCount int) *PolyCollection {
	t.Helper()
	for i := len(ax.Artists) - 1; i >= 0; i-- {
		polys, ok := ax.Artists[i].(*PolyCollection)
		if ok && len(polys.Polygons) == faceCount {
			return polys
		}
	}
	t.Fatalf("Bar3D did not add PolyCollection with %d faces", faceCount)
	return nil
}

func countFaceColorAlpha(colors []render.Color, alpha float64) int {
	count := 0
	for _, color := range colors {
		if approx(color.A, alpha, 1e-12) {
			count++
		}
	}
	return count
}

func assertVoxelCollectionColors(t *testing.T, voxel *PolyCollection, face, edge render.Color) {
	t.Helper()
	if voxel == nil {
		t.Fatal("missing voxel collection")
	}
	if len(voxel.FaceColors) == 0 {
		t.Fatal("voxel has no visible face colors")
	}
	for i, got := range voxel.FaceColors {
		if got != face {
			t.Fatalf("voxel face color %d = %+v, want %+v", i, got, face)
		}
	}
	if got := voxel.EdgeColor; got != edge {
		t.Fatalf("voxel edge color = %+v, want %+v", got, edge)
	}
	if got, want := voxel.EdgeWidth, 1.0; got != want {
		t.Fatalf("voxel edge width = %v, want %v", got, want)
	}
	if array := voxel.Array(); len(array) != 0 {
		t.Fatalf("voxel scalar array = %v, want non-scalar-mappable PolyCollection", array)
	}
	mapping := voxel.ScalarMap()
	if mapping.Colormap != "" || mapping.Norm != nil || mapping.VMin != 0 || mapping.VMax != 0 {
		t.Fatalf("voxel scalar map = %+v, want no scalar-map metadata", mapping)
	}
}
