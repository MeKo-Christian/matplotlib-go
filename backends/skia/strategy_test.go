package skia

import (
	"testing"

	"github.com/cwbudde/matplotlib-go/backends"
)

func TestBackendStrategyDocumentsBuildAndDependencyPolicy(t *testing.T) {
	strategy := BackendStrategy()

	if strategy.BuildTag != "skia" {
		t.Fatalf("BuildTag = %q, want skia", strategy.BuildTag)
	}
	if strategy.Binding != BindingExternalCAPI {
		t.Fatalf("Binding = %q, want %q", strategy.Binding, BindingExternalCAPI)
	}
	// The render mode and GPU status track the skiagpu build tag: with the tag the
	// GPU scaffold is selectable (StatusPlanned), without it GPU stays deferred and
	// the default mode is CPU.
	wantMode := ModeCPU
	wantGPU := StatusDeferred
	if gpuBuildEnabled {
		wantMode = ModeGPU
		wantGPU = StatusPlanned
	}
	if strategy.DefaultMode != wantMode {
		t.Fatalf("DefaultMode = %q, want %q", strategy.DefaultMode, wantMode)
	}
	if strategy.CPUStatus != StatusImplemented {
		t.Fatalf("CPUStatus = %q, want %q", strategy.CPUStatus, StatusImplemented)
	}
	if strategy.GPUStatus != wantGPU {
		t.Fatalf("GPUStatus = %q, want %q", strategy.GPUStatus, wantGPU)
	}
	if strategy.CIDefault != CIDefaultStub {
		t.Fatalf("CIDefault = %q, want %q", strategy.CIDefault, CIDefaultStub)
	}
	if len(strategy.RequiredLibraries) == 0 {
		t.Fatal("strategy should name required external libraries")
	}
}

func TestModeCapabilitiesExposesCPUGPUSplit(t *testing.T) {
	cpu := ModeCapabilities(ModeCPU)
	gpu := ModeCapabilities(ModeGPU)
	if len(cpu) == 0 {
		t.Fatal("CPU mode should advertise optional capabilities")
	}
	if len(gpu) == 0 {
		t.Fatal("GPU mode should advertise optional capabilities")
	}
	// The GPU set is currently a CPU-readback scaffold, so it must at least cover
	// every CPU capability; the helper exists so the sets can diverge later.
	gpuSet := make(map[backends.Capability]bool, len(gpu))
	for _, c := range gpu {
		gpuSet[c] = true
	}
	for _, c := range cpu {
		if !gpuSet[c] {
			t.Fatalf("GPU mode is missing CPU capability %q", c)
		}
	}
	// An unknown mode falls back to the CPU set rather than returning nil.
	if got := ModeCapabilities(RenderMode("bogus")); len(got) != len(cpu) {
		t.Fatalf("unknown mode returned %d capabilities, want CPU set len %d", len(got), len(cpu))
	}
}

func TestNativePathRequirementsDocumentDeferredExternalABI(t *testing.T) {
	requirements := NativePathRequirements()
	if len(requirements) == 0 {
		t.Fatal("NativePathRequirements returned no deferred native paths")
	}

	byPrimitive := make(map[string]NativePathRequirement, len(requirements))
	for _, req := range requirements {
		if req.Primitive == "" {
			t.Fatalf("native path requirement missing primitive: %#v", req)
		}
		if req.Status != StatusDeferred {
			t.Fatalf("%s status = %q, want %q until external ABI lands", req.Primitive, req.Status, StatusDeferred)
		}
		if len(req.Capabilities) == 0 {
			t.Fatalf("%s should map to at least one renderer capability", req.Primitive)
		}
		if len(req.ExternalEntrypoints) == 0 {
			t.Fatalf("%s should name external Skia entrypoints", req.Primitive)
		}
		if req.BlockedBy == "" {
			t.Fatalf("%s should document its blocker", req.Primitive)
		}
		byPrimitive[req.Primitive] = req
	}

	for _, primitive := range []string{
		"SkCanvas::drawAtlas",
		"SkVertices",
		"tiled SkShader",
		"SkSurface::MakeRenderTarget",
	} {
		if _, ok := byPrimitive[primitive]; !ok {
			t.Fatalf("NativePathRequirements missing %s", primitive)
		}
	}

	assertRequirementHasCapability(t, byPrimitive["SkCanvas::drawAtlas"], backends.MarkerBatch)
	assertRequirementHasCapability(t, byPrimitive["SkVertices"], backends.QuadMeshBatch)
	assertRequirementHasCapability(t, byPrimitive["tiled SkShader"], backends.NativeHatcher)
	assertRequirementHasCapability(t, byPrimitive["SkSurface::MakeRenderTarget"], backends.GPUAccel)
}

func assertRequirementHasCapability(t *testing.T, req NativePathRequirement, capability backends.Capability) {
	t.Helper()
	for _, got := range req.Capabilities {
		if got == capability {
			return
		}
	}
	t.Fatalf("%s capabilities = %v, missing %s", req.Primitive, req.Capabilities, capability)
}
