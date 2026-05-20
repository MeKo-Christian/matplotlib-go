package test

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
)

func TestAGGNativeParityFixturesRequireNativeCapabilities(t *testing.T) {
	renderer, err := backends.Create(backends.AGG, backends.TestDefaultConfig(64, 64))
	if err != nil {
		t.Fatalf("Create(AGG): %v", err)
	}

	seen := 0
	for _, c := range examplecatalog.Cases() {
		if c.NativeBackend == "" {
			continue
		}
		seen++
		if c.NativeBackend != string(backends.AGG) {
			t.Fatalf("%s NativeBackend = %q, want %q", c.ID, c.NativeBackend, backends.AGG)
		}
		if !c.FixtureOnly {
			t.Fatalf("%s is AGG-native but not FixtureOnly", c.ID)
		}
		if len(c.NativeCapabilities) == 0 {
			t.Fatalf("%s is AGG-native but has no NativeCapabilities", c.ID)
		}
		for _, name := range c.NativeCapabilities {
			capability := backends.Capability(name)
			if !knownCapability(capability) {
				t.Fatalf("%s names unknown native capability %q", c.ID, name)
			}
			status := backends.RendererCapabilityStatus(backends.AGG, renderer, capability)
			if status != backends.CapabilityNative {
				t.Fatalf("%s requires %s, AGG status = %s, want native", c.ID, capability, status)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no AGG-native parity fixtures cataloged")
	}
}

func TestAGGNativeParityFixturesAreSplitFromRendererNeutralCases(t *testing.T) {
	nativeIDs := map[string]bool{}
	for _, c := range aggNativeCases() {
		if c.NativeBackend != string(backends.AGG) {
			t.Fatalf("aggNativeCases included %s for backend %q", c.ID, c.NativeBackend)
		}
		nativeIDs[c.ID] = true
	}
	for _, c := range rendererNeutralCases() {
		if c.NativeBackend != "" {
			t.Fatalf("rendererNeutralCases included native fixture %s for %q", c.ID, c.NativeBackend)
		}
		if nativeIDs[c.ID] {
			t.Fatalf("%s appears in both native and renderer-neutral case lists", c.ID)
		}
	}
	for _, id := range []string{"large_scatter", "mixed_collection", "quad_mesh", "gouraud_triangles", "clip_path_batch"} {
		if !nativeIDs[id] {
			t.Fatalf("aggNativeCases missing %s", id)
		}
	}
}

func knownCapability(capability backends.Capability) bool {
	for _, known := range backends.AllCapabilities {
		if capability == known {
			return true
		}
	}
	return false
}
