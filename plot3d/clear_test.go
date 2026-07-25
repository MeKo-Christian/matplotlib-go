package plot3d

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
)

func TestAxes3DClearViaEmbedding(t *testing.T) {
	fig := core.NewFigure(400, 300)
	ax, err := AddAxes(fig, geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	if err != nil {
		t.Fatalf("AddAxes: %v", err)
	}
	_, _ = ax.Plot([]float64{0, 1}, []float64{0, 1})
	ax.Clear()
	if got := len(ax.Artists); got != 0 {
		t.Fatalf("after Axes3D.Clear: %d artists remain, want 0", got)
	}
}
