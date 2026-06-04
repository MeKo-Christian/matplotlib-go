package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestAxesQuiverGridScalarMap(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	q := ax.QuiverGrid(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{
			{1, 0},
			{0, -1},
		},
		[][]float64{
			{0, 1},
			{-1, 0},
		},
		QuiverOptions{
			CGrid: [][]float64{
				{1, 2},
				{3, 4},
			},
			Label: "field",
		},
	)
	if q == nil {
		t.Fatal("expected quiver artist")
	}
	if len(q.Anchors) != 4 {
		t.Fatalf("len(q.Anchors) = %d, want 4", len(q.Anchors))
	}
	if len(q.ScalarColors) != 4 {
		t.Fatalf("len(q.ScalarColors) = %d, want 4", len(q.ScalarColors))
	}
	mapping := q.ScalarMap()
	if mapping.Colormap != "viridis" || mapping.VMin != 1 || mapping.VMax != 4 {
		t.Fatalf("unexpected scalar map %+v", mapping)
	}
}

func TestAxesQuiverGridAcceptsSharedNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	q := ax.QuiverGrid(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{1, 0}, {0, -1}},
		[][]float64{{0, 1}, {-1, 0}},
		QuiverOptions{
			CGrid: [][]float64{{1, 10}, {100, 1000}},
			Norm:  LogNorm{VMin: 1, VMax: 1000},
		},
	)
	if q == nil {
		t.Fatal("expected quiver artist")
	}
	mapping := q.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("quiver norm = %#v, want log norm", mapping.Norm)
	}
	if got := mapping.Normalize(10); got < 0.32 || got > 0.34 {
		t.Fatalf("log-normalized scalar 10 = %v, want about 1/3", got)
	}
}

func TestAxesBarbsGridAcceptsSharedNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	b := ax.BarbsGrid(
		[]float64{0, 1},
		[]float64{0, 1},
		[][]float64{{1, 0}, {0, -1}},
		[][]float64{{0, 1}, {-1, 0}},
		BarbsOptions{
			CGrid: [][]float64{{1, 10}, {100, 1000}},
			Norm:  LogNorm{VMin: 1, VMax: 1000},
		},
	)
	if b == nil {
		t.Fatal("expected barbs artist")
	}
	mapping := b.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("barbs norm = %#v, want log norm", mapping.Norm)
	}
}

func TestQuiverPathsRespectPivot(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.2},
		Max: geom.Pt{X: 0.85, Y: 0.8},
	})
	ax.SetXLim(0, 2)
	ax.SetYLim(0, 2)
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	tail := &Quiver{
		Anchors:        []geom.Pt{{X: 1, Y: 1}},
		U:              []float64{1},
		V:              []float64{0},
		Color:          render.Color{R: 0.2, G: 0.3, B: 0.7, A: 1},
		Pivot:          vectorPivotTail,
		Angles:         quiverAnglesUV,
		Units:          "dots",
		ScaleUnits:     "dots",
		Width:          4,
		Scale:          1,
		ScaleSet:       true,
		HeadWidth:      3,
		HeadLength:     5,
		HeadAxisLength: 4.5,
		MinShaft:       1,
		MinLength:      1,
		forceLengthPx:  20,
	}
	middle := *tail
	middle.Pivot = vectorPivotMiddle

	tailPaths, _ := tail.pathsForContext(ctx)
	middlePaths, _ := middle.pathsForContext(ctx)

	tailBounds, ok := pathBounds(tailPaths[0])
	if !ok {
		t.Fatal("expected tail quiver path bounds")
	}
	middleBounds, ok := pathBounds(middlePaths[0])
	if !ok {
		t.Fatal("expected middle quiver path bounds")
	}
	if math.Abs(tailBounds.Min.X) > 1e-6 {
		t.Fatalf("tail pivot should start at the anchor, got min x %.2f", tailBounds.Min.X)
	}
	if !(middleBounds.Min.X < tailBounds.Min.X && middleBounds.Max.X < tailBounds.Max.X) {
		t.Fatalf("middle pivot should shift path around the anchor: tail=%+v middle=%+v", tailBounds, middleBounds)
	}
}

func TestQuiverExplicitAngles(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.2},
		Max: geom.Pt{X: 0.85, Y: 0.8},
	})
	ax.SetXLim(0, 2)
	ax.SetYLim(0, 2)
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	q := &Quiver{
		Anchors:        []geom.Pt{{X: 1, Y: 1}},
		U:              []float64{1},
		V:              []float64{0},
		AngleValues:    []float64{90},
		Color:          render.Color{R: 0.2, G: 0.3, B: 0.7, A: 1},
		Pivot:          vectorPivotTail,
		Angles:         quiverAnglesUV,
		Units:          "dots",
		ScaleUnits:     "dots",
		Width:          4,
		Scale:          1,
		ScaleSet:       true,
		HeadWidth:      3,
		HeadLength:     5,
		HeadAxisLength: 4.5,
		MinShaft:       1,
		MinLength:      1,
		forceLengthPx:  16,
	}
	vector, ok := q.directionVectorAt(ctx, 0)
	if !ok {
		t.Fatal("expected display vector")
	}
	if math.Abs(vector.X) > 1e-6 || vector.Y <= 0 {
		t.Fatalf("explicit 90-degree angle should point vertically, got %+v", vector)
	}
}

