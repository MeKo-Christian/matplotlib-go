package core

import (
	"io/fs"
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

func TestLightSourceExampleNeedListIsDocumented(t *testing.T) {
	refDir := filepath.Join("..", "test", "matplotlib_ref", "plots")
	forbidden := []string{"LightSource", "lightsource=", "hillshade(", "shade_rgb(", "blend_mode="}
	err := filepath.WalkDir(refDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".py" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		for _, phrase := range forbidden {
			if strings.Contains(src, phrase) {
				t.Fatalf("%s unexpectedly requires LightSource image-lighting API via %q", path, phrase)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan matplotlib reference plots: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource Example Need List",
		"No committed Python parity fixture imports `LightSource`",
		"No 2D image fixture",
		"`hillshade`, `shade`, or `shade_rgb`",
		"`mplot3d_terrain`",
		"`plot_surface(..., cmap=\"viridis\")`",
		"Matplotlib disables",
		"surface face shading when a colormap is present",
		"`mplot3d_bar3d`",
		"`mplot3d_voxels`",
		"`mplot3d_trisurf3d`",
		"`shade3DFaceColor`",
		"does not require a broad `LightSource` API",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource need-list docs missing %q", phrase)
		}
	}
}

func TestLightSourceHillshadeCoreOmissionIsDocumented(t *testing.T) {
	for _, dir := range []string{"core", "color"} {
		root := filepath.Join("..", dir)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || filepath.Ext(path) != ".go" || filepath.Base(path) == "lightsource_inventory_test.go" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			src := string(data)
			for _, phrase := range []string{"type LightSource", "func Hillshade", "Hillshade("} {
				if strings.Contains(src, phrase) {
					t.Fatalf("%s unexpectedly implements LightSource hillshade API via %q", path, phrase)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s for hillshade API: %v", dir, err)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource Hillshade Core Decision",
		"`hillshade` remains intentionally omitted",
		"`core.LightSource` or `color.LightSource` type is added",
		"`azdeg=315`, `altdeg=45`",
		"`vert_exag=1`, `dx=1`, `dy=1`, and `fraction=1`",
		"No committed",
		"parity fixture requires grayscale hillshade output",
		"`shade3DFaceColor` remains the supported 3D face-shading path",
		"Revisit this decision when a shaded-relief image fixture is added",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource hillshade decision docs missing %q", phrase)
		}
	}
}

func TestLightSourceBlendModeOmissionIsDocumented(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource RGB Blend Mode Decision",
		"`shade` and `shade_rgb` remain intentionally omitted",
		"`blend_overlay`",
		"`blend_soft_light`",
		"`blend_hsv`",
		"callable",
		"blend modes",
		"No committed parity fixture requires RGB shaded-relief blend output",
		"Go",
		"port should not expose a partial LightSource blend API",
		"Existing colormap lookup and mplot3d face shading remain unchanged",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource blend-mode decision docs missing %q", phrase)
		}
	}
}

func TestLightSourceImagePathIntegrationOmissionIsDocumented(t *testing.T) {
	for _, name := range []string{"image.go", "image_api.go", "matrix_helpers.go"} {
		path := filepath.Join("..", "core", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), "LightSource") || strings.Contains(string(data), "Hillshade") {
			t.Fatalf("%s unexpectedly wires LightSource into image rendering", path)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource Image Path Integration Decision",
		"LightSource path is connected",
		"`Image2D`, `imshow`, `matshow`, or",
		"transformed-image rendering",
		"The AGG image backend remains a scalar-image",
		"renderer",
		"No static image fixture requires shaded-relief rendering",
		"avoids coupling image resampling to unimplemented",
		"hillshade/blend semantics",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource image integration docs missing %q", phrase)
		}
	}
}

func TestLightSourceSurfacePathIntegrationDecisionIsDocumented(t *testing.T) {
	for _, name := range []string{"axes3d_contour_surface.go", "axes3d_bar_voxel.go", "axes3d_projection.go"} {
		path := filepath.Join("..", "core", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(data)
		if strings.Contains(src, "LightSource") || strings.Contains(src, "lightsource") {
			t.Fatalf("%s unexpectedly exposes a LightSource surface option", path)
		}
	}

	data, err := os.ReadFile(filepath.Join("..", "docs", "matplotlib-migration-notes.md"))
	if err != nil {
		t.Fatalf("read migration notes: %v", err)
	}
	doc := string(data)
	requiredDocs := []string{
		"Phase 17.75.5 LightSource Surface Path Integration Decision",
		"no public `LightSource` object is connected",
		"`Surface`",
		"`Trisurf`",
		"`Bar3D`",
		"`Voxels`",
		"`shade3DFaceColor` remains the supported mplot3d face-shading implementation",
		"`LightSource(azdeg=225, altdeg=19.4712)`",
		"`[0.3, 1]`",
		"preserves alpha",
		"Colormapped `plot_surface` and `plot_trisurf` paths keep shading disabled",
		"Do not add a `lightsource` option",
	}
	for _, phrase := range requiredDocs {
		if !strings.Contains(doc, phrase) {
			t.Fatalf("LightSource surface integration docs missing %q", phrase)
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
