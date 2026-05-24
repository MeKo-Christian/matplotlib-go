package core

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/transform"
)

// createTestDrawContext creates a valid DrawContext for testing
func createTestDrawContext() *DrawContext {
	return &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 10),
			YScale:      transform.NewLinear(0, 10),
			AxesToPixel: transform.NewAffine(geom.Affine{A: 100, D: -100, E: 50, F: 450}), // 500x500 viewport
		},
		RC:   style.Default,
		Clip: geom.Rect{Min: geom.Pt{X: 0, Y: 0}, Max: geom.Pt{X: 500, Y: 500}},
	}
}

func TestScatter2D_Draw(t *testing.T) {
	// Create a scatter with basic data
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: 2, Y: 2},
		},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Marker: MarkerCircle,
	}

	// Test that Draw doesn't panic with null renderer
	renderer := &render.NullRenderer{}

	// Create a proper DrawContext with transforms
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatterUsesIndependentShapeColorCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	palette := fig.RC.Palette()

	firstLine := ax.Plot([]float64{0, 1}, []float64{0, 1})
	scatter := ax.Scatter([]float64{0.5}, []float64{0.5})
	secondLine := ax.Plot([]float64{0, 1}, []float64{1, 0})

	if got, want := firstLine.Col, palette[0]; got != want {
		t.Fatalf("first line color = %+v, want %+v", got, want)
	}
	if got, want := scatter.Color, palette[0]; got != want {
		t.Fatalf("scatter color = %+v, want independent shape cycle first color %+v", got, want)
	}
	if got, want := secondLine.Col, palette[1]; got != want {
		t.Fatalf("second line color = %+v, want line cycle second color %+v", got, want)
	}
}

func TestScatterDefaultsUseMatplotlibFaceEdges(t *testing.T) {
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	scatter := ax.Scatter([]float64{0.5}, []float64{0.5})
	if scatter == nil {
		t.Fatal("Scatter returned nil")
	}
	if got, want := scatter.EdgeColor, scatter.Color; got != want {
		t.Fatalf("scatter default edge color = %+v, want face color %+v", got, want)
	}
	if got, want := scatter.EdgeWidth, 1.0; got != want {
		t.Fatalf("scatter default edge width = %v, want Matplotlib linewidth %v", got, want)
	}
}

func TestScatterHalfFilledMarkerDrawsSplitFillAndWholeEdge(t *testing.T) {
	markerStyle := NewMarkerStyle(MarkerSquare)
	markerStyle.FillStyle = MarkerFillTop
	scatter := &Scatter2D{
		XY:          []geom.Pt{{X: 0.5, Y: 0.5}},
		Size:        9,
		Color:       render.Color{R: 0.2, G: 0.7, B: 0.8, A: 1},
		EdgeColor:   render.Color{R: 0.05, G: 0.2, B: 0.25, A: 1},
		EdgeWidth:   2,
		MarkerStyle: markerStyle,
	}

	r := &recordingRenderer{}
	ctx := &DrawContext{
		DataToPixel: Transform2D{
			XScale:      transform.NewLinear(0, 1),
			YScale:      transform.NewLinear(0, 1),
			AxesToPixel: transform.NewAffine(geom.Identity()),
		},
		RC: style.Default,
	}
	scatter.Draw(r, ctx)

	if len(r.pathCalls) != 2 {
		t.Fatalf("path calls = %d, want primary half fill and whole edge", len(r.pathCalls))
	}
	if got, want := r.pathCalls[0].paint.Fill, scatter.Color; got != want {
		t.Fatalf("primary half fill = %+v, want %+v", got, want)
	}
	if got := r.pathCalls[0].paint.Stroke.A; got != 0 {
		t.Fatalf("primary half stroke alpha = %v, want 0", got)
	}
	if got := r.pathCalls[1].paint.Fill.A; got != 0 {
		t.Fatalf("edge pass fill alpha = %v, want 0", got)
	}
	if got, want := r.pathCalls[1].paint.Stroke, scatter.EdgeColor; got != want {
		t.Fatalf("edge stroke = %+v, want %+v", got, want)
	}
}

