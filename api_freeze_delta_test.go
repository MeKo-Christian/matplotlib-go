package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type freezeDeltaArtifact struct {
	SchemaVersion int               `json:"schema_version"`
	Baseline      freezeDeltaSource `json:"baseline"`
	Freeze        freezeDeltaSource `json:"freeze"`
	Rows          []freezeDeltaRow  `json:"rows"`
}

type freezeDeltaSource struct {
	Path        string `json:"path"`
	SHA256      string `json:"sha256"`
	SymbolCount int    `json:"symbol_count"`
}

type freezeDeltaRow struct {
	Direction     string `json:"direction"`
	Package       string `json:"package"`
	ID            string `json:"id"`
	Category      string `json:"category"`
	Note          string `json:"note"`
	TargetPackage string `json:"target_package"`
	TargetID      string `json:"target_id"`
	SourcePackage string `json:"source_package"`
	SourceID      string `json:"source_id"`
}

// TestPublicAPIFreezeDeltaIsReconciled is the prerequisite guard: the
// live freeze may differ from the tiering baseline only in ways the
// delta artifact accounts for row by row. The artifact pins both inputs by
// hash, so any change to the exported surface or to the tiering decisions
// invalidates it until docs/plans/generate_api_freeze_delta.py is rerun --
// and that generator refuses to emit an artifact with an unclassified symbol.
func TestPublicAPIFreezeDeltaIsReconciled(t *testing.T) {
	root := repoRootForAPIAudit(t)
	deltaPath := filepath.Join(root, "docs", "plans", "api-freeze-delta.json")
	data, err := os.ReadFile(deltaPath)
	if err != nil {
		t.Fatalf("read public API freeze delta artifact: %v", err)
	}
	var delta freezeDeltaArtifact
	if err := json.Unmarshal(data, &delta); err != nil {
		t.Fatalf("decode public API freeze delta artifact: %v", err)
	}
	if delta.SchemaVersion != 1 {
		t.Fatalf("freeze delta schema = %d, want 1", delta.SchemaVersion)
	}

	baseline := loadFreezeDeltaSource(t, root, delta.Baseline, "tiering baseline")
	freeze := loadFreezeDeltaSource(t, root, delta.Freeze, "public API freeze")

	removed, added := 0, 0
	for i := range delta.Rows {
		row := &delta.Rows[i]
		key := tieringKey(row.Package, row.ID)
		if strings.TrimSpace(row.Category) == "" || strings.TrimSpace(row.Note) == "" {
			t.Errorf("%s: every delta row needs a category and a note", key)
		}
		switch row.Direction {
		case "removed":
			removed++
			if _, ok := baseline[key]; !ok {
				t.Errorf("%s: removed row is not a tiering baseline symbol", key)
			}
			if _, ok := freeze[key]; ok {
				t.Errorf("%s: removed row is still in the freeze", key)
			}
			if row.TargetID != "" {
				target := tieringKey(row.TargetPackage, row.TargetID)
				if _, ok := freeze[target]; !ok {
					t.Errorf("%s: replacement %s is not frozen", key, target)
				}
			}
		case "added":
			added++
			if _, ok := freeze[key]; !ok {
				t.Errorf("%s: added row is not in the freeze", key)
			}
			if _, ok := baseline[key]; ok {
				t.Errorf("%s: added row already existed in the tiering baseline", key)
			}
			if row.SourceID != "" {
				source := tieringKey(row.SourcePackage, row.SourceID)
				if _, ok := baseline[source]; !ok {
					t.Errorf("%s: source %s is not a baseline symbol", key, source)
				}
			}
		default:
			t.Errorf("%s: invalid direction %q", key, row.Direction)
		}
	}

	if got := len(baseline); got != delta.Baseline.SymbolCount {
		t.Errorf("tiering baseline has %d symbols, artifact records %d", got, delta.Baseline.SymbolCount)
	}
	if got := len(freeze); got != delta.Freeze.SymbolCount {
		t.Errorf("public API freeze has %d symbols, artifact records %d", got, delta.Freeze.SymbolCount)
	}
	if want := delta.Baseline.SymbolCount - removed + added; want != delta.Freeze.SymbolCount {
		t.Errorf("delta does not balance: %d baseline - %d removed + %d added = %d, freeze has %d",
			delta.Baseline.SymbolCount, removed, added, want, delta.Freeze.SymbolCount)
	}

	// The generator only emits rows for symbols that actually differ, so a
	// complete artifact plus a balanced count means the two sides agree
	// everywhere else.
	for key := range baseline {
		if _, ok := freeze[key]; !ok {
			if !freezeDeltaHas(delta.Rows, "removed", key) {
				t.Errorf("%s: dropped from the freeze with no delta row", key)
			}
		}
	}
	for key := range freeze {
		if _, ok := baseline[key]; !ok {
			if !freezeDeltaHas(delta.Rows, "added", key) {
				t.Errorf("%s: added to the freeze with no delta row", key)
			}
		}
	}
}

func freezeDeltaHas(rows []freezeDeltaRow, direction, key string) bool {
	for i := range rows {
		if rows[i].Direction == direction && tieringKey(rows[i].Package, rows[i].ID) == key {
			return true
		}
	}
	return false
}

// loadFreezeDeltaSource verifies the recorded hash and returns the source's
// symbol keys. Both inputs use a package list with per-package symbol ids, so
// one decoder covers them.
func loadFreezeDeltaSource(t *testing.T, root string, src freezeDeltaSource, what string) map[string]struct{} {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(src.Path))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", what, err)
	}
	sum := sha256.Sum256(raw)
	if got := hex.EncodeToString(sum[:]); got != src.SHA256 {
		t.Fatalf("%s hash = %s, artifact records %s; rerun docs/plans/generate_api_freeze_delta.py and review the delta",
			what, got, src.SHA256)
	}

	var doc struct {
		Symbols []struct {
			Package string `json:"package"`
			ID      string `json:"id"`
		} `json:"symbols"`
		Packages []struct {
			Dir     string `json:"dir"`
			Symbols []struct {
				ID string `json:"id"`
			} `json:"symbols"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode %s: %v", what, err)
	}

	keys := make(map[string]struct{})
	add := func(pkg, id string) {
		key := tieringKey(pkg, id)
		if _, exists := keys[key]; exists {
			t.Errorf("%s: duplicate symbol %s", what, key)
		}
		keys[key] = struct{}{}
	}
	for _, symbol := range doc.Symbols {
		add(symbol.Package, symbol.ID)
	}
	for _, pkg := range doc.Packages {
		for _, symbol := range pkg.Symbols {
			add(pkg.Dir, symbol.ID)
		}
	}
	if len(keys) == 0 {
		t.Fatalf("%s: %q decoded to no symbols", what, src.Path)
	}
	return keys
}
