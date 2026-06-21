package widgets_gallery

import (
	"math"
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func TestPlotUsesMatplotlibWidgetVisualStyle(t *testing.T) {
	fig := Plot()
	if fig.RC.WidgetVisualStyle != style.WidgetVisualMatplotlib {
		t.Fatalf("widget visual style = %q, want %q", fig.RC.WidgetVisualStyle, style.WidgetVisualMatplotlib)
	}

	button := findWidget[*core.Button](fig)
	if button == nil {
		t.Fatal("parity widgets gallery should include a button")
	}
	if button.FaceColor != (render.Color{R: 0.85, G: 0.85, B: 0.85, A: 1}) {
		t.Fatalf("button face = %+v, want Matplotlib default gray", button.FaceColor)
	}
}

func TestPlotUsesMatplotlibReferenceAxesLayout(t *testing.T) {
	fig := Plot()
	if len(fig.Children) < 8 {
		t.Fatalf("axes count = %d, want at least 8 reference axes", len(fig.Children))
	}
	assertRectEqual(t, fig.Children[0].RectFraction, geom.Rect{Min: geom.Pt{X: 0.06, Y: 0.36}, Max: geom.Pt{X: 0.70, Y: 0.92}}, "main axes")
	assertRectEqual(t, fig.Children[1].RectFraction, geom.Rect{Min: geom.Pt{X: 0.76, Y: 0.36}, Max: geom.Pt{X: 0.94, Y: 0.92}}, "aux axes")
	assertRectEqual(t, fig.Children[2].RectFraction, geom.Rect{Min: geom.Pt{X: 0.06, Y: 0.23}, Max: geom.Pt{X: 0.22, Y: 0.30}}, "button axes")
	if fig.Children[1].XAxis != fig.Children[0].XAxis {
		t.Fatal("aux axes should share the main x-axis object like Matplotlib sharex")
	}
}

func findWidget[T any](fig *core.Figure) T {
	var zero T
	if fig == nil {
		return zero
	}
	for _, ax := range fig.Children {
		if ax == nil {
			continue
		}
		for _, art := range ax.WidgetArtists {
			if widget, ok := art.(T); ok {
				return widget
			}
		}
	}
	return zero
}

func assertRectEqual(t *testing.T, got, want geom.Rect, label string) {
	t.Helper()
	const tol = 1e-12
	if math.Abs(got.Min.X-want.Min.X) > tol ||
		math.Abs(got.Min.Y-want.Min.Y) > tol ||
		math.Abs(got.Max.X-want.Max.X) > tol ||
		math.Abs(got.Max.Y-want.Max.Y) > tol {
		t.Fatalf("%s rect = %+v, want %+v", label, got, want)
	}
}
