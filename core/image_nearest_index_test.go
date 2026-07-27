package core

import (
	"math"
	"testing"
)

// aggIroundReference and aggSpanReference are independent transcriptions of
// AGG's nearest-neighbour dest->src map, written from the C++ rather than from
// the implementation, so the assertions below are not just the implementation
// restated.
//
// The two axes take different paths. Within one scanline the source row is
// constant, so the interpolator reduces to a single rounding; the source column
// is walked by an integer DDA between the two rounded ends of the span.
func aggIroundReference(dst, src, index int) int {
	u := (float64(index) + 0.5) * float64(src) / float64(dst)
	fixed := math.Trunc(u*256 + 0.5)
	return int(math.Floor(fixed / 256))
}

func aggSpanReference(dst, src int) []int {
	scale := float64(dst) / float64(src)
	iround := func(v float64) int {
		if v < 0 {
			return int(math.Ceil(v - 0.5))
		}
		return int(math.Floor(v + 0.5))
	}
	begin := iround(0.5 / scale * 256)
	end := iround((float64(dst) + 0.5) / scale * 256)

	count := dst
	delta := end - begin
	left := int(math.Trunc(float64(delta) / float64(count)))
	rem := delta % count
	mod := rem
	value := begin
	if mod <= 0 {
		mod += count
		rem += count
		left--
	}
	mod -= count

	out := make([]int, dst)
	for x := range out {
		out[x] = value >> 8
		mod += rem
		value += left
		if mod > 0 {
			mod -= count
			value++
		}
	}
	return out
}

// Boundaries taken from real parity cases, where the port and Matplotlib
// disagreed on which source cell a whole destination column or row belongs to.
// Each row records what the arithmetic per-pixel rules would answer, so the
// case fails loudly if it stops straddling a boundary.
func TestSpanColumnsFollowTheAggDDA(t *testing.T) {
	cases := []struct {
		name         string
		target       int
		source       int
		index        int
		want         int
		wantByFloor  int
		wantByIround int
	}{
		{"colormap_diverging x", 500, 9, 55, 1, 0, 1},
		{"image_heatmap x", 544, 3, 362, 2, 1, 2},
		// Here the DDA sides with plain floor and the per-pixel rounding is the
		// one that is wrong, which is why neither arithmetic rule can stand in
		// for the interpolator.
		{"asinh_norm_image x", 397, 7, 283, 4, 4, 5},
		{"lognorm_imshow x", 397, 6, 330, 4, 4, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			columns := aggSpanColumns(tc.target, tc.source, float64(tc.target)/float64(tc.source), 0)
			if len(columns) != tc.target {
				t.Fatalf("aggSpanColumns returned %d columns, want %d", len(columns), tc.target)
			}
			if got := columns[tc.index]; got != tc.want {
				t.Fatalf("column %d of (%d, %d) = %d, want %d", tc.index, tc.target, tc.source, got, tc.want)
			}

			floored := int(math.Floor((float64(tc.index) + 0.5) * float64(tc.source) / float64(tc.target)))
			if floored != tc.wantByFloor {
				t.Fatalf("plain floor gave %d, want %d — the case no longer straddles a boundary", floored, tc.wantByFloor)
			}
			if rounded := aggIroundReference(tc.target, tc.source, tc.index); rounded != tc.wantByIround {
				t.Fatalf("per-pixel rounding gave %d, want %d — the case no longer straddles a boundary", rounded, tc.wantByIround)
			}
		})
	}
}

func TestSpanColumnsMatchAggAcrossFullScans(t *testing.T) {
	for _, dims := range [][2]int{{500, 9}, {544, 3}, {496, 21}, {397, 7}, {397, 6}, {278, 65}, {260, 5}, {7, 7}} {
		target, source := dims[0], dims[1]
		want := aggSpanReference(target, source)
		got := aggSpanColumns(target, source, float64(target)/float64(source), 0)
		for index := range target {
			expected := want[index]
			if expected >= source {
				expected = source - 1
			}
			if expected < 0 {
				expected = 0
			}
			if got[index] != expected {
				t.Fatalf("column %d of (%d, %d) = %d, want %d", index, target, source, got[index], expected)
			}
		}
	}
}

// The row is constant along a scanline, so it keeps the single rounding rather
// than the DDA. Plain floor is a different rule: a coordinate landing within
// 1/512 of a cell below a boundary snaps up onto it.
func TestScaledNearestIndexRoundsTheRowLikeAgg(t *testing.T) {
	for _, dims := range [][2]int{{500, 9}, {278, 5}, {278, 65}, {260, 5}, {270, 3}} {
		target, source := dims[0], dims[1]
		for index := range target {
			want := aggIroundReference(target, source, index)
			if want >= source {
				want = source - 1
			}
			if got := scaledNearestIndex(index, target, source); got != want {
				t.Fatalf("scaledNearestIndex(%d, %d, %d) = %d, want %d", index, target, source, got, want)
			}
		}
	}
}

// origin="lower" reflects the destination row, not the resulting source index.
// The rounding rule makes the two inequivalent: reflecting the index computes
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
		want := float64(aggIroundReference(targetHeight, rows, targetHeight-1-dstY))

		got, ok := img.scaledScalarValue(dstY, 0, targetHeight, targetWidth, rows, cols)
		if !ok {
			t.Fatalf("scaledScalarValue(%d) not ok", dstY)
		}
		if got != want {
			t.Fatalf("row for dstY=%d: got data row %v, want %v", dstY, got, want)
		}
	}
}
