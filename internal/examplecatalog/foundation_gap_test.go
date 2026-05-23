package examplecatalog

import (
	"path/filepath"
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
		"image-class-breadth",
		"colorbar-orientation-ticks",
		"colors-norms-lightsource",
		"pyplot-wrapper-surface",
		"backend-canvas-manager-lifecycle",
	}
	for _, id := range want {
		if _, ok := LookupFoundationAPIGap(id); !ok {
			t.Fatalf("missing required foundation API gap %q", id)
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
