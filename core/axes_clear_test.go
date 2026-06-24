package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
)

func testAxes() *Axes {
	fig := NewFigure(400, 300)
	return fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
}

func TestAxesClearResetsState(t *testing.T) {
	ax := testAxes()
	ax.Plot([]float64{0, 1, 2}, []float64{0, 1, 4})
	ax.SetTitle("hello")
	ax.XLabel = "x"
	ax.YLabel = "y"
	ax.XAxisTop = NewXAxis()
	ax.YAxisRight = NewYAxis()
	ax.SetXLim(-5, 5)

	ax.Clear()

	if lines := FindobjType[*Line2D](ax); len(lines) != 0 {
		t.Fatalf("after Clear: %d lines remain, want 0", len(lines))
	}
	if ax.Title != "" || ax.XLabel != "" || ax.YLabel != "" {
		t.Fatalf("after Clear: labels not reset: title=%q x=%q y=%q", ax.Title, ax.XLabel, ax.YLabel)
	}
	if ax.XAxisTop != nil || ax.YAxisRight != nil {
		t.Fatalf("after Clear: secondary axes not cleared")
	}
	if ax.xLimitsManual {
		t.Fatalf("after Clear: xLimitsManual still set")
	}
	if ax.XAxis == nil || ax.YAxis == nil {
		t.Fatalf("after Clear: primary axes missing")
	}
	if ax.XScale == nil || ax.YScale == nil {
		t.Fatalf("after Clear: scales not re-established")
	}
}

func TestAxesClaAliasesClear(t *testing.T) {
	ax := testAxes()
	ax.Plot([]float64{0, 1}, []float64{0, 1})
	ax.Cla()
	if lines := FindobjType[*Line2D](ax); len(lines) != 0 {
		t.Fatalf("after Cla: %d lines remain, want 0", len(lines))
	}
}

func TestAxes3DClearViaEmbedding(t *testing.T) {
	fig := NewFigure(400, 300)
	ax3d, err := fig.AddAxes3D(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}
	ax3d.Plot([]float64{0, 1}, []float64{0, 1})
	ax3d.Clear()
	if lines := FindobjType[*Line2D](ax3d.Axes); len(lines) != 0 {
		t.Fatalf("after Axes3D.Clear: %d lines remain, want 0", len(lines))
	}
}
