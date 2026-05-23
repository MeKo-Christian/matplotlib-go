package examplecatalog

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type publicSurfaceArtifact struct {
	SchemaVersion int                `json:"schema_version"`
	SourceRoot    string             `json:"source_root"`
	Modules       []string           `json:"modules"`
	Rows          []PublicSurfaceRow `json:"rows"`
}

func TestPublicSurfaceInventoryContainsLandmarkUpstreamRows(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	if artifact.SchemaVersion != 1 {
		t.Fatalf("schema version = %d, want 1", artifact.SchemaVersion)
	}
	want := []string{
		"artist.py:class:Artist",
		"axis.py:class:Axis",
		"ticker.py:class:EngFormatter",
		"scale.py:registry:scale:log",
		"transforms.py:class:Transform",
		"lines.py:class:Line2D",
		"markers.py:registry:marker:*",
		"markers.py:registry:fillstyle:left",
		"collections.py:class:PathCollection",
		"patches.py:registry:arrowstyle:->",
		"text.py:class:Text",
		"legend.py:class:Legend",
		"offsetbox.py:class:OffsetBox",
		"image.py:registry:interpolation:lanczos",
		"colorbar.py:class:Colorbar",
		"cm.py:class:ColormapRegistry",
		"colors.py:class:Normalize",
		"pyplot.py:function:plot",
		"backend_bases.py:class:FigureCanvasBase",
		"backend_tools.py:registry:tool:home",
		"widgets.py:class:Button",
		"animation.py:class:FuncAnimation",
	}
	for _, id := range want {
		if !publicSurfaceArtifactHasRow(artifact, id) {
			t.Fatalf("public surface inventory missing landmark row %q", id)
		}
	}
}

func TestPublicSurfaceInventoryMatchesExtractor(t *testing.T) {
	root := repoRoot(t)
	cmd := exec.Command("python3", filepath.Join(root, "internal", "examplecatalog", "extract_public_surface.py"))
	cmd.Dir = root
	got, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("run public surface extractor: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("run public surface extractor: %v", err)
	}
	want, err := os.ReadFile(filepath.Join(root, "test", "testdata", "parity_surface", "upstream_public_surface.json"))
	if err != nil {
		t.Fatalf("read committed public surface artifact: %v", err)
	}
	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		t.Fatal("committed public surface artifact differs from extractor output; rerun internal/examplecatalog/extract_public_surface.py")
	}
}

func TestPublicSurfaceParityRowsClassifyLandmarkRows(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	for _, row := range PublicSurfaceParityRows() {
		validatePublicSurfaceParityRow(t, artifact, row)
	}
	want := []string{
		"artist.py:class:Artist",
		"lines.py:class:Line2D",
		"markers.py:registry:marker:*",
		"image.py:registry:interpolation:lanczos",
		"pyplot.py:function:plot",
		"widgets.py:class:Button",
		"animation.py:class:FuncAnimation",
	}
	for _, upstreamID := range want {
		if _, ok := LookupPublicSurfaceParityByUpstreamID(upstreamID); !ok {
			t.Fatalf("missing public surface parity classification for %q", upstreamID)
		}
	}
}

func TestPublicSurfaceParityRowsCoverCommittedInventory(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	rows := PublicSurfaceParityRowsForSurface(artifact.Rows)
	if len(rows) != len(artifact.Rows) {
		classified := map[string]bool{}
		for _, row := range rows {
			classified[row.UpstreamID] = true
		}
		var missing []string
		for _, surface := range artifact.Rows {
			if !classified[surface.ID] {
				missing = append(missing, surface.ID)
			}
		}
		t.Fatalf("classified %d public-surface rows, want %d; missing: %v", len(rows), len(artifact.Rows), missing)
	}

	seenIDs := map[string]bool{}
	seenUpstream := map[string]bool{}
	for _, row := range rows {
		validatePublicSurfaceParityRow(t, artifact, row)
		if seenIDs[row.ID] {
			t.Fatalf("duplicate public surface parity row ID %q", row.ID)
		}
		seenIDs[row.ID] = true
		if seenUpstream[row.UpstreamID] {
			t.Fatalf("duplicate public surface upstream classification for %q", row.UpstreamID)
		}
		seenUpstream[row.UpstreamID] = true
	}
}

func TestPublicSurfaceParityRowsReferenceExistingLocalArtifacts(t *testing.T) {
	artifact := loadPublicSurfaceArtifact(t)
	root := repoRoot(t)
	for _, row := range PublicSurfaceParityRowsForSurface(artifact.Rows) {
		for _, path := range row.GoFiles {
			requireFile(t, filepath.Join(root, path))
		}
		for _, id := range row.CatalogIDs {
			if _, ok := Lookup(id); !ok {
				t.Fatalf("%s references missing catalog case %q", row.ID, id)
			}
		}
		for _, id := range row.ExampleIDs {
			c, ok := Lookup(id)
			if !ok {
				t.Fatalf("%s references missing example case %q", row.ID, id)
			}
			if !c.Showcase {
				t.Fatalf("%s references non-showcase example %q", row.ID, id)
			}
		}
	}
}

func loadPublicSurfaceArtifact(t *testing.T) publicSurfaceArtifact {
	t.Helper()
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "test", "testdata", "parity_surface", "upstream_public_surface.json"))
	if err != nil {
		t.Fatalf("read public surface artifact: %v", err)
	}
	var artifact publicSurfaceArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("parse public surface artifact: %v", err)
	}
	return artifact
}

func publicSurfaceArtifactHasRow(artifact publicSurfaceArtifact, id string) bool {
	for _, row := range artifact.Rows {
		if row.ID == id {
			return true
		}
	}
	return false
}

func validatePublicSurfaceParityRow(t *testing.T, artifact publicSurfaceArtifact, row PublicSurfaceParity) {
	t.Helper()
	if row.ID == "" || row.UpstreamID == "" || row.FeatureCoverageID == "" || row.Status == "" || row.Note == "" {
		t.Fatalf("incomplete public surface parity row: %+v", row)
	}
	if !validPublicSurfaceParityStatus(row.Status) {
		t.Fatalf("%s has invalid status %q", row.ID, row.Status)
	}
	if !publicSurfaceArtifactHasRow(artifact, row.UpstreamID) {
		t.Fatalf("%s references missing upstream public surface row %q", row.ID, row.UpstreamID)
	}
	if _, ok := LookupFeatureCoverage(row.FeatureCoverageID); !ok {
		t.Fatalf("%s references missing feature coverage row %q", row.ID, row.FeatureCoverageID)
	}
	if len(row.GoFiles) == 0 {
		t.Fatalf("%s has no local Go file reference", row.ID)
	}
}

func validPublicSurfaceParityStatus(status PublicSurfaceParityStatus) bool {
	switch status {
	case PublicSurfaceDirectEquivalent,
		PublicSurfaceIdiomaticEquivalent,
		PublicSurfacePartial,
		PublicSurfaceNotStarted,
		PublicSurfaceIntentionalOmission:
		return true
	default:
		return false
	}
}