func TestQuiverUVDirectionUsesMatplotlibComplexAngle(t *testing.T) {
	ctx := createTestDrawContext()
	q := &Quiver{
		Anchors:       []geom.Pt{{X: 1, Y: 1}},
		U:             []float64{0},
		V:             []float64{1},
		Pivot:         vectorPivotTail,
		Angles:        quiverAnglesUV,
		Units:         "dots",
		ScaleUnits:    "dots",
		Width:         4,
		Scale:         1,
		ScaleSet:      true,
		forceLengthPx: 12,
	}

	vector, ok := q.directionVectorAt(ctx, 0)
	if !ok {
		t.Fatal("expected display vector")
	}
	if math.Abs(vector.X) > 1e-9 || vector.Y <= 0 {
		t.Fatalf("angles='uv' should follow Matplotlib np.angle(U+Vi), got %+v", vector)
	}
}

func TestQuiverKeyDrawOverlay(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.SetXLim(0, 2)
	ax.SetYLim(0, 2)

	q := ax.Quiver(
		[]float64{0.5, 1.0},
		[]float64{0.5, 1.0},
		[]float64{1, 0.5},
		[]float64{0.25, 0.75},
		QuiverOptions{
			Color:      &render.Color{R: 0.2, G: 0.4, B: 0.7, A: 1},
			ScaleUnits: "dots",
			Units:      "dots",
			Scale:      floatPtr(24),
			Width:      floatPtr(4),
		},
	)
	if q == nil {
		t.Fatal("expected quiver artist")
	}
	ax.QuiverKey(q, 0.8, 0.2, 1, "1 unit", QuiverKeyOptions{
		Coords:   Coords(CoordAxes),
		LabelPos: "E",
	})

	var ren vectorTextRenderer
	DrawFigure(fig, &ren)
	if ren.pathCount == 0 {
		t.Fatal("expected quiver key path draw")
	}
	if len(ren.texts) == 0 || ren.texts[0] != "1 unit" {
		t.Fatalf("unexpected quiver key labels %v", ren.texts)
	}
}

func TestQuiverKeyDefaultLabelSeparationMatchesMatplotlib(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(unitRect())
	q := ax.Quiver([]float64{0}, []float64{0}, []float64{1}, []float64{0})
	if q == nil {
		t.Fatal("expected quiver artist")
	}

	key := ax.QuiverKey(q, 0.8, 0.2, 1, "1 unit", QuiverKeyOptions{LabelPos: "E"})
	if key == nil {
		t.Fatal("expected quiver key")
	}
	if key.LabelSep != 10 {
		t.Fatalf("default label separation = %v, want 10 px at 100 dpi", key.LabelSep)
	}
}

func TestBarbsDefaultLengthUsesPoints(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(unitRect())
	length := 6.0

	barbs := ax.Barbs(
		[]float64{0},
		[]float64{0},
		[]float64{10},
		[]float64{0},
		BarbsOptions{Length: &length},
	)
	if barbs == nil {
		t.Fatal("expected barbs artist")
	}
	if barbs.Units != "points" {
		t.Fatalf("barb units = %q, want points", barbs.Units)
	}
}

func TestBarbsPointsLengthMatchesMatplotlibCollectionScale(t *testing.T) {
	length := 6.0
	b := &Barbs{
		Anchors:    []geom.Pt{{X: 1, Y: 1}},
		U:          []float64{10},
		V:          []float64{0},
		BarbColor:  render.Color{A: 1},
		FlagColor:  render.Color{A: 1},
		LineWidth:  1,
		Pivot:      vectorPivotTip,
		Length:     length,
		Units:      "points",
		Sizes:      defaultBarbSizes(nil),
		Increments: defaultBarbIncrements(nil),
		Rounding:   true,
	}

	collection := b.asPathCollection(createTestDrawContext())
	if len(collection.Paths) != 1 || len(collection.Paths[0].V) < 2 {
		t.Fatalf("expected one display-space barb path, got %+v", collection.Paths)
	}
	wantLength := length * (100.0 / 72.0) * (length / 2.0)
	if !approx(collection.Paths[0].V[1].X, -wantLength, 1e-9) {
		t.Fatalf("barb staff endpoint X = %v, want %v", collection.Paths[0].V[1].X, -wantLength)
	}
}

