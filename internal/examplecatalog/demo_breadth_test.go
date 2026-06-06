package examplecatalog

import "testing"

func TestDemoBreadthAuditCoversRequiredPriorities(t *testing.T) {
	want := []string{
		"marker-grid",
		"advanced-scatter",
		"bar-variants",
		"fill-variants",
		"histogram-variants",
		"named-colors",
		"colormap-families",
		"image-variants",
		"colorbar-norms-extensions",
		"mathtext-gallery",
		"ticks-scales-formatters",
		"text-layout-gallery",
		"annotation-legend-offsetbox",
		"mplot3d-gallery",
		"projection-toolkit-gallery",
		"unstructured-triangulation",
		"mixed-raster-vector-output",
	}
	for _, id := range want {
		if _, ok := LookupDemoBreadthGap(id); !ok {
			t.Fatalf("missing demo breadth gap %q", id)
		}
	}
}

func TestDemoBreadthGapsReferenceCatalogCases(t *testing.T) {
	for _, gap := range DemoBreadthGaps() {
		if gap.ID == "" || gap.Title == "" || gap.Topic == "" {
			t.Fatalf("demo breadth gap has missing identity fields: %+v", gap)
		}
		if gap.Need == "" || gap.CurrentCoverage == "" || gap.RecommendedDemo == "" {
			t.Fatalf("%s has incomplete audit text: %+v", gap.ID, gap)
		}
		if gap.Priority == "" {
			t.Fatalf("%s has empty priority", gap.ID)
		}
		if len(gap.CatalogIDs) == 0 {
			t.Fatalf("%s has no catalog case references", gap.ID)
		}
		for _, id := range gap.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", gap.ID, id)
			}
		}
		for _, id := range gap.ShowcaseIDs {
			c, ok := Lookup(id)
			if !ok {
				t.Fatalf("%s references missing showcase %q", gap.ID, id)
			}
			if !c.Showcase {
				t.Fatalf("%s references non-showcase %q", gap.ID, id)
			}
		}
		for _, id := range gap.WebDemoIDs {
			if _, ok := LookupWebDemo(id); !ok {
				t.Fatalf("%s references missing web demo %q", gap.ID, id)
			}
		}
	}
}

func TestDemoBreadthGapPrioritiesAreActionable(t *testing.T) {
	for _, gap := range DemoBreadthGaps() {
		switch gap.Priority {
		case DemoBreadthHigh, DemoBreadthMedium, DemoBreadthLow:
		default:
			t.Fatalf("%s has unsupported priority %q", gap.ID, gap.Priority)
		}
		if gap.Priority == DemoBreadthHigh && len(gap.TargetFeatures) < 2 {
			t.Fatalf("%s high-priority gap should name multiple target features", gap.ID)
		}
	}
}

func TestHighPriorityDemoBreadthGapsPointToThinCoverage(t *testing.T) {
	for _, gap := range DemoBreadthGaps() {
		if gap.Priority != DemoBreadthHigh {
			continue
		}
		coveredByThinRowOrFixture := false
		for _, id := range gap.CatalogIDs {
			c, ok := Lookup(id)
			if ok && !c.Showcase {
				coveredByThinRowOrFixture = true
				break
			}
		}
		for _, row := range FeatureCoverageMatrix() {
			if row.Breadth != BreadthThin && row.Breadth != BreadthFixtureOnly && row.UserShowcase != CoveragePending {
				continue
			}
			for _, id := range gap.CatalogIDs {
				if containsString(row.CatalogIDs, id) {
					coveredByThinRowOrFixture = true
					break
				}
			}
			if coveredByThinRowOrFixture {
				break
			}
		}
		if !coveredByThinRowOrFixture {
			t.Fatalf("%s is high-priority but does not map to a thin coverage row or non-showcase fixture", gap.ID)
		}
	}
}

func TestHighImpactToolkitOutputDemoBreadthGapsNameShowcases(t *testing.T) {
	want := map[string][]string{
		"mplot3d-gallery":            {"mplot3d_gallery"},
		"mixed-raster-vector-output": {"mixed_raster_vector"},
	}
	for gapID, showcaseIDs := range want {
		gap, ok := LookupDemoBreadthGap(gapID)
		if !ok {
			t.Fatalf("missing high-impact toolkit/output demo breadth gap %q", gapID)
		}
		for _, id := range showcaseIDs {
			if !containsString(gap.ShowcaseIDs, id) {
				t.Fatalf("%s ShowcaseIDs = %v, want %q", gapID, gap.ShowcaseIDs, id)
			}
		}
	}
}
