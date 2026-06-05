package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestColorbarMutableMappableUpdateContractIsDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorbar.py"): {
			"mappable.colorbar_cid = mappable.callbacks.connect(",
			"'changed', self.update_normal)",
			"def update_normal(self, mappable=None):",
			"self.set_alpha(self.mappable.get_alpha())",
			"self.cmap = self.mappable.cmap",
			"if self.mappable.norm != self.norm:",
			"self._reset_locator_formatter_scale()",
		},
		filepath.Join("..", "third_party", "matplotlib", "lib", "matplotlib", "colorizer.py"): {
			"def set_clim(self, vmin=None, vmax=None):",
			"self.norm.callbacks.process('changed')",
			"def changed(self):",
			"self.callbacks.process('changed')",
		},
		filepath.Join("..", "core", "scalar_mappable.go"): {
			"type ScalarMappable interface",
			"ScalarMap() ScalarMapInfo",
			"type ScalarMapInfo struct",
			"Colormap string",
			"Norm     ScalarNormalizer",
		},
		filepath.Join("..", "core", "collection_common.go"): {
			"ScalarValues   []float64",
			"func (c *Collection) setArray(values []float64) error",
			"func (c *Collection) setColormap(name string)",
			"func (c *Collection) setNorm(norm ScalarNormalizer) error",
			"func (c *Collection) setCLim(vmin, vmax float64) error",
		},
		filepath.Join("..", "core", "collection_path.go"): {
			"func (c *PathCollection) SetArray(values []float64) error",
			"func (c *PathCollection) SetColormap(name string)",
			"func (c *PathCollection) SetNorm(norm ScalarNormalizer) error",
			"func (c *PathCollection) SetCLim(vmin, vmax float64) error",
			"c.refreshScalarMappedColors()",
		},
		filepath.Join("..", "core", "collection_quadmesh.go"): {
			"func (m *QuadMesh) SetArray(values []float64) error",
			"func (m *QuadMesh) SetColormap(name string)",
			"func (m *QuadMesh) SetNorm(norm ScalarNormalizer) error",
			"func (m *QuadMesh) SetCLim(vmin, vmax float64) error",
			"m.refreshScalarMappedColors()",
		},
		filepath.Join("..", "core", "colorbar.go"): {
			"Mappable    ScalarMappable",
			"mappable.ScalarMap().Resolved()",
			"func syncColorbarMapping(ax *Axes)",
			"mapping := cb.currentMapping()",
			"configureColorbarScale(ax, mapping, ax.colorbarLocation, ax.colorbarTicks, ax.colorbarBounds, ax.colorbarExtend)",
			"func (c *Colorbar) currentMapping() ScalarMapInfo",
			"if c.Mappable != nil",
			"mapping = c.Mappable.ScalarMap()",
		},
		filepath.Join("..", "core", "colorbar_test.go"): {
			"TestFigureColorbarSyncsMutableCollectionMapping",
			"pc.SetColormap(\"plasma\")",
			"pc.SetCLim(-1, 2)",
		},
		filepath.Join("..", "test", "parity", "collection_mutable_scalarmap", "plot.go"): {
			"mesh.SetColormap(\"plasma\")",
			"mesh.SetCLim(-0.5, 1.0)",
			"mesh.SetArray([]float64{",
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
				t.Fatalf("%s missing mutable colorbar marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Mutable Mappable Update Contract",
		"Matplotlib colorbars attach to scalar mappables through `mappable.callbacks.connect('changed', Colorbar.update_normal)`",
		"`Colorbar.update_normal` pulls alpha, colormap, and norm from the mappable",
		"Go colorbars store the typed `ScalarMappable` handle and refresh through `syncColorbarMapping` during layout or draw",
		"The supported Go update path is explicit mutation followed by redraw",
		"`SetArray`, `SetColormap`, `SetNorm`, and `SetCLim` refresh scalar-derived collection colors and scalar-map metadata",
		"Colorbar cmap, norm, and clim changes synchronize from `ScalarMap()` on redraw",
		"scalar-array changes are supported on PathCollection, LineCollection, PatchCollection, and QuadMesh",
		"alpha is not part of `ScalarMapInfo` and colorbar alpha does not follow mappable alpha automatically",
		"Matplotlib callback registries and `Colorbar.update_normal` side effects remain documented residuals",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar mutable update docs missing %q", phrase)
		}
	}
}

func TestColorbarMutableUpdateTestsAndOmissionsAreDocumented(t *testing.T) {
	sourceRequirements := map[string][]string{
		filepath.Join("..", "core", "colorbar_test.go"): {
			"TestFigureColorbarSyncsMutableCollectionMapping",
			"pc.SetCLim(-1, 2)",
			"pc.SetColormap(\"plasma\")",
			"TestFigureColorbarSyncsMutableCollectionNormScale",
			"pc.SetNorm(LogNorm{VMin: 0.1, VMax: 10})",
			"want transform.Log",
		},
		filepath.Join("..", "core", "collection_test.go"): {
			"TestPathCollectionSetArrayRefreshesMappedFacesAndFaceEdges",
			"TestLineCollectionSetArrayRefreshesStrokeColors",
			"TestQuadMeshSetArrayRefreshesFlatColorsAndFaceEdges",
			"TestQuadMeshSetArrayKeepsBadUnderOverColorsAfterMappingChanges",
		},
		filepath.Join("..", "test", "parity", "collection_mutable_scalarmap", "plot.go"): {
			"mesh.SetArray([]float64{",
			"mesh.SetColormap(\"plasma\")",
			"mesh.SetCLim(-0.5, 1.0)",
			"core.ColorbarOptions{Label: \"updated\"}",
		},
		filepath.Join("..", "core", "rasterization.go"): {
			"func (a *ArtistRasterization) SetAlpha(alpha float64)",
		},
		filepath.Join("..", "core", "scalar_mappable.go"): {
			"type ScalarMapInfo struct",
			"Colormap string",
			"Norm     ScalarNormalizer",
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
				t.Fatalf("%s missing colorbar update test marker %q", path, phrase)
			}
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	docText := strings.Join(strings.Fields(string(data)), " ")
	requiredDocs := []string{
		"Phase 17.75.5 Colorbar Mutable Update Tests and Omissions",
		"`TestFigureColorbarSyncsMutableCollectionMapping` covers post-creation `SetCLim` and `SetColormap` synchronization",
		"`TestFigureColorbarSyncsMutableCollectionNormScale` covers post-creation `SetNorm` synchronization to a log colorbar scale",
		"Collection tests cover `SetArray` refresh for PathCollection, LineCollection, and QuadMesh",
		"`collection_mutable_scalarmap` provides the visible parity fixture for `SetArray`, `SetColormap`, `SetCLim`, and a colorbar",
		"Alpha mutation remains documented as an omission from colorbar synchronization because `ScalarMapInfo` carries colormap, norm, vmin, and vmax, not artist alpha",
		"Matplotlib callback-driven immediate `update_normal` redraw semantics remain outside the Go API",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(docText, phrase) {
			t.Fatalf("colorbar mutable update test docs missing %q", phrase)
		}
	}
}
