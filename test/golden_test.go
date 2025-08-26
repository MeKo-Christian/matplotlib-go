package test

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"matplotlib-go/backends/gobasic"
	"matplotlib-go/core"
	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
	"matplotlib-go/test/imagecmp"
	"matplotlib-go/transform"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden images instead of comparing")

func TestBasicLine_Golden(t *testing.T) {
	runGoldenTest(t, "basic_line", renderBasicLine)
}

func TestJoinsCaps_Golden(t *testing.T) {
	runGoldenTest(t, "joins_caps", renderJoinsCaps)
}

func TestDashes_Golden(t *testing.T) {
	runGoldenTest(t, "dashes", renderDashes)
}

func TestScatterBasic_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_basic", renderScatterBasic)
}

func TestScatterMarkerTypes_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_marker_types", renderScatterMarkerTypes)
}

func TestScatterAdvanced_Golden(t *testing.T) {
	runGoldenTest(t, "scatter_advanced", renderScatterAdvanced)
}

func TestBarBasic_Golden(t *testing.T) {
	runGoldenTest(t, "bar_basic", renderBarBasic)
}

func TestBarHorizontal_Golden(t *testing.T) {
	runGoldenTest(t, "bar_horizontal", renderBarHorizontal)
}

func TestBarGrouped_Golden(t *testing.T) {
	runGoldenTest(t, "bar_grouped", renderBarGrouped)
}

func TestFillBasic_Golden(t *testing.T) {
	runGoldenTest(t, "fill_basic", renderFillBasic)
}

func TestFillBetween_Golden(t *testing.T) {
	runGoldenTest(t, "fill_between", renderFillBetween)
}

func TestFillStacked_Golden(t *testing.T) {
	runGoldenTest(t, "fill_stacked", renderFillStacked)
}

func TestMultiSeriesBasic_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_basic", renderMultiSeriesBasic)
}

func TestMultiSeriesColorCycle_Golden(t *testing.T) {
	runGoldenTest(t, "multi_series_color_cycle", renderMultiSeriesColorCycle)
}

// runGoldenTest is a helper function for golden image testing
func runGoldenTest(t *testing.T, testName string, renderFunc func() *gobasic.Renderer) {
	// Render the plot
	r := renderFunc()
	img := r.GetImage()

	goldenPath := "../testdata/golden/" + testName + ".png"

	if *updateGolden {
		// Update the golden image
		err := imagecmp.SavePNG(img, goldenPath)
		if err != nil {
			t.Fatalf("Failed to update golden image: %v", err)
		}
		t.Skip("Updated golden image")
		return
	}

	// Load the expected golden image
	want, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("Failed to load golden image %s: %v", goldenPath, err)
	}

	// Compare with tolerance
	diff, err := imagecmp.ComparePNG(img, want, 1) // ≤1 LSB tolerance
	if err != nil {
		t.Fatalf("Image comparison failed: %v", err)
	}

	// Check if images are within tolerance
	if !diff.Identical {
		// Save debug images
		artifactsDir := "../_artifacts"
		if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
			t.Logf("Warning: could not create artifacts directory: %v", err)
		} else {
			// Save the rendered image
			gotPath := filepath.Join(artifactsDir, testName+"_got.png")
			if err := imagecmp.SavePNG(img, gotPath); err != nil {
				t.Logf("Warning: could not save got image: %v", err)
			}

			// Save the diff image
			diffPath := filepath.Join(artifactsDir, testName+"_diff.png")
			if err := imagecmp.SaveDiffImage(img, want, 1, diffPath); err != nil {
				t.Logf("Warning: could not save diff image: %v", err)
			}

			t.Logf("Debug images saved to %s/", artifactsDir)
		}

		t.Fatalf("Golden image mismatch: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
			diff.MaxDiff, diff.MeanAbs, diff.PSNR)
	}

	t.Logf("Golden image match: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB",
		diff.MaxDiff, diff.MeanAbs, diff.PSNR)
}

// renderBasicLine creates the same basic line plot as examples/lines/basic.go
func renderBasicLine() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.15},
		Max: geom.Pt{X: 0.95, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 1)

	// Create a line with some sample data
	line := &core.Line2D{
		XY: []geom.Pt{
			{X: 0, Y: 0},
			{X: 1, Y: 0.2},
			{X: 3, Y: 0.9},
			{X: 6, Y: 0.4},
			{X: 10, Y: 0.8},
		},
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 0, A: 1}, // black line
	}

	// Add the line to the axes
	ax.Add(line)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderJoinsCaps creates a plot demonstrating different line joins and caps
