package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLightSourceAlgorithmInventoryIsDocumented(t *testing.T) {
	colors := readUpstreamColorsPy(t)
	upstreamRequired := []string{
		"def __init__(self, azdeg=315, altdeg=45, hsv_min_val=0, hsv_max_val=1,",
		"hsv_min_sat=1, hsv_max_sat=0",
		"az = np.radians(90 - self.azdeg)",
		"alt = np.radians(self.altdeg)",
		"def hillshade(self, elevation, vert_exag=1, dx=1, dy=1, fraction=1.):",
		"dy = -dy",
		"e_dy, e_dx = np.gradient(vert_exag * elevation, dy, dx)",
		"normal[..., 0] = -e_dx",
		"normal[..., 1] = -e_dy",
		"normal[..., 2] = 1",
		"return self.shade_normals(normal, fraction)",
		"intensity = normals.dot(self.direction)",
		"intensity *= fraction",
		"if (imax - imin) > 1e-6:",
		"intensity = np.clip(intensity, 0, 1)",
		"def shade(self, data, cmap, norm=None, blend_mode='overlay'",
		"def shade_rgb(self, rgb, elevation, fraction=1., blend_mode='hsv'",
		"'hsv': self.blend_hsv",
		"'soft': self.blend_soft_light",
		"'overlay': self.blend_overlay",
		"return 2 * intensity * rgb + (1 - 2 * intensity) * rgb**2",
		"low = 2 * intensity * rgb",
		"high = 1 - 2 * (1 - intensity) * (1 - rgb)",
	}
	for _, phrase := range upstreamRequired {
		if !strings.Contains(colors, phrase) {
			t.Fatalf("upstream LightSource algorithm missing %q", phrase)
		}
	}

	art3d := readUpstreamMPLToolkitsFile(t, "mplot3d", "art3d.py")
	for _, phrase := range []string{
		"mcolors.LightSource(azdeg=225, altdeg=19.4712)",
		"in_norm = mcolors.Normalize(-1, 1)",
		"out_norm = mcolors.Normalize(0.3, 1).inverse",
		"colors[:, 3] = alpha",
	} {
		if !strings.Contains(art3d, phrase) {
			t.Fatalf("upstream 3D shade algorithm missing %q", phrase)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource Algorithm Inventory",
		"`azdeg=315` and `altdeg=45`",
		"`dy = -dy`",
		"`fraction` multiplies the dot-product intensity",
		"`shade`",
		"`blend_mode='overlay'`",
		"`shade_rgb`",
		"`blend_mode='hsv'`",
		"`blend_overlay`",
		"`blend_soft_light`",
		"`blend_hsv`",
		"3D collection face shading is separate",
		"`LightSource(azdeg=225, altdeg=19.4712)`",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource inventory docs missing %q", phrase)
		}
	}
}

func readUpstreamMPLToolkitsFile(t *testing.T, parts ...string) string {
	t.Helper()
	pathParts := append([]string{"..", "third_party", "matplotlib", "lib", "mpl_toolkits"}, parts...)
	data, err := os.ReadFile(filepath.Join(pathParts...))
	if err != nil {
		t.Fatalf("read upstream mpl_toolkits file %v: %v", parts, err)
	}
	return string(data)
}
