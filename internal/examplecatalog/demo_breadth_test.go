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

func TestProjectionToolkitDemoBreadthGapNamesGalleryShowcase(t *testing.T) {
	gap, ok := LookupDemoBreadthGap("projection-toolkit-gallery")
	if !ok {
		t.Fatal("missing projection-toolkit-gallery demo breadth gap")
	}
	if !containsString(gap.ShowcaseIDs, "projection_toolkit_gallery") {
		t.Fatalf("projection-toolkit-gallery ShowcaseIDs = %v, want projection_toolkit_gallery", gap.ShowcaseIDs)
	}
}

func TestUnstructuredTriangulationDemoBreadthGapNamesGalleryShowcase(t *testing.T) {
	gap, ok := LookupDemoBreadthGap("unstructured-triangulation")
	if !ok {
		t.Fatal("missing unstructured-triangulation demo breadth gap")
	}
	if !containsString(gap.ShowcaseIDs, "triangulation_gallery") {
		t.Fatalf("unstructured-triangulation ShowcaseIDs = %v, want triangulation_gallery", gap.ShowcaseIDs)
	}
}

func TestTicksScalesFormattersDemoBreadthGapNamesDedicatedGalleryShowcase(t *testing.T) {
	gap, ok := LookupDemoBreadthGap("ticks-scales-formatters")
	if !ok {
		t.Fatal("missing ticks-scales-formatters demo breadth gap")
	}
	if !containsString(gap.ShowcaseIDs, "ticks_scales_formatters_gallery") {
		t.Fatalf("ticks-scales-formatters ShowcaseIDs = %v, want ticks_scales_formatters_gallery", gap.ShowcaseIDs)
	}
}

func TestMediumPriorityDemoBreadthGapsHaveExamplesOrFollowups(t *testing.T) {
	for _, gap := range DemoBreadthGaps() {
		if gap.Priority != DemoBreadthMedium {
			continue
		}
		if len(gap.ShowcaseIDs) == 0 && gap.RecommendedDemo == "" {
			t.Fatalf("%s has no user-facing showcase or scheduled follow-up rationale", gap.ID)
		}
	}
}

func TestHighPriorityDemoBreadthGapsAreClosedOrSplit(t *testing.T) {
	stalePhrases := []string{
		"not isolated",
		"Only a",
		"remain fixture-heavy",
		"fixture-only",
		"parity-only",
		"Add a",
		"Promote a",
	}
	for _, gap := range DemoBreadthGaps() {
		if gap.Priority != DemoBreadthHigh {
			continue
		}
		if len(gap.ShowcaseIDs) == 0 {
			t.Fatalf("%s has no user-facing showcase and no split", gap.ID)
		}
		for _, phrase := range stalePhrases {
			if containsStringFragment(gap.CurrentCoverage, phrase) {
				t.Fatalf("%s CurrentCoverage still uses stale example-breadth wording %q: %s", gap.ID, phrase, gap.CurrentCoverage)
			}
			if containsStringFragment(gap.RecommendedDemo, phrase) {
				t.Fatalf("%s RecommendedDemo still uses stale example-breadth wording %q: %s", gap.ID, phrase, gap.RecommendedDemo)
			}
		}
		if containsStringFragment(gap.CurrentCoverage, "residual") || containsStringFragment(gap.RecommendedDemo, "residual") {
			if gap.ImplementationSplitRationale == "" {
				t.Fatalf("%s mentions residual work but has no split rationale", gap.ID)
			}
		}
	}
}

func containsStringFragment(s, fragment string) bool {
	for i := 0; i+len(fragment) <= len(s); i++ {
		if s[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
