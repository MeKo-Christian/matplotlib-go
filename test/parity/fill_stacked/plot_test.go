package fill_stacked

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
)

func TestPlotFirstLayerMatchesMatplotlibArgumentOrder(t *testing.T) {
	fig := Plot()
	if len(fig.Children) != 1 {
		t.Fatalf("axes count = %d, want 1", len(fig.Children))
	}
	if len(fig.Children[0].Artists) == 0 {
		t.Fatal("expected stacked fill artists")
	}
	fill, ok := fig.Children[0].Artists[0].(*core.Fill2D)
	if !ok {
		t.Fatalf("first artist = %T, want *core.Fill2D", fig.Children[0].Artists[0])
	}
	if len(fill.Y1) == 0 || fill.Y1[0] != 0 {
		t.Fatalf("first fill lower boundary starts at %v, want Matplotlib fill_between(x, 0, layer1) order", fill.Y1)
	}
	if len(fill.Y2) == 0 || fill.Y2[0] != 1 {
		t.Fatalf("first fill upper boundary starts at %v, want layer1", fill.Y2)
	}
}
