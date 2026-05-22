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