func TestScatter2D_EmptyData(t *testing.T) {
	// Test with empty data
	scatter := &Scatter2D{
		XY:     []geom.Pt{},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with empty data
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_VariableSizesAndColors(t *testing.T) {
	// Test with variable sizes and colors
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		Sizes: []float64{3.0, 7.0},
		Colors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
		},
		Size:   5.0,                                  // fallback size
		Color:  render.Color{R: 0, G: 0, B: 1, A: 1}, // fallback color
		Marker: MarkerSquare,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_SizeUsesMatplotlibAreaSemantics(t *testing.T) {
	ctx := createTestDrawContext()
	ctx.RC.DPI = 144

	scatter := &Scatter2D{
		XY:   []geom.Pt{{X: 0, Y: 0}},
		Size: 36,
	}
	pc := scatter.toPathCollection(nil, ctx)
	want := 6.0 * 144.0 / 72.0
	if got := pc.Size; got != want {
		t.Fatalf("scatter scale = %v, want sqrt(area)*dpi/72 = %v", got, want)
	}
}

func TestScatterCircleMarkerPrototypeMatchesMatplotlibUnitMarker(t *testing.T) {
	circle := (&Scatter2D{Marker: MarkerCircle}).markerPrototypePath()
	bounds, ok := pathBounds(circle)
	if !ok {
		t.Fatal("circle marker prototype has no bounds")
	}
	if !approx(bounds.Min.X, -0.5, 1e-12) ||
		!approx(bounds.Min.Y, -0.5, 1e-12) ||
		!approx(bounds.Max.X, 0.5, 1e-12) ||
		!approx(bounds.Max.Y, 0.5, 1e-12) {
		t.Fatalf("circle marker bounds = %+v, want Matplotlib unit circle scaled to [-0.5,0.5]", bounds)
	}

	point := (&Scatter2D{Marker: MarkerPoint}).markerPrototypePath()
	bounds, ok = pathBounds(point)
	if !ok {
		t.Fatal("point marker prototype has no bounds")
	}
	if !approx(bounds.Min.X, -0.25, 1e-12) ||
		!approx(bounds.Min.Y, -0.25, 1e-12) ||
		!approx(bounds.Max.X, 0.25, 1e-12) ||
		!approx(bounds.Max.Y, 0.25, 1e-12) {
		t.Fatalf("point marker bounds = %+v, want Matplotlib point marker scaled to [-0.25,0.25]", bounds)
	}
}

func TestScatterAreaFromRadius(t *testing.T) {
	got := ScatterAreaFromRadius(8, 100)
	want := math.Pi * 8 * 72.0 / 100.0 * (8 * 72.0 / 100.0)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("ScatterAreaFromRadius(8, 100) = %v, want %v", got, want)
	}

	for _, tc := range []struct {
		radius float64
		dpi    float64
	}{
		{radius: 0, dpi: 100},
		{radius: 8, dpi: 0},
		{radius: -1, dpi: 100},
	} {
		if got := ScatterAreaFromRadius(tc.radius, tc.dpi); got != 0 {
			t.Fatalf("ScatterAreaFromRadius(%v, %v) = %v, want 0", tc.radius, tc.dpi, got)
		}
	}
}

func TestScatter2D_AllMarkerTypes(t *testing.T) {
	markerTypes := []MarkerType{
		MarkerCircle, MarkerSquare, MarkerTriangle, MarkerDiamond, MarkerPlus, MarkerCross,
		MarkerPixel, MarkerPoint, MarkerTriangleDown, MarkerTriangleLeft, MarkerTriangleRight,
		MarkerTriDown, MarkerTriUp, MarkerTriLeft, MarkerTriRight, MarkerOctagon, MarkerPentagon,
		MarkerStar, MarkerHexagon1, MarkerHexagon2, MarkerFilledX, MarkerFilledPlus,
		MarkerThinDiamond, MarkerVLine, MarkerHLine, MarkerTickLeft, MarkerTickRight,
		MarkerTickUp, MarkerTickDown, MarkerCaretLeft, MarkerCaretRight, MarkerCaretUp,
		MarkerCaretDown, MarkerCaretLeftBase, MarkerCaretRightBase, MarkerCaretUpBase,
		MarkerCaretDownBase, MarkerNone,
	}

	for _, markerType := range markerTypes {
		scatter := &Scatter2D{
			XY:     []geom.Pt{{X: 0, Y: 0}},
			Size:   5.0,
			Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
			Marker: markerType,
		}

		renderer := &render.NullRenderer{}
		ctx := createTestDrawContext()

		err := renderer.Begin(geom.Rect{})
		if err != nil {
			t.Fatalf("Failed to begin rendering for marker %v: %v", markerType, err)
		}

		// Should not panic for any marker type
		scatter.Draw(renderer, ctx)

		err = renderer.End()
		if err != nil {
			t.Fatalf("Failed to end rendering for marker %v: %v", markerType, err)
		}
	}
}

func TestMarkerTypeFromStringCoversMatplotlibAliases(t *testing.T) {
	for marker, want := range map[string]MarkerType{
		".":    MarkerPoint,
		",":    MarkerPixel,
		"o":    MarkerCircle,
		"v":    MarkerTriangleDown,
		"^":    MarkerTriangleUp,
		"<":    MarkerTriangleLeft,
		">":    MarkerTriangleRight,
		"1":    MarkerTriDown,
		"2":    MarkerTriUp,
		"3":    MarkerTriLeft,
		"4":    MarkerTriRight,
		"8":    MarkerOctagon,
		"s":    MarkerSquare,
		"p":    MarkerPentagon,
		"P":    MarkerFilledPlus,
		"*":    MarkerStar,
		"h":    MarkerHexagon1,
		"H":    MarkerHexagon2,
		"+":    MarkerPlus,
		"x":    MarkerCross,
		"X":    MarkerFilledX,
		"D":    MarkerDiamond,
		"d":    MarkerThinDiamond,
		"|":    MarkerVLine,
		"_":    MarkerHLine,
		"":     MarkerNone,
		" ":    MarkerNone,
		"none": MarkerNone,
		"None": MarkerNone,
	} {
		got, ok := MarkerTypeFromString(marker)
		if !ok {
			t.Fatalf("MarkerTypeFromString(%q) returned !ok", marker)
		}
		if got != want {
			t.Fatalf("MarkerTypeFromString(%q) = %v, want %v", marker, got, want)
		}
	}
	if _, ok := MarkerTypeFromString("not-a-marker"); ok {
		t.Fatal("unknown marker unexpectedly resolved")
	}
}

func TestMarkerStyleTupleMarkers(t *testing.T) {
	styles := []MarkerStyle{
		NewTupleMarkerStyle(5, MarkerTuplePolygon, 0),
		NewTupleMarkerStyle(5, MarkerTupleStar, 15),
		NewTupleMarkerStyle(6, MarkerTupleAsterisk, 30),
	}
	for _, style := range styles {
		scatter := &Scatter2D{MarkerStyle: style}
		path := scatter.markerPrototypePath()
		if len(path.C) == 0 || !path.Validate() {
			t.Fatalf("tuple style %+v produced invalid path: %+v", style, path)
		}
	}
}

func TestMarkerStyleFillNoneUsesFaceAsStrokeFallback(t *testing.T) {
	scatter := &Scatter2D{
		XY: []geom.Pt{{X: 1, Y: 1}},
		MarkerStyle: MarkerStyle{
			Type:      MarkerCircle,
			FillStyle: MarkerFillNone,
		},
		Color: render.Color{R: 0.25, G: 0.5, B: 0.75, A: 1},
		Size:  36,
	}
	pc := scatter.toPathCollection(&render.NullRenderer{}, createTestDrawContext())
	if pc.FaceColor.A != 0 {
		t.Fatalf("fillstyle none face alpha = %v, want 0", pc.FaceColor.A)
	}
	if pc.EdgeColor.A == 0 {
		t.Fatal("fillstyle none should fall back to face color for the outline")
	}
}

func TestScatter2D_ZOrder(t *testing.T) {
	scatter := &Scatter2D{z: 3.5}

	if got := scatter.Z(); got != 3.5 {
		t.Errorf("Expected Z() = 3.5, got %v", got)
	}
}

func TestScatter2D_Bounds(t *testing.T) {
	// Test empty scatter
	scatter := &Scatter2D{}
	bounds := scatter.Bounds(nil)
	expected := geom.Rect{}
	if bounds != expected {
		t.Errorf("Expected Bounds() = %v, got %v", expected, bounds)
	}

	// Test scatter with points
	scatter = &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 2, Y: 1},
			{X: 1, Y: 2},
		},
		Size: 5.0,
	}
	bounds = scatter.Bounds(nil)

	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 2 || bounds.Max.Y != 2 {
		t.Errorf("Expected point-center bounds [0,0]-[2,2], got %v", bounds)
	}

	// Test with variable sizes
	scatter = &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		Sizes: []float64{3.0, 10.0}, // Max size is 10.0
		Size:  5.0,                  // fallback size
	}
	bounds = scatter.Bounds(nil)

	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 1 || bounds.Max.Y != 1 {
		t.Errorf("Expected variable marker sizes not to affect data bounds, got %v", bounds)
	}
}

