package test

import (
	"image"
	"image/color"
	"math"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// === from alpha_residual_diagnostics_test.go ==========================

type alphaResidualCase struct {
	name             string
	axes             image.Rectangle
	threshold        uint8
	maxHighDiffRatio float64
}

func TestAlphaResidualDiagnostics(t *testing.T) {
	cases := []alphaResidualCase{
		{
			name:             "fill_stacked",
			axes:             fixtureAxesRect(640, 360, 0.1, 0.1, 0.9, 0.9),
			threshold:        32,
			maxHighDiffRatio: 0.030,
		},
		{
			name:             "hist_strategies",
			axes:             fixtureAxesRect(640, 360, 0.12, 0.12, 0.95, 0.90),
			threshold:        32,
			maxHighDiffRatio: 0.015,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := parity.Render(tc.name)
			if err != nil {
				t.Fatalf("render parity example %s: %v", tc.name, err)
			}
			want, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "matplotlib_ref", tc.name+".png"))
			if err != nil {
				t.Fatalf("load matplotlib reference: %v", err)
			}
			diag := alphaResidualDiagnostics(got, want, tc.axes, tc.threshold)
			t.Logf(
				"axes=%v threshold=%d highDiff=%d/%d (%.4f) meanAbs=%.3f rmse=%.3f bbox=%v",
				tc.axes,
				tc.threshold,
				diag.highDiff,
				diag.total,
				diag.highDiffRatio(),
				diag.meanAbs(),
				diag.rmse(),
				diag.bbox,
			)
			if diag.highDiffRatio() > tc.maxHighDiffRatio {
				t.Fatalf("high-diff ratio %.4f exceeds %.4f", diag.highDiffRatio(), tc.maxHighDiffRatio)
			}
		})
	}
}

type alphaResidualSummary struct {
	total      int
	highDiff   int
	sumAbs     float64
	sumSquared float64
	bbox       image.Rectangle
	haveBBox   bool
}

func (s alphaResidualSummary) highDiffRatio() float64 {
	if s.total == 0 {
		return 0
	}
	return float64(s.highDiff) / float64(s.total)
}

func (s alphaResidualSummary) meanAbs() float64 {
	if s.total == 0 {
		return 0
	}
	return s.sumAbs / float64(s.total)
}

func (s alphaResidualSummary) rmse() float64 {
	if s.total == 0 {
		return 0
	}
	return math.Sqrt(s.sumSquared / float64(s.total))
}

func alphaResidualDiagnostics(got, want image.Image, rect image.Rectangle, threshold uint8) alphaResidualSummary {
	bounds := got.Bounds().Intersect(want.Bounds()).Intersect(rect)
	summary := alphaResidualSummary{bbox: image.Rectangle{}}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gotColor := color.RGBAModel.Convert(got.At(x, y)).(color.RGBA)
			wantColor := color.RGBAModel.Convert(want.At(x, y)).(color.RGBA)
			diffR := absByteDiff(gotColor.R, wantColor.R)
			diffG := absByteDiff(gotColor.G, wantColor.G)
			diffB := absByteDiff(gotColor.B, wantColor.B)
			diffA := absByteDiff(gotColor.A, wantColor.A)
			maxDiff := maxByte4(diffR, diffG, diffB, diffA)
			mean := float64(diffR+diffG+diffB+diffA) / 4.0
			summary.total++
			summary.sumAbs += mean
			summary.sumSquared += (float64(diffR)*float64(diffR) + float64(diffG)*float64(diffG) + float64(diffB)*float64(diffB) + float64(diffA)*float64(diffA)) / 4.0
			if maxDiff <= threshold {
				continue
			}
			summary.highDiff++
			pt := image.Pt(x, y)
			if !summary.haveBBox {
				summary.bbox = image.Rectangle{Min: pt, Max: pt.Add(image.Pt(1, 1))}
				summary.haveBBox = true
				continue
			}
			summary.bbox = summary.bbox.Union(image.Rectangle{Min: pt, Max: pt.Add(image.Pt(1, 1))})
		}
	}
	return summary
}

func fixtureAxesRect(width, height int, minX, minY, maxX, maxY float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(float64(width)*minX)),
		int(math.Floor(float64(height)*(1-maxY))),
		int(math.Ceil(float64(width)*maxX)),
		int(math.Ceil(float64(height)*(1-minY))),
	)
}

func absByteDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxByte4(a, b, c, d uint8) uint8 {
	out := a
	if b > out {
		out = b
	}
	if c > out {
		out = c
	}
	if d > out {
		out = d
	}
	return out
}
