package core

import (
	"strings"
	"testing"

	"github.com/cwbudde/matplotlib-go/internal/diag"
)

// A Gouraud-shaded mesh drawn to a renderer without Gouraud-triangle support
// must surface the silent downgrade to flat shading as a warning — once, not
// on every redraw.
func TestGouraudMeshWarnsOnceWhenRendererLacksCapability(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	mesh := &QuadMesh{
		Shading: MeshShadingGouraud,
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1},
		Values:  [][]float64{{0, 0.5, 1}, {0.25, 0.5, 0.75}},
	}

	r := &recordingRenderer{} // embeds render.NullRenderer; no GouraudTriangleDrawer
	ctx := createTestDrawContext()

	mesh.Draw(r, ctx)
	mesh.Draw(r, ctx)

	if len(warnings) != 1 {
		t.Fatalf("expected exactly 1 downgrade warning across two draws, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(strings.ToLower(warnings[0]), "gouraud") {
		t.Fatalf("warning should mention Gouraud, got %q", warnings[0])
	}
}

// A flat mesh on the same renderer must not warn — the downgrade notice is
// specific to an unsatisfiable Gouraud request.
func TestFlatMeshDoesNotWarn(t *testing.T) {
	var warnings []string
	restore := diag.SetHandler(func(m string) { warnings = append(warnings, m) })
	defer restore()

	mesh := &QuadMesh{
		Shading: MeshShadingFlat,
		XEdges:  []float64{0, 1, 2},
		YEdges:  []float64{0, 1},
		Values:  [][]float64{{0, 0.5}},
	}
	mesh.Draw(&recordingRenderer{}, createTestDrawContext())

	if len(warnings) != 0 {
		t.Fatalf("flat mesh should not warn, got %v", warnings)
	}
}
