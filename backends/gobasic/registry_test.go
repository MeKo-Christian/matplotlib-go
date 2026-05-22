package gobasic

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
)

func TestGoBasicRendererCapabilityStatus(t *testing.T) {
	renderer, err := backends.Create(backends.GoBasic, backends.TestDefaultConfig(120, 80))
	if err != nil {
		t.Fatalf("Create(GoBasic) error: %v", err)
	}

	tests := []struct {
		capability backends.Capability
		want       backends.CapabilityStatus
	}{
		{backends.AntiAliasing, backends.CapabilityNative},
		{backends.PathClip, backends.CapabilityNative},
		{backends.DPIAware, backends.CapabilityNative},
		{backends.TextShaping, backends.CapabilityNative},
		{backends.TextPathing, backends.CapabilityNative},
		{backends.RotatedText, backends.CapabilityNative},
		{backends.VerticalText, backends.CapabilityNative},
		{backends.ImageTransform, backends.CapabilityNative},
		{backends.PathEffects, backends.CapabilityNative},
		{backends.PNGExport, backends.CapabilityNative},
		{backends.MarkerBatch, backends.CapabilityFallback},
		{backends.PathCollectionBatch, backends.CapabilityFallback},
		{backends.QuadMeshBatch, backends.CapabilityFallback},
		{backends.NativeHatcher, backends.CapabilityFallback},
		{backends.FontHinting, backends.CapabilityUnsupported},
		{backends.TextBounds, backends.CapabilityUnsupported},
		{backends.PatternFill, backends.CapabilityUnsupported},
		{backends.GradientFill, backends.CapabilityUnsupported},
		{backends.GouraudTriangleBatch, backends.CapabilityUnsupported},
		{backends.SVGExport, backends.CapabilityUnsupported},
	}

	for _, tc := range tests {
		got := backends.RendererCapabilityStatus(backends.GoBasic, renderer, tc.capability)
		if got != tc.want {
			t.Fatalf("GoBasic RendererCapabilityStatus(%s) = %s, want %s", tc.capability, got, tc.want)
		}
	}

	if err := backends.VerifyRendererCapabilities(backends.GoBasic, renderer); err != nil {
		t.Fatalf("VerifyRendererCapabilities(GoBasic) error = %v", err)
	}
}

func TestGoBasicBackendContractSuite(t *testing.T) {
	suite := backends.NewTestSuite(backends.GoBasic, backends.TestDefaultConfig(96, 64))
	suite.RunAll(t)
}
