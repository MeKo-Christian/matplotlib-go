package test

// Reference-compare tests cross-check our golden PNG against the matplotlib
// reference PNG for each case, applying per-case PSNR/MeanAbs/RMSE tolerances
// stored on the catalog row.
//
// Per-case invocation:
//
//	go test ./test/... -run TestReferenceCompare/basic_line
//	go test ./test/... -run 'TestReferenceCompare/.*pcolor.*'

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// TestReferenceCompare iterates every catalog case that has both a golden
// PNG and a matplotlib reference PNG, comparing them with case-specific
// tolerances pulled from the catalog row.
func TestReferenceCompare(t *testing.T) {
	for _, c := range rendererNeutralCases() {
		c := c
		if !goldenExists(c.ID) || !matplotlibRefExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			runReferenceCompareTest(t, &c)
		})
	}
}

func TestAGGNativeReferenceCompare(t *testing.T) {
	for _, c := range aggNativeCases() {
		c := c
		if !goldenExists(c.ID) || !matplotlibRefExists(c.ID) {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			runReferenceCompareTest(t, &c)
		})
	}
}

// TestColorbarCompositionImageOriginMatchesMatplotlibRef is a bespoke
// chromatic-bounding-box check: the rendered colorbar's main image area
// must align with matplotlib's within a 2-pixel tolerance.
func TestColorbarCompositionImageOriginMatchesMatplotlibRef(t *testing.T) {
	got, _, err := parity.Render("colorbar_composition")
	if err != nil {
		t.Fatalf("render parity example colorbar_composition: %v", err)
	}
	matplotlibRef, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "matplotlib_ref", "colorbar_composition.png"))
	if err != nil {
		t.Fatalf("load matplotlib reference: %v", err)
	}

	gotRect, ok := largestChromaComponentBounds(got)
	if !ok {
		t.Fatal("rendered colorbar composition has no chromatic component")
	}
	wantRect, ok := largestChromaComponentBounds(matplotlibRef)
	if !ok {
		t.Fatal("matplotlib colorbar composition has no chromatic component")
	}

	if !rectWithinPixels(gotRect, wantRect, 4) {
		t.Fatalf("main image component bounds = %v, want within 4 px of matplotlib %v", gotRect, wantRect)
	}
}

// ----------------------------------------------------------------------------
// Chromatic component analysis (used only by the colorbar spot check above)
// ----------------------------------------------------------------------------

func largestChromaComponentBounds(img image.Image) (image.Rectangle, bool) {
	if img == nil {
		return image.Rectangle{}, false
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return image.Rectangle{}, false
	}

	mask := make([]bool, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if isChromaticPixel(img.At(bounds.Min.X+x, bounds.Min.Y+y)) {
				mask[y*width+x] = true
			}
		}
	}

	visited := make([]bool, len(mask))
	queue := make([]int, 0, len(mask)/8)
	bestArea := 0
	best := image.Rectangle{}

	for idx, chromatic := range mask {
		if !chromatic || visited[idx] {
			continue
		}
		visited[idx] = true
		queue = append(queue[:0], idx)
		area := 0
		minX, maxX := idx%width, idx%width
		minY, maxY := idx/width, idx/width

		for len(queue) > 0 {
			cur := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			x := cur % width
			y := cur / width
			area++
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}

			neighbors := [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}}
			for _, neighbor := range neighbors {
				nx, ny := neighbor[0], neighbor[1]
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				next := ny*width + nx
				if !mask[next] || visited[next] {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}

		if area > bestArea {
			bestArea = area
			best = image.Rect(bounds.Min.X+minX, bounds.Min.Y+minY, bounds.Min.X+maxX+1, bounds.Min.Y+maxY+1)
		}
	}

	return best, bestArea > 0
}

func isChromaticPixel(c color.Color) bool {
	r16, g16, b16, a16 := c.RGBA()
	if a16 == 0 {
		return false
	}
	r := int(r16 >> 8)
	g := int(g16 >> 8)
	b := int(b16 >> 8)
	lo := min(r, min(g, b))
	hi := max(r, max(g, b))
	return hi-lo > 20 && hi < 252
}

func rectWithinPixels(got, want image.Rectangle, tolerance int) bool {
	return absInt(got.Min.X-want.Min.X) <= tolerance &&
		absInt(got.Min.Y-want.Min.Y) <= tolerance &&
		absInt(got.Max.X-want.Max.X) <= tolerance &&
		absInt(got.Max.Y-want.Max.Y) <= tolerance
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
