package examplecatalog

import "testing"

func TestPhase9EnumerableCatalogAuditsAreComplete(t *testing.T) {
	audits := Phase9EnumerableCatalogAudits()
	if len(audits) == 0 {
		t.Fatal("missing Phase 9 enumerable catalog audits")
	}

	seen := map[string]bool{}
	for _, audit := range audits {
		if audit.ID == "" || audit.Title == "" || audit.Decision == "" {
			t.Fatalf("incomplete Phase 9 audit row: %+v", audit)
		}
		if seen[audit.ID] {
			t.Fatalf("duplicate Phase 9 audit row %q", audit.ID)
		}
		seen[audit.ID] = true
		if len(audit.UpstreamSources) == 0 {
			t.Fatalf("%s has no upstream source anchors", audit.ID)
		}
		if len(audit.GuardTests) == 0 {
			t.Fatalf("%s has no machine-checked guard tests", audit.ID)
		}
		if audit.Decision == Phase9DecisionImplemented && len(audit.CatalogIDs) == 0 {
			t.Fatalf("%s is implemented but has no catalog-visible fixture IDs", audit.ID)
		}
		for _, id := range audit.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", audit.ID, id)
			}
		}
	}

	for _, id := range []string{
		"colormap-registry",
		"marker-registry",
		"named-colors",
		"image-interpolation-registry",
		"patch-style-registries",
		"hatch-registry",
		"sketch-xkcd",
		"figimage",
		"rcparams",
	} {
		if !seen[id] {
			t.Fatalf("missing Phase 9 audit row %q", id)
		}
	}
}

func TestPhase9ExitCriteriaAreSatisfied(t *testing.T) {
	criteria := Phase9ExitCriteria()
	if len(criteria) != 3 {
		t.Fatalf("Phase 9 exit criteria count = %d, want 3", len(criteria))
	}
	for _, criterion := range criteria {
		if criterion.ID == "" || criterion.Criterion == "" || criterion.Evidence == "" {
			t.Fatalf("incomplete Phase 9 exit criterion: %+v", criterion)
		}
		if criterion.Status != Phase9ExitSatisfied {
			t.Fatalf("%s status = %s, want %s", criterion.ID, criterion.Status, Phase9ExitSatisfied)
		}
	}
}