func TestBarbsDirectionUsesRawVectorAngleLikeMatplotlib(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.15, Y: 0.2},
		Max: geom.Pt{X: 0.85, Y: 0.8},
	})
	ax.SetXLim(0, 100)
	ax.SetYLim(0, 1)
	ctx := newAxesDrawContext(ax, fig, fig.DisplayRect(), ax.adjustedLayout(fig))

	b := &Barbs{
		Anchors:    []geom.Pt{{X: 50, Y: 0.5}},
		U:          []float64{10},
		V:          []float64{10},
		BarbColor:  render.Color{A: 1},
		FlagColor:  render.Color{A: 1},
		LineWidth:  1,
		Pivot:      vectorPivotTip,
		Length:     18,
		Units:      "dots",
		Sizes:      defaultBarbSizes(nil),
		Increments: defaultBarbIncrements(nil),
		Rounding:   true,
	}

	collection := b.asPathCollection(ctx)
	if len(collection.Paths) != 1 || len(collection.Paths[0].V) < 2 {
		t.Fatalf("expected one display-space barb path, got %+v", collection.Paths)
	}
	end := collection.Paths[0].V[1]
	want := -18 / math.Sqrt2
	if !approx(end.X, want, 1e-9) || !approx(end.Y, want, 1e-9) {
		t.Fatalf("barb staff endpoint = %+v, want raw 45-degree angle endpoint (%v, %v)", end, want, want)
	}
}

func TestBarbsFindTailsAndFlip(t *testing.T) {
	b := &Barbs{
		U:          []float64{35},
		V:          []float64{0},
		BarbColor:  render.Color{R: 0.3, G: 0.2, B: 0.1, A: 1},
		FlagColor:  render.Color{R: 0.3, G: 0.2, B: 0.1, A: 1},
		LineWidth:  1,
		Alpha:      1,
		Pivot:      vectorPivotTip,
		Length:     18,
		Units:      "dots",
		Sizes:      defaultBarbSizes(nil),
		Increments: defaultBarbIncrements(nil),
		Rounding:   true,
		Flip:       []bool{true},
	}

	nFlags, nBarbs, half, empty := b.findTails(35)
	if nFlags != 0 || nBarbs != 3 || !half || empty {
		t.Fatalf("unexpected tail decomposition flags=%d barbs=%d half=%v empty=%v", nFlags, nBarbs, half, empty)
	}
	path := b.barbGlyphPath(18, 0, false)
	if len(path.C) == 0 {
		t.Fatal("expected barb glyph path")
	}
}

func TestBarbGlyphPathMatchesMatplotlibDisplayGeometry(t *testing.T) {
	b := &Barbs{
		U:          []float64{20, 50, 5},
		V:          []float64{0, 0, 0},
		BarbColor:  render.Color{R: 0.3, G: 0.2, B: 0.1, A: 1},
		FlagColor:  render.Color{R: 0.3, G: 0.2, B: 0.1, A: 1},
		LineWidth:  1,
		Alpha:      1,
		Pivot:      vectorPivotTip,
		Length:     28,
		Units:      "dots",
		Sizes:      defaultBarbSizes(nil),
		Increments: defaultBarbIncrements(nil),
		Rounding:   true,
	}

	full := b.barbGlyphPath(28, 0, false)
	if len(full.C) == 0 || full.C[len(full.C)-1] != geom.ClosePath {
		t.Fatalf("full barb path should close like Matplotlib PolyCollection, got commands %+v", full.C)
	}
	wantFull := []geom.Pt{
		{X: 0, Y: 0},
		{X: -28, Y: 0},
		{X: -31.5, Y: 11.2},
		{X: -28, Y: 0},
		{X: -24.5, Y: 0},
		{X: -28, Y: 11.2},
		{X: -24.5, Y: 0},
	}
	assertPathVerticesClose(t, full, wantFull)

	flag := b.barbGlyphPath(28, 1, false)
	wantFlag := []geom.Pt{
		{X: 0, Y: 0},
		{X: -28, Y: 0},
		{X: -24.5, Y: 11.2},
		{X: -21, Y: 0},
	}
	assertPathVerticesClose(t, flag, wantFlag)

	half := b.barbGlyphPath(28, 2, false)
	wantHalf := []geom.Pt{
		{X: 0, Y: 0},
		{X: -28, Y: 0},
		{X: -22.75, Y: 0},
		{X: -24.5, Y: 5.6},
		{X: -22.75, Y: 0},
	}
	assertPathVerticesClose(t, half, wantHalf)
}

