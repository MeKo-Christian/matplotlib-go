package core_test

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/cwbudde/matplotlib-go/backends/all"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
	"github.com/cwbudde/matplotlib-go/test/imagecmp"
)

func TestCrossBackendVectorExportMatchesRasterGeometryAtMultipleDPIs(t *testing.T) {
	fig := vectorExportDPIProbeFigure()
	backends := []struct {
		name      string
		ext       string
		rasterize func(*testing.T, string, int) image.Image
		maxRMSE   float64
		maxMean   float64
	}{
		{name: "pdf", ext: ".pdf", rasterize: rasterizeVectorDPIProbePDF, maxRMSE: 28, maxMean: 5},
		{name: "svg", ext: ".svg", rasterize: rasterizeVectorDPIProbeSVG, maxRMSE: 28, maxMean: 5},
	}

	for _, dpi := range []int{72, 160} {
		pngPath := filepath.Join(t.TempDir(), fmt.Sprintf("probe-%d.png", dpi))
		if err := fig.Save(pngPath, render.WithSaveDPI(float64(dpi))); err != nil {
			t.Fatalf("save PNG at %d DPI: %v", dpi, err)
		}
		want := loadVectorDPIProbePNG(t, pngPath)

		for _, backend := range backends {
			backend := backend
			t.Run(fmt.Sprintf("%s/%ddpi", backend.name, dpi), func(t *testing.T) {
				vectorPath := filepath.Join(t.TempDir(), "probe"+backend.ext)
				if err := fig.Save(vectorPath, render.WithSaveDPI(float64(dpi))); err != nil {
					t.Fatalf("save %s at %d DPI: %v", backend.name, dpi, err)
				}
				got := backend.rasterize(t, vectorPath, dpi)
				if got.Bounds().Size() != want.Bounds().Size() {
					t.Fatalf("%s raster size at %d DPI = %v, want PNG size %v", backend.name, dpi, got.Bounds().Size(), want.Bounds().Size())
				}
				diff, err := imagecmp.ComparePNG(got, want, 255)
				if err != nil {
					t.Fatalf("compare %s at %d DPI: %v", backend.name, dpi, err)
				}
				t.Logf("%s %d DPI: RMSE=%.2f MeanAbs=%.2f MaxDiff=%d", backend.name, dpi, diff.RMSE, diff.MeanAbs, diff.MaxDiff)
				if diff.RMSE > backend.maxRMSE || diff.MeanAbs > backend.maxMean {
					t.Fatalf("%s geometry mismatch at %d DPI: RMSE=%.2f (max %.2f), MeanAbs=%.2f (max %.2f)",
						backend.name, dpi, diff.RMSE, backend.maxRMSE, diff.MeanAbs, backend.maxMean)
				}
			})
		}
	}
}

func vectorExportDPIProbeFigure() *core.Figure {
	fig := core.NewFigure(320, 200, style.WithDPI(160))
	fig.Artists = append(fig.Artists, core.ArtistFunc(func(r render.Renderer, ctx *core.DrawContext) {
		page := ctx.FigureRect
		pointScale := ctx.RC.DPI / 72
		panel := func(x0, y0, x1, y1 float64) geom.Path {
			var p geom.Path
			p.MoveTo(geom.Pt{X: page.W() * x0, Y: page.H() * y0})
			p.LineTo(geom.Pt{X: page.W() * x1, Y: page.H() * y0})
			p.LineTo(geom.Pt{X: page.W() * x1, Y: page.H() * y1})
			p.LineTo(geom.Pt{X: page.W() * x0, Y: page.H() * y1})
			p.Close()
			return p
		}

		r.Path(panel(0.08, 0.15, 0.46, 0.85), &render.Paint{
			Fill: render.Color{R: 0.9, G: 0.2, B: 0.15, A: 0.35}, Stroke: render.Color{R: 0.45, A: 1}, LineWidth: 2 * pointScale,
		})
		// This opaque panel follows a translucent paint and catches PDF
		// ExtGState leakage in the visual path as well as the structural test.
		r.Path(panel(0.27, 0.30, 0.65, 0.70), &render.Paint{
			Fill: render.Color{R: 0.15, G: 0.45, B: 0.9, A: 1}, Stroke: render.Color{B: 0.45, A: 1}, LineWidth: 3 * pointScale,
		})

		var marker geom.Path
		marker.MoveTo(geom.Pt{X: -1, Y: -1})
		marker.LineTo(geom.Pt{X: 1, Y: -1})
		marker.LineTo(geom.Pt{X: 0, Y: 1})
		marker.Close()
		items := make([]render.MarkerItem, 0, 4)
		for i, frac := range []float64{0.70, 0.78, 0.86, 0.94} {
			items = append(items, render.MarkerItem{
				Offset:    geom.Pt{X: page.W() * frac, Y: page.H() * (0.3 + 0.13*float64(i))},
				Transform: geom.Affine{A: 8 * pointScale, D: 8 * pointScale},
				Paint: render.Paint{
					Fill: render.Color{R: 0.95, G: 0.7, B: 0.1, A: 1}, Stroke: render.Color{R: 0.1, G: 0.1, B: 0.1, A: 1}, LineWidth: 3 * pointScale,
				},
			})
		}
		if drawer, ok := r.(render.MarkerDrawer); ok {
			drawer.DrawMarkers(render.MarkerBatch{Marker: marker, Items: items})
		}
	}))
	return fig
}

func rasterizeVectorDPIProbePDF(t *testing.T, path string, dpi int) image.Image {
	t.Helper()
	command, err := exec.LookPath("pdftoppm")
	if err != nil {
		t.Skip("pdftoppm is required for PDF vector export regression coverage")
	}
	outBase := filepath.Join(t.TempDir(), "page")
	runVectorDPIProbeCommand(t, command, "-singlefile", "-png", "-r", fmt.Sprint(dpi), path, outBase)
	return loadVectorDPIProbePNG(t, outBase+".png")
}

func rasterizeVectorDPIProbeSVG(t *testing.T, path string, dpi int) image.Image {
	t.Helper()
	outPath := filepath.Join(t.TempDir(), "page.png")
	if command, err := exec.LookPath("rsvg-convert"); err == nil {
		runVectorDPIProbeCommand(t, command, "--dpi-x", fmt.Sprint(dpi), "--dpi-y", fmt.Sprint(dpi), "--output", outPath, path)
		return loadVectorDPIProbePNG(t, outPath)
	}
	command, err := exec.LookPath("convert")
	if err != nil {
		t.Skip("rsvg-convert or ImageMagick convert is required for SVG vector export regression coverage")
	}
	runVectorDPIProbeCommand(t, command, "-density", fmt.Sprint(dpi), "-background", "white", "-alpha", "remove", path, outPath)
	return loadVectorDPIProbePNG(t, outPath)
}

func runVectorDPIProbeCommand(t *testing.T, command string, args ...string) {
	t.Helper()
	cmd := exec.Command(command, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s failed: %v\n%s", filepath.Base(command), err, out)
	}
}

func loadVectorDPIProbePNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}