func TestScatterAutoScaleIgnoresMarkerSize(t *testing.T) {
	fig := NewFigure(800, 600)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	size := 500.0
	ax.Scatter([]float64{0, 2}, []float64{1, 3}, ScatterOptions{Size: &size})

	ax.AutoScale(0.05)
	xMin, xMax := ax.XScale.Domain()
	yMin, yMax := ax.YScale.Domain()

	if !floatApprox(xMin, -0.1, 1e-12) || !floatApprox(xMax, 2.1, 1e-12) {
		t.Fatalf("x limits = [%v, %v], want [-0.1, 2.1]", xMin, xMax)
	}
	if !floatApprox(yMin, 0.9, 1e-12) || !floatApprox(yMax, 3.1, 1e-12) {
		t.Fatalf("y limits = [%v, %v], want [0.9, 3.1]", yMin, yMax)
	}
}

func TestScatter2D_EdgeColors(t *testing.T) {
	// Test with edge colors and width
	scatter := &Scatter2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
		},
		EdgeColors: []render.Color{
			{R: 1, G: 0, B: 0, A: 1},
			{R: 0, G: 1, B: 0, A: 1},
		},
		EdgeColor: render.Color{R: 0, G: 0, B: 1, A: 1}, // fallback
		EdgeWidth: 2.0,
		Size:      5.0,
		Color:     render.Color{R: 0.5, G: 0.5, B: 0.5, A: 1},
		Marker:    MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with edge colors
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_AlphaTransparency(t *testing.T) {
	// Test with alpha transparency
	scatter := &Scatter2D{
		XY:     []geom.Pt{{X: 0, Y: 0}},
		Size:   5.0,
		Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
		Alpha:  0.5, // 50% transparent
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should not panic with alpha transparency
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}
}

