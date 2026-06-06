package examplecatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCasesHaveStableUniqueIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Cases() {
		if c.ID == "" {
			t.Fatal("catalog case has empty ID")
		}
		if seen[c.ID] {
			t.Fatalf("duplicate catalog ID %q", c.ID)
		}
		seen[c.ID] = true
		if c.Topic == "" {
			t.Fatalf("%s has empty topic", c.ID)
		}
		if c.Title == "" {
			t.Fatalf("%s has empty title", c.ID)
		}
		if c.Width <= 0 || c.Height <= 0 || c.DPI <= 0 {
			t.Fatalf("%s dimensions/DPI = %dx%d @ %d", c.ID, c.Width, c.Height, c.DPI)
		}
	}
}

func TestCatalogReferencesCommittedParityImages(t *testing.T) {
	root := repoRoot(t)
	for _, c := range Cases() {
		requireFile(t, filepath.Join(root, "testdata", "golden", c.ID+".png"))
		requireFile(t, filepath.Join(root, "testdata", "matplotlib_ref", c.ID+".png"))
	}
}

func TestCatalogSourcePathsExistWhenRecorded(t *testing.T) {
	root := repoRoot(t)
	for _, c := range Cases() {
		if c.GoPath != "" {
			requireFile(t, filepath.Join(root, c.GoPath))
		}
		if c.PythonPath != "" {
			requireFile(t, filepath.Join(root, c.PythonPath))
		}
	}
}

func TestShowcaseRowsHaveGalleryMetadataAndRunnableSource(t *testing.T) {
	root := repoRoot(t)
	for _, c := range Cases() {
		if !c.Showcase {
			continue
		}
		t.Run(c.ID, func(t *testing.T) {
			if c.Title == "" {
				t.Fatal("showcase title is empty")
			}
			if strings.TrimSpace(c.Description) == "" || c.Description == c.Title {
				t.Fatalf("showcase description = %q, want gallery-ready text", c.Description)
			}
			if !strings.HasPrefix(c.GoPath, "examples/"+c.ID+"/") {
				t.Fatalf("showcase GoPath = %q, want examples/%s/...", c.GoPath, c.ID)
			}
			data, err := os.ReadFile(filepath.Join(root, c.GoPath))
			if err != nil {
				t.Fatalf("read showcase source %s: %v", c.GoPath, err)
			}
			source := string(data)
			if strings.Contains(source, "package main") {
				t.Fatalf("%s is a command package, want importable showcase package", c.GoPath)
			}
			if !strings.Contains(source, "func Plot() *core.Figure") {
				t.Fatalf("%s missing canonical Plot() *core.Figure snippet", c.GoPath)
			}
			if !strings.Contains(source, "func Render() image.Image") {
				t.Fatalf("%s missing runnable Render() image.Image snippet", c.GoPath)
			}
		})
	}
}

func TestCatalogIncludesProjectionAndToolkitFixtures(t *testing.T) {
	want := []string{
		"geo_aitoff_axes",
		"geo_hammer_axes",
		"geo_lambert_axes",
		"radar_basic",
		"skewt_basic",
		"mplot3d_basic",
		"mplot3d_terrain",
	}
	for _, id := range want {
		if _, ok := Lookup(id); !ok {
			t.Fatalf("missing projection/toolkit parity catalog case %q", id)
		}
	}
}

func TestCatalogIncludesMathTextFixtures(t *testing.T) {
	want := []string{
		"mathtext_basic",
		"mathtext_fractions",
		"mathtext_integrals",
		"mathtext_matrices",
		"mathtext_inline_labels",
	}
	for _, id := range want {
		if _, ok := Lookup(id); !ok {
			t.Fatalf("missing MathText parity catalog case %q", id)
		}
	}
}

func TestCatalogIncludesMixedRasterVectorFixture(t *testing.T) {
	c, ok := Lookup("mixed_raster_vector")
	if !ok {
		t.Fatal("missing mixed raster/vector parity catalog case")
	}
	if c.Topic != "raster" {
		t.Fatalf("mixed_raster_vector topic = %q, want raster", c.Topic)
	}
	if c.SVGGoldenFamily != "mixed_raster" {
		t.Fatalf("mixed_raster_vector SVGGoldenFamily = %q, want mixed_raster", c.SVGGoldenFamily)
	}
}

