package mathtext_basic

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestPlotUsesAnnotationOffsetPoints(t *testing.T) {
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

	if annotation.OffsetX != 34 || annotation.OffsetY != -26 || annotation.OffsetUnits != core.AnnotationOffsetPoints {
		t.Fatalf("annotation offset = (%g, %g) units=%v, want raw point offsets (34, -26)",
			annotation.OffsetX, annotation.OffsetY, annotation.OffsetUnits)
	}
	if annotation.ArrowStyle.Name != "->" {
		t.Fatalf("annotation arrow style = %q, want ->", annotation.ArrowStyle.Name)
	}
}
