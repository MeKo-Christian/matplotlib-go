package test

import (
	"flag"
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
