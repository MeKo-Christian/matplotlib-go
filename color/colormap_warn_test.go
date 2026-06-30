package color

import (
	"errors"
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

func TestGetColormapStrictErrorsOnUnknownName(t *testing.T) {
	// Strict variant never warns and never falls back; it returns the error.
	restore := diag.SetHandler(func(string) { t.Fatal("strict lookup must not warn") })
	defer restore()

	if _, err := GetColormapStrict("definitely-not-a-real-colormap"); !errors.Is(err, ErrUnknownColormap) {
		t.Fatalf("GetColormapStrict err = %v, want ErrUnknownColormap", err)
	}
}

func TestGetColormapStrictReturnsKnownAndReversed(t *testing.T) {
	cm, err := GetColormapStrict("viridis")
	if err != nil {
		t.Fatalf("GetColormapStrict(viridis) err = %v", err)
	}
	if cm.Name() != "viridis" {
		t.Fatalf("name = %q, want viridis", cm.Name())
	}
	if _, err := GetColormapStrict("plasma_r"); err != nil {
		t.Fatalf("GetColormapStrict(plasma_r) err = %v, want nil", err)
	}
}
