package examplecatalog

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowserDemoCoverageRowsHaveStableIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, row := range BrowserDemoCoverageRows() {
		if row.ID == "" || row.Title == "" {
			t.Fatalf("browser demo coverage row has missing identity: %+v", row)
		}
		if seen[row.ID] {
			t.Fatalf("duplicate browser demo coverage row ID %q", row.ID)
		}
		seen[row.ID] = true
		if row.Status == "" || row.Action == "" || row.Rationale == "" {
			t.Fatalf("%s has incomplete status/action/rationale: %+v", row.ID, row)
		}
	}
}

func TestWebDemoReferenceModulesAreReconciled(t *testing.T) {
	root := repoRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "test", "matplotlib_ref", "webdemos", "*.py"))
	if err != nil {
		t.Fatal(err)
	}
	active := map[string]bool{}
	for _, c := range WebDemos() {
		active[c.WebDemoID] = true
	}
	for _, file := range files {
		id := strings.TrimSuffix(filepath.Base(file), ".py")
		if id == "__init__" {
			continue
		}
		if active[id] {
			continue
		}
		row, ok := LookupBrowserDemoCoverage("webref-" + id)
		if !ok {
			t.Fatalf("unreconciled web reference module %q", id)
		}
		if row.ReferenceModule != id {
			t.Fatalf("%s ReferenceModule = %q, want %q", row.ID, row.ReferenceModule, id)
		}
	}
}

func TestCLIOnlyShowcasesHaveBrowserCoverageRows(t *testing.T) {
	for _, c := range Cases() {
		if !c.Showcase || c.WebDemoID != "" {
			continue
		}
		row, ok := LookupBrowserDemoCoverage("showcase-" + c.ID)
		if !ok {
			t.Fatalf("showcase %q has no browser demo coverage row", c.ID)
		}
		if !containsString(row.CatalogIDs, c.ID) {
			t.Fatalf("%s CatalogIDs = %v, want %q", row.ID, row.CatalogIDs, c.ID)
		}
	}
}

func TestBrowserDemoCoverageReferencesCatalogAndWebDemos(t *testing.T) {
	for _, row := range BrowserDemoCoverageRows() {
		for _, id := range row.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", row.ID, id)
			}
		}
		if row.ActiveWebDemoID != "" {
			if _, ok := LookupWebDemo(row.ActiveWebDemoID); !ok {
				t.Fatalf("%s references missing active web demo %q", row.ID, row.ActiveWebDemoID)
			}
		}
	}
}
