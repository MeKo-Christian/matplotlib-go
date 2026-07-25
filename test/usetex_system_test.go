package test

import (
	"flag"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cwbudde/matplotlib-go/backends/agg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

var updateUseTeXGolden = flag.Bool("update-usetex-golden", false, "Update system-TeX golden fixtures instead of comparing")

func TestUseTeXArtistPipelineWithSystemToolchain(t *testing.T) {
	requireSystemTeXCommand(t, "latex")
	requireSystemTeXCommand(t, "dvipng")

	img := renderUseTeXFigure(t)
	if got := countNonWhite(img); got == 0 {
		t.Fatal("UseTeX artist pipeline rendered a blank image")
	}
}

func TestUseTeXGoldenWithSystemToolchain(t *testing.T) {
	requireSystemTeXCommand(t, "latex")
	requireSystemTeXCommand(t, "dvipng")

	got := renderUseTeXFigure(t)
	goldenPath := filepath.Join("..", "testdata", "usetex_golden", "basic.png")
	if *updateUseTeXGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("create UseTeX golden dir: %v", err)
		}
		if err := imagecmp.SavePNG(got, goldenPath); err != nil {
			t.Fatalf("update UseTeX golden %s: %v", goldenPath, err)
		}
		t.Skip("updated UseTeX golden image")
	}
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		t.Fatalf("system TeX golden %s is missing (rerun with -update-usetex-golden on a host with latex+dvipng)", goldenPath)
	}

	want, err := imagecmp.LoadPNG(goldenPath)
	if err != nil {
		t.Fatalf("load UseTeX golden %s: %v", goldenPath, err)
	}
	diff, err := imagecmp.ComparePNG(got, want, 1)
	if err != nil {
		t.Fatalf("compare UseTeX golden %s: %v", goldenPath, err)
	}
	if !diff.Identical {
		artifactsDir := writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "usetex_golden"))
		savePNGOrFail(t, got, filepath.Join(artifactsDir, "basic_got.png"))
		savePNGOrFail(t, want, filepath.Join(artifactsDir, "basic_want.png"))
		if err := imagecmp.SaveDiffImage(got, want, 1, filepath.Join(artifactsDir, "basic_diff.png")); err != nil {
			t.Fatalf("save UseTeX golden diff: %v", err)
		}
		t.Fatalf("UseTeX golden mismatch: MaxDiff=%d, MeanAbs=%.2f, PSNR=%.2fdB", diff.MaxDiff, diff.MeanAbs, diff.PSNR)
	}
}

func renderUseTeXFigure(t *testing.T) *image.RGBA {
	t.Helper()

	fig := core.NewFigure(220, 140)
	fig.RC.UseTeX = true
	fig.Text(0.5, 0.5, `$\alpha_i^2 + \frac{1}{2}$`, core.TextOptions{
		HAlign:   core.TextAlignCenter,
		VAlign:   core.TextVAlignMiddle,
		FontSize: 14,
	})

	r, err := agg.New(220, 140, render.Color{R: 1, G: 1, B: 1, A: 1})
	if err != nil {
		t.Fatalf("agg.New: %v", err)
	}
	core.DrawFigure(fig, r)
	return r.Image()
}

func requireSystemTeXCommand(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("skipping system TeX artist pipeline test: %s not found on PATH", name)
	}
}

func countNonWhite(img *image.RGBA) int {
	if img == nil {
		return 0
	}
	count := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if img.RGBAAt(x, y) != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
				count++
			}
		}
	}
	return count
}
