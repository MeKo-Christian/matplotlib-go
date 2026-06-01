package spectrum_variants

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/style"
)

func TestPlotUsesMatplotlibDefaultTextColor(t *testing.T) {
	fig := Plot()
	if fig.RC.TextColor != style.Default.TextColor {
		t.Fatalf("figure text color = %v, want Matplotlib default %v", fig.RC.TextColor, style.Default.TextColor)
	}
}
