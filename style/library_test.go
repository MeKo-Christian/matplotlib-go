package style

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStyleLibraryRegistersDiscoveredThemes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "paper.mplstyle")
	if err := os.WriteFile(path, []byte("figure.dpi: 180\naxes.facecolor: \"#f0f1f2\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := LoadStyleLibrary(dir)
	if err != nil {
		t.Fatalf("LoadStyleLibrary() error = %v", err)
	}
	if len(report.Paths) != 1 || report.Paths[0] != dir {
		t.Fatalf("paths = %v, want [%q]", report.Paths, dir)
	}
	if len(report.Loaded) != 1 || report.Loaded[0].Name != "paper" || report.Loaded[0].Path != path {
		t.Fatalf("loaded = %+v", report.Loaded)
	}
	if len(report.Loaded[0].Report.Applied) != 2 {
		t.Fatalf("applied count = %d, want 2", len(report.Loaded[0].Report.Applied))
	}

	theme, ok := LookupTheme("paper")
	if !ok {
		t.Fatal("expected discovered theme to be registered")
	}
	if got, want := theme.RC.DPI, 180.0; got != want {
		t.Fatalf("theme DPI = %v, want %v", got, want)
	}
	if got := theme.RC.AxesBackground; got.R != 0xf0/255.0 || got.G != 0xf1/255.0 || got.B != 0xf2/255.0 {
		t.Fatalf("axes background = %+v", got)
	}
}

func TestDiscoverStyleLibraryUsesDefaultSearchPaths(t *testing.T) {
	config := t.TempDir()
	styleDir := filepath.Join(config, "matplotlib", "stylelib")
	if err := os.MkdirAll(styleDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(styleDir, "lab.mplstyle"), []byte("font.size: 14\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	home := t.TempDir()
	t.Setenv("MATPLOTLIB_GO_STYLELIB", "")
	t.Setenv("MPLSTYLEPATH", "")
	t.Setenv("MPLCONFIGDIR", "")
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("HOME", home)

	themes, report, err := DiscoverStyleLibrary()
	if err != nil {
		t.Fatalf("DiscoverStyleLibrary() error = %v", err)
	}
	if _, ok := themes["lab"]; !ok {
		t.Fatalf("expected lab theme in discovered map, got %v", mapKeys(themes))
	}
	if len(report.Loaded) != 1 || report.Loaded[0].Name != "lab" {
		t.Fatalf("loaded = %+v", report.Loaded)
	}
}

func TestLoadStyleLibraryAllowsExplicitFileAndLastThemeWins(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	firstPath := filepath.Join(first, "shared.mplstyle")
	secondPath := filepath.Join(second, "shared.mplstyle")
	if err := os.WriteFile(firstPath, []byte("figure.dpi: 101\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(first) error = %v", err)
	}
	if err := os.WriteFile(secondPath, []byte("figure.dpi: 202\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(second) error = %v", err)
	}

	report, err := LoadStyleLibrary(firstPath, second)
	if err != nil {
		t.Fatalf("LoadStyleLibrary() error = %v", err)
	}
	if len(report.Loaded) != 2 {
		t.Fatalf("loaded count = %d, want 2", len(report.Loaded))
	}
	theme, ok := LookupTheme("shared")
	if !ok {
		t.Fatal("expected shared theme to be registered")
	}
	if got, want := theme.RC.DPI, 202.0; got != want {
		t.Fatalf("theme DPI = %v, want %v", got, want)
	}
}

func TestDiscoverStyleLibraryReportsInvalidExplicitPaths(t *testing.T) {
	dir := t.TempDir()
	goodPath := filepath.Join(dir, "good.mplstyle")
	badPath := filepath.Join(dir, "bad.mplstyle")
	if err := os.WriteFile(goodPath, []byte("font.size: 13\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(good) error = %v", err)
	}
	if err := os.WriteFile(badPath, []byte("lines.linewidth: no\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(bad) error = %v", err)
	}
	missing := filepath.Join(dir, "missing")

	themes, report, err := DiscoverStyleLibrary(dir, missing)
	if err == nil {
		t.Fatal("expected error for bad style and missing explicit path")
	}
	var libErr StyleLibraryError
	if !errors.As(err, &libErr) {
		t.Fatalf("error = %T %[1]v, want StyleLibraryError", err)
	}
	if len(libErr) != 2 {
		t.Fatalf("library error count = %d, want 2", len(libErr))
	}
	if _, ok := themes["good"]; !ok {
		t.Fatalf("expected good theme despite skipped paths, got %v", mapKeys(themes))
	}
	if len(report.Loaded) != 1 || report.Loaded[0].Name != "good" {
		t.Fatalf("loaded = %+v", report.Loaded)
	}
	if len(report.Skipped) != 2 {
		t.Fatalf("skipped = %+v", report.Skipped)
	}
}

func TestLoadStyleLibraryRejectsNonStyleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.txt")
	if err := os.WriteFile(path, []byte("figure.dpi: 120\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := LoadStyleLibrary(path)
	if err == nil {
		t.Fatal("expected non-style file error")
	}
	if len(report.Loaded) != 0 || len(report.Skipped) != 1 {
		t.Fatalf("report = %+v", report)
	}
}

func mapKeys(themes map[string]Theme) []string {
	keys := make([]string, 0, len(themes))
	for key := range themes {
		keys = append(keys, key)
	}
	return keys
}
