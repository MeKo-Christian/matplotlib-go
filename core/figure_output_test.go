package core_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	_ "github.com/cwbudde/matplotlib-go/backends/agg"
	_ "github.com/cwbudde/matplotlib-go/backends/pdf"
	_ "github.com/cwbudde/matplotlib-go/backends/pgf"
	_ "github.com/cwbudde/matplotlib-go/backends/ps"
	_ "github.com/cwbudde/matplotlib-go/backends/svg"
	"github.com/cwbudde/matplotlib-go/core"
	"github.com/cwbudde/matplotlib-go/geom"
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/cwbudde/matplotlib-go/style"
)

func outputTestFigure() *core.Figure {
	fig := core.NewFigure(80, 60)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	_, _ = ax.Plot([]float64{0, 1}, []float64{0, 1}, core.PlotOptions{})
	return fig
}

func TestFigureSaveSelectsBackendFromExtension(t *testing.T) {
	fig := outputTestFigure()
	for _, ext := range []string{".png", ".svg", ".pdf", ".ps", ".eps", ".pgf"} {
		t.Run(ext, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "figure"+ext)
			if err := fig.Save(path); err != nil {
				t.Fatalf("Save(%q): %v", ext, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat output: %v", err)
			}
			if info.Size() == 0 {
				t.Fatal("saved output is empty")
			}
		})
	}
}

func TestFigureWriteToPNG(t *testing.T) {
	fig := outputTestFigure()
	var out bytes.Buffer
	if err := fig.WriteTo(&out, "png"); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("decode PNG: %v", err)
	}
	if got, want := img.Bounds().Dx(), 80; got != want {
		t.Fatalf("image width = %d, want %d", got, want)
	}
}

func TestFigureWriteToSVGForwardsOptions(t *testing.T) {
	fig := outputTestFigure()
	var out bytes.Buffer
	if err := fig.WriteTo(
		&out,
		"svg",
		render.WithSVGMetadata(map[string]string{"Title": "writer title"}),
	); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "writer title") {
		t.Fatal("SVG output does not contain forwarded metadata")
	}
}

func TestFigureImageReturnsDetachedRGBA(t *testing.T) {
	fig := outputTestFigure()
	first, err := fig.Image()
	if err != nil {
		t.Fatalf("Image: %v", err)
	}
	if got, want := first.Bounds().Dx(), 80; got != want {
		t.Fatalf("image width = %d, want %d", got, want)
	}
	first.Pix[0] ^= 0xff

	second, err := fig.Image()
	if err != nil {
		t.Fatalf("second Image: %v", err)
	}
	if first.Pix[0] == second.Pix[0] {
		t.Fatal("mutating returned image affected a later render")
	}
}

func TestFigureImageFrameOffIsTransparent(t *testing.T) {
	fig := core.NewFigure(20, 20)
	fig.RC.Background = [4]float64{0.8, 0.4, 0.2, 1}
	fig.RC.AxesBackground = render.Color{R: 0.1, G: 0.2, B: 0.3, A: 1}
	fig.RC.Figure.FrameOn = false
	fig.AddAxes(geom.Rect{
		Min: geom.Pt{},
		Max: geom.Pt{X: 1, Y: 1},
	})

	got, err := fig.Image()
	if err != nil {
		t.Fatalf("Image: %v", err)
	}
	if pixel := got.RGBAAt(10, 10); pixel != (color.RGBA{}) {
		t.Fatalf("axes interior pixel = %#v, want transparent", pixel)
	}
}

func TestFigureImageReturnsPremultipliedRGBA(t *testing.T) {
	fig := core.NewFigure(3, 2)
	fig.RC.Background = [4]float64{1, 0, 0, 0.25}

	got, err := fig.Image()
	if err != nil {
		t.Fatalf("Image: %v", err)
	}
	pixel := got.RGBAAt(0, 0)
	if pixel.A == 0 {
		t.Fatal("semi-transparent background became fully transparent")
	}
	if pixel.R != pixel.A || pixel.G != 0 || pixel.B != 0 {
		t.Fatalf("RGBAAt(0, 0) = %#v, want premultiplied red with R == A", pixel)
	}
}

func TestFigureOutputRejectsInvalidInputs(t *testing.T) {
	var nilFigure *core.Figure
	if err := nilFigure.Save("out.png"); err == nil {
		t.Fatal("nil Figure.Save succeeded")
	}
	if err := outputTestFigure().WriteTo(nil, "png"); err == nil {
		t.Fatal("WriteTo with nil writer succeeded")
	}
	if err := outputTestFigure().WriteTo(&bytes.Buffer{}, "gif"); err == nil {
		t.Fatal("WriteTo with unsupported format succeeded")
	}
}

