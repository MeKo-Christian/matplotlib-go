package plot_variants

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestReferenceLineDashesMatchMatplotlibPointUnits(t *testing.T) {
	fig := Plot()
	if len(fig.Children) < 2 {
		t.Fatalf("figure children = %d, want fill-between axes", len(fig.Children))
	}
	fillAx := fig.Children[1]

	var segments []*core.Segment2D
	for _, art := range fillAx.Artists {
		if seg, ok := art.(*core.Segment2D); ok && len(seg.Dashes) > 0 {
			segments = append(segments, seg)
		}
	}
	if len(segments) != 2 {
		t.Fatalf("dashed reference lines = %d, want hline and vline", len(segments))
	}

	wantHLine := []float64{4 * 36.0 / DPI, 3 * 36.0 / DPI}
	wantVLine := []float64{2 * 36.0 / DPI, 2 * 36.0 / DPI}
	assertDashSeq(t, segments[0].Dashes, wantHLine)
	assertDashSeq(t, segments[1].Dashes, wantVLine)
}

func assertDashSeq(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("dash sequence length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("dash sequence = %v, want %v", got, want)
		}
	}
}
