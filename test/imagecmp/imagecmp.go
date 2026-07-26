// Package imagecmp provides utilities for comparing images in tests,
// particularly for golden image testing.
package imagecmp

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// DiffResult contains metrics from comparing two images.
//
// PSNR is derived from RMSE as 20*log10(255/RMSE) and therefore carries no
// information RMSE does not: it is a monotone restatement, useful for reading
// logs, not as a second independent gate. Until Phase 3.1 the two were computed
// from separate accumulators and looked independent, but only because the PSNR
// accumulator squared its differences in uint8 arithmetic and wrapped mod 256
// for every difference above 15.
//
// DiffPixels, Clusters and LargestCluster describe the *shape* of the residual
// rather than its amplitude, which is what MeanAbs, RMSE and PSNR — all
// whole-image averages — structurally cannot do. A wholly misplaced glyph is a
// few hundred pixels at near-maximum amplitude; averaged over a 640x360 frame
// it disappears into a small RMSE, but it is unmistakable as one dense cluster.
// See docs/plans/phase3-tolerance-audit.md.
type DiffResult struct {
	MaxDiff   uint8   // Maximum per-channel difference found
	MeanAbs   float64 // Mean absolute difference across all channels
	RMSE      float64 // Root-mean-square error across all channels
	PSNR      float64 // Peak Signal-to-Noise Ratio in dB, = 20*log10(255/RMSE)
	Identical bool    // True if images are pixel-perfect identical

	// DiffPixels counts pixels whose largest per-channel difference exceeds
	// tolerance — the same population SaveDiffImage highlights.
	DiffPixels int
	// Clusters counts the 8-connected components those pixels form.
	Clusters int
	// LargestCluster is the pixel count of the biggest such component.
	LargestCluster int
}

// ComparePNG compares two images and returns difference metrics.
// The tolerance parameter specifies the maximum allowed per-channel difference
// before considering pixels different (typically 1 for ≤1 LSB tolerance).
func ComparePNG(got, want image.Image, tolerance uint8) (DiffResult, error) {
	gotBounds := got.Bounds()
	wantBounds := want.Bounds()

	// Check dimensions match
	if gotBounds.Size() != wantBounds.Size() {
		return DiffResult{}, fmt.Errorf("image dimensions differ: got %v, want %v",
			gotBounds.Size(), wantBounds.Size())
	}

	var maxDiff uint8
	var sumDiff float64
	var numPixels int64
	var sumSquaredError float64
	identical := true

	width := gotBounds.Dx()
	height := gotBounds.Dy()
	// Mask of pixels exceeding tolerance, used below to measure residual shape.
	mask := make([]bool, width*height)
	var diffPixels int

	// Iterate through all pixels
	for y := gotBounds.Min.Y; y < gotBounds.Max.Y; y++ {
		for x := gotBounds.Min.X; x < gotBounds.Max.X; x++ {
			gotColor := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
			wantColor := color.RGBAModel.Convert(want.At(x, y)).(color.RGBA)

			// Calculate per-channel differences
			diffR := absDiff(gotColor.R, wantColor.R)
			diffG := absDiff(gotColor.G, wantColor.G)
			diffB := absDiff(gotColor.B, wantColor.B)
			diffA := absDiff(gotColor.A, wantColor.A)

			// Track maximum difference
			channelMax := max4(diffR, diffG, diffB, diffA)
			if channelMax > maxDiff {
				maxDiff = channelMax
			}

			// Calculate mean absolute difference
			channelSum := float64(diffR + diffG + diffB + diffA)
			sumDiff += channelSum / 4.0 // Average per pixel

			// Accumulate squared error in float64: squaring a uint8 difference
			// in uint8 arithmetic overflows for anything above 15.
			squaredError := squareDiff(diffR) + squareDiff(diffG) + squareDiff(diffB) + squareDiff(diffA)
			sumSquaredError += squaredError / 4.0 // Average per pixel

			numPixels++

			// Check if pixel exceeds tolerance
			if channelMax > tolerance {
				identical = false
				mask[(y-gotBounds.Min.Y)*width+(x-gotBounds.Min.X)] = true
				diffPixels++
			}
		}
	}

	clusters, largestCluster := clusterStats(mask, width, height)

	// Calculate metrics
	meanAbs := sumDiff / float64(numPixels)

	// Calculate PSNR (Peak Signal-to-Noise Ratio)
	var psnr float64
	var rmse float64
	if sumSquaredError == 0 {
		psnr = math.Inf(1) // Perfect match
	} else {
		mse := sumSquaredError / float64(numPixels)
		rmse = math.Sqrt(mse)
		psnr = 20 * math.Log10(255/rmse)
	}

	return DiffResult{
		MaxDiff:        maxDiff,
		MeanAbs:        meanAbs,
		RMSE:           rmse,
		PSNR:           psnr,
		Identical:      identical && maxDiff <= tolerance,
		DiffPixels:     diffPixels,
		Clusters:       clusters,
		LargestCluster: largestCluster,
	}, nil
}