func renderJoinsCaps() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 6)

	// L-shaped path to demonstrate joins
	joinPath := []geom.Pt{
		{X: 1, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 3}, {X: 5, Y: 3},
	}

	// Miter join line (thick red)
	miterLine := &core.Line2D{
		XY:  joinPath,
		W:   8.0,
		Col: render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1},
	}
	ax.Add(miterLine)

	// Straight line for caps demo
	capPath := []geom.Pt{
		{X: 7, Y: 5}, {X: 9, Y: 5},
	}

	// Thick blue line with round caps
	capLine := &core.Line2D{
		XY:  capPath,
		W:   8.0,
		Col: render.Color{R: 0.2, G: 0.2, B: 0.8, A: 1},
	}
	ax.Add(capLine)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderDashes creates a plot demonstrating dash patterns
func renderDashes() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 5)

	// Multiple horizontal lines with different dash patterns
	lines := []struct {
		y      float64
		dashes []float64
		color  render.Color
	}{
		{4, []float64{}, render.Color{R: 0, G: 0, B: 0, A: 1}},             // solid
		{3, []float64{5, 2}, render.Color{R: 0.8, G: 0, B: 0, A: 1}},       // basic dash
		{2, []float64{3, 1, 1, 1}, render.Color{R: 0, G: 0.6, B: 0, A: 1}}, // dash-dot
		{1, []float64{1, 1}, render.Color{R: 0, G: 0, B: 0.8, A: 1}},       // dotted
	}

	for _, lineSpec := range lines {
		// Create line data
		path := []geom.Pt{
			{X: 1, Y: lineSpec.y}, {X: 9, Y: lineSpec.y},
		}

		line := &core.Line2D{
			XY:     path,
			W:      3.0,
			Col:    lineSpec.color,
			Dashes: lineSpec.dashes,
		}
		ax.Add(line)
	}

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderScatterBasic creates a basic scatter plot for golden testing
func renderScatterBasic() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 10)

	// Basic scatter with circles
	basicPoints := []geom.Pt{
		{X: 2, Y: 3}, {X: 4, Y: 6}, {X: 6, Y: 4},
		{X: 8, Y: 7}, {X: 3, Y: 8}, {X: 7, Y: 2},
	}

	scatter := &core.Scatter2D{
		XY:     basicPoints,
		Size:   8.0,
		Color:  render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
		Marker: core.MarkerCircle,
		Alpha:  1.0,
	}
	ax.Add(scatter)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderScatterMarkerTypes creates a plot showing all marker types
func renderScatterMarkerTypes() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 8)

	// All marker types with different colors
	markerTypes := []core.MarkerType{
		core.MarkerCircle, core.MarkerSquare, core.MarkerTriangle,
		core.MarkerDiamond, core.MarkerPlus, core.MarkerCross,
	}

	colors := []render.Color{
		{R: 1, G: 0, B: 0, A: 1}, // red
		{R: 0, G: 1, B: 0, A: 1}, // green  
		{R: 0, G: 0, B: 1, A: 1}, // blue
		{R: 1, G: 1, B: 0, A: 1}, // yellow
		{R: 1, G: 0, B: 1, A: 1}, // magenta
		{R: 0, G: 1, B: 1, A: 1}, // cyan
	}

	for i, markerType := range markerTypes {
		x := float64(1 + i)
		y := float64(4)

		scatter := &core.Scatter2D{
			XY:     []geom.Pt{{X: x, Y: y}},
			Size:   12.0,
			Color:  colors[i],
			Marker: markerType,
			Alpha:  1.0,
		}
		ax.Add(scatter)
	}

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderScatterAdvanced creates an advanced scatter plot with edges, alpha, and variable sizes
func renderScatterAdvanced() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 10)

	// Variable sizes and colors with edge support
	points := []geom.Pt{
		{X: 2, Y: 2}, {X: 4, Y: 4}, {X: 6, Y: 6}, {X: 8, Y: 8},
		{X: 2, Y: 8}, {X: 4, Y: 6}, {X: 6, Y: 4}, {X: 8, Y: 2},
	}

	sizes := []float64{6, 10, 14, 18, 8, 12, 16, 20}

	fillColors := []render.Color{
		{R: 1, G: 0.5, B: 0.5, A: 1}, {R: 0.5, G: 1, B: 0.5, A: 1},
		{R: 0.5, G: 0.5, B: 1, A: 1}, {R: 1, G: 1, B: 0.5, A: 1},
		{R: 1, G: 0.5, B: 1, A: 1}, {R: 0.5, G: 1, B: 1, A: 1},
		{R: 0.8, G: 0.8, B: 0.8, A: 1}, {R: 0.3, G: 0.3, B: 0.3, A: 1},
	}

	edgeColors := []render.Color{
		{R: 0.5, G: 0, B: 0, A: 1}, {R: 0, G: 0.5, B: 0, A: 1},
		{R: 0, G: 0, B: 0.5, A: 1}, {R: 0.5, G: 0.5, B: 0, A: 1},
		{R: 0.5, G: 0, B: 0.5, A: 1}, {R: 0, G: 0.5, B: 0.5, A: 1},
		{R: 0.4, G: 0.4, B: 0.4, A: 1}, {R: 0, G: 0, B: 0, A: 1},
	}

	scatter := &core.Scatter2D{
		XY:         points,
		Sizes:      sizes,
		Colors:     fillColors,
		EdgeColors: edgeColors,
		EdgeWidth:  2.0,
		Alpha:      0.8,
		Marker:     core.MarkerCircle,
	}
	ax.Add(scatter)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderBarBasic creates a basic vertical bar chart for golden testing
