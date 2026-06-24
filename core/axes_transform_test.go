package core

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

// drawNull renders the figure through a no-op renderer so the persistent
// transform graph is exercised exactly as in a real draw.
func drawNull(fig *Figure) {
	var r render.NullRenderer
	DrawFigure(fig, &r)
}

func TestAxesTransformReusedWhenSizeUnchanged(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{})

	drawNull(fig)
	if ax.axesBbox == nil {
		t.Fatal("axesBbox was not initialized by the draw")
	}
	v1 := ax.axesBbox.Version()

	// Redrawing at the same size must not invalidate the axes->pixel transform.
	drawNull(fig)
	if v2 := ax.axesBbox.Version(); v2 != v1 {
		t.Fatalf("redraw at same size invalidated axesBbox: version %d -> %d", v1, v2)
	}
}

func TestAxesTransformInvalidatedOnResize(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.1}, Max: geom.Pt{X: 0.9, Y: 0.9}})
	ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{})

	drawNull(fig)
	v1 := ax.axesBbox.Version()
	before := ax.transData.Apply(geom.Pt{X: 1, Y: 1})

	// Resize the figure: the axes pixel rectangle changes, so the persistent
	// transform must be invalidated and recompute a new mapping.
	fig.SizePx = geom.Pt{X: 800, Y: 600}
	drawNull(fig)

	if v2 := ax.axesBbox.Version(); v2 == v1 {
		t.Fatalf("resize did not invalidate axesBbox (version stayed %d)", v1)
	}
	after := ax.transData.Apply(geom.Pt{X: 1, Y: 1})
	if approxPtCore(before, after, 1e-9) {
		t.Fatalf("transData did not change after resize: %+v == %+v", before, after)
	}
}