func TestCatalogIncludesPatchStyleMatrixFixture(t *testing.T) {
	c, ok := Lookup("patch_style_matrix")
	if !ok {
		t.Fatal("missing Phase 12.4 focused patch style parity catalog case")
	}
	if c.Topic != "patches" {
		t.Fatalf("patch_style_matrix topic = %q, want patches", c.Topic)
	}
	if !c.FixtureOnly {
		t.Fatal("patch_style_matrix should be fixture-only, not a gallery showcase")
	}
}

func TestCatalogIncludesLegendLayoutMatrixFixture(t *testing.T) {
	c, ok := Lookup("legend_layout_matrix")
	if !ok {
		t.Fatal("missing Phase 12.4 focused legend layout parity catalog case")
	}
	if c.Topic != "legend" {
		t.Fatalf("legend_layout_matrix topic = %q, want legend", c.Topic)
	}
	if !c.FixtureOnly {
		t.Fatal("legend_layout_matrix should be fixture-only, not a gallery showcase")
	}
}

func TestCatalogIncludesTextAnnotationMatrixFixture(t *testing.T) {
	c, ok := Lookup("text_annotation_matrix")
	if !ok {
		t.Fatal("missing Phase 12.4 focused text/annotation parity catalog case")
	}
	if c.Topic != "annotation" {
		t.Fatalf("text_annotation_matrix topic = %q, want annotation", c.Topic)
	}
	if !c.FixtureOnly {
		t.Fatal("text_annotation_matrix should be fixture-only, not a gallery showcase")
	}
}

func TestCatalogIncludesImshowInterpolationMatrixFixture(t *testing.T) {
	c, ok := Lookup("imshow_interpolation_matrix")
	if !ok {
		t.Fatal("missing Phase 12.5 focused interpolation matrix parity catalog case")
	}
	if c.Topic != "image" {
		t.Fatalf("imshow_interpolation_matrix topic = %q, want image", c.Topic)
	}
	if !c.FixtureOnly {
		t.Fatal("imshow_interpolation_matrix should be fixture-only, not a gallery showcase")
	}
	if c.Width < 600 || c.Height < 300 {
		t.Fatalf("imshow_interpolation_matrix dimensions = %dx%d, want compact gallery coverage", c.Width, c.Height)
	}
}

func TestCatalogIncludesColorImageGalleryShowcases(t *testing.T) {
	want := map[string]string{
		"named_colors_gallery":      "color",
		"colormap_families_gallery": "colormap",
		"image_variants_gallery":    "image",
		"colorbar_variants_gallery": "colorbar",
	}
	for id, topic := range want {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("missing Phase 18.2 showcase catalog case %q", id)
		}
		if c.Topic != topic {
			t.Fatalf("%s topic = %q, want %q", id, c.Topic, topic)
		}
		if !c.Showcase {
			t.Fatalf("%s should be a user-facing showcase", id)
		}
		if c.FixtureOnly {
			t.Fatalf("%s should not be fixture-only", id)
		}
	}
}

func TestCatalogIncludesTextGalleryShowcases(t *testing.T) {
	want := map[string]string{
		"mathtext_gallery":                    "mathtext",
		"text_layout_gallery":                 "text",
		"annotation_legend_offsetbox_gallery": "annotation",
	}
	for id, topic := range want {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("missing Phase 18.2 showcase catalog case %q", id)
		}
		if c.Topic != topic {
			t.Fatalf("%s topic = %q, want %q", id, c.Topic, topic)
		}
		if !c.Showcase {
			t.Fatalf("%s should be a user-facing showcase", id)
		}
		if c.FixtureOnly {
			t.Fatalf("%s should not be fixture-only", id)
		}
	}
}

func TestCatalogIncludesHighImpactToolkitOutputShowcases(t *testing.T) {
	want := map[string]string{
		"mplot3d_gallery":     "mplot3d",
		"mixed_raster_vector": "raster",
	}
	for id, topic := range want {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("missing high-impact toolkit/output showcase catalog case %q", id)
		}
		if c.Topic != topic {
			t.Fatalf("%s topic = %q, want %q", id, c.Topic, topic)
		}
		if !c.Showcase {
			t.Fatalf("%s should be a user-facing showcase", id)
		}
		if c.FixtureOnly {
			t.Fatalf("%s should not be fixture-only", id)
		}
		if c.GoPath != "examples/"+id+"/example.go" {
			t.Fatalf("%s GoPath = %q, want examples/%s/example.go", id, c.GoPath, id)
		}
	}
}