func renderBarBasic() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 6)
	ax.YScale = transform.NewLinear(0, 10)

	// Basic vertical bar chart
	bar := &core.Bar2D{
		X:           []float64{1, 2, 3, 4, 5},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.6,
		Color:       render.Color{R: 0.2, G: 0.6, B: 0.8, A: 1}, // blue
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderBarHorizontal creates a horizontal bar chart for golden testing
func renderBarHorizontal() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(0, 6)

	// Horizontal bar chart
	bar := &core.Bar2D{
		X:           []float64{1, 2, 3, 4, 5},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.6,
		Color:       render.Color{R: 0.8, G: 0.4, B: 0.2, A: 1}, // orange
		Baseline:    0,
		Orientation: core.BarHorizontal,
	}
	ax.Add(bar)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderBarGrouped creates a grouped bar chart with variable colors and edges
func renderBarGrouped() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 7)
	ax.YScale = transform.NewLinear(0, 10)

	// First series - shifted left
	bar1 := &core.Bar2D{
		X:           []float64{1.2, 2.2, 3.2, 4.2, 5.2},
		Heights:     []float64{3, 7, 2, 8, 5},
		Width:       0.35,
		Color:       render.Color{R: 0.8, G: 0.2, B: 0.2, A: 1}, // red
		EdgeColor:   render.Color{R: 0.5, G: 0, B: 0, A: 1},     // dark red edge
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar1)

	// Second series - shifted right
	bar2 := &core.Bar2D{
		X:           []float64{1.8, 2.8, 3.8, 4.8, 5.8},
		Heights:     []float64{5, 4, 6, 3, 7},
		Width:       0.35,
		Color:       render.Color{R: 0.2, G: 0.8, B: 0.2, A: 1}, // green
		EdgeColor:   render.Color{R: 0, G: 0.5, B: 0, A: 1},     // dark green edge
		EdgeWidth:   1.0,
		Baseline:    0,
		Orientation: core.BarVertical,
	}
	ax.Add(bar2)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderFillBasic creates a basic fill to baseline for golden testing
func renderFillBasic() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 10)
	ax.YScale = transform.NewLinear(-1, 3)

	// Create simple curve data
	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{0.5, 1.8, 2.3, 1.2, 2.8, 1.9, 2.1, 1.5, 0.8}

	// Fill to baseline with edge
	fill := &core.Fill2D{
		X:         x,
		Y1:        y,
		Baseline:  0,
		Color:     render.Color{R: 0.3, G: 0.7, B: 0.9, A: 0.7}, // semi-transparent blue
		EdgeColor: render.Color{R: 0.1, G: 0.3, B: 0.5, A: 1.0}, // darker blue edge
		EdgeWidth: 2.0,
		Alpha:     1.0,
	}
	ax.Add(fill)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderFillBetween creates a fill between two curves for golden testing
func renderFillBetween() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 6.28)
	ax.YScale = transform.NewLinear(-1.5, 1.5)

	// Generate sine and cosine curves
	n := 50
	x := make([]float64, n)
	y1 := make([]float64, n) // sine
	y2 := make([]float64, n) // cosine * 0.8

	for i := 0; i < n; i++ {
		t := 6.28 * float64(i) / float64(n-1)
		x[i] = t
		y1[i] = math.Sin(t)
		y2[i] = 0.8 * math.Cos(t)
	}

	// Fill between curves
	fill := core.FillBetween(x, y1, y2, render.Color{R: 0.8, G: 0.3, B: 0.3, A: 0.6})
	fill.EdgeColor = render.Color{R: 0.5, G: 0.1, B: 0.1, A: 1.0}
	fill.EdgeWidth = 1.5

	ax.Add(fill)

	// Add the curves themselves as lines
	sineLine := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 1, G: 0, B: 0, A: 1}, // red
	}
	cosLine := &core.Line2D{
		XY:  make([]geom.Pt, n),
		W:   2.0,
		Col: render.Color{R: 0, G: 0, B: 1, A: 1}, // blue
	}

	for i := 0; i < n; i++ {
		sineLine.XY[i] = geom.Pt{X: x[i], Y: y1[i]}
		cosLine.XY[i] = geom.Pt{X: x[i], Y: y2[i]}
	}

	ax.Add(sineLine)
	ax.Add(cosLine)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderFillStacked creates a stacked area chart for golden testing