// clusterStats labels the 8-connected components of mask and returns their
// count along with the size of the largest one. Connectivity matches the
// 3x3 structuring element the audit generator uses
// (docs/plans/generate_phase3_tolerance_audit.py), so the two agree per case.
func clusterStats(mask []bool, width, height int) (clusters, largest int) {
	if width <= 0 || height <= 0 {
		return 0, 0
	}

	visited := make([]bool, len(mask))
	stack := make([]int, 0, 64)

	for start := range mask {
		if !mask[start] || visited[start] {
			continue
		}
		clusters++
		visited[start] = true
		stack = append(stack[:0], start)
		size := 0

		for len(stack) > 0 {
			cur := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			size++
			x := cur % width
			y := cur / width

			for dy := -1; dy <= 1; dy++ {
				ny := y + dy
				if ny < 0 || ny >= height {
					continue
				}
				for dx := -1; dx <= 1; dx++ {
					nx := x + dx
					if nx < 0 || nx >= width || (dx == 0 && dy == 0) {
						continue
					}
					next := ny*width + nx
					if !mask[next] || visited[next] {
						continue
					}
					visited[next] = true
					stack = append(stack, next)
				}
			}
		}

		if size > largest {
			largest = size
		}
	}

	return clusters, largest
}

// LoadPNG loads a PNG image from the given file path.
func LoadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PNG file %s: %w", path, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG file %s: %w", path, err)
	}

	return img, nil
}

// SavePNG saves an image to the given file path as PNG.
func SavePNG(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create PNG file %s: %w", path, err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG file %s: %w", path, err)
	}

	return nil
}

// HashPNG computes a SHA256 hash of the image's raw RGBA data.
// This provides a deterministic fingerprint for CI assertions.
func HashPNG(img image.Image) string {
	bounds := img.Bounds()
	hasher := sha256.New()

	// Convert to RGBA and hash raw bytes
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			hasher.Write([]byte{rgba.R, rgba.G, rgba.B, rgba.A})
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// SaveDiffImage creates a visual diff image highlighting differences between two images.
// Pixels that differ by more than threshold are highlighted in red.
func SaveDiffImage(got, want image.Image, threshold uint8, outputPath string) error {
	bounds := got.Bounds()
	diffImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gotColor := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
			wantColor := color.RGBAModel.Convert(want.At(x, y)).(color.RGBA)

			// Calculate maximum channel difference
			diffR := absDiff(gotColor.R, wantColor.R)
			diffG := absDiff(gotColor.G, wantColor.G)
			diffB := absDiff(gotColor.B, wantColor.B)
			diffA := absDiff(gotColor.A, wantColor.A)

			maxDiff := max4(diffR, diffG, diffB, diffA)

			if maxDiff > threshold {
				// Highlight differences in bright red
				diffImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			} else {
				// Show original pixel (from 'got') for context
				diffImg.Set(x, y, gotColor)
			}
		}
	}

	return SavePNG(diffImg, outputPath)
}

// Helper functions

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func squareDiff(v uint8) float64 {
	f := float64(v)
	return f * f
}

func max4(a, b, c, d uint8) uint8 {
	maxVal := a
	if b > maxVal {
		maxVal = b
	}
	if c > maxVal {
		maxVal = c
	}
	if d > maxVal {
		maxVal = d
	}
	return maxVal
}
