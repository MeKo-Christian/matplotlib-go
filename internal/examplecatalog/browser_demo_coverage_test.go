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

func TestPlannedWebReferenceModulesAreActiveOrReferenceOnly(t *testing.T) {
	wantActive := []string{
		"annotations",
		"bars",
		"errorbars",
		"fills",
		"heatmap",
		"histogram",
		"lines",
		"patches",
		"scatter",
		"subplots",
	}
	for _, module := range wantActive {
		row, ok := LookupBrowserDemoCoverage("webref-" + module)
		if !ok {
			t.Fatalf("missing web reference coverage row for %q", module)
		}
		if row.ReferenceModule != module {
			t.Fatalf("%s ReferenceModule = %q, want %q", row.ID, row.ReferenceModule, module)
		}
		if row.Status != BrowserDemoActive {
			t.Fatalf("%s Status = %s, want %s", row.ID, row.Status, BrowserDemoActive)
		}
		if row.ActiveWebDemoID == "" {
			t.Fatalf("%s has no active web demo mapping", row.ID)
		}
		active, ok := LookupWebDemo(row.ActiveWebDemoID)
		if !ok {
			t.Fatalf("%s maps to missing web demo %q", row.ID, row.ActiveWebDemoID)
		}
		if !containsString(row.CatalogIDs, active.ID) {
			t.Fatalf("%s CatalogIDs = %v, want active catalog case %q", row.ID, row.CatalogIDs, active.ID)
		}
	}

	row, ok := LookupBrowserDemoCoverage("webref-radialforce")
	if !ok {
		t.Fatal("missing radialforce web reference coverage row")
	}
	if row.Status != BrowserDemoReferenceOnly {
		t.Fatalf("radialforce Status = %s, want %s", row.Status, BrowserDemoReferenceOnly)
	}
	if row.ActiveWebDemoID != "" {
		t.Fatalf("radialforce ActiveWebDemoID = %q, want empty reference-only mapping", row.ActiveWebDemoID)
	}
}

func TestActiveWebReferenceModulesMapToCatalogCases(t *testing.T) {
	for _, row := range BrowserDemoCoverageRows() {
		if row.ReferenceModule == "" || row.Status != BrowserDemoActive {
			continue
		}
		if row.ActiveWebDemoID == "" {
			t.Fatalf("%s is active but has no active web demo ID", row.ID)
		}
		c, ok := LookupWebDemo(row.ActiveWebDemoID)
		if !ok {
			t.Fatalf("%s maps to missing web demo %q", row.ID, row.ActiveWebDemoID)
		}
		if !containsString(row.CatalogIDs, c.ID) {
			t.Fatalf("%s CatalogIDs = %v, want active web demo catalog case %q", row.ID, row.CatalogIDs, c.ID)
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

func TestAnimationGalleryBrowserCoverageIsActive(t *testing.T) {
	row, ok := LookupBrowserDemoCoverage("showcase-animation_gallery")
	if !ok {
		t.Fatal("missing animation gallery browser coverage row")
	}
	if row.Status != BrowserDemoActive {
		t.Fatalf("animation gallery browser coverage status = %s, want %s", row.Status, BrowserDemoActive)
	}
	if row.ActiveWebDemoID != "animation" {
		t.Fatalf("animation gallery ActiveWebDemoID = %q, want animation", row.ActiveWebDemoID)
	}
	if _, ok := LookupWebDemo(row.ActiveWebDemoID); !ok {
		t.Fatalf("animation gallery active web demo %q does not resolve", row.ActiveWebDemoID)
	}
}
