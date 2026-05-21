package parity

import "testing"

func TestFigureReturnsPDFGoldenSubset(t *testing.T) {
	ids := []string{
		"basic_line",
		"bar_basic",
		"scatter_basic",
		"hist_basic",
		"mesh_contour_tri",
		"image_heatmap",
		"polar_axes",
		"patch_showcase",
		"text_labels_strict",
		"imshow_clipped",
		"imshow_transformed",
		"mixed_raster_vector",
	}
	for _, id := range ids {
		id := id
		t.Run(id, func(t *testing.T) {
			fig, c, err := Figure(id)
			if err != nil {
				t.Fatalf("Figure(%q): %v", id, err)
			}
			if c.ID != id {
				t.Fatalf("Figure(%q) returned case %q", id, c.ID)
			}
			if fig == nil {
				t.Fatalf("Figure(%q) returned nil figure", id)
			}
		})
	}
}
