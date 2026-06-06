package examplecatalog

import (
	"path/filepath"
	"testing"
)

func TestFeatureCoverageRowsHaveStableIDsAndStatuses(t *testing.T) {
	seen := map[string]bool{}
	for _, row := range FeatureCoverageMatrix() {
		if row.ID == "" {
			t.Fatal("feature coverage row has empty ID")
		}
		if seen[row.ID] {
			t.Fatalf("duplicate feature coverage row ID %q", row.ID)
		}
		seen[row.ID] = true
		if row.Title == "" {
			t.Fatalf("%s has empty title", row.ID)
		}
		if len(row.UpstreamModules) == 0 && len(row.UpstreamGalleryFamilies) == 0 {
			t.Fatalf("%s has no upstream reference", row.ID)
		}
		if row.GoEquivalent == "" || row.ParityFixture == "" || row.UserShowcase == "" || row.BrowserDemo == "" || row.Breadth == "" {
			t.Fatalf("%s has an empty coverage status: %+v", row.ID, row)
		}
	}
}

func TestFeatureCoverageCoversFundamentalMatplotlibAreas(t *testing.T) {
	want := []string{
		"artist",
		"axes",
		"figure-layout",
		"axis-ticker-scale",
		"transforms",
		"lines",
		"collections",
		"patches",
		"text-annotation-legend",
		"image",
		"colorbar",
		"colors-cm",
		"pyplot-state",
		"renderer-backends",
		"widgets-events-animation",
	}
	for _, id := range want {
		if _, ok := LookupFeatureCoverage(id); !ok {
			t.Fatalf("missing feature coverage row %q", id)
		}
	}
}

func TestFeatureCoverageReferencesCatalogRowsAndWebDemos(t *testing.T) {
	for _, row := range FeatureCoverageMatrix() {
		for _, id := range row.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", row.ID, id)
			}
		}
		for _, id := range row.ExampleIDs {
			c, ok := Lookup(id)
			if !ok {
				t.Fatalf("%s references missing showcase case %q", row.ID, id)
			}
			if !c.Showcase {
				t.Fatalf("%s references non-showcase example %q", row.ID, id)
			}
		}
		for _, id := range row.WebDemoIDs {
			if _, ok := LookupWebDemo(id); !ok {
				t.Fatalf("%s references missing web demo %q", row.ID, id)
			}
		}
	}
}

func TestFeatureCoverageReferencesExistingGoFiles(t *testing.T) {
	root := repoRoot(t)
	for _, row := range FeatureCoverageMatrix() {
		for _, path := range row.GoFiles {
			requireFile(t, filepath.Join(root, path))
		}
	}
}

func TestImplementedFeatureCoverageHasFixtureOrOmission(t *testing.T) {
	for _, row := range FeatureCoverageMatrix() {
		if row.GoEquivalent != CoverageImplemented {
			continue
		}
		if len(row.CatalogIDs) == 0 && row.IntentionalOmission == "" {
			t.Fatalf("%s is implemented but has no parity fixture or intentional omission", row.ID)
		}
	}
}

// TestPartialFeatureCoverageRowsArePrecise enforces the Phase 17 exit criterion
// that every `partial` core feature row is either moved to implemented /
// intentional-omission or kept as a smaller precise partial row: a partial
// GoEquivalent row must document the remaining scope (or carry an explicit
// intentional omission) rather than leaving the partial status unexplained.
func TestPartialFeatureCoverageRowsArePrecise(t *testing.T) {
	for _, row := range FeatureCoverageMatrix() {
		if row.GoEquivalent != CoveragePartial {
			continue
		}
		if row.IntentionalOmission != "" {
			continue
		}
		if !partialNoteDocumentsRemainingScope(row.Notes) {
			t.Fatalf("%s is a partial core feature row without precise remaining-scope notes; "+
				"move it to implemented/intentional-omission or document what is still missing", row.ID)
		}
	}
}
