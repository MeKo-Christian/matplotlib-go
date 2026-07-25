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

type publicAPITieringArtifact struct {
	SchemaVersion int `json:"schema_version"`
	Baseline      struct {
		Path        string `json:"path"`
		SHA256      string `json:"sha256"`
		SymbolCount int    `json:"symbol_count"`
	} `json:"baseline"`
	Symbols []publicAPITieringRow `json:"symbols"`
}

type publicAPITieringRow struct {
	Package       string `json:"package"`
	ID            string `json:"id"`
	Disposition   string `json:"disposition"`
	Rationale     string `json:"rationale"`
	Replacement   string `json:"replacement"`
	TargetPackage string `json:"target_package"`
}

func TestPublicAPITieringMatchesPreBreakSnapshot(t *testing.T) {
	root := repoRootForAPIAudit(t)
	tieringPath := filepath.Join(root, "docs", "plans", "phase2-public-api-tiering.json")
	data, err := os.ReadFile(tieringPath)
	if err != nil {
		t.Fatalf("read Phase 2 API tiering artifact: %v", err)
	}
	var tiering publicAPITieringArtifact
	if err := json.Unmarshal(data, &tiering); err != nil {
		t.Fatalf("decode Phase 2 API tiering artifact: %v", err)
	}
	if tiering.SchemaVersion != 1 {
		t.Fatalf("Phase 2 API tiering schema = %d, want 1", tiering.SchemaVersion)
	}

	frozenPath := filepath.Join(root, filepath.FromSlash(tiering.Baseline.Path))
	frozenData, err := os.ReadFile(frozenPath)
	if err != nil {
		t.Fatalf("read frozen public API: %v", err)
	}
	sum := sha256.Sum256(frozenData)
	if got := hex.EncodeToString(sum[:]); got != tiering.Baseline.SHA256 {
		t.Fatalf("Phase 2 API baseline hash = %s, frozen API hash = %s; review and regenerate the tiering decisions", tiering.Baseline.SHA256, got)
	}

	var frozen stableAPIArtifact
	if err := json.Unmarshal(frozenData, &frozen); err != nil {
		t.Fatalf("decode frozen public API: %v", err)
	}
	want := make(map[string]stableAPISymbol)
	for _, pkg := range frozen.Packages {
		for _, symbol := range pkg.Symbols {
			key := tieringKey(pkg.Dir, symbol.ID)
			if _, exists := want[key]; exists {
				t.Fatalf("duplicate frozen public API symbol %q", key)
			}
			want[key] = symbol
		}
	}
	if tiering.Baseline.SymbolCount != len(want) {
		t.Fatalf("Phase 2 API baseline count = %d, frozen API has %d symbols", tiering.Baseline.SymbolCount, len(want))
	}
	if len(tiering.Symbols) != len(want) {
		t.Fatalf("Phase 2 API tiering has %d rows, want %d", len(tiering.Symbols), len(want))
	}

	got := make(map[string]publicAPITieringRow, len(tiering.Symbols))
	for _, row := range tiering.Symbols {
		key := tieringKey(row.Package, row.ID)
		if _, exists := got[key]; exists {
			t.Errorf("duplicate Phase 2 API tiering row %q", key)
			continue
		}
		got[key] = row
		if _, exists := want[key]; !exists {
			t.Errorf("Phase 2 API tiering contains unknown symbol %q", key)
		}
		switch row.Disposition {
		case "keep":
			if row.TargetPackage != "" || row.Replacement != "" {
				t.Errorf("%s: keep row must not name a target or replacement", key)
			}
		case "demote":
			if strings.TrimSpace(row.TargetPackage) == "" {
				t.Errorf("%s: demote row must name target_package", key)
			}
			requireTieringExplanation(t, key, &row)
		case "delete":
			if row.TargetPackage != "" {
				t.Errorf("%s: delete row must not name target_package", key)
			}
			requireTieringExplanation(t, key, &row)
		default:
			t.Errorf("%s: invalid disposition %q", key, row.Disposition)
		}
	}
	for key := range want {
		if _, exists := got[key]; !exists {
			t.Errorf("frozen public API symbol %q has no Phase 2 tiering row", key)
		}
	}

	assertPublicAPITieringLandmarks(t, got, want)
}

func requireTieringExplanation(t *testing.T, key string, row *publicAPITieringRow) {
	t.Helper()
	if strings.TrimSpace(row.Rationale) == "" {
		t.Errorf("%s: %s row must have a rationale", key, row.Disposition)
	}
	if strings.TrimSpace(row.Replacement) == "" {
		t.Errorf("%s: %s row must name a replacement or migration path", key, row.Disposition)
	}
}

func assertPublicAPITieringLandmarks(t *testing.T, got map[string]publicAPITieringRow, want map[string]stableAPISymbol) {
	t.Helper()
	introspection := []string{
		"func Findobj",
		"func FindobjType",
		"func Getp",
		"func GetpAll",
		"func Setp",
		"method ArtistRasterization.Property",
		"method ArtistRasterization.PropertyNames",
		"method ArtistRasterization.SetProperty",
		"method Axes.Findobj",
		"method Figure.Findobj",
		"method Line2D.Property",
		"method Line2D.PropertyNames",
		"method Line2D.SetProperty",
		"type PropertyBag",
	}
	for _, id := range introspection {
		assertTieringDisposition(t, got, "core", id, "delete")
	}
	for _, id := range []string{
		"method Axes.BarUnits",
		"method Axes.FillBetweenUnits",
		"method Axes.PlotUnits",
		"method Axes.ScatterUnits",
	} {
		assertTieringDisposition(t, got, "core", id, "delete")
	}
	assertTieringDisposition(t, got, "render", "type CapabilityBridgeReporter", "delete")
	assertTieringDisposition(t, got, "render", "type RendererModeReporter", "demote")
	if row := got[tieringKey("render", "type RendererModeReporter")]; row.TargetPackage != "backends" {
		t.Errorf("render.RendererModeReporter target = %q, want backends", row.TargetPackage)
	}

	for key, symbol := range want {
		if !strings.HasPrefix(key, "render\x00") ||
			!strings.HasPrefix(symbol.Declaration, "type ") ||
			!strings.Contains(symbol.Declaration, " interface{") {
			continue
		}
		if symbol.ID == "type CapabilityBridgeReporter" || symbol.ID == "type RendererModeReporter" {
			continue
		}
		if row := got[key]; row.Disposition != "keep" {
			t.Errorf("%s: renderer SPI disposition = %q, want keep", key, row.Disposition)
		}
	}
}

func assertTieringDisposition(t *testing.T, rows map[string]publicAPITieringRow, pkg, id, want string) {
	t.Helper()
	key := tieringKey(pkg, id)
	row, ok := rows[key]
	if !ok {
		t.Errorf("missing landmark Phase 2 API decision %q", key)
		return
	}
	if row.Disposition != want {
		t.Errorf("%s: disposition = %q, want %q", key, row.Disposition, want)
	}
}

func tieringKey(pkg, id string) string {
	return pkg + "\x00" + id
}
