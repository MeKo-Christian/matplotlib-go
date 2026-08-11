package test

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/cwbudde/matplotlib-go/backends/pdf"
	_ "github.com/cwbudde/matplotlib-go/backends/ps"
	_ "github.com/cwbudde/matplotlib-go/backends/svg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
	"github.com/cwbudde/matplotlib-go/test/parity"
)

// TestMathTextVectorBackendsRasterizeCloseToAGGGoldens is a coarse "the glyphs
// landed in roughly the right places" check: it compares an external
// rasterizer's view of our PDF/PS/SVG output against the AGG golden, so it
// carries whole-glyph antialiasing and hinting differences by construction.
//
// The bound was a PSNR floor of 18 dB until fixed imagecmp's
// overflowing PSNR accumulator, which had been flattering every case; the floor
// is now stated as its RMSE equivalent and recalibrated against measurement.
// Across the 15 subtests RMSE measures 19.3 (mathtext_basic/svg) to 32.1
// (mathtext_integrals/ps), so 40 leaves ~25% headroom for rasterizer version
// drift on the three external tools.
func TestMathTextVectorBackendsRasterizeCloseToAGGGoldens(t *testing.T) {
	backends := []struct {
		name      string
		ext       string
		render    func(*testing.T, string, string)
		rasterize func(*testing.T, string, image.Point) image.Image
		maxRMSE   float64
		maxMean   float64
	}{
		{name: "pdf", ext: ".pdf", render: renderMathTextPDF, rasterize: rasterizePDF, maxRMSE: 40, maxMean: 18},
		{name: "ps", ext: ".ps", render: renderMathTextPS, rasterize: rasterizePS, maxRMSE: 40, maxMean: 18},
		{name: "svg", ext: ".svg", render: renderMathTextSVG, rasterize: rasterizeSVG, maxRMSE: 40, maxMean: 18},
	}

	for _, id := range mathTextFixtureIDs {
		id := id
		want, err := imagecmp.LoadPNG(filepath.Join("..", "testdata", "golden", id+".png"))
		if err != nil {
			t.Fatalf("load AGG MathText golden %s: %v", id, err)
		}
		for _, backend := range backends {
			backend := backend
			t.Run(id+"/"+backend.name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), id+backend.ext)
				backend.render(t, id, path)
				got := backend.rasterize(t, path, want.Bounds().Size())
				diff, err := imagecmp.ComparePNG(got, want, 255)
				if err != nil {
					t.Fatalf("compare rasterized %s against AGG golden: %v", backend.name, err)
				}
				t.Logf("%s %s: RMSE=%.2f MeanAbs=%.2f MaxDiff=%d", id, backend.name, diff.RMSE, diff.MeanAbs, diff.MaxDiff)
				if diff.RMSE > backend.maxRMSE || diff.MeanAbs > backend.maxMean {
					artifactsDir := writableArtifactsDir(t, filepath.Join("..", "testdata", "_artifacts", "mathtext_vector_raster"))
					prefix := id + "_" + backend.name
					savePNGOrFail(t, got, filepath.Join(artifactsDir, prefix+"_got.png"))
					savePNGOrFail(t, want, filepath.Join(artifactsDir, prefix+"_agg_golden.png"))
					if err := imagecmp.SaveDiffImage(got, want, 10, filepath.Join(artifactsDir, prefix+"_diff.png")); err != nil {
						t.Fatalf("save raster diff: %v", err)
					}
					t.Fatalf("%s %s raster mismatch: RMSE=%.2f (max %.2f), MeanAbs=%.2f (max %.2f), MaxDiff=%d",
						id, backend.name, diff.RMSE, backend.maxRMSE, diff.MeanAbs, backend.maxMean, diff.MaxDiff)
				}
			})
		}
	}
}

func renderMathTextPDF(t *testing.T, id, path string) {
	t.Helper()
	fig := mathTextFigure(t, id)
	if err := fig.Save(path); err != nil {
		t.Fatalf("SavePDF %s: %v", path, err)
	}
}

func renderMathTextPS(t *testing.T, id, path string) {
	t.Helper()
	fig := mathTextFigure(t, id)
	if err := fig.Save(path); err != nil {
		t.Fatalf("SavePS %s: %v", path, err)
	}
}

func renderMathTextSVG(t *testing.T, id, path string) {
	t.Helper()
	fig := mathTextFigure(t, id)
	if err := fig.Save(path); err != nil {
		t.Fatalf("SaveSVG %s: %v", path, err)
	}
}

func mathTextFigure(t *testing.T, id string) *core.Figure {
	t.Helper()
	fig, _, err := parity.Figure(id)
	if err != nil {
		t.Fatalf("parity figure %s: %v", id, err)
	}
	return fig
}

func rasterizePDF(t *testing.T, path string, size image.Point) image.Image {
	t.Helper()
	pdftoppm := requireRasterizer(t, "pdftoppm")
	outBase := filepath.Join(t.TempDir(), "page")
	runRasterizer(t, pdftoppm, "-singlefile", "-png", "-scale-to-x", fmt.Sprint(size.X), "-scale-to-y", fmt.Sprint(size.Y), path, outBase)
	return loadRasterPNG(t, outBase+".png")
}

func rasterizePS(t *testing.T, path string, size image.Point) image.Image {
	t.Helper()
	gs := requireRasterizer(t, "gs")
	outPath := filepath.Join(t.TempDir(), "page.png")
	runRasterizer(
		t, gs,
		"-dSAFER",
		"-dBATCH",
		"-dNOPAUSE",
		"-sDEVICE=pngalpha",
		"-r72",
		fmt.Sprintf("-g%dx%d", size.X, size.Y),
		"-dTextAlphaBits=4",
		"-dGraphicsAlphaBits=4",
		"-sOutputFile="+outPath,
		path,
	)
	return loadRasterPNG(t, outPath)
}

func rasterizeSVG(t *testing.T, path string, size image.Point) image.Image {
	t.Helper()
	convert := requireRasterizer(t, "convert")
	outPath := filepath.Join(t.TempDir(), "page.png")
	runRasterizer(t, convert, "-background", "white", "-alpha", "remove", path, "-resize", fmt.Sprintf("%dx%d!", size.X, size.Y), outPath)
	return loadRasterPNG(t, outPath)
}

func requireRasterizer(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("skipping MathText vector raster comparison: %s not found on PATH", name)
	}
	return path
}

func runRasterizer(t *testing.T, command string, args ...string) {
	t.Helper()
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("skipping MathText vector raster comparison: %s failed: %v\n%s", filepath.Base(command), err, string(out))
	}
}

func loadRasterPNG(t *testing.T, path string) image.Image {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("rasterizer did not produce %s: %v", path, err)
	}
	img, err := imagecmp.LoadPNG(path)
	if err != nil {
		t.Fatalf("load rasterized PNG %s: %v", path, err)
	}
	if s := img.Bounds().Size(); s.X <= 0 || s.Y <= 0 {
		t.Fatalf("invalid rasterized PNG size for %s: %s", path, s)
	}
	return img
}
