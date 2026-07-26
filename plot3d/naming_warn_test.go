package plot3d

import (
	"strings"
	"testing"
)

// PlotSurface is a well-known Matplotlib name, but this port's PlotSurface
// merely draws a connected line strip (it delegates to Plot3D). That silent
// divergence must be surfaced, once per axes, pointing at the real Surface API.
func TestPlotSurfaceWarnsAndPointsToSurface(t *testing.T) {
	got, restore := captureWarnings()
	defer restore()

	fig := NewFigure(400, 300)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	x := []float64{0, 1, 2, 3}
	ax.PlotSurface(x, x, x, PlotOptions{})
	ax.PlotSurface(x, x, x, PlotOptions{})

	if len(*got) != 1 {
		t.Fatalf("PlotSurface should warn exactly once per axes, got %d: %v", len(*got), *got)
	}
	if !strings.Contains((*got)[0], "Surface") {
		t.Fatalf("warning %q should point to the real Surface API", (*got)[0])
	}
}

// Voxel (singular) draws unstructured wireframe prisms via Bar3D, not the
// filled cubes Matplotlib's voxels() produces. Surface the trap and point at
// the real Voxels API.
func TestVoxelWarnsAndPointsToVoxels(t *testing.T) {
	got, restore := captureWarnings()
	defer restore()

	fig := NewFigure(400, 300)
	ax, err := AddAxes(fig, unitRect())
	if err != nil {
		t.Fatalf("AddAxes3D: %v", err)
	}

	one := []float64{0, 1}
	ax.Voxel(one, one, one, one, one, one, PlotOptions{})
	ax.Voxel(one, one, one, one, one, one, PlotOptions{})

	if len(*got) != 1 {
		t.Fatalf("Voxel should warn exactly once per axes, got %d: %v", len(*got), *got)
	}
	if !strings.Contains((*got)[0], "Voxels") {
		t.Fatalf("warning %q should point to the real Voxels API", (*got)[0])
	}
}
