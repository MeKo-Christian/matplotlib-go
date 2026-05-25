package widgets_gallery

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/core"
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
