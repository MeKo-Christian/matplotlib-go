package examplecatalog

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFoundationAPIGapAuditCoversRequiredModules(t *testing.T) {
	want := []string{
		"artist.py",
		"axis.py",
		"ticker.py",
		"scale.py",
		"transforms.py",
		"lines.py",
		"collections.py",
		"patches.py",
		"text.py",
		"image.py",
		"colorbar.py",
		"cm.py",
		"colors.py",
		"pyplot.py",
		"backend_bases.py",
	}
	for _, module := range want {
		if len(FoundationGapsForUpstreamModule(module)) == 0 {
			t.Fatalf("missing foundation API gap audit for upstream module %q", module)
		}
	}
}

func TestFoundationAPIGapsHaveDecisionsAndCoverageRows(t *testing.T) {
	seen := map[string]bool{}
	for _, gap := range FoundationAPIGapAudit() {
		if gap.ID == "" {
			t.Fatal("foundation API gap has empty ID")
		}
		if seen[gap.ID] {
			t.Fatalf("duplicate foundation API gap ID %q", gap.ID)
		}
		seen[gap.ID] = true
		if gap.CoverageID == "" {
			t.Fatalf("%s has empty coverage row ID", gap.ID)
		}
		if _, ok := LookupFeatureCoverage(gap.CoverageID); !ok {
			t.Fatalf("%s references missing coverage row %q", gap.ID, gap.CoverageID)
		}
		if gap.Title == "" || gap.CurrentEquivalent == "" || gap.Gap == "" {
			t.Fatalf("%s has incomplete audit text: %+v", gap.ID, gap)
		}
		if len(gap.UpstreamModules) == 0 {
			t.Fatalf("%s has no upstream module references", gap.ID)
		}
		if gap.Decision == "" {
			t.Fatalf("%s has no implementation decision", gap.ID)
		}
		if gap.Rationale == "" {
			t.Fatalf("%s has no decision rationale", gap.ID)
		}
	}
}

func TestFoundationAPIGapsCoverRequiredFundamentals(t *testing.T) {
	want := []string{
		"artist-properties-callbacks",
		"artist-clipping-transform",
		"ticker-formatter-catalog",
		"tick-artist-model",
		"transform-bbox-paths",
		"line2d-marker-data-semantics",
		"collection-variants-setters",
		"collection-scalar-mapping",
		"patch-style-registries",
		"text-font-layout",
		"text-font-properties",
		"annotation-coordinate-model",
		"offsetbox-static-layout",
		"legend-handler-layout",
		"image-class-breadth",
		"colorbar-orientation-ticks",
		"colors-norms-lightsource",
		"pyplot-wrapper-surface",
		"backend-canvas-manager-lifecycle",
		"widget-interaction-scope",
		"animation-playback-writers",
	}
	for _, id := range want {
		if _, ok := LookupFoundationAPIGap(id); !ok {
			t.Fatalf("missing required foundation API gap %q", id)
		}
	}
}

func TestPatchTextLegendFoundationGapsAreSplitBySurface(t *testing.T) {
	for _, tc := range []struct {
		id      string
		modules []string
	}{
		{"patch-style-registries", []string{"patches.py", "hatch.py"}},
		{"text-font-layout", []string{"text.py", "textpath.py"}},
		{"text-font-properties", []string{"text.py", "font_manager.py", "textpath.py"}},
		{"annotation-coordinate-model", []string{"text.py", "patches.py"}},
		{"offsetbox-static-layout", []string{"offsetbox.py"}},
		{"legend-handler-layout", []string{"legend.py", "legend_handler.py"}},
	} {
		gap, ok := LookupFoundationAPIGap(tc.id)
		if !ok {
			t.Fatalf("missing foundation API gap %q", tc.id)
		}
		if !sameStrings(gap.UpstreamModules, tc.modules) {
			t.Fatalf("%s UpstreamModules = %v, want %v", tc.id, gap.UpstreamModules, tc.modules)
		}
	}
}

func TestImagePyplotBackendFoundationGapsAreSplitBySurface(t *testing.T) {
	for _, tc := range []struct {
		id      string
		modules []string
	}{
		{"image-class-breadth", []string{"image.py"}},
		{"pyplot-wrapper-surface", []string{"pyplot.py", "_pylab_helpers.py"}},
		{"backend-canvas-manager-lifecycle", []string{"backend_bases.py", "backend_tools.py"}},
		{"widget-interaction-scope", []string{"widgets.py"}},
		{"animation-playback-writers", []string{"animation.py"}},
	} {
		gap, ok := LookupFoundationAPIGap(tc.id)
		if !ok {
			t.Fatalf("missing foundation API gap %q", tc.id)
		}
		if !sameStrings(gap.UpstreamModules, tc.modules) {
			t.Fatalf("%s UpstreamModules = %v, want %v", tc.id, gap.UpstreamModules, tc.modules)
		}
	}
}

func TestImageGapClassifiesPColorFast(t *testing.T) {
	gap, ok := LookupFoundationAPIGap("image-class-breadth")
	if !ok {
		t.Fatal("missing image-class-breadth foundation gap")
	}
	if !strings.Contains(gap.CurrentEquivalent, "PColorFast") {
		t.Fatalf("image-class-breadth CurrentEquivalent = %q, want PColorFast classification", gap.CurrentEquivalent)
	}
	if strings.Contains(gap.Gap, "pcolorfast") {
		t.Fatalf("image-class-breadth Gap still treats pcolorfast as unresolved: %q", gap.Gap)
	}
}

// TestImplementGapsHaveCatalogCoverageOrReclassification enforces the // exit criterion that every GapDecisionImplement row is either implemented with
// catalog coverage or deliberately reclassified with rationale. A gap that is
// still marked "implement" must point at a coverage row that carries at least
// one catalog case; otherwise it should be reclassified to an idiomatic
// equivalent or intentional omission with rationale.
func TestImplementGapsHaveCatalogCoverageOrReclassification(t *testing.T) {
	for _, gap := range FoundationAPIGapAudit() {
		if gap.Decision != GapDecisionImplement {
			continue
		}
		if gap.Rationale == "" {
			t.Fatalf("%s is GapDecisionImplement without rationale", gap.ID)
		}
		row, ok := LookupFeatureCoverage(gap.CoverageID)
		if !ok {
			t.Fatalf("%s is GapDecisionImplement but references missing coverage row %q", gap.ID, gap.CoverageID)
		}
		if len(row.CatalogIDs) == 0 {
			t.Fatalf("%s is GapDecisionImplement but coverage row %q has no catalog coverage; "+
				"implement catalog parity or reclassify the gap with rationale", gap.ID, gap.CoverageID)
		}
	}
}

func TestFoundationAPIGapsReferenceExistingGoFiles(t *testing.T) {
	root := repoRoot(t)
	for _, gap := range FoundationAPIGapAudit() {
		for _, path := range gap.GoFiles {
			requireFile(t, filepath.Join(root, path))
		}
	}
}
