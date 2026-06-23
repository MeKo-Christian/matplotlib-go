package core

import (
	"reflect"
	"testing"

	"github.com/cwbudde/matplotlib-go/cycler"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
)

func multiPropCycle(t *testing.T) *cycler.Cycler {
	t.Helper()
	red := render.Color{R: 1, A: 1}
	green := render.Color{G: 1, A: 1}
	c, err := cycler.New("color", red, green).Concat(cycler.New("linestyle", "-", "--"))
	if err != nil {
		t.Fatalf("concat linestyle: %v", err)
	}
	if c, err = c.Concat(cycler.New("marker", "o", "s")); err != nil {
		t.Fatalf("concat marker: %v", err)
	}
	if c, err = c.Concat(cycler.New("linewidth", 1.5, 3.0)); err != nil {
		t.Fatalf("concat linewidth: %v", err)
	}
	return c
}

func TestPlotConsumesMultiPropertyCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	fig.RC.PropCycle = multiPropCycle(t)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	first := ax.Plot([]float64{0, 1}, []float64{0, 1})
	second := ax.Plot([]float64{0, 1}, []float64{1, 0})

	// First step: solid line, circle marker, width 1.5.
	if got, want := first.Col, (render.Color{R: 1, A: 1}); got != want {
		t.Fatalf("first color = %+v, want %+v", got, want)
	}
	if first.Dashes != nil {
		t.Fatalf("first dashes = %v, want nil (solid)", first.Dashes)
	}
	if !first.MarkerSet || first.Marker != MarkerCircle {
		t.Fatalf("first marker set=%v type=%v, want circle", first.MarkerSet, first.Marker)
	}
	if first.W != 1.5 {
		t.Fatalf("first width = %v, want 1.5", first.W)
	}

	// Second step: dashed line, square marker, width 3.0.
	if got, want := second.Col, (render.Color{G: 1, A: 1}); got != want {
		t.Fatalf("second color = %+v, want %+v", got, want)
	}
	if want := lineStyleToDashes("--", 3.0); !reflect.DeepEqual(second.Dashes, want) {
		t.Fatalf("second dashes = %v, want %v", second.Dashes, want)
	}
	if !second.MarkerSet || second.Marker != MarkerSquare {
		t.Fatalf("second marker set=%v type=%v, want square", second.MarkerSet, second.Marker)
	}
	if second.W != 3.0 {
		t.Fatalf("second width = %v, want 3.0", second.W)
	}
}

func TestPlotExplicitOptionsOverrideCycle(t *testing.T) {
	fig := NewFigure(640, 480)
	fig.RC.PropCycle = multiPropCycle(t)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})

	lw := 7.0
	marker := MarkerTriangle
	dashes := []float64{2, 2}
	line := ax.Plot([]float64{0, 1}, []float64{0, 1}, PlotOptions{
		LineWidth: &lw,
		Marker:    &marker,
		Dashes:    dashes,
	})

	if line.W != 7.0 {
		t.Fatalf("width = %v, want explicit 7.0", line.W)
	}
	if line.Marker != MarkerTriangle {
		t.Fatalf("marker = %v, want explicit triangle", line.Marker)
	}
	if !reflect.DeepEqual(line.Dashes, dashes) {
		t.Fatalf("dashes = %v, want explicit %v", line.Dashes, dashes)
	}
}

func TestPlotColorOnlyCycleUnchanged(t *testing.T) {
	// With no PropCycle (the default), lines keep the historical defaults:
	// solid, no marker, width 2.0, colors from the palette.
	fig := NewFigure(640, 480)
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{}, Max: geom.Pt{X: 1, Y: 1}})
	palette := fig.RC.Palette()

	line := ax.Plot([]float64{0, 1}, []float64{0, 1})
	if line.Dashes != nil || line.MarkerSet || line.W != 2.0 {
		t.Fatalf("color-only defaults changed: dashes=%v markerSet=%v w=%v", line.Dashes, line.MarkerSet, line.W)
	}
	if line.Col != palette[0] {
		t.Fatalf("color = %+v, want palette[0] %+v", line.Col, palette[0])
	}
}