func TestCatalogIncludesWidgetsGalleryShowcase(t *testing.T) {
	c, ok := Lookup("widgets_gallery")
	if !ok {
		t.Fatal("missing Phase 12.5 widgets gallery catalog case")
	}
	if c.Topic != "widgets" {
		t.Fatalf("widgets_gallery topic = %q, want widgets", c.Topic)
	}
	if !c.Showcase {
		t.Fatal("widgets_gallery should be a user-facing showcase")
	}
}

func TestCatalogIncludesAnimationGalleryShowcase(t *testing.T) {
	c, ok := Lookup("animation_gallery")
	if !ok {
		t.Fatal("missing Phase 12.5 animation gallery catalog case")
	}
	if c.Topic != "animation" {
		t.Fatalf("animation_gallery topic = %q, want animation", c.Topic)
	}
	if !c.Showcase {
		t.Fatal("animation_gallery should be a user-facing showcase")
	}
}

func TestCatalogSplitsAGGNativeParityFixtures(t *testing.T) {
	want := map[string][]string{
		"large_scatter":     {"pathcollectionbatch"},
		"mixed_collection":  {"pathcollectionbatch"},
		"quad_mesh":         {"quadmeshbatch"},
		"gouraud_triangles": {"gouraudtrianglebatch"},
		"clip_path_batch":   {"pathclip", "quadmeshbatch"},
	}
	for id, capabilities := range want {
		c, ok := Lookup(id)
		if !ok {
			t.Fatalf("missing AGG-native parity catalog case %q", id)
		}
		if c.NativeBackend != "agg" {
			t.Fatalf("%s NativeBackend = %q, want agg", id, c.NativeBackend)
		}
		if !sameStrings(c.NativeCapabilities, capabilities) {
			t.Fatalf("%s NativeCapabilities = %v, want %v", id, c.NativeCapabilities, capabilities)
		}
	}
}

func TestCatalogIncludesSVGStructuralFamilies(t *testing.T) {
	want := map[string]string{
		"bar":           "bar_basic",
		"errorbar":      "errorbar_basic",
		"hist":          "hist_basic",
		"collection":    "mixed_collection",
		"image":         "image_heatmap",
		"clipped_polar": "polar_axes",
		"hatch_bars":    "patch_showcase",
		"text_layout":   "text_labels_strict",
		"mathtext":      "mathtext_basic",
	}
	got := map[string]string{}
	for _, c := range Cases() {
		if c.SVGGoldenFamily != "" {
			got[c.SVGGoldenFamily] = c.ID
		}
	}
	for family, id := range want {
		if got[family] != id {
			t.Fatalf("SVG structural family %q = %q, want %q", family, got[family], id)
		}
	}
}

