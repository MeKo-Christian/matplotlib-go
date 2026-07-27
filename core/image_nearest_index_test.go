package core

import (
	"math"
	"testing"
)

// aggNearestReference is an independent transcription of AGG's nearest-neighbour
// dest->src map, written from the C++ rather than from nearestSourceIndex, so
// the assertions below are not just the implementation restated.
func aggNearestReference(dst, src, index int) int {
	u := (float64(index) + 0.5) * float64(src) / float64(dst)
	fixed := math.Trunc(u*256 + 0.5)
	return int(math.Floor(fixed / 256))
}

// Four boundaries taken from real parity cases. Each is a destination
// coordinate where the source coordinate lands inside the 1/512 gap below a
// cell boundary, so AGG takes the cell on the right and a plain floor takes the
// cell on the left. Each one is a whole mis-drawn column or row in a parity
// case.
func TestScaledNearestIndexTakesTheAggSideOfACellBoundary(t *testing.T) {
	cases := []struct {
		name        string
		target      int
		source      int
		index       int
		want        int
		wantByFloor int
	}{
		{"colormap_diverging x", 500, 9, 55, 1, 0},
		{"image_heatmap x", 544, 3, 362, 2, 1},
		{"specgram_psd x", 496, 21, 401, 17, 16},
		{"specgram_psd y", 278, 65, 239, 56, 55},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scaledNearestIndex(tc.index, tc.target, tc.source); got != tc.want {
				t.Fatalf("scaledNearestIndex(%d, %d, %d) = %d, want %d",
					tc.index, tc.target, tc.source, got, tc.want)
			}
			floored := int(math.Floor((float64(tc.index) + 0.5) * float64(tc.source) / float64(tc.target)))
			if floored != tc.wantByFloor {
				t.Fatalf("plain floor gave %d, want %d — the test case no longer straddles a boundary",
					floored, tc.wantByFloor)
			}
			// The neighbours must be unaffected: this is a single-column flip,
			// not a shift of the whole mapping.
			for _, delta := range []int{-1, 1} {
				neighbour := tc.index + delta
				want := aggNearestReference(tc.target, tc.source, neighbour)
				if got := scaledNearestIndex(neighbour, tc.target, tc.source); got != want {
					t.Errorf("scaledNearestIndex(%d, %d, %d) = %d, want %d",
						neighbour, tc.target, tc.source, got, want)
				}
			}
		})
	}
}

func TestScaledNearestIndexMatchesAggAcrossFullScans(t *testing.T) {
	for _, dims := range [][2]int{{500, 9}, {544, 3}, {496, 21}, {278, 65}, {260, 5}, {7, 7}} {
		target, source := dims[0], dims[1]
		for index := range target {
			want := aggNearestReference(target, source, index)
			if want >= source {
				want = source - 1
			}
			if got := scaledNearestIndex(index, target, source); got != want {
				t.Fatalf("scaledNearestIndex(%d, %d, %d) = %d, want %d",
					index, target, source, got, want)
			}
		}
	}
}

// origin="lower" reflects the destination row, not the resulting source index.
// The boundary rule makes the two inequivalent: reflecting the index computes
// (rows-1)-fp(u) where Matplotlib computes fp(rows-u), and those disagree at
// exactly the coordinates that snap up onto a cell edge.
func TestScaledScalarValueReflectsTheDestinationRowForLowerOrigin(t *testing.T) {
	const (
		targetHeight = 278
		targetWidth  = 4
		rows         = 65
		cols         = 2
	)

	data := make([][]float64, rows)
	for r := range data {
		data[r] = []float64{float64(r), float64(r)}
	}
	img := &Image2D{Data: data, Origin: ImageOriginLower}

	for dstY := range targetHeight {
		want := float64(aggNearestReference(targetHeight, rows, targetHeight-1-dstY))

		got, ok := img.scaledScalarValue(dstY, 0, targetHeight, targetWidth, rows, cols)
		if !ok {
			t.Fatalf("scaledScalarValue(%d) not ok", dstY)
		}
		if got != want {
			t.Fatalf("row for dstY=%d: got data row %v, want %v", dstY, got, want)
		}
	}
}
