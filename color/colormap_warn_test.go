package color

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

func TestGetColormapWarnsOnUnknownName(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	cm := GetColormap("definitely-not-a-real-colormap")

	// Fallback to the default is preserved (no panic, usable colormap).
	if cm.Name() != defaultColormapName {
		t.Fatalf("fallback colormap = %q, want %q", cm.Name(), defaultColormapName)
	}
	if len(warnings) == 0 {
		t.Fatal("expected a warning for an unknown colormap name, got none")
	}
}

func TestGetColormapDoesNotWarnForKnownName(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	_ = GetColormap("viridis")
	_ = GetColormap("plasma_r") // reversed variant of a known map

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for known colormaps, got %v", warnings)
	}
}
