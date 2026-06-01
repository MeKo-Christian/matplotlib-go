package mathtext_basic

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	common "github.com/cwbudde/matplotlib-go/internal/parityutil"
)

func TestPlotConvertsAnnotationOffsetPoints(t *testing.T) {
	fig := Plot()
	if len(fig.Children) != 1 {
		t.Fatalf("axes count = %d, want 1", len(fig.Children))
	}

	var annotation *core.Annotation
	for _, artist := range fig.Children[0].Artists {
		if ann, ok := artist.(*core.Annotation); ok && ann.Content == `$\Delta y \approx \frac{1}{2}$` {
			annotation = ann
			break
		}
	}
	if annotation == nil {
		t.Fatal("math annotation not found")
	}

	wantX := common.ReferencePointsToPixels(34)
	wantY := common.ReferencePointsToPixels(-26)
	if math.Abs(annotation.OffsetX-wantX) > 1e-9 || math.Abs(annotation.OffsetY-wantY) > 1e-9 {
		t.Fatalf("annotation offset = (%g, %g), want points converted to pixels (%g, %g)",
			annotation.OffsetX, annotation.OffsetY, wantX, wantY)
	}
}