func TestFigureSaveKeepsVectorPageSizeIndependentOfSaveDPI(t *testing.T) {
	fig := core.NewFigure(1400, 1150, style.WithDPI(160))
	wants := map[string]string{
		".pdf": "/MediaBox [0 0 630 517.5]",
		".svg": `width="630pt" height="517.5pt" viewBox="0 0 630 517.5"`,
		".ps":  "%%HiResBoundingBox: 0 0 630 517.5",
		".eps": "%%HiResBoundingBox: 0 0 630 517.5",
		".pgf": `\pgfpathrectangle{\pgfpoint{0pt}{0pt}}{\pgfpoint{630pt}{517.5pt}}`,
	}

	for ext, want := range wants {
		ext, want := ext, want
		t.Run(ext, func(t *testing.T) {
			for _, dpi := range []float64{72, 160} {
				path := filepath.Join(t.TempDir(), "figure"+ext)
				if err := fig.Save(path, render.WithSaveDPI(dpi)); err != nil {
					t.Fatalf("Save(%s, dpi=%v): %v", ext, dpi, err)
				}
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("ReadFile(%s): %v", ext, err)
				}
				if !bytes.Contains(data, []byte(want)) {
					t.Fatalf("%s at dpi=%v missing physical size %q", ext, dpi, want)
				}
				if ext == ".ps" && !bytes.Contains(data, []byte("<< /PageSize [630 517.5] >> setpagedevice")) {
					t.Fatalf("PS at dpi=%v missing exact PageSize", dpi)
				}
				if ext == ".pgf" && !bytes.Contains(data, []byte(`\pgftransformscale{1.00375}`)) {
					t.Fatalf("PGF at dpi=%v missing PostScript-point to TeX-point conversion", dpi)
				}
			}
		})
	}
}

func TestFigureSaveDPIControlsRasterAndSVGEmbeddedImageResolution(t *testing.T) {
	fig := core.NewFigure(160, 80, style.WithDPI(160))
	ax := fig.AddAxes(geom.Rect{Min: geom.Pt{X: 0.1, Y: 0.15}, Max: geom.Pt{X: 0.9, Y: 0.85}})
	line, err := ax.Plot([]float64{0, 1}, []float64{0, 1}, core.PlotOptions{})
	if err != nil {
		t.Fatalf("Plot: %v", err)
	}
	line.SetRasterized(true)

	wantPixels := map[int]image.Point{
		72:  {X: 72, Y: 36},
		160: {X: 160, Y: 80},
	}
	dataURI := regexp.MustCompile(`data:image/png;base64,([^" ]+)`)
	for _, dpi := range []int{72, 160} {
		t.Run(fmt.Sprintf("dpi_%d", dpi), func(t *testing.T) {
			pngPath := filepath.Join(t.TempDir(), "figure.png")
			if err := fig.Save(pngPath, render.WithSaveDPI(float64(dpi))); err != nil {
				t.Fatalf("Save PNG: %v", err)
			}
			pngFile, err := os.Open(pngPath)
			if err != nil {
				t.Fatalf("Open PNG: %v", err)
			}
			cfg, err := png.DecodeConfig(pngFile)
			_ = pngFile.Close()
			if err != nil {
				t.Fatalf("DecodeConfig PNG: %v", err)
			}
			if got := image.Pt(cfg.Width, cfg.Height); got != wantPixels[dpi] {
				t.Fatalf("PNG size at %d DPI = %v, want %v", dpi, got, wantPixels[dpi])
			}

			svgPath := filepath.Join(t.TempDir(), "figure.svg")
			if err := fig.Save(svgPath, render.WithSaveDPI(float64(dpi))); err != nil {
				t.Fatalf("Save SVG: %v", err)
			}
			svgData, err := os.ReadFile(svgPath)
			if err != nil {
				t.Fatalf("Read SVG: %v", err)
			}
			if !bytes.Contains(svgData, []byte(`width="72pt" height="36pt" viewBox="0 0 72 36"`)) {
				t.Fatalf("SVG physical page changed at %d DPI", dpi)
			}
			match := dataURI.FindSubmatch(svgData)
			if len(match) != 2 {
				t.Fatalf("SVG at %d DPI has no embedded rasterized PNG", dpi)
			}
			raw, err := base64.StdEncoding.DecodeString(string(match[1]))
			if err != nil {
				t.Fatalf("decode SVG image data: %v", err)
			}
			embedded, err := png.DecodeConfig(bytes.NewReader(raw))
			if err != nil {
				t.Fatalf("decode embedded SVG PNG: %v", err)
			}
			if got := image.Pt(embedded.Width, embedded.Height); got != wantPixels[dpi] {
				t.Fatalf("embedded SVG raster at %d DPI = %v, want %v", dpi, got, wantPixels[dpi])
			}
		})
	}
}
