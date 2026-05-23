package examplecatalog

import (
	"path/filepath"
	"testing"
)

func TestCatalogCasesHaveCanonicalParitySourcePairs(t *testing.T) {
	root := repoRoot(t)
	for _, c := range Cases() {
		requireFile(t, filepath.Join(root, "test", "parity", c.ID, "plot.go"))
		requireFile(t, filepath.Join(root, "test", "parity", c.ID, "plot.py"))
	}
}

func TestCatalogCasesHaveVisibleMatplotlibReferenceModules(t *testing.T) {
	root := repoRoot(t)
	for _, c := range Cases() {
		requireFile(t, filepath.Join(root, "test", "matplotlib_ref", "plots", c.ID+".py"))
	}
}

func TestPublicFixtureOnlyCasesHaveDemoBreadthAccounting(t *testing.T) {
	for _, c := range Cases() {
		if !c.FixtureOnly || !isPublicFixtureOnlyTopic(c.Topic) {
			continue
		}
		if referenceConsistencyFixtureOnlyClassification(c.ID) != "" {
			continue
		}
		if !demoBreadthGapReferencesCase(c.ID) {
			t.Fatalf("public fixture-only case %q has no demo-breadth gap or reference-consistency classification", c.ID)
		}
	}
}

func TestReferenceConsistencyClassificationsReferenceCatalogCases(t *testing.T) {
	for _, item := range ReferenceConsistencyClassifications() {
		if item.CaseID == "" || item.Classification == "" || item.Rationale == "" {
			t.Fatalf("incomplete reference consistency classification: %+v", item)
		}
		c, ok := Lookup(item.CaseID)
		if !ok {
			t.Fatalf("reference consistency classification references missing case %q", item.CaseID)
		}
		if !c.FixtureOnly {
			t.Fatalf("reference consistency classification %q points at non-fixture-only case", item.CaseID)
		}
	}
}

func demoBreadthGapReferencesCase(id string) bool {
	for _, gap := range DemoBreadthGaps() {
		if containsString(gap.CatalogIDs, id) {
			return true
		}
	}
	return false
}