func TestParityValidationClusters(t *testing.T) {
	want := map[string][]string{
		ValidationClusterLayoutText: {
			"figure_labels_composition",
			"text_labels_strict",
			"axes_top_right_inverted",
			"axes_control_surface",
			"transform_coordinates",
			"annotation_composition",
			"colorbar_composition",
		},
		ValidationClusterImageMeshColorbar: {
			"image_heatmap",
			"imshow_clipped",
			"imshow_transformed",
			"spy_image",
			"pcolor_flat",
			"pcolormesh_gouraud",
			"asinh_norm_image",
			"boundarynorm_pcolormesh",
			"collection_mutable_scalarmap",
			"colorbar_boundary_values",
			"colorbar_horizontal_ticks",
			"lognorm_imshow",
			"twoslope_norm_image",
			"colorbar_extensions",
		},
		ValidationClusterProjection3D: {
			"mplot3d_basic",
			"mplot3d_terrain",
			"mplot3d_plot3d",
			"mplot3d_scatter3d",
			"mplot3d_surface3d",
			"mplot3d_wire3d",
			"mplot3d_trisurf3d",
			"mplot3d_bar3d",
			"mplot3d_voxels",
			"mplot3d_quiver3d",
			"mplot3d_errorbar3d",
			"mplot3d_stem3d",
			"mplot3d_fill_between3d",
			"mplot3d_contour3d",
			"mplot3d_contourf3d",
			"mplot3d_tricontour3d",
			"mplot3d_tricontourf3d",
			"mplot3d_bar2d_zdir",
			"mplot3d_text3d",
			"polar_axes",
			"geo_mollweide_axes",
			"geo_aitoff_axes",
			"geo_hammer_axes",
			"geo_lambert_axes",
			"radar_basic",
			"skewt_basic",
		},
	}

	for clusterID, caseIDs := range want {
		cluster, ok := LookupValidationCluster(clusterID)
		if !ok {
			t.Fatalf("missing parity validation cluster %q", clusterID)
		}
		if strings.TrimSpace(cluster.Description) == "" {
			t.Fatalf("validation cluster %q has empty description", clusterID)
		}
		for _, id := range caseIDs {
			if !containsString(cluster.CaseIDs, id) {
				t.Fatalf("validation cluster %q missing required case %q", clusterID, id)
			}
			if _, ok := Lookup(id); !ok {
				t.Fatalf("validation cluster %q references missing catalog case %q", clusterID, id)
			}
		}
	}
}

func TestValidationClustersReferenceArtifactsExist(t *testing.T) {
	root := repoRoot(t)
	for _, cluster := range ValidationClusters() {
		cluster := cluster
		t.Run(cluster.ID, func(t *testing.T) {
			for _, id := range cluster.CaseIDs {
				requireFile(t, filepath.Join(root, "testdata", "golden", id+".png"))
				requireFile(t, filepath.Join(root, "testdata", "matplotlib_ref", id+".png"))
			}
		})
	}
}

func TestValidationClustersHaveStableMembers(t *testing.T) {
	seenClusters := map[string]bool{}
	for _, cluster := range ValidationClusters() {
		if cluster.ID == "" {
			t.Fatal("validation cluster has empty ID")
		}
		if seenClusters[cluster.ID] {
			t.Fatalf("duplicate validation cluster ID %q", cluster.ID)
		}
		seenClusters[cluster.ID] = true
		if len(cluster.CaseIDs) == 0 {
			t.Fatalf("validation cluster %q has no cases", cluster.ID)
		}

		seenCases := map[string]bool{}
		for _, id := range cluster.CaseIDs {
			if seenCases[id] {
				t.Fatalf("validation cluster %q repeats case %q", cluster.ID, id)
			}
			seenCases[id] = true
			if _, ok := Lookup(id); !ok {
				t.Fatalf("validation cluster %q references missing catalog case %q", cluster.ID, id)
			}
		}
	}
}

