package core

import "testing"

func TestAxesSetAxisBelowTruePlacesGridsBelowPatches(t *testing.T) {
	ax := &Axes{}
	grid := ax.AddGrid(AxisLeft)

	ax.SetAxisBelow(true)

	if got, want := grid.Z(), 0.5; got != want {
		t.Fatalf("existing grid zorder = %v, want %v", got, want)
	}
	if !(grid.Z() < defaultPatchZ) {
		t.Fatalf("existing grid zorder = %v, want below patch zorder %v", grid.Z(), defaultPatchZ)
	}

	later := ax.AddGrid(AxisBottom)
	if got, want := later.Z(), 0.5; got != want {
		t.Fatalf("new grid zorder = %v, want %v", got, want)
	}
}
