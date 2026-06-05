package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestColorImageColorbarFixtureGapInventoryIsDocumented(t *testing.T) {
	imageIDs := []string{
		"image_heatmap",
		"imshow_clipped",
		"imshow_transformed",
		"imshow_bilinear",
		"imshow_bicubic",
		"imshow_interpolation_matrix",
		"image_alpha",
		"matshow_basic",
		"spy_marker",
		"spy_image",
		"arrays_showcase",
	}
	colorIDs := []string{
		"colormap_diverging",
		"colormap_qualitative",
		"colormap_cyclic",
		"named_colors",
	}
	colorbarIDs := []string{
		"asinh_norm_image",
		"boundarynorm_pcolormesh",
		"collection_mutable_scalarmap",
		"colorbar_boundary_values",
		"colorbar_horizontal_ticks",
		"lognorm_imshow",
		"twoslope_norm_image",
		"colorbar_extensions",
	}
	validationIDs := []string{
		"image_heatmap",
		"imshow_clipped",
		"imshow_transformed",
		"spy_marker",
		"spy_image",
		"asinh_norm_image",
		"boundarynorm_pcolormesh",
		"collection_mutable_scalarmap",
		"colorbar_boundary_values",
		"colorbar_horizontal_ticks",
		"lognorm_imshow",
		"twoslope_norm_image",
		"colorbar_extensions",
		"colorbar_composition",
	}
	referenceIDs := []string{
		"image_heatmap",
		"imshow_clipped",
		"imshow_transformed",
		"image_alpha",
		"matshow_basic",
		"spy_marker",
		"spy_image",
		"arrays_showcase",
		"colormap_diverging",
		"colormap_qualitative",
		"colormap_cyclic",
		"named_colors",
	}
	referenceIDs = append(referenceIDs, colorbarIDs...)

	sourceRequirements := map[string][]string{
		filepath.Join("..", "internal", "examplecatalog", "catalog.go"):               append(append(append([]string{}, imageIDs...), colorIDs...), colorbarIDs...),
		filepath.Join("..", "internal", "examplecatalog", "feature_coverage.go"):      append(append(append([]string{}, imageIDs...), colorIDs...), colorbarIDs...),
		filepath.Join("..", "internal", "examplecatalog", "validation_clusters.go"):   validationIDs,
		filepath.Join("..", "internal", "examplecatalog", "parity_validation.go"):     validationIDs,
		filepath.Join("..", "test", "parity", "registry.go"):                          append(append([]string{}, imageIDs...), colorbarIDs...),
		filepath.Join("..", "test", "matplotlib_ref", "plots", "__init__.py"):         referenceIDs,
		filepath.Join("..", "internal", "examplecatalog", "public_surface_parity.go"): {"LightSource", "BivarColormap", "FuncNorm", "colorbar_composition"},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing fixture inventory marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Fixture Gap Inventory",
		"Color fixtures currently cover named colors plus diverging, qualitative, and cyclic scalar colormaps",
		"Image fixtures cover heatmap, clipped, transformed, bilinear, bicubic, interpolation-matrix, alpha, matshow, spy-marker, spy-image, and arrays-showcase paths",
		"Norm and colorbar fixtures cover log, asinh, two-slope, boundary/discrete, mutable scalar-map, explicit ticks, and extensions",
		"Missing high-priority fixture work is ordered as norm/FuncNorm/boundary/two-slope and colormap/discrete triplets before transformed-image/LightSource and colorbar placement/formatter/update triplets",
		"LightSource and shaded-relief image fixtures remain absent by design",
		"bivariate and multivariate colormap fixture work remains deferred until a 2D or multi-component colorbar contract exists",
		"Potential weak coverage: scalar colormap families are represented by only three fixture rows and colorbar composition is the only showcase example",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("fixture gap inventory docs missing %q", phrase)
		}
	}
}

func TestColorImageColorbarFixtureTripletsAreDocumented(t *testing.T) {
	tripletIDs := []string{
		"colormap_diverging",
		"colormap_qualitative",
		"colormap_cyclic",
		"named_colors",
		"image_heatmap",
		"imshow_clipped",
		"imshow_transformed",
		"imshow_interpolation_matrix",
		"image_alpha",
		"matshow_basic",
		"spy_marker",
		"spy_image",
		"asinh_norm_image",
		"boundarynorm_pcolormesh",
		"collection_mutable_scalarmap",
		"colorbar_boundary_values",
		"colorbar_horizontal_ticks",
		"lognorm_imshow",
		"twoslope_norm_image",
		"colorbar_extensions",
	}
	sourceRequirements := map[string][]string{
		filepath.Join("..", "internal", "examplecatalog", "catalog.go"): tripletIDs,
		filepath.Join("..", "testdata", "golden"): {
			"colormap_diverging.png",
			"colormap_qualitative.png",
			"colormap_cyclic.png",
			"image_heatmap.png",
			"imshow_clipped.png",
			"imshow_transformed.png",
			"image_alpha.png",
			"matshow_basic.png",
			"spy_marker.png",
			"spy_image.png",
			"asinh_norm_image.png",
			"boundarynorm_pcolormesh.png",
			"collection_mutable_scalarmap.png",
			"colorbar_boundary_values.png",
			"colorbar_horizontal_ticks.png",
			"lognorm_imshow.png",
			"twoslope_norm_image.png",
			"colorbar_extensions.png",
		},
		filepath.Join("..", "test", "matplotlib_ref", "plots", "__init__.py"): {
			"colormap_diverging",
			"colormap_qualitative",
			"colormap_cyclic",
			"asinh_norm_image",
			"boundarynorm_pcolormesh",
			"collection_mutable_scalarmap",
			"colorbar_boundary_values",
			"colorbar_horizontal_ticks",
			"lognorm_imshow",
			"twoslope_norm_image",
			"colorbar_extensions",
		},
	}
	for path, phrases := range sourceRequirements {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		var src string
		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				t.Fatalf("read dir %s: %v", path, err)
			}
			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				names = append(names, entry.Name())
			}
			src = strings.Join(names, " ")
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			src = string(data)
		}
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing fixture triplet marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Color/Image/Colorbar Fixture Triplets",
		"Committed triplets cover scalar colormap swatches, named colors, image heatmap/clipped/transformed/interpolation/alpha/matshow/spy paths, and the colorbar norm/update/extension set",
		"Existing norm triplets cover LogNorm, AsinhNorm, TwoSlopeNorm, BoundaryNorm, and Normalize-backed mutable scalar maps",
		"No new FuncNorm, LightSource, shaded-relief, bivariate, or multivariate triplet is added for this phase because those APIs are documented omissions",
		"Colorbar placement/formatter/boundary/extension/update-contract coverage is represented by `colorbar_composition`, `colorbar_horizontal_ticks`, `colorbar_boundary_values`, `colorbar_extensions`, and `collection_mutable_scalarmap`",
		"Focused visual checks for this phase are `TestGolden` and `TestMatplotlibRef` on colormap_diverging, colormap_qualitative, colormap_cyclic, asinh_norm_image, boundarynorm_pcolormesh, collection_mutable_scalarmap, colorbar_boundary_values, colorbar_horizontal_ticks, lognorm_imshow, twoslope_norm_image, and colorbar_extensions",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("fixture triplet docs missing %q", phrase)
		}
	}
}