func TestParityFixValidationTargetsNameClusters(t *testing.T) {
	wantCaseIDs := []string{
		"fill_basic",
		"fill_stacked",
		"errorbar_basic",
		"boxplot_basic",
		"text_labels_strict",
		"mathtext_basic",
		"mathtext_fractions",
		"mathtext_integrals",
		"mathtext_matrices",
		"mathtext_inline_labels",
		"image_heatmap",
		"imshow_clipped",
		"imshow_transformed",
		"spy_marker",
		"spy_image",
		"axes_top_right_inverted",
		"axes_control_surface",
		"transform_coordinates",
		"figure_labels_composition",
		"colorbar_composition",
		"annotation_composition",
		"legend_layout_matrix",
		"text_annotation_matrix",
		"patch_showcase",
		"patch_style_matrix",
		"mesh_contour_tri",
		"plot_variants",
		"spectrum_variants",
		"specialty_depth",
		"stem_plot",
		"specialty_artists",
		"units_overview",
		"units_dates",
		"units_categories",
		"units_custom_converter",
		"vector_fields",
		"polar_axes",
		"geo_mollweide_axes",
		"geo_aitoff_axes",
		"geo_hammer_axes",
		"geo_lambert_axes",
		"radar_basic",
		"skewt_basic",
		"mplot3d_basic",
		"mplot3d_terrain",
		"mplot3d_plot3d",
		"mplot3d_scatter3d",
		"mplot3d_surface3d",
		"mplot3d_wire3d",
		"mplot3d_trisurf3d",
		"mplot3d_bar3d",
		"mplot3d_voxels",
		"mplot3d_quiver3d",
		"mplot3d_errorbar3d",
		"mplot3d_stem3d",
		"mplot3d_fill_between3d",
		"mplot3d_contour3d",
		"mplot3d_contourf3d",
		"mplot3d_tricontour3d",
		"mplot3d_tricontourf3d",
		"mplot3d_bar2d_zdir",
		"mplot3d_text3d",
		"unstructured_showcase",
		"arrays_showcase",
		"axisartist_showcase",
		"axes_grid1_showcase",
		"pcolor_flat",
		"pcolormesh_gouraud",
		"hist2d_weighted_density",
		"asinh_norm_image",
		"boundarynorm_pcolormesh",
		"collection_mutable_scalarmap",
		"colorbar_boundary_values",
		"colorbar_horizontal_ticks",
		"lognorm_imshow",
		"twoslope_norm_image",
		"colorbar_extensions",
	}

	clusterCases := map[string][]string{}
	for _, cluster := range ValidationClusters() {
		clusterCases[cluster.ID] = cluster.CaseIDs
	}

	got := map[string]ParityFixValidationTarget{}
	for _, target := range ParityFixValidationTargets() {
		if _, ok := Lookup(target.CaseID); !ok {
			t.Fatalf("parity validation target references missing catalog case %q", target.CaseID)
		}
		if _, exists := got[target.CaseID]; exists {
			t.Fatalf("duplicate parity validation target for %q", target.CaseID)
		}
		if len(target.ClusterIDs) == 0 {
			t.Fatalf("parity validation target %q names no clusters", target.CaseID)
		}
		for _, clusterID := range target.ClusterIDs {
			caseIDs, ok := clusterCases[clusterID]
			if !ok {
				t.Fatalf("parity validation target %q references missing cluster %q", target.CaseID, clusterID)
			}
			if !containsString(caseIDs, target.CaseID) {
				t.Fatalf("parity validation target %q names cluster %q, but the cluster does not include that case", target.CaseID, clusterID)
			}
		}
		got[target.CaseID] = target
	}

	for _, id := range wantCaseIDs {
		if _, ok := got[id]; !ok {
			t.Fatalf("parity case %q has no validation target cluster", id)
		}
	}
	if len(got) != len(wantCaseIDs) {
		t.Fatalf("parity validation target count = %d, want %d", len(got), len(wantCaseIDs))
	}
}

func TestNoCatalogIDsInLibraryImplementation(t *testing.T) {
	root := repoRoot(t)
	scanDirs := []string{"core", "render", "backends"}
	var quotedIDs []string
	for _, c := range Cases() {
		quotedIDs = append(quotedIDs, `"`+c.ID+`"`, "`"+c.ID+"`")
	}

	for _, dir := range scanDirs {
		dir := dir
		t.Run(dir, func(t *testing.T) {
			err := filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				source := string(data)
				for _, quotedID := range quotedIDs {
					if strings.Contains(source, quotedID) {
						rel, relErr := filepath.Rel(root, path)
						if relErr != nil {
							rel = path
						}
						t.Fatalf("%s contains catalog case ID %s; parity hardening forbids fixture-specific library logic", rel, quotedID)
					}
				}
				return nil
			})
			if err != nil {
				t.Fatalf("scan %s: %v", dir, err)
			}
		})
	}
}

func TestWebDemosAreParityCasesWithReferences(t *testing.T) {
	root := repoRoot(t)
	seen := map[string]bool{}
	for _, c := range WebDemos() {
		if c.WebDemoID == "" {
			t.Fatalf("%s has empty WebDemoID", c.ID)
		}
		if seen[c.WebDemoID] {
			t.Fatalf("duplicate web demo ID %q", c.WebDemoID)
		}
		seen[c.WebDemoID] = true
		if _, ok := Lookup(c.ID); !ok {
			t.Fatalf("web demo %q does not resolve to a parity case", c.WebDemoID)
		}
		requireFile(t, filepath.Join(root, "test", "matplotlib_ref", "webdemos", c.WebDemoID+".py"))
	}
	if len(seen) < 8 {
		t.Fatalf("web demo catalog has %d entries, want a curated but varied set", len(seen))
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func requireFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}