func renderFillStacked() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 8)

	// Create stacked data
	x := []float64{1, 2, 3, 4, 5, 6, 7}
	layer1 := []float64{1, 1.5, 2, 1.8, 2.2, 1.9, 1.6}
	layer2 := make([]float64, len(layer1))
	layer3 := make([]float64, len(layer1))

	// Stack the layers
	for i := range layer1 {
		layer2[i] = layer1[i] + 1.5 + 0.3*math.Sin(float64(i))
		layer3[i] = layer2[i] + 1.2 + 0.4*math.Cos(float64(i))
	}

	// Bottom layer (to baseline)
	fill1 := core.FillToBaseline(x, layer1, 0, render.Color{R: 0.8, G: 0.2, B: 0.2, A: 0.8})
	fill1.EdgeColor = render.Color{R: 0.5, G: 0, B: 0, A: 1}
	fill1.EdgeWidth = 1.0

	// Middle layer (between layer1 and layer2)
	fill2 := core.FillBetween(x, layer1, layer2, render.Color{R: 0.2, G: 0.8, B: 0.2, A: 0.8})
	fill2.EdgeColor = render.Color{R: 0, G: 0.5, B: 0, A: 1}
	fill2.EdgeWidth = 1.0

	// Top layer (between layer2 and layer3)
	fill3 := core.FillBetween(x, layer2, layer3, render.Color{R: 0.2, G: 0.2, B: 0.8, A: 0.8})
	fill3.EdgeColor = render.Color{R: 0, G: 0, B: 0.5, A: 1}
	fill3.EdgeWidth = 1.0

	ax.Add(fill1)
	ax.Add(fill2)
	ax.Add(fill3)

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderMultiSeriesBasic creates a plot with multiple series using different plot types
func renderMultiSeriesBasic() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 8)
	ax.YScale = transform.NewLinear(0, 6)

	// Generate sample data
	x1 := []float64{1, 2, 3, 4, 5, 6}
	y1 := []float64{1.5, 2.8, 2.2, 3.5, 3.8, 4.2}

	x2 := []float64{1.5, 2.5, 3.5, 4.5, 5.5}
	y2 := []float64{2.2, 3.1, 2.9, 4.1, 4.5}

	x3 := []float64{2, 3, 4, 5}
	y3 := []float64{3.8, 2.5, 4.8, 3.2}

	// Use convenience methods with automatic color cycling
	ax.Plot(x1, y1, core.PlotOptions{Label: "Series 1"})
	ax.Scatter(x2, y2, core.ScatterOptions{Label: "Series 2"})
	
	width := 0.4
	ax.Bar(x3, y3, core.BarOptions{Label: "Series 3", Width: &width})

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}

// renderMultiSeriesColorCycle creates a plot demonstrating automatic color cycling
func renderMultiSeriesColorCycle() *gobasic.Renderer {
	// Create a figure with dimensions 640x360
	fig := core.NewFigure(640, 360)

	// Add axes that take up most of the figure space
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})

	// Set up coordinate scales
	ax.XScale = transform.NewLinear(0, 2*math.Pi)
	ax.YScale = transform.NewLinear(-1.2, 1.2)

	// Generate sine waves with different frequencies
	nPoints := 50
	x := make([]float64, nPoints)
	for i := 0; i < nPoints; i++ {
		x[i] = 2 * math.Pi * float64(i) / float64(nPoints-1)
	}

	// Create 4 different sine waves with automatic color cycling
	for freq := 1; freq <= 4; freq++ {
		y := make([]float64, nPoints)
		for i := 0; i < nPoints; i++ {
			y[i] = math.Sin(float64(freq) * x[i])
		}

		label := fmt.Sprintf("f=%d", freq)
		ax.Plot(x, y, core.PlotOptions{Label: label})
	}

	// Create a GoBasic renderer with white background
	r := gobasic.New(640, 360, render.Color{R: 1, G: 1, B: 1, A: 1})

	// Render the figure
	core.DrawFigure(fig, r)

	return r
}
