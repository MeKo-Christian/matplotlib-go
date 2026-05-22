package agg

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
)

func TestAGGBackend_PopulatesSaveFormats(t *testing.T) {
	info, ok := backends.DefaultRegistry.Get(backends.AGG)
	if !ok {
		t.Fatal("AGG backend not registered")
	}
	if info.SaveFormats == nil {
		t.Fatal("AGG SaveFormats map is nil")
	}
	if _, ok := info.SaveFormats[".png"]; !ok {
		t.Fatalf("AGG SaveFormats missing .png handler; got keys: %v", mapKeys(info.SaveFormats))
	}
}

func TestAGGRendererCapabilityStatus(t *testing.T) {
	renderer, err := backends.Create(backends.AGG, backends.TestDefaultConfig(120, 80))
	if err != nil {
		t.Fatalf("Create(AGG) error: %v", err)
	}

	tests := []struct {
		capability backends.Capability
		want       backends.CapabilityStatus
	}{
		{backends.VectorOutput, backends.CapabilityUnsupported},
		{backends.DPIAware, backends.CapabilityNative},
		{backends.TextShaping, backends.CapabilityNative},
		{backends.FontHinting, backends.CapabilityNative},
		{backends.TextBounds, backends.CapabilityNative},
		{backends.TextPathing, backends.CapabilityNative},
		{backends.RotatedText, backends.CapabilityNative},
		{backends.VerticalText, backends.CapabilityNative},
		{backends.ImageTransform, backends.CapabilityNative},
		{backends.RGBABuffer, backends.CapabilityNative},
		{backends.BufferRegion, backends.CapabilityNative},
		{backends.OffscreenFilter, backends.CapabilityNative},
		{backends.PatternFill, backends.CapabilityNative},
		{backends.GradientFill, backends.CapabilityNative},
		{backends.PathEffects, backends.CapabilityNative},
		{backends.MarkerBatch, backends.CapabilityNative},
		{backends.PathCollectionBatch, backends.CapabilityNative},
		{backends.QuadMeshBatch, backends.CapabilityNative},
		{backends.GouraudTriangleBatch, backends.CapabilityNative},
		{backends.NativeHatcher, backends.CapabilityNative},
		{backends.PNGExport, backends.CapabilityNative},
		{backends.SVGExport, backends.CapabilityUnsupported},
	}

	for _, tc := range tests {
		got := backends.RendererCapabilityStatus(backends.AGG, renderer, tc.capability)
		if got != tc.want {
			t.Fatalf("AGG RendererCapabilityStatus(%s) = %s, want %s", tc.capability, got, tc.want)
		}
	}
}

func mapKeys(m map[string]backends.SaveHandler) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
