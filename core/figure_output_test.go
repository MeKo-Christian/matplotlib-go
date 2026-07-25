package core_test

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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
)

func outputTestFigure() *core.Figure {
	fig := core.NewFigure(80, 60)
	ax := fig.AddAxes(geom.Rect{
		Min: geom.Pt{X: 0.1, Y: 0.1},
		Max: geom.Pt{X: 0.9, Y: 0.9},
	})
	ax.Plot([]float64{0, 1}, []float64{0, 1})
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