func TestAxesStreamplotProducesLinesAndArrows(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	x := []float64{0, 1, 2, 3}
	y := []float64{0, 1, 2}
	u := [][]float64{
		{1, 1, 1, 1},
		{1, 1, 1, 1},
		{1, 1, 1, 1},
	}
	v := [][]float64{
		{0.2, 0.2, 0.2, 0.2},
		{0.2, 0.2, 0.2, 0.2},
		{0.2, 0.2, 0.2, 0.2},
	}

	set := ax.Streamplot(x, y, u, v, StreamplotOptions{
		Density:     0.35,
		StartPoints: []geom.Pt{{X: 0.5, Y: 0.5}, {X: 0.5, Y: 1.5}},
		ArrowCount:  intPtr(2),
		Label:       "stream",
	})
	if set == nil || set.Lines == nil || set.Arrows == nil {
		t.Fatal("expected streamplot set")
	}
	if len(set.Lines.Segments) == 0 {
		t.Fatal("expected streamline segments")
	}
	if len(set.Arrows.Anchors) == 0 {
		t.Fatal("expected sampled streamline arrows")
	}
}

func TestAxesStreamplotAcceptsSharedNorm(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	set := ax.Streamplot(
		[]float64{0, 1, 2},
		[]float64{0, 1, 2},
		[][]float64{
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
		[][]float64{
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		StreamplotOptions{
			CGrid: [][]float64{
				{1, 10, 100},
				{1, 10, 100},
				{1, 10, 100},
			},
			Norm:       LogNorm{VMin: 1, VMax: 100},
			ArrowCount: intPtr(1),
		},
	)
	if set == nil {
		t.Fatal("expected streamplot set")
	}
	mapping := set.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("streamplot norm = %#v, want log norm", mapping.Norm)
	}
	if set.Lines == nil || len(set.Lines.Colors) == 0 {
		t.Fatal("expected scalar-colored stream lines")
	}
}

func TestAxesStreamplotColorbarUsesLineScalarMap(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	set := ax.Streamplot(
		[]float64{0, 1, 2},
		[]float64{0, 1, 2},
		[][]float64{
			{1, 1, 1},
			{1, 1, 1},
			{1, 1, 1},
		},
		[][]float64{
			{0, 0, 0},
			{0, 0, 0},
			{0, 0, 0},
		},
		StreamplotOptions{
			CGrid: [][]float64{
				{1, 10, 100},
				{1, 10, 100},
				{1, 10, 100},
			},
			Norm:       LogNorm{VMin: 1, VMax: 100},
			ArrowCount: intPtr(0),
		},
	)
	if set == nil {
		t.Fatal("expected streamplot set")
	}
	cbAx := fig.AddColorbar(ax, set)
	if cbAx == nil || len(cbAx.Artists) == 0 {
		t.Fatal("expected colorbar for scalar-colored streamplot")
	}
	cb, ok := cbAx.Artists[0].(*Colorbar)
	if !ok {
		t.Fatalf("colorbar artist type = %T, want *Colorbar", cbAx.Artists[0])
	}
	if cb.Mapping.Norm == nil || cb.Mapping.Norm.NormName() != "log" {
		t.Fatalf("colorbar norm = %#v, want log norm", cb.Mapping.Norm)
	}
}

func TestStreamplotSetScalarMapDelegatesArrowMappable(t *testing.T) {
	set := &StreamplotSet{
		Arrows: &Quiver{
			ScalarColors: []float64{1, 10, 100},
			Colormap:     "viridis",
			VMin:         1,
			VMax:         100,
			Norm:         LogNorm{VMin: 1, VMax: 100},
		},
	}
	mapping := set.ScalarMap()
	if mapping.Norm == nil || mapping.Norm.NormName() != "log" {
		t.Fatalf("streamplot scalar norm = %#v, want log norm", mapping.Norm)
	}
}

func assertPathVerticesClose(t *testing.T, got geom.Path, want []geom.Pt) {
	t.Helper()
	if len(got.V) != len(want) {
		t.Fatalf("path vertices = %d, want %d: %+v", len(got.V), len(want), got.V)
	}
	for i := range want {
		if math.Abs(got.V[i].X-want[i].X) > 1e-9 || math.Abs(got.V[i].Y-want[i].Y) > 1e-9 {
			t.Fatalf("vertex %d = %+v, want %+v", i, got.V[i], want[i])
		}
	}
}

type vectorTextRenderer struct {
	render.NullRenderer
	pathCount int
	texts     []string
}

func (r *vectorTextRenderer) Path(_ geom.Path, _ *render.Paint) {
	r.pathCount++
}

func (r *vectorTextRenderer) DrawText(text string, _ geom.Pt, _ float64, _ render.Color) {
	r.texts = append(r.texts, text)
}

func (r *vectorTextRenderer) MeasureText(text string, size float64, _ string) render.TextMetrics {
	return render.TextMetrics{
		W:       float64(len(text)) * size * 0.5,
		H:       size,
		Ascent:  size * 0.8,
		Descent: size * 0.2,
	}
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}