func TestScatter2D_AlphaEdgeCases(t *testing.T) {
	testCases := []struct {
		name  string
		alpha float64
	}{
		{"Zero alpha", 0.0},
		{"Negative alpha", -0.5},
		{"Greater than 1 alpha", 1.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scatter := &Scatter2D{
				XY:     []geom.Pt{{X: 0, Y: 0}},
				Size:   5.0,
				Color:  render.Color{R: 1, G: 0, B: 0, A: 1},
				Alpha:  tc.alpha,
				Marker: MarkerCircle,
			}

			renderer := &render.NullRenderer{}
			ctx := createTestDrawContext()

			err := renderer.Begin(geom.Rect{})
			if err != nil {
				t.Fatalf("Failed to begin rendering: %v", err)
			}

			// Should not panic with edge case alpha values
			scatter.Draw(renderer, ctx)

			err = renderer.End()
			if err != nil {
				t.Fatalf("Failed to end rendering: %v", err)
			}
		})
	}
}

func TestScatter2D_LargeDataset(t *testing.T) {
	// Test performance with many points
	const numPoints = 10000
	points := make([]geom.Pt, numPoints)
	sizes := make([]float64, numPoints)
	colors := make([]render.Color, numPoints)

	for i := 0; i < numPoints; i++ {
		points[i] = geom.Pt{X: float64(i), Y: float64(i % 100)}
		sizes[i] = float64(3 + (i % 10))
		colors[i] = render.Color{
			R: float64(i%256) / 255.0,
			G: float64((i*2)%256) / 255.0,
			B: float64((i*3)%256) / 255.0,
			A: 1.0,
		}
	}

	scatter := &Scatter2D{
		XY:     points,
		Sizes:  sizes,
		Colors: colors,
		Marker: MarkerCircle,
	}

	renderer := &render.NullRenderer{}
	ctx := createTestDrawContext()

	err := renderer.Begin(geom.Rect{})
	if err != nil {
		t.Fatalf("Failed to begin rendering: %v", err)
	}

	// Should handle large dataset without issues
	scatter.Draw(renderer, ctx)

	err = renderer.End()
	if err != nil {
		t.Fatalf("Failed to end rendering: %v", err)
	}

	// Test bounds calculation with large dataset
	bounds := scatter.Bounds(nil)
	if bounds == (geom.Rect{}) {
		t.Error("Expected non-empty bounds for large dataset")
	}
}
