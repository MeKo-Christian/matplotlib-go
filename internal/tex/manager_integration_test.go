package tex

import (
	"os"
	"os/exec"
	"testing"
)

func TestManagerRenderWithSystemTeXToolchain(t *testing.T) {
	latex, ok := systemCommandPath(t, "latex")
	if !ok {
		return
	}
	dvipng, ok := systemCommandPath(t, "dvipng")
	if !ok {
		return
	}

	manager := NewManager(ManagerConfig{
		CacheDir:      t.TempDir(),
		LaTeXCommand:  latex,
		DVIPNGCommand: dvipng,
	})

	first, err := manager.Render(`$\alpha_i^2 + \frac{1}{2}$`, 12, 120, "DejaVu Serif")
	if err != nil {
		t.Fatalf("Render with system TeX toolchain: %v", err)
	}
	if first.PNGPath == "" {
		t.Fatal("Render returned an empty PNG path")
	}
	if _, err := os.Stat(first.PNGPath); err != nil {
		t.Fatalf("rendered PNG does not exist at %s: %v", first.PNGPath, err)
	}
	if first.Image == nil || first.Image.Bounds().Dx() == 0 || first.Image.Bounds().Dy() == 0 {
		t.Fatalf("Render returned an empty image: %+v", first.Image)
	}
	if first.Metrics.W <= 0 || first.Metrics.H <= 0 || first.Metrics.Ascent <= 0 {
		t.Fatalf("Render returned invalid metrics: %+v", first.Metrics)
	}

	second, err := manager.Render(`$\alpha_i^2 + \frac{1}{2}$`, 12, 120, "DejaVu Serif")
	if err != nil {
		t.Fatalf("cached render with system TeX toolchain: %v", err)
	}
	if second.PNGPath != first.PNGPath {
		t.Fatalf("cached render path changed: first=%q second=%q", first.PNGPath, second.PNGPath)
	}
}

func systemCommandPath(t *testing.T, name string) (string, bool) {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("skipping system TeX integration test: %s not found on PATH", name)
		return "", false
	}
	return path, true
}
