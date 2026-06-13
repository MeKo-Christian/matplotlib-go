package line2d_markers

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
)

func TestPlotLinesMirrorMatplotlibButtCapstyle(t *testing.T) {
	fig := Plot()
	if len(fig.Children) != 1 {
		t.Fatalf("axes count = %d, want 1", len(fig.Children))
	}
	for i, art := range fig.Children[0].Artists {
		line, ok := art.(*core.Line2D)
		if !ok {
			continue
		}
		if !line.LineCapSet || line.LineCap != render.CapButt {
			t.Fatalf("line artist %d cap = set:%v %v, want Python solid_capstyle='butt'", i, line.LineCapSet, line.LineCap)
		}
	}
}
