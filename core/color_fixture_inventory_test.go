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
		"Phase 17.6.5 Fixture Gap Inventory",
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
		"Phase 17.6.5 Color/Image/Colorbar Fixture Triplets",
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

func TestColorImageColorbarMetadataAndMigrationNotesAreDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "internal", "examplecatalog", "public_surface_parity.go"): {
			"idPrefix:          \"image\"",
			"idPrefix:          \"colorbar\"",
			"idPrefix:          \"colorizer\"",
			"idPrefix:          \"cm\"",
			"idPrefix:          \"colors\"",
			"colors-normalize-class",
			"colors-boundarynorm-class",
			"colors-asinhnorm-class",
			"colors-twoslope-norm-class",
			"FuncNorm is intentionally omitted",
			"LightSource as an intentional omission",
			"bivar/multivar colormaps",
			"pyplot-imshow",
			"pyplot-colorbar",
		},
		filepath.Join("..", "docs", "matplotlib-parity-status.md"): {
			"image.py:class:AxesImage",
			"colorbar.py:class:Colorbar",
			"cm.py:function:get_cmap",
			"colors.py:class:ColorConverter",
			"pyplot.py:function:imshow",
			"pyplot.py:function:colorbar",
			"mutable mappable clim/colormap/norm updates",
			"LightSource and bivar/multivar colormaps have explicit intentional-omission rows",
		},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing metadata marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.6.5 Metadata and Migration Notes",
		"Public-surface metadata marks color conversion, scalar colormaps, norms, images, colorbars, and colorizer routing with Phase 17.6.5 notes",
		"Implemented fixture IDs are attached to image, colorbar, colors-cm, Normalize, BoundaryNorm, AsinhNorm, TwoSlopeNorm, pyplot imshow, pyplot colorbar, current-image, and current-mappable rows",
		"Intentional omissions remain recorded for FuncNorm as a concrete type, LightSource, bivariate colormaps, and multivariate colormaps",
		"Migration notes summarize typed Go API differences for dynamic Python color inputs, callback-driven colorbar updates, custom colorbar formatters, gridspec and multi-parent colorbar helpers, and omitted shaded-relief or multi-component colorbars",
		"`docs/matplotlib-parity-status.md` is generated from `internal/examplecatalog` and carries the updated color/image/colorbar row notes",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("metadata and migration docs missing %q", phrase)
		}
	}
}

func TestFinalColorStatusRegenerationIsDocumented(t *testing.T) {
	// The per-milestone PLAN.md markers were retired when the roadmap was
	// restructured (completed phases now live in git history); the generated
	// docs and their guard tests remain the guarded surface.
	sourceRequirements := map[string][]string{
		filepath.Join("..", "docs", "matplotlib-parity-status.md"): {
			"Generated from `internal/examplecatalog` and `test/testdata/parity_surface/upstream_public_surface.json`.",
			"colorbar.py:class:Colorbar",
			"colors.py:class:ColorConverter",
			"image.py:class:AxesImage",
			"mutable mappable clim/colormap/norm updates",
		},
		filepath.Join("..", "internal", "examplecatalog", "public_surface_inventory_test.go"): {
			"TestMatplotlibParityStatusDocIsCurrent",
			"TestPublicSurfaceParityRowsCoverCommittedInventory",
			"TestPublicSurfaceParityRowsReferenceExistingLocalArtifacts",
		},
	}
	for path, phrases := range sourceRequirements {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		for _, phrase := range phrases {
			if !strings.Contains(src, phrase) {
				t.Fatalf("%s missing final color status marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.6.5 Final Color Status Regeneration",
		"`docs/matplotlib-parity-status.md` was regenerated from `cmd/paritystatusdoc` after color, image, norm, and colorbar metadata updates",
		"The final sweep covers Golden, MatplotlibRef, and ReferenceCompare for color, norm, image, and colorbar fixtures",
		"The catalog/doc freshness checks are `TestMatplotlibParityStatusDocIsCurrent`, `TestPublicSurfaceParityRowsCoverCommittedInventory`, and `TestPublicSurfaceParityRowsReferenceExistingLocalArtifacts`",
		"`17.6.5.7` and `17.6.5.7.4` are closed only after those checks pass",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("final color status docs missing %q", phrase)
		}
	}
}
